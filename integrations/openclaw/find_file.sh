#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 \"query words\" [limit]"
  exit 1
fi

QUERY="$1"
LIMIT="${2:-10}"

curl -sS http://127.0.0.1:4819/v1/find \
  -H 'Content-Type: application/json' \
  -d "{\"query\":\"${QUERY}\",\"limit\":${LIMIT}}"
