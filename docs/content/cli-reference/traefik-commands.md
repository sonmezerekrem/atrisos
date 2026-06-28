---
title: Traefik Commands
section: cli-reference
group: Traefik
order: 1
icon: folder
description: Manage the shared Traefik reverse proxy instance.
---

# Traefik Commands

## `atrisos traefik up`

Start the managed Traefik instance.

```sh
atrisos traefik up
```

## `atrisos traefik down`

Stop the managed Traefik instance. Running stacks will lose routing but containers keep running.

```sh
atrisos traefik down
```

## `atrisos traefik restart`

Restart Traefik (e.g. after changing global Traefik settings).

```sh
atrisos traefik restart
```

## `atrisos traefik status`

Show Traefik container status and a summary of active routers.

```sh
atrisos traefik status
```

## `atrisos traefik logs`

Tail Traefik logs.

```sh
atrisos traefik logs
atrisos traefik logs --lines 200
```

## `atrisos traefik dashboard`

Open the Traefik dashboard in the default browser (or print the URL).

```sh
atrisos traefik dashboard
```
