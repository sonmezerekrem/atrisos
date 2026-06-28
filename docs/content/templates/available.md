---
title: Available Templates
section: templates
group: Using Templates
order: 2
icon: folder
description: Built-in stack init templates shipped with atrisos.
---

# Available Templates

These templates are defined in `templates/manifest.json` in the repository.

| Template | Description |
|----------|-------------|
| `basic` | Single service with optional domain routing |
| `webapp-postgres` | Web service backed by PostgreSQL with backup enabled |
| `postgres` | Standalone PostgreSQL 16 Alpine database with scheduled backups |
| `valkey` | Valkey 8.1 key-value store (Redis-compatible) with persistence |
| `mongo` | MongoDB 8 document database with scheduled backups |
| `wordpress` | WordPress with MySQL 8 and automatic TLS |
| `registry` | Private container image registry (registry:2) with TLS via Traefik |

## Using a specific template

```sh
atrisos init myapp --template webapp-postgres
```

The wizard still prompts for template-specific values (domains, passwords, ports) with sensible auto-generated defaults.

## Listing templates locally

```sh
atrisos init --list-templates
```

This reads from the local cache (or fetches first if empty).
