# Codex FileCairn Skill

Use this skill when Codex needs to find local files quickly or register files it just wrote.

## Requirements

1. `filecairn serve` is running.
2. API base URL is `http://127.0.0.1:4819`.

## Supported actions

1. Find candidate files:
- Call `POST /v1/find` with `{query, limit}`.
- Return best matches with path, score, and match reasoning.

2. Register generated files:
- Call `POST /v1/register-created` with `{path, agent, share}`.
- Use `agent="codex"` unless another agent identity is required.

3. Refresh index on demand:
- Call `POST /v1/scan` with `{roots:[...]}`.
- Use `{all:true}` only with explicit user confirmation.

## Guardrails

1. Do not submit forms, execute files, or modify unrelated files as part of search.
2. Do not include secrets in query text or API payload fields.
3. Register each newly created file so future searches can find it.
