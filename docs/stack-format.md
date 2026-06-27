# Stack Format

Each application is a self-contained directory called a **stack**. atrisos discovers stacks from a root directory and/or registered paths.

## Directory layout

```
myapp/
├── compose.yml          # Podman/Docker Compose file (required)
├── compose.override.yml # Optional: merged on top of compose.yml automatically
├── .env                 # Environment variables (required, may be empty)
├── .env.example         # Committed placeholder values (recommended)
└── config.yml           # atrisos configuration (required)
```

`docker-compose.yml` is also accepted as an alias for `compose.yml`. If both exist, `compose.yml` takes precedence.

If `compose.override.yml` exists, atrisos deep-merges it with `compose.yml` before applying Traefik label injection — same semantics as `docker compose` override files. No config is needed; presence of the file is the signal.

---

## compose.yml

Write a standard Compose file describing your services, volumes, and networks — **with no Traefik-related content whatsoever**. No labels, no Traefik network references, no special ports for routing. atrisos handles all of that by merging your compose file with the `domains` config at runtime before invoking `podman compose`.

```yaml
services:
  web:
    image: nginx:alpine

  api:
    image: myorg/api:latest
    environment:
      - DB_URL=${DB_URL}
    depends_on:
      - db

  db:
    image: postgres:16-alpine
    volumes:
      - db_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_PASSWORD=${DB_PASSWORD}

volumes:
  db_data:
```

Rules for compose.yml:
- Do not add `labels` with `traefik.*` keys — atrisos injects them.
- Do not reference `atrisos_net` in `networks` — atrisos injects it for routed services.
- Do not map host ports (e.g. `ports: - "3000:3000"`) for services routed through Traefik — traffic goes through the shared network.
- Internal service-to-service communication works normally via Compose's default network.

---

## .env

Standard dotenv format. Variables are available inside `compose.yml` via `${VAR_NAME}`.

```env
APP_ENV=production
SECRET_KEY=changeme
DB_PASSWORD=hunter2
DB_URL=postgresql://postgres:hunter2@db:5432/myapp
```

Commit a `.env.example` with placeholder values; add `.env` to `.gitignore`.

---

## config.yml

atrisos-specific configuration for the stack.

```yaml
# ── Metadata ────────────────────────────────────────────────
name: "My App"
description: "Short description shown in TUI detail panel"
tags:
  - web
  - production
meta:                           # free-form key-value pairs, displayed in TUI
  owner: "backend-team"
  repo: "https://github.com/org/myapp"
  docs: "https://wiki.internal/myapp"
  # any string key-value pairs

# ── Domain routing ──────────────────────────────────────────
domains:
  - service: "web"              # must match a service name in compose.yml
    host: "myapp.example.com"   # public hostname
    port: 3000                  # container port the service listens on
    path_prefix: "/"            # optional, default "/"
    tls: true                   # optional: true | staging | false (default: true)
                               #   true    → production Let's Encrypt (trusted cert)
                               #   staging → LE staging CA (no rate limits, untrusted cert)
                               #   false   → HTTP only, no certificate
    middlewares: []             # optional Traefik middleware names (advanced)

  - service: "api"
    host: "api.example.com"
    port: 8080

# ── Update behavior ─────────────────────────────────────────
update:
  mode: "manual"               # "manual" | "watch" — overrides global default

# ── Boot ────────────────────────────────────────────────────
auto_start: false              # if true, atrisos installs a systemd/launchd unit
                               # so this stack starts automatically after a reboot

# ── Backup ──────────────────────────────────────────────────
backup:
  enabled: false
  schedule: "0 2 * * *"       # cron syntax — atrisos installs a systemd timer / launchd
                               # plist on `atrisos up` to trigger this automatically
  destination: "~/backups/myapp"
  volumes:
    - db_data

# ── Notifications ────────────────────────────────────────────
notify:
  webhook: "https://ntfy.sh/myapp-alerts"   # POST JSON on: unexpected exit,
                                             # backup failure, cert expiry warning
```

### domains array

Each entry in `domains` maps one hostname (optionally with a path prefix) to one service in your `compose.yml`. You can have:

- **One domain → one service** (most common)
- **Multiple domains → different services** (e.g. frontend on `app.example.com`, API on `api.example.com`)
- **Multiple domains → same service** (e.g. `www.example.com` and `example.com` both pointing to `web`)
- **Path-based routing on the same host** (e.g. `/` → `web`, `/api` → `api` on the same domain)

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `service` | yes | — | Service name from `compose.yml` |
| `host` | yes | — | Public hostname (e.g. `myapp.example.com`) |
| `port` | yes | — | Container port the service listens on |
| `path_prefix` | no | `/` | Route only requests matching this path prefix |
| `tls` | no | `true` | `true` = production LE cert, `staging` = LE staging (no rate limits, untrusted), `false` = HTTP only |
| `middlewares` | no | `[]` | Traefik middleware names to attach to this router |

### Minimal config.yml (no routing)

```yaml
name: "Internal Tool"
description: "No public domain needed"
```

---

## Validation rules

atrisos validates `config.yml` on load and reports errors before attempting to start anything.

| Field | Rule |
|-------|------|
| `domains[*].service` | Must match a service name in `compose.yml` |
| `domains[*].host` | Must be a valid hostname. If `tls: true` or `tls: staging`, must be a public domain (ACME cannot issue certs for bare IPs or `.local`) |
| `domains[*].port` | Integer 1–65535 |
| `domains[*].path_prefix` | Must start with `/` |
| `backup.schedule` | Valid 5-field cron expression |
| `backup.destination` | Valid local path or `s3://` URI |
| `meta` | All values must be strings |
| `notify.webhook` | Must be a valid HTTP/HTTPS URL |

---

## Full example

**compose.yml** — plain services, nothing Traefik-related:
```yaml
services:
  web:
    image: ghcr.io/myorg/myapp:latest
    environment:
      - DATABASE_URL=${DATABASE_URL}
    depends_on:
      - db

  api:
    image: ghcr.io/myorg/myapp-api:latest
    environment:
      - DATABASE_URL=${DATABASE_URL}

  db:
    image: postgres:16-alpine
    volumes:
      - db_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_PASSWORD=${DB_PASSWORD}

volumes:
  db_data:
```

**.env**:
```env
DATABASE_URL=postgresql://postgres:supersecret@db:5432/myapp
DB_PASSWORD=supersecret
```

**config.yml**:
```yaml
name: "My App"
description: "Main web application"
tags:
  - web
  - production
meta:
  owner: "backend-team"
  repo: "https://github.com/myorg/myapp"

domains:
  - service: "web"
    host: "myapp.example.com"
    port: 3000
  - service: "api"
    host: "api.example.com"
    port: 8080

update:
  mode: manual

auto_start: true

backup:
  enabled: true
  schedule: "0 3 * * *"
  destination: "~/backups/myapp"
  volumes:
    - db_data
```

What atrisos does with this before running `podman compose`:
- Injects Traefik labels onto the `web` service for `myapp.example.com:3000`
- Injects Traefik labels onto the `api` service for `api.example.com:8080`
- Attaches `atrisos_net` to `web` and `api` (not `db`, which has no domain entry)
- Adds the `atrisos_net` external network declaration at the top-level `networks` section
- Passes the merged compose document to `podman compose` — your original `compose.yml` is never modified
