APP_NAME := app
PKG := ./...

.PHONY: build run test test-int test-int-all fmt vet lint tools dev-up dev-down prod-up prod-down migrate-up migrate-down migrate-new swagger sqlc sqlc-gen

build: sqlc-gen
	go build $(PKG)

run: sqlc-gen
	go run ./cmd/api

test: sqlc-gen
	go test -race -cover $(PKG)

test-int: sqlc-gen
	@echo "Running integration tests (requires Docker)..."
	go test -tags=integration ./internal/tests/integration -v

test-int-all: sqlc-gen
	@echo "Starting compose for integration tests..."
	docker compose -f docker-compose.test.yml up -d
	@echo "Running integration tests (DB+Redis)..."
	DB_HOST?=localhost DB_PORT?=55432 DB_USER?=appsechub DB_PASSWORD?=devpassword DB_NAME?=appsechub REDIS_ADDR?=localhost:56379 \
	go test -tags=integration ./internal/tests/integration -v || RET=$$?; \
	printf "\nStopping compose...\n"; docker compose -f docker-compose.test.yml down -v; \
	exit $${RET:-0}

fmt:
	go fmt $(PKG)

vet:
	go vet $(PKG)

lint:
	@golangci-lint run || echo "Install golangci-lint or run: make tools"

tools:
	@echo "Installing dev tools..."
	@go install github.com/cosmtrek/air@latest
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.61.0
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

sqlc:
	@which sqlc >/dev/null 2>&1 || go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@sqlc version

sqlc-gen: sqlc
	@echo "Generating sqlc code..."
	@sqlc generate

dev-up:
	docker compose -f docker-compose.dev.yml up

dev-down:
	docker compose -f docker-compose.dev.yml down -v

prod-up:
	docker compose up -d --build

prod-down:
	docker compose down -v

migrate-up:
	@echo "Migrations run automatically on app start; use this target if you wire CLI support."

migrate-down:
	@echo "Migrations run automatically on app start; use this target if you wire CLI support."


# Create a new SQL migration pair using golang-migrate (local binary if present, or Docker fallback)
# Usage: make migrate-new name=add_users_index
migrate-new:
	@test -n "$(name)" || (echo "Usage: make migrate-new name=<migration_name>" && exit 1)
	@if command -v migrate >/dev/null 2>&1; then \
		migrate create -seq -ext sql -dir migrations $(name); \
	else \
		echo "migrate not found, using Docker image migrate/migrate"; \
		docker run --rm -v $(PWD)/migrations:/migrations -w /migrations migrate/migrate create -seq -ext sql -dir /migrations $(name); \
	fi


