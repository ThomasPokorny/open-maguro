package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestHealthCheck(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("health check request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %s", body["status"])
	}
}

func TestAgentTaskCRUD(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create
	createBody := `{
		"name": "Test cron task",
		"cron_expression": "0 9 * * *",
		"prompt": "Say hello"
	}`
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", createBody)
	assertStatus(t, resp, http.StatusCreated)

	var created map[string]any
	parseJSON(t, resp, &created)

	id := created["id"].(string)
	if created["name"] != "Test cron task" {
		t.Fatalf("expected name 'Test cron task', got %v", created["name"])
	}
	if created["task_type"] != "cron" {
		t.Fatalf("expected task_type 'cron', got %v", created["task_type"])
	}
	if created["enabled"] != true {
		t.Fatalf("expected enabled true, got %v", created["enabled"])
	}
	if created["system_agent"] != false {
		t.Fatalf("expected system_agent false, got %v", created["system_agent"])
	}

	// Get
	resp = doRequest(t, "GET", srv.URL+"/api/v1/agent-tasks/"+id, "")
	assertStatus(t, resp, http.StatusOK)

	var fetched map[string]any
	parseJSON(t, resp, &fetched)
	if fetched["id"] != id {
		t.Fatalf("expected id %s, got %v", id, fetched["id"])
	}

	// List
	resp = doRequest(t, "GET", srv.URL+"/api/v1/agent-tasks", "")
	assertStatus(t, resp, http.StatusOK)

	var list []map[string]any
	parseJSON(t, resp, &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 task, got %d", len(list))
	}

	// Update
	resp = doRequest(t, "PATCH", srv.URL+"/api/v1/agent-tasks/"+id, `{"name": "Updated name", "enabled": false}`)
	assertStatus(t, resp, http.StatusOK)

	var updated map[string]any
	parseJSON(t, resp, &updated)
	if updated["name"] != "Updated name" {
		t.Fatalf("expected name 'Updated name', got %v", updated["name"])
	}
	if updated["enabled"] != false {
		t.Fatalf("expected enabled false, got %v", updated["enabled"])
	}

	// Delete
	resp = doRequest(t, "DELETE", srv.URL+"/api/v1/agent-tasks/"+id, "")
	assertStatus(t, resp, http.StatusNoContent)

	// Verify deleted
	resp = doRequest(t, "GET", srv.URL+"/api/v1/agent-tasks/"+id, "")
	assertStatus(t, resp, http.StatusNotFound)
}

func TestScheduledTaskCreation(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	runAt := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
	createBody := `{
		"name": "One-time task",
		"prompt": "Do something once",
		"run_at": "` + runAt + `"
	}`
	resp := doRequest(t, "POST", srv.URL+"/api/v1/scheduled-tasks", createBody)
	assertStatus(t, resp, http.StatusCreated)

	var created map[string]any
	parseJSON(t, resp, &created)
	if created["task_type"] != "one_time" {
		t.Fatalf("expected task_type 'one_time', got %v", created["task_type"])
	}
}

func TestSystemAgentFlag(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create system agent
	createBody := `{
		"name": "Lifeline check",
		"cron_expression": "*/5 * * * *",
		"prompt": "Check system health",
		"system_agent": true
	}`
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", createBody)
	assertStatus(t, resp, http.StatusCreated)

	var created map[string]any
	parseJSON(t, resp, &created)
	if created["system_agent"] != true {
		t.Fatalf("expected system_agent true, got %v", created["system_agent"])
	}
}

func TestMCPServerManagement(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// List (empty)
	resp := doRequest(t, "GET", srv.URL+"/api/v1/mcp-servers", "")
	assertStatus(t, resp, http.StatusOK)

	var list []map[string]any
	parseJSON(t, resp, &list)
	if len(list) != 0 {
		t.Fatalf("expected 0 MCP servers, got %d", len(list))
	}

	// Add
	addBody := `{
		"name": "test-mcp",
		"command": "npx",
		"args": ["-y", "test-mcp-server"],
		"env": {"API_KEY": "test123"}
	}`
	resp = doRequest(t, "POST", srv.URL+"/api/v1/mcp-servers", addBody)
	assertStatus(t, resp, http.StatusCreated)

	// List (1 server)
	resp = doRequest(t, "GET", srv.URL+"/api/v1/mcp-servers", "")
	assertStatus(t, resp, http.StatusOK)
	parseJSON(t, resp, &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 MCP server, got %d", len(list))
	}
	if list[0]["name"] != "test-mcp" {
		t.Fatalf("expected name 'test-mcp', got %v", list[0]["name"])
	}

	// Remove
	resp = doRequest(t, "DELETE", srv.URL+"/api/v1/mcp-servers/test-mcp", "")
	assertStatus(t, resp, http.StatusNoContent)

	// Verify removed
	resp = doRequest(t, "GET", srv.URL+"/api/v1/mcp-servers", "")
	assertStatus(t, resp, http.StatusOK)
	parseJSON(t, resp, &list)
	if len(list) != 0 {
		t.Fatalf("expected 0 MCP servers after removal, got %d", len(list))
	}
}

func TestSkillCRUD(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create
	createBody := `{
		"title": "Test Skill",
		"content": "This is a test skill with instructions."
	}`
	resp := doRequest(t, "POST", srv.URL+"/api/v1/skills", createBody)
	assertStatus(t, resp, http.StatusCreated)

	var created map[string]any
	parseJSON(t, resp, &created)

	id := created["id"].(string)
	if created["title"] != "Test Skill" {
		t.Fatalf("expected title 'Test Skill', got %v", created["title"])
	}

	// Get
	resp = doRequest(t, "GET", srv.URL+"/api/v1/skills/"+id, "")
	assertStatus(t, resp, http.StatusOK)

	var fetched map[string]any
	parseJSON(t, resp, &fetched)
	if fetched["id"] != id {
		t.Fatalf("expected id %s, got %v", id, fetched["id"])
	}

	// List
	resp = doRequest(t, "GET", srv.URL+"/api/v1/skills", "")
	assertStatus(t, resp, http.StatusOK)

	var list []map[string]any
	parseJSON(t, resp, &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(list))
	}

	// Update (partial)
	resp = doRequest(t, "PATCH", srv.URL+"/api/v1/skills/"+id, `{"title": "Updated Skill"}`)
	assertStatus(t, resp, http.StatusOK)

	var updated map[string]any
	parseJSON(t, resp, &updated)
	if updated["title"] != "Updated Skill" {
		t.Fatalf("expected title 'Updated Skill', got %v", updated["title"])
	}
	if updated["content"] != "This is a test skill with instructions." {
		t.Fatalf("expected content preserved, got %v", updated["content"])
	}

	// Delete
	resp = doRequest(t, "DELETE", srv.URL+"/api/v1/skills/"+id, "")
	assertStatus(t, resp, http.StatusNoContent)

	// Verify deleted
	resp = doRequest(t, "GET", srv.URL+"/api/v1/skills/"+id, "")
	assertStatus(t, resp, http.StatusNotFound)
}

func TestAgentSkillAssociation(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create a skill
	resp := doRequest(t, "POST", srv.URL+"/api/v1/skills", `{"title": "Slack API", "content": "Use Slack API to send messages."}`)
	assertStatus(t, resp, http.StatusCreated)
	var skill map[string]any
	parseJSON(t, resp, &skill)
	skillID := skill["id"].(string)

	// Create an agent task
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Slack bot",
		"cron_expression": "0 9 * * *",
		"prompt": "Send morning greeting"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var task map[string]any
	parseJSON(t, resp, &task)
	taskID := task["id"].(string)

	// Attach skill to agent
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks/"+taskID+"/skills/"+skillID, "")
	assertStatus(t, resp, http.StatusNoContent)

	// List skills for agent
	resp = doRequest(t, "GET", srv.URL+"/api/v1/agent-tasks/"+taskID+"/skills", "")
	assertStatus(t, resp, http.StatusOK)
	var skills []map[string]any
	parseJSON(t, resp, &skills)
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0]["title"] != "Slack API" {
		t.Fatalf("expected title 'Slack API', got %v", skills[0]["title"])
	}

	// Detach skill
	resp = doRequest(t, "DELETE", srv.URL+"/api/v1/agent-tasks/"+taskID+"/skills/"+skillID, "")
	assertStatus(t, resp, http.StatusNoContent)

	// Verify detached
	resp = doRequest(t, "GET", srv.URL+"/api/v1/agent-tasks/"+taskID+"/skills", "")
	assertStatus(t, resp, http.StatusOK)
	parseJSON(t, resp, &skills)
	if len(skills) != 0 {
		t.Fatalf("expected 0 skills after detach, got %d", len(skills))
	}
}

func TestGlobalSkillAccessFlag(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	createBody := `{
		"name": "Global agent",
		"cron_expression": "0 * * * *",
		"prompt": "Do things",
		"global_skill_access": true
	}`
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", createBody)
	assertStatus(t, resp, http.StatusCreated)

	var created map[string]any
	parseJSON(t, resp, &created)
	if created["global_skill_access"] != true {
		t.Fatalf("expected global_skill_access true, got %v", created["global_skill_access"])
	}
}

func TestAgentTaskRun(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create an agent task
	createBody := `{
		"name": "Runnable task",
		"cron_expression": "0 9 * * *",
		"prompt": "Say hello"
	}`
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", createBody)
	assertStatus(t, resp, http.StatusCreated)

	var created map[string]any
	parseJSON(t, resp, &created)
	id := created["id"].(string)

	// Run the task immediately
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks/"+id+"/run", "")
	assertStatus(t, resp, http.StatusAccepted)

	var runResp map[string]any
	parseJSON(t, resp, &runResp)
	if runResp["status"] != "accepted" {
		t.Fatalf("expected status 'accepted', got %v", runResp["status"])
	}

	// Run with invalid ID returns 400
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks/not-a-uuid/run", "")
	assertStatus(t, resp, http.StatusBadRequest)

	// Run with non-existent ID returns 404
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks/00000000-0000-0000-0000-000000000000/run", "")
	assertStatus(t, resp, http.StatusNotFound)
}

func TestAgentChaining(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create three agents
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Agent A",
		"cron_expression": "0 9 * * *",
		"prompt": "Do step A"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agentA map[string]any
	parseJSON(t, resp, &agentA)
	agentAID := agentA["id"].(string)

	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Agent B",
		"cron_expression": "0 10 * * *",
		"prompt": "Do step B"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agentB map[string]any
	parseJSON(t, resp, &agentB)
	agentBID := agentB["id"].(string)

	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Error Handler",
		"cron_expression": "0 11 * * *",
		"prompt": "Handle errors"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agentC map[string]any
	parseJSON(t, resp, &agentC)
	agentCID := agentC["id"].(string)

	// Set Agent A's on_success to Agent B and on_failure to Agent C
	resp = doRequest(t, "PATCH", srv.URL+"/api/v1/agent-tasks/"+agentAID, `{
		"on_success_task_id": "`+agentBID+`",
		"on_failure_task_id": "`+agentCID+`"
	}`)
	assertStatus(t, resp, http.StatusOK)
	var updated map[string]any
	parseJSON(t, resp, &updated)
	if updated["on_success_task_id"] != agentBID {
		t.Fatalf("expected on_success_task_id %s, got %v", agentBID, updated["on_success_task_id"])
	}
	if updated["on_failure_task_id"] != agentCID {
		t.Fatalf("expected on_failure_task_id %s, got %v", agentCID, updated["on_failure_task_id"])
	}

	// Verify via GET
	resp = doRequest(t, "GET", srv.URL+"/api/v1/agent-tasks/"+agentAID, "")
	assertStatus(t, resp, http.StatusOK)
	var fetched map[string]any
	parseJSON(t, resp, &fetched)
	if fetched["on_success_task_id"] != agentBID {
		t.Fatalf("expected on_success_task_id %s on GET, got %v", agentBID, fetched["on_success_task_id"])
	}
	if fetched["on_failure_task_id"] != agentCID {
		t.Fatalf("expected on_failure_task_id %s on GET, got %v", agentCID, fetched["on_failure_task_id"])
	}

	// Verify chaining fields appear in list response
	resp = doRequest(t, "GET", srv.URL+"/api/v1/agent-tasks", "")
	assertStatus(t, resp, http.StatusOK)
	var list []map[string]any
	parseJSON(t, resp, &list)
	found := false
	for _, task := range list {
		if task["id"] == agentAID {
			found = true
			if task["on_success_task_id"] != agentBID {
				t.Fatalf("expected on_success_task_id in list, got %v", task["on_success_task_id"])
			}
		}
	}
	if !found {
		t.Fatalf("agent A not found in list")
	}
}

func TestAgentChainingOnCreate(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create target agent first
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Target Agent",
		"cron_expression": "0 9 * * *",
		"prompt": "Target task"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var target map[string]any
	parseJSON(t, resp, &target)
	targetID := target["id"].(string)

	// Create agent with chaining set at creation time
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Source Agent",
		"cron_expression": "0 8 * * *",
		"prompt": "Source task",
		"on_success_task_id": "`+targetID+`"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var created map[string]any
	parseJSON(t, resp, &created)
	if created["on_success_task_id"] != targetID {
		t.Fatalf("expected on_success_task_id %s on create, got %v", targetID, created["on_success_task_id"])
	}
}

func TestAgentChainingCycleDetection(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create Agent A
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Agent A",
		"cron_expression": "0 9 * * *",
		"prompt": "Do A"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agentA map[string]any
	parseJSON(t, resp, &agentA)
	agentAID := agentA["id"].(string)

	// Create Agent B with on_success pointing to A
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Agent B",
		"cron_expression": "0 10 * * *",
		"prompt": "Do B",
		"on_success_task_id": "`+agentAID+`"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agentB map[string]any
	parseJSON(t, resp, &agentB)
	agentBID := agentB["id"].(string)

	// Try to set A's on_success to B — should fail with 409 (A -> B -> A cycle)
	resp = doRequest(t, "PATCH", srv.URL+"/api/v1/agent-tasks/"+agentAID, `{
		"on_success_task_id": "`+agentBID+`"
	}`)
	assertStatus(t, resp, http.StatusConflict)
	var errResp map[string]any
	parseJSON(t, resp, &errResp)
	errMsg, ok := errResp["error"].(string)
	if !ok || errMsg == "" {
		t.Fatalf("expected error message about circular chain, got %v", errResp)
	}

	// Also test on_failure cycle: Create Agent C -> A, then try A.on_failure = C
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Agent C",
		"cron_expression": "0 11 * * *",
		"prompt": "Do C",
		"on_failure_task_id": "`+agentAID+`"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agentC map[string]any
	parseJSON(t, resp, &agentC)
	agentCID := agentC["id"].(string)

	resp = doRequest(t, "PATCH", srv.URL+"/api/v1/agent-tasks/"+agentAID, `{
		"on_failure_task_id": "`+agentCID+`"
	}`)
	assertStatus(t, resp, http.StatusConflict)
	resp.Body.Close()
}

func TestExecutionsList(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// List all executions (empty)
	resp := doRequest(t, "GET", srv.URL+"/api/v1/executions", "")
	assertStatus(t, resp, http.StatusOK)

	var list []map[string]any
	parseJSON(t, resp, &list)
	if len(list) != 0 {
		t.Fatalf("expected 0 executions, got %d", len(list))
	}
}

func TestAgentWorkspaceLifecycle(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create an agent
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Workspace agent",
		"cron_expression": "0 9 * * *",
		"prompt": "Use workspace"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var created map[string]any
	parseJSON(t, resp, &created)
	agentID := created["id"].(string)

	// Verify workspace directory was created
	workspaceDir := GetWorkspaceRoot(t) + "/" + agentID
	info, err := os.Stat(workspaceDir)
	if err != nil {
		t.Fatalf("expected workspace directory to exist at %s, got error: %v", workspaceDir, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %s to be a directory", workspaceDir)
	}

	// Delete the agent
	resp = doRequest(t, "DELETE", srv.URL+"/api/v1/agent-tasks/"+agentID, "")
	assertStatus(t, resp, http.StatusNoContent)

	// Verify workspace directory was removed
	if _, err := os.Stat(workspaceDir); !os.IsNotExist(err) {
		t.Fatalf("expected workspace directory to be removed after agent delete, but it still exists")
	}
}

func TestExecutionResponseShape(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create and run an agent to generate an execution
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Shape test agent",
		"cron_expression": "0 9 * * *",
		"prompt": "Say hello"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var task map[string]any
	parseJSON(t, resp, &task)
	taskID := task["id"].(string)

	// Trigger execution (will fail since no claude CLI, but creates the record)
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks/"+taskID+"/run", "")
	assertStatus(t, resp, http.StatusAccepted)
	resp.Body.Close()

	// Wait briefly for the execution record to be created
	time.Sleep(500 * time.Millisecond)

	// Check executions for this agent
	resp = doRequest(t, "GET", srv.URL+"/api/v1/agent-tasks/"+taskID+"/executions", "")
	assertStatus(t, resp, http.StatusOK)
	var executions []map[string]any
	parseJSON(t, resp, &executions)

	if len(executions) == 0 {
		t.Skip("no execution record created (expected in test env without claude CLI)")
	}

	// Verify response shape includes triggered_by_execution_id
	exec := executions[0]
	requiredFields := []string{"id", "status", "created_at"}
	for _, field := range requiredFields {
		if _, ok := exec[field]; !ok {
			t.Fatalf("execution response missing required field: %s", field)
		}
	}
	// triggered_by_execution_id should be absent (omitempty) or null for non-chained executions
	if val, ok := exec["triggered_by_execution_id"]; ok && val != nil {
		t.Fatalf("expected triggered_by_execution_id to be nil for non-chained execution, got %v", val)
	}
}

func TestKanbanTaskCRUD(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create an agent task first (prerequisite)
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Kanban Worker",
		"cron_expression": "0 9 * * *",
		"prompt": "Process tasks"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agent map[string]any
	parseJSON(t, resp, &agent)
	agentID := agent["id"].(string)

	// Create kanban task
	resp = doRequest(t, "POST", srv.URL+"/api/v1/kanban-tasks", `{
		"title": "Write report",
		"description": "Generate the Q1 report",
		"agent_task_id": "`+agentID+`"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var created map[string]any
	parseJSON(t, resp, &created)
	kanbanID := created["id"].(string)

	if created["title"] != "Write report" {
		t.Fatalf("expected title 'Write report', got %v", created["title"])
	}
	if created["status"] != "todo" {
		t.Fatalf("expected status 'todo', got %v", created["status"])
	}
	if created["agent_task_id"] != agentID {
		t.Fatalf("expected agent_task_id %s, got %v", agentID, created["agent_task_id"])
	}

	// Get by ID
	resp = doRequest(t, "GET", srv.URL+"/api/v1/kanban-tasks/"+kanbanID, "")
	assertStatus(t, resp, http.StatusOK)
	var fetched map[string]any
	parseJSON(t, resp, &fetched)
	if fetched["id"] != kanbanID {
		t.Fatalf("expected id %s, got %v", kanbanID, fetched["id"])
	}

	// List all
	resp = doRequest(t, "GET", srv.URL+"/api/v1/kanban-tasks", "")
	assertStatus(t, resp, http.StatusOK)
	var list []map[string]any
	parseJSON(t, resp, &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 kanban task, got %d", len(list))
	}

	// List with agent_id filter
	resp = doRequest(t, "GET", srv.URL+"/api/v1/kanban-tasks?agent_id="+agentID, "")
	assertStatus(t, resp, http.StatusOK)
	parseJSON(t, resp, &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 kanban task for agent, got %d", len(list))
	}

	// List with status filter (task may have been picked up by executor already,
	// so check that status filter works rather than asserting specific counts)
	resp = doRequest(t, "GET", srv.URL+"/api/v1/kanban-tasks?status=done", "")
	assertStatus(t, resp, http.StatusOK)
	// Just verify the endpoint works with filters — don't assert counts since
	// the kanban executor races with the test

	// Update
	resp = doRequest(t, "PATCH", srv.URL+"/api/v1/kanban-tasks/"+kanbanID, `{
		"title": "Updated report task"
	}`)
	assertStatus(t, resp, http.StatusOK)
	var updated map[string]any
	parseJSON(t, resp, &updated)
	if updated["title"] != "Updated report task" {
		t.Fatalf("expected title 'Updated report task', got %v", updated["title"])
	}
	if updated["description"] != "Generate the Q1 report" {
		t.Fatalf("expected description preserved, got %v", updated["description"])
	}

	// Delete
	resp = doRequest(t, "DELETE", srv.URL+"/api/v1/kanban-tasks/"+kanbanID, "")
	assertStatus(t, resp, http.StatusNoContent)

	// Verify deleted
	resp = doRequest(t, "GET", srv.URL+"/api/v1/kanban-tasks/"+kanbanID, "")
	assertStatus(t, resp, http.StatusNotFound)
}

func TestKanbanTaskPickup(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create an agent
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Auto Worker",
		"cron_expression": "0 9 * * *",
		"prompt": "You are a task worker"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agent map[string]any
	parseJSON(t, resp, &agent)
	agentID := agent["id"].(string)

	// Create a kanban task — the executor should pick it up
	resp = doRequest(t, "POST", srv.URL+"/api/v1/kanban-tasks", `{
		"title": "Auto task",
		"description": "Should be picked up automatically",
		"agent_task_id": "`+agentID+`"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var created map[string]any
	parseJSON(t, resp, &created)
	kanbanID := created["id"].(string)

	// Wait for the kanban executor to process (will fail since no claude CLI,
	// but should transition from todo -> progress -> failed)
	time.Sleep(1 * time.Second)

	resp = doRequest(t, "GET", srv.URL+"/api/v1/kanban-tasks/"+kanbanID, "")
	assertStatus(t, resp, http.StatusOK)
	var result map[string]any
	parseJSON(t, resp, &result)

	// Should have been picked up (moved from todo to progress or failed)
	status := result["status"].(string)
	if status == "todo" {
		t.Fatalf("expected kanban task to be picked up (progress or failed), but still 'todo'")
	}
}

func TestAgentTaskWithoutCron(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create agent without cron_expression (kanban-only agent)
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Kanban-only agent",
		"prompt": "Process kanban tasks"
	}`)
	assertStatus(t, resp, http.StatusCreated)

	var created map[string]any
	parseJSON(t, resp, &created)
	id := created["id"].(string)

	if created["name"] != "Kanban-only agent" {
		t.Fatalf("expected name 'Kanban-only agent', got %v", created["name"])
	}
	if created["task_type"] != "cron" {
		t.Fatalf("expected task_type 'cron', got %v", created["task_type"])
	}
	// cron_expression should be absent (omitempty) or null
	if val, ok := created["cron_expression"]; ok && val != nil {
		t.Fatalf("expected cron_expression to be nil, got %v", val)
	}

	// Should be usable for kanban tasks
	resp = doRequest(t, "POST", srv.URL+"/api/v1/kanban-tasks", `{
		"title": "Test task for kanban-only agent",
		"agent_task_id": "`+id+`"
	}`)
	assertStatus(t, resp, http.StatusCreated)

	// Verify the agent can be fetched
	resp = doRequest(t, "GET", srv.URL+"/api/v1/agent-tasks/"+id, "")
	assertStatus(t, resp, http.StatusOK)
	var fetched map[string]any
	parseJSON(t, resp, &fetched)
	if val, ok := fetched["cron_expression"]; ok && val != nil {
		t.Fatalf("expected cron_expression to be nil on GET, got %v", val)
	}
}

func TestExecutionPurge(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create an agent and trigger execution to generate a record
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Purge test agent",
		"cron_expression": "0 9 * * *",
		"prompt": "Say hello"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agent map[string]any
	parseJSON(t, resp, &agent)
	agentID := agent["id"].(string)

	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks/"+agentID+"/run", "")
	assertStatus(t, resp, http.StatusAccepted)
	resp.Body.Close()

	// Wait for execution record to be created
	time.Sleep(1 * time.Second)

	// Verify we have at least 1 execution
	resp = doRequest(t, "GET", srv.URL+"/api/v1/executions", "")
	assertStatus(t, resp, http.StatusOK)
	var execs []map[string]any
	parseJSON(t, resp, &execs)
	if len(execs) == 0 {
		t.Skip("no execution records created in test env")
	}

	// Purge without older_than should fail
	resp = doRequest(t, "DELETE", srv.URL+"/api/v1/executions", "")
	assertStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	// Purge with older_than=0d should delete everything (0 days ago = now)
	// But our records were just created, so "older than now" means everything
	// Use a future timestamp to delete everything
	futureTS := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
	resp = doRequest(t, "DELETE", srv.URL+"/api/v1/executions?older_than="+futureTS, "")
	assertStatus(t, resp, http.StatusOK)
	var purgeResult map[string]any
	parseJSON(t, resp, &purgeResult)
	deleted := purgeResult["deleted"].(float64)
	if deleted == 0 {
		t.Fatalf("expected at least 1 deleted execution, got 0")
	}

	// Verify executions are gone
	resp = doRequest(t, "GET", srv.URL+"/api/v1/executions", "")
	assertStatus(t, resp, http.StatusOK)
	parseJSON(t, resp, &execs)
	if len(execs) != 0 {
		t.Fatalf("expected 0 executions after purge, got %d", len(execs))
	}
}

func TestExecutionPurgeDurationFormat(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Test that duration formats are accepted (e.g., "30d", "24h")
	// Even with no records, the endpoint should return 200 with deleted: 0
	resp := doRequest(t, "DELETE", srv.URL+"/api/v1/executions?older_than=30d", "")
	assertStatus(t, resp, http.StatusOK)
	var result map[string]any
	parseJSON(t, resp, &result)
	if result["deleted"].(float64) != 0 {
		t.Fatalf("expected 0 deleted with no records, got %v", result["deleted"])
	}

	// Test with hours format
	resp = doRequest(t, "DELETE", srv.URL+"/api/v1/executions?older_than=24h", "")
	assertStatus(t, resp, http.StatusOK)
	parseJSON(t, resp, &result)

	// Test with invalid format
	resp = doRequest(t, "DELETE", srv.URL+"/api/v1/executions?older_than=garbage", "")
	assertStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()
}

func TestKanbanTaskExecutionLogging(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create an agent
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Logging Worker",
		"prompt": "Process tasks and log"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agent map[string]any
	parseJSON(t, resp, &agent)
	agentID := agent["id"].(string)

	// Create a kanban task
	resp = doRequest(t, "POST", srv.URL+"/api/v1/kanban-tasks", `{
		"title": "Logged kanban task",
		"description": "This should create an execution record",
		"agent_task_id": "`+agentID+`"
	}`)
	assertStatus(t, resp, http.StatusCreated)

	// Wait for the kanban executor to process (will fail since no claude CLI)
	time.Sleep(2 * time.Second)

	// Check that an execution record was created for this agent
	resp = doRequest(t, "GET", srv.URL+"/api/v1/agent-tasks/"+agentID+"/executions", "")
	assertStatus(t, resp, http.StatusOK)
	var executions []map[string]any
	parseJSON(t, resp, &executions)

	if len(executions) == 0 {
		t.Fatalf("expected at least 1 execution record from kanban task, got 0")
	}

	// Verify the execution has the kanban task name prefix
	exec := executions[0]
	taskName, ok := exec["task_name"].(string)
	if !ok {
		t.Fatalf("expected task_name to be a string, got %v", exec["task_name"])
	}
	if taskName != "[kanban] Logged kanban task" {
		t.Fatalf("expected task_name '[kanban] Logged kanban task', got %s", taskName)
	}

	// Execution should also appear in the global executions list
	resp = doRequest(t, "GET", srv.URL+"/api/v1/executions", "")
	assertStatus(t, resp, http.StatusOK)
	var allExecs []map[string]any
	parseJSON(t, resp, &allExecs)
	found := false
	for _, e := range allExecs {
		if tn, ok := e["task_name"].(string); ok && tn == "[kanban] Logged kanban task" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("kanban execution not found in global executions list")
	}
}

func TestTeamCRUD(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create
	resp := doRequest(t, "POST", srv.URL+"/api/v1/teams", `{
		"title": "Backend Team",
		"description": "Handles backend services",
		"color": "#ff5733"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var created map[string]any
	parseJSON(t, resp, &created)
	teamID := created["id"].(string)

	if created["title"] != "Backend Team" {
		t.Fatalf("expected title 'Backend Team', got %v", created["title"])
	}
	if created["description"] != "Handles backend services" {
		t.Fatalf("expected description, got %v", created["description"])
	}
	if created["color"] != "#ff5733" {
		t.Fatalf("expected color '#ff5733', got %v", created["color"])
	}

	// Get
	resp = doRequest(t, "GET", srv.URL+"/api/v1/teams/"+teamID, "")
	assertStatus(t, resp, http.StatusOK)
	var fetched map[string]any
	parseJSON(t, resp, &fetched)
	if fetched["id"] != teamID {
		t.Fatalf("expected id %s, got %v", teamID, fetched["id"])
	}

	// List
	resp = doRequest(t, "GET", srv.URL+"/api/v1/teams", "")
	assertStatus(t, resp, http.StatusOK)
	var list []map[string]any
	parseJSON(t, resp, &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 team, got %d", len(list))
	}

	// Update
	resp = doRequest(t, "PATCH", srv.URL+"/api/v1/teams/"+teamID, `{"title": "Platform Team", "color": "#00ff00"}`)
	assertStatus(t, resp, http.StatusOK)
	var updated map[string]any
	parseJSON(t, resp, &updated)
	if updated["title"] != "Platform Team" {
		t.Fatalf("expected title 'Platform Team', got %v", updated["title"])
	}
	if updated["color"] != "#00ff00" {
		t.Fatalf("expected color '#00ff00', got %v", updated["color"])
	}
	if updated["description"] != "Handles backend services" {
		t.Fatalf("expected description preserved, got %v", updated["description"])
	}

	// Delete
	resp = doRequest(t, "DELETE", srv.URL+"/api/v1/teams/"+teamID, "")
	assertStatus(t, resp, http.StatusNoContent)

	// Verify deleted
	resp = doRequest(t, "GET", srv.URL+"/api/v1/teams/"+teamID, "")
	assertStatus(t, resp, http.StatusNotFound)
}

func TestTeamDefaultColor(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create without color — should get default
	resp := doRequest(t, "POST", srv.URL+"/api/v1/teams", `{
		"title": "Default Color Team"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var created map[string]any
	parseJSON(t, resp, &created)
	if created["color"] != "#6366f1" {
		t.Fatalf("expected default color '#6366f1', got %v", created["color"])
	}
}

func TestAgentTeamAssignment(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create a team
	resp := doRequest(t, "POST", srv.URL+"/api/v1/teams", `{
		"title": "Data Team",
		"color": "#2dd4bf"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var team map[string]any
	parseJSON(t, resp, &team)
	teamID := team["id"].(string)

	// Create agent with team_id
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Data Agent",
		"prompt": "Process data",
		"team_id": "`+teamID+`"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agent map[string]any
	parseJSON(t, resp, &agent)
	agentID := agent["id"].(string)
	if agent["team_id"] != teamID {
		t.Fatalf("expected team_id %s on create, got %v", teamID, agent["team_id"])
	}

	// Create agent without team
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Solo Agent",
		"prompt": "Work alone"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var soloAgent map[string]any
	parseJSON(t, resp, &soloAgent)

	// Filter agents by team_id
	resp = doRequest(t, "GET", srv.URL+"/api/v1/agent-tasks?team_id="+teamID, "")
	assertStatus(t, resp, http.StatusOK)
	var filtered []map[string]any
	parseJSON(t, resp, &filtered)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 agent in team, got %d", len(filtered))
	}
	if filtered[0]["name"] != "Data Agent" {
		t.Fatalf("expected 'Data Agent', got %v", filtered[0]["name"])
	}

	// Update agent to remove team (set to null via PATCH)
	resp = doRequest(t, "PATCH", srv.URL+"/api/v1/agent-tasks/"+agentID, `{
		"team_id": null
	}`)
	assertStatus(t, resp, http.StatusOK)

	// Team filter should now return 0
	resp = doRequest(t, "GET", srv.URL+"/api/v1/agent-tasks?team_id="+teamID, "")
	assertStatus(t, resp, http.StatusOK)
	parseJSON(t, resp, &filtered)
	if len(filtered) != 0 {
		t.Fatalf("expected 0 agents in team after removal, got %d", len(filtered))
	}
}

func TestTeamDeleteUnassignsAgents(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create team
	resp := doRequest(t, "POST", srv.URL+"/api/v1/teams", `{
		"title": "Temp Team",
		"color": "#ff0000"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var tm map[string]any
	parseJSON(t, resp, &tm)
	teamID := tm["id"].(string)

	// Create agent in team
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Team Agent",
		"prompt": "Work in team",
		"team_id": "`+teamID+`"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agent map[string]any
	parseJSON(t, resp, &agent)
	agentID := agent["id"].(string)

	// Delete team
	resp = doRequest(t, "DELETE", srv.URL+"/api/v1/teams/"+teamID, "")
	assertStatus(t, resp, http.StatusNoContent)

	// Agent should still exist but team_id should be null (ON DELETE SET NULL)
	resp = doRequest(t, "GET", srv.URL+"/api/v1/agent-tasks/"+agentID, "")
	assertStatus(t, resp, http.StatusOK)
	var fetched map[string]any
	parseJSON(t, resp, &fetched)
	if val, ok := fetched["team_id"]; ok && val != nil {
		t.Fatalf("expected team_id to be null after team deletion, got %v", val)
	}
}

func TestKanbanTaskTeamFilter(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create team
	resp := doRequest(t, "POST", srv.URL+"/api/v1/teams", `{
		"title": "Kanban Team",
		"color": "#0000ff"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var tm map[string]any
	parseJSON(t, resp, &tm)
	teamID := tm["id"].(string)

	// Create agent in team
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Team Worker",
		"prompt": "Work",
		"team_id": "`+teamID+`"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agent map[string]any
	parseJSON(t, resp, &agent)
	agentID := agent["id"].(string)

	// Create agent NOT in team
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Solo Worker",
		"prompt": "Work alone"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var soloAgent map[string]any
	parseJSON(t, resp, &soloAgent)
	soloAgentID := soloAgent["id"].(string)

	// Create kanban task for team agent
	resp = doRequest(t, "POST", srv.URL+"/api/v1/kanban-tasks", `{
		"title": "Team task",
		"agent_task_id": "`+agentID+`"
	}`)
	assertStatus(t, resp, http.StatusCreated)

	// Create kanban task for solo agent
	resp = doRequest(t, "POST", srv.URL+"/api/v1/kanban-tasks", `{
		"title": "Solo task",
		"agent_task_id": "`+soloAgentID+`"
	}`)
	assertStatus(t, resp, http.StatusCreated)

	// Filter kanban tasks by team
	resp = doRequest(t, "GET", srv.URL+"/api/v1/kanban-tasks?team_id="+teamID, "")
	assertStatus(t, resp, http.StatusOK)
	var filtered []map[string]any
	parseJSON(t, resp, &filtered)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 kanban task in team, got %d", len(filtered))
	}
	if filtered[0]["title"] != "Team task" {
		t.Fatalf("expected 'Team task', got %v", filtered[0]["title"])
	}
}

// Helpers

func doRequest(t *testing.T, method, url, body string) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func TestOpenWorkspace(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create an agent (workspace dir is auto-created)
	resp := doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks", `{
		"name": "Workspace open test",
		"prompt": "Do stuff"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var agent map[string]any
	parseJSON(t, resp, &agent)
	agentID := agent["id"].(string)

	// Open workspace — should return 200 with path
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks/"+agentID+"/open-workspace", "")
	assertStatus(t, resp, http.StatusOK)
	var result map[string]string
	parseJSON(t, resp, &result)
	if result["path"] == "" {
		t.Fatal("expected path in response")
	}
	expectedSuffix := "/" + agentID
	if !bytes.HasSuffix([]byte(result["path"]), []byte(expectedSuffix)) {
		t.Fatalf("expected path to end with %s, got %s", expectedSuffix, result["path"])
	}

	// Non-existent agent — should return 404
	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks/00000000-0000-0000-0000-000000000000/open-workspace", "")
	assertStatus(t, resp, http.StatusNotFound)

	// Delete agent, then try open-workspace — should return 404 (workspace gone)
	resp = doRequest(t, "DELETE", srv.URL+"/api/v1/agent-tasks/"+agentID, "")
	assertStatus(t, resp, http.StatusNoContent)

	resp = doRequest(t, "POST", srv.URL+"/api/v1/agent-tasks/"+agentID+"/open-workspace", "")
	assertStatus(t, resp, http.StatusNotFound)
}

func TestSkillEnvironmentSecrets(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create skill with environment_secrets
	resp := doRequest(t, "POST", srv.URL+"/api/v1/skills", `{
		"title": "Linear API",
		"content": "Use the Linear GraphQL API. Your API key is in $LINEAR_API_KEY.",
		"environment_secrets": {"LINEAR_API_KEY": "lin_api_secret123", "LINEAR_WEBHOOK_SECRET": "whsec_abc"}
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var created map[string]any
	parseJSON(t, resp, &created)
	skillID := created["id"].(string)

	// Verify secret_keys are returned (sorted)
	keys := created["secret_keys"].([]any)
	if len(keys) != 2 {
		t.Fatalf("expected 2 secret_keys, got %d", len(keys))
	}
	if keys[0] != "LINEAR_API_KEY" || keys[1] != "LINEAR_WEBHOOK_SECRET" {
		t.Fatalf("unexpected secret_keys: %v", keys)
	}

	// Verify raw values are NOT in the response
	respBody := doRequest(t, "GET", srv.URL+"/api/v1/skills/"+skillID, "")
	assertStatus(t, respBody, http.StatusOK)
	body, _ := io.ReadAll(respBody.Body)
	respBody.Body.Close()
	bodyStr := string(body)
	if strings.Contains(bodyStr, "lin_api_secret123") {
		t.Fatal("secret value leaked in GET response")
	}
	if strings.Contains(bodyStr, "whsec_abc") {
		t.Fatal("secret value leaked in GET response")
	}
	// But secret_keys should be there
	if !strings.Contains(bodyStr, "LINEAR_API_KEY") {
		t.Fatal("expected secret key name in response")
	}

	// List skills — verify secrets not leaked
	resp = doRequest(t, "GET", srv.URL+"/api/v1/skills", "")
	assertStatus(t, resp, http.StatusOK)
	listBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if strings.Contains(string(listBody), "lin_api_secret123") {
		t.Fatal("secret value leaked in list response")
	}

	// PATCH to update secrets
	resp = doRequest(t, "PATCH", srv.URL+"/api/v1/skills/"+skillID, `{
		"environment_secrets": {"NEW_KEY": "new_val"}
	}`)
	assertStatus(t, resp, http.StatusOK)
	var updated map[string]any
	parseJSON(t, resp, &updated)
	updatedKeys := updated["secret_keys"].([]any)
	if len(updatedKeys) != 1 || updatedKeys[0] != "NEW_KEY" {
		t.Fatalf("expected [NEW_KEY], got %v", updatedKeys)
	}

	// PATCH without environment_secrets — should preserve existing
	resp = doRequest(t, "PATCH", srv.URL+"/api/v1/skills/"+skillID, `{
		"title": "Linear API v2"
	}`)
	assertStatus(t, resp, http.StatusOK)
	parseJSON(t, resp, &updated)
	preservedKeys := updated["secret_keys"].([]any)
	if len(preservedKeys) != 1 || preservedKeys[0] != "NEW_KEY" {
		t.Fatalf("expected secrets preserved, got %v", preservedKeys)
	}

	// PATCH with empty secrets — should clear
	resp = doRequest(t, "PATCH", srv.URL+"/api/v1/skills/"+skillID, `{
		"environment_secrets": {}
	}`)
	assertStatus(t, resp, http.StatusOK)
	parseJSON(t, resp, &updated)
	clearedKeys := updated["secret_keys"].([]any)
	if len(clearedKeys) != 0 {
		t.Fatalf("expected empty secret_keys after clear, got %v", clearedKeys)
	}
}

func TestSkillWithoutSecrets(t *testing.T) {
	srv, cleanup := SetupTestServer(t)
	defer cleanup()

	// Create skill without secrets — should work and return empty secret_keys
	resp := doRequest(t, "POST", srv.URL+"/api/v1/skills", `{
		"title": "Plain Skill",
		"content": "Just instructions, no secrets"
	}`)
	assertStatus(t, resp, http.StatusCreated)
	var created map[string]any
	parseJSON(t, resp, &created)
	keys := created["secret_keys"].([]any)
	if len(keys) != 0 {
		t.Fatalf("expected empty secret_keys, got %v", keys)
	}
}

func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d: %s", expected, resp.StatusCode, string(body))
	}
}

func parseJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
}
