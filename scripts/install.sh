#!/usr/bin/env sh
set -e

VERSION=""
PREFIX=""

# Parse flags
while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --prefix)  PREFIX="$2";  shift 2 ;;
    *) echo "Unknown flag: $1"; exit 1 ;;
  esac
done

# Detect OS and arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

ASSET="atrisos-${OS}-${ARCH}"

# Fetch latest version if not specified
if [ -z "$VERSION" ]; then
  VERSION="$(curl -fsSL https://api.github.com/repos/sonmezerekrem/atrisos/releases/latest \
    | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest version. Use --version to specify one."
    exit 1
  fi
fi

echo "→ Installing atrisos $VERSION ($OS/$ARCH)..."

# Determine install prefix
if [ -z "$PREFIX" ]; then
  if [ "$OS" = "darwin" ] && [ -d "/opt/homebrew" ]; then
    PREFIX="/opt/homebrew"
  else
    PREFIX="/usr/local"
  fi
fi

INSTALL_DIR="$PREFIX/bin"
INSTALL_PATH="$INSTALL_DIR/atrisos"

# Download to a temp file
TMP="$(mktemp)"
curl -fsSL "https://github.com/sonmezerekrem/atrisos/releases/download/${VERSION}/${ASSET}" -o "$TMP"
chmod +x "$TMP"

# Install (may need sudo)
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "$INSTALL_PATH"
else
  echo "→ Installing to $INSTALL_PATH (sudo required)..."
  sudo mv "$TMP" "$INSTALL_PATH"
fi

echo "✓ atrisos $VERSION installed to $INSTALL_PATH"
echo ""
echo "Run: atrisos --help"
