#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <agent-name> <target-path> [share-mode]"
  echo "Example: echo 'content' | $0 openclaw /tmp/note.md full"
  exit 1
fi

AGENT="$1"
TARGET="$2"
SHARE="${3:-full}"
mkdir -p "$(dirname "$TARGET")"
cat > "$TARGET"

if command -v fileatlas >/dev/null 2>&1; then
  fileatlas register-created --path "$TARGET" --agent "$AGENT" --share "$SHARE" >/dev/null
  echo "Wrote + registered: $TARGET"
else
  echo "Wrote file but fileatlas command not found in PATH: $TARGET"
fi
