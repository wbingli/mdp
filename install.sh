#!/bin/sh
set -e

REPO="wbingli/mdp"
INSTALL_DIR="${MDP_INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Darwin) OS="darwin" ;;
  Linux)  OS="linux" ;;
  *)      echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)             echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Get latest version
if command -v curl >/dev/null 2>&1; then
  FETCH="curl -fsSL"
  FETCH_REDIRECT="curl -fsSL -o"
elif command -v wget >/dev/null 2>&1; then
  FETCH="wget -qO-"
  FETCH_REDIRECT="wget -qO"
else
  echo "Error: curl or wget required" >&2
  exit 1
fi

VERSION="$($FETCH "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)"
if [ -z "$VERSION" ]; then
  echo "Error: could not determine latest version" >&2
  exit 1
fi

BINARY="mdp-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}"

echo "Installing mdp ${VERSION} (${OS}/${ARCH})..."

mkdir -p "$INSTALL_DIR"

if [ "$FETCH_REDIRECT" ]; then
  $FETCH_REDIRECT "$INSTALL_DIR/mdp" "$URL"
else
  $FETCH "$URL" > "$INSTALL_DIR/mdp"
fi

chmod +x "$INSTALL_DIR/mdp"

echo "Installed mdp to ${INSTALL_DIR}/mdp"

# Check if install dir is in PATH
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "Add ${INSTALL_DIR} to your PATH:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac
