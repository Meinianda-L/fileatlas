#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="$ROOT_DIR/bin"
TARGET_DIR="${HOME}/.local/bin"

mkdir -p "$BIN_DIR"
mkdir -p "$TARGET_DIR"

cd "$ROOT_DIR"
go build -o "$BIN_DIR/filecairn" ./cmd/filecairn
cp "$BIN_DIR/filecairn" "$TARGET_DIR/filecairn"
chmod +x "$TARGET_DIR/filecairn"

echo "Installed: $TARGET_DIR/filecairn"
echo "If needed, add ~/.local/bin to PATH:"
echo '  export PATH="$HOME/.local/bin:$PATH"'
