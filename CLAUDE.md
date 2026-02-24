# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development
```bash
wails dev          # Full dev mode — Go hot-reload backend + Vite frontend (http://localhost:34115)
wails build        # Production build → stacknest.exe
```

### Frontend only (inside `frontend/`)
```bash
npm run build      # TypeScript type-check + Vite production build
npm run preview    # Serve the built frontend
```

No test suite exists. No linter config beyond TypeScript strict mode.

After modifying any exported `App` method in `app.go`, run `wails dev` once so Wails regenerates the TypeScript bindings in `frontend/wailsjs/go/main/App.ts`. The frontend imports all backend calls from there.

---

## Architecture

Stacknest is a **Wails v2** desktop app: Go backend exposed as IPC methods, React+TypeScript frontend.

### IPC Bridge

- **Frontend → Backend**: Import from `../../wailsjs/go/main/App` (auto-generated). Every public method on `app.go`'s `App` struct becomes callable.
- **Backend → Frontend**: `runtime.EventsEmit(ctx, "event:name", payload)` in Go; `EventsOn("event:name", cb)` in frontend (`serviceStore.ts` listens for `services:updated`, `log:line`, `term:output`, `term:exit`).

### Backend (`app.go` + `internal/`)

`app.go` owns a single `App` struct with one field per internal manager:

| Field | Package | Responsibility |
|---|---|---|
| `svcMgr` | `internal/services` | Start/stop/restart Apache, Nginx, MySQL, PHP, Redis via `exec.Cmd`; polls status every 3s and emits `services:updated` |
| `vhostMgr` | `internal/vhost` | Writes Apache `.conf` files + edits `/etc/hosts` |
| `phpSwitcher` | `internal/phpswitch` | Scans for PHP installs, switches active version, persists to `php_versions.json` |
| `cfgEditor` | `internal/configeditor` | Read/write service config files with timestamped backup |
| `sslMgr` | `internal/ssl` | Generates local CA + per-domain certs (RSA-2048, stdlib only), installs CA in OS trust store |
| `adminerSrv` | `internal/database` | Spawns `php -S` to serve Adminer; opens HeidiSQL |
| `termSession` | `internal/terminal` | PTY session via `go-pty`; streams output as `term:output` events |
| `logCancel` | `internal/logs` | `fsnotify`-based log tail; streams as `log:line` events |

All managers are initialized in `NewApp()` with `cfg.RootPath` (Laragon root, default `C:\laragon`). Config persists to `%APPDATA%\Stacknest\config.json`.

Cross-platform branching (`runtime.GOOS`) is in each manager — not in `app.go`.

### Frontend (`frontend/src/`)

- **State**: Single Zustand store at `store/serviceStore.ts`. All backend data lives here. Components read from store; actions call Wails IPC then update state.
- **Routing**: `App.tsx` switches on `useState<Page>`. Pages are in `pages/`. Add a new page by: (1) add to `Page` type in `Sidebar.tsx`, (2) add nav item in `Sidebar.tsx`, (3) add import + `case` in `App.tsx`.
- **Styling**: Tailwind CSS. Dark theme palette: background `#0f1420`, card `#1e2535`, border `#2a3347`, accent blue-400/500.
- **Terminal**: `pages/Terminal.tsx` uses `@xterm/xterm` + `@xterm/addon-fit`. Go streams PTY output as string events.
- **Config Editor**: `pages/ConfigEditor.tsx` uses `@uiw/react-codemirror` with `@codemirror/legacy-modes` for Apache/Nginx/INI syntax.

### SSL Certificate flow
`sslMgr.EnsureCA()` → generates CA at `rootPath/ssl/ca.{crt,key}`. `GenerateCert(domain)` writes to `rootPath/vhosts/{domain}.{crt,key}` — exactly where `vhostMgr` references them in Apache SSL config. No changes to vhost manager needed when generating certs.

### Adminer flow
`adminerSrv.Start()` reads active PHP path from `rootPath/php_versions.json`, finds `C:\laragon\etc\apps\adminer\index.php`, runs `php -S 127.0.0.1:<port>`, then Go calls `runtime.BrowserOpenURL` to open the browser.
