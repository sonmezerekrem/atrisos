---
title: Stack Lifecycle
section: cli-reference
group: Stack Commands
order: 1
icon: folder
description: Start, stop, restart, update, and watch stacks.
---

# Stack Lifecycle

## `atrisos up <stack>`

Start a stack. If Traefik is not running, starts it first. If the stack has `auto_start: true` or `backup.enabled: true`, installs the corresponding systemd/launchd units.

```sh
atrisos up myapp
atrisos up myapp --pull    # pull latest images before starting
atrisos up myapp --build   # build images from Dockerfile before starting
atrisos up --all           # start all discovered stacks sequentially
atrisos up --tag production  # start all stacks tagged "production"
```

## `atrisos down <stack>`

Stop a stack and remove its containers. Volumes are preserved. Removes any installed systemd/launchd units for that stack.

```sh
atrisos down myapp
atrisos down myapp --volumes  # also remove named volumes (destructive)
atrisos down --all            # stop all discovered stacks sequentially
atrisos down --tag production # stop all stacks tagged "production"
```

## `atrisos restart <stack>`

Stop and start a stack.

```sh
atrisos restart myapp
```

## `atrisos update <stack>`

Pull latest images and recreate containers with zero additional downtime where possible.

```sh
atrisos update myapp
atrisos update myapp --pull     # explicit pull before recreate (default)
atrisos update myapp --no-pull  # recreate without pulling (e.g. to apply .env changes)
atrisos update --all            # update all discovered stacks sequentially
atrisos update --tag production # update all stacks tagged "production"
```

## `atrisos watch <stack>`

Start a stack in watch mode: monitors the stack directory for changes to `compose.yml`, `.env`, and `config.yml`, and automatically re-applies changes. Stays in foreground; Ctrl+C stops watching (stack keeps running).

```sh
atrisos watch myapp
```
