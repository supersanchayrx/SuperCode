package llm

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"supercode/internal/config"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func BuildChatRequest(systemPrompt, message string) []byte {

	var messages []Message

	if systemPrompt != "" {

		systemMessage := Message{Role: "system",
			Content: ""}

		systemMessage.Content = systemPrompt

		messages = append(messages, systemMessage)
	}

	userMessage := Message{
		Role:    "user",
		Content: message,
	}

	messages = append(messages, userMessage)

	reqBody := ChatRequest{
		Model:    config.ModelID,
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)

	if err != nil {
		log.Fatal(err)
	}

	return jsonData
}

func HandleRequest(data []byte) (string, error) {

	sentRequest := SendRequest(data)

	client := &http.Client{}

	resp, err := client.Do(sentRequest)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API Error (%d) %s", resp.StatusCode, string(body))
	}

	return ParseResponse(body)
}

func SendRequest(data []byte) *http.Request {
	req, err := http.NewRequest("POST", config.EndPoint, bytes.NewBuffer(data))

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", "Bearer "+config.SuperCodeApiKey)
	req.Header.Set("Content-Type", "application/json")

	return req
}

func ParseResponse(responseBody []byte) (string, error) {

	var finalResponse = ""
	var chatReponse ChatResponse

	err := json.Unmarshal(responseBody, &chatReponse)

	if err != nil {
		log.Fatal(err)
	}

	if len(chatReponse.Choices) > 0 {
		finalResponse = chatReponse.Choices[0].Message.Content
	}

	return finalResponse, err
}
