# Documentation content

Place Markdown files here. Run `make docs` from the repo root to regenerate `nav.json`.

## Top-level sections (menu tabs)

| Section | ID | Sidebar groups |
|---------|-----|----------------|
| Overview | `overview` | Introduction, Core Concepts, Configuration |
| Agents | `agents` | Prompts |
| CLI Reference | `cli-reference` | Basics, Stack Commands, Discovery & Init, Interactive, Traefik, Meta |
| Templates | `templates` | Using Templates, Template Format, Contributing Templates |
| Contribution | `contribution` | Getting Started, Guidelines, Process |

Use subdirectories for multi-page sections — e.g. `cli-reference/usage.md` → `#cli-reference/usage`.

## Frontmatter

```yaml
---
title: Page Title
section: cli-reference
group: Stack Commands
order: 1
icon: info
description: Short summary shown under the page title.
---
```

| Field | Required | Description |
|-------|----------|-------------|
| `title` | no | Page title (defaults to first `#` heading) |
| `section` | no | Top tab id (see table above) |
| `group` | no | Sidebar group within the section |
| `order` | no | Sort order within the group |
| `icon` | no | Icon key: `info`, `pin`, `grid`, `folder`, `users`, `flag` |
| `description` | no | Subtitle under the page title |
| `id` | no | URL hash id (default: path without `.md`) |

## Local preview

```sh
make docs-serve
# open http://localhost:8080
```
