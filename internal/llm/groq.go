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
	ToolCalls []groqToolCall `json:"tool_calls,omitempty"`
}

type groqChatCompletionResponse struct {
	Choices []struct {
		Message groqMessage `json:"message"`
	} `json:"choices"`
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
		Model:             "qwen/qwen3.6-27b",
		Messages:          messages,
		Tools:             groqTools,
		ParallelToolCalls: true,
	})
	if err != nil {
		return nil, err
	}

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
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("groq api error: status %d, body %s", res.StatusCode, data)
	}

	var completion groqChatCompletionResponse
	if err := json.Unmarshal(data, &completion); err != nil {
		return nil, err
	}

	if len(completion.Choices) == 0 {
		return nil, errors.New("groq api returned no choices")
	}

	msg := completion.Choices[0].Message
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
