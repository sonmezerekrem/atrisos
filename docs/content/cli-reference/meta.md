---
title: Version & Meta
section: cli-reference
group: Meta
order: 1
icon: info
description: Version info, self-update, and shell completion.
---

# Version & Meta

## `atrisos version`

Print version, build info, detected platform, and latest available version (checked in background, cached 24 hours).

```sh
atrisos version
```

## `atrisos self-update`

Download the latest atrisos release from GitHub, verify its checksum, and replace the running binary in place.

```sh
atrisos self-update
atrisos self-update --version v0.5.0   # pin to a specific version
```

## `atrisos completion <shell>`

Generate shell completion scripts.

```sh
atrisos completion bash   > /etc/bash_completion.d/atrisos
atrisos completion zsh    > ~/.zsh/completions/_atrisos
atrisos completion fish   > ~/.config/fish/completions/atrisos.fish
```
