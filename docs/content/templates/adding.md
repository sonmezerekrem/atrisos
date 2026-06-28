---
title: Adding Templates
section: templates
group: Contributing Templates
order: 1
icon: pin
description: How to create and submit a new init template.
---

# Adding a New Template

1. Create a new directory under `templates/<name>/`.
2. Add `template.yml` with prompts.
3. Add `.tmpl` files for each file to generate (`compose.yml.tmpl`, `config.yml.tmpl`, `.env.tmpl`, `.env.example.tmpl`).
4. Update `templates/manifest.json` with the new entry (`display`, `description`, `iconUrl` — a direct image URL for the docs grid) and a new `version` timestamp.
5. Open a pull request — once merged to `main`, the template is available to all users on next cache refresh.

## Checklist

- [ ] `template.yml` has clear prompts with defaults where sensible
- [ ] Generated `config.yml` follows the stack format (domains, backup, etc.)
- [ ] `compose.yml.tmpl` has **no Traefik content** — routing is injected by atrisos
- [ ] `.env.example.tmpl` documents all required env vars without secrets
- [ ] `manifest.json` version bumped

## Test locally

```sh
# Copy your template into the cache dir or run from a dev build
atrisos init teststack --template your-template --dir /tmp/teststack
atrisos validate teststack
```

See the **Template Format** page for file structure and variable syntax.
