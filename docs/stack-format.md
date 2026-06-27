# Stack Format

Each application is a self-contained directory called a **stack**. atrisos discovers stacks from a root directory and/or registered paths.

## Directory layout

```
myapp/
├── compose.yml        # Podman/Docker Compose file (required)
├── .env               # Environment variables (required, may be empty)
└── config.yml         # atrisos configuration (required)
```

`docker-compose.yml` is also accepted as an alias for `compose.yml`. If both exist, `compose.yml` takes precedence.

---

## compose.yml

Standard Compose format. Write it exactly as you would for `docker compose` — atrisos passes it through to `podman compose` unchanged, with one addition: atrisos injects Traefik labels at runtime based on `config.yml` so you don't write labels by hand.

Minimum example:

```yaml
services:
  web:
    image: nginx:alpine
    expose:
      - "80"
```

Use `expose` (not `ports`) for services you want routed through Traefik. atrisos only injects Traefik labels on the service named as the entry point in `config.yml` (defaults to the first service).

---

## .env

Standard dotenv format. All variables here are available inside `compose.yml` via `${VAR_NAME}`.

```env
APP_ENV=production
SECRET_KEY=changeme
DB_PASSWORD=hunter2
```

Commit a `.env.example` with placeholder values; add `.env` to `.gitignore`.

---

## config.yml

atrisos-specific configuration for the stack. Full schema:

```yaml
# ── Metadata ────────────────────────────────────────────────
name: "My App"                 # display name in TUI (default: directory name)
description: "Short description shown in TUI detail panel"
tags:
  - web
  - production
meta:                           # free-form key-value pairs, shown in TUI
  owner: "backend-team"
  repo: "https://github.com/org/myapp"
  docs: "https://wiki.internal/myapp"
  # any additional keys you find useful

# ── Domain routing (Traefik) ────────────────────────────────
domain:
  host: "myapp.example.com"    # required for Traefik routing
  path_prefix: "/"             # optional, default "/"
  service: "web"               # which compose service to route to (default: first service)
  port: 80                     # container port to forward to (default: first exposed port)
  tls: true                    # enable HTTPS via ACME (default: true)
  acme_email: ""               # override global acme_email for this stack (optional)
  middlewares: []              # list of Traefik middleware names (advanced, optional)

# ── Update behavior ─────────────────────────────────────────
update:
  mode: "manual"               # "manual" | "watch" — overrides global default
                               # manual: user runs `atrisos update <stack>`
                               # watch: auto-restart on file changes in stack dir

# ── Backup ──────────────────────────────────────────────────
backup:
  enabled: false
  schedule: "0 2 * * *"       # cron syntax — when to run backups
  destination: "~/backups/myapp"  # local path or s3://bucket/prefix
  volumes:                     # which named volumes to back up (default: all)
    - myapp_data
```

### Minimal config.yml (no Traefik routing)

```yaml
name: "My App"
description: "Internal tool, no public domain needed"
```

### config.yml with domain routing

```yaml
name: "Ghost Blog"
description: "Personal blog"
tags:
  - web
  - blog
meta:
  owner: "ekrem"

domain:
  host: "blog.example.com"
  service: "ghost"
  port: 2368
  tls: true

update:
  mode: watch
```

---

## Validation rules

atrisos validates config.yml on load and reports errors clearly:

| Field | Rule |
|-------|------|
| `domain.host` | Must be a valid hostname. If `tls: true`, must be a real public domain (ACME won't issue certs for IPs or `.local`). |
| `domain.service` | Must match a service name defined in `compose.yml`. |
| `domain.port` | Must be in the range 1–65535. |
| `backup.schedule` | Must be valid cron syntax (5-field). |
| `backup.destination` | Must be a valid local path or `s3://` URI. |
| `meta` | Values must be strings. |

---

## Example: full stack

```
myapp/
├── compose.yml
├── .env
├── .env.example
└── config.yml
```

**compose.yml**
```yaml
services:
  web:
    image: ghcr.io/myorg/myapp:latest
    expose:
      - "3000"
    environment:
      - DATABASE_URL=${DATABASE_URL}
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

**.env**
```env
DATABASE_URL=postgresql://postgres:${DB_PASSWORD}@db:5432/myapp
DB_PASSWORD=supersecret
```

**config.yml**
```yaml
name: "My App"
description: "Main web application"
tags:
  - web
  - production
meta:
  owner: "backend-team"

domain:
  host: "myapp.example.com"
  service: "web"
  port: 3000
  tls: true

update:
  mode: manual

backup:
  enabled: true
  schedule: "0 3 * * *"
  destination: "~/backups/myapp"
  volumes:
    - db_data
```
