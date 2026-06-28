---
title: Overview
section: templates
group: Using Templates
order: 1
icon: info
description: How atrisos init fetches and caches stack templates.
---

# Templates Overview

`atrisos init` creates new stacks from templates. Templates live in the `templates/` directory of the Atrisos repository and are fetched from GitHub at runtime.

## How templates are fetched

1. On `atrisos init`, Atrisos checks `~/.config/atrisos/templates-cache/manifest.json` for a cached version.
2. If online, Atrisos compares the local manifest against the remote one (fetched from `raw.githubusercontent.com/sonmezerekrem/atrisos/main/templates/manifest.json`). If the remote is newer, the full template set is re-downloaded.
3. If offline (or GitHub is unreachable), the local cache is used as-is.
4. On first run with no cache, Atrisos fetches all templates before presenting the wizard.

The cache is stored at `~/.config/atrisos/templates-cache/`.

## Forcing a cache refresh

```sh
atrisos init --refresh-templates   # re-download all templates from GitHub then run wizard
```

## Quick start

```sh
atrisos init myapp                  # interactive wizard
atrisos init myapp --template basic # skip template prompt
atrisos init --list-templates       # list available templates
```

See **Available Templates** for the built-in catalog and **Template Format** for how templates are structured.
