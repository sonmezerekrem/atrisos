---
title: Getting Started
section: contribution
group: Getting Started
order: 1
icon: pin
description: Development setup and project overview for contributors.
---

# Getting Started

Thank you for your interest in contributing. Atrisos is an open-source project — bug reports, documentation improvements, and new stack templates are all welcome.

## Development setup

```sh
git clone https://github.com/sonmezerekrem/atrisos
cd atrisos
make build                       # ./atrisos
cd app && go test ./internal/... # run tests
```

Go 1.26+ required.

## What you can contribute

- **Code** — bug fixes and features in `app/` (Go CLI + TUI)
- **Documentation** — pages in `docs/content/`
- **Stack templates** — new `atrisos init` templates in `templates/`
- **Issues & PRs** — reports, discussions, and pull requests

See the other pages in this section for guidelines on each area.
