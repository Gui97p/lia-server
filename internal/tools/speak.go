package tools

import "github.com/Gui97p/lia-server/internal/llm"

var SpeakDefinition = llm.ToolDefinition{
	Name:        "speak",
	Description: "Fala com o usuário. Use na maioria dos momentos chave.",
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"text": map[string]any{"type": "string", "description": "mensagem a ser reproduzida"},
			"mode": map[string]any{"type": "string", "description": "fire_and_forget: não espera antes do próximo passo. wait: espera a fala terminar antes do próximo passo.", "enum": []string{"fire_and_forget", "wait"}},
		},
		"required": []string{"text", "mode"},
	},
}
