# Planner e Executor

Essa é a decisão arquitetural central.

**A LLM não executa ferramentas diretamente.**

```
Planner   →  "O que precisa ser feito?"   →  produz plano
Executor  →  "Como executar cada etapa?"  →  executa deterministicamente
```

## Agent como camada de orquestração

O loop Planner → Executor → (falha) → Replanner → novo plano (ver [Modelo híbrido de planejamento](#modelo-híbrido-de-planejamento)) é formalizado como um tipo `Agent` no código, em vez de ficar implícito em `main.go`. O `Agent` não substitui Planner/Executor — ele é quem orquestra o ciclo entre eles.

Isso abre espaço para múltiplos tipos de `Agent` no futuro (ex: um agent simples para comandos diretos, um agent de pesquisa, um agent de monitoramento contínuo), todos reaproveitando a mesma infraestrutura de Tasks, Tools, Memory e Executor — mas essa variação é para depois: por ora existe só um tipo de `Agent`. Quando existirem vários, surge a pergunta de **qual `Agent` usar para qual Task** (um roteador/dispatcher) — não é uma decisão a tomar agora, só um problema previsível quando #4 sair do papel.

## Planner

Recebe a intenção do usuário e o contexto atual. Produz um macro-workflow — uma sequência de etapas de alto nível.

O Planner não sabe se "abrir o Spotify" vai ser executado via AppleScript, PowerShell ou API. Ele pede uma capability. O sistema resolve quem pode fornecê-la.

Cada instrução processada pelo Planner carrega uma **proveniência** (`comando_direto`, `memoria_injetada`, `agendado` ou `evento` — ver [Memória: proveniência e segurança](memory.md#proveniência-e-segurança-contra-prompt-injection) e [Tasks e Eventos: autorização de Tasks sem sessão](tasks-and-events.md#autorização-de-tasks-sem-sessão-viva)). Isso é usado pelo Executor para decidir se uma tool sensível pode rodar direto ou precisa de confirmação.

### Abstração de LLM

O Planner chama uma LLM através de `internal/llm`, que expõe uma interface pequena (`Client.Complete(ctx, apiKey, messages, tools)`) em vez de acoplar diretamente ao SDK de um provider específico — mesmo padrão de interface já usado em `internal/memory` (`Store`). `GroqClient` e `GeminiClient` implementam essa interface, cada um traduzindo pro formato de request/response do respectivo provider.

### Roteamento multi-provider (`RoutingClient`)

O que o `Planner` de fato enxerga não é um `Client` único, e sim `llm.RouterClient` (`Complete(ctx, keys providers.Providers, messages, tools)`) — recebe o mapa inteiro de keys disponíveis pro usuário, não uma key isolada. `BaseRouterClient` (`internal/llm/router.go`) é a implementação: guarda um `map[providers.ProviderName]Client` e uma `Priority []providers.ProviderName` (hoje `["groq", "gemini"]`), e por chamada:

1. Pula qualquer provider sem key configurada pelo usuário (funciona com o que estiver disponível — nem todo usuário precisa ter todos os providers).
2. Pula o provider se ele estiver em **cooldown** (60s após um rate limit, guardado por *key*, não por provider — importante porque `BaseRouterClient` é compartilhado entre todas as goroutines de usuário; um cooldown por provider vazaria entre usuários diferentes).
3. Pula o Groq especificamente se uma estimativa rasa de tokens do prompt (`len(chars)/4`) passar de um teto — proteção contra o limite de TPM baixo do free tier.
4. Chama o provider. Erro real (401, 500, etc.) **não** cai pro próximo da lista — só `llm.ErrRateLimit` (sentinela, `errors.Is`) aciona o fallback. Mascarar um erro real de configuração do usuário como se fosse falta de disponibilidade seria mais confuso que útil.

Prompt caching do Groq foi avaliado e descartado por ora: só está disponível pra modelos GPT-OSS, que não suportam parallel tool calls (usado o tempo todo aqui — ver [`speak` intercalado com outras tools](tools-and-capabilities.md#speak-como-capability)). Trocar de modelo por causa do cache custaria mais do que economizaria.

## Executor

Recebe o workflow do Planner. Executa cada etapa consultando o registry de capabilities disponíveis no momento (quais clients estão conectados, quais tools estão registradas).

É determinístico — não toma decisões sobre o que fazer, apenas sobre como fazer dentro do que está disponível.

## Estrutura do Workflow

Cada step tem, no mínimo:

```go
type Step struct {
	ID         string
	Capability string
	Params     map[string]any
	DependsOn  []string // vazio = depende implicitamente do step anterior
	Provenance string   // "comando_direto" | "memoria_injetada"
	RetryPolicy *RetryPolicy // nil = herda o default da tool/global
}
```

Campos serão adicionados conforme necessidade real aparecer (princípio "casos primeiro") — não antecipar campos que nenhum caso de uso ainda pede.

### Referência a resultado de outro step (`$fromStep`)

Um parâmetro de step pode não ter valor literal disponível no momento do planejamento — só existe depois que outro step roda (ex: `moveWindow` precisa do identificador de uma janela que só `activeWindows` consegue informar, em runtime). Isso independe de o workflow ser sequencial ou DAG — mesmo um workflow linear de 2 steps já precisa disso.

O Planner expressa isso como uma referência simbólica em vez de um valor literal:

```json
{
  "id": "step-2",
  "capability": "moveWindow",
  "params": {
    "window": { "$fromStep": "step-1", "match": { "app": "discord" } },
    "monitor": 2
  },
  "dependsOn": ["step-1"]
}
```

Antes de disparar o step, o Executor resolve `$fromStep`: pega o resultado do step referenciado, aplica o filtro de `match` (igualdade exata em campo estruturado, nunca correspondência aproximada de texto) e substitui pelo valor real. Se o filtro encontrar exatamente um resultado, segue normalmente. **Zero ou múltiplos resultados** não é adivinhado pelo Executor — é tratado como falha sem retry, que escala para Replanning (ver [Tasks e Eventos](tasks-and-events.md)), podendo virar uma pergunta de desambiguação ao usuário.

### Sequencial agora, DAG depois

Por ora, workflows são listas lineares de steps (`DependsOn` vazio, execução em ordem). DAG (steps paralelos, dependências não-lineares) é a evolução natural — necessário a partir do caso de uso 3 do roadmap ("baixe o relatório, analise, compare"), quando houver necessidade real de paralelismo.

O campo `DependsOn` já existe desde o início justamente para que essa transição não exija redesenhar a estrutura do Workflow — sequencial é só o caso em que cada step depende apenas do anterior.

### Retry por step — hierarquia

Ordem de precedência (mais específico vence):

```
RetryPolicy do step (override do Planner)
      ↓ (se nil)
RetryPolicy default da tool (definida no registro da tool)
      ↓ (se nil)
RetryPolicy global do sistema
```

Ver [Tasks e Eventos](tasks-and-events.md) para a política default (3 tentativas, backoff exponencial).

## Modelo híbrido de planejamento

Nem tudo de uma vez, nem um passo por vez. O modelo preferido é híbrido:

```
           Planner
              │
              ▼
       Macro-workflow
              │
              ▼
          Executor
              │
      ┌───────┴───────┐
      │               │
  resultados       falhas
      │               │
      ▼               ▼
  continua        Replanner
                      │
                      ▼
                novo workflow
```

Planejamento antecipado quando possível. Replanejamento quando a realidade exigir.

### Como o replan é sinalizado hoje

Qualquer tool (server-side ou client-side) pode disparar um novo ciclo de planejamento devolvendo `session.ToolResult{NeedsReplan: true}` — não é um mecanismo exclusivo de uma tool específica. O `Executor` trata isso genericamente: assim que um step retorna `NeedsReplan: true`, ele para de executar o resto do Workflow daquele turno (qualquer step colocado depois dele **não roda**, silenciosamente — por isso o prompt do sistema instrui a sempre deixar uma tool desse tipo por último no plano).

Duas tools usam isso hoje:

- **`searchWeb`** dispara `NeedsReplan` sozinho, sempre — o resultado de uma busca é inútil sem o modelo reconsiderar o plano com ele em mãos.
- **`replan`** (`internal/tools/replan.go`) não faz nada além de sinalizar isso — existe pro modelo pedir reconsideração explicitamente em casos que nenhuma outra tool cobre. Ao contrário de `speak`, `replan` tem um `Handler` normal registrado no `Registry`, então ela entra automaticamente em todo plano pelo mesmo loop genérico que adiciona qualquer tool com definição+handler — não precisa de nenhum caso especial no Planner.

O resultado bruto da tool que disparou o replan **não** é persistido como mensagem de conversa — só um marcador curto (`"Executou searchWeb."`) entra no histórico permanente. O conteúdo de verdade (ex: os resultados da busca) é injetado como `extraContext` efêmero, só pro próximo `Plan` dessa mesma requisição — nunca reaparece em turnos futuros. Isso existe especificamente para não estourar o orçamento de tokens: sem isso, um resultado de busca ficaria sendo reenviado em toda chamada seguinte enquanto estivesse dentro da janela de histórico.

## Em aberto

- Critério exato de quando uma falha dispara replanning completo vs apenas retry pontual (além do número de tentativas esgotado) — depende de casos reais ainda não vividos.
