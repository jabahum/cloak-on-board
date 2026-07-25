# Deployment Guide

This guide describes how to deploy the Keycloak Onboarder platform in development, staging, and production environments.

---

# Table of Contents

- Architecture
- Prerequisites
- Development Deployment
- Production Deployment
- Docker Compose
- Environment Variables
- Keycloak Configuration
- PostgreSQL
- Reverse Proxy
- SSL/TLS
- Backups
- Monitoring
- Upgrading
- Troubleshooting

---

# Architecture

```
                    Internet
                        │
                        ▼
                Reverse Proxy (Nginx)
                        │
          ┌─────────────┴─────────────┐
          ▼                           ▼
     Frontend                    Backend API
                                       │
                    ┌──────────────────┴──────────────────┐
                    ▼                                     ▼
              PostgreSQL                           Keycloak
```

The backend communicates with:

- PostgreSQL
- Keycloak Admin REST API

The frontend communicates only with the backend.

---

# Prerequisites

Install:

- Docker
- Docker Compose
- Git

Verify Docker.

```bash
docker --version
docker compose version
```

---

# Development Deployment

Clone the repository.

```bash
git clone https://github.com/<organization>/keycloak-onboarder.git

cd keycloak-onboarder
```

Copy environment variables.

```bash
cp .env.example .env
```

Start all services.

```bash
docker compose up --build
```

Services

| Service    | URL                   |
| ---------- | --------------------- |
| Frontend   | http://localhost:3000 |
| Backend    | http://localhost:9000 |
| Keycloak   | http://localhost:8080 |
| PostgreSQL | localhost:5432        |

---

# Starting Individual Services

Backend

```bash
docker compose up backend
```

Frontend

```bash
docker compose up frontend
```

Keycloak

```bash
docker compose up keycloak
```

Database

```bash
docker compose up postgres
```

Run migrations

```bash
docker compose up migrate
```

---

# Production Deployment

Use a dedicated production compose file.

```
docker-compose.prod.yml
```

Deploy.

```bash
docker compose -f docker-compose.prod.yml up -d --build
```

Verify.

```bash
docker ps
```

---

# Environment Variables

## Backend

```env
APP_ENV=production

SERVER_PORT=9000

DATABASE_URL=postgres://user:password@postgres:5432/onboarder?sslmode=disable

AUTH_MODE=keycloak

KEYCLOAK_PUBLIC_URL=https://auth.example.com

KEYCLOAK_INTERNAL_URL=http://keycloak:8080

KEYCLOAK_BASE_URL=https://auth.example.com

KEYCLOAK_REALM=onboarder

KEYCLOAK_AUDIENCE=keycloak-onboarder-ui

KEYCLOAK_ADMIN_CLIENT_ID=onboarder-admin

KEYCLOAK_ADMIN_CLIENT_SECRET=<secret>

CREDENTIAL_ENCRYPTION_KEYS=v2:<base64-32-byte-key>,v1:<previous-key>

SECRET_DELIVERY_TTL_MINUTES=10

DRIFT_CHECK_INTERVAL_MINUTES=15

ALLOWED_ORIGINS=https://onboarder.example.com

```

Production startup fails unless Keycloak authentication is fully configured.
API-key mode is development/test only.

Production also fails without `CREDENTIAL_ENCRYPTION_KEYS`. Put the newest key
first and retain previous versions while rotating encryption material. Do not
reuse the authentication-realm administrative client for managed realms.

For each target realm, create a confidential service-account client with
`realm-management/manage-clients`, `view-clients`, `query-clients`, and
`view-realm`. Add it through the realm-connections API or admin UI. The secret
is write-only and encrypted immediately.

Promotion order is controlled by the database. Production-like environments
must be marked protected so deployment, rollback, reconciliation, and secret
rotation use an approval with a different reviewer.

---

## Frontend

```env
VITE_API_BASE_URL=https://api.example.com/api/v1
VITE_KEYCLOAK_URL=https://auth.example.com
VITE_KEYCLOAK_REALM=onboarder
VITE_KEYCLOAK_CLIENT_ID=keycloak-onboarder-ui
```

---

# Keycloak

Create a confidential client.

```
Client ID

onboarder-admin
```

Enable

- Client Authentication
- Service Accounts

Assign roles

```
realm-management/manage-clients

realm-management/manage-realm

realm-management/view-clients
```

Copy the client secret.

Save it using the Settings page.

---

# PostgreSQL

Run migrations.

```bash
docker compose up migrate
```

Reset database.

```bash
docker compose down -v
```

Back up database.

```bash
docker exec postgres pg_dump \
-U postgres onboarder \
> backup.sql
```

Restore database.

```bash
cat backup.sql | docker exec -i postgres psql \
-U postgres onboarder
```

---

# Reverse Proxy

Example Nginx configuration.

```nginx
server {

    listen 80;

    server_name onboarder.example.com;

    location / {

        proxy_pass http://frontend:3000;

    }

    location /api {

        proxy_pass http://backend:9000;

    }

}
```

---

# HTTPS

Install Let's Encrypt.

```bash
sudo apt install certbot
```

Generate certificates.

```bash
sudo certbot --nginx
```

Verify.

```
https://onboarder.example.com
```

---

# Docker Commands

Start

```bash
docker compose up -d
```

Stop

```bash
docker compose down
```

Restart

```bash
docker compose restart
```

Logs

```bash
docker compose logs -f
```

Backend logs

```bash
docker compose logs -f backend
```

Frontend logs

```bash
docker compose logs -f frontend
```

---

# Health Checks

Backend

```bash
curl http://localhost:9000/api/v1/health
```

Frontend

```
http://localhost:3000
```

Keycloak

```
http://localhost:8080
```

---

# Monitoring

Recommended

- Prometheus
- Grafana
- Loki

Monitor

- CPU
- Memory
- API latency
- Request rate
- Provisioning jobs
- Database connections

---

# Logging

All backend requests include

- Request ID
- Method
- Path
- Duration
- Status Code
- IP Address

---

# Upgrading

Pull latest changes.

```bash
git pull
```

Rebuild.

```bash
docker compose build
```

Restart.

```bash
docker compose up -d
```

Run migrations.

```bash
docker compose up migrate
```

---

# Backup Strategy

Daily

- PostgreSQL
- Docker volumes

Weekly

- Full server snapshot

Monthly

- Restore backup into staging

---

# Disaster Recovery

Restore database.

Restore Docker volumes.

Run

```bash
docker compose up
```

Verify

- Applications
- Templates
- Jobs
- Keycloak connectivity

---

# Troubleshooting

## CORS

Verify

```
ALLOWED_ORIGINS
```

---

## Cannot connect to PostgreSQL

Check

```bash
docker compose logs postgres
```

---

## Cannot connect to Keycloak

Verify

```
KEYCLOAK_BASE_URL
```

Verify

```
KEYCLOAK_ADMIN_CLIENT_SECRET
```

---

## Migrations failed

Run

```bash
docker compose up migrate
```

---

## Backend crashes

View

```bash
docker compose logs backend
```

---

## Frontend cannot reach backend

Verify

```
VITE_API_BASE_URL
```

Open browser developer tools.

Check Network tab.

---

# Production Checklist

- HTTPS enabled
- PostgreSQL secured
- Keycloak secured
- Strong passwords
- Secrets outside repository
- Automatic backups
- Monitoring enabled
- Log aggregation configured
- Firewall configured
- Resource limits configured
- Health checks enabled
- SSL certificates renewed automatically

---

# Related Documentation

- [Project README](../README.md)
- [Backend Guide](../backend/README.md)
- [Frontend Guide](../frontend/README.md)
- [Architecture Guide](ARCHITECTURE.md)
