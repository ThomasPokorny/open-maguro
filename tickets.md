## Maguro Tasks

[x] one shot tasks execution logs deleted — fixed: migration was correct (ON DELETE SET NULL), added `GET /api/v1/executions` endpoint for discoverability
[x] add a selection of allowed tools to the agent config — added `allowed_tools` field (comma-separated, additive to global `ALLOWED_TOOLS` env var)
[x] create a new agent property called "system_agent" — boolean field (default false) to distinguish internal system agents from user agents
[x] update documentation accordingly — CLAUDE.md, README.md, CLAUDE-AGENT.md updated
[x] in depth api testing — testcontainers-go + httptest, 6 e2e tests (health, CRUD, scheduled tasks, system agent, MCP servers, executions), CLAUDE.md updated with test commands
[x] goose migrations should run automatically on every startup of the application — embedded via go:embed, runs goose.Up in main.go before pool creation
[x] create skills api — CRUD at /api/v1/skills (title + content TEXT), migration 008, internal/skill/ package
[x] skills can be added to agents — agent_skills join table, POST/DELETE /api/v1/agent-tasks/{id}/skills/{skillId}
[x] global skill access — global_skill_access bool on agent_tasks, threaded through full stack
[x] add skills to execution prompt — executor injects skills as "These are your skills:\n## Title\ncontent" prefix
[x] update agent md, readme and claude md — all updated with skill management API and guidance
[x] add two seed skills with obfuscated api keys (Slack API + Linear GraphQL) — scripts/seed_skills.sh with Slack API and Linear GraphQL skills
[x] create agent run api which allows one to execute agents immediately. add information about it in the agent md file and readme
[x] add system prompt: 
# OpenMaguro🐟 Agent Orchestrator.
"You are an agent orchestrated by the `OpenMaguro🐟 Agent Orchestrator` project. Similarly to OpenClaw, users can create agents and schedule them fulfilling different tasks. This is a task running in the background. So there is no means of getting additional tool calls whitelisted. Try to fulfill the user request by all means."
to the beginning of every agent execution
[x] Brainstorm further features for OpenMaguro and add a list including required implementation tasks (including testing) to a section below:
[x] When my macbook (or any other machine currently running open maguro) is down, a missed cornjob will just be lost, we need a heartbeat every 10minutes checking for agents lost and in a somewhat dead state — heartbeat loop in scheduler (10min interval, 24h cron lookback, 2h stale execution marking)
[x] ### 6. Agent Chaining (on_success / on_failure triggers)
Enable workflows where one agent triggers another.
- [x] Add migration: `on_success_task_id` and `on_failure_task_id` nullable FK columns on agent_tasks
- [x] After execution completes, check triggers and run the linked task
- [x] Prevent circular chains (validate on create/update)
- [x] Add `triggered_by_execution_id` to task_executions for traceability
- [x] Add tests for chain execution and cycle detection
- [x] Update docs
- [x] Create new lovable prompt to also reflect those changes. The agent configuration ui, should enable one to create those agents
[x] Kanban Task Queue — agents pick up assigned work items sequentially, maintain work-log.md in workspace. Migration 010, internal/kanban/ + internal/kanban_executor/, full CRUD API with status filters, auto-pickup on create, restart recovery.
[x] Add `EXECUTION_RETENTION_DAYS` env var (default: 30)
[x] Add a periodic cleanup goroutine (runs daily) that deletes executions older than retention period
[x] Add `DELETE /api/v1/executions` endpoint for manual purge with optional `?older_than=` query param
[x] Add test: verify old executions are cleaned up
[x] Update docs
[x] Agent Teams - As a user I want to be able to create teams - and assign agents to them. A team should have a title and a optional description. each agent can be in one swarm team exactly (null able fk on agent). furthermore teams can have a hex color
- [x] migration and crud api for teams
- [x] possibility to assign agent to team
- [x] possibility to filter agent by team
- [x] possibility to filter tasks by team
### 9. Execution Metrics & Observability
No visibility into system health or execution costs.
- [] Add `duration_ms` computed field to execution responses
- [] Add `GET /api/v1/stats` endpoint: total tasks, executions by status (last 24h/7d/30d), avg duration
- [] Optionally expose Prometheus-compatible `/metrics` endpoint
- [] Add tests for stats endpoint
- [] Update docs

## Feature Backlog

### 1. Execution Concurrency Limits
Currently unlimited goroutines can spawn. A worker pool with a configurable max prevents resource exhaustion.
- [] Add `MAX_CONCURRENT_EXECUTIONS` env var to config (default: 5)
- [] Implement a semaphore/worker pool in the executor
- [] Queue excess executions and process them in order
- [] Add test: verify tasks are queued when pool is full
- [] Update CLAUDE.md and README with new config

### 2. Execution Timeout
Tasks can hang indefinitely. Per-task timeouts give control back.
- [] Add migration: `ALTER TABLE agent_tasks ADD COLUMN timeout_seconds INTEGER`
- [] Add `timeout_seconds` field to domain, DTOs, sqlc queries
- [] Enforce timeout via `context.WithTimeout` in executor before calling claude CLI
- [] Record `timeout` status when exceeded (status already exists in domain)
- [] Add tests for timeout behavior
- [] Update docs

### 3. Retry on Failure
Failed tasks stay failed. Automatic retries with backoff improve reliability.
- [] Add migration: `max_retries` (int, default 0) and `retry_delay_seconds` (int, default 60) columns on agent_tasks
- [] Add `attempt_number` column on task_executions
- [] Implement retry logic in executor: on failure, schedule re-execution up to max_retries
- [] Add `triggered_by` field to executions (`cron`, `api`, `retry`) for traceability
- [] Add tests for retry behavior (succeeds after retry, gives up after max)
- [] Update docs

### 4. Execution Log Retention & Cleanup ✅ DONE

### 5. Webhook Notifications
No way for external systems to know when tasks complete.
- [] Add migration: create `webhooks` table (id, url, events, created_at)
- [] Add `internal/webhook/` package (handler, service, repository, DTOs)
- [] CRUD endpoints: `POST/GET/DELETE /api/v1/webhooks`
- [] Fire HTTP POST to registered webhooks on execution status changes (success, failure)
- [] Payload: `{event, agent_task_id, execution_id, status, task_name, timestamp}`
- [] Fire-and-forget with short timeout, log failures
- [] Add tests for webhook CRUD and delivery
- [] Update docs

### 6. Agent Chaining (on_success / on_failure triggers) ✅ DONE

### 7. Pagination & Filtering
All list endpoints return everything. Pagination is needed at scale.
- [] Add `?limit=`, `?offset=`, `?status=` query params to execution list endpoints
- [] Add `?enabled=`, `?system_agent=` filters to agent task list
- [] Return pagination metadata in response (`total`, `limit`, `offset`)
- [] Update sqlc queries with dynamic filtering
- [] Add tests for filtered and paginated responses
- [] Update docs


### 10. Agent Workspaces (Per-Agent Disk Directory) ✅ DONE
Give every agent its own persistent workspace directory where it can read/write files across runs.
- [x] Add `WORKSPACE_ROOT` env var to config (default: `~/.maguro/workspaces`)
- [x] Create `{WORKSPACE_ROOT}/{agent-id}/` directory when agent is created
- [x] Delete `{WORKSPACE_ROOT}/{agent-id}/` directory when agent is deleted
- [x] Executor sets `cmd.Dir` to agent's workspace before running claude CLI
- [x] System prompt tells the agent about its workspace path
- [x] Add tests for workspace lifecycle (create/delete)
- [x] Update docs (CLAUDE.md, README.md, CLAUDE-AGENT.md)

### 11. Agent Templates
Reduce boilerplate for common agent patterns.
- [] Add migration: create `agent_templates` table (id, name, prompt_template, default_cron, default_tools)
- [] CRUD endpoints: `POST/GET/PATCH/DELETE /api/v1/templates`
- [] `POST /api/v1/agent-tasks` accepts optional `template_id` to pre-fill fields
- [] Add tests for template CRUD and agent creation from template
- [] Update docs