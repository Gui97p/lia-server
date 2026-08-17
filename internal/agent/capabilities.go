package agent

import "github.com/Gui97p/lia-server/internal/llm"

var knownCapabilities = map[string]llm.ToolDefinition{
	"openApp": {
		Name:        "openApp",
		Description: "Abre um aplicativo no dispositivo do usuário",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"app": map[string]any{"type": "string", "description": "nome do aplicativo a abrir"},
			},
			"required": []string{"app"},
		},
	},
}
