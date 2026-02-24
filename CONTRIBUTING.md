# Contributing to Stacknest

Thank you for your interest in contributing! This document covers how to set up a development environment, the conventions we follow, and the process for submitting changes.

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Project Overview](#project-overview)
3. [Development Workflow](#development-workflow)
4. [Code Conventions](#code-conventions)
5. [Adding Features](#adding-features)
6. [Submitting a Pull Request](#submitting-a-pull-request)
7. [Reporting Bugs](#reporting-bugs)

---

## Getting Started

### Prerequisites

| Tool | Min version | Install |
|---|---|---|
| Go | 1.23 | https://go.dev/dl |
| Node.js | 18 LTS | https://nodejs.org |
| Wails CLI | v2.11 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| WebView2 | — | Bundled on Windows 11; on Windows 10: https://developer.microsoft.com/microsoft-edge/webview2 |

Verify your environment:

```bash
wails doctor
```

All prerequisites must pass before continuing.

### Clone and install

```bash
git clone <repo-url>
cd stacknest

# Go dependencies
go mod download

# Frontend dependencies
cd frontend && npm install && cd ..
```

### Start the dev server

```bash
wails dev
```

This starts hot-reload for both the Go backend and the Vite frontend simultaneously. The app window opens automatically. The frontend is also accessible at `http://localhost:34115` for browser devtools.

---

## Project Overview

```
stacknest/
├── app.go              # All IPC methods exposed to the frontend (App struct)
├── main.go             # Wails entrypoint — window config, tray, lifecycle
├── internal/           # Backend packages (one responsibility per package)
│   ├── config/         # Config struct + load/save
│   ├── configeditor/   # Read/write service configs with timestamped backup
│   ├── database/       # Adminer + HeidiSQL launcher
│   ├── downloader/     # Binary download + active-version tracking
│   ├── logs/           # fsnotify log tailer
│   ├── phpswitch/      # PHP version switching
│   ├── portcheck/      # Port conflict detection + kill
│   ├── services/       # Start/stop/restart via exec.Cmd; status polling
│   ├── ssl/            # Local CA + TLS cert generation
│   ├── terminal/       # PTY session
│   ├── tray/           # System tray
│   └── vhost/          # Virtual host .conf writer + hosts file editor
└── frontend/
    └── src/
        ├── App.tsx             # Page router
        ├── components/         # Shared UI components
        ├── pages/              # One file per page
        ├── store/
        │   └── serviceStore.ts # Central Zustand store; all IPC calls live here
        └── types/
            └── index.ts        # Shared TypeScript interfaces
```

### IPC bridge

- **Frontend → Backend**: every exported method on `App` in `app.go` is callable from TypeScript via `../../wailsjs/go/main/App`. **Do not edit `wailsjs/` manually** — Wails regenerates it on every `wails dev` run.
- **Backend → Frontend**: `runtime.EventsEmit(ctx, "event:name", payload)` in Go; `EventsOn("event:name", cb)` in `serviceStore.ts`.

---

## Development Workflow

### Branching

```
main          — stable, always buildable
feat/<name>   — new feature
fix/<name>    — bug fix
refactor/<name> — internal cleanup with no behaviour change
```

Create a branch before making any changes:

```bash
git checkout -b feat/my-feature
```

### After changing `app.go`

Any time you **add, rename, or remove** an exported method on the `App` struct, Wails must regenerate the TypeScript bindings:

```bash
wails dev   # bindings are regenerated on startup
# or, without opening the window:
wails generate module
```

### Checking your work

```bash
# Go static analysis
go vet ./...

# TypeScript type-check + frontend build (must succeed with zero errors)
cd frontend && npm run build
```

Neither command should produce errors before you open a pull request.

---

## Code Conventions

### Go

- Format with `gofmt` (no custom linter config).
- One package per internal concern — do not add logic directly to `app.go`; delegate to `internal/`.
- Cross-platform branching (`runtime.GOOS`) belongs inside the relevant `internal/` package, not in `app.go`.
- Return errors to the caller; avoid `log.Fatal` or `os.Exit` inside packages.

### TypeScript / React

- **Strict mode** — `tsconfig.json` has `strict: true`. Do not use `any` unless there is no alternative.
- **No inline styles** — use Tailwind utility classes only.
- **State** — all backend calls go through `serviceStore.ts`. Components read from the store; they do not call IPC directly (exception: one-off calls in `pages/` where the data is not shared globally).
- **No prop drilling** — if data is needed in more than two components, put it in the store.

### UI / Styling

Dark palette used throughout:

| Token | Value |
|---|---|
| Page background | `#0f1420` |
| Card / panel | `#1e2535` |
| Border | `#2a3347` |
| Accent | `blue-400` / `blue-500` |
| Success | `green-400` / `green-500` |
| Danger | `red-400` / `red-500` |

Service icon colours (Lucide):

| Service | Colour |
|---|---|
| Apache | `text-orange-400` |
| Nginx | `text-green-400` |
| MySQL | `text-blue-400` |
| PHP | `text-purple-400` |
| Redis | `text-red-400` |

---

## Adding Features

### New page

1. Add the page ID to the `Page` type union in [frontend/src/components/Sidebar.tsx](frontend/src/components/Sidebar.tsx).
2. Add a nav item (with a Lucide icon) to `navItems` in the same file.
3. Create `frontend/src/pages/MyPage.tsx`.
4. Add an `import` and a `case 'mypage':` in [frontend/src/App.tsx](frontend/src/App.tsx).

### New backend API

1. Add the business logic to the relevant `internal/` package (or create a new one).
2. Add an exported method to `App` in `app.go` that delegates to the package.
3. Run `wails dev` — the TypeScript binding is generated automatically.
4. Import and call from `serviceStore.ts` (or directly from a component for one-off calls).

### New service (beyond Apache/Nginx/MySQL/PHP/Redis)

1. Add a `ServiceName` constant in `internal/services/types.go`.
2. Implement `build<Name>Cmd` in `internal/services/manager.go`.
3. Add the service to the `services` slice in `NewManager`.
4. Add entries to `serviceBinPaths`, `serviceDataPaths`, and `serviceLogPaths` in `app.go`.
5. Add the service to the version catalog in `internal/downloader/catalog.go`.
6. Add a Lucide icon mapping in `ServiceCard.tsx`, `Binaries.tsx`, and `Sidebar.tsx`.

---

## Submitting a Pull Request

1. Make sure `go vet ./...` and `cd frontend && npm run build` both pass with **zero errors**.
2. Keep commits focused — one logical change per commit.
3. Write a clear PR description:
   - **What** changed
   - **Why** (link to an issue if applicable)
   - **How to test** it manually
4. Target the `main` branch.
5. A maintainer will review and may request changes before merging.

---

## Reporting Bugs

Open an issue with:

- Stacknest version (or commit hash)
- OS and version
- Steps to reproduce
- Expected vs actual behaviour
- Relevant log output (from **Log Viewer** or the terminal where `wails dev` is running)
