# Generic HTTP Agent Skill (FileCairn)

Use this skill for any model runtime that can send local HTTP requests.

## Requirements

1. Start FileCairn API: `filecairn serve`
2. Set base URL: `http://127.0.0.1:4819`

## API operations

1. Health check:
- `GET /v1/health`

2. Search files:
- `POST /v1/find`
- Body: `{ "query": "notes draft", "limit": 10 }`

3. Trigger scan:
- `POST /v1/scan`
- Body: `{ "roots": ["/path/a", "/path/b"] }` or `{ "all": true }`

4. Register new file created by agent:
- `POST /v1/register-created`
- Body: `{ "path": "/abs/path/new.md", "agent": "agent-name", "share": "private|summary|full" }`

## Recommended usage flow

1. Query with `/v1/find`.
2. Show top candidates to user.
3. After writing a new file, call `/v1/register-created`.
4. Use `/v1/scan` only when index refresh is needed.

## Guardrails

1. Keep all traffic local unless user explicitly configures otherwise.
2. Never send secret values inside `query`, `path`, or `agent` fields.
3. Confirm before full-home scans.
