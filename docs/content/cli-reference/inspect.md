---
title: Inspect & Debug
section: cli-reference
group: Stack Commands
order: 2
icon: info
description: Render merged compose, logs, status, validation, and outdated checks.
---

# Inspect & Debug

## `atrisos render <stack>`

Print the merged compose document that atrisos would pass to `podman compose` — the original `compose.yml` with Traefik labels and `atrisos_net` injected. Useful for debugging routing configuration without actually starting anything.

```sh
atrisos render myapp           # print merged compose YAML to stdout
atrisos render myapp --diff    # show diff vs original compose.yml
```

## `atrisos logs <stack>`

Tail logs from all containers in a stack.

```sh
atrisos logs myapp
atrisos logs myapp --service web     # logs from a specific service
atrisos logs myapp --lines 100       # show last N lines (default: 50)
atrisos logs myapp --follow          # keep streaming (default: true if TTY)
atrisos logs myapp --no-follow       # print and exit
atrisos logs myapp --timestamps      # include timestamps
```

## `atrisos status`

Show status of all stacks.

```sh
atrisos status
atrisos status myapp    # detailed status for one stack
```

Output columns: `NAME`, `STATUS`, `CONTAINERS`, `DOMAIN`, `UPDATED`.

## `atrisos ps <stack>`

Show individual containers in a stack (equivalent to `podman compose ps`).

```sh
atrisos ps myapp
```

## `atrisos validate <stack>`

Dry-run validation of a stack's files — checks config.yml schema, compose.yml syntax, domain-to-service cross-references, and presence of `.env`. All errors reported at once, nothing is started.

```sh
atrisos validate myapp
atrisos validate --all    # validate all discovered stacks
```

## `atrisos outdated [stack]`

Check whether newer image versions are available in the registry for images used by a stack.

```sh
atrisos outdated          # check all stacks
atrisos outdated myapp    # check a specific stack
```
