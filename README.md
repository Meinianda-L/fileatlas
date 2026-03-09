# FileAtlas

FileAtlas is a local-first file indexer and search engine for personal workspaces.
It scans selected folders, builds an inverted index, and exposes both a CLI and a local HTTP API so tools and agents can find files quickly.

## Why This Project Exists

Desktop search is often either too shallow (filename-only) or too heavy (cloud sync, background daemons, complicated setup).
FileAtlas keeps the workflow simple:

- run on your machine
- choose scan roots explicitly
- keep content indexing opt-in
- expose clean interfaces for automation

## Core Features

- Parallel recursive scanning with incremental re-indexing
- Ranked retrieval with token, label, recency, and path boosts
- Content indexing disabled by default until you enable it
- Local API for external tools (`/v1/find`, `/v1/scan`, `/v1/register-created`)
- Agent-created file registration and audit log
- Configurable provider metadata for model-driven workflows

## Install

### Option 1: Build and install locally

```bash
git clone https://github.com/Meinianda-L/fileatlas.git
cd fileatlas
./scripts/install.sh
```

### Option 2: Build without installing

```bash
go build -o ./bin/fileatlas ./cmd/fileatlas
./bin/fileatlas help
```

If `fileatlas` is not found after install, add this to your shell profile:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Quick Start

```bash
fileatlas init
fileatlas scan
fileatlas find "meeting notes"
```

`init` walks through:

1. provider setup (name, endpoint, model, API key env name)
2. content indexing permission
3. scan scope (home directory or custom roots)
4. initial index build

## CLI Reference

```bash
fileatlas init
fileatlas scan [--all] [--roots /a,/b]
fileatlas find [--limit 20] <query>
fileatlas register-created --path <file> --agent <name> [--share private|summary|full]
fileatlas serve [--addr 127.0.0.1:4819]
fileatlas status
fileatlas content on|off
```

## API Reference

Start the local API server:

```bash
fileatlas serve
```

### `GET /v1/health`
Returns service health.

### `GET /v1/status`
Returns current config summary and index stats.

### `POST /v1/find`
Request body:

```json
{ "query": "invoice april", "limit": 10 }
```

### `POST /v1/scan`
Request body examples:

```json
{ "all": true }
```

```json
{ "roots": ["/Users/alex/Documents"] }
```

### `POST /v1/register-created`
Request body:

```json
{ "path": "/abs/path/file.md", "agent": "openclaw", "share": "full" }
```

## OpenClaw Integration

Included scripts:

- `integrations/openclaw/find_file.sh`
- `integrations/openclaw/register_created.sh`
- `scripts/agent_write.sh`

The skill contract for linked agents is documented in `skills/openclaw-fileatlas-skill.md`.

## Data Layout

By default, data is stored in `~/.fileatlas` (or `$FILEATLAS_HOME`):

- `config.json` - runtime config
- `files.json` - current file records
- `inverted.json` - token to file-id map
- `agent_created.jsonl` - append-only agent write log

## Privacy and Safety Defaults

- No file deletion or modification during scan/search
- No automatic web uploads in indexing/search code paths
- Content read disabled until explicitly enabled
- Full-home scan requires confirmation in CLI

## Algorithms

See `docs/ALGORITHMS.md` for indexing and ranking details.

## Development

```bash
go fmt ./...
go test ./...
go build ./cmd/fileatlas
```

## Roadmap

- File watcher mode for near-real-time index refresh
- Optional SQLite backend for larger datasets
- More tool adapters beyond OpenClaw
- Optional interactive TUI for browsing results

## License

MIT
