# Contributing to Keycloak Onboarder

Thank you for your interest in contributing to Keycloak Onboarder!

We welcome bug reports, feature requests, documentation improvements, and code contributions.

---

# Getting Started

Before contributing, please read:

- README.md
- docs/DEVELOPMENT.md
- docs/ARCHITECTURE.md

Set up the development environment.

```bash
git clone https://github.com/<organization>/keycloak-onboarder.git

cd keycloak-onboarder

docker compose up --build
```

---

# Branching Strategy

Create feature branches from `main`.

Examples

```
feature/client-scopes

feature/template-import

feature/notifications

bugfix/cors

hotfix/job-status
```

---

# Commit Messages

Use Conventional Commits.

Examples

```
feat: add client scope provisioning

fix: resolve CORS issue

docs: update deployment guide

refactor: simplify provisioning service

test: add application service tests
```

---

# Pull Requests

Before opening a pull request:

- Ensure the project builds successfully.
- Run tests.
- Update documentation where applicable.
- Include database migrations if required.
- Keep pull requests focused on a single feature or fix.

---

# Coding Standards

## Backend

- Use `go fmt`.
- Keep handlers thin.
- Place business logic in services.
- Place persistence logic in repositories.
- Return consistent API responses.

## Frontend

- Use TypeScript.
- Avoid `any`.
- Reuse shared components.
- Centralize API calls.
- Keep pages focused.

---

# Reporting Bugs

Please include:

- Description
- Steps to reproduce
- Expected behavior
- Actual behavior
- Logs
- Screenshots (if applicable)

---

# Feature Requests

Describe:

- Problem
- Proposed solution
- Alternatives considered
- Additional context

---

# Code Reviews

All contributions require review before merging.

Reviewers will evaluate:

- Readability
- Maintainability
- Tests
- Documentation
- Security

---

Thank you for contributing!