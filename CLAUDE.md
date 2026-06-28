# atrisos — CLAUDE.md

Go module: `github.com/sonmezerekrem/atrisos`  
Go version: 1.24.2

## What this project is

atrisos is a CLI/TUI tool for managing Podman Compose stacks with automatic Traefik routing. Users write a plain `compose.yml` with zero Traefik content; atrisos injects routing labels and network attachment at runtime before invoking `podman compose`. Targets macOS (Apple Silicon + Intel) and Ubuntu/Debian (amd64 + arm64).

## Build and test

```sh
make build          # builds ./atrisos binary (dev version)
make install        # builds + installs to /usr/local/bin
make lint           # go vet ./...
make build-all      # cross-compile all four targets to dist/

cd app && go test ./internal/...   # run all tests (84 tests, no integration setup needed)
cd app && go vet ./...
```

The `Version` variable is injected via ldflags: `-X github.com/sonmezerekrem/atrisos/cmd.Version=$(VERSION)`.

## Directory layout

```
app/                     # all Go source code
  main.go                # calls cmd.Execute()
  cmd/                   # cobra commands, one file per subcommand group
  internal/
    config/              # global config: ~/.config/atrisos/config.yml
    registry/            # registry.json — root dir + extra registered paths
    stack/               # Stack struct, loader, discovery, types
    compose/             # compose merge pipeline + podman compose runner
    traefik/             # label generation, network, managed Traefik stack
    podman/              # machine management (macOS) + container status polling
    backup/              # restic backup invocation
    restic/              # auto-download restic binary to ~/.config/atrisos/bin/
    scheduler/           # systemd timers (Linux) + launchd plists (macOS)
    notify/              # webhook POST on events
    outdated/            # compare local vs remote image digests
    selfupdate/          # GitHub release fetch + atomic binary replace
    templates/           # fetch templates from GitHub, local cache management
  tui/                   # bubbletea TUI (app.go, stacklist.go, detail.go, logs.go, styles.go)
docs/                    # architecture and user documentation
templates/               # bundled init templates (manifest.json + per-template dirs)
scripts/install.sh       # curl-pipe installer
.goreleaser.yml          # release builds
.github/workflows/       # ci.yml (go build + vet), release.yml (goreleaser on v* tags)
```

## Key architectural patterns

### Compose merge pipeline

The core invariant: user `compose.yml` files are **never modified**. atrisos reads the file into `map[string]interface{}` (`internal/compose/types.go`), deep-merges Traefik labels and `atrisos_net` network attachment for each service with a `domains` entry, writes the result to a temp file, and passes that to `podman compose -f <tmpfile>`. The temp file is deleted after the command exits.

`internal/compose/merge.go` — `Merge(doc, stackCfg, stackPath)` is the entry point. `LoadAndMerge(dir, cfg)` also handles optional `compose.override.yml`.

`atrisos render <stack>` prints the merged document without running anything.

### Router naming and collision avoidance

Router names follow the pattern `<stack-dir>-<service>-<hash>` where `<hash>` is the first 6 hex characters of the SHA-256 of the stack's absolute path. Two stacks named `myapp` in different directories get different hashes, preventing Traefik router collisions.

`internal/traefik/labels.go` — `RouterName(stackDir, service, stackAbsPath)`.

### No daemon

atrisos has no background process. Stack state is read live from Podman on every status call. Scheduled backups and auto-start-on-boot are delegated to the OS scheduler (systemd timers on Linux, launchd plists on macOS) installed on `atrisos up` and removed on `atrisos down`.

### Stack discovery

`internal/stack/discover.go` — walks the root directory one level deep and merges in extra paths from `~/.config/atrisos/registry.json`. A directory is a stack if it contains `compose.yml` or `docker-compose.yml`. Results are deduped and sorted by name.

### SELinux handling

`internal/compose/merge.go` — `selinuxEnforcing()` runs `getenforce` at merge time. If enforcing, `:z` is appended to all bind-mount volume strings in the merged document (named volumes are untouched).

### TUI

`tui/app.go` — `AppModel` is the root bubbletea model. Two panels: `panelList` (stack list with status indicators) and `panelLogs` (streaming log viewport). Log streaming runs `podman compose logs -f` as a subprocess and pipes lines through a channel. Buffer is capped at 2000 lines. Outdated image check runs in a background goroutine on startup and sends an `outdatedResultMsg` when done.

`tui/styles.go` — all lipgloss styles with `AdaptiveColor` for light/dark terminal support.

### Global config vs stack config

Global: `~/.config/atrisos/config.yml` — ACME email, stacks root, default update mode, network name, Traefik settings.  
Per-stack: `<stack-dir>/config.yml` — name, description, tags, domains array, update mode override, auto_start, backup, notify.

`internal/config/config.go` — `Load()` reads global config with safe defaults for missing file (first-run wizard in `cmd/root.go` creates it).

### Circular import avoidance

`internal/traefik/manager.go` uses `os/exec` directly for compose operations instead of importing `internal/compose`, which in turn imports `internal/traefik/labels`. Keep this separation.

## Stack format

Each stack is a directory containing:
- `compose.yml` — standard Compose file, **no Traefik content**
- `.env` — environment variables
- `config.yml` — atrisos config (name, domains, tags, backup, notify, etc.)
- `compose.override.yml` — optional, deep-merged before label injection

The `domains` array in `config.yml` is what drives Traefik label injection. Services not listed there are left on the stack's internal network only.

## TLS modes

`tls` field on each `domains` entry:
- `true` — production Let's Encrypt (default)
- `staging` — LE staging CA (no rate limits, untrusted cert; use during setup)
- `false` — HTTP only

Two ACME resolvers (`letsencrypt` and `letsencrypt-staging`) are always present in the managed Traefik compose.

## Managed Traefik

Traefik runs as a Compose stack at `~/.config/atrisos/traefik/`. atrisos generates and owns those files — they should not be edited by hand. Started automatically on first `atrisos up`. Controlled via `atrisos traefik up/down/restart/status/logs`.

Podman socket path is platform-specific:
- Linux: `/run/user/<UID>/podman/podman.sock`
- macOS: queried from `podman machine inspect`

## Release

Tag `v*` → GitHub Actions runs goreleaser → publishes four binaries to GitHub Releases:
- `atrisos-linux-amd64`, `atrisos-linux-arm64`
- `atrisos-darwin-amd64`, `atrisos-darwin-arm64`

`CGO_ENABLED=0` — all builds are fully static.

`cmd.Version` is set by goreleaser ldflags. When building locally it defaults to `dev` unless overridden: `make build VERSION=v0.2.0`.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Stack not found |
| 3 | Stack validation error |
| 4 | Podman not found or not running |
| 5 | Traefik error |
