.PHONY: dev dev-backend dev-frontend build test lint format emulators seed
SEED_FLAGS ?=
dev:
	./scripts/dev.sh
dev-backend:
	cd backend && set -a && . ./.env && set +a && go run ./cmd/api
dev-frontend:
	cd frontend && npm run dev
build:
	cd backend && go build -o bin/api ./cmd/api
	cd frontend && npm run build
test:
	cd backend && go test ./...
	cd frontend && npm test
lint:
	cd backend && go vet ./...
	cd frontend && npm run lint
format:
	cd backend && gofmt -w .
	cd frontend && npm run format
emulators:
	cd firebase && firebase emulators:start
seed:
	cd backend && set -a && . ./.env && set +a && go run ./cmd/seed $(SEED_FLAGS)
