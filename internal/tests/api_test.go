package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
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
