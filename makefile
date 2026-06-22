FRONTEND_DIR = ./web/default
FRONTEND_CLASSIC_DIR = ./web/classic
BACKEND_DIR = .

.PHONY: all build-frontend build-frontend-classic build-all-frontends start-backend dev dev-api dev-web dev-web-classic docs-check docs-diff docs-publish snapshot snapshot-list snapshot-show snapshot-diff snapshot-restore

all: build-all-frontends start-backend

build-frontend:
	@echo "Building default frontend..."
	@cd $(FRONTEND_DIR) && bun install && DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat ../../VERSION) bun run build

build-frontend-classic:
	@echo "Building classic frontend..."
	@cd $(FRONTEND_CLASSIC_DIR) && bun install && VITE_REACT_APP_VERSION=$(cat ../../VERSION) bun run build

build-all-frontends: build-frontend build-frontend-classic

start-backend:
	@echo "Starting backend dev server..."
	@cd $(BACKEND_DIR) && go run main.go &

dev-api:
	@echo "Starting backend services (docker)..."
	@docker compose -f docker-compose.dev.yml up -d

dev-web:
	@echo "Starting frontend dev server..."
	@cd $(FRONTEND_DIR) && bun install && bun run dev

dev-web-classic:
	@echo "Starting classic frontend dev server..."
	@cd $(FRONTEND_CLASSIC_DIR) && bun install && bun run dev

dev: dev-api dev-web

docs-check:
	@bash scripts/docs/check-api-docs.sh

docs-diff:
	@go run ./scripts/docs/generate-doc-diff "$(or $(FROM),docs/api/current.md)" "$(or $(TO),docs/api/current.md)"

docs-publish:
	@go run ./scripts/docs/publish-api-docs

snapshot:
	@./scripts/git-snapshot create "$(or $(MSG),manual snapshot)"

snapshot-list:
	@./scripts/git-snapshot list

snapshot-show:
	@./scripts/git-snapshot show "$(or $(REF),latest)"

snapshot-diff:
	@./scripts/git-snapshot diff "$(or $(REF),latest)"

snapshot-restore:
	@./scripts/git-snapshot restore "$(or $(REF),latest)"
