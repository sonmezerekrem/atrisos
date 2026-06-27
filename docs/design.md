# Design

## Goals

1. Single binary that works on macOS and Ubuntu/Debian without root.
2. Every stack is a self-contained directory — portable, version-controllable.
3. Traefik routing and TLS require zero manual label work from the user.
4. Daemonless: no background atrisos daemon, Podman handles runtime state.

## Architecture overview

```
atrisos binary (Go)
│
├── CLI layer (cobra)            # subcommands, flags, output formatting
├── TUI layer (bubbletea)        # interactive dashboard launched by `atrisos tui` or bare `atrisos`
│
├── Stack manager                # discover, parse, and operate on stacks
│   ├── Discovery                # scan root dir + registered paths
│   ├── Stack loader             # reads compose.yml + .env + config.yml
│   └── Compose runner           # wraps `podman compose` invocations
│
├── Traefik manager              # manages the shared Traefik Compose stack
│   ├── Label generator          # builds Traefik labels from config.yml
│   └── Network manager          # ensures shared Podman network exists
│
└── Global config                # reads ~/.config/atrisos/config.yml
```

## Key design decisions

### Podman, not Docker

Podman is daemonless and rootless by default. atrisos invokes `podman compose` (Podman v4.7+ has built-in compose support) or falls back to `podman-compose` (Python wrapper). On macOS, Podman runs inside a lightweight VM (`podman machine`) — atrisos handles machine init on first run.

### No atrisos daemon

atrisos does not run a background service. Stack state is read live from Podman each time a status command runs. The only persistent process atrisos manages is the Traefik container itself (and the user's app containers), which are owned by Podman.

### Shared Podman network

All stacks and Traefik share a single Podman network named `atrisos_net`. Traefik discovers containers on this network via Docker-compatible labels. atrisos creates this network at first use if it doesn't exist.

### Stack discovery

atrisos maintains a registry file at `~/.config/atrisos/registry.json` that stores:
- The configured root directory (default `~/atrisos-stacks`)
- A list of extra registered paths (absolute paths to individual stack dirs)

Discovery walks the root dir (one level deep, each subdirectory is a candidate stack) and merges in extra paths. A directory is recognized as a stack if it contains `compose.yml` or `docker-compose.yml`.

### Update modes

Two modes, configurable globally and per-stack:

- `manual` — user runs `atrisos update <stack>` to pull images and recreate containers
- `watch` — atrisos watches the stack directory for file changes and re-applies automatically

The global default is set in `~/.config/atrisos/config.yml`. Each stack's `config.yml` can override.

## Go project structure

```
atrisos/
├── main.go
├── cmd/                    # cobra commands (one file per subcommand group)
│   ├── root.go
│   ├── up.go
│   ├── down.go
│   ├── update.go
│   ├── watch.go
│   ├── logs.go
│   ├── status.go
│   ├── register.go
│   ├── list.go
│   ├── init.go
│   └── traefik.go
├── internal/
│   ├── config/             # global config loading
│   ├── stack/              # Stack struct, loader, discovery
│   ├── compose/            # podman compose wrapper
│   ├── traefik/            # Traefik label gen, network, managed stack
│   ├── registry/           # registry.json read/write
│   ├── watcher/            # fsnotify-based file watcher
│   └── backup/             # backup scheduling and execution
├── tui/                    # bubbletea TUI components
│   ├── app.go              # root model
│   ├── stacklist.go        # stack list panel
│   ├── detail.go           # stack detail panel
│   └── logs.go             # log streaming panel
├── docs/
├── scripts/
│   └── install.sh          # curl-pipe installer
└── go.mod
```

## Key Go dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI subcommands and flags |
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/lipgloss` | TUI styling |
| `github.com/charmbracelet/bubbles` | TUI components (list, viewport, spinner) |
| `gopkg.in/yaml.v3` | Parse compose.yml and config.yml |
| `github.com/joho/godotenv` | Parse .env files |
| `github.com/fsnotify/fsnotify` | Watch stack dirs for changes |
| `github.com/pelletier/go-toml` | Optional: TOML alt for global config |

## Platform-specific behavior

### macOS

- Podman requires a running `podman machine`. atrisos checks on startup and offers to init/start it if needed.
- Homebrew is the expected way to have Podman installed, but the installer script checks and guides the user.
- `launchd` plist generation for auto-start-on-boot stacks (future scope).

### Linux (Ubuntu/Debian)

- Podman installed via apt (`podman` package, Debian bookworm+ or Ubuntu 22.04+).
- `systemd --user` unit generation for auto-start-on-boot stacks (future scope).
- `podman compose` or `podman-compose` must be present; atrisos checks at startup.

## TUI layout

```
┌─────────────────────────────────────────────────────────┐
│ atrisos                              [traefik: running]  │
├─────────────────┬───────────────────────────────────────┤
│ STACKS          │ myapp                                  │
│                 │ ─────────────────────────────────────  │
│ ▶ myapp    ●   │ Status:   running (3 containers)       │
│   postgres  ●   │ Domain:   https://myapp.example.com   │
│   redis     ●   │ Updated:  2 hours ago                  │
│   grafana   ○   │                                        │
│             │   │ Containers                             │
│             │   │   web     running   Up 2h              │
│             │   │   worker  running   Up 2h              │
│             │   │   db      running   Up 2h              │
│             │   │                                        │
│             │   │ [u] update  [r] restart  [l] logs      │
│             │   │ [↑↓] navigate  [enter] select  [q] quit│
└─────────────┴───────────────────────────────────────────┘
```

Status indicators: `●` running, `○` stopped, `◑` partial (some containers down), `↺` updating.
