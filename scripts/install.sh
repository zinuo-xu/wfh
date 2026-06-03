#!/bin/sh
# wfh installer — privacy-first WFH activity tracker
# Usage: curl -sfL https://raw.githubusercontent.com/zinuo-xu/wfh/main/scripts/install.sh | sh

set -e

BINARY="wfh"
REPO="zinuo-xu/wfh"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux)   ;;
    darwin)  ;;
    *)       echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Get latest release
echo "Fetching latest release..."
LATEST=$(curl -sfL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)

if [ -z "$LATEST" ]; then
    echo "Could not determine latest version. Install Go and build from source instead."
    echo "  git clone https://github.com/$REPO.git"
    echo "  cd wfh && make build"
    exit 1
fi

echo "Downloading $BINARY $LATEST for $OS/$ARCH..."

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST/${BINARY}_${LATEST}_${OS}_${ARCH}.tar.gz"
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

curl -sfL "$DOWNLOAD_URL" | tar xz -C "$TMP_DIR"

if [ ! -f "$TMP_DIR/$BINARY" ]; then
    echo "Error: binary not found in archive"
    exit 1
fi

# Install
install -d "$INSTALL_DIR"
install -m 755 "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"

echo ""
echo "wfh has been installed to $INSTALL_DIR/$BINARY"
echo ""
echo "Quick start:"
echo "  wfh start          # Start the tracking daemon"
echo "  wfh status         # Check daemon status"
echo "  wfh today          # View today's activity"
echo "  wfh week           # View weekly digest"
echo ""
echo "Privacy-first: all data stays on your machine. No telemetry. No cloud."
