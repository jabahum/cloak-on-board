# Development Guide

This guide is intended for developers contributing to the Keycloak Onboarder project. It covers project setup, development workflow, coding standards, testing, debugging, and contribution guidelines.

---

# Table of Contents

- Development Environment
- Project Setup
- Running the Application
- Repository Structure
- Development Workflow
- Coding Standards
- Git Workflow
- Database Migrations
- Adding New Modules
- Testing
- Debugging
- Logging
- Release Process
- Best Practices

---

# Development Environment

## Prerequisites

Install the following tools:

| Tool           | Version                       |
| -------------- | ----------------------------- |
| Go             | 1.25+                         |
| Node.js        | 22+                           |
| npm            | 10+                           |
| Docker         | Latest                        |
| Docker Compose | Latest                        |
| Git            | Latest                        |
| PostgreSQL     | Optional (Docker recommended) |

---

# Clone the Repository

```bash
git clone https://github.com/<organization>/keycloak-onboarder.git

cd keycloak-onboarder
```

---

# Start Development Environment

```bash
docker compose up --build
```

Services started:

- Frontend
- Backend
- PostgreSQL
- Keycloak
- Migration Runner

---

# Project Structure

```text
keycloak-onboarder/

backend/

frontend/

docs/

docker-compose.yml

docker-compose.prod.yml
```

---

# Backend Structure

```text
internal/

applications/

templates/

settings/

provisioning/

keycloak/

middleware/

database/

response/

server/
```

Each module follows the same layout.

```
handler.go

service.go

repository.go

model.go

dto.go
```

---

# Frontend Structure

```text
src/

api/

pages/

components/

layouts/

styles/

router/

types/
```

---

# Development Workflow

Typical workflow

```
Create Feature Branch

↓

Develop

↓

Run Tests

↓

Commit

↓

Push

↓

Open Pull Request

↓

Code Review

↓

Merge
```

---

# Running Individual Components

Backend

```bash
docker compose up backend
```

Frontend

```bash
docker compose up frontend
```

Database

```bash
docker compose up postgres
```

Keycloak

```bash
docker compose up keycloak
```

---

# Running Without Docker

Backend

```bash
cd backend

go run ./cmd/api
```

Frontend

```bash
cd frontend

npm install

npm run dev
```

Authentication requires the `onboarder` development realm and these backend
variables:

```text
AUTH_MODE=keycloak
KEYCLOAK_PUBLIC_URL=http://localhost:8080
KEYCLOAK_INTERNAL_URL=http://localhost:8080
KEYCLOAK_REALM=onboarder
KEYCLOAK_AUDIENCE=keycloak-onboarder-ui
```

The realm import includes development-only admin, manager, and viewer examples.
Assign production users through the Administrators, Managers, or Viewers group;
never reuse the development credentials.

---

# Coding Standards

## Go

Follow standard Go formatting.

```bash
go fmt ./...
```

Lint

```bash
golangci-lint run
```

Naming

- Packages are lowercase
- Interfaces end with "er" where appropriate
- Keep handlers thin
- Put business logic in services
- Database access belongs in repositories

---

## TypeScript

Run linting.

```bash
npm run lint
```

Formatting.

```bash
npm run format
```

Guidelines

- Prefer functional components
- Avoid `any`
- Define shared models under `src/types`
- Use the centralized API client

---

# Git Workflow

Branch naming

```
feature/application-import

feature/client-scopes

bugfix/cors

hotfix/provisioning
```

Commit examples

```
feat: add client scope provisioning

fix: resolve CORS issue

docs: update deployment guide

refactor: simplify provisioning service
```

---

# Database Migrations

Create a migration.

```bash
migrate create -ext sql -dir backend/migrations add_client_scopes
```

Run migrations.

```bash
docker compose up migrate
```

Migration naming

```
001_create_tables

002_add_templates

003_add_jobs
```

---

# Adding a New Backend Module

Create

```
internal/<module>/

handler.go

service.go

repository.go

model.go

dto.go
```

Register the routes in

```
internal/server/routes.go
```

Inject dependencies in

```
internal/server/server.go
```

---

# Adding a Frontend Page

Create

```
pages/

NewPage.tsx
```

Add API functions

```
src/api
```

Add models

```
src/types
```

Register route

```
router/
```

---

# Testing

Backend

Run all tests.

```bash
go test ./...
```

Run package.

```bash
go test ./internal/applications
```

Verbose.

```bash
go test -v ./...
```

---

Frontend

Run linting.

```bash
npm run lint
```

Build.

```bash
npm run build
```

---

# Debugging

Backend logs

```bash
docker compose logs -f backend
```

Frontend logs

Browser Developer Tools

Network tab

Console

---

# Logging

Every request includes

- Request ID
- Method
- Path
- Duration
- Status
- Client IP

Example

```
[HTTP]

RequestID=...

POST /applications

201 Created

Duration=14ms
```

---

# Environment Variables

Backend

```
DATABASE_URL

KEYCLOAK_BASE_URL

KEYCLOAK_REALM

KEYCLOAK_ADMIN_CLIENT_ID

KEYCLOAK_ADMIN_CLIENT_SECRET

ALLOWED_ORIGINS

API_AUTH_TOKEN
```

Frontend

```
VITE_API_BASE_URL
```

---

# Pull Request Checklist

Before opening a PR:

- Code builds successfully
- Tests pass
- Lint passes
- Documentation updated
- Migration included (if required)
- API changes documented
- Screenshots included (UI changes)

---

# Release Process

1. Merge into `main`
2. Create release tag
3. Build Docker images
4. Run migrations
5. Deploy backend
6. Deploy frontend
7. Verify health endpoint
8. Smoke test provisioning

---

# Best Practices

Backend

- Keep handlers thin
- Business logic belongs in services
- Database access belongs in repositories
- Return consistent API responses
- Never log secrets

Frontend

- Reuse components
- Keep pages focused
- Centralize API calls
- Reuse shared types
- Avoid duplicated state

General

- Keep modules independent
- Write documentation alongside features
- Prefer composition over duplication
- Use meaningful commit messages
- Review before merging

---

# Future Improvements

- GitHub Actions CI
- Automated testing
- End-to-end tests
- Code coverage reports
- Dependency scanning
- Security scanning
- Automatic documentation generation
- Release automation

---

# Related Documentation

- [Project README](../README.md)
- [Backend Guide](../backend/README.md)
- [Frontend Guide](../frontend/README.md)
- [Architecture Guide](ARCHITECTURE.md)
- [API Reference](API.md)
- [Deployment Guide](DEPLOYMENT.md)
