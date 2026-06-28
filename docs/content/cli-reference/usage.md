---
title: Usage & Flags
section: cli-reference
group: Basics
order: 1
icon: info
description: CLI invocation, global flags, exit codes, and environment variables.
---

# Usage & Flags

## Usage

```
atrisos [flags]
atrisos <command> [flags] [args]
```

Running `atrisos` with no arguments launches the TUI.

## Global flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config <path>` | `~/.config/atrisos/config.yml` | Path to global config file |
| `--root <path>` | (from config) | Override the stacks root directory for this invocation |
| `--verbose`, `-v` | false | Show verbose output including raw Podman commands |
| `--no-color` | false | Disable color output |
| `--no-emoji` | false | Disable emoji status indicators |

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Stack not found |
| 3 | Stack validation error (bad config.yml or compose.yml) |
| 4 | Podman not found or not running |
| 5 | Traefik error |

## Environment variables

All global flags can also be set via environment variables (flag takes precedence):

| Variable | Equivalent flag |
|----------|----------------|
| `ATRISOS_CONFIG` | `--config` |
| `ATRISOS_ROOT` | `--root` |
| `ATRISOS_NO_COLOR` | `--no-color` |
| `ATRISOS_VERBOSE` | `--verbose` |
| `NO_COLOR` | `--no-color` (standard) |
