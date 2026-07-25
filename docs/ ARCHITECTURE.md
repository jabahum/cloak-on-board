# Architecture Guide

This document describes the overall architecture of the Keycloak Onboarder platform, its major components, data flow, and design decisions.

---

# Table of Contents

- Overview
- Design Principles
- High-Level Architecture
- System Components
- Backend Architecture
- Frontend Architecture
- Database Architecture
- Keycloak Integration
- Provisioning Workflow
- Sequence Diagrams
- Module Responsibilities
- Data Model
- Security Architecture
- Deployment Architecture
- Future Architecture

---

# Overview

Keycloak Onboarder is a platform for automating the onboarding of applications into Keycloak.

Instead of manually configuring Keycloak through the Admin Console, administrators register an application once and the platform automatically provisions:

- Clients
- Roles
- Client Scopes
- Protocol Mappers
- Redirect URIs
- Web Origins
- Client Secrets

## Phase 4 delivery architecture

The authentication realm is a control-plane dependency and is never selected
as an implicit deployment target. `environments` define strict promotion
order; `realm_connections` bind one managed realm to an environment;
`application_deployments` hold the realm-specific client UUID and current
snapshot.

Snapshots are immutable canonical JSON documents. Only `redirect_uris`,
`web_origins`, and `enabled` may be overridden per environment. Promotion
copies a snapshot, never mutable application state, and requires the
immediately preceding deployed environment. Protected environments use the
existing approval system.

Drift checks read the target client and compare only managed fields:
identity, display metadata, client type/flows, enabled state, redirect URIs,
web origins, and client roles. Lists are sorted and Keycloak defaults are
normalized before hashing. Reconciliation reapplies the snapshot and performs
a post-write verification.

Realm credentials and temporary secret deliveries use AES-256-GCM with
versioned keys. Connection IDs and delivery identity are authenticated as
additional data. Rotation is a single-attempt mutation and delivery
ciphertext is deleted atomically on first consumption.

```text
Authentication realm ──JWT──> API
                              │
Application ──> Snapshot ──> Deployment ──> Explicit realm connection
                              │
                              ├── Drift run/findings
                              └── Approval ──> promotion/reconcile/rotation
```

---

# Design Principles

The platform is designed around the following principles:

- Simplicity
- Idempotent provisioning
- Modular architecture
- Clear separation of concerns
- Stateless REST APIs
- Template-driven provisioning
- Extensible integrations
- Production readiness

---

# High-Level Architecture

```text
                    +------------------------+
                    |      React Frontend    |
                    +-----------+------------+
                                |
                                |
                                ▼
                    +------------------------+
                    |      Backend API       |
                    |         (Go)           |
                    +-----------+------------+
                                |
        +-----------------------+-----------------------+
        |                       |                       |
        ▼                       ▼                       ▼
+---------------+      +----------------+      +----------------+
| PostgreSQL    |      |   Keycloak     |      | Future Plugins |
| Metadata      |      | Admin REST API |      | Notifications  |
+---------------+      +----------------+      +----------------+
```

---

# System Components

## Frontend

Responsibilities

- User Interface
- Application Registration
- Provisioning
- Settings
- Job Monitoring

Technology

- React
- TypeScript
- Carbon Design System

---

## Backend

Responsibilities

- REST API
- Business Logic
- Provisioning Engine
- Keycloak Integration
- Job Tracking

Technology

- Go
- Gin
- PostgreSQL

---

## Database

Stores

- Applications
- Templates
- Jobs
- Job Steps
- Settings

---

## Keycloak

Provides

- Identity
- Authentication
- Authorization

The platform communicates with Keycloak using the Admin REST API.

---

# Backend Architecture

```text
HTTP Request
      │
      ▼
Gin Router
      │
      ▼
Middleware
      │
      ▼
Handler
      │
      ▼
Service
      │
      ▼
Repository
      │
      ▼
PostgreSQL
```

Each layer has a single responsibility.

---

# Backend Modules

```text
internal/

applications/

templates/

settings/

provisioning/

keycloak/

middleware/

database/

server/

response/
```

---

# Provisioning Engine

The provisioning engine orchestrates all Keycloak operations.

Workflow

```text
Application

↓

Load Settings

↓

Authenticate

↓

Create Client

↓

Create Roles

↓

Assign Client Scopes

↓

Create Protocol Mappers

↓

Retrieve Secret

↓

Persist Metadata

↓

Complete Job
```

Each provisioning step is independently tracked.

---

# Frontend Architecture

```text
React

↓

Pages

↓

Components

↓

API Layer

↓

Backend API
```

---

# Frontend Structure

```text
src/

api/

components/

pages/

layouts/

styles/

types/

router/
```

The frontend does **not** communicate directly with Keycloak.

---

# Database Architecture

```text
Applications
      │
      │
      ├──────── Jobs
      │             │
      │             └──────── Job Steps
      │
      ├──────── Templates
      │
      └──────── Settings
```

---

# Core Entities

## Applications

Stores

- Name
- Slug
- Type
- Redirect URIs
- Web Origins
- Roles
- Provisioning Status

---

## Templates

Stores reusable onboarding configurations.

---

## Settings

Stores platform configuration.

---

## Jobs

Tracks provisioning execution.

---

## Job Steps

Tracks individual provisioning operations.

---

# Keycloak Integration

The backend communicates exclusively through the Keycloak Admin REST API.

Supported operations

- Create Client
- Update Client
- Delete Client
- Create Roles
- Create Scopes
- Create Protocol Mappers
- Retrieve Client Secret

Future operations

- Import Existing Clients
- Synchronize Configuration
- Rotate Secrets

---

# Provisioning Workflow

```text
User

↓

Create Application

↓

Save Metadata

↓

Provision

↓

Provisioner

↓

Authenticate to Keycloak

↓

Create Client

↓

Configure Client

↓

Create Roles

↓

Assign Scopes

↓

Create Mappers

↓

Retrieve Secret

↓

Update Database

↓

Success
```

---

# Sequence Diagram

```text
User
 │
 │ Create Application
 ▼
Frontend
 │
 │ POST /applications
 ▼
Backend
 │
 │ Save
 ▼
Database
 │
 │
 ▼

User
 │
 │ Provision
 ▼
Backend
 │
 │
 ▼
Provisioner
 │
 │ Authenticate
 ▼
Keycloak
 │
 │ OK
 ▼
Provisioner
 │
 │ Create Client
 ▼
Keycloak
 │
 │ Client Created
 ▼
Provisioner
 │
 │ Save Metadata
 ▼
Database
 │
 ▼
Frontend
```

---

# Middleware

Every request passes through

```text
Request ID

↓

Logger

↓

Recovery

↓

CORS

↓

Authentication

↓

Handler
```

---

# Security Architecture

Authentication

Current

- API Token

Future

- Keycloak JWT

Authorization

Future

- RBAC

Secrets

- Stored securely
- Never exposed in logs

---

# Deployment Architecture

Development

```text
Docker Compose

↓

Frontend

↓

Backend

↓

PostgreSQL

↓

Keycloak
```

Production

```text
Internet

↓

Nginx

↓

Frontend

↓

Backend

↓

PostgreSQL

↓

Keycloak
```

---

# Extension Points

The platform is designed for future expansion.

Planned modules

- Notification Service
- Approval Workflow
- Secret Rotation
- Configuration Drift Detection
- Audit Logs
- Multi-Realm Support
- Multi-Environment Promotion
- Plugin SDK

---

# Future Microservice Architecture

Current

```text
Frontend

↓

Backend

↓

PostgreSQL
```

Future

```text
                    API Gateway
                          │
     ┌──────────────┬──────────────┬──────────────┐
     ▼              ▼              ▼
Applications   Provisioning   Notifications
Service         Service         Service
     │              │              │
     └──────────────┴──────────────┘
                    │
                    ▼
                PostgreSQL

                    │

                    ▼

                Keycloak
```

The current monolith has been intentionally structured so that each module can later be extracted into an independent service if needed.

---

# Phase 3 Security Architecture

Keycloak authenticates browser users with authorization code flow and PKCE.
The API validates RS256 access tokens against cached JWKS, issuer, lifetime,
and audience before resolving an effective realm role.

```text
Keycloak login
      │
      ▼
JWT authentication → permission middleware → handler/service
                                             │
                         ┌───────────────────┼───────────────────┐
                         ▼                   ▼                   ▼
                    Audit log          Approval request     Notification
                                             │
                                      Admin decision
                                             │
                                             ▼
                                  Existing provisioning engine
```

Approval requests store immutable proposed payloads and the application
configuration version. Only an accepted, non-stale request reaches the existing
Keycloak mutation services. Audit records are append-only, and notifications
are isolated by Keycloak subject.

---

# Design Decisions

## Why Go?

- Fast
- Simple
- Excellent concurrency
- Small binaries
- Great Docker support

## Why React?

- Mature ecosystem
- TypeScript support
- Component-based
- Carbon Design System compatibility

## Why PostgreSQL?

- Reliable
- JSON support
- Strong transactional guarantees
- Excellent performance

## Why Keycloak?

- Open Source
- OpenID Connect
- OAuth2
- SAML
- Enterprise-ready

---

# Related Documentation

- [Project README](../README.md)
- [Backend Guide](../backend/README.md)
- [Frontend Guide](../frontend/README.md)
- [Deployment Guide](DEPLOYMENT.md)
- [API Reference](API.md)
