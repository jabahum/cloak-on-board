# Security Policy

## Supported Versions

| Version | Supported |
| ------- | --------- |
| 0.1.x   | ✅         |
| <0.1    | ❌         |

---

# Reporting a Vulnerability

Please do **not** create public GitHub issues for security vulnerabilities.

Instead:

- Email the maintainers.
- Include detailed reproduction steps.
- Include affected versions.
- Include proof-of-concept where possible.

We will acknowledge reports within 72 hours.

---

# Security Best Practices

- Enable HTTPS
- Use strong PostgreSQL passwords
- Use confidential Keycloak clients
- Restrict service account permissions
- Rotate secrets regularly
- Enable backups
- Restrict CORS origins
- Keep dependencies updated

## Authentication and authorization

Production deployments must use `AUTH_MODE=keycloak`. The API validates
RS256 signatures, issuer, lifetime, and the configured `KEYCLOAK_AUDIENCE`.
API-key mode is restricted to development and test environments.

Realm roles map to these effective roles:

| Role | Access |
| --- | --- |
| viewer | Read applications, templates, jobs, profile, and personal notifications |
| manager | Viewer access plus drafts, imports, and approval submission |
| admin | Approval review, Keycloak mutations, settings, templates, and audit logs |

Backend permission middleware is authoritative. UI visibility is not a
security boundary.

## Secret handling

- Application list and detail responses never include client secrets.
- Settings responses never include the Keycloak admin client secret.
- Audit data recursively redacts secrets, passwords, tokens, and authorization headers.
- Tokens and credentials must never be added to notifications or error messages.

The users and credentials in the development realm import are examples for
local development only. Production users must be provisioned separately and
assigned `admin`, `manager`, or `viewer` through Keycloak groups or realm-role
mappings.

## Audit and approvals

All authenticated mutations create append-only audit records with actor,
request ID, result, and request metadata. Manager-requested provisioning,
linked-client updates, and Keycloak deletion require an administrator decision.
Requesters cannot approve their own requests.
