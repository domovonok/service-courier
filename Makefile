.DEFAULT_GOAL := run

MAIN ?= ./cmd/service-courier/main.go
BIN ?= app
GOOSE_CMD = goose -dir migrations $(GOOSE_DRIVER) $(GOOSE_DBSTRING)

.PHONY: migrate migrate-down up down run build

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
