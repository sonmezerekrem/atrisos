# Atrisos — CLAUDE.md

Go module: `github.com/sonmezerekrem/atrisos`  
Go version: 1.26

## What this project is

Atrisos is a CLI/TUI tool for managing Podman Compose stacks with automatic Traefik routing. Users write a plain `compose.yml` with zero Traefik content; Atrisos injects routing labels and network attachment at runtime before invoking `podman compose`. Targets macOS (Apple Silicon + Intel) and Ubuntu/Debian (amd64 + arm64).

## Build and test

```sh
make build          # builds ./atrisos binary (dev version)
make install        # builds + installs to /usr/local/bin
make lint           # go vet ./...
make build-all      # cross-compile all four targets to dist/
make docs           # regenerate docs/nav.json from docs/content/
make docs-serve     # regenerate + serve docs at http://localhost:8080

cd app && go test ./internal/...   # run all tests (no integration setup needed)
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
docs/                    # GitHub Pages documentation site
  index.html             # docs shell
  app.js                 # React docs UI (no build step)
  build.mjs              # scans content/ → nav.json; copies templates manifest → templates.json
  nav.json               # generated navigation (run `make docs`)
  templates.json         # generated template catalog
  content/               # markdown pages (YAML frontmatter)
    getting-started.md
    install.md
    stack-format.md
    traefik.md
    global-config.md
    design.md
    cli-reference/
    templates/
    agents/
  assets/                # static assets (logo, favicon)
templates/               # bundled init templates (manifest.json + per-template dirs)
  basic/ ghost/ minio/ mongo/ mysql/ postgres/ registry/ seaweedfs/ umami/ valkey/ wordpress/
scripts/install.sh       # curl-pipe installer
.goreleaser.yml          # release builds
.github/workflows/
  ci.yml                 # go build + vet + test/coverage on push/PR
  release.yml            # goreleaser on v* tags
  docs.yml               # GitHub Pages deploy on push to main
.claude/skills/release/  # /release skill — version bump, release notes, tag + push
```

## Key architectural patterns

### Compose merge pipeline

The core invariant: user `compose.yml` files are **never modified**. Atrisos reads the file into `map[string]interface{}` (`internal/compose/types.go`), deep-merges Traefik labels and `atrisos_net` network attachment for each service with a `domains` entry, writes the result to a temp file, and passes that to `podman compose -f <tmpfile>`. The temp file is deleted after the command exits.

`internal/compose/merge.go` — `Merge(doc, stackCfg, stackPath)` is the entry point. `LoadAndMerge(dir, cfg)` also handles optional `compose.override.yml`.

`atrisos render <stack>` prints the merged document without running anything.

### Router naming and collision avoidance

Router names follow the pattern `<stack-dir>-<service>-<hash>` where `<hash>` is the first 6 hex characters of the SHA-256 of the stack's absolute path. Two stacks named `myapp` in different directories get different hashes, preventing Traefik router collisions.

`internal/traefik/labels.go` — `RouterName(stackDir, service, stackAbsPath)`.

### No daemon

Atrisos has no background process. Stack state is read live from Podman on every status call. Scheduled backups and auto-start-on-boot are delegated to the OS scheduler (systemd timers on Linux, launchd plists on macOS) installed on `atrisos up` and removed on `atrisos down`.

`internal/scheduler/scheduler.go` — `RemoveBackupTimer` is a no-op when `backup.enabled` is false; `linuxRemoveUnit` skips the systemctl call if the unit file does not exist (prevents spurious "Failed to disable unit" errors).

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

### Template system

Templates live in `templates/` and are fetched from GitHub raw URLs at runtime, cached at `~/.config/atrisos/templates-cache/`. Each template has a `template.yml` (prompts), `compose.yml.tmpl`, `config.yml.tmpl`, `.env.tmpl`, `.env.example.tmpl`.

`internal/templates/templates.go` — `Prompt.Generate` field drives auto-generation: `"random_password"` (24-char alphanumeric via `crypto/rand`) or `"traefik_me_domain"` (`<slug>-<rand8hex>.traefik.me`).

After all prompts, `tls` is auto-set to `"false"` for `.traefik.me`, `.nip.io`, `.sslip.io` domains and `"true"` for everything else (`cmd/init.go`).

All compose templates use fully qualified image names (`docker.io/library/postgres:16-alpine`) to avoid Podman unqualified-name errors on Linux.

## Stack format

Each stack is a directory containing:
- `compose.yml` — standard Compose file, **no Traefik content**
- `.env` — environment variables
- `config.yml` — Atrisos config (name, domains, tags, backup, notify, etc.)
- `compose.override.yml` — optional, deep-merged before label injection

The `domains` array in `config.yml` is what drives Traefik label injection. Services not listed there are left on the stack's internal network only.

## TLS modes

`tls` field on each `domains` entry:
- `true` — production Let's Encrypt (default)
- `staging` — LE staging CA (no rate limits, untrusted cert; use during setup)
- `false` — HTTP only

Two ACME resolvers (`letsencrypt` and `letsencrypt-staging`) are always present in the managed Traefik compose.

## Managed Traefik

Traefik runs as a Compose stack at `~/.config/atrisos/traefik/`. Atrisos generates and owns those files — they should not be edited by hand. Started automatically on first `atrisos up`. Controlled via `atrisos traefik up/down/restart/status/logs`.

Podman socket path is platform-specific:
- Linux: `/run/user/<UID>/podman/podman.sock`
- macOS: queried from `podman machine inspect`

## Documentation site

`docs/` is a self-contained GitHub Pages app (React, no build step). Content lives in `docs/content/` as Markdown with YAML frontmatter. `docs/build.mjs` scans content and writes `nav.json` (navigation tree) and `templates.json` (template catalog grid).

Run `make docs` after adding or renaming pages. The `docs.yml` workflow deploys automatically on push to `main`.

## Release

Use the `/release` skill to create a release. It handles pre-flight checks, version bump, release notes, tag creation, and post-release cleanup.

Manual process: tag `v*` → GitHub Actions runs goreleaser → publishes four binaries to GitHub Releases:
- `atrisos-linux-amd64`, `atrisos-linux-arm64`
- `atrisos-darwin-amd64`, `atrisos-darwin-arm64`

`CGO_ENABLED=0` — all builds are fully static.

`cmd.Version` is set by goreleaser ldflags. When building locally it defaults to `dev` unless overridden: `make build VERSION=v0.3.0`.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Stack not found |
| 3 | Stack validation error |
| 4 | Podman not found or not running |
| 5 | Traefik error |
