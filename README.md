# Stacknest

A cross-platform desktop app for managing local development services — Apache, Nginx, MySQL, PHP, Redis. Built with [Wails v2](https://wails.io): Go backend + React/TypeScript frontend rendered in a native WebView.

---

## Features

- **Dashboard** — start/stop/restart individual services; toggle which services are included in *Start All*
- **Binaries** — download and switch between multiple versions of each service
- **Config Editor** — edit service config files (`httpd.conf`, `nginx.conf`, `my.ini`, `php.ini`) with syntax highlighting and automatic timestamped backups
- **Virtual Hosts** — add/remove Apache virtual hosts; hosts file updated automatically (UAC prompt on Windows when not running as admin)
- **SSL** — generate a local CA + per-domain TLS certificates; install CA into the OS trust store
- **Database** — launch Adminer in-browser or open HeidiSQL
- **Log Viewer** — real-time log tailing per service
- **Terminal** — embedded PTY terminal
- **System Tray** — minimize to tray

---

## Tech Stack

| Layer | Technology |
|---|---|
| Desktop shell | Wails v2.11 |
| Backend | Go 1.23 |
| Frontend | React 18 + TypeScript 5 |
| Styling | Tailwind CSS v3 |
| State | Zustand v4 |
| Editor | CodeMirror 6 (`@uiw/react-codemirror`) |
| Terminal | xterm.js v6 |
| Icons | Lucide React |
| Build tool | Vite 5 |

---

## Prerequisites

| Tool | Min version | Install |
|---|---|---|
| Go | 1.23 | https://go.dev/dl |
| Node.js | 18 LTS | https://nodejs.org |
| Wails CLI | v2.11 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| WebView2 | — | Bundled on Windows 11; on Windows 10 install the runtime from https://developer.microsoft.com/microsoft-edge/webview2 |

Verify your environment:

```bash
wails doctor
```

---

## Project Structure

```
stacknest/
├── app.go                    # App struct — all IPC methods exposed to the frontend
├── main.go                   # Wails entrypoint (window config, tray, lifecycle hooks)
├── go.mod
├── build/                    # Wails build assets (icons, manifests) — committed to git
│   └── bin/                  # Compiled output — gitignored
├── frontend/
│   ├── src/
│   │   ├── App.tsx           # Page router (useState<Page>)
│   │   ├── components/       # ServiceCard, Sidebar, PortConflictModal, …
│   │   ├── pages/            # Dashboard, Binaries, ConfigEditor, VHosts, SSL, …
│   │   ├── store/
│   │   │   └── serviceStore.ts  # Single Zustand store; all IPC calls live here
│   │   └── types/
│   │       └── index.ts      # Shared TypeScript interfaces
│   └── wailsjs/              # Auto-generated IPC bindings — do not edit manually
└── internal/
    ├── config/               # Config struct; loads/saves to %APPDATA%/Stacknest/
    ├── configeditor/         # Read/write service config files with auto-backup
    ├── database/             # Adminer PHP server + HeidiSQL launcher
    ├── downloader/           # Binary download, version catalog, active-version tracking
    ├── logs/                 # fsnotify-based log tailer
    ├── phpswitch/            # Scan PHP installs, switch active version
    ├── portcheck/            # Detect and kill port-occupying processes
    ├── services/             # Start/stop/restart services via exec.Cmd; polls status every 3s
    ├── ssl/                  # Local CA + per-domain cert generation (stdlib RSA-2048)
    ├── terminal/             # PTY session (go-pty)
    ├── tray/                 # System tray icon
    └── vhost/                # Apache vhost .conf writer + hosts file editor
```

### IPC bridge

- **Frontend → Backend**: import from `../../wailsjs/go/main/App` (auto-generated). Every exported method on the `App` struct is callable from TypeScript.
- **Backend → Frontend**: `runtime.EventsEmit(ctx, "event:name", payload)` in Go; `EventsOn("event:name", cb)` in `serviceStore.ts`.

| Event | Payload | Purpose |
|---|---|---|
| `services:updated` | `ServiceInfo[]` | Polled every 3s; updates dashboard |
| `log:line` | `LogEntry` | Real-time log tail |
| `term:output` | `string` | PTY output stream |
| `term:exit` | — | Shell process exited |
| `binary:progress` | `{service, version, pct}` | Download progress |
| `binary:done` | `{service, version, error}` | Download finished |

---

## Setup

```bash
# 1. Clone the repository
git clone <repo-url>
cd stacknest

# 2. Install Go dependencies
go mod download

# 3. Install frontend dependencies
cd frontend && npm install && cd ..
```

---

## Development

```bash
# Full hot-reload dev mode (Go backend + Vite frontend simultaneously)
wails dev
```

The app window opens automatically. The frontend is also served at `http://localhost:34115` for browser-based devtools access.

> **Important:** after adding or renaming any exported method on `App` in `app.go`, run `wails dev` once so Wails regenerates the TypeScript bindings in `frontend/wailsjs/go/main/App.ts`.

### Frontend only (faster UI iteration)

```bash
cd frontend
npm run build     # TypeScript type-check + Vite production build
npm run preview   # Serve built frontend at localhost:4173
```

---

## Build

```bash
# Production build → build/bin/stacknest.exe
wails build

# With devtools enabled
wails build -devtools
```

The output binary is fully self-contained — the compiled frontend is embedded via Go's `embed` package.

---

## Configuration

The config file is persisted automatically on first launch:

| OS | Path |
|---|---|
| Windows | `%APPDATA%\Stacknest\config.json` |
| macOS | `~/Library/Application Support/Stacknest/config.json` |
| Linux | `~/.config/stacknest/config.json` |

`RootPath` (where binaries, data, logs, and vhosts are stored):

| Mode | Path |
|---|---|
| Production | Same directory as the executable |
| `wails dev` | Repository root (current working directory) |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for setup instructions, code conventions, and the pull request process.

---

## License

MIT
