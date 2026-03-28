.PHONY: help setup start scheduler batch-run mcp-build frontend-dev backend-dev db-up db-down dev kill-dev seed migrate docker-up docker-down docker-build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# Setup & Start
# =============================================================================
setup: ## Initial setup (install deps + DB + migrate + seed)
	@echo "\033[36m=== Step 1/5: Install frontend dependencies ===\033[0m"
	cd frontend && npm install
	@echo ""
	@echo "\033[36m=== Step 2/5: Copy environment files ===\033[0m"
	@if [ ! -f backend/.env ]; then \
		cp backend/.env.example backend/.env; \
		echo "  created backend/.env"; \
	else \
		echo "  backend/.env already exists (skipped)"; \
	fi
	@if [ ! -f frontend/.env.local ]; then \
		cp frontend/.env.local.example frontend/.env.local; \
		echo "  created frontend/.env.local"; \
	else \
		echo "  frontend/.env.local already exists (skipped)"; \
	fi
	@echo ""
	@echo "\033[36m=== Step 3/5: Start PostgreSQL ===\033[0m"
	docker compose up -d db
	@echo "  waiting for PostgreSQL to be ready..."
	@sleep 3
	@echo ""
	@echo "\033[36m=== Step 4/5: Run migrations ===\033[0m"
	$(MAKE) migrate
	@echo ""
	@echo "\033[36m=== Step 5/5: Seed sample data ===\033[0m"
	$(MAKE) seed
	@echo ""
	@echo "\033[32m=== Setup complete! Run 'make start' to launch the app ===\033[0m"

start: kill-dev ## Start frontend + backend
	$(MAKE) -j2 backend-start frontend-dev

backend-start: ## Start backend server (go run, no hot-reload)
	cd backend && set -a && . .env && set +a && go run ./cmd/server

# =============================================================================
# Scheduler & Batch
# =============================================================================
scheduler: ## Start the local job scheduler (data collection + notifications)
	cd backend && set -a && . .env && set +a && go run ./cmd/scheduler

batch-run: ## Run a single batch job (usage: make batch-run JOB=trend-fetch)
	cd backend && set -a && . .env && set +a && BATCH_JOB_TYPE=$(JOB) go run ./cmd/batch

mcp-build: ## Build the MCP server binary for Claude Code
	cd backend && go build -o bin/trendbird-mcp ./cmd/mcp

# =============================================================================
# Development
# =============================================================================
frontend-dev: ## Start frontend development server
	cd frontend && npm run dev

backend-dev: ## Start backend development server (hot-reload)
	cd backend && set -a && . .env && set +a && air

db-up: ## Start PostgreSQL database
	docker compose up -d db

db-down: ## Stop and remove database container
	docker compose down

kill-dev: ## Kill processes on dev ports (3000, 8080)
	-lsof -ti:3000 | xargs kill -9 2>/dev/null || true
	-lsof -ti:8080 | xargs kill -9 2>/dev/null || true

dev: kill-dev ## Start all services (backend + frontend)
	$(MAKE) -j2 backend-dev frontend-dev

PSQL ?= $(shell which psql 2>/dev/null || echo /opt/homebrew/opt/postgresql@17/bin/psql)
migrate: ## Run database migrations
	@cd backend && set -a && . .env && set +a && \
	for f in migrations/*.up.sql; do \
		echo "Running: $$f"; \
		$(PSQL) "$$DATABASE_URL" -f "$$f"; \
	done

seed: ## Seed database with sample data
	cd backend && set -a && . .env && set +a && go run ./cmd/seed

# =============================================================================
# Docker Compose (full-stack)
# =============================================================================
docker-up: ## Start all services via Docker Compose (DB + Backend + Scheduler + Frontend)
	docker compose up -d --build

docker-down: ## Stop all Docker Compose services
	docker compose down

docker-build: ## Build all Docker images without starting
	docker compose build
