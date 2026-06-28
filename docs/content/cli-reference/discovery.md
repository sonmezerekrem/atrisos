---
title: Discovery & Init
section: cli-reference
group: Discovery & Init
order: 1
icon: pin
description: List, register, unregister stacks and run the init wizard.
---

# Discovery & Init

## `atrisos list`

List all discovered stacks (from root dir and registered paths).

```sh
atrisos list
atrisos list --format table     # default
atrisos list --format json      # JSON output for scripting
atrisos list --format plain     # one name per line
```

## `atrisos register <path>`

Register a stack directory that lives outside the root directory.

```sh
atrisos register ./myapp
atrisos register /opt/services/monitoring
```

The path is stored in `~/.config/atrisos/registry.json`.

## `atrisos unregister <stack>`

Remove a stack from the registry. Does not delete any files.

```sh
atrisos unregister myapp
```

## `atrisos init [name]`

Interactive wizard to create a new stack directory from a template. Templates are fetched from the `templates/` directory in the atrisos GitHub repository (`main` branch) and cached locally in `~/.config/atrisos/templates-cache/`. Subsequent calls use the cache and work offline; atrisos checks for updates when online.

```sh
atrisos init                          # prompts for name and template, creates in root dir
atrisos init myapp                    # create stack named "myapp", prompts for template
atrisos init myapp --template basic   # skip template prompt, use "basic" template
atrisos init myapp --dir ./myapp      # create at a specific path instead of root dir
atrisos init --list-templates         # list available templates without creating anything
```

The wizard asks:

1. Stack name and description
2. Template to use (lists available templates with descriptions)
3. Template-specific prompts (e.g. image name, domain, port)
4. Whether to enable backups

See the Templates section for available templates and how to contribute new ones.
