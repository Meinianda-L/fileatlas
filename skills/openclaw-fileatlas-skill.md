# OpenClaw FileAtlas Skill

Use this skill when OpenClaw must find local files or register newly created files.

## Required runtime

1. `fileatlas serve` must be running.
2. API default address: `http://127.0.0.1:4819`.

## Operations

1. Find files:
- Call `POST /v1/find` with `{query, limit}`.
- Return top results with path, score, and why matched.

2. Register agent-created file:
- Call `POST /v1/register-created` with `{path, agent, share}`.
- Use `agent="openclaw"` and `share="full"` by default.

3. Trigger scan:
- Call `POST /v1/scan` with `{all:true}` or `{roots:[...]}` only after user approval.

## Guardrails

1. Never expose secret environment values in requests.
2. Never assume content visibility is globally enabled.
3. Use `register-created` for every file written by OpenClaw.
