# atrisos

Manage Podman Compose stacks with automatic Traefik routing and TLS. Write a plain `compose.yml` with no Traefik content — atrisos injects routing labels at runtime.

- **Zero Traefik config** — domains go in `config.yml`, atrisos handles labels
- **Auto TLS** — Let's Encrypt certs issued on first request (staging CA available)
- **TUI dashboard** — interactive stack list, log viewer, container health
- **No daemon** — reads state live from Podman, delegates scheduling to systemd/launchd
- macOS (Apple Silicon + Intel) and Ubuntu/Debian (amd64 + arm64)

---

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/sonmezerekrem/atrisos/main/scripts/install.sh | sh
```

**Requirements**:
- **macOS**: `brew install podman`
- **Linux**: `sudo apt install podman` (Ubuntu 22.04+ / Debian 12+)

---

## Quick start

```sh
atrisos          # first run — setup wizard (ACME email, stacks root)
atrisos init myapp           # create a stack from a template
atrisos up myapp             # start it (Traefik starts automatically)
atrisos tui                  # open the dashboard
```

Stack directory layout:

```
myapp/
├── compose.yml      # standard Compose file — no Traefik content
├── .env             # environment variables
└── config.yml       # atrisos config: name, domains, backup, notify
```

`config.yml` example:

```yaml
name: "My App"

domains:
  - service: web
    host: myapp.example.com
    port: 3000
    tls: true      # true | staging | false
```

atrisos merges Traefik labels into the compose document at runtime. Your `compose.yml` is never modified.

---

## Commands

```sh
atrisos up <stack>           # start stack (Traefik auto-starts)
atrisos down <stack>         # stop stack
atrisos restart <stack>      # restart
atrisos update <stack>       # pull latest images and recreate
atrisos logs <stack>         # tail logs
atrisos status               # status of all stacks
atrisos render <stack>       # print merged compose YAML (debug routing)
atrisos validate <stack>     # dry-run config + compose validation
atrisos outdated             # check for newer image versions
atrisos backup <stack>       # manual restic backup of named volumes
atrisos traefik status       # Traefik container status
atrisos self-update          # update atrisos binary in place
```

Bulk operations: `--all` or `--tag <tag>` on `up`, `down`, `update`.

---

## Documentation

- [Stack format & config.yml schema](docs/stack-format.md)
- [Traefik integration & compose merge pipeline](docs/traefik.md)
- [CLI reference](docs/cli-reference.md)
- [Global config](docs/global-config.md)
- [Installation & first run](docs/install.md)
- [Init templates](docs/templates.md)

---

## Build from source

```sh
git clone https://github.com/sonmezerekrem/atrisos
cd atrisos
make build          # ./atrisos
make install        # /usr/local/bin/atrisos
go test ./internal/...
```
