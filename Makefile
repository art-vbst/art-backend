.PHONY: help build run test clean sqlc migrate-create migrate-up migrate-down migrate-reset migrate-status migrate-force lint-sql fix-sql

ifneq (,$(wildcard .env))
    include .env
    export
endif

help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make sqlc           - Generate Go code from SQL"
	@echo "  make migrate-create - Create a new migration (usage: make migrate-create name=create_users_table)"
	@echo "  make migrate-up     - Run all pending migrations"
	@echo "  make migrate-down   - Rollback the last migration"
	@echo "  make migrate-reset  - Drop all migrations and re-apply them (WARNING: destructive)"
	@echo "  make migrate-status - Show migration status"
	@echo "  make migrate-force  - Force set migration version (usage: make migrate-force version=1)"
	@echo "  make lint-sql       - Lint SQL files"
	@echo "  make fix-sql        - Fix SQL files"

build:
	@echo "Building application..."
	go build -o bin/server cmd/server/main.go

run:
	@echo "Running application..."
	go run cmd/server/main.go

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/

sqlc:
	@echo "Generating code with sqlc..."
	sqlc generate

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)"
	migrate create -ext sql -dir internal/platform/db/migrations -seq $(name)

migrate-up:
	@echo "Running migrations up..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL environment variable is not set"; \
		exit 1; \
	fi
	migrate -path internal/platform/db/migrations -database "$(DB_URL)" up

migrate-down:
	@echo "Rolling back migration..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL environment variable is not set"; \
		exit 1; \
	fi
	migrate -path internal/platform/db/migrations -database "$(DB_URL)" down 1

migrate-reset:
	@echo "WARNING: This will drop all migrations and re-apply them!"
	@echo "Press Ctrl+C to cancel, or wait 3 seconds to continue..."
	@sleep 3
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL environment variable is not set"; \
		exit 1; \
	fi
	@echo "Dropping all migrations..."
	migrate -path internal/platform/db/migrations -database "$(DB_URL)" down -all || true
	@echo "Re-applying all migrations..."
	migrate -path internal/platform/db/migrations -database "$(DB_URL)" up

migrate-status:
	@echo "Migration status..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL environment variable is not set"; \
		exit 1; \
	fi
	migrate -path internal/platform/db/migrations -database "$(DB_URL)" version

migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "Error: version is required. Usage: make migrate-force version=1"; \
		exit 1; \
	fi
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL environment variable is not set"; \
		exit 1; \
	fi
	@echo "Forcing migration version to $(version)..."
	migrate -path internal/platform/db/migrations -database "$(DB_URL)" force $(version)

lint-sql:
	sqlfluff lint internal/platform/db/migrations/
	sqlfluff lint internal/platform/db/queries/

fix-sql:
	sqlfluff fix internal/platform/db/migrations/
	sqlfluff fix internal/platform/db/queries/