.PHONY: run run-admin build test tidy fmt vet clean db-up db-down db-drop migrate-up migrate-down migrate-create

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

run:
	go run ./cmd/server

run-admin:
	go run ./cmd/admin

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
