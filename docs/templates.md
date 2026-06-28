# Stack Init Templates

`atrisos init` creates new stacks from templates. Templates live in the `templates/` directory of the Atrisos repository and are fetched from GitHub at runtime.

---

## How templates are fetched

1. On `atrisos init`, Atrisos checks `~/.config/atrisos/templates-cache/manifest.json` for a cached version.
2. If online, Atrisos compares the local manifest against the remote one (fetched from `raw.githubusercontent.com/sonmezerekrem/atrisos/main/templates/manifest.json`). If the remote is newer, the full template set is re-downloaded.
3. If offline (or GitHub is unreachable), the local cache is used as-is.
4. On first run with no cache, Atrisos fetches all templates before presenting the wizard.

The cache is stored at `~/.config/atrisos/templates-cache/`.

---

## Template directory structure

```
templates/
├── manifest.json          # index of available templates
├── basic/
│   ├── template.yml       # template metadata
│   ├── compose.yml.tmpl   # Go text/template for compose.yml
│   ├── config.yml.tmpl    # Go text/template for config.yml
│   ├── .env.tmpl          # Go text/template for .env
│   └── .env.example.tmpl  # Go text/template for .env.example
├── nginx/
│   └── ...
├── postgres/
│   └── ...
└── ghost/
    └── ...
```

---

## manifest.json

Lists all available templates with display names and descriptions shown in the `atrisos init` wizard.

```json
{
  "version": "2026-06-27T00:00:00Z",
  "templates": [
    {
      "name": "basic",
      "display": "Basic",
      "description": "Minimal stack: single service, one domain entry, no database"
    },
    {
      "name": "nginx",
      "display": "Nginx static site",
      "description": "Nginx serving static files from a volume"
    },
    {
      "name": "postgres",
      "display": "App + PostgreSQL",
      "description": "Web service backed by a PostgreSQL database with backup enabled"
    },
    {
      "name": "ghost",
      "display": "Ghost blog",
      "description": "Ghost CMS with MySQL and Let's Encrypt"
    }
  ]
}
```

---

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

---

## Template files (.tmpl)

Template files use Go's `text/template` syntax. Variables are the answers to the prompts defined in `template.yml`, plus a set of built-in variables.

### Built-in variables

| Variable | Value |
|----------|-------|
| `{{.Name}}` | Stack name as entered in the wizard |
| `{{.DirName}}` | Stack directory name (same as Name, slugified) |

### Example: basic/compose.yml.tmpl

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

### Example: basic/config.yml.tmpl

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

### Example: basic/.env.tmpl

```
# Add environment variables for {{.Name}} here
```

### Example: basic/.env.example.tmpl

```
# Copy to .env and fill in values
# Add environment variables for {{.Name}} here
```

---

## Adding a new template

1. Create a new directory under `templates/<name>/`.
2. Add `template.yml` with prompts.
3. Add `.tmpl` files for each file to generate.
4. Update `manifest.json` with the new entry and a new `version` timestamp.
5. Open a pull request — once merged to `main`, the template is available to all users on next cache refresh.

---

## Forcing a cache refresh

```sh
atrisos init --refresh-templates   # re-download all templates from GitHub then run wizard
```
