# FileAtlas

FileAtlas is a local file indexer with CLI and HTTP API access.

## Features

- Recursive scan with incremental re-indexing
- Ranked search using token match, labels, and recency
- Content indexing is **off by default** until explicitly enabled
- Provider configuration supports multiple endpoints/models
- Local HTTP API for agent integrations
- Agent-created files can be registered and indexed immediately

## Install

```bash
cd /Users/kevin/fileatlas
./scripts/install.sh
```

If `fileatlas` is not found, add:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## First Run

```bash
fileatlas init
```

`init` asks for:

1. Provider configuration
2. Content indexing permission
3. Scan scope (`home` or selected roots)
4. Initial index build

## Commands

```bash
fileatlas init
fileatlas scan [--all] [--roots /a,/b]
fileatlas find [--limit 20] <query>
fileatlas register-created --path <file> --agent <name> [--share full]
fileatlas serve [--addr 127.0.0.1:4819]
fileatlas status
fileatlas content on|off
```

## Search Example

```bash
fileatlas find "toefl speaking timer"
```

## API

Start server:

```bash
fileatlas serve
```

Endpoints:

- `GET /v1/health`
- `GET /v1/status`
- `POST /v1/find` with `{ "query": "toefl timer", "limit": 10 }`
- `POST /v1/scan` with `{ "all": true }` or `{ "roots": ["/Users/kevin/Documents"] }`
- `POST /v1/register-created` with `{ "path": "...", "agent": "openclaw", "share": "full" }`

## Integration Scripts (OpenClaw)

- `integrations/openclaw/find_file.sh`
- `integrations/openclaw/register_created.sh`

## Register Agent-Created Files

Direct call:

```bash
fileatlas register-created --path /abs/path/file.md --agent openclaw --share full
```

Wrapper:

```bash
echo "content" | ./scripts/agent_write.sh openclaw /tmp/demo.md full
```

## Data Files

Stored in `~/.fileatlas` (or `$FILEATLAS_HOME`):

- `config.json`
- `files.json`
- `inverted.json`
- `agent_created.jsonl`

## Safety Defaults

- No file move/delete actions
- No network upload built into scan/search path
- Content indexing disabled until confirmed
- Full-home scan requires explicit confirmation

## Internals

- Ranking/indexing details: `docs/ALGORITHMS.md`
