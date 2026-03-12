#!/bin/sh
set -e

REPO="antonzaytsev/jira-cli"
INSTALL_DIR="${HOME}/.local/bin"

ARCH=$(uname -m)
case "$ARCH" in
  arm64|aarch64) ARCH="arm64" ;;
  x86_64)        ARCH="x86_64" ;;
  *)             echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

VERSION=$(curl -sI "https://github.com/${REPO}/releases/latest" | grep -i '^location:' | sed 's|.*/v||;s/[[:space:]]//g')
if [ -z "$VERSION" ]; then
  echo "Failed to determine latest version" >&2
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/v${VERSION}/jira_${VERSION}_macOS_${ARCH}.tar.gz"
TMP=$(mktemp -d)

echo "Downloading jira v${VERSION} (${ARCH})..."
curl -sL "$URL" | tar -xz -C "$TMP"

mkdir -p "$INSTALL_DIR"
mv "$TMP/jira_${VERSION}_macOS_${ARCH}/bin/jira" "$INSTALL_DIR/jira"
rm -rf "$TMP"

if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo ""
  echo "Add this to your shell config (~/.zshrc or ~/.bashrc):"
  echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi

echo "Installed jira v${VERSION} to ${INSTALL_DIR}/jira"
