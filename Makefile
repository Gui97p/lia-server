.PHONY: run run-admin dev build test tidy fmt vet clean db-up db-down db-drop migrate-up migrate-down migrate-create migrate-force migrate-reset

-include .env
export

BIN_DIR := bin

ifneq ($(shell command -v podman-compose 2>/dev/null),)
COMPOSE := podman-compose
else ifneq ($(shell command -v docker-compose 2>/dev/null),)
COMPOSE := docker-compose
else
COMPOSE := docker compose
endif

default: build

run:
	go run ./cmd/server

run-admin:
	go run ./cmd/admin

dev:
	air

build:
	go build -o $(BIN_DIR)/lia-server ./cmd/server
	go build -o $(BIN_DIR)/lia-admin ./cmd/admin

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)

db-up:
	$(COMPOSE) up -d

db-down:
	$(COMPOSE) down

db-drop:
	$(COMPOSE) down -v

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	migrate -database "$(DATABASE_URL)" -path migrations up

migrate-down:
	migrate -database "$(DATABASE_URL)" -path migrations down 1

migrate-force:
	migrate -database "$(DATABASE_URL)" -path migrations force $(version)

migrate-reset:
	$(COMPOSE) exec postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -c "DROP TABLE IF EXISTS schema_migrations;"
