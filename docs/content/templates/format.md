---
title: Template Format
section: templates
group: Template Format
order: 1
icon: grid
description: Directory layout, manifest.json, template.yml, and .tmpl files.
---

# Template Format

## Directory structure

```
templates/
├── manifest.json          # index of available templates
├── basic/
│   ├── template.yml       # template metadata
│   ├── compose.yml.tmpl   # Go text/template for compose.yml
│   ├── config.yml.tmpl    # Go text/template for config.yml
│   ├── .env.tmpl          # Go text/template for .env
│   └── .env.example.tmpl  # Go text/template for .env.example
└── webapp-postgres/
    └── ...
```

## manifest.json

Lists all available templates with display names and descriptions shown in the `atrisos init` wizard.

```json
{
  "version": "2026-06-28T20:00:00Z",
  "templates": [
    {
      "name": "basic",
      "display": "Basic",
      "description": "Single service with optional domain routing"
    }
  ]
}
```

Bump `version` (ISO timestamp) whenever templates change so clients know to refresh the cache.

## template.yml

Metadata file inside each template directory. Defines the prompts the wizard will ask, and the variables available in `.tmpl` files.

```yaml
name: "basic"
display: "Basic"
description: "Minimal stack with a single service and optional domain"

prompts:
  - name: image
    label: "Docker image (e.g. nginx:alpine)"
    type: string
    required: true

  - name: port
    label: "Container port the service listens on"
    type: int
    default: 80

  - name: domain
    label: "Domain hostname (leave blank to skip Traefik routing)"
    type: string
    required: false
```

### Prompt types

| Type | Input | Notes |
|------|-------|-------|
| `string` | Free text | |
| `int` | Integer | Validated as number |
| `bool` | yes/no | Rendered as checkbox in wizard |
| `select` | Choice list | Requires an `options` list in `template.yml` |

## Template files (.tmpl)

Template files use Go's `text/template` syntax. Variables are the answers to the prompts defined in `template.yml`, plus built-in variables.

### Built-in variables

| Variable | Value |
|----------|-------|
| `{{.Name}}` | Stack name as entered in the wizard |
| `{{.DirName}}` | Stack directory name (same as Name, slugified) |

### Example: compose.yml.tmpl

```yaml
services:
  web:
    image: {{.image}}
{{- if .domain}}
    expose:
      - "{{.port}}"
{{- else}}
    ports:
      - "{{.port}}:{{.port}}"
{{- end}}
```

### Example: config.yml.tmpl

```yaml
name: "{{.Name}}"
description: ""
tags: []
meta: {}
{{- if .domain}}

domains:
  - service: "web"
    host: "{{.domain}}"
    port: {{.port}}
    tls: true
{{- end}}

update:
  mode: manual

backup:
  enabled: false
```
