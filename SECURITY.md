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
- Managed-realm administrator secrets are encrypted with AES-256-GCM. The
  realm-connection UUID is authenticated as additional data.
- `CREDENTIAL_ENCRYPTION_KEYS` is a comma-separated, newest-first keyring in
  `version:base64` form. Each decoded key must be exactly 32 bytes. Retain old
  versions until every stored credential has been re-encrypted.
- Production refuses to start without encryption keys. Legacy settings
  credentials are migrated once and their plaintext column is cleared.
- Rotated client secrets are encrypted at rest, expire after
  `SECRET_DELIVERY_TTL_MINUTES`, are recipient-bound, and are atomically
  consumed. The ciphertext is deleted when consumed.
- Secret-delivery responses use `Cache-Control: no-store` and `Pragma:
  no-cache`. Expired or consumed deliveries return `410`.
- Rotation is never automatically retried because a lost successful response
  would make a second rotation unsafe.

## Realm separation

The realm configured by `KEYCLOAK_REALM` authenticates cloak-on-board users
only. Every managed Keycloak operation resolves an enabled realm connection
from its deployment; an application-level UUID is never reused across realms.
Use distinct least-privilege service-account clients for every managed realm.

The users and credentials in the development realm import are examples for
local development only. Production users must be provisioned separately and
assigned `admin`, `manager`, or `viewer` through Keycloak groups or realm-role
mappings.

## Audit and approvals

All authenticated mutations create append-only audit records with actor,
request ID, result, and request metadata. Manager-requested provisioning,
linked-client updates, and Keycloak deletion require an administrator decision.
Requesters cannot approve their own requests.
