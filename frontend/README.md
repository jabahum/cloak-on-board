# Frontend

The frontend provides the user interface for the Keycloak Onboarder platform. It enables administrators and developers to register applications, manage onboarding templates, configure Keycloak settings, provision clients, and monitor provisioning jobs.

---

# Features

- Dashboard
- Application Management
- Keycloak Provisioning
- Provisioning Jobs
- Template Management
- Platform Settings
- Responsive UI
- Carbon Design System
- REST API Integration
- Type-safe API Models

---

# Technology Stack

| Component    | Technology           |
| ------------ | -------------------- |
| Framework    | React 19             |
| Language     | TypeScript           |
| Build Tool   | Vite                 |
| UI Framework | Carbon Design System |
| HTTP Client  | Axios                |
| Routing      | React Router         |
| Icons        | Carbon Icons         |

---

# Directory Structure

```text
frontend/

├── public/
│
├── src/
│
│   ├── api/
│   │   ├── applications.ts
│   │   ├── provisioning.ts
│   │   ├── settings.ts
│   │   ├── templates.ts
│   │   └── client.ts
│   │
│   ├── assets/
│   │
│   ├── components/
│   │
│   ├── hooks/
│   │
│   ├── layouts/
│   │
│   ├── pages/
│   │   ├── applications/
│   │   ├── dashboard/
│   │   ├── jobs/
│   │   ├── settings/
│   │   └── templates/
│   │
│   ├── router/
│   │
│   ├── styles/
│   │
│   ├── types/
│   │
│   ├── App.tsx
│   └── main.tsx
│
├── Dockerfile
├── package.json
├── vite.config.ts
└── tsconfig.json
```

---

# Installing Dependencies

```bash
npm install
```

or

```bash
pnpm install
```

---

# Development

Start the development server.

```bash
npm run dev
```

Default URL

```
http://localhost:3000
```

---

# Production Build

Generate a production build.

```bash
npm run build
```

Preview the production build.

```bash
npm run preview
```

---

# Environment Variables

Create a `.env` file.

```env
VITE_API_BASE_URL=http://localhost:9000/api/v1
```

Example production configuration.

```env
VITE_API_BASE_URL=https://api.onboarder.example.com/api/v1
```

---

# Docker

Build the frontend.

```bash
docker compose build frontend
```

Start the frontend.

```bash
docker compose up frontend
```

View logs.

```bash
docker compose logs -f frontend
```

---

# Routing

The application currently exposes the following routes.

| Route             | Description         |
| ----------------- | ------------------- |
| /                 | Dashboard           |
| /applications     | List applications   |
| /applications/new | Create application  |
| /applications/:id | Application details |
| /templates        | Templates           |
| /jobs             | Provisioning jobs   |
| /settings         | Platform settings   |

---

# Pages

## Dashboard

Displays:

- Platform overview
- Provisioning statistics
- Recent activity

---

## Applications

Manage applications.

Features

- List
- Create
- View
- Delete
- Provision

---

## Templates

Displays onboarding templates.

Templates include

- React SPA
- Backend API
- Flutter Mobile
- Machine-to-Machine

---

## Provisioning Jobs

Displays

- Running jobs
- Failed jobs
- Completed jobs

Each job displays

- Status
- Steps
- Errors

---

## Settings

Configure

- Keycloak URL
- Realm
- Admin Client
- Admin Secret

---

# API Layer

All HTTP requests are centralized in

```
src/api
```

Current modules

```
applications.ts

templates.ts

settings.ts

provisioning.ts
```

Every API uses

```
client.ts
```

which exposes a shared Axios instance.

---

# Type Definitions

Application models live under

```
src/types
```

Examples

```
Application

ProvisioningJob

Settings

Template
```

Keeping all models centralized avoids duplication.

---

# Styling

The application uses

- Carbon Design System
- SCSS
- CSS Modules (optional)

Shared styles are located in

```
src/styles
```

---

# Icons

Icons are provided by

Carbon Icons

Example

```tsx
import { Add } from "@carbon/icons-react";
```

---

# API Configuration

The API client is configured in

```
src/api/client.ts
```

Example

```ts
export const api = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL
});
```

---

# Production Deployment

Recommended stack

```
Browser

↓

Nginx

↓

React Static Files

↓

Backend API
```

The frontend should be served as static files.

Recommended web servers

- Nginx
- Apache
- Caddy

---

# Production Checklist

Before deployment

- Build production bundle
- Configure API URL
- Enable HTTPS
- Enable compression
- Configure cache headers
- Enable Content Security Policy
- Enable security headers

---

# Useful Commands

Install

```bash
npm install
```

Development

```bash
npm run dev
```

Build

```bash
npm run build
```

Preview

```bash
npm run preview
```

Lint

```bash
npm run lint
```

Format

```bash
npm run format
```

---

# Future Improvements

- Dark Mode
- Notifications
- Global Search
- Theme Switching
- Internationalization
- User Profile
- Role Management
- Live Provisioning Progress
- WebSocket Notifications

---

# Related Documentation

- [Project README](../README.md)
- [Backend Guide](../backend/README.md)
- [Deployment Guide](../docs/DEPLOYMENT.md)
- [Architecture Guide](../docs/ARCHITECTURE.md)