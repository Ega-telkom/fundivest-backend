# Makefile
.PHONY: help build up down logs clean test deploy
CONTAINER_ENGINE ?= docker

ifeq ($(shell command -v podman 2> /dev/null),)
    CONTAINER_ENGINE := docker
else
    CONTAINER_ENGINE := podman
endif

help:
	@echo "Available commands:"
	@echo "  make build         - Build docker images"
	@echo "  make up            - Start all services"
	@echo "  make down          - Stop all services"
	@echo "  make logs          - Show logs"
	@echo "  make clean         - Clean up volumes"
	@echo "  make test          - Run tests"
	@echo "  make deploy        - Deploy to production"

build:
	$(CONTAINER_ENGINE)-compose build

up:
	$(CONTAINER_ENGINE)-compose up -d
	@echo "Services started!"
	@echo "API: http://localhost:8080"

down:
	$(CONTAINER_ENGINE)-compose down

logs:
	$(CONTAINER_ENGINE)-compose logs -f

logs-api:
	$(CONTAINER_ENGINE)-compose logs -f api

logs-worker:
	$(CONTAINER_ENGINE)-compose logs -f worker

clean:
	$(CONTAINER_ENGINE)-compose down -v
	rm -rf storage/*

restart:
	$(CONTAINER_ENGINE)-compose restart

# Testing
test-setup:
	$(CONTAINER_ENGINE)-compose -f docker-compose.test.yml up -d
	sleep 3

test-teardown:
	$(CONTAINER_ENGINE)-compose -f docker-compose.test.yml down -v

test: test-setup
	@echo "=== TESTING START ==="
	go test ./... -v -coverprofile=coverage.out
	@echo "=== TESTING ENDED ==="
	$(MAKE) test-teardown
	
# CI
ci-test:
	go test ./internal/... -v -race -coverprofile=coverage.out
	go test ./tests/integration/... -v

ci-lint:
	golangci-lint run --timeout=5m

ci: ci-lint ci-test
	
# Database
db-migrate:
	@echo "Running migrations..."
	@for file in migrations/*.sql; do \
		echo "Applying $$file..."; \
		$(CONTAINER_ENGINE)-compose exec -T postgres psql -U certuser -d certdb -f - < $$file; \
	done
	@echo "Migrations complete!"

db-migrate-single:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make db-migrate-single FILE=001_create_certificates.sql"; \
		exit 1; \
	fi
	$(CONTAINER_ENGINE)-compose exec -T postgres psql -U certuser -d certdb -f - < migrations/$(FILE)

db-rollback:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make db-rollback FILE=001_create_certificates_down.sql"; \
		exit 1; \
	fi
	$(CONTAINER_ENGINE)-compose exec -T postgres psql -U certuser -d certdb -f - < migrations/$(FILE)

db-status:
	$(CONTAINER_ENGINE)-compose exec postgres psql -U certuser -d certdb -c "SELECT * FROM schema_migrations ORDER BY version;"

db-shell:
	$(CONTAINER_ENGINE)-compose exec postgres psql -U certuser -d certdb
	
# Start infrastructure only
dev-up:
	$(CONTAINER_ENGINE)-compose -f docker-compose.dev.yml up postgres valkey minio gotenberg -d
	@echo "Infrastructure started!"
	@echo "Postgres: localhost:5432"
	@echo "Valkey: localhost:6379"
	@echo "Gotenberg: localhost:3000"

dev-down:
	$(CONTAINER_ENGINE)-compose -f docker-compose.dev.yml down

# Run API with hot reload
dev-api:
	air -c .air.toml

# Run Worker with hot reload
dev-worker:
	air -c .air-worker.toml

# Run both (requires tmux/screen)
dev:
	@echo "Starting development environment..."
	$(MAKE) dev-up
	@echo "Run 'make dev-api' in one terminal and 'make dev-worker' in another"

# Deployment
deploy-build:
	$(CONTAINER_ENGINE)-compose -f docker-compose.prod.yml build

deploy-up:
	$(CONTAINER_ENGINE)-compose -f docker-compose.prod.yml up -d

deploy-down:
	$(CONTAINER_ENGINE)-compose -f docker-compose.prod.yml down