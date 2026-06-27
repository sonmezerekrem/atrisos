# Traefik Integration

atrisos manages a single shared Traefik instance that routes traffic to all stacks. Users write zero Traefik-related content in their `compose.yml` — domain configuration lives exclusively in `config.yml`, and atrisos generates everything else.

---

## How it works

```
Internet / LAN
      │
      ▼
  Traefik container           (managed by atrisos at ~/.config/atrisos/traefik/)
      │
      │   reads labels from containers on atrisos_net
      │
   ┌──┴──────────────────────────────────────────────────┐
   │                    atrisos_net                       │
   ├──────────────────────────────────────────────────────┤
   │  myapp-web    ← Host(`myapp.example.com`), port 3000 │
   │  myapp-api    ← Host(`api.example.com`),  port 8080  │
   │  blog-ghost   ← Host(`blog.example.com`), port 2368  │
   │  myapp-db     (no domain entry — not on this network)│
   └──────────────────────────────────────────────────────┘
```

1. User runs `atrisos up myapp`.
2. atrisos reads `compose.yml` and `config.yml`, builds a merged compose document in memory, and passes it to `podman compose`. The original `compose.yml` is never written to.
3. Traefik detects new containers on `atrisos_net` via the Podman socket and picks up their labels immediately — no Traefik restart.
4. ACME certificates are issued per domain on first HTTPS request.

---

## Compose merge pipeline

This is the core mechanism that keeps user compose files clean.

### Input

`compose.yml` (user-authored, no Traefik content):
```yaml
services:
  web:
    image: myorg/app:latest
    environment:
      - DATABASE_URL=${DATABASE_URL}
  db:
    image: postgres:16-alpine
    volumes:
      - db_data:/var/lib/postgresql/data
volumes:
  db_data:
```

`config.yml` domains section:
```yaml
domains:
  - service: "web"
    host: "myapp.example.com"
    port: 3000
    tls: true
```

### What atrisos generates in memory

atrisos deep-merges the following additions into the parsed compose structure:

```yaml
services:
  web:
    # everything from the original web service, plus:
    networks:
      - default          # original default network preserved
      - atrisos_net      # injected: shared Traefik network
    labels:
      # injected: all Traefik routing labels (see Label generation below)
      traefik.enable: "true"
      traefik.http.routers.myapp-web.rule: "Host(`myapp.example.com`)"
      traefik.http.routers.myapp-web.entrypoints: "websecure"
      traefik.http.routers.myapp-web.tls: "true"
      traefik.http.routers.myapp-web.tls.certresolver: "letsencrypt"
      traefik.http.services.myapp-web.loadbalancer.server.port: "3000"
      traefik.http.routers.myapp-web-http.rule: "Host(`myapp.example.com`)"
      traefik.http.routers.myapp-web-http.entrypoints: "web"
      traefik.http.routers.myapp-web-http.middlewares: "https-redirect"
  db:
    # untouched — no domain entry for this service

networks:
  default: {}              # original default network preserved
  atrisos_net:             # injected: external network declaration
    external: true

volumes:
  db_data: {}              # unchanged
```

The `db` service is untouched because it has no `domains` entry in `config.yml`. It remains on the stack's default internal network only.

### Execution

The merged document is written to a temporary file (e.g. `/tmp/atrisos-myapp-<hash>.yml`) and passed to `podman compose`:

```sh
podman compose -f /tmp/atrisos-myapp-<hash>.yml --project-name myapp up -d
```

The temp file is deleted after `podman compose` exits. The user's `compose.yml` is never touched.

To inspect the merged compose document without running it:

```sh
atrisos render myapp          # print merged compose YAML to stdout
atrisos render myapp --diff   # show diff between original compose.yml and merged output
```

---

## Label generation

For each entry in `config.yml`'s `domains` array, atrisos generates a set of Traefik labels. The router name is `<stack-dir-name>-<service-name>` (lowercased, non-alphanumeric chars replaced with `-`).

### HTTPS (tls: true, default)

Input:
```yaml
domains:
  - service: "web"
    host: "myapp.example.com"
    port: 3000
    tls: true
```

Generated labels on the `web` container:
```
traefik.enable=true

# HTTPS router
traefik.http.routers.myapp-web.rule=Host(`myapp.example.com`)
traefik.http.routers.myapp-web.entrypoints=websecure
traefik.http.routers.myapp-web.tls=true
traefik.http.routers.myapp-web.tls.certresolver=letsencrypt

# Service (load balancer target)
traefik.http.services.myapp-web.loadbalancer.server.port=3000

# HTTP → HTTPS redirect router
traefik.http.routers.myapp-web-http.rule=Host(`myapp.example.com`)
traefik.http.routers.myapp-web-http.entrypoints=web
traefik.http.routers.myapp-web-http.middlewares=https-redirect
```

The `https-redirect` middleware is defined globally in the managed Traefik config — atrisos creates it once.

### HTTP only (tls: false)

Input:
```yaml
domains:
  - service: "web"
    host: "myapp.example.com"
    port: 3000
    tls: false
```

Generated labels:
```
traefik.enable=true
traefik.http.routers.myapp-web.rule=Host(`myapp.example.com`)
traefik.http.routers.myapp-web.entrypoints=web
traefik.http.services.myapp-web.loadbalancer.server.port=3000
```

### Path prefix routing

Input:
```yaml
domains:
  - service: "api"
    host: "myapp.example.com"
    port: 8080
    path_prefix: "/api"
```

The rule becomes:
```
traefik.http.routers.myapp-api.rule=Host(`myapp.example.com`) && PathPrefix(`/api`)
```

### Multiple domains on one stack

Each `domains` entry generates its own independent router. Two entries pointing to different services:

```yaml
domains:
  - service: "web"
    host: "myapp.example.com"
    port: 3000
  - service: "api"
    host: "api.example.com"
    port: 8080
```

Produces labels on `web`:
```
traefik.http.routers.myapp-web.rule=Host(`myapp.example.com`)
traefik.http.services.myapp-web.loadbalancer.server.port=3000
...
```

And labels on `api`:
```
traefik.http.routers.myapp-api.rule=Host(`api.example.com`)
traefik.http.services.myapp-api.loadbalancer.server.port=8080
...
```

### Two domains pointing to the same service

```yaml
domains:
  - service: "web"
    host: "myapp.example.com"
    port: 3000
  - service: "web"
    host: "www.myapp.example.com"
    port: 3000
```

Generates two routers (`myapp-web-0`, `myapp-web-1`) both pointing to the same service backend. The service entry is deduplicated.

---

## Network: atrisos_net

atrisos creates and owns a Podman network named `atrisos_net` (configurable in global config):

```sh
podman network create atrisos_net
```

Only services with at least one entry in `config.yml`'s `domains` array are attached to `atrisos_net`. Services with no domain entry (databases, caches, workers) run only on the stack's internal default network, which is the correct isolation posture.

---

## Managed Traefik stack

Traefik runs as a Compose stack at `~/.config/atrisos/traefik/`. atrisos manages it — do not edit these files directly.

### Internal compose.yml

```yaml
services:
  traefik:
    image: traefik:v3
    command:
      - "--providers.docker=true"
      - "--providers.docker.network=atrisos_net"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge=true"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
      - "--certificatesresolvers.letsencrypt.acme.email=${ACME_EMAIL}"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
      - "--api.dashboard=true"
      - "--log.level=INFO"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ${PODMAN_SOCKET}:/var/run/docker.sock:ro
      - letsencrypt:/letsencrypt
    networks:
      - atrisos_net

networks:
  atrisos_net:
    external: true

volumes:
  letsencrypt:
```

`PODMAN_SOCKET` is computed at runtime:
- **Linux**: `/run/user/<UID>/podman/podman.sock`
- **macOS**: path from `podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}'`

---

## TLS / ACME

- Certificates are issued via the HTTP-01 challenge — port 80 must be reachable from the internet.
- Certs are stored in the `atrisos_letsencrypt` Podman volume and persist across Traefik restarts.
- ACME email is read from the global config (`traefik.acme_email`) and can be overridden per-domain entry with a `acme_email` field.

**Rate limits**: Let's Encrypt allows 5 duplicate cert requests per week per domain. Use a staging domain or set `tls: false` during development.

---

## Traefik dashboard

Disabled externally by default. To access it locally:

```sh
atrisos traefik dashboard   # prints URL: http://localhost:8080/dashboard/
```

To expose it via a domain, set in global config:

```yaml
traefik:
  dashboard:
    enabled: true
    host: "traefik.example.com"
```

---

## Traefik commands

```sh
atrisos traefik up         # start (runs automatically on first `atrisos up`)
atrisos traefik down       # stop
atrisos traefik restart    # restart
atrisos traefik status     # container status + active router summary
atrisos traefik logs       # tail Traefik logs
atrisos traefik dashboard  # print dashboard URL
```

---

## Troubleshooting

| Symptom | Likely cause |
|---------|-------------|
| `502 Bad Gateway` | App container not running or wrong `port` in `config.yml` |
| `404 page not found` (Traefik default) | Router rule mismatch — check `host` exactly matches the request hostname |
| Certificate error in browser | ACME HTTP-01 failed — check port 80 is reachable and the domain resolves to this host |
| Labels not picked up | Service not attached to `atrisos_net` — run `atrisos render myapp` to verify the merged compose includes the network |
| `podman.sock not found` | Socket path detection failed — run `podman info --format '{{.Host.RemoteSocket.Path}}'` and file a bug |
| Multiple domains, wrong service gets traffic | Check router names with `atrisos traefik status` — name collision between stacks is possible if two stacks share a service name; use unique stack directory names |
