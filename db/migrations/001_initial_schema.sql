-- +goose Up

CREATE TABLE IF NOT EXISTS teams (
    id TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT '#6366f1',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS agent_tasks (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    cron_expression TEXT,
    prompt TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    task_type TEXT NOT NULL DEFAULT 'cron',
    run_at DATETIME,
    mcp_config TEXT,
    allowed_tools TEXT,
    system_agent BOOLEAN NOT NULL DEFAULT 0,
    global_skill_access BOOLEAN NOT NULL DEFAULT 0,
    on_success_task_id TEXT REFERENCES agent_tasks(id) ON DELETE SET NULL,
    on_failure_task_id TEXT REFERENCES agent_tasks(id) ON DELETE SET NULL,
    team_id TEXT REFERENCES teams(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_agent_tasks_enabled ON agent_tasks(enabled);

CREATE TABLE IF NOT EXISTS task_executions (
    id TEXT NOT NULL PRIMARY KEY,
    agent_task_id TEXT REFERENCES agent_tasks(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'success', 'failure', 'timeout')),
    started_at DATETIME,
    finished_at DATETIME,
    summary TEXT,
    error TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    task_name TEXT,
    triggered_by_execution_id TEXT REFERENCES task_executions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_task_executions_agent_task ON task_executions(agent_task_id);
CREATE INDEX IF NOT EXISTS idx_task_executions_status ON task_executions(status);
CREATE INDEX IF NOT EXISTS idx_task_executions_created ON task_executions(created_at DESC);

CREATE TABLE IF NOT EXISTS skills (
    id TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    environment_secrets TEXT
);

CREATE TABLE IF NOT EXISTS agent_skills (
    agent_task_id TEXT NOT NULL REFERENCES agent_tasks(id) ON DELETE CASCADE,
    skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    PRIMARY KEY (agent_task_id, skill_id)
);

CREATE TABLE IF NOT EXISTS kanban_tasks (
    id TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    agent_task_id TEXT NOT NULL REFERENCES agent_tasks(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'todo' CHECK (status IN ('todo', 'progress', 'done', 'failed')),
    result TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_kanban_tasks_agent ON kanban_tasks(agent_task_id);
CREATE INDEX IF NOT EXISTS idx_kanban_tasks_status ON kanban_tasks(status);

-- +goose Down

DROP TABLE IF EXISTS kanban_tasks;
DROP TABLE IF EXISTS agent_skills;
DROP TABLE IF EXISTS skills;
DROP TABLE IF EXISTS task_executions;
DROP TABLE IF EXISTS agent_tasks;
DROP TABLE IF EXISTS teams;
