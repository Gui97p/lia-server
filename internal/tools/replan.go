package tools

import (
	"context"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/session"
)

var ReplanDefinition = llm.ToolDefinition{
	Name:        "replan",
	Description: "Sinaliza que o plano deve retornar a você para replanejamento.",
	Parameters: map[string]any{
		"type":       "object",
		"properties": map[string]any{},
		"required":   []string{},
	},
}

func NewReplanHandler() Handler {
	return func(ctx context.Context, sess *session.Session, params map[string]any) (session.ToolResult, error) {
		return session.ToolResult{Success: true, NeedsReplan: true}, nil
	}
}
