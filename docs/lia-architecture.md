# Lia — Arquitetura do Sistema

> Documento vivo. Reflete decisões tomadas e hipóteses em aberto. Evolui junto com o projeto.

---

## Visão

A Lia não é um chatbot. É uma plataforma de inteligência pessoal — uma única IA que acompanha o usuário em todos os dispositivos, mantém memória contínua, reconhece quem está falando, e é capaz de planejar e executar tarefas complexas com múltiplas etapas, falhas e recuperações.

A referência é o Jarvis: uma inteligência que age, não apenas responde.

```
                      LIA
                       │
          ┌────────────┼────────────┐
          │            │            │
       Usuário A    Usuário B    Usuário C
          │
    ┌─────┴─────┐
  Desktop     Mobile
```

A identidade da Lia é única. Usuários, sessões, permissões e memórias possuem seus próprios contextos.

---

## Princípios

- **Fluidez** — a experiência deve ser natural. A Lia não esquece, não recomeça, não perde contexto ao trocar de device.
- **Casos primeiro** — a arquitetura é descoberta através de casos de uso reais, do simples ao complexo. Não desenhamos o que não precisamos ainda.
- **Falhas são normais** — o sistema é projetado para lidar com falhas como parte do fluxo, não como exceções arquiteturais.
- **Separação clara** — servidor é responsável pelo significado e inteligência. Cliente é responsável pela experiência de interação em tempo real.

---

## Stack de Linguagens

A linguagem é escolhida pelo que faz mais sentido em cada contexto:

| Componente | Linguagem | Motivo |
|---|---|---|
| Server (core) | Go | Concorrência nativa, performance, binário único |
| Clients (desktop, CLI, Discord) | Rust | Performance, baixo consumo, nativo |
| IoT / hardware | C++ | Acesso direto ao hardware |
| Treinamento de modelos (wake word) | Python | Ecossistema de ML sem alternativa |

---

## Identidade, Autenticação e Autorização

Três conceitos distintos que não devem ser confundidos:

```
Identificação      →  "Provavelmente é Guilherme"
Autenticação       →  "Essa sessão pertence a Guilherme"
Autorização        →  "Guilherme pode fazer isso?"
```

### Identificação por voz

A Lia reconhece quem está falando através do perfil de voz. Isso é identificação — não autenticação. Mais de uma pessoa pode usar o mesmo dispositivo.

```
Áudio recebido
      │
      ▼
Reconhecimento de voz
      │
      ├── Reconheceu → usa perfil do usuário identificado
      │
      └── Não reconheceu → usuário anônimo (permissões mínimas)
```

O reconhecimento de voz contribui para um **nível de confiança**, não para um booleano `authenticated = true/false`.

### Níveis de confiança

Em vez de autenticado/não autenticado, o sistema opera com níveis:

```
ANONYMOUS     → voz não reconhecida ou sem voz
IDENTIFIED    → voz reconhecida (provavelmente é X)
AUTHENTICATED → sessão autenticada por credencial forte
TRUSTED       → autenticado + contexto conhecido + histórico
```

Permissões e capabilities disponíveis variam por nível.

### Modelo de dados

```sql
users        (id, username, password_hash, groq_api_key, created_at)
groups       (id, name, created_at)
user_groups  (user_id, group_id)
voice_profiles (id, user_id, embedding, created_at)
```

### Gestão de usuários

Operações via CLI Go local — sem rota pública de registro:

```bash
./lia-admin users create --username yure --key gsk_...
./lia-admin users delete --username yure
```

### JWT

Login via HTTP retorna JWT de longa duração (1 ano). Sem refresh tokens.

```json
{
  "user_id": "uuid",
  "username": "gui",
  "groq_api_key": "gsk_...",
  "group_ids": ["amigos"],
  "trust_level": "authenticated"
}
```

---

## Memória

### Escopos

Memória não é apenas "fatos sobre o usuário". Existem quatro escopos:

```
GLOBAL   → conhecimento da Lia sobre o mundo (não pertence a ninguém)
USER     → fatos sobre um usuário específico
GROUP    → fatos compartilhados entre membros de um grupo
PRIVATE  → fatos que a Lia possui mas nenhum usuário pode acessar diretamente
```

> **Memória não é autorização.** A Lia pode saber algo sem nenhum usuário poder consultar esse dado.

### Injeção no contexto

No contexto de cada conversa, a Lia recebe:
- Memórias do usuário atual (USER)
- Memórias dos grupos do usuário (GROUP)
- Memórias globais relevantes (GLOBAL)

Memórias PRIVATE nunca são expostas diretamente.

### Tools de memória

```
saveMemory(fact, category, scope, target_id?)
updateMemory(id, fact)
deleteMemory(id)
```

IDs são expostos no system prompt para que a Lia saiba o que pode gerenciar.

### Evolução futura

Adição de `pgvector` para busca semântica — recuperar memórias por relevância em vez de injetar tudo.

---

## Planner e Executor

Essa é a decisão arquitetural central.

**A LLM não executa ferramentas diretamente.**

```
Planner   →  "O que precisa ser feito?"   →  produz plano
Executor  →  "Como executar cada etapa?"  →  executa deterministicamente
```

### Planner

Recebe a intenção do usuário e o contexto atual. Produz um macro-workflow — uma sequência de etapas de alto nível.

O Planner não sabe se "abrir o Spotify" vai ser executado via AppleScript, PowerShell ou API. Ele pede uma capability. O sistema resolve quem pode fornecê-la.

### Executor

Recebe o workflow do Planner. Executa cada etapa consultando o registry de capabilities disponíveis no momento (quais clients estão conectados, quais tools estão registradas).

É determinístico — não toma decisões sobre o que fazer, apenas sobre como fazer dentro do que está disponível.

### Modelo híbrido de planejamento

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

---

## Tasks e Máquina de Estados

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

---

## Ciclo de Vida de Tools

Tools têm ciclo de vida próprio — cancelar uma Task não cancela automaticamente uma tool já em execução no client.

```
tool.request
     │
     ▼
tool.accepted
     │
     ▼
tool.started
     │
     ▼
tool.progress (opcional)
     │
     ▼
tool.completed
```

E para cancelamento:

```
tool.cancel
     │
     ▼
tool.cancelled (se possível)
```

### Separação de capabilities

```
Capability    →  o que precisa ser feito ("abrir aplicativo")
Tool          →  abstração que representa essa capability
Implementation →  como é feito em cada plataforma (PowerShell, AppleScript, etc.)
```

### Registry de capabilities

O Executor consulta um registry que mapeia capabilities para implementations disponíveis no momento. Clients anunciam suas capabilities ao conectar.

```json
{
  "capabilities": {
    "openApp": ["discord", "vscode", "spotify"],
    "activeWindows": ["discord", "vscode"],
    "default": ["exit", "setClock"]
  }
}
```

---

## Arquitetura de Eventos

A comunicação entre componentes é event-driven. Em vez de chamadas diretas:

```
Planner → Executor
```

Os componentes se comunicam via eventos:

```
Planner
   │
   ▼
 event
   │
   ▼
Executor
```

Isso permite:
- Execução distribuída
- Logs automáticos de tudo que acontece
- Cancelamento em qualquer ponto
- Retries sem acoplamento
- Recuperação após desconexão
- Observabilidade completa

---

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

---

## Cliente vs Servidor

### Cliente

Responsável pela experiência de interação em tempo real:

- Microfone e speaker
- VAD (Voice Activity Detection)
- Wake word (inferência local do modelo .onnx)
- Captura e reprodução de áudio
- Buffering e streaming
- Estado da conexão
- Interrupção de fala
- Latência percebida
- Tools locais (openApp, windowManager, etc.)

### Servidor

Responsável pelo significado e inteligência:

- STT (Whisper via Groq)
- TTS (Edge TTS)
- LLM (Groq)
- Planner
- Executor
- Memory
- Tool registry
- Orquestração de workflows
- Autenticação e autorização

STT/TTS no servidor não significa que o cliente precisa esperar respostas completas. A comunicação pode ser streaming quando necessário.

---

## Áudio

### Fluxo completo

```
Microfone → VAD → Wake Word (local) → captura áudio
→ POST /audio/transcribe (STT no server)
→ socket: mensagem de texto
→ Planner/Executor
→ POST /audio/speak (TTS no server)
→ client recebe bytes e reproduz
```

### Wake Word

Modelo `hey_lia.onnx` treinado via OpenWakeWord. Inferência local no client — zero latência de rede para detecção.

### Sem streaming de áudio por agora

TTS recebe texto completo e gera áudio de uma vez. Streaming de texto conflita com entonação natural do TTS — prioridade para qualidade de voz. Revisitar quando necessário.

---

## Transporte

Protocolo ainda em aberto (Socket.IO vs alternativas). O que está decidido é que a comunicação é bidirecional e persistente — não HTTP puro.

Rotas HTTP permanecem para operações pontuais:
- `POST /auth/login`
- `POST /audio/transcribe`
- `POST /audio/speak`

---

## Banco de Dados

PostgreSQL como banco principal. Suporta múltiplas conexões simultâneas e prepara para `pgvector` no futuro.

```sql
users          (id, username, password_hash, groq_api_key, created_at)
groups         (id, name, created_at)
user_groups    (user_id, group_id)
voice_profiles (id, user_id, embedding, created_at)
messages       (id, user_id, role, content, created_at)
memories       (id, user_id, group_id, scope, fact, category, created_at)
tasks          (id, user_id, state, workflow, created_at, updated_at)
```

---

## Observabilidade

Logs em toda parte. Em um sistema event-driven com múltiplos componentes, saber o que aconteceu e quando é crítico.

Todo evento tem: timestamp, tipo, user_id, task_id (quando aplicável), payload, resultado.

---

## Estrutura do Repositório

```
lia/
├── server/                  # Go
│   ├── cmd/
│   │   ├── server/main.go
│   │   └── admin/main.go    # CLI de gestão
│   ├── internal/
│   │   ├── agent/
│   │   │   ├── planner.go
│   │   │   ├── executor.go
│   │   │   └── prompt.go
│   │   ├── llm/
│   │   │   └── client.go
│   │   ├── memory/
│   │   │   ├── store.go     # interface
│   │   │   └── postgres.go
│   │   ├── tools/
│   │   │   ├── registry.go
│   │   │   └── server/
│   │   ├── audio/
│   │   │   ├── transcriber.go
│   │   │   └── tts.go
│   │   ├── auth/
│   │   │   └── jwt.go
│   │   └── tasks/
│   │       └── state.go
│   ├── go.mod
│   └── go.sum
│
├── clients/                 # Rust
│   ├── desktop/
│   ├── cli/
│   └── discord/
│
├── ml/                      # Python
│   └── wakeword/
│       └── train.py
│
├── shared/
│   └── api-contract.md
│
└── docs/
    ├── lia-architecture.md
    └── lia-mindmap.svg
```

---

## Divisão de Responsabilidades

**Gui (server Go)**
- Core da IA (Planner, Executor, LLM, memória)
- PostgreSQL e modelos de dados
- Protocolo de comunicação e sessões
- Autenticação, autorização e níveis de confiança
- Áudio (STT e TTS)
- CLI de gestão de usuários
- Observabilidade e logs

**Yure (clients Rust)**
- Desktop App (wake word, áudio local, tools locais)
- CLI
- Discord Bot
- SDK de conexão com o server

---

## Roadmap

### Fase atual — descoberta por casos

Desenvolver a arquitetura através de casos reais, do simples ao complexo:

1. "Abra o Spotify."
2. "Abra o Spotify e coloque aquela playlist que eu estava ouvindo ontem."
3. "Baixe o relatório, analise, compare com o anterior e mande um resumo se houver mudança."
4. "Saí de casa. Apague as luzes, tranque a porta, faça backup, e me avise quando terminar. Se der erro, tente resolver."

Cada caso revela uma necessidade arquitetural real.

### Decisões abertas

Estas perguntas serão respondidas pelos casos de uso, não antecipadamente:

- Planner sequencial vs DAG
- Estrutura exata do Workflow
- Como memória é consultada durante planejamento
- Event Bus — implementação
- Protocolo de transporte (Socket.IO vs alternativas)
- Retry policy
- Cancelamento distribuído
- Persistência e recuperação de tasks após reboot
- Comunicação entre serviços futuros
- Como clients anunciam capabilities formalmente

---

## Visão de Longo Prazo

- Reconhecimento de voz por perfil para identificação passiva
- Consciência social: quem está online, memórias de grupo, mensagens entre usuários
- Tools colaborativas entre usuários
- Suporte a IoT e dispositivos físicos (C++)
- Múltiplos assistentes especializados orquestrados pelo mesmo core
