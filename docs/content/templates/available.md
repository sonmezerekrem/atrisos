---
title: Available Templates
section: templates
group: Using Templates
order: 2
icon: folder
description: Built-in stack init templates shipped with atrisos.
---

# Available Templates

Browse built-in init templates below. Each card shows the template name, description, and the `--template` id used with `atrisos init`.

## Using a specific template

```sh
atrisos init myapp --template wordpress
```

The wizard still prompts for template-specific values (domains, passwords, ports) with sensible auto-generated defaults.

## Listing templates locally

```sh
atrisos init --list-templates
```

This reads from the local cache (or fetches first if empty).
