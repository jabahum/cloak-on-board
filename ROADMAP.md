# Product Roadmap

This file is the source of truth for cloak-on-board feature status.

Status legend:

- `[x]` Implemented and available in the repository.
- `[~]` Partially implemented; the remaining scope is stated explicitly.
- `[ ]` Not implemented.

## Release summary

| Release | Status | Outcome |
| --- | --- | --- |
| Phase 1 — MVP/Foundation | Complete | Application registry, templates, provisioning, Keycloak integration, jobs, and the initial UI |
| Phase 2 — Client Lifecycle | Complete | Import, update, delete, client scopes, and protocol mappers |
| Phase 3 — Security and Governance | Complete | Keycloak login, RBAC, audit logs, notifications, and approvals |
| Phase 4 — Delivery Platform | Implemented with follow-ups | Multi-realm environments, promotion, drift, rotation, OpenAPI, and SDK; remaining gaps are listed below |
| Phase 5 — Automation Ecosystem | Not started | Plugins, infrastructure-as-code, operators, GitOps, and CI/CD |

---

## Phase 1 — MVP/Foundation

### Project and runtime foundation

- [x] Go backend using Gin and pgx.
- [x] React and TypeScript frontend using Vite and Carbon.
- [x] PostgreSQL persistence.
- [x] Docker development and production builds.
- [x] Docker Compose services for the API, UI, PostgreSQL, Keycloak, and migrations.
- [x] Versioned forward and rollback SQL migrations.
- [x] Health endpoint at `GET /api/v1/health`.
- [x] Environment-based backend and frontend configuration.

### Application registry

- [x] Create applications.
- [x] List applications.
- [x] View application details.
- [x] Edit application metadata and configuration.
- [x] Store application name, slug/client ID, description, type, owner, and enabled state.
- [x] Store redirect URIs, web origins, and client roles.
- [x] Track draft, approval, provisioning, success, and failure states.
- [x] Application creation wizard and detail UI.
- [x] Application type support for frontend, mobile, backend, and machine-to-machine clients.

### Templates

- [x] Built-in templates for React SPA, backend API, mobile, and machine-to-machine clients.
- [x] Template listing and detail APIs.
- [x] Default-template seeding.
- [x] Template selection in the application wizard.
- [ ] Custom template authoring, editing, and deletion. This is tracked in the unassigned backlog.

### Initial Keycloak provisioning

- [x] Configurable Keycloak URL, realm, administrative client ID, and secret.
- [x] Service-account authentication to the Keycloak Admin REST API.
- [x] Public-client provisioning.
- [x] Confidential-client provisioning.
- [x] PKCE configuration for browser and mobile clients.
- [x] Redirect URI and web-origin configuration.
- [x] Client-role creation.
- [x] Default protocol-mapper creation.
- [x] Confidential-client secret retrieval during legacy provisioning.
- [x] Persist the linked Keycloak client UUID and client ID for the original single-realm workflow.
- [x] Idempotent create-or-get behavior for clients and roles.

### Provisioning jobs

- [x] Create a job for provisioning operations.
- [x] Track pending, running, succeeded, and failed states.
- [x] Track start and completion timestamps.
- [x] Track individual provisioning steps.
- [x] Record sanitized execution errors.
- [x] Jobs API and job-monitoring UI.

### HTTP and operational middleware

- [x] Request IDs.
- [x] Structured request logging.
- [x] Panic recovery.
- [x] CORS enforcement.
- [x] Consistent success and error envelopes.
- [x] Request IDs on API errors.
- [x] Development API-key authentication mode.

### MVP completion note

The original MVP is complete. Features previously listed in the README as
“planned”—automatic migrations, client scopes, import, approvals, audit,
multi-realm support, drift, and rotation—were delivered in Phases 2–4 below.

---

## Phase 2 — Client Lifecycle

### Existing-client import

- [x] Search clients in Keycloak.
- [x] Import an existing client into the application registry.
- [x] Preserve the Keycloak UUID rather than treating client ID as a realm-global identity.
- [x] Import core client settings, URLs, flows, enabled state, and roles.
- [x] Prevent duplicate local management of the same client.
- [x] Import UI.

### Client updates and deletion

- [x] Update local application configuration.
- [x] Synchronize linked Keycloak client metadata, URLs, flows, enabled state, and roles.
- [x] Delete only the local application record.
- [x] Optionally delete the linked Keycloak client.
- [x] Confirmation UI for destructive Keycloak deletion.
- [x] Keep the local record if an upstream deletion fails.
- [x] Treat an already-missing Keycloak client as deleted.

### Client scopes

- [x] List default, optional, and available client scopes.
- [x] Assign a default client scope.
- [x] Assign an optional client scope.
- [x] Change a scope’s assignment type.
- [x] Remove an assigned scope.
- [x] Client-scope management UI.

### Protocol mappers

- [x] List protocol mappers.
- [x] Create protocol mappers.
- [x] Update protocol mappers.
- [x] Delete protocol mappers.
- [x] Default user-attribute, client-role, and realm-role mapper helpers.
- [x] Protocol-mapper management UI.

---

## Phase 3 — Security and Governance

### User authentication

- [x] Keycloak authorization-code login with PKCE.
- [x] Cached single initialization of the Keycloak browser adapter.
- [x] JWT signature verification using Keycloak JWKS.
- [x] Issuer, audience, and token-lifetime validation.
- [x] Current-user endpoint at `GET /api/v1/auth/me`.
- [x] Production enforcement of Keycloak authentication.
- [x] Logout and session-expiry behavior.
- [x] Authentication realm remains a control-plane concern, separate from Phase 4 managed realms.

### Role-based access control

- [x] Viewer, manager, and administrator effective roles.
- [x] Backend permission middleware as the authoritative security boundary.
- [x] UI permission guards and conditional navigation.
- [x] Permissions for reads, drafts, approvals, client administration, settings, audit, environments, realm connections, promotion, drift, reconciliation, and rotation.

### Audit logs

- [x] Append-only audit records for authenticated mutations.
- [x] Actor subject, username, email, and effective role.
- [x] Action, resource, application, request ID, result, status, IP address, and user agent.
- [x] Recursive redaction of secrets, passwords, tokens, and authorization data.
- [x] Filtering and pagination.
- [x] Administrator-only audit API and UI.

### Notifications

- [x] Durable per-user notifications.
- [x] Unread counts.
- [x] Mark one or all notifications as read.
- [x] Recipient isolation by Keycloak subject.
- [x] Deduplication keys for retryable workflows.
- [x] Approval, promotion, drift, reconciliation, and rotation lifecycle notifications.
- [x] Notification UI.

### Approval workflows

- [x] One approval system shared by legacy provisioning and Phase 4 workflows.
- [x] Submit, list, inspect, approve, reject, cancel, and retry actions.
- [x] Immutable request payload and application version at submission.
- [x] Stale application-version rejection.
- [x] Requester self-approval prevention.
- [x] Independent administrator review.
- [x] Execution job and sanitized failure tracking.
- [x] Unsafe secret-rotation retries explicitly prohibited.
- [x] Approval review UI with request payload and decision comments.

---

## Phase 4 — Delivery Platform

### Multi-environment and multi-realm model

- [x] Ordered development, staging, and protected production environments.
- [x] Environment create, read, update, and guarded delete APIs.
- [x] Explicit realm connections assigned to environments.
- [x] Realm-connection create, read, update, connection-test, and disable APIs.
- [x] Disabled and mismatched connections rejected for managed operations.
- [x] Realm credentials never returned by APIs.
- [x] AES-256-GCM credential encryption.
- [x] Versioned encryption keyring through `CREDENTIAL_ENCRYPTION_KEYS`.
- [x] Realm-connection UUID used as authenticated additional data.
- [x] Production startup fails for missing or invalid encryption material.
- [x] Legacy settings credential migrated once and its plaintext value cleared.
- [x] Authentication realm configuration remains independent from managed target realms.
- [x] Development imports for separate development, staging, and production-like target realms.
- [~] Environment and realm administration UI. Listing, creation, and
  connection testing are available; edit, delete, and disable controls still
  need to be exposed in the UI.

### Immutable snapshots and authoritative deployments

- [x] Immutable application snapshots with a database update-prevention trigger.
- [x] Canonical JSON configuration and SHA-256 configuration hashes.
- [x] Snapshot versions scoped to an application.
- [x] Duplicate configuration snapshots resolve idempotently.
- [~] Deployment records are authoritative for Phase 4 environment workflows.
  The original single-realm provisioning endpoints still retain
  application-level Keycloak UUID fields for backward compatibility.
- [x] The same client ID can be managed in multiple realms with different UUIDs.
- [x] Legacy single-realm application linkage migrates into a default deployment.
- [x] Controlled overrides limited to redirect URIs, web origins, and enabled state.
- [x] Application Deployments tab and global deployment matrix.

### Deployment, promotion, and rollback

- [x] Deploy a snapshot to an explicit environment and realm connection.
- [x] Promote the exact immutable source snapshot.
- [x] Enforce environment order.
- [x] Prevent skipping the immediately preceding deployed environment.
- [x] Reject promotion of a stale source snapshot.
- [x] Require approval for protected environments.
- [x] Preserve source deployments when a destination operation fails.
- [x] Idempotent Keycloak client, role, scope, and mapper synchronization.
- [x] Roll back using the previous deployed snapshot.
- [x] Promotion, rollback, and deployment jobs carry environment, connection, deployment, and snapshot identifiers.
- [x] Promotion review uses the existing Phase 3 approval workflow.

### Configuration drift

- [x] Read-only manual drift checks.
- [x] Optional scheduled drift checks.
- [x] PostgreSQL advisory lock prevents multiple schedulers from running concurrently.
- [x] Bounded scheduler concurrency.
- [x] Drift runs and path-level findings.
- [x] `unknown`, `checking`, `in_sync`, `drifted`, and `error` lifecycle states.
- [x] Canonical ordering for lists and managed objects.
- [x] Comparison of managed client metadata, flow/type, enabled state, URLs, origins, roles, client scopes, and protocol mappers.
- [x] Ignore unmanaged Keycloak fields and unmanaged scope/mapper entries.
- [~] Missing and changed findings are recorded. Explicit `extra` findings for
  individual list members are not yet emitted; list differences currently
  appear as a changed managed field.
- [x] Target outage handling without leaking upstream responses.
- [x] Approval-required reconciliation.
- [x] Post-reconciliation verification.
- [x] Drift-detected, check-failed, and resolved notifications.
- [x] Drift history and check UI.

### Secret rotation and delivery

- [x] Explicit deployment- and realm-specific Keycloak secret rotation.
- [x] Confidential-client validation.
- [x] Public clients rejected.
- [x] Approval action `rotate_client_secret`.
- [x] Rotation is a single-attempt mutation with no automatic retry.
- [x] AES-256-GCM encryption for temporary secret deliveries.
- [x] Recipient-bound delivery.
- [x] Configurable expiry.
- [x] Background expiry cleanup and notification.
- [x] Atomic one-time consumption.
- [x] `404` for another recipient and `410` for consumed or expired deliveries.
- [x] Ciphertext deletion after consumption or expiry.
- [x] `Cache-Control: no-store` and `Pragma: no-cache`.
- [x] Rotated values excluded from applications, jobs, approvals, notifications, audit logs, and errors.
- [x] Rotation request UI and reveal-once notification UI.

### OpenAPI and TypeScript SDK

- [~] Committed OpenAPI 3.1 contract at `openapi/openapi.yaml`. It covers the
  primary application and Phase 4 APIs, but not every older Phase 1–3 endpoint.
- [x] Validated OpenAPI structure.
- [x] ESM TypeScript package at `sdk/typescript`.
- [x] Public package exports and declaration files.
- [x] Async token provider.
- [x] Request-ID propagation.
- [x] Typed `ApiError`.
- [x] `AbortSignal` support.
- [x] Pagination query support.
- [x] Retry policy limited to safe reads.
- [x] SDK methods for applications, templates, settings, scopes, mappers, environments, realm connections, deployments, snapshots, promotions, drift, approvals, jobs, notifications, audit logs, and one-time delivery.
- [x] SDK README and usage example.
- [x] Mocked HTTP unit tests.
- [x] Typecheck, test, build, and package-export scripts.
- [x] SDK is handwritten; no generated artifacts or stale-generation check are required.

### Remaining Phase 4 follow-ups

- [ ] Expand the OpenAPI document to describe every Phase 1–3 route, request,
  response, permission, and error—not only the primary and Phase 4 surface.
- [ ] Emit granular `extra` drift findings for surplus managed roles, scopes,
  mappers, URLs, and origins instead of reporting only a changed collection.
- [ ] Route the Phase 2 client-scope and protocol-mapper editing APIs through an
  explicit deployment/realm connection. They currently operate through the
  backward-compatible single-realm settings path.
- [ ] Remove the legacy application-level Keycloak UUID and secret columns after
  all original provisioning and Phase 2 operations have migrated to
  deployment-scoped storage.
- [ ] Add environment edit/delete and realm-connection edit/disable controls to
  the frontend. The backend and SDK operations already exist.
- [ ] Add dedicated integration tests for scheduler lock contention, delivery
  expiry timing, wrong-recipient retrieval, rollback failure, and disabled
  connection rejection. Core behavior exists, while the current automated
  suite and live checks do not cover every edge case.

### Phase 4 verification completed

- [x] Backend formatting, unit tests, and `go vet`.
- [x] Frontend lint and production build using Node 22.
- [x] SDK installation, typecheck, tests, build, and package export validation.
- [x] OpenAPI validation.
- [x] Migration up, rollback, and reapplication on disposable PostgreSQL databases.
- [x] Separate authentication, development, staging, and production-like realms.
- [x] Same client ID deployed into multiple realms with distinct UUIDs.
- [x] Immutable snapshot promoted from development to staging.
- [x] Protected production-like promotion approved by a different user.
- [x] Deliberate out-of-band change detected and reconciled through approval.
- [x] Confidential-client secret rotated and consumed exactly once.
- [x] No-cache headers, ciphertext deletion, and absence of secret leakage verified.
- [x] Temporary test clients, applications, snapshots, deliveries, approvals, jobs, tokens, and relaxed security settings cleaned up.

---

## Phase 5 — Automation Ecosystem

Phase 5 has not started.

### Plugin marketplace

- [ ] Plugin manifest and compatibility model.
- [ ] Plugin discovery and installation.
- [ ] Trust, signing, permissions, and sandbox policy.
- [ ] Marketplace administration UI.
- [ ] Plugin lifecycle and version upgrades.

### Terraform export

- [ ] Export application snapshots as Terraform configuration.
- [ ] Export realm-specific deployment settings without secrets.
- [ ] Stable resource naming and dependency ordering.
- [ ] Drift-safe import guidance.

### Kubernetes operator

- [ ] Custom resource definitions for applications and deployments.
- [ ] Reconciliation controller.
- [ ] Status conditions and events.
- [ ] Kubernetes Secret integration.
- [ ] Upgrade and rollback strategy.

### GitOps integration

- [ ] Repository-backed desired configuration.
- [ ] Pull-request promotion workflow.
- [ ] Signed commit and policy validation.
- [ ] Reconciliation status reporting.
- [ ] Conflict handling between UI and Git-managed resources.

### CI/CD templates

- [ ] Reusable validation workflow.
- [ ] Snapshot creation and promotion workflow.
- [ ] Approval-aware protected-environment workflow.
- [ ] SDK examples for common CI providers.
- [ ] Release and rollback examples.

---

## Unassigned backlog

These items are not implemented and do not currently have a committed phase:

- [ ] Developer self-service portal distinct from the administrator UI.
- [ ] Custom template create, update, delete, import, and export.
- [ ] External secret-manager integrations such as Vault or cloud secret stores.
- [ ] Keycloak group management.
- [ ] Realm-role and identity-governance management.
- [ ] Multi-tenant isolation above the current multi-realm model.
- [ ] General configuration export outside the planned Terraform exporter.
- [ ] Generated documentation portal.
- [ ] Dedicated metrics, alerting, and health-monitoring subsystem.
- [ ] Service decomposition; the current system remains a modular monolith.

Items move from this backlog into a numbered phase only when scope and
acceptance criteria have been agreed.
