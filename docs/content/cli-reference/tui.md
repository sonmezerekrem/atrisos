---
title: TUI
section: cli-reference
group: Interactive
order: 1
icon: grid
description: Interactive dashboard and keyboard shortcuts.
---

# TUI

## `atrisos tui`

Launch the interactive TUI dashboard. Same as running bare `atrisos`.

```sh
atrisos tui
```

## Keyboard shortcuts

| Key | Action |
|-----|--------|
| `↑` / `↓` or `j` / `k` | Navigate stack list |
| `Enter` | Select / focus stack detail |
| `u` | Update selected stack |
| `r` | Restart selected stack |
| `s` | Start selected stack |
| `x` | Stop selected stack |
| `l` | Open log viewer for selected stack |
| `e` | Open shell in the focused service (suspends TUI, returns on exit) |
| `i` | Show stack info / config.yml summary |
| `o` | Show `atrisos outdated` results for selected stack |
| `t` | Show Traefik status panel |
| `/` | Filter stacks by name or tag |
| `?` | Show help |
| `q` or `Ctrl+C` | Quit |
