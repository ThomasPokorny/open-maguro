# OpenMaguro🐟 Dashboard — Kanban Board Update

Add a new **Kanban Board** page to the existing OpenMaguro🐟 dashboard. The dashboard already has Agents, Skills, and Execution Logs views. This prompt adds a full kanban task board as a new top-level page.

The app talks to a REST API at `http://localhost:8080`. All data comes from this API — there is no local database or auth.

---

## Design & Style

Continue the existing dark navy theme — deep navy (`#1a2332`), slate blue (`#334155`), muted teal accents (`#2dd4bf`), soft steel grays. Dark background, lighter dark panels for cards, crisp white/light gray text.

---

## Navigation Update

Add **"Board"** as a new tab/page in the header navigation alongside the existing Agents, Skills, and Logs views. The Board should feel like a **primary feature** — give it equal prominence to Agents.

---

## Kanban Board Page

A **three-column kanban board** with columns: **Todo**, **In Progress**, **Done**.

### Board Layout

Three vertical columns side by side, each with a header and a scrollable card list:

| Column | Status filter | Header color accent |
|---|---|---|
| **Todo** | `todo` | Gray/neutral |
| **In Progress** | `progress` | Teal/blue pulse or glow |
| **Done** | `done` | Green/success |

There is also a **Failed** state — show failed tasks in the Done column with a red badge, or add an optional 4th "Failed" column. Keep it clean — failed tasks should be visible but not dominate.

### Agent Filter

At the top of the board, add an **agent filter dropdown**:
- "All Agents" (default — shows kanban tasks from all agents)
- Lists all agent tasks by name (fetched from `GET /api/v1/agent-tasks`)
- Selecting an agent filters the board to only that agent's tasks via `?agent_id={uuid}`

### Kanban Task Cards

Each card in a column shows:

| Element | Description |
|---|---|
| **Title** | Task title (bold, truncated if long) |
| **Description** | First 2 lines of description (truncated, lighter text) |
| **Agent** | Agent name badge/chip (small, colored) — resolve from agent_task_id |
| **Time** | Relative timestamp ("3 min ago") — use `created_at` for todo, `updated_at` for progress/done |
| **Result** | For done/failed tasks, show a small expandable section or tooltip with the `result` field |

Card click or expand should show the full description and full result text.

### Create Task

A **"+ New Task"** button at the top of the Todo column (or top of the board). Opens an inline form or modal:

- `title` (text input, required, max 255 chars)
- `description` (textarea, optional)
- `agent_task_id` (dropdown, required — select from existing agent tasks, show agent name)

On submit: `POST /api/v1/kanban-tasks`. The task appears in the Todo column. The agent picks it up automatically — it will move to In Progress within seconds.

### Real-time Feel

Poll `GET /api/v1/kanban-tasks` every **5 seconds** (or use a configurable interval) to refresh the board. When a card moves from one status to another, animate it sliding to the new column. This gives the illusion of real-time as agents pick up and complete tasks.

### Done Task Auto-Hide

The API already filters done tasks older than 2 hours from the default list. The UI doesn't need to handle this — just refetch and render what the API returns. Optionally add a "Show all done" toggle that fetches `?status=done` to see historical completed tasks.

### Card Actions

- **Delete**: small trash icon on each card. Confirmation dialog before `DELETE /api/v1/kanban-tasks/{id}`.
- **Edit**: click the card to open an edit view with `title` and `description` fields. Save via `PATCH /api/v1/kanban-tasks/{id}`. Only allow editing tasks in `todo` status (in-progress/done tasks are read-only).

---

## Agent Configuration — Chaining Fields

In the **Agents** view (already built), ensure the expanded agent edit form includes these chaining fields:

- `on_success_task_id` — dropdown selecting from existing agent tasks, or "None"
- `on_failure_task_id` — dropdown selecting from existing agent tasks, or "None"

When either is set, show a visual chain indicator on the collapsed agent row (e.g. a small "→ Agent Name" badge or chain link icon).

**Error handling for chaining:** The API returns `409 Conflict` with `{"error": "circular chain detected: task {id} would create a cycle"}` if the user tries to create a circular chain (A → B → A). Display this error clearly as a toast or inline error message near the dropdown.

---

## API Reference — Kanban Tasks

**Base URL:** `http://localhost:8080`

All endpoints return JSON. Errors return `{"error": "message"}`.

### Create Kanban Task

```
POST /api/v1/kanban-tasks
Content-Type: application/json

{
  "title": "Write Q1 report",
  "description": "Generate the quarterly sales report from KPI data",
  "agent_task_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `title` | string | Yes | Task title (1–255 chars) |
| `description` | string | No | Detailed task instructions (defaults to "") |
| `agent_task_id` | uuid | Yes | Agent to assign this task to |

**Response `201`:** Created kanban task object (status will be `"todo"`).

**Response `422`:** Validation error — missing required fields or title too long.
```json
{"error": "Key: 'CreateRequest.Title' Error:Field validation for 'Title' failed on the 'required' tag"}
```

**Response `500`:** Agent task ID doesn't exist (FK constraint).
```json
{"error": "failed to create kanban task"}
```

### List Kanban Tasks

```
GET /api/v1/kanban-tasks
GET /api/v1/kanban-tasks?agent_id={uuid}
GET /api/v1/kanban-tasks?status={status}
GET /api/v1/kanban-tasks?agent_id={uuid}&status={status}
```

| Param | Type | Description |
|---|---|---|
| `agent_id` | uuid | Filter by assigned agent |
| `status` | string | Filter by status: `todo`, `progress`, `done`, `failed` |

**Response `200`:** Array of kanban task objects.

**Response `400`:** Invalid `agent_id` format.
```json
{"error": "invalid agent_id"}
```

**Important:** The default list (no `?status=` filter) automatically hides done tasks older than 2 hours. Pass `?status=done` to see all done tasks regardless of age.

### Get Kanban Task

```
GET /api/v1/kanban-tasks/{id}
```

**Response `200`:** Kanban task object.

**Response `400`:** Invalid UUID format.
```json
{"error": "invalid id"}
```

**Response `404`:** Task not found.
```json
{"error": "kanban task not found"}
```

### Update Kanban Task

```
PATCH /api/v1/kanban-tasks/{id}
Content-Type: application/json

{
  "title": "Updated title",
  "description": "Updated description"
}
```

All fields optional. Only provided fields are updated.

| Field | Type | Description |
|---|---|---|
| `title` | string | New title (1–255 chars) |
| `description` | string | New description |
| `agent_task_id` | uuid | Reassign to a different agent |

**Response `200`:** Updated kanban task object.

**Response `404`:** Task not found.

**Note:** The UI should only allow editing tasks in `todo` status. Once an agent picks up a task (status `progress`/`done`/`failed`), edits are not meaningful.

### Delete Kanban Task

```
DELETE /api/v1/kanban-tasks/{id}
```

**Response `204`:** No content.

**Response `400`:** Invalid UUID.

### Kanban Task Response Shape

```json
{
  "id": "uuid",
  "title": "Write Q1 report",
  "description": "Generate the quarterly sales report",
  "agent_task_id": "uuid",
  "status": "todo",
  "result": null,
  "created_at": "2026-03-11T10:00:00Z",
  "updated_at": "2026-03-11T10:00:00Z"
}
```

| Field | Type | Description |
|---|---|---|
| `id` | uuid | Unique task ID |
| `title` | string | Task title |
| `description` | string | Task description (may be empty `""`) |
| `agent_task_id` | uuid | Assigned agent's ID |
| `status` | string | One of: `todo`, `progress`, `done`, `failed` |
| `result` | string or null | Agent's output (populated when done/failed) |
| `created_at` | ISO 8601 | Creation timestamp |
| `updated_at` | ISO 8601 | Last status change timestamp |

---

## API Reference — Agent Tasks (for dropdowns)

The Board page needs the agent list for the "assign to agent" dropdown and the agent filter.

**List agents:**
```
GET /api/v1/agent-tasks
```
**Response `200`:** Array of agent task objects. Use `id` and `name` fields for dropdowns.

**Agent task response shape (relevant fields):**
```json
{
  "id": "uuid",
  "name": "string",
  "cron_expression": "0 6 * * *",
  "enabled": true,
  "on_success_task_id": "uuid or null",
  "on_failure_task_id": "uuid or null"
}
```

**Update agent (for chaining config):**
```
PATCH /api/v1/agent-tasks/{id}
Content-Type: application/json

{"on_success_task_id": "uuid-of-next-agent"}
```

**Response `200`:** Updated agent object.

**Response `409`:** Circular chain detected.
```json
{"error": "circular chain detected: task 550e8400-... would create a cycle"}
```

---

## Interaction Details

- **Polling:** Fetch `GET /api/v1/kanban-tasks` (with current agent filter if set) every 5 seconds. Diff the results and animate cards that changed columns.
- **Optimistic UI:** When creating a task, immediately add it to the Todo column before the API responds. Remove it if the request fails.
- **Agent name resolution:** Fetch `GET /api/v1/agent-tasks` once on page load and cache it. Use the agent list to resolve `agent_task_id` → agent name for card badges and dropdowns.
- **Delete confirmation:** Always show a confirmation dialog before deleting a kanban task.
- **Toast notifications:** Show brief toasts for: task created, task deleted, errors.
- **Empty states:** Show friendly empty state messages in columns: "No tasks waiting", "No tasks in progress", "No completed tasks".
- **Responsive:** On narrow screens, stack columns vertically or allow horizontal scrolling.
