package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const groqThreshold = 6000

type GroqClient struct {
	Model  string
	Logger *slog.Logger
}

func NewGroqClient(model string, logger *slog.Logger) *GroqClient {
	return &GroqClient{Model: model, Logger: logger}
}

type groqChatCompletionRequest struct {
	Model           string         `json:"model"`
	Messages        []Message      `json:"messages"`
	ResponseFormat  map[string]any `json:"response_format"`
	ReasoningFormat string         `json:"reasoning_format,omitempty"`
	ReasoningEffort string         `json:"reasoning_effort,omitempty"`
	MaxTokens       int            `json:"max_tokens,omitempty"`
}

type groqMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Reasoning string `json:"reasoning,omitempty"`
}

type groqChatCompletionResponse struct {
	Choices []struct {
		Message groqMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type groqPlanStep struct {
	ID         string         `json:"id"`
	DependsOn  []string       `json:"dependsOn"`
	Capability string         `json:"capability"`
	Params     map[string]any `json:"params"`
}

type groqPlan struct {
	Steps []groqPlanStep `json:"steps"`
}

func (c *GroqClient) Complete(ctx context.Context, apiKey string, messages []Message, tools []ToolDefinition) (*CompletionResult, error) {
	requestBody, err := json.Marshal(groqChatCompletionRequest{
		Model:    c.Model,
		Messages: messages,
		ResponseFormat: map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "plan",
				"strict": true,
				"schema": BuildPlanSchema(tools),
			},
		},
		ReasoningFormat: "parsed",
		ReasoningEffort: "low",
		MaxTokens:       2048,
	})
	if err != nil {
		return nil, err
	}

	data, err := c.doWithRateLimitRetry(ctx, apiKey, requestBody)
	if err != nil {
		return nil, err
	}

	var completion groqChatCompletionResponse
	if err := json.Unmarshal(data, &completion); err != nil {
		return nil, err
	}

	if len(completion.Choices) == 0 {
		return nil, errors.New("groq api returned no choices")
	}

	msg := completion.Choices[0].Message

	if c.Logger != nil {
		c.Logger.Info("groq completion",
			"reasoning", msg.Reasoning,
			"content", msg.Content,
			"prompt_tokens", completion.Usage.PromptTokens,
			"completion_tokens", completion.Usage.CompletionTokens,
			"total_tokens", completion.Usage.TotalTokens,
		)
	}

	var plan groqPlan
	if err := json.Unmarshal([]byte(msg.Content), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan: %w", err)
	}

	toolCalls := []ToolCall{}
	for _, step := range plan.Steps {
		toolCalls = append(toolCalls, ToolCall{ID: step.ID, DependsOn: step.DependsOn, Name: step.Capability, Params: step.Params})
	}

	return &CompletionResult{Steps: toolCalls}, nil
}

func (c *GroqClient) doWithRateLimitRetry(ctx context.Context, apiKey string, requestBody []byte) ([]byte, error) {
	client := http.Client{}

	request, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	res, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusOK {
		return data, nil
	}

	if res.StatusCode == http.StatusTooManyRequests {
		retryAfter, ok := parseRetryAfter(res.Header.Get("Retry-After"))
		if c.Logger != nil {
			c.Logger.Warn("groq rate limited, failing fast so the router can fall back", "retry_after", retryAfter, "body", string(data))
		}
		return nil, &RateLimitError{RetryAfter: retryAfter, HasRetryAfter: ok}
	}

	var apiErr struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(data, &apiErr) == nil && apiErr.Error.Code == "json_validate_failed" {
		if c.Logger != nil {
			c.Logger.Warn("groq failed to generate a schema-matching response, falling back", "body", string(data))
		}
		return nil, ErrGenerationFailed
	}

	return nil, fmt.Errorf("groq api error: status %d, body %s", res.StatusCode, data)
}
