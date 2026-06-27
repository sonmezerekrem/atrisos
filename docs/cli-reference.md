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

Start a stack. If Traefik is not running, starts it first.

```sh
atrisos up myapp
atrisos up myapp --pull   # pull latest images before starting
atrisos up myapp --build  # build images from Dockerfile before starting
```

### `atrisos down <stack>`

Stop a stack and remove its containers. Volumes are preserved.

```sh
atrisos down myapp
atrisos down myapp --volumes  # also remove named volumes (destructive)
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
atrisos update myapp --pull   # explicit pull before recreate (default behavior)
atrisos update myapp --no-pull  # recreate without pulling (e.g. to apply .env changes)
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

Interactive wizard to create a new stack directory with boilerplate `compose.yml`, `.env`, and `config.yml`.

```sh
atrisos init              # prompts for name, creates in root dir
atrisos init myapp        # create stack named "myapp" in root dir
atrisos init myapp --dir ./myapp  # create at a specific path
```

The wizard asks:
1. Stack name and description
2. Image name (optional)
3. Domain hostname (optional)
4. Service port
5. Whether to enable backups

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
| `i` | Show stack info / config.yml summary |
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

Print version, build info, and detected platform.

```sh
atrisos version
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
