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
	"saveMemory": {
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
	},
	"updateMemory": {
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
	},
	"deleteMemory": {
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
	},
}
