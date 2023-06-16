package openai

import (
	"10-typing/models"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

var (
	client            = &http.Client{}
	chatCompletionUrl = "https://api.openai.com/v1/chat/completions"
	systemPrompt      = "You return a text that will be used for practicing 10 finger typing. " +
		"You will receive several input variables. The length of the text, if the text should include punctuation or" +
		" not (if it includes punctuation then include punctuation and capital letters at the beginning of a new sentence, " +
		"if not then all the letters should be lowercase), number of special characters and number of numbers that should be " +
		"used in the text. Only return the text itself, not any explanation. The text doesn't have to make sense."
)

// GPTText sends a request to the OpenAI API and returns the generated text
func GPTText(input *models.CreateTextInput) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")

	requestBody := ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Content: systemPrompt, Role: "system"}, {Content: input.String(), Role: "user"},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		marshallingError := errors.New("error marshalling struct to JSON")
		return "", marshallingError
	}

	req, err := http.NewRequest("POST", chatCompletionUrl, bytes.NewReader(jsonBody))
	if err != nil {
		return "", errors.New("error creating request to OpenAI")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("error sending request to OpenAI")
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("error reading response body")
	}

	var response ChatCompletionResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return "", errors.New("error unmarshalling response body")
	}

	return response.Choices[0].Message.Content, nil
}
