---
title: Code & Docs
section: contribution
group: Guidelines
order: 1
icon: info
description: Guidelines for code changes and documentation updates.
---

# Code & Docs Guidelines

## Code

- Keep changes focused; match existing patterns in the package you're editing
- Run `make lint` and `cd app && go test ./internal/...` before opening a PR
- No background daemon — stack state is read live from Podman
- User `compose.yml` files must never be modified by atrisos (merge pipeline writes temp files only)

For architecture context, see **Design** and **Stack Format** in the Overview section.

## Documentation

Docs live in `docs/content/`. Each page uses YAML frontmatter:

```yaml
---
title: Page Title
section: overview          # overview | agents | cli-reference | templates | contribution
group: Introduction
order: 1
icon: info
description: Short summary.
---
```

After adding or editing a page:

```sh
make docs        # regenerate nav.json
make docs-serve  # preview at http://localhost:8080
```

Split large topics into multiple pages within a section rather than one long page. Use `group` to organize the sidebar.

## Stack templates

New init templates go in `templates/<name>/` with a `manifest.json` entry. See the **Templates** section for format and submission steps.
