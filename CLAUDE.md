# OpenMaguro

Scheduled Claude Code SDK agent task orchestrator with a REST API.

## Tech Stack
- Go 1.24+, Chi router, pgx/v5 + sqlc, Goose migrations
- PostgreSQL 17
- robfig/cron/v3 for task scheduling
- Claude CLI (`claude`) for agent execution — must be installed and on PATH

## Architecture
- 3-tier: handler (controller) / service / repository
- DTOs at the handler layer, domain entities for service/repo
- Feature-based package organization under internal/
- Repository interfaces defined by consumers (services), not producers

## Key Commands
- `docker compose up -d` — start Postgres
- `go run cmd/api/main.go` — start the API server
- `goose -dir db/migrations postgres "$DATABASE_URL" up` — run migrations
- `sqlc generate` — regenerate database code from queries
- `go test ./internal/tests/... -v` — run e2e API tests (requires Podman/Docker running)

## Testing
- E2e tests live in `internal/tests/` using testcontainers-go (spins up a real Postgres per test)
- Auto-detects Podman on macOS; also works with Docker
- `go test ./internal/tests/... -v -count=1` — run without cache
- **IMPORTANT**: After any significant change (new endpoints, schema changes, migrations, business logic, DTO changes), you MUST:
  1. Update or add tests in `internal/tests/api_test.go` to cover the change
  2. Run `go test ./internal/tests/... -v -count=1` to verify all tests pass
  3. Fix any failures before considering the change complete

## Environment Variables
- DATABASE_URL (required): Postgres connection string
- PORT (default: 8080): HTTP server port
- LOG_LEVEL (default: info): Logging level
- MCP_CONFIG_PATH: Path to global MCP config file (mcp.json). Used as default for all task executions unless overridden per-task.
- ALLOWED_TOOLS (default: `Bash(curl*),Bash(npx*),WebSearch,WebFetch,mcp__*`): Comma-separated list of tool patterns auto-approved for agent execution.
- WORKSPACE_ROOT (default: `~/.maguro/workspaces`): Root directory for per-agent workspaces. Each agent gets `{WORKSPACE_ROOT}/{agent-id}/`.
- EXECUTION_RETENTION_DAYS (default: 30): Number of days to keep execution logs. A daily cleanup purges older records automatically.
- MAGURO_SECRET_KEY: Hex-encoded 32-byte AES-256 key for encrypting skill secrets. Auto-generated to `~/.maguro/.secret_key` if not set.

## Project Layout
- cmd/api/ — application entry point
- internal/domain/ — entity structs (no dependencies)
- internal/agent_task/ — AgentTask feature (handler, service, repo, DTOs)
- internal/task_execution/ — TaskExecution feature
- internal/scheduled_task/ — One-time scheduled task endpoint
- internal/skill/ — Skills feature (CRUD + agent-skill associations)
- internal/kanban/ — Kanban task feature (CRUD, assigned to agents)
- internal/kanban_executor/ — Kanban worker pool (one worker per agent, sequential processing)
- internal/team/ — Team feature (CRUD, agent grouping with hex color)
- internal/crypto/ — AES-256-GCM encryption for skill secrets
- internal/mcp_config/ — MCP server config management (read/write mcp.json)
- internal/config/ — configuration loading
- internal/database/ — database connection pool
- internal/executor/ — Claude CLI execution (shells out to `claude` CLI)
- internal/scheduler/ — Cron + one-time task scheduler (triggers executor)
- internal/router/ — Chi router setup
- internal/sqlcgen/ — sqlc generated code (do not edit manually)
- db/migrations/ — Goose SQL migration files
- db/queries/ — sqlc SQL query files

## API Endpoints
- GET /health — health check
- POST /api/v1/agent-tasks — create agent task
- GET /api/v1/agent-tasks — list agent tasks
- GET /api/v1/agent-tasks/{id} — get agent task
- PATCH /api/v1/agent-tasks/{id} — partial update agent task
- DELETE /api/v1/agent-tasks/{id} — delete agent task
- POST /api/v1/agent-tasks/{id}/run — trigger immediate execution of agent task
- POST /api/v1/agent-tasks/{id}/open-workspace — open agent's workspace directory in file explorer
- POST /api/v1/scheduled-tasks — create one-time scheduled task (auto-deletes after execution)
- Agent chaining: `on_success_task_id` / `on_failure_task_id` on agent tasks trigger follow-up agents with parent output as context
- Heartbeat: scheduler checks every 10 min for missed cron jobs (24h lookback) and marks stale executions (>2h running) as failed
- GET /api/v1/mcp-servers — list configured MCP servers
- POST /api/v1/mcp-servers — add an MCP server to global config
- DELETE /api/v1/mcp-servers/{name} — remove an MCP server from global config
- POST /api/v1/skills — create skill
- GET /api/v1/skills — list skills
- GET /api/v1/skills/{id} — get skill
- PATCH /api/v1/skills/{id} — partial update skill
- DELETE /api/v1/skills/{id} — delete skill
- GET /api/v1/agent-tasks/{id}/skills — list skills for agent
- POST /api/v1/agent-tasks/{id}/skills/{skillId} — attach skill to agent
- DELETE /api/v1/agent-tasks/{id}/skills/{skillId} — detach skill from agent
- GET /api/v1/agent-tasks/{taskId}/executions — list executions for a task
- GET /api/v1/executions — list all executions (includes orphaned one-shot task logs + kanban executions)
- GET /api/v1/executions/{id} — get execution by id
- DELETE /api/v1/executions?older_than= — purge old executions (accepts RFC3339 timestamp or duration: '30d', '24h')
- POST /api/v1/kanban-tasks — create kanban task (title, description, agent_task_id)
- GET /api/v1/kanban-tasks — list kanban tasks (done >2h filtered, ?agent_id= ?status= ?team_id= filters)
- GET /api/v1/kanban-tasks/{id} — get kanban task
- PATCH /api/v1/kanban-tasks/{id} — update kanban task
- DELETE /api/v1/kanban-tasks/{id} — delete kanban task
- POST /api/v1/teams — create team (title, description, color)
- GET /api/v1/teams — list teams
- GET /api/v1/teams/{id} — get team
- PATCH /api/v1/teams/{id} — partial update team
- DELETE /api/v1/teams/{id} — delete team (unassigns agents via ON DELETE SET NULL)
- Agent tasks support `team_id` field for team assignment, filterable via `?team_id=` on list endpoints

## Conventions
- Use log/slog for all logging
- All handler functions are standard http.HandlerFunc
- Nullable DB fields use pointer types in Go structs
- All timestamps are TIMESTAMPTZ in Postgres and time.Time in Go
- API routes are under /api/v1
- Health check at GET /health
