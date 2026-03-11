#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

"$ROOT_DIR/scripts/install.sh"

if ! command -v filecairn >/dev/null 2>&1; then
  export PATH="$HOME/.local/bin:$PATH"
fi

echo "Starting interactive setup..."
filecairn init

echo "To start API for linked agents:"
echo "  filecairn serve"
