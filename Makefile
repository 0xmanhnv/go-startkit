APP_NAME := appsechub
PKG := ./...

.PHONY: build run test fmt vet lint tools dev-up dev-down prod-up prod-down migrate-up migrate-down swagger

build:
	go build $(PKG)

run:
	go run ./cmd/api

test:
	go test -race -cover $(PKG)

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


