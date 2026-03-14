package tests

import (
	"context"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"open-maguro/internal/agent_task"
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

func init() {
	// Auto-detect Podman socket if DOCKER_HOST is not set
	if os.Getenv("DOCKER_HOST") == "" {
		if out, err := exec.Command("podman", "machine", "inspect", "--format", "{{.ConnectionInfo.PodmanSocket.Path}}").Output(); err == nil {
			socket := strings.TrimSpace(string(out))
			if socket != "" {
				os.Setenv("DOCKER_HOST", "unix://"+socket)
			}
		}
	}
	// Ryuk doesn't work reliably with Podman
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}

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

// SetupTestServer spins up a Postgres testcontainer, runs migrations,
// wires the full application, and returns an httptest.Server.
// The caller should defer cleanup().
func SetupTestServer(t *testing.T) (server *httptest.Server, cleanup func()) {
	t.Helper()
	ctx := context.Background()

	// Start Postgres container
	pgContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("test_maguro"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		pgContainer.Terminate(ctx)
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Run goose migrations
	if err := runGooseMigrations(connStr); err != nil {
		pgContainer.Terminate(ctx)
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Connect pool
	pool, err := database.NewPool(ctx, connStr)
	if err != nil {
		pgContainer.Terminate(ctx)
		t.Fatalf("failed to create pool: %v", err)
	}

	validate := validator.New()

	// Wire up repos
	agentTaskRepo := agent_task.NewPostgresRepository(pool)
	taskExecRepo := task_execution.NewPostgresRepository(pool)
	skillRepo := skill.NewPostgresRepository(pool)

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

	teamRepo := team.NewPostgresRepository(pool)
	teamService := team.NewService(teamRepo)
	teamHandler := team.NewHandler(teamService, validate)

	kanbanRepo := kanban.NewPostgresRepository(pool)
	kanbanService := kanban.NewService(kanbanRepo)
	kExec := kanbanexec.New(kanbanRepo, agentTaskRepo, taskExecRepo, exec)
	kanbanHandler := kanban.NewHandler(kanbanService, validate,
		kanban.WithOnTaskCreated(kExec.Enqueue),
	)

	r := router.New(agentTaskHandler, taskExecHandler, scheduledTaskHandler, mcpConfigHandler, skillHandler, kanbanHandler, teamHandler)
	srv := httptest.NewServer(r)

	return srv, func() {
		kExec.Stop()
		srv.Close()
		pool.Close()
		pgContainer.Terminate(ctx)
	}
}
