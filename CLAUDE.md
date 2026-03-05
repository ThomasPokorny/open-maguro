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

## Environment Variables
- DATABASE_URL (required): Postgres connection string
- PORT (default: 8080): HTTP server port
- LOG_LEVEL (default: info): Logging level

## Project Layout
- cmd/api/ — application entry point
- internal/domain/ — entity structs (no dependencies)
- internal/agent_task/ — AgentTask feature (handler, service, repo, DTOs)
- internal/task_execution/ — TaskExecution feature
- internal/scheduled_task/ — One-time scheduled task endpoint
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
- POST /api/v1/scheduled-tasks — create one-time scheduled task (auto-deletes after execution)
- GET /api/v1/agent-tasks/{taskId}/executions — list executions for a task
- GET /api/v1/executions/{id} — get execution by id

## Conventions
- Use log/slog for all logging
- All handler functions are standard http.HandlerFunc
- Nullable DB fields use pointer types in Go structs
- All timestamps are TIMESTAMPTZ in Postgres and time.Time in Go
- API routes are under /api/v1
- Health check at GET /health
