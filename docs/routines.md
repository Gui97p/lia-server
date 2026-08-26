# Routines

Design discutido, nada disso existe em código ainda — registrado aqui pra não se perder até virar um caso de uso real (ver [fase de descoberta por casos](roadmap-and-responsibilities.md#fase-atual--descoberta-por-casos)).

## O problema

Pedidos como "Lia, preparar pra dormir" (desligar PC, apagar luz, ligar ventilador) são uma sequência **fixa e conhecida** de ações, não um fato sobre o mundo/usuário. Não deveria depender de [memória](memory.md) — uma memória é um fato que o modelo *reinterpreta* toda vez que é injetado no contexto; pedir pro modelo reconstruir a sequência certa de tool calls a partir de uma descrição em prosa, toda vez, é frágil (risco de esquecer um passo, trocar o nome de um device, variar a ordem) e caro em token à toa, já que é uma sequência 100% determinística que não deveria precisar ser "redescoberta" a cada execução.

Mesmo raciocínio que já levou a separar `behavior_rules` de `memories` no schema (ver [Banco de Dados](database.md#behavior_rules-é-separada-de-memories)): categorias com formato e risco diferentes não deveriam compartilhar a mesma tabela/mecanismo só porque parecem "conteúdo textual configurável" à primeira vista.

## Storage

Uma tabela própria, seguindo o mesmo padrão de catálogo já usado por capabilities/behavior_rules (Postgres, editável só via `lia-admin`, sem precisar de deploy pra registrar uma rotina nova):

```sql
routines (
  id, user_id,           -- null user_id = rotina global, mesmo padrão de escopo de memories
  name, description,
  steps jsonb,            -- []Step — mesmo formato que o Workflow já usa
  created_at, updated_at
)
```

`steps` é literalmente uma lista de `Step{Capability, Params, TargetDevice}` (ver [estrutura do Workflow](planner-and-executor.md#estrutura-do-workflow) e [roteamento de device](transport-and-audio.md#client-vs-device)) — nada de formato novo pra aprender.

## Como o modelo vê isso

Uma única tool dinâmica, `runRoutine(name: enum[...])`, com o enum populado pelos nomes de rotina cadastrados (+ descrição curta de cada uma no schema). Preferido a uma tool por rotina (`runRoutine_dormir`, etc.) — mantém o catálogo de tools enxuto conforme o número de rotinas cresce, mesmo instinto que já guiou o corte de tokens dos schemas existentes (ver [Tools e Capabilities](tools-and-capabilities.md)).

## Execução: determinística, sem replan no meio

O handler de `runRoutine` **não** deveria disparar um replan pedindo pro modelo reconstruir os passos a partir de uma descrição — isso reintroduziria exatamente o problema que rotina existe pra evitar. Em vez disso, `runRoutine` é tratado como caso especial no `Executor`, no mesmo espírito de como `speak` já é hoje (não passa pelo `tools.Registry` genérico, porque precisa de controle de fluxo privilegiado — ver [`speak` como capability](tools-and-capabilities.md#speak-como-capability)):

1. Executor encontra um step com capability `runRoutine`.
2. Busca os `steps` da rotina no Postgres (por `name`).
3. Executa esses steps sequencialmente, dentro do mesmo `Execute()` — sem nenhuma chamada de LLM extra no meio.

O modelo só decide **qual** rotina rodar; nunca reconstrói **o que** ela faz.

## Falha no meio de uma rotina

Mesmo tratamento que qualquer falha de step do Workflow já tem hoje (ver [Falhas](tasks-and-events.md#falhas)) — não é um mecanismo novo, os steps expandidos da rotina são, na prática, só mais steps do Workflow em execução.

## Ações irreversíveis não pedem confirmação de novo

O prompt hoje exige confirmação explícita do usuário antes de uma ação sensível/irreversível. Dentro de uma rotina, isso não deveria disparar de novo a cada execução — a aprovação já aconteceu no momento em que alguém (Gui/Yure) revisou e cadastrou os `steps` via `lia-admin`. Mesmo raciocínio que já justifica `behavior_rules` ser editável só via CLI: autoria humana deliberada substitui confirmação ad-hoc.

## Composição com device routing

Cada step de uma rotina pode ter seu próprio `target_device`, então "dormir" vira literalmente:

```json
[
  { "capability": "toggleLight", "target_device": "luz-quarto", "params": { "state": "off" } },
  { "capability": "shutdown", "target_device": "pc" },
  { "capability": "toggleFan", "target_device": "ventilador-quarto", "params": { "state": "on" } }
]
```

Nenhum mecanismo novo — [roteamento pra device nomeado](transport-and-audio.md#client-vs-device) já cobre isso.

## Composição com Tasks agendadas/por evento

Ponto que ainda não estava conectado: uma Task com `trigger_type: scheduled` ou `trigger_type: event` (ver [Autorização de Tasks sem sessão viva](tasks-and-events.md#autorização-de-tasks-sem-sessão-viva)) que existe só pra rodar uma rotina específica ("toda noite às 22h, rotina dormir") **não precisa passar pelo Planner/LLM nenhuma vez** — já se sabe deterministicamente qual rotina rodar, o agendamento só dispara o `Executor` direto nos `steps` daquela rotina. Isso é uma economia real (zero chamadas de LLM pra esse fluxo inteiro), e uma questão de segurança a menos (nada de replanning especulativo rodando sem sessão viva por trás).

## Em aberto

- **Rotina com parâmetro** — hoje o desenho assume rotina sem argumento ("dormir" sempre faz a mesma coisa). Um caso tipo "modo cinema pro quarto" vs "modo cinema pra sala" — rotina parametrizável (`target_device` base injetado, ou um `$param` nos steps) ainda não tem formato definido.
- **Colisão de nome entre rotina e capability real** — `runRoutine(name: "openApp")` não devia ser possível; validação de nome ainda não desenhada (provavelmente: `lia-admin routines add` rejeita nome já usado por uma capability).
- **Rotina que chama outra rotina** — permitir ou proibir composição (rotina A tem um step que roda rotina B)? Risco de ciclo. Não decidido — provável que a resposta inicial seja "proibir", e revisitar só se um caso real pedir.
