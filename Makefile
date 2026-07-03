################################################################################
# Keycloak Onboarder
#
# Development & Deployment Makefile
################################################################################

COMPOSE=docker compose

################################################################################
# HELP
################################################################################

.DEFAULT_GOAL := help

help:
	@echo ""
	@echo "Keycloak Onboarder"
	@echo ""
	@echo "Available Commands"
	@echo ""
	@echo "Development"
	@echo "  make dev              Start development environment"
	@echo "  make rebuild          Rebuild development environment"
	@echo "  make stop             Stop development environment"
	@echo "  make clean            Stop and remove containers & volumes"
	@echo ""
	@echo "Production"
	@echo "  make prod             Start production environment"
	@echo "  make prod-build       Build production images"
	@echo "  make prod-stop        Stop production"
	@echo ""
	@echo "Logs"
	@echo "  make logs             All logs"
	@echo "  make backend          Backend logs"
	@echo "  make frontend         Frontend logs"
	@echo "  make keycloak         Keycloak logs"
	@echo "  make postgres         PostgreSQL logs"
	@echo ""
	@echo "Utilities"
	@echo "  make migrate          Run migrations"
	@echo "  make restart          Restart all services"
	@echo "  make ps               Running containers"
	@echo ""
	@echo "Backend"
	@echo "  make gofmt            Format Go code"
	@echo "  make gotest           Run Go tests"
	@echo ""
	@echo "Frontend"
	@echo "  make npm-install      Install npm packages"
	@echo "  make lint             Run frontend lint"
	@echo "  make build            Build frontend"
	@echo ""
	@echo "Docker"
	@echo "  make prune            Remove unused Docker resources"
	@echo ""

################################################################################
# DEVELOPMENT
################################################################################

dev:
	$(COMPOSE) up --build

rebuild:
	$(COMPOSE) down
	$(COMPOSE) up --build

stop:
	$(COMPOSE) down

restart:
	$(COMPOSE) restart

clean:
	$(COMPOSE) down -v --remove-orphans

################################################################################
# PRODUCTION
################################################################################

prod:
	$(COMPOSE) -f docker-compose.prod.yml --env-file .env up -d --build

prod-build:
	$(COMPOSE) -f docker-compose.prod.yml --env-file .env build

prod-stop:
	$(COMPOSE) -f docker-compose.prod.yml down

################################################################################
# LOGS
################################################################################

logs:
	$(COMPOSE) logs -f

backend:
	$(COMPOSE) logs -f backend

frontend:
	$(COMPOSE) logs -f frontend

keycloak:
	$(COMPOSE) logs -f keycloak

postgres:
	$(COMPOSE) logs -f postgres

################################################################################
# STATUS
################################################################################

ps:
	$(COMPOSE) ps

################################################################################
# MIGRATIONS
################################################################################

migrate:
	$(COMPOSE) up migrate

################################################################################
# GO
################################################################################

gofmt:
	cd backend && go fmt ./...

gotest:
	cd backend && go test ./...

################################################################################
# FRONTEND
################################################################################

npm-install:
	cd frontend && npm install

lint:
	cd frontend && npm run lint

build:
	cd frontend && npm run build

################################################################################
# DOCKER
################################################################################

prune:
	docker system prune -f

################################################################################
# DEVELOPMENT UTILITIES
################################################################################

shell-backend:
	$(COMPOSE) exec backend sh

shell-postgres:
	$(COMPOSE) exec postgres sh

psql:
	$(COMPOSE) exec postgres psql -U postgres onboarder

shell-keycloak:
	$(COMPOSE) exec keycloak sh

################################################################################
# DATABASE
################################################################################

db-reset:
	$(COMPOSE) down -v
	$(COMPOSE) up -d postgres
	$(COMPOSE) up migrate

################################################################################
# KEYCLOAK
################################################################################

keycloak-restart:
	$(COMPOSE) restart keycloak

################################################################################
# FRONTEND
################################################################################

frontend-restart:
	$(COMPOSE) restart frontend

################################################################################
# BACKEND
################################################################################

backend-restart:
	$(COMPOSE) restart backend

################################################################################
# FULL RESET
################################################################################

reset:
	$(COMPOSE) down -v --remove-orphans
	docker volume prune -f
	$(COMPOSE) up --build

################################################################################
# VERIFY
################################################################################

health:
	curl http://localhost:9000/api/v1/health

################################################################################
# VERSION
################################################################################

version:
	@echo "Keycloak Onboarder"
	@echo "Docker Compose:"
	@docker compose version
	@echo ""
	@echo "Docker:"
	@docker version --format '{{.Server.Version}}'
	@echo ""
	@echo "Go:"
	@go version
	@echo ""
	@echo "Node:"
	@node --version
	@echo ""
	@echo "npm:"
	@npm --version


################################################################################
# SEEDING
################################################################################

seed:
	curl -X POST http://localhost:9000/api/v1/templates/seed

################################################################################
# DATABASE BACKUP / RESTORE
################################################################################

backup-db:
	mkdir -p backups
	$(COMPOSE) exec -T postgres pg_dump -U postgres onboarder > backups/onboarder_$$(date +%Y%m%d_%H%M%S).sql

restore-db:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make restore-db FILE=backups/onboarder_YYYYMMDD_HHMMSS.sql"; \
		exit 1; \
	fi
	$(COMPOSE) exec -T postgres psql -U postgres onboarder < $(FILE)

################################################################################
# KEYCLOAK REALM EXPORT
################################################################################

export-realm:
	mkdir -p infrastructure/keycloak/export
	$(COMPOSE) exec keycloak /opt/keycloak/bin/kc.sh export \
		--realm onboarder \
		--file /opt/keycloak/data/import/onboarder-realm-export.json

################################################################################
# PRODUCTION LOGS
################################################################################

prod-logs:
	$(COMPOSE) -f docker-compose.prod.yml --env-file .env logs -f

prod-backend:
	$(COMPOSE) -f docker-compose.prod.yml --env-file .env logs -f backend

prod-frontend:
	$(COMPOSE) -f docker-compose.prod.yml --env-file .env logs -f frontend

prod-nginx:
	$(COMPOSE) -f docker-compose.prod.yml --env-file .env logs -f nginx

################################################################################
# PRODUCTION HEALTH
################################################################################

prod-health:
	curl -k https://localhost/health