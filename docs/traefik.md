# Traefik Integration

atrisos manages a single shared Traefik instance that acts as the reverse proxy for all stacks. Users never write Traefik labels by hand — atrisos generates them from each stack's `config.yml`.

---

## How it works

```
Internet / LAN
      │
      ▼
  Traefik container           (atrisos_traefik stack, managed by atrisos)
      │
      │   reads labels from containers on atrisos_net
      │
   ┌──┴──────────────┐
   │   atrisos_net   │        (shared Podman network)
   ├─────────────────┤
   │  myapp-web      │  ← label: traefik.http.routers.myapp.rule=Host(`myapp.example.com`)
   │  blog-ghost     │  ← label: traefik.http.routers.blog.rule=Host(`blog.example.com`)
   │  grafana-web    │  ← label: traefik.http.routers.grafana.rule=Host(`grafana.example.com`)
   └─────────────────┘
```

1. On `atrisos up <stack>`, atrisos reads `config.yml`, generates Traefik labels, and merges them into the compose run arguments.
2. Traefik watches containers on the `atrisos_net` network and picks up labels immediately — no Traefik restart needed.
3. TLS certificates are obtained automatically via ACME (Let's Encrypt) per domain.

---

## Managed Traefik stack

Traefik itself runs as a Compose stack stored at `~/.config/atrisos/traefik/`. atrisos manages this stack internally — users do not edit it directly (though advanced users can inspect it).

### Generated compose.yml (internal)

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
      - /run/user/1000/podman/podman.sock:/var/run/docker.sock:ro
      - letsencrypt:/letsencrypt
    networks:
      - atrisos_net

networks:
  atrisos_net:
    external: true

volumes:
  letsencrypt:
```

Notes:
- Uses the Podman socket (rootless) at `/run/user/<UID>/podman/podman.sock` — path computed at runtime.
- On macOS with `podman machine`, the socket path differs and is computed from `podman machine inspect`.

---

## Label generation

For a stack with this `config.yml`:

```yaml
name: "My App"
domain:
  host: "myapp.example.com"
  service: "web"
  port: 3000
  tls: true
```

atrisos generates and injects these labels on the `web` service at runtime:

```
traefik.enable=true
traefik.http.routers.myapp.rule=Host(`myapp.example.com`)
traefik.http.routers.myapp.entrypoints=websecure
traefik.http.routers.myapp.tls=true
traefik.http.routers.myapp.tls.certresolver=letsencrypt
traefik.http.services.myapp.loadbalancer.server.port=3000
traefik.http.routers.myapp-redirect.rule=Host(`myapp.example.com`)
traefik.http.routers.myapp-redirect.entrypoints=web
traefik.http.routers.myapp-redirect.middlewares=redirect-to-https
traefik.http.middlewares.redirect-to-https.redirectscheme.scheme=https
```

The router name is derived from the stack directory name (sanitized: lowercased, non-alphanumeric chars replaced with `-`).

### Path prefix routing

If `path_prefix` is set (e.g. `/api`), the rule becomes:
```
Host(`myapp.example.com`) && PathPrefix(`/api`)
```

### Custom middlewares

If `domain.middlewares` lists middleware names, they are appended to the router's middleware chain. The middlewares must be defined in Traefik's static or dynamic config — atrisos does not create them.

---

## Network: atrisos_net

atrisos creates a Podman network named `atrisos_net` on first use:

```sh
podman network create atrisos_net
```

All stacks that use Traefik routing must join this network. atrisos automatically adds the network to the compose file at runtime if `domain.host` is set:

```yaml
# injected at runtime into every routed stack
networks:
  default:
    name: atrisos_net
    external: true
```

Stacks without a `domain.host` in `config.yml` are not attached to `atrisos_net` and run in isolation.

---

## TLS / ACME (Let's Encrypt)

- Certificates are obtained via HTTP-01 challenge — port 80 must be reachable from the internet.
- Certificates are stored in a Podman volume (`atrisos_letsencrypt`) and persist across Traefik restarts.
- The ACME email is set globally in `~/.config/atrisos/config.yml` and can be overridden per stack in `config.yml`.

### Rate limits

Let's Encrypt enforces rate limits (5 duplicate cert requests per week). During development:
- Use a subdomain you control rather than a production domain.
- Or temporarily set `tls: false` in `config.yml` to skip ACME.

### Self-signed / local mode (future)

A `local` mode using Traefik's built-in self-signed certs is planned for LAN/homelab use where public ACME is not available.

---

## Traefik dashboard

The Traefik dashboard is enabled but not exposed via a domain by default. To access it:

```sh
atrisos traefik dashboard
# opens http://localhost:8080/dashboard/ via SSH tunnel or direct if local
```

To expose it via a domain, add to `~/.config/atrisos/config.yml`:

```yaml
traefik:
  dashboard:
    enabled: true
    host: "traefik.example.com"
```

---

## Traefik commands

```sh
atrisos traefik up       # start managed Traefik (done automatically on first `atrisos up`)
atrisos traefik down     # stop Traefik
atrisos traefik restart  # restart Traefik
atrisos traefik status   # show Traefik container status
atrisos traefik logs     # tail Traefik logs
atrisos traefik dashboard  # open dashboard
```

---

## Troubleshooting

| Symptom | Likely cause |
|---------|-------------|
| `502 Bad Gateway` | App container not running or wrong port in `config.yml` |
| `404 page not found` (Traefik 404) | No matching router — check `domain.host` matches the request hostname exactly |
| Certificate error in browser | ACME challenge failed — check port 80 is open and domain resolves to this server |
| Labels not picked up | Container not on `atrisos_net` — check atrisos injected the network correctly with `podman inspect <container>` |
| `podman.sock not found` | Podman socket path wrong — run `podman info --format '{{.Host.RemoteSocket.Path}}'` and report as a bug |
