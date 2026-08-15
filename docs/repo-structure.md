# Estrutura do Repositório

Este repositório contém só o server (Go). Clients (Rust), treinamento de wake word (Python) e outros componentes vivem em repositórios próprios.

```
lia-server/
├── cmd/
│   ├── server/main.go
│   └── admin/               # CLI de gestão (users, tokens)
├── internal/
│   ├── agent/               # planner, executor, prompt (stubs)
│   ├── audio/               # STT / TTS (stubs)
│   ├── auth/                # JWT, trust levels
│   ├── config/              # .env → Config
│   ├── crypto/              # AES-GCM (Groq key em repouso)
│   ├── db/                  # pgxpool
│   ├── llm/                 # client Groq (stub)
│   ├── memory/              # store + postgres (stubs)
│   ├── messages/            # store + postgres
│   ├── session/             # Session, Hub, Writer (registry em memória)
│   ├── tasks/               # máquina de estados (stub)
│   ├── tools/               # registry + tools server-side (stubs)
│   ├── transport/           # HTTP + WebSocket (handshake, router, health)
│   └── users/               # store + postgres
├── docs/
├── migrations/              # SQL versionado (golang-migrate)
├── docker-compose.yml
├── Makefile
├── go.mod
└── go.sum
```

## Convenções

- `cmd/` — pontos de entrada (`func main()`). Cada subpasta é um binário compilado independentemente (`go build ./cmd/server`, `go build ./cmd/admin`), compartilhando a lógica em `internal/`.
- `internal/` — o compilador Go impede que qualquer coisa fora deste module importe pacotes daqui. Como este repo é só o server, praticamente toda a lógica mora aqui.
- Organização por domínio (`agent/`, `memory/`, `auth/`, `session/`, etc.), não por camada técnica (não há `controllers/`, `models/`, `services/`) — cada pasta concentra tudo que aquele conceito precisa.
- Padrão de interface + implementação (`store.go` interface, `postgres.go` implementação) onde faz sentido trocar o backend concreto sem afetar quem consome.

## Transporte vs sessão

- `transport/` — borda de rede: `Accept`/`Read`/`Write`/`Close` do WebSocket, handshake `auth`, envelope JSON, router de eventos, `GET /health`. Não exporta `*websocket.Conn` para o resto do sistema.
- `session/` — unidade de conexão autenticada (`Session`: `ConnID`, user, capabilities, `Writer`) e o `Hub` (mapa `ConnID → *Session`). Agent/tools/roteamento multi-device importam `session`, nunca `transport`.

Hoje o `transport.Server` cria e detém o `Hub` (Register/Unregister no ciclo de vida do WS). Quando o Executor precisar escolher uma conn por capability, o hub (ou uma interface fina) passa a ser injetado também no agent — o **tipo** continua em `session/`.
