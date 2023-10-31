package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAiService struct {
	ApiKey string
}

type ChatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message ChatMessage `json:"message"`
}

const chatCompletionUrl = "https://api.openai.com/v1/chat/completions"

var client = &http.Client{}

// GenerateTypingText sends a request to the OpenAI API and returns the generated text
func (oais *OpenAiService) GenerateTypingText(input CreateTextInput) (string, error) {
	systemPrompt := "You return a text that will be used for practicing 10 finger typing. " +
		"You will receive several input variables. The length of the text, if the text should include punctuation or" +
		" not (if it includes punctuation then include punctuation and capital letters at the beginning of a new sentence, " +
		"if not then all the letters should be lowercase), number of special characters and number of numbers that should be " +
		"used in the text. Only return the text itself, not any explanation. The text doesn't have to make sense."
	requestBody := oais.createRequestBody(input, systemPrompt)

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		marshallingError := fmt.Errorf("error marshalling struct to JSON: %w", err)
		return "", marshallingError
	}

	req, err := oais.createRequest(jsonBody)
	if err != nil {
		return "", fmt.Errorf("error creating request to OpenAI: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request to OpenAI: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var response ChatCompletionResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response body: %w", err)
	}

	return response.Choices[0].Message.Content, nil
}

func (oais *OpenAiService) createRequestBody(input CreateTextInput, systemPrompt string) ChatCompletionRequest {
	return ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ChatMessage{
			{Content: systemPrompt, Role: "system"}, {Content: input.String(), Role: "user"},
		},
	}
}

func (oais *OpenAiService) createRequest(jsonBody []byte) (*http.Request, error) {
	req, err := http.NewRequest("POST", chatCompletionUrl, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+oais.ApiKey)

	return req, nil
}
