package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/Gui97p/lia-server/internal/websearch"
)

var SearchWebDefinition = llm.ToolDefinition{
	Name:        "searchWeb",
	Description: "Busca informações atualizadas na internet. Use quando o pedido depender de informação que você não sabe de cor ou que pode ter mudado (notícias, preços, eventos recentes, fatos específicos) — não use pra conversa geral ou fatos estáveis que você já sabe.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Termos de busca, como você digitaria num motor de busca.",
			},
		},
		"required": []string{"query"},
	},
}

func NewSearchWebHandler(client websearch.Client) Handler {
	return func(ctx context.Context, sess *session.Session, params map[string]any) (session.ToolResult, error) {
		query, ok := params["query"].(string)
		if !ok || len(query) == 0 {
			err := fmt.Errorf("query is needed")
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		results, err := client.Search(ctx, query)
		if err != nil {
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		resultJSON, err := json.Marshal(results)
		if err != nil {
			return session.ToolResult{Success: false, Error: err.Error()}, err
		}

		return session.ToolResult{Success: true, Result: resultJSON, NeedsReplan: true}, nil
	}
}
