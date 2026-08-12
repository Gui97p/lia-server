# Lia Server

Server (Go) da Lia — uma plataforma de inteligência pessoal, não um chatbot. A Lia acompanha o usuário em múltiplos dispositivos, mantém memória contínua, reconhece quem está falando, e planeja/executa tarefas complexas com múltiplas etapas, falhas e recuperações. A referência é o Jarvis: uma inteligência que age, não apenas responde.

Este repositório contém apenas o server. Clients (desktop, CLI, Discord — em Rust) e treinamento de wake word (Python) vivem em repositórios próprios.

Projeto pessoal, de uso interno, em fase inicial de descoberta de arquitetura através de casos de uso reais.

## Stack

- **Go** — core do server (Planner, Executor, LLM, memória, auth)
- **PostgreSQL** — banco principal
- **WebSocket** — transporte bidirecional persistente com os clients

## Desenvolvimento

```bash
cp .env.example .env   # ajustar valores conforme necessário
make db-up             # sobe o Postgres local (Podman/Docker Compose)
make run                # roda o server
```

Outros comandos disponíveis no `Makefile` (`make build`, `make test`, `make db-down`, `make db-drop`, etc.).

## Documentação

A arquitetura completa do sistema — identidade e autenticação, memória, Planner/Executor, tasks e eventos, transporte, banco de dados, observabilidade, e as decisões ainda em aberto — está em [`docs/`](docs/).
