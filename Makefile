.DEFAULT_GOAL := run

MAIN           ?= ./cmd/service-courier/main.go
BIN            ?= app
MIGRATIONS_DIR ?= ./migrations
GOOSE          ?= goose

GOOSE_CMD = $(GOOSE) -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) $(GOOSE_DBSTRING)

.PHONY: up down migrate migrate-down run build

up:
	docker compose up -d

down:
	docker compose down

migrate:
	$(GOOSE_CMD) up

migrate-down:
	$(GOOSE_CMD) down

run: up migrate
	go run $(MAIN)

build: up migrate
	go build -o $(BIN) $(MAIN)
