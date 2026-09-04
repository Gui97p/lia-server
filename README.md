# Lia Server

Go backend for Lia, a personal intelligence platform designed to persist context, reason across tasks, and execute actions on behalf of the user.

Unlike traditional chatbots, Lia is built around long-term memory, planning, execution, and continuous interaction across devices.

This is a personal project developed for real-world use. The architecture evolves through practical use cases and experimentation rather than extensive upfront design.

## Stack

* Go
* PostgreSQL
* WebSocket
* SearXNG

## Requirements

* Go 1.26+
* PostgreSQL
* SearXNG

## Getting Started

Clone the repository:

```bash
git clone git@github.com:Gui97p/lia-server.git
cd lia-server
```

Install the migration CLI:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Create the required configuration files:

```bash
cp .env.example .env
cp searxng/settings.example.yml searxng/settings.yml
```

Fill in the required values before running the application.

## Development

Start the local database:

```bash
make db-up
```

Apply pending migrations:

```bash
make migrate-up
```

Run the server:

```bash
make run
```

Additional commands are available through the Makefile:

```bash
make build
make test
make db-down
make db-drop
make migrate-down
make migrate-create name=<migration_name>
```

## Hot Reload

For automatic rebuilds during development, install Air:

```bash
go install github.com/air-verse/air@latest
```

Then run:

```bash
make dev
```

Configuration is defined in `.air.toml`.

## Goals

* Persistent memory
* Multi-step planning
* Task execution
* Failure recovery and replanning
* Multi-device synchronization
* Provider-agnostic LLM integration
