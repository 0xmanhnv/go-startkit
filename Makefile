APP_NAME := app
PKG := ./...

.PHONY: build run test test-int test-int-all fmt vet lint tools dev-up dev-down prod-up prod-down migrate-up migrate-down swagger

build:
	go build $(PKG)

run:
	go run ./cmd/api

test:
	go test -race -cover $(PKG)

test-int:
	@echo "Running integration tests (requires Docker)..."
	go test -tags=integration ./internal/tests/integration -v

test-int-all:
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


