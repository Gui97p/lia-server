# Estrutura do Repositório

Este repositório contém só o server (Go). Clients (Rust), treinamento de wake word (Python) e outros componentes vivem em repositórios próprios.

```
lia-server/
├── cmd/
│   ├── server/main.go
│   └── admin/main.go        # CLI de gestão
├── internal/
│   ├── config/
│   │   └── config.go        # carregamento e validação de .env
│   ├── db/
│   │   └── db.go            # conexão/pool com Postgres (pgxpool)
│   ├── agent/
│   │   ├── planner.go
│   │   ├── executor.go
│   │   └── prompt.go
│   ├── llm/
│   │   └── client.go
│   ├── memory/
│   │   ├── store.go         # interface
│   │   └── postgres.go
│   ├── tools/
│   │   ├── registry.go
│   │   └── server/
│   ├── audio/
│   │   ├── transcriber.go
│   │   └── tts.go
│   ├── auth/
│   │   └── jwt.go
│   └── tasks/
│       └── state.go
├── docs/                     # esta documentação
├── docker-compose.yml        # Postgres local
├── Makefile
├── go.mod
└── go.sum
```

## Convenções

- `cmd/` — pontos de entrada (`func main()`). Cada subpasta é um binário compilado independentemente (`go build ./cmd/server`, `go build ./cmd/admin`), compartilhando a lógica em `internal/`.
- `internal/` — o compilador Go impede que qualquer coisa fora deste module importe pacotes daqui. Como este repo é só o server, praticamente toda a lógica mora aqui.
- Organização por domínio (`agent/`, `memory/`, `auth/`, etc.), não por camada técnica (não há `controllers/`, `models/`, `services/`) — cada pasta concentra tudo que aquele conceito precisa.
- Padrão de interface + implementação (`store.go` interface, `postgres.go` implementação) onde faz sentido trocar o backend concreto sem afetar quem consome.
