<p align="center">
  <img src="docs/assets/logo.svg" alt="Atrisos logo" width="72" height="72">
</p>

# Atrisos

![CI](https://github.com/sonmezerekrem/atrisos/actions/workflows/ci.yml/badge.svg)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

CLI + TUI for Podman Compose stacks with automatic Traefik routing and TLS. Write a plain `compose.yml` — atrisos injects routing labels at runtime and manages certificates, backups, and scheduling.

Works on macOS (Apple Silicon + Intel) and Ubuntu/Debian (amd64 + arm64).

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/sonmezerekrem/atrisos/main/scripts/install.sh | sh
```

The script installs Podman if needed and places `atrisos` in your PATH. See [Installation](https://sonmezerekrem.github.io/atrisos/#install) for requirements and manual install options.

## Quick start

```sh
atrisos                      # first run: setup wizard (ACME email, stacks root)
atrisos init myapp           # create a stack from a template
atrisos up myapp             # start it — Traefik starts automatically
atrisos tui                  # open the interactive dashboard
```

## Documentation

Full docs — stack format, CLI reference, templates, Traefik, and more:

**https://sonmezerekrem.github.io/atrisos/**

## License

[MIT](LICENSE)
