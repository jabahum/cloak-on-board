# API Reference

This document describes the REST API exposed by the Keycloak Onboarder backend.

---

# Base URL

Development

```
http://localhost:9000/api/v1
```

Production

```
https://api.onboarder.example.com/api/v1
```

---

# Content Type

All requests use JSON.

```
Content-Type: application/json
```

---

# Authentication

Currently, the API supports optional bearer token authentication.

```
Authorization: Bearer <token>
```

When `API_AUTH_TOKEN` is not configured, authentication is disabled for development.

Future versions will support Keycloak-issued JWT access tokens.

---

# Common Response Format

## Success

```json
{
  "data": {}
}
```

## Error

```json
{
  "error": "validation failed",
  "request_id": "3a6f79bb-acde-4c11-b6c4-7b85f6e4d8aa"
}
```

---

# HTTP Status Codes

| Code | Meaning               |
| ---- | --------------------- |
| 200  | Success               |
| 201  | Created               |
| 204  | No Content            |
| 400  | Bad Request           |
| 401  | Unauthorized          |
| 404  | Not Found             |
| 409  | Conflict              |
| 422  | Validation Error      |
| 500  | Internal Server Error |

---

# Health

## Get Health

```
GET /health
```

Response

```json
{
  "status": "ok"
}
```

---

# Applications

Applications represent systems that will be onboarded into Keycloak.

---

## List Applications

```
GET /applications
```

Response

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Health BI",
      "slug": "health-bi",
      "app_type": "frontend",
      "status": "provisioned"
    }
  ]
}
```

---

## Get Application

```
GET /applications/{id}
```

Example

```
GET /applications/ef3e4e3f-45aa-44d1
```

---

## Create Application

```
POST /applications
```

Request

```json
{
  "name": "Health BI",
  "slug": "health-bi",
  "description": "Integrated Health BI",
  "app_type": "frontend",
  "owner_name": "HMIS Team",
  "owner_email": "hmis@example.org",
  "redirect_uris": [
    "http://localhost:3000/*"
  ],
  "web_origins": [
    "http://localhost:3000"
  ],
  "roles": [
    "admin",
    "manager",
    "viewer"
  ]
}
```

Response

```json
{
  "data": {
    "id": "uuid"
  }
}
```

---

## Update Application

```
PUT /applications/{id}
```

Updates an existing application.

---

## Delete Application

```
DELETE /applications/{id}
```

Response

```
204 No Content
```

---

## Provision Application

```
POST /applications/{id}/provision
```

Starts a provisioning job.

Response

```json
{
  "data": {
    "id": "job-id",
    "status": "running"
  }
}
```

---

# Templates

Templates provide reusable onboarding defaults.

---

## List Templates

```
GET /templates
```

---

## Get Template

```
GET /templates/{id}
```

---

## Seed Default Templates

```
POST /templates/seed
```

Response

```json
{
  "data": {
    "message": "Templates seeded successfully"
  }
}
```

---

# Settings

Settings define how the platform connects to Keycloak.

---

## Get Settings

```
GET /settings
```

Response

```json
{
  "data": {
    "keycloak_base_url": "http://keycloak:8080",
    "keycloak_realm": "master",
    "keycloak_admin_client_id": "onboarder-admin"
  }
}
```

---

## Save Settings

```
PUT /settings
```

Request

```json
{
  "keycloak_base_url": "http://keycloak:8080",
  "keycloak_realm": "master",
  "keycloak_admin_client_id": "onboarder-admin",
  "keycloak_admin_client_secret": "secret"
}
```

---

# Provisioning Jobs

Every provisioning operation creates a job.

---

## List Jobs

```
GET /jobs
```

Response

```json
{
  "data": [
    {
      "id": "uuid",
      "status": "running"
    }
  ]
}
```

---

## Get Job

```
GET /jobs/{id}
```

Response

```json
{
  "data": {
    "id": "uuid",
    "status": "completed",
    "steps": [
      {
        "step_name": "create_client",
        "status": "completed"
      },
      {
        "step_name": "create_roles",
        "status": "completed"
      }
    ]
  }
}
```

---

# Provisioning Workflow

Provisioning consists of several tracked steps.

1. Authenticate with Keycloak
2. Create Client
3. Configure Client
4. Create Roles
5. Assign Client Scopes
6. Create Protocol Mappers
7. Retrieve Client Secret
8. Persist Metadata
9. Mark Job Complete

Each step is persisted to the database.

---

# Validation Rules

## Applications

| Field         | Rule                         |
| ------------- | ---------------------------- |
| name          | Required                     |
| slug          | Required, unique             |
| app_type      | Required                     |
| redirect_uris | Required for frontend/mobile |
| owner_email   | Valid email                  |

---

# Error Codes

| Error                 | Description                |
| --------------------- | -------------------------- |
| validation_failed     | Invalid request payload    |
| application_not_found | Application does not exist |
| template_not_found    | Template does not exist    |
| provisioning_failed   | Provisioning failed        |
| keycloak_error        | Keycloak returned an error |
| database_error        | Database operation failed  |

---

# Versioning

Current version

```
v1
```

Base path

```
/api/v1
```

Future versions will use:

```
/api/v2
```

without breaking existing clients.

---

# Pagination (Future)

Future list endpoints will support:

```
?page=1

&page_size=20

&sort=name

&order=asc
```

---

# Filtering (Future)

Examples

```
GET /applications?status=provisioned

GET /applications?app_type=frontend

GET /jobs?status=running
```

---

# OpenAPI

Future versions of the platform will expose an OpenAPI (Swagger) specification.

```
/swagger/index.html

/openapi.json
```

---

# Related Documentation

- [Project README](../README.md)
- [Backend Guide](../backend/README.md)
- [Frontend Guide](../frontend/README.md)
- [Architecture Guide](ARCHITECTURE.md)
- [Deployment Guide](DEPLOYMENT.md)