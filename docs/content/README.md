# Documentation content

Place Markdown files here. Run `make docs` from the repo root to regenerate `nav.json`.

## Frontmatter

Each page should start with YAML frontmatter:

```yaml
---
title: Page Title
group: Introduction
order: 1
icon: info
description: Short summary shown under the page title.
---
```

| Field | Required | Description |
|-------|----------|-------------|
| `title` | no | Page title (defaults to first `#` heading) |
| `group` | no | Sidebar group name (default: `Documentation`) |
| `order` | no | Sort order within the group |
| `icon` | no | Sidebar icon key (Hugeicons stroke): `info`, `pin`, `grid`, `folder`, `users`, `flag` |
| `description` | no | Subtitle under the page title |
| `id` | no | URL hash id (default: file path without `.md`) |

Subdirectories are supported — e.g. `agents/create-stack.md` becomes `#agents/create-stack`.

## Local preview

```sh
make docs-serve
# open http://localhost:8080
```
