---
title: Design
group: Core Concepts
order: 1
icon: grid
description: Architecture, components, and key design decisions behind atrisos.
---

# Design

## Goals

1. Single binary that works on macOS and Ubuntu/Debian without root.
2. Every stack is a self-contained directory — portable, version-controllable.
3. Traefik routing and TLS require zero manual label work from the user.
4. Daemonless: no background Atrisos daemon, Podman handles runtime state.

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

Podman is daemonless and rootless by default. Atrisos invokes `podman compose` (Podman v4.7+ has built-in compose support) or falls back to `podman-compose` (Python wrapper).

On macOS, containers cannot run directly on the host — they require a Linux kernel. Podman solves this with `podman machine`, a lightweight Linux VM (Apple Virtualization framework on Apple Silicon, QEMU on Intel). On first run Atrisos silently creates and starts a machine named `atrisos` if none exists, showing a progress indicator. Subsequent runs start the machine automatically if it is stopped. On Linux, Podman runs natively — no VM involved.

### No Atrisos Daemon

Atrisos does not run a background service. Stack state is read live from Podman each time a status command runs. The only persistent process Atrisos manages is the Traefik container itself (and the user's app containers), which are owned by Podman.

### Shared Podman network

All stacks and Traefik share a single Podman network named `atrisos_net`. Traefik discovers containers on this network via Docker-compatible labels. Atrisos creates this network at first use if it doesn't exist.

### Stack discovery

Atrisos maintains a registry file at `~/.config/atrisos/registry.json` that stores:
- The configured root directory (default `~/atrisos-stacks`)
- A list of extra registered paths (absolute paths to individual stack dirs)

Discovery walks the root dir (one level deep, each subdirectory is a candidate stack) and merges in extra paths. A directory is recognized as a stack if it contains `compose.yml` or `docker-compose.yml`.

### Update modes

Two modes, configurable globally and per-stack:

- `manual` — user runs `atrisos update <stack>` to pull images and recreate containers
- `watch` — Atrisos watches the stack directory for file changes and re-applies via `podman compose up -d` (Compose reconciles the diff — unchanged containers are left running, only changed services are recreated)

The global default is set in `~/.config/atrisos/config.yml`. Each stack's `config.yml` can override.

### Bundled restic

Atrisos ships with a bundled `restic` binary for volume backups. On first backup run, Atrisos downloads the correct restic release for the current OS/arch from the official restic GitHub releases and stores it at `~/.config/atrisos/bin/restic`. Users do not need to install restic separately. Atrisos verifies the binary checksum after download. The bundled restic is only used by Atrisos internals — it is not added to the user's PATH.

### Traefik router naming and collision avoidance

Traefik router names follow the pattern `<stack-dir>-<service>-<hash>` where `<hash>` is the first 6 characters of the SHA-256 of the stack's absolute path. This prevents collisions when two stacks in different locations happen to share the same directory name and service name. Example: two stacks both named `myapp` with a `web` service get routers `myapp-web-a3f291` and `myapp-web-d8c104`.

### TUI log streaming

The log panel uses a single goroutine that runs `podman compose logs -f --timestamps --no-color` as a subprocess and streams lines into a bubbletea `viewport` component via a channel. All services are multiplexed into one stream, each line prefixed with the service name. The viewport supports keyboard scrolling and buffers the last **2000 lines** (hardcoded). Older lines are dropped as new ones arrive. The goroutine is cancelled when the log panel is closed.

### Backup scheduling via system scheduler

Atrisos does not run a daemon, so backup schedules are handed off to the OS scheduler. When a stack has `backup.enabled: true` and a `backup.schedule` cron expression, `atrisos up` installs a scheduler unit and `atrisos down` removes it.

- **Linux**: generates a systemd user service + timer pair under `~/.config/systemd/user/`:
  - `atrisos-backup-<stack>.service` — runs `atrisos backup <stack>`
  - `atrisos-backup-<stack>.timer` — cron schedule converted to `OnCalendar=` syntax
  - Enabled via `systemctl --user enable --now atrisos-backup-<stack>.timer`
- **macOS**: generates a launchd plist at `~/Library/LaunchAgents/io.atrisos.backup.<stack>.plist` with `StartCalendarInterval` keys parsed from the cron expression. Loaded via `launchctl load`.

### Auto-start on boot

Stacks with `auto_start: true` in `config.yml` are registered with the OS init system so their containers start after a reboot without manual intervention.

- **Linux**: generates `~/.config/systemd/user/atrisos-<stack>.service` with `WantedBy=default.target`. Enabled via `systemctl --user enable atrisos-<stack>`. The service runs `atrisos up <stack>`.
- **macOS**: generates `~/Library/LaunchAgents/io.atrisos.<stack>.plist` with `RunAtLoad: true`. Loaded via `launchctl load`.

Both are installed on `atrisos up` when `auto_start: true` and removed on `atrisos down`.

### Bulk operations

`atrisos up --all`, `atrisos down --all`, and `atrisos update --all` operate on all discovered stacks sequentially in discovery order (root dir stacks first, then registered extra paths, both alphabetically). A single stack failure prints an error and continues to the next stack rather than aborting the whole operation. All three also accept `--tag <tag>` to operate only on stacks whose `config.yml` includes that tag.

### ACME staging

The `tls` field in each `domains` entry accepts three values:

- `true` — production Let's Encrypt (trusted cert, rate-limited)
- `staging` — Let's Encrypt staging CA (no rate limits, browser shows untrusted warning)
- `false` — HTTP only, no certificate

Traefik uses a separate certificate resolver (`letsencrypt-staging`) for staging domains, configured in the managed Traefik compose file alongside the production resolver.

### Webhook notifications

Stacks with a `notify.webhook` URL in `config.yml` POST a JSON payload to that URL on these events:

- Unexpected container exit (container stops without `atrisos down`)
- Backup failure
- TLS certificate expiry within 7 days

Payload format:
```json
{
  "event": "container_exit",
  "stack": "myapp",
  "service": "web",
  "timestamp": "2026-06-27T14:00:00Z",
  "message": "Container myapp-web exited with code 1"
}
```

Webhook URL is per-stack only (no global fallback). Stacks without a `notify` block send no notifications. Compatible with Slack, Discord, ntfy, and any service that accepts a POST with a JSON body.

### Container exec and shell

`atrisos exec <stack> <service> -- <command>` and `atrisos shell <stack> <service>` wrap `podman exec`. In the TUI, pressing `e` on a selected service suspends the TUI and opens a shell directly in the terminal, then returns to the TUI on exit.

### SELinux auto-detection

On Linux, atrisos runs `getenforce` at startup. If SELinux is enforcing, it appends `:z` to all bind-mount volume entries in the merged compose document (named volumes are unaffected — SELinux relabelling only applies to bind mounts). Named volumes managed by Podman do not need this label.

### Compose override files

If a stack directory contains `compose.override.yml`, atrisos deep-merges it with `compose.yml` before applying Traefik label injection. This follows the same merge semantics as `docker compose` (service keys in the override layer replace or extend the base). No config needed — the file's presence is the signal.

### Port conflict detection

Before starting Traefik, atrisos checks whether the configured HTTP and HTTPS ports (default 80 and 443) are already bound using a TCP dial attempt. If either port is taken, atrisos exits immediately with a message identifying the conflicting process (via `lsof -i :<port>` on macOS/Linux).

### Container health checks in TUI

The TUI reads Podman health check state for each container (`podman ps --format json` includes `Health.Status`). Three states are surfaced:

- `healthy` — health check passing
- `starting` — container running but health check not yet passed
- `unhealthy` — health check failing

The stack list uses a distinct indicator for stacks with unhealthy containers (`⚠` instead of `●`). The detail panel shows per-service health status alongside running/stopped.

### Image update awareness

`atrisos outdated` queries the registry for each image used across all stacks and compares the remote digest against the locally pulled digest. Services with available updates are listed. The TUI runs this check in a background goroutine on startup and shows a small `↑` badge next to stack names with available updates.

### Config validation

`atrisos validate <stack>` performs a dry-run check: config.yml schema, compose.yml syntax (via `podman compose config`), cross-references between domains service names and compose services, presence of `.env`. All errors are collected and reported at once rather than failing on the first one.

### Self-update

`atrisos self-update` fetches the latest release from GitHub, verifies the checksum, and replaces the running binary in place. `atrisos version` shows both the current version and the latest available (checked in the background, cached for 24 hours).

### Stack export and import

`atrisos export <stack>` creates a `.tar.gz` containing `compose.yml`, `config.yml`, `compose.override.yml` (if present), and `.env.example`. The `.env` file is intentionally excluded to avoid exporting secrets. `atrisos import <file.tar.gz>` extracts into the stacks root directory and prompts the user to create a `.env` from `.env.example`.

### Stack init templates

`atrisos init` fetches template files from the `templates/` directory in the atrisos GitHub repository (`main` branch) at runtime. Templates are cached locally in `~/.config/atrisos/templates-cache/` after the first download so subsequent `atrisos init` calls work offline. atrisos checks for a newer cache version (by comparing a manifest file) only when online. See [templates.md](templates.md) for the template format and available templates.

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
│   ├── backup.go
│   ├── exec.go
│   ├── validate.go
│   ├── outdated.go
│   ├── export.go
│   ├── selfupdate.go
│   └── traefik.go
├── internal/
│   ├── config/             # global config loading
│   ├── stack/              # Stack struct, loader, discovery
│   ├── compose/            # podman compose wrapper + merge pipeline
│   ├── traefik/            # Traefik label gen, network, managed stack
│   ├── registry/           # registry.json read/write
│   ├── watcher/            # fsnotify-based file watcher
│   ├── backup/             # backup scheduling and restic invocation
│   ├── restic/             # bundled restic binary management (download, verify, exec)
│   ├── scheduler/          # systemd timer / launchd plist generation and management
│   ├── notify/             # webhook POST on events
│   ├── outdated/           # registry digest comparison for image update checks
│   ├── selfupdate/         # GitHub release fetch, checksum verify, binary replace
│   ├── health/             # Podman health check state polling
│   └── templates/          # GitHub template fetching and local cache management
├── tui/                    # bubbletea TUI components
│   ├── app.go              # root model
│   ├── stacklist.go        # stack list panel
│   ├── detail.go           # stack detail panel
│   └── logs.go             # log streaming panel (multiplexed viewport)
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
| `text/template` (stdlib) | Render stack init templates |

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
│   postgres  ● ↑ │ Domain:   https://myapp.example.com   │
│   redis     ●   │ Updated:  2 hours ago                  │
│   grafana   ⚠   │                                        │
│                 │ Containers                             │
│                 │   web     ● running  healthy   Up 2h   │
│                 │   worker  ● running  healthy   Up 2h   │
│                 │   db      ● running  starting  Up 10s  │
│                 │                                        │
│                 │ [u]update [r]restart [l]logs [e]shell  │
│                 │ [↑↓]navigate  [/]filter  [?]help  [q]quit│
└─────────────────┴───────────────────────────────────────┘
```

Stack list indicators: `●` running, `○` stopped, `◑` partial, `↺` updating, `⚠` unhealthy container, `↑` image update available.
Detail panel health states: `healthy`, `starting`, `unhealthy` (sourced from Podman health check).
