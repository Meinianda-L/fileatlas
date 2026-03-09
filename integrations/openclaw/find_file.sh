#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 \"query words\" [limit]"
  exit 1
fi

QUERY="$1"
LIMIT="${2:-10}"

json_escape() {
  local s="$1"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  s="${s//$'\n'/\\n}"
  printf '%s' "$s"
}

PAYLOAD="$(printf '{\"query\":\"%s\",\"limit\":%s}' "$(json_escape "$QUERY")" "$LIMIT")"

curl -sS http://127.0.0.1:4819/v1/find \
  -H 'Content-Type: application/json' \
  -d "$PAYLOAD"
