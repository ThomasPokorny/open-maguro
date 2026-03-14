# OpenMaguro🐟

```
▛▌▛▌█▌▛▌▄▖▛▛▌▀▌▛▌▌▌▛▘▛▌
▙▌▙▌▙▖▌▌  ▌▌▌█▌▙▌▙▌▌ ▙▌
  ▌            ▄▌      

OpenMaguro🐟 v0.1 — swim upstream, think downstream.
```

Scheduled Claude Code SDK agent task orchestrator. Define agent tasks with cron schedules and track their execution history via a REST API. Includes a React dashboard for managing agents, skills, kanban tasks, and teams.

**Fully self-contained** — single binary, embedded SQLite database, no external dependencies. Just run it.

## Quick Start

```bash
# Build and run
go build -o maguro ./cmd/api
./maguro
```

That's it. Open `http://localhost:8080` — the dashboard and API are served from the same port.

The SQLite database is created automatically at `~/.maguro/maguro.db`. Migrations run on startup. No Docker, no Postgres, no `.env` file needed.

### Prerequisites

- **Go 1.24+** (to build)
- **Claude CLI** (`claude`) installed and on PATH (for agent execution)

### Development Mode

```bash
# Terminal 1: API server
go run ./cmd/api

# Terminal 2: Frontend with hot reload (proxies API to :8080)
cd maguro-dashboard && npm run dev
```

The Vite dev server runs on `:5173` and proxies `/api` requests to the Go backend on `:8080`.

## Environment Variables

All settings have sensible defaults — no configuration required to get started.

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | `~/.maguro/maguro.db` | Path to SQLite database file |
| `PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Logging level |
| `MCP_CONFIG_PATH` | — | Path to global MCP config file (mcp.json) |
| `ALLOWED_TOOLS` | `Bash(curl*),Bash(npx*),WebSearch,WebFetch,mcp__*` | Comma-separated tool patterns auto-approved for agents |
| `WORKSPACE_ROOT` | `~/.maguro/workspaces` | Root directory for per-agent workspaces |
| `EXECUTION_RETENTION_DAYS` | `30` | Days to keep execution logs (daily cleanup purges older) |
| `MAGURO_SECRET_KEY` | auto-generated | Hex-encoded 32-byte AES-256 key for encrypting skill secrets |

Override any of these via environment variables or an optional `.env` file.

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
| `cron_expression` | string | No | — | Cron schedule expression (omit for non-scheduled agents) |
| `prompt` | string | Yes | — | Instruction for the Claude Code SDK agent |
| `enabled` | bool | No | `true` | Whether the task is active |
| `mcp_config` | string | No | — | Path to custom MCP config file (overrides global) |
| `allowed_tools` | string | No | — | Comma-separated extra tool patterns (additive to global) |
| `system_agent` | bool | No | `false` | Mark as internal system agent |
| `global_skill_access` | bool | No | `false` | Grant access to all skills (instead of only assigned ones) |
| `on_success_task_id` | uuid | No | — | Agent task to trigger when this task succeeds |
| `on_failure_task_id` | uuid | No | — | Agent task to trigger when this task fails |
| `team_id` | uuid | No | — | Team to assign this agent to |

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

Query params: `?team_id={uuid}` to filter by team.

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

#### Open Agent Workspace

```
POST /api/v1/agent-tasks/{id}/open-workspace
```

Opens the agent's workspace directory in the system file explorer (Finder on macOS, `xdg-open` on Linux, Explorer on Windows).

Response `200`:
```json
{"path": "/Users/you/.maguro/workspaces/550e8400-..."}
```

Response `404`: Agent not found or workspace directory doesn't exist.

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

### Agent Workspaces

Each agent gets its own persistent workspace directory at `{WORKSPACE_ROOT}/{agent-id}/`. The directory is:

- **Created** automatically when the agent is created
- **Deleted** automatically when the agent is deleted
- **Set as working directory** (`cwd`) for every claude CLI execution
- **Communicated to the agent** via the system prompt so it knows where to read/write files

Files persist across runs, allowing agents to maintain state, notes, intermediate results, or configuration between executions. One-time scheduled tasks do not get workspaces.

Configure the root with `WORKSPACE_ROOT` env var (default: `~/.maguro/workspaces`).

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
| `content` | string | Yes | Skill content (markdown, instructions) |
| `environment_secrets` | object | No | Key-value map of secrets injected as env vars at runtime |

Response `201`: Skill object. The response includes `secret_keys` (array of key names) but never the secret values.

**Example with secrets:**
```json
{
  "title": "Linear API",
  "content": "Use the Linear GraphQL API. Your API key is in $LINEAR_API_KEY.",
  "environment_secrets": {
    "LINEAR_API_KEY": "lin_api_..."
  }
}
```

The agent receives `LINEAR_API_KEY` as an environment variable at execution time. The secret is encrypted at rest (AES-256-GCM) and never returned by the API.

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

### Kanban Tasks

Assign work items to agents. Each agent processes its queue sequentially — one task at a time. The agent maintains a `work-log.md` in its workspace for context across tasks.

#### Create Kanban Task

```
POST /api/v1/kanban-tasks
```

Request body:
```json
{
  "title": "Write Q1 report",
  "description": "Generate the quarterly report from the KPI data in workspace",
  "agent_task_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `title` | string | Yes | Task title (max 255 chars) |
| `description` | string | No | Detailed task description |
| `agent_task_id` | uuid | Yes | Agent to assign this task to |

Response `201`: Kanban task object with `status: "todo"`. The assigned agent's worker picks it up automatically.

#### List Kanban Tasks

```
GET /api/v1/kanban-tasks
GET /api/v1/kanban-tasks?agent_id={uuid}
GET /api/v1/kanban-tasks?status=todo
GET /api/v1/kanban-tasks?team_id={uuid}
GET /api/v1/kanban-tasks?agent_id={uuid}&status=progress
```

Response `200`: Array of kanban task objects. Done tasks older than 2 hours are hidden from the default list (pass `?status=done` to see all).

| Status | Description |
|---|---|
| `todo` | Queued, waiting for agent |
| `progress` | Agent is working on it |
| `done` | Completed successfully |
| `failed` | Agent failed to complete |

#### Get / Update / Delete

```
GET /api/v1/kanban-tasks/{id}
PATCH /api/v1/kanban-tasks/{id}
DELETE /api/v1/kanban-tasks/{id}
```

---

### Teams

Organize agents into teams. Each agent can belong to one team (nullable FK). Deleting a team unassigns its agents (ON DELETE SET NULL).

#### Create Team

```
POST /api/v1/teams
```

Request body:
```json
{
  "title": "Data Team",
  "description": "Agents that handle data processing",
  "color": "#6366f1"
}
```

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `title` | string | Yes | — | Team name (max 255 chars) |
| `description` | string | No | `""` | Team description |
| `color` | string | No | `#6366f1` | Hex color code |

Response `201`: Team object.

#### List Teams

```
GET /api/v1/teams
```

Response `200`: Array of team objects (ordered by created_at DESC).

#### Get / Update / Delete Team

```
GET /api/v1/teams/{id}
PATCH /api/v1/teams/{id}
DELETE /api/v1/teams/{id}
```

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

#### Purge Old Executions

```
DELETE /api/v1/executions?older_than=30d
DELETE /api/v1/executions?older_than=24h
DELETE /api/v1/executions?older_than=2026-01-01T00:00:00Z
```

| Param | Type | Required | Description |
|---|---|---|---|
| `older_than` | string | Yes | Duration (`30d`, `24h`) or RFC3339 timestamp |

Response `200`:
```json
{"deleted": 42}
```

Response `400`: Missing or invalid `older_than` parameter.

**Automatic cleanup:** A daily background job purges executions older than `EXECUTION_RETENTION_DAYS` (default: 30 days).

---

## Tech Stack

### Backend
- **Go 1.24+** with Chi router
- **SQLite** (embedded, via `modernc.org/sqlite` — pure Go, no CGO)
- **sqlc** for type-safe SQL query generation
- **Goose** for database migrations (run automatically on startup)

### Frontend (`maguro-dashboard/`)
- **React 18** + TypeScript
- **Vite** for dev server and builds
- **Tailwind CSS** + **shadcn/ui** component library
- **TanStack Query** for data fetching
- **React Router** for navigation

## Development

```bash
# Regenerate sqlc code after changing queries
sqlc generate

# Build backend
go build -o maguro ./cmd/api

# Install frontend dependencies
cd maguro-dashboard && npm install

# Start frontend dev server
cd maguro-dashboard && npm run dev
```

Migrations run automatically on server startup (embedded via `go:embed` + goose).

### Data Directory

All data lives under `~/.maguro/`:

```
~/.maguro/
├── maguro.db          # SQLite database
├── .secret_key        # Auto-generated encryption key
└── workspaces/        # Per-agent workspace directories
    └── {agent-id}/
```

## Project Structure

```
open-maguro/
├── cmd/api/              # Go API entry point
├── internal/             # Go backend packages
├── db/                   # Migrations and SQL queries
├── maguro-dashboard/     # React frontend
│   ├── src/components/   # UI views (Agents, Skills, Board, Logs)
│   ├── src/lib/api.ts    # Typed API client
│   └── src/pages/        # Route pages
└── scripts/              # Utility scripts
```

## Testing

### Backend
E2e API tests use a temporary SQLite database per test — no external dependencies needed.

```bash
# Run all tests
go test ./internal/tests/... -v

# Run without cache
go test ./internal/tests/... -v -count=1
```

### Frontend

```bash
cd maguro-dashboard
npm test            # run once
npm run test:watch  # watch mode
```