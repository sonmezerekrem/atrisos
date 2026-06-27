# Installation

## Requirements

### Linux (Ubuntu/Debian)

- Ubuntu 22.04+ or Debian 12+
- `podman` (v4.7+ recommended for built-in compose support)
- `podman-compose` OR Podman v4.7+ (`podman compose` subcommand)
- Ports 80 and 443 open on the host (for Traefik)

Install Podman on Ubuntu/Debian:
```sh
sudo apt update && sudo apt install -y podman
# Optional: install podman-compose if using Podman < 4.7
sudo apt install -y podman-compose
```

### macOS

- macOS 13+ (Ventura or later)
- `podman` installed via Homebrew
- `podman machine` running (atrisos manages this for you)

Install Podman on macOS:
```sh
brew install podman
```

---

## Install atrisos

```sh
curl -fsSL https://get.atrisos.io/install.sh | sh
```

The script:
1. Detects your OS and architecture (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64)
2. Downloads the latest release binary from GitHub
3. Places it at `/usr/local/bin/atrisos` (Linux) or `/opt/homebrew/bin/atrisos` (macOS Apple Silicon) or `/usr/local/bin/atrisos` (macOS Intel)
4. Makes it executable
5. Prints next steps

To install a specific version:
```sh
curl -fsSL https://get.atrisos.io/install.sh | sh -s -- --version v0.3.0
```

To install to a custom location:
```sh
curl -fsSL https://get.atrisos.io/install.sh | sh -s -- --prefix ~/.local
```

---

## First run

```sh
atrisos version    # verify installation
atrisos            # launch TUI — will run first-time setup wizard
```

The setup wizard:
1. Creates `~/.config/atrisos/config.yml` with defaults
2. Prompts for your ACME email (for Let's Encrypt TLS)
3. Prompts for your stacks root directory (default `~/atrisos-stacks`)
4. **macOS only**: if no `podman machine` exists, silently creates and starts one named `atrisos` with a progress indicator (~1–2 min, one-time). On subsequent runs atrisos starts the machine automatically if it is stopped.
5. Creates the `atrisos_net` Podman network
6. Starts the managed Traefik container

No additional tools need to be installed for backup support — atrisos downloads and manages its own `restic` binary the first time a backup runs.

---

## Creating your first stack

```sh
atrisos init myapp        # interactive wizard
cd ~/atrisos-stacks/myapp
# edit compose.yml and config.yml
atrisos up myapp
```

---

## Updating atrisos

```sh
curl -fsSL https://get.atrisos.io/install.sh | sh
```

The installer is idempotent and will replace the existing binary.

---

## Uninstall

```sh
atrisos traefik down          # stop Traefik
atrisos down --all            # stop all stacks (optional)
rm "$(which atrisos)"         # remove binary
rm -rf ~/.config/atrisos      # remove config (optional — preserves your stacks)
```

---

## Building from source

Requirements: Go 1.22+

```sh
git clone https://github.com/sonmezerekrem/atrisos
cd atrisos
go build -o atrisos .
sudo mv atrisos /usr/local/bin/
```
