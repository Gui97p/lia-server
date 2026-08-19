package agent

import "github.com/Gui97p/lia-server/internal/llm"

var knownCapabilities = map[string]llm.ToolDefinition{
	"speak": {
		Name:        "speak",
		Description: "Conversa com o usuário. Essa tool deve ser usada na maioria dos momentos chave.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"text": map[string]any{"type": "string", "description": "mensagem a ser reproduzida"},
				"mode": map[string]any{"type": "string", "description": "modo que define o comportamento do server quanto à fala. fire_and_forget: continua pro próximo step sem esperar. wait: espera a fala terminar antes do próximo step. wait_and_replan: espera terminar e reconsidera o plano — use quando o que falar depois depender de um resultado ainda não conhecido, ou quando a fala for uma pergunta ao usuário.", "enum": []string{"fire_and_forget", "wait", "wait_and_replan"}},
			},
			"required": []string{"text", "mode"},
		},
	},
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
