package tests

import (
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-playground/validator/v10"

	"open-maguro/internal/agent_task"
	"open-maguro/internal/crypto"
	"open-maguro/internal/database"
	"open-maguro/internal/executor"
	"open-maguro/internal/kanban"
	kanbanexec "open-maguro/internal/kanban_executor"
	"open-maguro/internal/mcp_config"
	"open-maguro/internal/router"
	"open-maguro/internal/scheduled_task"
	"open-maguro/internal/scheduler"
	"open-maguro/internal/skill"
	"open-maguro/internal/task_execution"
	"open-maguro/internal/team"
)

// migrationsDir returns the absolute path to db/migrations.
func migrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "db", "migrations")
}

// testWorkspaceRoot stores the workspace root for the current test so tests can inspect it.
var testWorkspaceRoot string

// GetWorkspaceRoot returns the workspace root used by the current test server.
func GetWorkspaceRoot(t *testing.T) string {
	t.Helper()
	return testWorkspaceRoot
}

// SetupTestServer creates an in-memory SQLite database, runs migrations,
// wires the full application, and returns an httptest.Server.
// The caller should defer cleanup().
func SetupTestServer(t *testing.T) (server *httptest.Server, cleanup func()) {
	t.Helper()

	// Use a temp file for SQLite (in-memory doesn't work well with multiple connections)
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Run goose migrations
	if err := runGooseMigrations(dbPath); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Connect
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	validate := validator.New()

	// Generate test encryption key
	testSecretKey, err := crypto.GenerateKey()
	if err != nil {
		db.Close()
		t.Fatalf("failed to generate test secret key: %v", err)
	}

	// Wire up repos
	agentTaskRepo := agent_task.NewPostgresRepository(db)
	taskExecRepo := task_execution.NewPostgresRepository(db)
	skillRepo := skill.NewPostgresRepository(db, testSecretKey)

	// Wire up executor (no real claude CLI in tests) and scheduler
	workspaceRoot := t.TempDir() + "/workspaces"
	testWorkspaceRoot = workspaceRoot
	exec := executor.New(taskExecRepo, skillRepo, agentTaskRepo, "", nil, workspaceRoot)
	sched := scheduler.New(agentTaskRepo, agentTaskRepo, taskExecRepo, exec, 0)

	// Wire up services and handlers
	agentTaskService := agent_task.NewService(agentTaskRepo, workspaceRoot)
	agentTaskHandler := agent_task.NewHandler(agentTaskService, validate,
		agent_task.WithOnTaskChanged(sched.Reload),
		agent_task.WithRunner(exec),
	)

	taskExecService := task_execution.NewService(taskExecRepo)
	taskExecHandler := task_execution.NewHandler(taskExecService)

	scheduledTaskService := scheduled_task.NewService(agentTaskRepo)
	scheduledTaskHandler := scheduled_task.NewHandler(scheduledTaskService, validate,
		scheduled_task.WithOnTaskChanged(sched.Reload),
	)

	mcpConfigService := mcp_config.NewService(t.TempDir() + "/mcp.json")
	mcpConfigHandler := mcp_config.NewHandler(mcpConfigService, validate)

	skillService := skill.NewService(skillRepo)
	skillHandler := skill.NewHandler(skillService, validate)

	teamRepo := team.NewPostgresRepository(db)
	teamService := team.NewService(teamRepo)
	teamHandler := team.NewHandler(teamService, validate)

	kanbanRepo := kanban.NewPostgresRepository(db)
	kanbanService := kanban.NewService(kanbanRepo)
	kExec := kanbanexec.New(kanbanRepo, agentTaskRepo, taskExecRepo, exec)
	kanbanHandler := kanban.NewHandler(kanbanService, validate,
		kanban.WithOnTaskCreated(kExec.Enqueue),
	)

	r := router.New(agentTaskHandler, taskExecHandler, scheduledTaskHandler, mcpConfigHandler, skillHandler, kanbanHandler, teamHandler)
	srv := httptest.NewServer(r)

	return srv, func() {
		srv.Close()
		sched.Stop()
		kExec.Stop()
		db.Close()
	}
}
