# Keycloak Onboarder

A lightweight platform for onboarding applications into **Keycloak** with minimal manual configuration.

The goal of this project is to eliminate repetitive Keycloak administration tasks by allowing administrators and developers to register an application once and automatically provision the required Keycloak resources.

---

# Features

## Application Management

- Register applications
- Import existing Keycloak clients
- Update application information
- Delete locally or delete linked Keycloak clients
- View application details
- Store application metadata
- Track provisioning status

## Keycloak Provisioning

- Create Keycloak clients
- Configure public and confidential clients
- Configure PKCE-enabled clients
- Configure redirect URIs
- Configure web origins
- Create client roles
- Create protocol mappers
- Assign client scopes
- Retrieve confidential client secrets
- Store provisioned client information

## Templates

The platform ships with reusable onboarding templates for:

- React SPA
- Backend API
- Flutter Mobile
- Machine-to-Machine Services

Each template provides sensible defaults for:

- Roles
- Redirect URIs
- Web Origins
- Client Scopes
- Protocol Mappers
- Client Configuration

## Provisioning Jobs

Every provisioning operation creates a job that tracks:

- Status
- Started time
- Completion time
- Individual provisioning steps
- Errors
- Execution history

## Security and Governance

- Keycloak user authentication with PKCE
- Viewer, manager, and administrator permissions
- Versioned approval requests for Keycloak mutations
- Append-only audit history
- Durable per-user in-app notifications

## Settings

Configure the Keycloak connection without modifying application code.

- Keycloak URL
- Realm
- Admin Client ID
- Admin Client Secret

---

# Technology Stack

## Backend

- Go
- Gin
- PostgreSQL
- pgx
- Docker

## Frontend

- React
- TypeScript
- Vite
- Carbon Design System
- Axios

## Infrastructure

- Docker Compose
- PostgreSQL
- Keycloak

---

# Repository Structure

```text
keycloak-onboarder/

├── backend/
│   ├── README.md
│   ├── cmd/
│   ├── internal/
│   ├── migrations/
│   └── Dockerfile
│
├── frontend/
│   ├── README.md
│   ├── src/
│   └── Dockerfile
│
├── docs/
│   ├── API.md
│   ├── ARCHITECTURE.md
│   ├── DEPLOYMENT.md
│   ├── DEVELOPMENT.md
│   └── CONTRIBUTING.md
│
├── docker-compose.yml
├── docker-compose.prod.yml
├── .env.example
└── README.md
```

---

# Documentation

The project documentation is split into focused guides.

| Document                                       | Description                                                                                  |
| ---------------------------------------------- | -------------------------------------------------------------------------------------------- |
| **[Backend Guide](backend/README.md)**         | Backend architecture, configuration, development, Docker, migrations, testing and deployment |
| **[Frontend Guide](frontend/README.md)**       | Frontend development, routing, API integration, building and deployment                      |
| **[Deployment Guide](docs/DEPLOYMENT.md)**     | Local, staging and production deployment instructions                                        |
| **[Architecture Guide](docs/ARCHITECTURE.md)** | System architecture, provisioning workflow and module overview                               |
| **[API Reference](docs/API.md)**               | REST API reference with request/response examples                                            |
| **[Development Guide](docs/DEVELOPMENT.md)**   | Local development workflow, coding standards and project conventions                         |

---

# Quick Start

## Prerequisites

- Docker
- Docker Compose

Clone the repository.

```bash
git clone <repository-url>

cd keycloak-onboarder
```

---

# Development

Start the complete development environment.

```bash
docker compose up --build
```

The following services will be available.

| Service         | URL                                 |
| --------------- | ----------------------------------- |
| Frontend        | http://localhost:3000               |
| Backend API     | http://localhost:9000               |
| Health Endpoint | http://localhost:9000/api/v1/health |
| Keycloak        | http://localhost:8080               |

## Start Individual Components

Backend

```bash
docker compose up backend
```

Frontend

```bash
docker compose up frontend
```

PostgreSQL

```bash
docker compose up postgres
```

Keycloak

```bash
docker compose up keycloak
```

Database Migrations

```bash
docker compose up migrate
```

Stop everything

```bash
docker compose down
```

Remove all containers and volumes

```bash
docker compose down -v
```

---

# Production

Production deployments should use the production compose file.

```bash
docker compose -f docker-compose.prod.yml up -d --build
```

Run migrations.

```bash
docker compose -f docker-compose.prod.yml up migrate
```

Stop production.

```bash
docker compose -f docker-compose.prod.yml down
```

Complete production instructions are available in:

**[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)**

---

# Provisioning Workflow

```text
Create Application
        │
        ▼
Select Template
        │
        ▼
Save Application
        │
        ▼
Provision Application
        │
        ▼
Create Keycloak Client
        │
        ▼
Create Client Roles
        │
        ▼
Assign Client Scopes
        │
        ▼
Create Protocol Mappers
        │
        ▼
Retrieve Client Secret
        │
        ▼
Update Application Status
        │
        ▼
Provision Complete
```

---

# Current API

## Applications

```
GET    /api/v1/applications
GET    /api/v1/applications/{id}
POST   /api/v1/applications
DELETE /api/v1/applications/{id}
POST   /api/v1/applications/{id}/provision
```

## Templates

```
GET  /api/v1/templates
GET  /api/v1/templates/{id}
POST /api/v1/templates/seed
```

## Settings

```
GET /api/v1/settings
PUT /api/v1/settings
```

## Provisioning Jobs

```
GET  /api/v1/jobs
GET  /api/v1/jobs/{id}
POST /api/v1/jobs
```

---

# MVP Roadmap

## Completed

- Project setup
- Docker environment
- Backend API
- React frontend
- Application management
- Templates
- Settings
- Provisioning jobs
- Keycloak Admin client
- Client provisioning
- Client roles
- Protocol mappers
- Client secret retrieval
- Middleware
  - Logger
  - Recovery
  - Request ID
  - CORS
  - API Authentication

## Planned

- Automatic database migrations
- Client scope management
- Existing client import
- Secret rotation
- Configuration drift detection
- Multi-environment support
- Multi-realm support
- Approval workflows
- Audit logging
- Developer Portal

---

# Long-Term Vision

The MVP is the foundation of a complete **Keycloak Application Lifecycle Management Platform**.

Future capabilities include:

- Environment promotion
- Approval workflows
- Configuration synchronization
- Secret rotation
- Application registry
- Audit logging
- Role templates
- Group management
- Health monitoring
- Developer portal
- Configuration export
- Documentation generation
- Multi-realm administration
- Multi-tenant support
- Identity governance

---

# License

MIT License
