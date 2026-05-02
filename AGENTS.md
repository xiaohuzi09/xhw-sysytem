# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based monorepo containing two independent modules:

- **`app/`** — `app-image-handle`: A Wails v3 desktop application for image template management and automated product uploading. Built with Go backend services and a Vue 3 + TypeScript frontend.
- **`backend/`** — `xhw-service`: A Gin-based REST API serving as the data backend. Uses MySQL (GORM), JWT auth, Swagger docs, RustFS/S3 object storage, and Volcano Engine ARK API for image recognition.

Each module has its own `go.mod` and must be operated from within its directory.

## Common Commands

### App (`app/`)

```bash
cd app

# Install Go dependencies
go mod tidy

# Development mode (hot reload, requires wails3 CLI and Task)
task dev
# Or directly:
wails3 dev -config ./build/config.yml -port 9245

# Build for current platform
task build
# Or directly:
wails3 build

# Cross-platform builds (handled by included Taskfiles)
task darwin:build
task windows:build
task linux:build
```

The app uses `task` (Taskfile.yml) which includes platform-specific Taskfiles from `build/`. The dev server runs Vite on port 9245 by default.

### App Frontend (`app/frontend/`)

```bash
cd app/frontend

# Install dependencies
npm install

# Vite dev server (standalone, without Wails bindings)
npm run dev

# Production build
npm run build
```

### Backend (`backend/`)

```bash
cd backend

# Install dependencies
go mod tidy

# Run the server (requires MySQL and `config/config.yaml`)
go run main.go

# Build
-go build -o xhw-service

# Generate/update Swagger docs (requires swag CLI)
swag init
```

The server starts on the port defined in `config.yaml` (default 8080). Swagger UI is at `http://localhost:8080/swagger/index.html`.

## Architecture

### App (`app-image-handle`)

**Wails v3 desktop application.** Go services expose methods to the frontend via Wails runtime. The frontend is a standard Vue 3 SPA that also makes HTTP requests to the backend API using axios.

**Frontend stack:** Vue 3 + TypeScript + Vite + Element Plus + UnoCSS + Vue Router + `@wailsio/runtime`.

**Go services (`services/`):**
- `ImageService` — Image template CRUD, file dialog handling, image resizing and compositing. Templates and runtime images are stored in the app working directory (`templates.json` and `images/`).
- `AutoUploadService` — Playwright browser automation for the Dianxiaomi e-commerce platform. Launches Chromium, logs in, and uploads product data including images.
- `GreetService` — Example service (can be removed).

Services are registered in `main.go` via `application.NewService(...)`. Wails v3 automatically generates TypeScript bindings to `frontend/bindings/` when running `wails3 dev`.

**Frontend structure:**
- `src/App.vue` — Root component.
- `src/components/` — Page components: `ImageManager.vue`, `AutoUpload.vue`, `TemplateList.vue`, `MaterialList.vue`, `CombineImage.vue`, `AddTemplate.vue`, `LoginView.vue`, `UserManage.vue`.
- `src/router/index.ts` — Vue Router configuration.
- `src/api/` — Axios API clients for backend endpoints (`auth.ts`, `user.ts`, `template.ts`, `material.ts`, `presign.ts`).

### Backend (`xhw-service`)

**Layer structure:**
- `main.go` — Entry point. Loads config, initializes DB, runs auto-migration, sets up router, starts server.
- `config/` — YAML-based config loader. `LoadConfig()` reads `config.yaml` (dev) or `config.prod.yaml` (when `ENV=prod` or `ENV=production`). Exposes a global `AppConfig *Config` variable used throughout the app.
- `models/` — GORM models (`User`, `Template`, `Material`). All embed a base `Model` (ID, timestamps, soft delete). `AutoMigrate()` runs on startup and also syncs user roles against the admin whitelist via `SyncUserRoles()`.
- `controllers/` — HTTP handlers. Request/response DTOs are defined here alongside handler functions.
- `services/` — Business logic. `UserService` handles auth, CRUD, and RBAC checks. `authz.go` contains shared authorization helpers.
- `middleware/` — `AuthMiddleware()` validates JWT, then fetches the **current** user from the database (so role changes take effect immediately). `RequireRoles(...)` enforces route-level RBAC.
- `utils/` — `Response` is the unified JSON envelope (`{code, message, data}`). `Success`, `Fail`, `Error` helpers are used in all controllers. `jwt.go` handles token generation/validation using `config.AppConfig.JWT`.
- `routes/router.go` — All routes are registered here. Public routes (`/auth/login`, `/auth/register`) are outside the `AuthMiddleware()` group.

**Auth & RBAC:**
- Two roles: `admin` and `user`.
- Admin usernames are listed in `config.yaml` under `auth.admin_usernames`. On registration and on every startup migration, users matching this whitelist are promoted to `admin`.
- Controllers extract the current user from Gin context keys (`user_id`, `username`, `user_role`) set by `AuthMiddleware`.
- Some endpoints (e.g., user list, user create, user delete) require `admin` role via `middleware.RequireRoles`.

**Database:**
- MySQL. GORM connection pool is configured in `config/database.go`.
- `models.AutoMigrate()` creates/updates tables on every startup.

**External integrations:**
- **RustFS/S3** — Presigned upload/download URLs, bucket/object management (`utils/rustfs.go`, `controllers/rustfs_controller.go`).
- **ARK API** — Volcano Engine vision model for generating product titles from images (`utils/ark.go`, `controllers/ark_controller.go`).

## Development Notes

- There are **no test files** in either module yet.
- The backend binary `xhw-service` is gitignored but may exist as an untracked file in `backend/`.
- Backend config files (`config.yaml`, `config.prod.yaml`) contain real credentials and secrets — they are tracked by git. Be cautious about exposing them.
- The backend uses `gin.SetMode(cfg.Server.Mode)` — set to `release` in production.
- The app's `AutoUploadService` requires Playwright browsers to be installed (`playwright.Install()` is called in code).
- The app frontend communicates with the backend REST API via axios (configured in `frontend/src/api/index.ts`). Ensure the backend is running when testing API-dependent frontend features.


<claude-mem-context>
# Memory Context

# [xhw-sysytem] recent context, 2026-05-02 2:03pm GMT+8

Legend: 🎯session 🔴bugfix 🟣feature 🔄refactor ✅change 🔵discovery ⚖️decision
Format: ID TIME TYPE TITLE
Fetch details: get_observations([IDs]) | Search: mem-search skill

Stats: 10 obs (5,881t read) | 0t work

### May 1, 2026
454 11:50p 🔵 Frontend UI architecture and current "AI style" design system identified
456 11:54p 🔵 Frontend architecture exploration to understand current design system before Apple-style redesign
### May 2, 2026
461 12:01a 🟣 Completed ImageManager main framework Apple-style redesign and started TemplateList business page conversion
463 " 🟣 Applied Apple-style redesign to AutoUpload, ConfigView, and UserManage components; identified remaining AI-style elements for cleanup
464 12:03a 🔄 Apple-style design system applied to entire frontend
465 12:04a 🟣 Completed Apple macOS UI redesign across all frontend components with successful build verification and dev server launch
466 1:50p ✅ Development branch pull requested
468 " 🔵 Git pull blocked by untracked AGENTS.md file conflict
467 " ✅ Development branch pull initiated
469 1:53p ✅ Development branch synchronized with remote after AGENTS.md conflict resolution
</claude-mem-context>