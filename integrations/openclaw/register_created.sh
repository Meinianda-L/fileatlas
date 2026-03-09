#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <path> <agent-name> [share-mode]"
  exit 1
fi

PATH_ARG="$1"
AGENT="$2"
SHARE="${3:-full}"

json_escape() {
  local s="$1"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  s="${s//$'\n'/\\n}"
  printf '%s' "$s"
}

PAYLOAD="$(printf '{\"path\":\"%s\",\"agent\":\"%s\",\"share\":\"%s\"}' \
  "$(json_escape "$PATH_ARG")" \
  "$(json_escape "$AGENT")" \
  "$(json_escape "$SHARE")")"

curl -sS http://127.0.0.1:4819/v1/register-created \
  -H 'Content-Type: application/json' \
  -d "$PAYLOAD"
