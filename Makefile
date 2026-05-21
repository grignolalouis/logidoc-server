.PHONY: dev dev-local build test lint prod sdk api-check

# Dev containerized with Air hot reload
dev:
	docker compose up --build

# Dev local (no container)
dev-local:
	go run ./cmd/server

# Build binary
build:
	go build -o bin/logidoc-server ./cmd/server

# Test
test:
	go test ./... -v -cover

# Lint
lint:
	golangci-lint run

# Prod containerized
prod:
	docker compose -f deploy/docker-compose.prod.yml up --build -d

# SDK generation
sdk:
	cd fern && fern check && fern generate --local

# OpenAPI validation
api-check:
	npx @redocly/cli lint openapi.yaml
