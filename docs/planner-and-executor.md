# Planner e Executor

Essa é a decisão arquitetural central.

**A LLM não executa ferramentas diretamente.**

```
Planner   →  "O que precisa ser feito?"   →  produz plano
Executor  →  "Como executar cada etapa?"  →  executa deterministicamente
```

## Planner

Recebe a intenção do usuário e o contexto atual. Produz um macro-workflow — uma sequência de etapas de alto nível.

O Planner não sabe se "abrir o Spotify" vai ser executado via AppleScript, PowerShell ou API. Ele pede uma capability. O sistema resolve quem pode fornecê-la.

Cada instrução processada pelo Planner carrega uma **proveniência** (`comando_direto` vs `memoria_injetada`) — ver [Memória: proveniência e segurança](memory.md#proveniência-e-segurança-contra-prompt-injection). Isso é usado pelo Executor para decidir se uma tool sensível pode rodar direto ou precisa de confirmação.

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

## Em aberto

- Critério exato de quando uma falha dispara replanning completo vs apenas retry pontual (além do número de tentativas esgotado) — depende de casos reais ainda não vividos.
