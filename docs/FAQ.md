# Frequently Asked Questions

## Why not configure Keycloak directly?

Keycloak Onboarder standardizes onboarding, reduces manual errors, tracks provisioning, and provides reusable templates.

---

## Does this replace Keycloak?

No. It uses the Keycloak Admin REST API and complements Keycloak.

---

## Can I import existing clients?

Not yet. This is planned for a future release.

---

## Can multiple realms be managed?

Not in the MVP. Multi-realm support is on the roadmap.

---

## Does it support mobile applications?

Yes. The platform includes a mobile application template with appropriate defaults.

---

## Can I customize templates?

Yes. Templates can be managed through the platform and extended over time.

---

## Is authentication required?

Authentication is optional in the MVP. Future releases will integrate with Keycloak using JWTs and RBAC.

---

## Where are client secrets stored?

Client secrets are stored in the platform database after provisioning confidential clients. In production, you should consider integrating with a dedicated secrets management solution.

---

## Can I use this in production?

The project is designed with production in mind, but you should review the deployment guide and security recommendations before deploying in a production environment.