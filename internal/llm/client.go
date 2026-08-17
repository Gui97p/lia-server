package llm

import "context"

type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]any
}

type ToolCall struct {
	Name   string
	Params map[string]any
}

type CompletionResult struct {
	Content  string
	ToolCall *ToolCall
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client interface {
	Complete(ctx context.Context, apiKey string, messages []Message, tools []ToolDefinition) (*CompletionResult, error)
}
