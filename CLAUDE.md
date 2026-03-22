# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Onefetch is a self-hosted Docker application for downloading files from 1fichier.com directly to a server, with optional Jellyfin media server integration. It consists of a Go backend and a SvelteKit frontend.

## Development Commands

### Frontend (`frontend/`)

```bash
pnpm install          # Install dependencies
pnpm run dev          # Dev server (proxies /api to localhost:3000)
pnpm run build        # Production build → ./build
pnpm run check        # TypeScript type checking
pnpm lint             # ESLint
pnpm lint:fix         # ESLint avec auto-fix
pnpm format           # Prettier (auto-format)
```

### Backend (`backend/`)

```bash
go mod download       # Install dependencies
go run main.go        # Run in development mode
go build -o onefetch-app .  # Production build
go test ./...         # Run tests
```

### Docker

```bash
docker compose up     # Run full stack (port 3030 → 3000)
```

## Architecture

### Backend (`backend/`)

Layered Go application using the Fiber web framework:

- **`internal/handler/`** — HTTP request/response handling
- **`internal/service/`** — Business logic
- **`internal/repository/`** — Data access via GORM (SQLite)
- **`internal/model/`** — Data models and enums
- **`internal/config/`** — Environment variable configuration
- **`pkg/client/`** — External API clients (1fichier, Jellyfin)
- **`pkg/sse/`** — Server-Sent Events manager for real-time progress
- **`pkg/worker/`** — Background download worker

Key environment variables (with defaults):
- `APP_PORT` (3000)
- `APP_DOWNLOAD_PATH` (./downloads)
- `APP_DATA_PATH` (./data)
- `APP_API_URL_1FICHIER`
- `APP_API_URL_JELLYFIN`

API routes: `/api/settings`, `/api/downloads`, `/api/downloads/streams` (SSE), `/api/files`. In production, the backend also serves the static SvelteKit build.

### Frontend (`frontend/`)

SvelteKit 5 app with static adapter, using Svelte Runes for state management:

- **`src/lib/api/`** — Centralized API client (`api.svelte.ts`) and SSE hook
- **`src/lib/state/`** — Page-level state (one file per route, e.g. `active-state.svelte.ts`)
- **`src/lib/components/`** — Reusable Svelte components
- **`src/lib/types/`** — TypeScript types
- **`src/routes/`** — SvelteKit routes: `/` (new download), `/active`, `/history`, `/files`, `/settings`

### Real-time Updates

Downloads report progress via SSE at `/api/downloads/streams`. The frontend connects via the `useSSE` hook in `api/use-sse.svelte.ts`, which updates the active downloads state reactively.

### Dev vs Production

In dev, Vite proxies `/api/*` to `http://localhost:3000`. In production, everything runs from the same origin (Go serves the static build from `./web`).

## API Specification

The full API is documented in `specs/openapi.yml`.
