.PHONY: help dev-backend dev-frontend test-backend test-frontend build-backend build-frontend lint

help:
	@echo "Available commands:"
	@echo "  make dev-backend   - Run the Go API server"
	@echo "  make dev-frontend  - Run the React frontend dev server with Bun"
	@echo "  make test-backend  - Run backend unit tests"
	@echo "  make test-frontend - Run frontend tests"
	@echo "  make build-backend - Compile Go backend binary"
	@echo "  make build-frontend- Build production frontend bundle"
	@echo "  make lint          - Run linters and checks"

dev-backend:
	cd backend && go run ./cmd/server/main.go

dev-frontend:
	cd frontend && bun run dev

test-backend:
	cd backend && go test -v ./...

test-frontend:
	cd frontend && bun test

build-backend:
	cd backend && go build -o bin/server ./cmd/server/main.go

build-frontend:
	cd frontend && bun run build

lint:
	cd backend && go vet ./...
