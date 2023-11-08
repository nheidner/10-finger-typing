package open_ai_repo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAiRepository struct {
	ApiKey string
}

func NewOpenAiRepository(apiKey string) *OpenAiRepository {
	return &OpenAiRepository{apiKey}
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

// GenerateTypingText sends a request to the OpenAI API and returns the generated text
func (oar *OpenAiRepository) GenerateTypingText(language string, punctuation bool, specialCharacters, numbers int) (string, error) {
	systemPrompt := "You return a text that will be used for practicing 10 finger typing. " +
		"You will receive several input variables. The length of the text, if the text should include punctuation or" +
		" not (if it includes punctuation then include punctuation and capital letters at the beginning of a new sentence, " +
		"if not then all the letters should be lowercase), number of special characters and number of numbers that should be " +
		"used in the text. Only return the text itself, not any explanation. The text doesn't have to make sense."
	requestBody := oar.createRequestBody(language, systemPrompt, punctuation, specialCharacters, numbers)

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		marshallingError := fmt.Errorf("error marshalling struct to JSON: %w", err)
		return "", marshallingError
	}

	req, err := oar.createRequest(jsonBody)
	if err != nil {
		return "", fmt.Errorf("error creating request to OpenAI: %w", err)
	}

	var client = &http.Client{}

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

func (oar *OpenAiRepository) createRequestBody(language, systemPrompt string, punctuation bool, specialCharacters, numbers int) ChatCompletionRequest {
	content := fmt.Sprintf("language: %s, punctuation: %t, number of special characters: %d, number of numbers: %d, length: 100 words", language, punctuation, specialCharacters, numbers)
	return ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ChatMessage{
			{Content: systemPrompt, Role: "system"}, {Content: content, Role: "user"},
		},
	}
}

func (oar *OpenAiRepository) createRequest(jsonBody []byte) (*http.Request, error) {
	req, err := http.NewRequest("POST", chatCompletionUrl, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+oar.ApiKey)

	return req, nil
}
