package tools

import (
	"context"
	"fmt"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/memories"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/google/uuid"
)

var DeleteMemoryDefinition = llm.ToolDefinition{
	Name:        "deleteMemory",
	Description: "Remove permanentemente uma memória existente. Use quando o usuário pedir explicitamente para esquecer algo, ou quando um fato salvo se tornou claramente inválido ou contraditório. Nunca invente um id — ele precisa ser um dos que já apareceram no contexto de memórias que você recebeu.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "string",
				"description": "ID da memória a ser removida — deve ser um dos IDs já mostrados no contexto de memórias.",
			},
		},
		"required": []string{"id"},
	},
}

func NewDeleteMemoryHandler(store memories.Store) Handler {
	return func(ctx context.Context, sess *session.Session, params map[string]any) (session.ToolResult, error) {
		idStr, ok := params["id"].(string)
		if !ok || len(idStr) == 0 {
			err := fmt.Errorf("id is needed")
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		err = store.Delete(ctx, id)
		if err != nil {
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		return session.ToolResult{
			Success: true,
		}, nil
	}
}
