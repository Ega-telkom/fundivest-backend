# Makefile
.PHONY: help
CONTAINER_ENGINE ?= $(shell command -v podman 2>/dev/null || echo docker)
BRANCH = trunk

help:
	@echo "Why, hello there!"
	@echo ""
	@echo "Development:"
	@echo "  make dev-up          - Start dev infrastructure only"
	@echo "  make dev-down        - Stop dev infrastructure"
	@echo "  make dev-api         - Run API with hot reload"
	@echo "  make dev-worker      - Run worker with hot reload"
	@echo "  make dev-full        - Run all infrastructure"
	@echo ""
	@echo "Production:"
	@echo "  make prod-build      - Build all services"
	@echo "  make prod-up         - Start all services"
	@echo "  make prod-down       - Stop all services"
	@echo "  make prod-logs       - All services Logs"
	@echo "  make prod-logs-api   - API Logs"
	@echo "  make prod-logs-worker- Worker Logs"
	@echo "  make prod-restart    - Restart all services"
	@echo ""
	@echo "Testing:"
	@echo "  make test            - Run all tests with test DB"
	@echo ""
	@echo "Database:"
	@echo "  make db-migrate      - Run all migrations"
	@echo "  make db-shell        - Open psql shell"
	@echo ""
	@echo "Deployment:"
	@echo "  make deploy          - Deploy to production"
	@echo "  make deploy-health   - Check production health"

# ============================================================================
# DEVELOPMENT
# ============================================================================

.PHONY: dev-up dev-down dev-api dev-worker

dev-up:
	$(CONTAINER_ENGINE) compose up postgres valkey minio gotenberg -d

dev-down:
	$(CONTAINER_ENGINE) compose down

dev-api:
	air -c .air.toml

dev-worker:
	air -c .air-worker.toml

dev-full:
	$(CONTAINER_ENGINE) compose up -d

# ============================================================================
# PRODUCTION
# ============================================================================

.PHONY: prod-build prod-up prod-down prod-logs prod-restart

prod-build:
	$(CONTAINER_ENGINE) compose -f docker-compose.prod.yml --env-file .env.production build

prod-up:
	$(CONTAINER_ENGINE) compose -f docker-compose.prod.yml --env-file .env.production up -d

prod-down:
	$(CONTAINER_ENGINE) compose -f docker-compose.prod.yml down

prod-logs:
	$(CONTAINER_ENGINE) compose -f docker-compose.prod.yml logs -f

prod-logs-api:
	$(CONTAINER_ENGINE) compose -f docker-compose.prod.yml logs -f api

prod-logs-worker:
	$(CONTAINER_ENGINE) compose -f docker-compose.prod.yml logs -f worker

prod-restart:
	$(CONTAINER_ENGINE) compose -f docker-compose.prod.yml --env-file .env.production restart api worker

# ============================================================================
# DEPLOYMENT
# ============================================================================

.PHONY: deploy deploy-health

deploy:
	@echo "Pulling from remote source..."
	@git pull origin $(BRANCH)
	@echo "Building from source..."
	@$(CONTAINER_ENGINE) compose -f docker-compose.prod.yml --env-file .env.production build
	@$(CONTAINER_ENGINE) compose -f docker-compose.prod.yml --env-file .env.production up -d --no-deps api worker
	@$(CONTAINER_ENGINE) image prune -f
	@echo "Deploy complete!"
	@$(MAKE) deploy-health

deploy-health:
	@echo "Checking health..."
	@curl -sf http://localhost:8080/health && echo "API healthy" || echo "API down"
	@$(CONTAINER_ENGINE) compose -f docker-compose.prod.yml ps

# ============================================================================
# TESTING
# ============================================================================

.PHONY: test

test:
	@echo "Running tests..."
	$(CONTAINER_ENGINE) compose -f docker-compose.test.yml up -d
	@sleep 3
	go test ./... -v -coverprofile=coverage.out
	$(CONTAINER_ENGINE) compose -f docker-compose.test.yml down -v
	@echo "Tests complete!"
