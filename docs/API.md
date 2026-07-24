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

Updates local metadata, URLs and roles. When the application is linked to a
Keycloak client, supported client settings and roles are synchronized through
a tracked `update_application` job.

Request

```json
{
  "name": "Health BI",
  "slug": "health-bi",
  "description": "Integrated Health BI",
  "app_type": "frontend",
  "owner_name": "HMIS Team",
  "owner_email": "hmis@example.org",
  "redirect_uris": ["https://health.example.org/*"],
  "web_origins": ["https://health.example.org"],
  "roles": ["admin", "viewer"],
  "enabled": true
}
```

---

## Delete Application

```
DELETE /applications/{id}?delete_keycloak=false
```

`delete_keycloak=false` removes only the local application. Set it explicitly
to `true` to delete the linked client by its stored Keycloak UUID before
removing the local application. A failed Keycloak deletion leaves the local
record intact. Missing Keycloak clients are treated as already deleted.

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

# Keycloak Client Import

Client secrets are never included in discovery or import responses.

## Search Realm Clients

```
GET /keycloak/clients?search=health
```

Each result contains its Keycloak UUID, client ID, display metadata and core
flow settings. `imported` is true when the UUID is already linked locally.

## Import Client

```
POST /applications/import
```

```json
{
  "keycloak_client_uuid": "5c387c42-ef66-4daf-8b72-bf617eb8e839",
  "name": "Health BI",
  "description": "Existing realm client",
  "app_type": "frontend",
  "owner_name": "HMIS Team",
  "owner_email": "hmis@example.org"
}
```

The server fetches the complete client representation and roles from
Keycloak, infers `app_type` when omitted, and creates the linked local record
transactionally. Duplicate slugs, client IDs and Keycloak UUIDs return `409`.

---

# Client Scopes

## List Assignments

```
GET /applications/{id}/client-scopes
```

Returns `default`, `optional`, and `available` realm client-scope arrays.

## Assign or Change Assignment

```
PUT /applications/{id}/client-scopes/{scopeId}
```

```json
{
  "type": "default"
}
```

`type` must be `default` or `optional`. Assignment is idempotent and moving a
scope removes its previous assignment type.

## Remove Assignment

```
DELETE /applications/{id}/client-scopes/{scopeId}?type=default
```

Returns `204 No Content`. `type` must be `default` or `optional`.

---

# Protocol Mappers

Supported mapper types are OIDC user-attribute, client-role and realm-role
mappers.

## List Mappers

```
GET /applications/{id}/protocol-mappers
```

## Create Mapper

```
POST /applications/{id}/protocol-mappers
```

```json
{
  "name": "department",
  "protocol": "openid-connect",
  "protocolMapper": "oidc-usermodel-attribute-mapper",
  "config": {
    "user.attribute": "department",
    "claim.name": "department",
    "jsonType.label": "String",
    "id.token.claim": "true",
    "access.token.claim": "true",
    "userinfo.token.claim": "true"
  }
}
```

## Update Mapper

```
PUT /applications/{id}/protocol-mappers/{mapperId}
```

The request body uses the same representation as create.

## Delete Mapper

```
DELETE /applications/{id}/protocol-mappers/{mapperId}
```

Returns `204 No Content`.

All scope and mapper endpoints return `409` when the local application is not
linked to a Keycloak client.

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
