.PHONY: run run-admin build test tidy fmt vet clean db-up db-down db-drop

BIN_DIR := bin

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
	podman-compose up -d

db-down:
	podman-compose down

db-drop:
	podman-compose down -v
