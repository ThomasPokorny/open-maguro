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
  "enabled": true,
  "timeout_seconds": 120
}
```

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `name` | string | Yes | — | Human-readable task name (max 255 chars) |
| `cron_expression` | string | Yes | — | Cron schedule expression |
| `prompt` | string | Yes | — | Instruction for the Claude Code SDK agent |
| `enabled` | bool | No | `true` | Whether the task is active |
| `timeout_seconds` | int | No | `60` | Max execution time (1–3600) |

Response `201`:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Daily report",
  "cron_expression": "0 6 * * *",
  "prompt": "Generate the daily sales report and email it to the team",
  "enabled": true,
  "timeout_seconds": 120,
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
  "run_at": "2026-03-05T13:00:00Z",
  "timeout_seconds": 60
}
```

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `name` | string | Yes | — | Human-readable task name |
| `prompt` | string | Yes | — | Instruction for the Claude Code SDK agent |
| `run_at` | string (RFC3339) | Yes | — | When to execute the task |
| `timeout_seconds` | int | No | `60` | Max execution time (1–3600) |

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

### Task Executions

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

To test it end-to-end:
docker compose up -d
goose -dir db/migrations postgres "$DATABASE_URL" up
go run cmd/api/main.go

# Create a task that fires every minute
curl -X POST localhost:8080/api/v1/agent-tasks \
-H 'Content-Type: application/json' \
-d '{"name":"test","cron_expression":"*/1 * * * *","prompt":"Say hello"}'
