---
title: CLI Reference
group: Reference
order: 2
icon: info
description: All atrisos CLI commands, flags, and exit codes.
---

# CLI Reference

## Usage

```
atrisos [flags]
atrisos <command> [flags] [args]
```

Running `atrisos` with no arguments launches the TUI.

---

## Global flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config <path>` | `~/.config/atrisos/config.yml` | Path to global config file |
| `--root <path>` | (from config) | Override the stacks root directory for this invocation |
| `--verbose`, `-v` | false | Show verbose output including raw Podman commands |
| `--no-color` | false | Disable color output |
| `--no-emoji` | false | Disable emoji status indicators |

---

## Stack commands

### `atrisos up <stack>`

Start a stack. If Traefik is not running, starts it first. If the stack has `auto_start: true` or `backup.enabled: true`, installs the corresponding systemd/launchd units.

```sh
atrisos up myapp
atrisos up myapp --pull    # pull latest images before starting
atrisos up myapp --build   # build images from Dockerfile before starting
atrisos up --all           # start all discovered stacks sequentially
atrisos up --tag production  # start all stacks tagged "production"
```

### `atrisos down <stack>`

Stop a stack and remove its containers. Volumes are preserved. Removes any installed systemd/launchd units for that stack.

```sh
atrisos down myapp
atrisos down myapp --volumes  # also remove named volumes (destructive)
atrisos down --all            # stop all discovered stacks sequentially
atrisos down --tag production # stop all stacks tagged "production"
```

### `atrisos restart <stack>`

Stop and start a stack.

```sh
atrisos restart myapp
```

### `atrisos update <stack>`

Pull latest images and recreate containers with zero additional downtime where possible.

```sh
atrisos update myapp
atrisos update myapp --pull     # explicit pull before recreate (default)
atrisos update myapp --no-pull  # recreate without pulling (e.g. to apply .env changes)
atrisos update --all            # update all discovered stacks sequentially
atrisos update --tag production # update all stacks tagged "production"
```

### `atrisos render <stack>`

Print the merged compose document that atrisos would pass to `podman compose` — the original `compose.yml` with Traefik labels and `atrisos_net` injected. Useful for debugging routing configuration without actually starting anything.

```sh
atrisos render myapp           # print merged compose YAML to stdout
atrisos render myapp --diff    # show diff vs original compose.yml
```

### `atrisos watch <stack>`

Start a stack in watch mode: monitors the stack directory for changes to `compose.yml`, `.env`, and `config.yml`, and automatically re-applies changes. Stays in foreground; Ctrl+C stops watching (stack keeps running).

```sh
atrisos watch myapp
```

### `atrisos logs <stack>`

Tail logs from all containers in a stack.

```sh
atrisos logs myapp
atrisos logs myapp --service web     # logs from a specific service
atrisos logs myapp --lines 100       # show last N lines (default: 50)
atrisos logs myapp --follow          # keep streaming (default: true if TTY)
atrisos logs myapp --no-follow       # print and exit
atrisos logs myapp --timestamps      # include timestamps
```

### `atrisos status`

Show status of all stacks.

```sh
atrisos status
atrisos status myapp    # detailed status for one stack
```

Output columns: `NAME`, `STATUS`, `CONTAINERS`, `DOMAIN`, `UPDATED`.

### `atrisos ps <stack>`

Show individual containers in a stack (equivalent to `podman compose ps`).

```sh
atrisos ps myapp
```

### `atrisos backup <stack>`

Manually trigger a backup for a stack regardless of its configured schedule. Backs up the named volumes listed in `backup.volumes` (or all named volumes if the list is empty) using the bundled restic binary.

```sh
atrisos backup myapp
atrisos backup myapp --dry-run   # show what would be backed up without running
```

Scheduled backups are triggered automatically by systemd timers (Linux) or launchd plists (macOS) installed when the stack is started. This command is for on-demand runs.

### `atrisos exec <stack> <service> -- <command>`

Run a one-off command inside a running container. Wraps `podman exec`.

```sh
atrisos exec myapp web -- ls /app
atrisos exec myapp db -- psql -U postgres
```

### `atrisos shell <stack> <service>`

Open an interactive shell in a running container. Tries `/bin/bash` first, falls back to `/bin/sh`.

```sh
atrisos shell myapp web
atrisos shell myapp db
```

### `atrisos validate <stack>`

Dry-run validation of a stack's files — checks config.yml schema, compose.yml syntax, domain-to-service cross-references, and presence of `.env`. All errors reported at once, nothing is started.

```sh
atrisos validate myapp
atrisos validate --all    # validate all discovered stacks
```

### `atrisos outdated [stack]`

Check whether newer image versions are available in the registry for images used by a stack.

```sh
atrisos outdated          # check all stacks
atrisos outdated myapp    # check a specific stack
```

### `atrisos export <stack>`

Package a stack into a portable `.tar.gz` containing `compose.yml`, `config.yml`, `compose.override.yml` (if present), and `.env.example`. The `.env` file is excluded to avoid exporting secrets.

```sh
atrisos export myapp                  # creates myapp.tar.gz in current dir
atrisos export myapp --output ~/backups/myapp.tar.gz
```

### `atrisos import <file>`

Extract a stack archive into the stacks root directory and prompt the user to create `.env` from `.env.example`.

```sh
atrisos import myapp.tar.gz
atrisos import myapp.tar.gz --dir /opt/stacks   # extract to a specific location
```

---

## Stack management commands

### `atrisos list`

List all discovered stacks (from root dir and registered paths).

```sh
atrisos list
atrisos list --format table     # default
atrisos list --format json      # JSON output for scripting
atrisos list --format plain     # one name per line
```

### `atrisos register <path>`

Register a stack directory that lives outside the root directory.

```sh
atrisos register ./myapp
atrisos register /opt/services/monitoring
```

The path is stored in `~/.config/atrisos/registry.json`.

### `atrisos unregister <stack>`

Remove a stack from the registry. Does not delete any files.

```sh
atrisos unregister myapp
```

### `atrisos init [name]`

Interactive wizard to create a new stack directory from a template. Templates are fetched from the `templates/` directory in the atrisos GitHub repository (`main` branch) and cached locally in `~/.config/atrisos/templates-cache/`. Subsequent calls use the cache and work offline; atrisos checks for updates when online.

```sh
atrisos init                          # prompts for name and template, creates in root dir
atrisos init myapp                    # create stack named "myapp", prompts for template
atrisos init myapp --template basic   # skip template prompt, use "basic" template
atrisos init myapp --dir ./myapp      # create at a specific path instead of root dir
atrisos init --list-templates         # list available templates without creating anything
```

The wizard asks:
1. Stack name and description
2. Template to use (lists available templates with descriptions)
3. Template-specific prompts (e.g. image name, domain, port)
4. Whether to enable backups

See [templates.md](templates.md) for available templates and how to contribute new ones.

---

## TUI command

### `atrisos tui`

Launch the interactive TUI dashboard. Same as running bare `atrisos`.

```sh
atrisos tui
```

### TUI keyboard shortcuts

| Key | Action |
|-----|--------|
| `↑` / `↓` or `j` / `k` | Navigate stack list |
| `Enter` | Select / focus stack detail |
| `u` | Update selected stack |
| `r` | Restart selected stack |
| `s` | Start selected stack |
| `x` | Stop selected stack |
| `l` | Open log viewer for selected stack |
| `e` | Open shell in the focused service (suspends TUI, returns on exit) |
| `i` | Show stack info / config.yml summary |
| `o` | Show `atrisos outdated` results for selected stack |
| `t` | Show Traefik status panel |
| `/` | Filter stacks by name or tag |
| `?` | Show help |
| `q` or `Ctrl+C` | Quit |

---

## Traefik commands

### `atrisos traefik up`

Start the managed Traefik instance.

```sh
atrisos traefik up
```

### `atrisos traefik down`

Stop the managed Traefik instance. Running stacks will lose routing but containers keep running.

```sh
atrisos traefik down
```

### `atrisos traefik restart`

Restart Traefik (e.g. after changing global Traefik settings).

```sh
atrisos traefik restart
```

### `atrisos traefik status`

Show Traefik container status and a summary of active routers.

```sh
atrisos traefik status
```

### `atrisos traefik logs`

Tail Traefik logs.

```sh
atrisos traefik logs
atrisos traefik logs --lines 200
```

### `atrisos traefik dashboard`

Open the Traefik dashboard in the default browser (or print the URL).

```sh
atrisos traefik dashboard
```

---

## Version and meta

### `atrisos version`

Print version, build info, detected platform, and latest available version (checked in background, cached 24 hours).

```sh
atrisos version
```

### `atrisos self-update`

Download the latest atrisos release from GitHub, verify its checksum, and replace the running binary in place.

```sh
atrisos self-update
atrisos self-update --version v0.5.0   # pin to a specific version
```

### `atrisos completion <shell>`

Generate shell completion scripts.

```sh
atrisos completion bash   > /etc/bash_completion.d/atrisos
atrisos completion zsh    > ~/.zsh/completions/_atrisos
atrisos completion fish   > ~/.config/fish/completions/atrisos.fish
```

---

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Stack not found |
| 3 | Stack validation error (bad config.yml or compose.yml) |
| 4 | Podman not found or not running |
| 5 | Traefik error |

---

## Environment variables

All global flags can also be set via environment variables (flag takes precedence):

| Variable | Equivalent flag |
|----------|----------------|
| `ATRISOS_CONFIG` | `--config` |
| `ATRISOS_ROOT` | `--root` |
| `ATRISOS_NO_COLOR` | `--no-color` |
| `ATRISOS_VERBOSE` | `--verbose` |
| `NO_COLOR` | `--no-color` (standard) |
