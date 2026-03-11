# OpenMaguro🐟 Dashboard — Lovable Prompt

Build a single-page dashboard for **OpenMaguro🐟**, a scheduled AI agent task orchestrator. The app talks to a REST API at `http://localhost:8080`. All data comes from this API — there is no local database or auth.

---

## Design & Style

Visually orient towards the **Open Claw** look and feel — clean, modern, slightly industrial "hacker like" with clear typography and generous spacing — but replace all reds/warm accents with **dark blue, navy, and calming cool-tone colors**. Think deep navy (`#1a2332`), slate blue (`#334155`), muted teal accents (`#2dd4bf`), and soft steel grays. The background should be dark (near-black navy), cards should be slightly lighter dark panels, and text should be crisp white/light gray.

- **Title:** "OpenMaguro🐟" displayed prominently in the header (include the fish emoji in the rendered title)
- Rounded corners on cards, subtle borders, no harsh shadows
- Monospace font for cron expressions, IDs, and code-like fields
- Smooth transitions for accordion expand/collapse

---

## Layout

### Header
- App title: **OpenMaguro🐟**
- View switcher: two tabs/buttons to toggle between **Agents** and **Skills** view
- Small collapsible **Execution Logs** section accessible from the header (e.g. a subtle "Logs" link or icon that opens a slide-out panel or small collapsible section at the bottom — not a main navigation item, keep it understated)

### Main Content

The main area renders either the **Agents** view or the **Skills** view depending on the active tab.

---

## Agents View

A vertical **accordion list** of all agent tasks. Each collapsed row shows:

| Element | Description |
|---|---|
| **Name** | Agent name (bold) |
| **Cron** | Cron expression in monospace, or "One-time" badge |
| **Enabled** | Toggle switch to enable/disable |
| **Last Run** | Relative timestamp of last execution (e.g. "3 min ago"), or "Never" |
| **▶️ Run** | Button to trigger immediate execution |

### Expanded Agent (Accordion)

When a row is expanded, show an **edit form** with all agent properties:

- `name` (text input)
- `cron_expression` (text input, monospace)
- `prompt` (textarea, multi-line)
- `enabled` (toggle)
- `mcp_config` (text input, optional)
- `allowed_tools` (text input, comma-separated, optional)
- `system_agent` (toggle)
- `global_skill_access` (toggle)

Below the form fields, show a **Skills Assignment** section:
- List of currently assigned skills (as removable chips/tags)
- A dropdown or autocomplete to attach additional skills from the full skill list
- If `global_skill_access` is on, show a note: "This agent has access to all skills" and hide the individual assignment UI

Action buttons at the bottom of the expanded section:
- **Save** — PATCH updates to the agent
- **Delete** — delete the agent (with confirmation)

### Create Agent

A **"+ New Agent"** button at the top of the list that opens a creation form (can be inline at the top or a modal). Fields: `name`, `cron_expression`, `prompt`. Optional: `enabled`, `mcp_config`, `allowed_tools`, `system_agent`, `global_skill_access`.

---

## Skills View

A simpler list/card layout for managing skills. Each skill card shows:
- **Title** (bold)
- **Content** preview (truncated to 2-3 lines)
- **Edit** / **Delete** buttons

### Expanded/Edit Skill
- `title` (text input)
- `content` (large textarea — this can be long markdown with API docs, credentials, etc.)
- **Save** / **Cancel** buttons

### Create Skill
A **"+ New Skill"** button that opens an inline form or modal with `title` and `content` fields.

---

## Execution Logs (Minor Section)

This is **not a primary view** — it should be a collapsible panel, slide-out drawer, or a small expandable section. Keep it understated.

- Shows a chronological list of all executions (most recent first)
- Each entry shows: `task_name`, `status` (with color-coded badges: green=success, red=failure, yellow=running, gray=pending), `started_at`, `finished_at`, duration
- Clicking an entry expands to show `summary` and `error` fields
- Optionally filter by status

---

## API Reference

**Base URL:** `http://localhost:8080`

All endpoints return JSON. Errors return `{"error": "message"}`.

### Agent Tasks

**List agents:**
```
GET /api/v1/agent-tasks
```
Response `200`: Array of agent task objects.

**Get agent:**
```
GET /api/v1/agent-tasks/{id}
```

**Create agent:**
```
POST /api/v1/agent-tasks
Content-Type: application/json

{
  "name": "Daily report",
  "cron_expression": "0 6 * * *",
  "prompt": "Generate the daily sales report",
  "enabled": true,
  "mcp_config": null,
  "allowed_tools": null,
  "system_agent": false,
  "global_skill_access": false
}
```
Response `201`: Created agent object.

**Update agent (partial):**
```
PATCH /api/v1/agent-tasks/{id}
Content-Type: application/json

{
  "name": "Updated name",
  "enabled": false
}
```
All fields are optional. Only provided fields are updated.

**Delete agent:**
```
DELETE /api/v1/agent-tasks/{id}
```
Response `204`: No content.

**Run agent immediately:**
```
POST /api/v1/agent-tasks/{id}/run
```
Response `202`: `{"status": "accepted"}`. Execution runs in background.

**Agent task response shape:**
```json
{
  "id": "uuid",
  "name": "string",
  "task_type": "cron",
  "cron_expression": "0 6 * * *",
  "prompt": "string",
  "run_at": null,
  "mcp_config": null,
  "allowed_tools": null,
  "enabled": true,
  "system_agent": false,
  "global_skill_access": false,
  "created_at": "2026-03-05T10:00:00Z",
  "updated_at": "2026-03-05T10:00:00Z"
}
```

### Skills

**List skills:**
```
GET /api/v1/skills
```
Response `200`: Array of skill objects.

**Get skill:**
```
GET /api/v1/skills/{id}
```

**Create skill:**
```
POST /api/v1/skills
Content-Type: application/json

{
  "title": "Slack API",
  "content": "Use the Slack Bot Token to send messages..."
}
```
Response `201`: Created skill object.

**Update skill (partial):**
```
PATCH /api/v1/skills/{id}
Content-Type: application/json

{
  "title": "Updated title"
}
```

**Delete skill:**
```
DELETE /api/v1/skills/{id}
```
Response `204`: No content.

**Skill response shape:**
```json
{
  "id": "uuid",
  "title": "string",
  "content": "string",
  "created_at": "2026-03-05T10:00:00Z",
  "updated_at": "2026-03-05T10:00:00Z"
}
```

### Agent ↔ Skill Associations

**List skills for an agent:**
```
GET /api/v1/agent-tasks/{id}/skills
```
Response `200`: Array of skill objects assigned to this agent.

**Attach skill to agent:**
```
POST /api/v1/agent-tasks/{id}/skills/{skillId}
```
Response `204`: No content. Idempotent.

**Detach skill from agent:**
```
DELETE /api/v1/agent-tasks/{id}/skills/{skillId}
```
Response `204`: No content.

### Executions

**List all executions:**
```
GET /api/v1/executions
```
Response `200`: Array of execution objects (most recent first). Includes orphaned entries from deleted one-shot tasks (`agent_task_id` will be null, `task_name` preserved).

**List executions for a specific agent:**
```
GET /api/v1/agent-tasks/{taskId}/executions
```

**Get single execution:**
```
GET /api/v1/executions/{id}
```

**Execution response shape:**
```json
{
  "id": "uuid",
  "agent_task_id": "uuid or null",
  "task_name": "string or null",
  "status": "pending|running|success|failure",
  "started_at": "2026-03-05T06:00:00Z",
  "finished_at": "2026-03-05T06:01:30Z",
  "summary": "string or null",
  "error": "string or null",
  "created_at": "2026-03-05T06:00:00Z"
}
```

---

## Interaction Details

- After creating/updating/deleting an agent or skill, **refetch the list** to keep UI in sync
- After clicking ▶️ Run, show a brief toast "Execution started" — do not wait for it to finish
- The **enabled toggle** on the agent list row should immediately PATCH `{"enabled": true/false}` without expanding the accordion
- For the "Last Run" column, fetch `GET /api/v1/agent-tasks/{id}/executions`, take the first entry's `started_at`, and display as relative time. Cache or lazy-load this per agent to avoid excessive requests on initial load
- Deletion of agents and skills should show a confirmation dialog before proceeding
- Use optimistic UI updates where sensible (toggles, deletes)
