# OpenMaguro

Scheduled Claude Code SDK agent task orchestrator. Define agent tasks with cron schedules and track their execution history via a REST API.

## Quick Start

```bash
# Start Postgres
docker compose up -d

# Copy env and adjust if needed
cp .env.example .env

# Run migrations
goose -dir db/migrations postgres "$DATABASE_URL" up

# Start the server
go run cmd/api/main.go
```

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | Postgres connection string |
| `PORT` | No | `8080` | HTTP server port |
| `LOG_LEVEL` | No | `info` | Logging level |
| `MCP_CONFIG_PATH` | No | — | Path to global MCP config file (mcp.json) |
| `ALLOWED_TOOLS` | No | `Bash(curl*),Bash(npx*),WebSearch,WebFetch,mcp__*` | Comma-separated tool patterns auto-approved for agents |

## API Endpoints

Base URL: `http://localhost:8080`

### Health Check

```
GET /health
```

Response `200`:
```json
{"status": "ok"}
```

---

### Agent Tasks

#### Create Agent Task

```
POST /api/v1/agent-tasks
```

Request body:
```json
{
  "name": "Daily report",
  "cron_expression": "0 6 * * *",
  "prompt": "Generate the daily sales report and email it to the team",
  "enabled": true
}
```

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `name` | string | Yes | — | Human-readable task name (max 255 chars) |
| `cron_expression` | string | Yes | — | Cron schedule expression |
| `prompt` | string | Yes | — | Instruction for the Claude Code SDK agent |
| `enabled` | bool | No | `true` | Whether the task is active |
| `mcp_config` | string | No | — | Path to custom MCP config file (overrides global) |
| `allowed_tools` | string | No | — | Comma-separated extra tool patterns (additive to global) |
| `system_agent` | bool | No | `false` | Mark as internal system agent |
| `global_skill_access` | bool | No | `false` | Grant access to all skills (instead of only assigned ones) |
| `on_success_task_id` | uuid | No | — | Agent task to trigger when this task succeeds |
| `on_failure_task_id` | uuid | No | — | Agent task to trigger when this task fails |

Response `201`:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Daily report",
  "cron_expression": "0 6 * * *",
  "prompt": "Generate the daily sales report and email it to the team",
  "enabled": true,
  "system_agent": false,
  "on_success_task_id": null,
  "on_failure_task_id": null,
  "created_at": "2026-03-05T10:00:00Z",
  "updated_at": "2026-03-05T10:00:00Z"
}
```

#### List Agent Tasks

```
GET /api/v1/agent-tasks
```

Response `200`: Array of agent task objects (ordered by created_at DESC).

#### Get Agent Task

```
GET /api/v1/agent-tasks/{id}
```

Response `200`: Agent task object.
Response `404`: `{"error": "agent task not found"}`

#### Update Agent Task (Partial)

```
PATCH /api/v1/agent-tasks/{id}
```

Request body (all fields optional):
```json
{
  "name": "Updated name",
  "enabled": false
}
```

Response `200`: Updated agent task object.

#### Delete Agent Task

```
DELETE /api/v1/agent-tasks/{id}
```

Response `204`: No content.

#### Run Agent Task Immediately

```
POST /api/v1/agent-tasks/{id}/run
```

Triggers immediate execution of the agent task in the background. No request body required.

Response `202`:
```json
{"status": "accepted"}
```

Response `404`: `{"error": "agent task not found"}`

Check execution results via `GET /api/v1/agent-tasks/{id}/executions`.

---

### Agent Chaining

Chain agents together so one triggers another on success or failure. Set `on_success_task_id` and/or `on_failure_task_id` when creating or updating an agent task.

```bash
curl -X PATCH http://localhost:8080/api/v1/agent-tasks/{id} \
  -H 'Content-Type: application/json' \
  -d '{"on_success_task_id": "uuid-of-next-agent"}'
```

When an agent completes, the chained agent receives the parent's output as context in its prompt. Circular chains are rejected at create/update time. Chained executions include `triggered_by_execution_id` for traceability.

---

### Heartbeat & Recovery

The scheduler runs a heartbeat every 10 minutes that:
- **Detects missed cron jobs**: looks back 24 hours, compares expected fire times against actual executions, and triggers any missed runs
- **Marks stale executions**: any execution stuck in `running` status for over 2 hours is marked as `failed`

This ensures tasks are recovered even after server restarts or downtime.

---

### Scheduled Tasks (One-Time)

#### Create Scheduled Task

```
POST /api/v1/scheduled-tasks
```

Request body:
```json
{
  "name": "Send Slack reminder",
  "prompt": "Send a message to #general on Slack saying 'Team standup in 5 minutes'",
  "run_at": "2026-03-05T13:00:00Z"
}
```

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `name` | string | Yes | — | Human-readable task name |
| `prompt` | string | Yes | — | Instruction for the Claude Code SDK agent |
| `run_at` | string (RFC3339) | Yes | — | When to execute the task |
| `mcp_config` | string | No | — | Path to custom MCP config file |
| `allowed_tools` | string | No | — | Comma-separated extra tool patterns |

Response `201`: Scheduled task object. The task auto-deletes after execution, but the execution log persists.

---

### MCP Servers

Manage MCP (Model Context Protocol) servers that give agents access to external tools. The global config at `MCP_CONFIG_PATH` is used for all executions unless a task specifies its own `mcp_config`.

#### List MCP Servers

```
GET /api/v1/mcp-servers
```

Response `200`:
```json
[
  {
    "name": "linear",
    "command": "npx",
    "args": ["-y", "linear-mcp-server"],
    "env": {
      "LINEAR_API_KEY": "lin_api_..."
    }
  }
]
```

#### Add MCP Server

```
POST /api/v1/mcp-servers
```

Request body:
```json
{
  "name": "notion",
  "command": "npx",
  "args": ["-y", "notion-mcp-server"],
  "env": {
    "NOTION_API_KEY": "ntn_..."
  }
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | Yes | Unique server identifier |
| `command` | string | Yes | Command to run |
| `args` | string[] | Yes | Command arguments |
| `env` | object | No | Environment variables (API keys, etc) |

Response `201`:
```json
{"status": "ok", "name": "notion"}
```

#### Remove MCP Server

```
DELETE /api/v1/mcp-servers/{name}
```

Response `204`: No content.

---

### Skills

Skills are reusable markdown documents (instructions, API references, credentials) that get injected into agent execution prompts. Attach skills to specific agents, or set `global_skill_access: true` on an agent to give it all skills.

#### Create Skill

```
POST /api/v1/skills
```

Request body:
```json
{
  "title": "Slack API",
  "content": "Use the Slack Bot Token xoxb-**** to send messages via the Slack API..."
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `title` | string | Yes | Skill name (max 255 chars) |
| `content` | string | Yes | Skill content (markdown, instructions, API keys) |

Response `201`: Skill object.

#### List Skills

```
GET /api/v1/skills
```

Response `200`: Array of skill objects (ordered by created_at DESC).

#### Get / Update / Delete Skill

```
GET /api/v1/skills/{id}
PATCH /api/v1/skills/{id}
DELETE /api/v1/skills/{id}
```

#### Attach Skill to Agent

```
POST /api/v1/agent-tasks/{id}/skills/{skillId}
```

Response `204`: No content. Idempotent — attaching twice is a no-op.

#### Detach Skill from Agent

```
DELETE /api/v1/agent-tasks/{id}/skills/{skillId}
```

Response `204`: No content.

#### List Skills for Agent

```
GET /api/v1/agent-tasks/{id}/skills
```

Response `200`: Array of skill objects attached to this agent.

---

### Task Executions

#### List All Executions

```
GET /api/v1/executions
```

Response `200`: Array of all execution objects (ordered by created_at DESC). Includes executions from deleted one-shot tasks (with `agent_task_id: null` and `task_name` preserved).

#### List Executions for an Agent Task

```
GET /api/v1/agent-tasks/{taskId}/executions
```

Response `200`:
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "agent_task_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "success",
    "started_at": "2026-03-05T06:00:00Z",
    "finished_at": "2026-03-05T06:01:30Z",
    "summary": "Generated and sent the daily sales report successfully",
    "triggered_by_execution_id": null,
    "created_at": "2026-03-05T06:00:00Z"
  }
]
```

| Status | Description |
|---|---|
| `pending` | Execution created, not yet started |
| `running` | Agent is currently executing |
| `success` | Completed successfully |
| `failure` | Completed with an error |
| `timeout` | Exceeded timeout_seconds |

#### Get Execution

```
GET /api/v1/executions/{id}
```

Response `200`: Task execution object.
Response `404`: `{"error": "execution not found"}`

## Tech Stack

- **Go** with Chi router
- **PostgreSQL 17** with pgx/v5 driver
- **sqlc** for type-safe SQL query generation
- **Goose** for database migrations

## Development

```bash
# Regenerate sqlc code after changing queries
sqlc generate

# Build
go build ./...
```

Migrations run automatically on server startup (embedded via `go:embed` + goose).

## Testing

E2e API tests use [testcontainers-go](https://golang.testcontainers.org/) to spin up a real Postgres instance per test. Works with both Podman and Docker.

```bash
# Run all tests
go test ./internal/tests/... -v

# Run without cache
go test ./internal/tests/... -v -count=1
```

Requires Podman or Docker to be running. On macOS with Podman, the socket is auto-detected.