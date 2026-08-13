# Tasks e Eventos

## Máquina de Estados de Task

Uma Task não é simplesmente `success/failed`. Ela tem ciclo de vida próprio:

```
CREATED
   │
   ▼
PLANNING
   │
   ▼
READY
   │
   ▼
RUNNING ──────────────────────┐
   │                          │
   ├── WAITING                │
   │     (aguarda resultado   │
   │      de tool ou usuário) │
   │                          │
   ├── BLOCKED                │
   │     (precisa de          │
   │      intervenção)        │
   │                          │
   ├── REPLANNING ────────────┘
   │
   ├── COMPLETED
   │
   ├── FAILED
   │
   └── CANCELLED
```

A máquina de estados definitiva ainda será refinada com casos de uso reais.

## Arquitetura de Eventos

A comunicação entre componentes é event-driven. Em vez de chamadas diretas (`Planner → Executor`), os componentes se comunicam via eventos:

```
Planner
   │
   ▼
 event
   │
   ▼
Executor
```

Isso permite: execução distribuída, logs automáticos de tudo que acontece, cancelamento em qualquer ponto, retries sem acoplamento, recuperação após desconexão, observabilidade completa.

### Implementação do Event Bus: in-process

Para um servidor único (sem múltiplas instâncias ainda), o Event Bus é in-process — `chan`/pub-sub interno do Go, sem infraestrutura externa (NATS, Redis Streams, etc.). Não introduzir essas dependências antes de haver necessidade real de múltiplas instâncias do servidor — seria construir para uma escala que não existe ainda.

## Retry policy

Default global: **3 tentativas, backoff exponencial**. Pode ser sobrescrito por tool ou por step (ver [hierarquia em Planner e Executor](planner-and-executor.md#retry-por-step--hierarquia)).

### Teto de chamadas de LLM por Task

O servidor usa a `groq_api_key` do próprio usuário para chamar o Groq em nome dele (ver [Identidade e Secrets](identity-auth-and-secrets.md)). Um bug que faça o Planner entrar em loop de replanning consome a cota/créditos do usuário sem limite, além de ser um risco funcional (um Planner "preso" tentando repetidamente uma ação real, como trancar porta). Por isso, além da retry policy por step, deve existir um **limite duro de chamadas de LLM por Task** (independente de quantos steps ou retries individuais aconteçam) — uma rede de segurança contra loops, não apenas uma questão de custo.

## Cancelamento distribuído

Quando o usuário cancela uma Task, isso precisa se propagar até uma tool já em execução no client:

```
tool.cancel
     │
     ▼
tool.cancelled (se possível)
```

## Falhas

Falha não significa necessariamente `Task → FAILED`. O sistema trata falhas como parte normal do fluxo:

```
Tool falhou
     │
     ▼
Executor analisa
     │
     ├── retry disponível → tenta novamente
     │         │
     │         ├── sucesso → continua workflow
     │         │
     │         └── falha → Planner
     │                        │
     │                        ▼
     │                   replaneja
     │
     └── sem retry → BLOCKED
                        │
                        ▼
               usuário precisa intervir
```

## Autorização de Tasks sem sessão viva

Todo o modelo de identidade (ver [Identidade, Autenticação e Secrets](identity-auth-and-secrets.md)) pressupõe uma sessão viva — um JWT, um device conectado. Isso não existe para uma Task disparada por agendamento ("toda sexta faça backup") ou por evento externo ("porta aberta há 5 minutos"). Sem sessão, não há `trust_level` nem `device_id` pra consultar.

Toda Task carrega um `trigger_type` e um snapshot congelado de autorização, capturado no momento em que a Task foi criada (não uma referência a uma sessão que pode nem existir mais quando a Task rodar):

```sql
tasks (
  ...,
  trigger_type,             -- 'user' | 'scheduled' | 'event'
  authorized_trust_level,   -- trust level do usuário no momento em que a Task foi configurada/criada
  ...
)
```

- `trigger_type: user` — origem é uma sessão viva, comportamento já descrito no resto da documentação.
- `trigger_type: scheduled` — proveniência `agendado` (ver [Memória: proveniência](memory.md#proveniência-e-segurança-contra-prompt-injection)). O usuário autorizou isso com antecedência; a Task pode executar ações sensíveis sem confirmação ao vivo, já que esse é o próprio propósito de agendar algo pra rodar sem ninguém por perto. Antes de executar, o Executor revalida que a autorização daquele usuário ainda é válida (equivalente ao `token_version` — se a conta foi revogada entre o agendamento e a execução, a Task não roda mesmo estando pronta).
- `trigger_type: event` — proveniência `evento`. A condição de disparo vem de um sinal externo (sensor, estado do ambiente), não de um comando ao vivo nem de uma janela de tempo que o usuário está presente. Se o efeito é só notificação, sem confirmação. Se o efeito é uma ação física automática a partir só do evento, aplica-se a mesma cautela de `memoria_injetada` — confirmação explícita antes de agir.

## Recuperação após reboot

Reiniciar o servidor com uma Task em `RUNNING` é um caso trivial de acontecer — a Task deve ir para `FAILED` ou `BLOCKED`, nunca ser retomada automaticamente. Retomar uma ação que pode ter efeito no mundo real (trancar porta, etc.) sem confirmação é arriscado demais, especialmente considerando o cenário abaixo.

### Risco de race: task completada mas evento não recebido

Se uma tool no client terminou e emitiu `tool.completed`, mas o servidor caiu antes de processar esse evento, o servidor não pode saber se a ação foi de fato concluída ou não. Retomar cegamente pode duplicar uma ação (ex: trancar a porta de novo quando já estava trancada — inofensivo aqui, mas o padrão geral é perigoso para tools não-idempotentes).

## Em aberto

- **Protocolo de reconciliação pós-reconexão** — quando um client reconecta após o servidor reiniciar, ele deveria reportar o último status conhecido de qualquer tool/task que estivesse em execução, para o servidor reconciliar em vez de assumir cegamente `FAILED`/`BLOCKED`. Isso ainda não tem formato definido — depende também do protocolo de anúncio de capabilities (ver [Tools e Capabilities](tools-and-capabilities.md)).
- **Timeout de cancelamento** — o que acontece se o client nunca confirmar `tool.cancelled`? Ainda não definido.
