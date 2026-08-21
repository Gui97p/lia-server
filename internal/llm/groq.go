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
	"time"
)

const groqThreshold = 6000

type GroqClient struct {
	Model  string
	Logger *slog.Logger
}

func NewGroqClient(model string, logger *slog.Logger) *GroqClient {
	return &GroqClient{Model: model, Logger: logger}
}

type groqFunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type groqTool struct {
	Type     string          `json:"type"`
	Function groqFunctionDef `json:"function"`
}

type groqChatCompletionRequest struct {
	Model             string     `json:"model"`
	Messages          []Message  `json:"messages"`
	Tools             []groqTool `json:"tools,omitempty"`
	ParallelToolCalls bool       `json:"parallel_tool_calls"`
	ReasoningFormat   string     `json:"reasoning_format,omitempty"`
	ReasoningEffort   string     `json:"reasoning_effort,omitempty"`
	MaxTokens         int        `json:"max_tokens,omitempty"`
}

type groqToolCall struct {
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type groqMessage struct {
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	Reasoning string         `json:"reasoning,omitempty"`
	ToolCalls []groqToolCall `json:"tool_calls,omitempty"`
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

func (c *GroqClient) Complete(ctx context.Context, apiKey string, messages []Message, tools []ToolDefinition) (*CompletionResult, error) {
	groqTools := make([]groqTool, 0, len(tools))
	for _, t := range tools {
		groqTools = append(groqTools, groqTool{
			Type: "function",
			Function: groqFunctionDef{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	requestBody, err := json.Marshal(groqChatCompletionRequest{
		Model:             c.Model,
		Messages:          messages,
		Tools:             groqTools,
		ParallelToolCalls: true,
		ReasoningFormat:   "parsed",
		ReasoningEffort:   "none",
		MaxTokens:         2048,
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
			"tool_calls", len(msg.ToolCalls),
			"prompt_tokens", completion.Usage.PromptTokens,
			"completion_tokens", completion.Usage.CompletionTokens,
			"total_tokens", completion.Usage.TotalTokens,
		)
	}
	if len(msg.ToolCalls) > 0 {
		toolCalls := make([]ToolCall, 0, len(msg.ToolCalls))
		for _, call := range msg.ToolCalls {
			var params map[string]any
			if err := json.Unmarshal([]byte(call.Function.Arguments), &params); err != nil {
				return nil, fmt.Errorf("invalid tool call arguments: %w", err)
			}

			toolCalls = append(toolCalls, ToolCall{Name: call.Function.Name, Params: params})
		}

		return &CompletionResult{
			ToolCalls: toolCalls,
		}, nil
	}

	return &CompletionResult{Content: msg.Content}, nil
}

func (c *GroqClient) doWithRateLimitRetry(ctx context.Context, apiKey string, requestBody []byte) ([]byte, error) {
	client := http.Client{}

	for attempt := 0; ; attempt++ {
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

		if res.StatusCode == http.StatusTooManyRequests && attempt < maxRateLimitRetries {
			wait := retryAfterDuration(res.Header.Get("Retry-After"))
			if c.Logger != nil {
				c.Logger.Warn("groq rate limited, retrying", "attempt", attempt+1, "wait", wait, "body", string(data))
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
			continue
		}

		if res.StatusCode == http.StatusTooManyRequests {
			return nil, ErrRateLimit
		}

		return nil, fmt.Errorf("groq api error: status %d, body %s", res.StatusCode, data)
	}
}
