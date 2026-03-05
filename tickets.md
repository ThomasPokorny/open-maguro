# OpenMaguro — Task Tracking

## Phase 1: Project Scaffolding (done)
- [x] Initialize go.mod with dependencies
- [x] Create docker-compose.yml for Postgres
- [x] Create Goose migration files for agent_tasks and task_executions
- [x] Set up sqlc configuration and write initial queries
- [x] Create domain entity structs
- [x] Create config package with env parsing
- [x] Create database connection package
- [x] Set up Chi router with health check endpoint
- [x] Create main.go entry point with graceful shutdown
- [x] AgentTask CRUD (handler/service/repo/dto)
- [x] TaskExecution read endpoints (handler/service/repo/dto)
- [x] Create CLAUDE.md

## Phase 2: Scheduling & Execution Engine (done)
- [x] Cron scheduler for enabled AgentTasks (robfig/cron/v3)
- [x] Claude CLI integration for agent execution (`claude --print --output-format json`)
- [x] Execution lifecycle management (pending -> running -> success/failure/timeout)
- [x] Timeout enforcement (context.WithTimeout per task)
- [x] Scheduler reload on task create/update/delete (callback pattern)
- [x] Graceful shutdown (stop scheduler -> drain HTTP -> close pool)

## Phase 3: Polish
- [ ] Error handling improvements (typed errors, consistent responses)
- [ ] Pagination for list endpoints
- [ ] Request logging middleware tuning
- [ ] Makefile for common commands
- [ ] API documentation
