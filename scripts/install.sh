#!/usr/bin/env sh
set -e

VERSION=""
PREFIX=""

while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --prefix)  PREFIX="$2";  shift 2 ;;
    *) echo "Unknown flag: $1"; exit 1 ;;
  esac
done

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

info()    { printf '\033[1;34m→\033[0m %s\n' "$*"; }
success() { printf '\033[1;32m✓\033[0m %s\n' "$*"; }
warn()    { printf '\033[1;33m!\033[0m %s\n' "$*"; }
die()     { printf '\033[1;31m✗\033[0m %s\n' "$*" >&2; exit 1; }

have() { command -v "$1" >/dev/null 2>&1; }

# ---------------------------------------------------------------------------
# Detect OS and arch
# ---------------------------------------------------------------------------

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) die "Unsupported architecture: $ARCH" ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) die "Unsupported OS: $OS" ;;
esac

# ---------------------------------------------------------------------------
# Install Podman
# ---------------------------------------------------------------------------

install_podman_macos() {
  if ! have brew; then
    info "Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    if [ -d "/opt/homebrew/bin" ]; then
      export PATH="/opt/homebrew/bin:$PATH"
    elif [ -d "/usr/local/bin" ]; then
      export PATH="/usr/local/bin:$PATH"
    fi
  fi

  if have podman; then
    success "Podman already installed ($(podman --version))"
  else
    info "Installing Podman via Homebrew..."
    brew install podman
    success "Podman installed"
  fi
}

install_podman_linux() {
  if have apt-get; then
    info "Installing Podman via apt..."
    sudo apt-get update -qq
    sudo apt-get install -y podman
    PODMAN_VER="$(podman --version 2>/dev/null | awk '{print $3}')"
    PODMAN_MAJOR="$(echo "$PODMAN_VER" | cut -d. -f1)"
    PODMAN_MINOR="$(echo "$PODMAN_VER" | cut -d. -f2)"
    if [ "${PODMAN_MAJOR:-0}" -lt 4 ] || { [ "${PODMAN_MAJOR:-0}" -eq 4 ] && [ "${PODMAN_MINOR:-0}" -lt 7 ]; }; then
      info "Podman < 4.7 detected — installing podman-compose..."
      sudo apt-get install -y podman-compose || true
    fi
    success "Podman installed"
  elif have dnf; then
    info "Installing Podman via dnf..."
    sudo dnf install -y podman
    success "Podman installed"
  elif have pacman; then
    info "Installing Podman via pacman..."
    sudo pacman -Sy --noconfirm podman
    success "Podman installed"
  else
    warn "Could not detect package manager — install Podman manually: https://podman.io/get-started"
  fi
}

if have podman; then
  success "Podman already installed ($(podman --version))"
else
  info "Podman not found — installing..."
  if [ "$OS" = "darwin" ]; then
    install_podman_macos
  else
    install_podman_linux
  fi
fi

# ---------------------------------------------------------------------------
# Configure unqualified-search-registries (Linux only)
# Without this, Podman rejects short image names like "postgres:16-alpine".
# ---------------------------------------------------------------------------

if [ "$OS" = "linux" ]; then
  REGISTRIES_CONF="/etc/containers/registries.conf"
  if ! grep -q "^unqualified-search-registries" "$REGISTRIES_CONF" 2>/dev/null; then
    info "Configuring Podman to search docker.io for unqualified image names..."
    printf '\nunqualified-search-registries = ["docker.io"]\n' \
      | sudo tee -a "$REGISTRIES_CONF" > /dev/null
    success "Registry search configured"
  fi
fi

# ---------------------------------------------------------------------------
# Verify podman compose is available
# ---------------------------------------------------------------------------

if podman compose version >/dev/null 2>&1; then
  : # built-in compose available
elif have podman-compose; then
  : # python wrapper available
else
  warn "Neither 'podman compose' nor 'podman-compose' found."
  if [ "$OS" = "linux" ] && have apt-get; then
    info "Installing podman-compose..."
    sudo apt-get install -y podman-compose
  else
    warn "Install it manually: https://github.com/containers/podman-compose"
  fi
fi

# ---------------------------------------------------------------------------
# Download atrisos binary
# ---------------------------------------------------------------------------

REPO="sonmezerekrem/atrisos"
ASSET="atrisos-${OS}-${ARCH}"

if [ -z "$PREFIX" ]; then
  if [ "$OS" = "darwin" ] && [ -d "/opt/homebrew" ]; then
    PREFIX="/opt/homebrew"
  else
    PREFIX="/usr/local"
  fi
fi

INSTALL_DIR="$PREFIX/bin"
INSTALL_PATH="$INSTALL_DIR/atrisos"

if [ -z "$VERSION" ]; then
  info "Fetching latest atrisos release..."
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  [ -z "$VERSION" ] && die "Failed to fetch latest version. Use --version to specify one."
fi

info "Installing atrisos $VERSION ($OS/$ARCH)..."

TMP="$(mktemp)"
curl -fsSL "https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}" -o "$TMP"
chmod +x "$TMP"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "$INSTALL_PATH"
else
  info "Installing to $INSTALL_PATH (sudo required)..."
  sudo mv "$TMP" "$INSTALL_PATH"
fi

success "atrisos $VERSION installed to $INSTALL_PATH"
echo ""
echo "  Run: atrisos"
echo "  (First run will complete setup — ACME email, stacks root, Podman machine on macOS)"
echo ""
