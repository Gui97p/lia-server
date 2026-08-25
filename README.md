# Lia Server

Go backend for Lia — a personal intelligence platform, not a chatbot. Lia follows the user across multiple devices, keeps persistent memory, and plans/executes multi-step tasks with failure recovery and replanning. The reference point is Jarvis: an intelligence that acts, not just responds.

Personal project, built for real, internal use — the architecture is discovered through actual use cases rather than designed upfront.

## Stack

- **Go** — core server (Planner, Executor, multi-provider LLM routing, memory, auth)
- **PostgreSQL** — primary datastore
- **WebSocket** — persistent, bidirectional transport with clients

## Setting up on a new machine

Prerequisites, from scratch:

1. **Go** (via the official tarball, not `apt` — avoids an outdated version):
   ```bash
   curl -LO https://go.dev/dl/go1.26.5.linux-amd64.tar.gz
   sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.5.linux-amd64.tar.gz
   echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
   source ~/.bashrc
   ```
2. **Podman + podman-compose** (or Docker — `docker-compose.yml` works with either):
   ```bash
   sudo apt install podman podman-compose   # Ubuntu
   sudo dnf install podman podman-compose   # Fedora
   ```
3. **`golang-migrate` CLI:**
   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```
4. **Clone the repo** (needs an SSH key registered on GitHub, or use HTTPS):
   ```bash
   git clone git@github.com:Gui97p/lia-server.git
   ```
5. **Recreate `.env` and `searxng/settings.yml`, then fill them in.** Both are in `.gitignore` on purpose (they hold secrets), so they don't come with the clone:
   ```bash
   cp .env.example .env
   cp searxng/settings.example.yml searxng/settings.yml
   ```

## Development

```bash
make db-up          # start the local Postgres container
make migrate-up      # apply pending migrations
make run             # run the server
```

More commands available in the `Makefile` (`make build`, `make test`, `make db-down`, `make db-drop`, `make migrate-down`, `make migrate-create name=...`, etc.).

### Hot reload (optional)

`make run` doesn't rebuild automatically on code changes. For that, install [`air`](https://github.com/air-verse/air) (active fork of the old `cosmtrek/air`) and use `make dev` instead of `make run`:

```bash
go install github.com/air-verse/air@latest
make dev
```

Configured in `.air.toml` (repo root) — rebuilds and restarts `cmd/server` on every `.go` change.

## Documentation

The full system architecture — identity and auth, memory, Planner/Executor, tasks and events, transport, database, observability, and open design questions — lives in [`docs/`](docs/).

The docs themselves are in Portuguese (this README is the English entry point) — the project is used daily in Portuguese, and that's the language the design decisions were actually made and argued in.
