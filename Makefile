# Makefile for ORA SCRUM Backend

# Variables
DB_URL=postgresql://postgres:postgres@localhost:5432/ora_scrum?sslmode=disable
MIGRATIONS_PATH=./internal/db/migrations

# Docker commands
.PHONY: docker-up
docker-up:
	docker compose up -d

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: docker-clean
docker-clean:
	docker compose down -v
	docker system prune -f

# Database commands
.PHONY: db-create
db-create:
	docker exec -it ora_scrum_db createdb --username=postgres --owner=postgres ora_scrum

.PHONY: db-drop
db-drop:
	docker exec -it ora_scrum_db dropdb --username=postgres ora_scrum

.PHONY: db-reset
db-reset: docker-clean docker-up
	@echo "Waiting for database to be ready..."
	@sleep 3
	@echo "Database reset complete!"

# Migration commands
.PHONY: migrate-up
migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

.PHONY: migrate-down
migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down

.PHONY: migrate-down-all
migrate-down-all:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down -all

.PHONY: migrate-force
migrate-force:
	@read -p "Enter version to force: " version; \
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $$version

.PHONY: migrate-version
migrate-version:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

.PHONY: migrate-create
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $$name

# App commands
.PHONY: run
run:
	go run main.go

.PHONY: build
build:
	go build -o bin/ora-scrum main.go

.PHONY: test
test:
	go test -v ./...

# Development workflow
.PHONY: dev
dev: db-reset migrate-up run

.PHONY: fresh
fresh: db-reset migrate-up
	@echo "✅ Fresh database ready!"
	@echo "Run 'make run' to start the application"

# CI/CD commands
.PHONY: ci-test
ci-test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: ci-build
ci-build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/ora-scrum main.go

# Help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make docker-up        - Start Docker containers"
	@echo "  make docker-down      - Stop Docker containers"
	@echo "  make docker-clean     - Remove containers and volumes"
	@echo ""
	@echo "  make db-reset         - Reset database (clean + up)"
	@echo ""
	@echo "  make migrate-up       - Run all migrations"
	@echo "  make migrate-down     - Rollback last migration"
	@echo "  make migrate-version  - Check current migration version"
	@echo "  make migrate-create   - Create new migration"
	@echo ""
	@echo "  make test             - Run tests"
	@echo ""
	@echo "  make dev              - Full dev setup (reset + migrate + run)"
	@echo "  make fresh            - Fresh database only (reset + migrate)"