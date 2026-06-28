# Atrisos Documentation

Documentation for [atrisos](https://github.com/sonmezerekrem/atrisos) — CLI + TUI for Podman Compose stacks with automatic Traefik routing.

## Documentation site

The docs site lives in this folder and is designed for [GitHub Pages](https://pages.github.com/).

**Top menu:** Overview · Agents · CLI Reference · Templates · Contribution

| Path | Purpose |
|------|---------|
| [`index.html`](index.html) | Documentation shell (loads `app.js`) |
| [`app.js`](app.js) | Docs UI (React, no build step) |
| [`content/`](content/) | **Markdown pages** — add `.md` files here |
| [`build.mjs`](build.mjs) | Scans `content/` and generates `nav.json` |
| [`nav.json`](nav.json) | Generated navigation (run `make docs`) |

### Add a page

1. Create a file in `content/` with YAML frontmatter (see [`content/README.md`](content/README.md))
2. Run `make docs` to regenerate `nav.json`
3. Preview with `make docs-serve` → http://localhost:8080

### Deploy to GitHub Pages

1. In repo **Settings → Pages**, set source to **GitHub Actions**
2. Push to `main` — the [docs workflow](../.github/workflows/docs.yml) builds and deploys automatically

Site URL: `https://<user>.github.io/atrisos/`

For a custom domain at the repo root, set `"basePath": ""` in `docs/build.mjs`.
