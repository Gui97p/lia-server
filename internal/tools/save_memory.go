package tools

import (
	"context"
	"fmt"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/memories"
	"github.com/Gui97p/lia-server/internal/session"
)

var SaveMemoryDefinition = llm.ToolDefinition{
	Name:        "saveMemory",
	Description: "Salva um novo fato permanente na memória, pra ser lembrado em conversas futuras. Use quando o usuário compartilhar uma informação, preferência ou fato relevante sobre si mesmo — não use para pedidos ou informações efêmeras que só importam nesta conversa (ex: 'abra o Spotify agora' não é um fato a guardar).",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"fact": map[string]any{
				"type":        "string",
				"description": "O fato em si, escrito de forma clara e objetiva, em texto corrido (ex: 'gosta de rock progressivo', 'trabalha como programador').",
			},
			"scope": map[string]any{
				"type":        "string",
				"description": "Escopo do fato. 'user': fato específico sobre o usuário atual (preferências, dados pessoais, rotina) — use esse na grande maioria dos casos. 'global': conhecimento geral sobre o mundo, que não pertence a nenhum usuário específico (raro — só use se for um fato universal, não relacionado a uma pessoa). 'private': um fato que você deve saber mas nunca deve revelar diretamente ao usuário se perguntada (raro, use com cautela e só quando fizer sentido de verdade).",
				"enum":        []string{"user", "global", "private"},
			},
			"category": map[string]any{
				"type":        "string",
				"description": "Categoria curta e livre pra agrupar o fato (ex: 'preferencias', 'trabalho', 'saude'). Opcional — pode omitir se não houver categoria clara.",
			},
		},
		"required": []string{"fact", "scope"},
	},
}

func NewSaveMemoryHandler(store memories.Store) Handler {
	return func(ctx context.Context, sess *session.Session, params map[string]any) (session.ToolResult, error) {
		scopeStr, ok := params["scope"].(string)
		if !ok || len(scopeStr) == 0 {
			err := fmt.Errorf("scope is needed")
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}
		scope := memories.MemoryScope(scopeStr)

		fact, ok := params["fact"].(string)
		if !ok || len(fact) == 0 {
			err := fmt.Errorf("fact is needed")
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		var memory *memories.Memory
		var err error

		switch scope {
		case memories.Global:
			memory, err = store.Create(ctx, scope, fact, nil)
		case memories.Private:
			memory, err = store.Create(ctx, scope, fact, nil)
		case memories.User:
			memory, err = store.Create(ctx, scope, fact, &sess.UserID)
		default:
			err = fmt.Errorf("invalid scope")
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}
		if err != nil {
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		category, ok := params["category"].(string)
		if ok {
			if len(category) > 0 {
				err = store.SetCategory(ctx, memory.ID, category)
				if err != nil {
					return session.ToolResult{Success: false, Error: err.Error()}, err
				}
			}
		}

		return session.ToolResult{
			Success: true,
		}, nil
	}
}
