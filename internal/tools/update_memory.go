package tools

import (
	"context"
	"fmt"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/memories"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/google/uuid"
)

var UpdateMemoryDefinition = llm.ToolDefinition{
	Name:        "updateMemory",
	Description: "Atualiza o conteúdo de uma memória já existente. Use apenas quando um fato salvo anteriormente mudou ou ficou desatualizado — nunca invente um id, ele precisa ser um dos que já apareceram no contexto de memórias que você recebeu.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "string",
				"description": "ID da memória a ser atualizada — deve ser um dos IDs já mostrados no contexto de memórias.",
			},
			"fact": map[string]any{
				"type":        "string",
				"description": "Novo conteúdo do fato, substituindo o anterior por completo.",
			},
		},
		"required": []string{"id", "fact"},
	},
}

func NewUpdateMemoryHandler(store memories.Store) Handler {
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

		fact, ok := params["fact"].(string)
		if !ok || len(fact) == 0 {
			err := fmt.Errorf("fact is needed")
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		err = store.SetFact(ctx, id, fact)
		if err != nil {
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		return session.ToolResult{
			Success: true,
		}, nil
	}
}
