# Lia Server

Server (Go) da Lia — uma plataforma de inteligência pessoal, não um chatbot. A Lia acompanha o usuário em múltiplos dispositivos, mantém memória contínua, reconhece quem está falando, e planeja/executa tarefas complexas com múltiplas etapas, falhas e recuperações. A referência é o Jarvis: uma inteligência que age, não apenas responde.

Este repositório contém apenas o server. Clients (desktop, CLI, Discord — em Rust) e treinamento de wake word (Python) vivem em repositórios próprios.

Projeto pessoal, de uso interno, em fase inicial de descoberta de arquitetura através de casos de uso reais.

## Stack

- **Go** — core do server (Planner, Executor, LLM, memória, auth)
- **PostgreSQL** — banco principal
- **WebSocket** — transporte bidirecional persistente com os clients

## Configurando em uma nova máquina

Pré-requisitos, do zero:

1. **Go** (via tarball oficial, não `apt` — evita versão desatualizada):
   ```bash
   curl -LO https://go.dev/dl/go1.26.5.linux-amd64.tar.gz
   sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.5.linux-amd64.tar.gz
   echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
   source ~/.bashrc
   ```
2. **Podman + podman-compose** (ou Docker — o `docker-compose.yml` funciona com ambos):
   ```bash
   sudo apt install podman podman-compose   # Ubuntu
   sudo dnf install podman podman-compose   # Fedora
   ```
3. **`golang-migrate` CLI:**
   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```
4. **Clonar o repo** (precisa de chave SSH cadastrada no GitHub, ou usar HTTPS):
   ```bash
   git clone git@github.com:Gui97p/lia-server.git
   ```
5. **Recriar o `.env`** — está no `.gitignore` de propósito (tem segredos), então não vem no clone:
   ```bash
   cp .env.example .env   # ajustar valores conforme necessário
   ```

## Desenvolvimento

```bash
make db-up          # sobe o Postgres local
make migrate-up      # aplica as migrations pendentes
make run             # roda o server
```

Outros comandos disponíveis no `Makefile` (`make build`, `make test`, `make db-down`, `make db-drop`, `make migrate-down`, `make migrate-create name=...`, etc.).

### Hot reload (opcional)

`make run` não reconstrói sozinho quando o código muda. Pra isso, instale o [`air`](https://github.com/air-verse/air) (fork ativo do antigo `cosmtrek/air`) e use `make dev` no lugar de `make run`:

```bash
go install github.com/air-verse/air@latest
make dev
```

Configuração em `.air.toml` (na raiz do repo) — reconstrói e reinicia o `cmd/server` a cada `.go` alterado.

## Documentação

A arquitetura completa do sistema — identidade e autenticação, memória, Planner/Executor, tasks e eventos, transporte, banco de dados, observabilidade, e as decisões ainda em aberto — está em [`docs/`](docs/).
