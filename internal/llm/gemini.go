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
	"strings"
)

type GeminiClient struct {
	Model  string
	Logger *slog.Logger
}

func NewGeminiClient(model string, logger *slog.Logger) *GeminiClient {
	return &GeminiClient{Model: model, Logger: logger}
}

type geminiPart struct {
	Text         string              `json:"text,omitempty"`
	FunctionCall *geminiFunctionCall `json:"functionCall,omitempty"`
}

type geminiFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiFunctionDeclaration struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFunctionDeclaration `json:"functionDeclarations"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
}

type geminiGenerateContentRequest struct {
	Contents          []geminiContent         `json:"contents"`
	SystemInstruction *geminiContent          `json:"systemInstruction,omitempty"`
	Tools             []geminiTool            `json:"tools,omitempty"`
	GenerationConfig  *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiGenerateContentResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func (c *GeminiClient) Complete(ctx context.Context, apiKey string, messages []Message, tools []ToolDefinition) (*CompletionResult, error) {
	var systemParts []string
	contents := make([]geminiContent, 0, len(messages))
	for _, m := range messages {
		if m.Role == "system" {
			systemParts = append(systemParts, m.Content)
			continue
		}

		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, geminiContent{Role: role, Parts: []geminiPart{{Text: m.Content}}})
	}

	var systemInstruction *geminiContent
	if len(systemParts) > 0 {
		systemInstruction = &geminiContent{Parts: []geminiPart{{Text: strings.Join(systemParts, "\n\n")}}}
	}

	if len(contents) > 0 && contents[len(contents)-1].Role == "model" {
		contents = append(contents, geminiContent{Role: "user", Parts: []geminiPart{{Text: "Continue."}}})
	}

	var geminiTools []geminiTool
	if len(tools) > 0 {
		declarations := make([]geminiFunctionDeclaration, 0, len(tools))
		for _, t := range tools {
			declarations = append(declarations, geminiFunctionDeclaration{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			})
		}
		geminiTools = []geminiTool{{FunctionDeclarations: declarations}}
	}

	requestBody, err := json.Marshal(geminiGenerateContentRequest{
		Contents:          contents,
		SystemInstruction: systemInstruction,
		Tools:             geminiTools,
		GenerationConfig:  &geminiGenerationConfig{MaxOutputTokens: 2048},
	})
	if err != nil {
		return nil, err
	}

	data, err := c.doWithRateLimitRetry(ctx, apiKey, requestBody)
	if err != nil {
		return nil, err
	}

	var completion geminiGenerateContentResponse
	if err := json.Unmarshal(data, &completion); err != nil {
		return nil, err
	}

	if len(completion.Candidates) == 0 {
		return nil, errors.New("gemini api returned no candidates")
	}

	parts := completion.Candidates[0].Content.Parts

	if c.Logger != nil {
		c.Logger.Info("gemini completion",
			"parts", len(parts),
			"prompt_tokens", completion.UsageMetadata.PromptTokenCount,
			"completion_tokens", completion.UsageMetadata.CandidatesTokenCount,
			"total_tokens", completion.UsageMetadata.TotalTokenCount,
		)
	}

	var toolCalls []ToolCall
	var content strings.Builder
	for _, p := range parts {
		if p.FunctionCall != nil {
			toolCalls = append(toolCalls, ToolCall{Name: p.FunctionCall.Name, Params: p.FunctionCall.Args})
			continue
		}
		content.WriteString(p.Text)
	}

	if len(toolCalls) > 0 {
		return &CompletionResult{ToolCalls: toolCalls}, nil
	}

	return &CompletionResult{Content: content.String()}, nil
}

func (c *GeminiClient) doWithRateLimitRetry(ctx context.Context, apiKey string, requestBody []byte) ([]byte, error) {
	client := http.Client{}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", c.Model)

	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-type", "application/json")
	request.Header.Set("x-goog-api-key", apiKey)

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
			c.Logger.Warn("gemini rate limited, failing fast", "retry_after", retryAfter, "body", string(data))
		}
		return nil, &RateLimitError{RetryAfter: retryAfter, HasRetryAfter: ok}
	}

	return nil, fmt.Errorf("gemini api error: status %d, body %s", res.StatusCode, data)
}
