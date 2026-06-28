---
title: Getting Started
section: overview
group: Introduction
order: 1
icon: info
description: Overview of atrisos — CLI and TUI for Podman Compose stacks with automatic Traefik routing.
---

# Atrisos

A CLI + TUI tool for managing Podman Compose stacks with automatic Traefik routing.

## What it does

Atrisos lets you run self-hosted applications as Compose stacks with a consistent folder layout. It handles:

- Starting/stopping/updating stacks via CLI commands or an interactive TUI
- Daemonless container management via Podman
- Automatic domain routing and TLS via a managed Traefik instance
- Volume backups per stack
- Stack discovery from a root directory and/or registered paths

## Platform support

- macOS (Apple Silicon + Intel)
- Ubuntu / Debian-based Linux

## Tech stack

- **Go** — main binary (CLI + TUI)
- **sh/bash** — install script, hook scripts
- **Podman** — container runtime (daemonless)
- **Traefik** — reverse proxy + TLS (ACME/Let's Encrypt)

## Docs index

| Doc | Contents |
|-----|----------|
| [design.md](design.md) | Architecture, components, project structure |
| [stack-format.md](stack-format.md) | Stack folder layout, config.yml schema |
| [traefik.md](traefik.md) | Managed Traefik setup, domain wiring, TLS |
| [cli-reference/usage.md](cli-reference/usage.md) | CLI usage, flags, and exit codes |
| [global-config.md](global-config.md) | Global Atrisos config file schema |
| [install.md](install.md) | Installation and first-run guide |
| [templates/overview.md](templates/overview.md) | Stack init templates: format, available templates, how to add new ones |
| [agents/create-stack.md](agents/create-stack.md) | AI agent prompt for generating Atrisos stacks — paste into any AI assistant |
