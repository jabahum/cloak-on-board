################################################################################
# Keycloak Onboarder
################################################################################

DEV_COMPOSE=docker compose -f docker-compose.dev.yml
PROD_COMPOSE=docker compose -f docker-compose.prod.yml --env-file .env

.DEFAULT_GOAL := help

help:
	@echo ""
	@echo "Keycloak Onboarder"
	@echo ""
	@echo "Development"
	@echo "  make dev              Start development environment"
	@echo "  make rebuild          Rebuild development environment"
	@echo "  make stop             Stop development environment"
	@echo "  make restart          Restart development services"
	@echo "  make clean            Stop and remove dev containers and volumes"
	@echo "  make reset            Full reset and rebuild"
	@echo ""
	@echo "Production"
	@echo "  make prod             Start production environment"
	@echo "  make prod-build       Build production images"
	@echo "  make prod-stop        Stop production environment"
	@echo "  make prod-logs        Production logs"
	@echo "  make prod-health      Production health check"
	@echo ""
	@echo "Logs"
	@echo "  make logs             All development logs"
	@echo "  make backend          Backend logs"
	@echo "  make frontend         Frontend logs"
	@echo "  make keycloak         Keycloak logs"
	@echo "  make postgres         PostgreSQL logs"
	@echo ""
	@echo "Database"
	@echo "  make migrate          Run migrations"
	@echo "  make psql             Open PostgreSQL shell"
	@echo "  make db-reset         Reset database"
	@echo "  make backup-db        Backup database"
	@echo "  make restore-db FILE=backups/file.sql"
	@echo ""
	@echo "Utilities"
	@echo "  make seed             Seed templates"
	@echo "  make health           Backend health check"
	@echo "  make ps               Show running containers"
	@echo "  make version          Show tool versions"
	@echo ""

################################################################################
# DEVELOPMENT
################################################################################

dev:
	$(DEV_COMPOSE) up --build

dev-detached:
	$(DEV_COMPOSE) up -d --build

rebuild:
	$(DEV_COMPOSE) down
	$(DEV_COMPOSE) up --build

stop:
	$(DEV_COMPOSE) down

restart:
	$(DEV_COMPOSE) restart

clean:
	$(DEV_COMPOSE) down -v --remove-orphans

reset:
	$(DEV_COMPOSE) down -v --remove-orphans
	docker volume prune -f
	$(DEV_COMPOSE) up --build

################################################################################
# PRODUCTION
################################################################################

prod:
	$(PROD_COMPOSE) up -d --build

prod-build:
	$(PROD_COMPOSE) build

prod-stop:
	$(PROD_COMPOSE) down

prod-restart:
	$(PROD_COMPOSE) restart

prod-logs:
	$(PROD_COMPOSE) logs -f

prod-backend:
	$(PROD_COMPOSE) logs -f backend

prod-frontend:
	$(PROD_COMPOSE) logs -f frontend

prod-nginx:
	$(PROD_COMPOSE) logs -f nginx

prod-health:
	curl -k https://localhost/health

################################################################################
# LOGS
################################################################################

logs:
	$(DEV_COMPOSE) logs -f

backend:
	$(DEV_COMPOSE) logs -f backend

frontend:
	$(DEV_COMPOSE) logs -f frontend

keycloak:
	$(DEV_COMPOSE) logs -f keycloak

postgres:
	$(DEV_COMPOSE) logs -f postgres

################################################################################
# STATUS
################################################################################

ps:
	$(DEV_COMPOSE) ps

prod-ps:
	$(PROD_COMPOSE) ps

################################################################################
# MIGRATIONS
################################################################################

migrate:
	$(DEV_COMPOSE) up migrate

prod-migrate:
	$(PROD_COMPOSE) up migrate

################################################################################
# SEEDING
################################################################################

seed:
	curl -X POST http://localhost:9000/api/v1/templates/seed

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
# SHELLS
################################################################################

shell-backend:
	$(DEV_COMPOSE) exec backend sh

shell-postgres:
	$(DEV_COMPOSE) exec postgres sh

shell-keycloak:
	$(DEV_COMPOSE) exec keycloak sh

psql:
	$(DEV_COMPOSE) exec postgres psql -U postgres onboarder

################################################################################
# DATABASE
################################################################################

db-reset:
	$(DEV_COMPOSE) down -v
	$(DEV_COMPOSE) up -d postgres
	$(DEV_COMPOSE) up migrate

backup-db:
	mkdir -p backups
	$(DEV_COMPOSE) exec -T postgres pg_dump -U postgres onboarder > backups/onboarder_$$(date +%Y%m%d_%H%M%S).sql

restore-db:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make restore-db FILE=backups/onboarder_YYYYMMDD_HHMMSS.sql"; \
		exit 1; \
	fi
	$(DEV_COMPOSE) exec -T postgres psql -U postgres onboarder < $(FILE)

################################################################################
# KEYCLOAK
################################################################################

keycloak-restart:
	$(DEV_COMPOSE) restart keycloak

export-realm:
	$(DEV_COMPOSE) exec keycloak /opt/keycloak/bin/kc.sh export \
		--realm onboarder \
		--file /opt/keycloak/data/import/onboarder-realm-export.json

################################################################################
# SERVICE RESTARTS
################################################################################

backend-restart:
	$(DEV_COMPOSE) restart backend

frontend-restart:
	$(DEV_COMPOSE) restart frontend

################################################################################
# VERIFY
################################################################################

health:
	curl http://localhost:9000/api/v1/health

################################################################################
# DOCKER
################################################################################

prune:
	docker system prune -f

################################################################################
# VERSION
################################################################################

version:
	@echo "Keycloak Onboarder"
	@echo ""
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