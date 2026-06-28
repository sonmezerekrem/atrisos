---
title: Pull Requests
section: contribution
group: Process
order: 1
icon: folder
description: Issue workflow, pull requests, releases, and license.
---

# Pull Requests

## Before you start

1. Open an issue to discuss larger changes before investing significant time
2. For small fixes (typos, docs, obvious bugs), a PR without a prior issue is fine

## Opening a PR

1. Fork the repo and create a branch from `main`
2. Make your changes with tests where applicable
3. Open a PR against `main` with a clear description of what changed and why
4. Link related issues when applicable

## Release process

Releases are tagged `v*` on GitHub. GitHub Actions runs GoReleaser and publishes binaries for:

- `atrisos-linux-amd64`, `atrisos-linux-arm64`
- `atrisos-darwin-amd64`, `atrisos-darwin-arm64`

You do not need to cut releases manually unless you are a maintainer.

## License

Contributions are accepted under the [MIT license](https://github.com/sonmezerekrem/atrisos/blob/main/LICENSE).
