package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"supercode/internal/config"
	"time"
)

var (
	endpoint   = config.EndPoint
	httpClient = &http.Client{Timeout: 3 * time.Minute}
)

type APIError struct {
	Status     int
	Message    string
	Upstream   bool
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	if e.Status == 0 {
		return e.Message
	}
	return fmt.Sprintf("API error (%d): %s", e.Status, e.Message)
}

func BuildRequest(model string, messages []Message, tools []ToolDef) ([]byte, error) {
	requestBody := ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Tools:    tools,
	}

	jsonData, err := json.Marshal(requestBody)

	if err != nil {
		return nil, fmt.Errorf("couldn't build request: %w", err)
	}

	return jsonData, nil
}

func send(key, model string, messages []Message, tools []ToolDef) (Message, error) {
	data, err := BuildRequest(model, messages, tools)

	if err != nil {
		return Message{}, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(data))

	if err != nil {
		return Message{}, fmt.Errorf("couldn't create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Title", "SuperCode")

	resp, err := httpClient.Do(req)

	if err != nil {
		return Message{}, &APIError{Message: "request failed: " + err.Error()}
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return Message{}, &APIError{Message: "couldn't read response: " + err.Error()}
	}

	if resp.StatusCode != http.StatusOK {
		return Message{}, newAPIError(resp.StatusCode, resp.Header, body)
	}

	return ParseResponse(body)
}

func ParseResponse(responseBody []byte) (Message, error) {
	var chatResponse ChatResponse

	if err := json.Unmarshal(responseBody, &chatResponse); err != nil {
		return Message{}, &APIError{Message: "couldn't parse response: " + err.Error()}
	}

	if chatResponse.Error != nil {
		return Message{}, errorFromInfo(codeOf(chatResponse.Error.Code, http.StatusBadGateway), nil, *chatResponse.Error)
	}

	if len(chatResponse.Choices) == 0 {
		return Message{}, &APIError{Message: "no choices in response"}
	}

	return chatResponse.Choices[0].Message, nil
}

func newAPIError(status int, header http.Header, body []byte) *APIError {
	var parsed struct {
		Error ErrorInfo `json:"error"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil || parsed.Error.Message == "" {
		return &APIError{
			Status:     status,
			Message:    shorten(strings.TrimSpace(string(body)), 300),
			RetryAfter: retryAfter(header, nil),
		}
	}

	return errorFromInfo(status, header, parsed.Error)
}

func errorFromInfo(status int, header http.Header, info ErrorInfo) *APIError {
	lower := strings.ToLower(info.Message)

	return &APIError{
		Status:     status,
		Message:    shorten(info.Message, 300),
		Upstream:   info.Metadata.ProviderName != "" || strings.Contains(lower, "upstream") || strings.Contains(lower, "provider"),
		RetryAfter: retryAfter(header, info.Metadata.Headers),
	}
}

func retryAfter(header http.Header, metadata map[string]any) time.Duration {
	if header != nil {
		if s := header.Get("Retry-After"); s != "" {
			if n, err := strconv.Atoi(s); err == nil && n > 0 {
				return time.Duration(n) * time.Second
			}
		}
		if d := untilReset(header.Get("X-RateLimit-Reset")); d > 0 {
			return d
		}
	}

	if v, ok := metadata["X-RateLimit-Reset"]; ok {
		return untilReset(fmt.Sprint(v))
	}

	return 0
}

func untilReset(value string) time.Duration {
	ms, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)

	if err != nil {
		return 0
	}

	return time.Until(time.UnixMilli(ms))
}

func codeOf(raw json.RawMessage, fallback int) int {
	n, err := strconv.Atoi(strings.Trim(string(raw), `" `))

	if err != nil || n == 0 {
		return fallback
	}

	return n
}

func shorten(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		return string(r[:max]) + "…"
	}
	return s
}
