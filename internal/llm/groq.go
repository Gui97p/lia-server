package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type GroqClient struct{}

type groqChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type groqChatCompletionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func (c *GroqClient) Complete(ctx context.Context, apiKey string, messages []Message) (string, error) {
	requestBody, err := json.Marshal(groqChatCompletionRequest{
		Model:    "openai/gpt-oss-120b",
		Messages: messages,
	})
	if err != nil {
		return "", err
	}

	client := http.Client{}

	request, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	res, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq api error: status %d, body %s", res.StatusCode, data)
	}

	var completion groqChatCompletionResponse
	if err := json.Unmarshal(data, &completion); err != nil {
		return "", err
	}

	if len(completion.Choices) == 0 {
		return "", errors.New("groq api returned no choices")
	}

	return completion.Choices[0].Message.Content, nil
}
