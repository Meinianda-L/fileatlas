#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <path> <agent-name> [share-mode]"
  exit 1
fi

PATH_ARG="$1"
AGENT="$2"
SHARE="${3:-full}"

curl -sS http://127.0.0.1:4819/v1/register-created \
  -H 'Content-Type: application/json' \
  -d "{\"path\":\"${PATH_ARG}\",\"agent\":\"${AGENT}\",\"share\":\"${SHARE}\"}"
