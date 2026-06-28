---
title: Maintenance
section: cli-reference
group: Stack Commands
order: 3
icon: folder
description: Backup, exec, shell, export, and import commands.
---

# Maintenance

## `atrisos backup <stack>`

Manually trigger a backup for a stack regardless of its configured schedule. Backs up the named volumes listed in `backup.volumes` (or all named volumes if the list is empty) using the bundled restic binary.

```sh
atrisos backup myapp
atrisos backup myapp --dry-run   # show what would be backed up without running
```

Scheduled backups are triggered automatically by systemd timers (Linux) or launchd plists (macOS) installed when the stack is started. This command is for on-demand runs.

## `atrisos exec <stack> <service> -- <command>`

Run a one-off command inside a running container. Wraps `podman exec`.

```sh
atrisos exec myapp web -- ls /app
atrisos exec myapp db -- psql -U postgres
```

## `atrisos shell <stack> <service>`

Open an interactive shell in a running container. Tries `/bin/bash` first, falls back to `/bin/sh`.

```sh
atrisos shell myapp web
atrisos shell myapp db
```

## `atrisos export <stack>`

Package a stack into a portable `.tar.gz` containing `compose.yml`, `config.yml`, `compose.override.yml` (if present), and `.env.example`. The `.env` file is excluded to avoid exporting secrets.

```sh
atrisos export myapp                  # creates myapp.tar.gz in current dir
atrisos export myapp --output ~/backups/myapp.tar.gz
```

## `atrisos import <file>`

Extract a stack archive into the stacks root directory and prompt the user to create `.env` from `.env.example`.

```sh
atrisos import myapp.tar.gz
atrisos import myapp.tar.gz --dir /opt/stacks   # extract to a specific location
```
