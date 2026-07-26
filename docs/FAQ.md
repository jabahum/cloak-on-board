# Frequently Asked Questions

## Why not configure Keycloak directly?

Keycloak Onboarder standardizes onboarding, reduces manual errors, tracks provisioning, and provides reusable templates.

---

## Does this replace Keycloak?

No. It uses the Keycloak Admin REST API and complements Keycloak.

---

## Can I import existing clients?

Yes. Phase 2 added Keycloak client discovery and import, including core client
configuration and client roles.

---

## Can multiple realms be managed?

Yes. Phase 4 added explicit environment-scoped realm connections. The
`onboarder` authentication realm remains separate from managed target realms.

---

## Does it support mobile applications?

Yes. The platform includes a mobile application template with appropriate defaults.

---

## Can I customize templates?

Yes. Templates can be managed through the platform and extended over time.

---

## Is authentication required?

Production requires Keycloak JWT authentication. Development may explicitly
use API-key mode. Viewer, manager, and administrator roles are enforced by the
backend.

---

## Where are client secrets stored?

Legacy provisioning data remains response-hidden. Phase 4 rotated secrets use
encrypted, recipient-bound, expiring one-time deliveries; ciphertext is
deleted after consumption or expiry. External secret-manager integration is
still in the unassigned backlog.

---

## Can I use this in production?

The project includes production authentication, encrypted managed-realm
credentials, approvals, audit logs, and production container configuration.
Review the deployment and security guides, provide production encryption keys,
use least-privilege realm service accounts, and perform your organization’s
own operational and security review.

For current and planned scope, see the authoritative
[product roadmap](../ROADMAP.md).
