package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"supercode/internal/config"
)

func BuildRequest(messages []Message, tools []ToolDef) ([]byte, error) {

	requestBody := ChatRequest{
		Model:    config.ModelID,
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

func BuildChatRequest(systemPrompt, message string) ([]byte, error) {

	var messages []Message

	if systemPrompt != "" {

		systemMessage := Message{Role: "system",
			Content: systemPrompt}

		messages = append(messages, systemMessage)
	}

	userMessage := Message{
		Role:    "user",
		Content: message,
	}

	messages = append(messages, userMessage)

	return BuildRequest(messages, nil)
}

func HandleRequest(data []byte) (Message, error) {

	var parsedMessage Message
	sentRequest, err := SendRequest(data)

	if err != nil {
		return parsedMessage, err
	}

	client := &http.Client{}

	resp, err := client.Do(sentRequest)

	if err != nil {
		return parsedMessage, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return parsedMessage, fmt.Errorf("couldn't read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return parsedMessage, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	return ParseResponse(body)
}

func SendRequest(data []byte) (*http.Request, error) {
	req, err := http.NewRequest("POST", config.EndPoint, bytes.NewBuffer(data))

	if err != nil {
		return nil, fmt.Errorf("couldn't create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.SuperCodeApiKey)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func ParseResponse(responseBody []byte) (Message, error) {

	var finalResponse Message
	var chatResponse ChatResponse

	err := json.Unmarshal(responseBody, &chatResponse)

	if err != nil {
		return finalResponse, fmt.Errorf("couldn't parse response: %w", err)
	}

	if len(chatResponse.Choices) == 0 {
		return finalResponse, fmt.Errorf("no choices in response")
	}

	finalResponse = chatResponse.Choices[0].Message
	return finalResponse, nil
}
