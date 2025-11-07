.PHONY: help build run test clean exec sqlc forward-stripe migrate-create migrate-up migrate-down migrate-reset migrate-status migrate-force lint-sql fix-sql psql

ENV ?= dev

ifdef env
    ENV := $(env)
endif

ifeq ($(ENV),dev)
    ENV_FILE := .env
else ifeq ($(ENV),stage)
    ENV_FILE := .env.stage
else ifeq ($(ENV),prod)
    ENV_FILE := .env.prod
else
    $(error Invalid environment: $(ENV). Must be one of: dev, stage, prod)
endif

ifneq (,$(wildcard $(ENV_FILE)))
    include $(ENV_FILE)
    export
endif

export ENV

help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make forward-stripe - Forward Stripe webhook to localhost"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make sqlc           - Generate Go code from SQL"
	@echo "  make exec <action>  - Execute a tool (e.g., make exec createuser)"
	@echo "  make migrate-create - Create a new migration (usage: make migrate-create name=create_users_table)"
	@echo "  make migrate-up     - Run all pending migrations"
	@echo "  make migrate-down   - Rollback the last migration"
	@echo "  make migrate-reset  - Drop all migrations and re-apply them (WARNING: destructive)"
	@echo "  make migrate-status - Show migration status"
	@echo "  make migrate-force  - Force set migration version (usage: make migrate-force version=1)"
	@echo "  make seed           - Seed the database"
	@echo "  make psql           - Connect to PostgreSQL database"
	@echo "  make lint-sql       - Lint SQL files"
	@echo "  make fix-sql        - Fix SQL files"
	@echo ""
	@echo "Environment:"
	@echo "  All commands accept an 'env' parameter (default: dev)"
	@echo "  Usage: make <command> env=<dev|stage|prod>"
	@echo "  Examples:"
	@echo "    make run env=dev     - Run with .env file (default)"
	@echo "    make run env=stage   - Run with .env.stage file"
	@echo "    make run env=prod    - Run with .env.prod file"
	@echo "    make psql env=stage  - Connect to stage database"

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

exec:
	@echo "Executing tool..."
	@if [ -z "$(filter-out exec,$(MAKECMDGOALS))" ]; then \
		echo "Error: action is required. Usage: make exec createuser"; \
		exit 1; \
	fi
	go run cmd/tools/main.go $(filter-out exec,$(MAKECMDGOALS))

ifneq (,$(filter exec,$(MAKECMDGOALS)))
%:
	@:
endif

sqlc:
	@echo "Generating code with sqlc..."
	sqlc generate

forward-stripe:
	@echo "Forwarding Stripe webhook to localhost..."
	stripe listen --forward-to http://localhost:8080/stripe/webhook

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)"
	migrate create -ext sql -dir internal/platform/db/schema/migrations -seq $(name)

migrate-up:
	@echo "Running migrations up..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL environment variable is not set"; \
		exit 1; \
	fi
	migrate -path internal/platform/db/schema/migrations -database "$(DB_URL)" up

migrate-down:
	@echo "Rolling back migration..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL environment variable is not set"; \
		exit 1; \
	fi
	migrate -path internal/platform/db/schema/migrations -database "$(DB_URL)" down 1

migrate-reset:
	@echo "WARNING: This will drop all migrations and re-apply them!"
	@echo "Press Ctrl+C to cancel, or wait 3 seconds to continue..."
	@sleep 3
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL environment variable is not set"; \
		exit 1; \
	fi
	@echo "Dropping all migrations..."
	migrate -path internal/platform/db/schema/migrations -database "$(DB_URL)" down -all || true
	@echo "Re-applying all migrations..."
	migrate -path internal/platform/db/schema/migrations -database "$(DB_URL)" up

migrate-status:
	@echo "Migration status..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL environment variable is not set"; \
		exit 1; \
	fi
	migrate -path internal/platform/db/schema/migrations -database "$(DB_URL)" version

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
	migrate -path internal/platform/db/schema/migrations -database "$(DB_URL)" force $(version)

seed:
	@echo "Seeding database..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL environment variable is not set"; \
		exit 1; \
	fi
	psql "$(DB_URL)" < internal/platform/db/seed/artworks.sql

psql:
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL environment variable is not set"; \
		exit 1; \
	fi
	psql "$(DB_URL)"

lint-sql:
	sqlfluff lint internal/platform/db/schema/migrations/
	sqlfluff lint internal/platform/db/schema/queries/

fix-sql:
	sqlfluff fix internal/platform/db/schema/migrations/
	sqlfluff fix internal/platform/db/schema/queries/