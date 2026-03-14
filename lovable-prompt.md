# OpenMaguro🐟 Dashboard — Team Swarms Update

Add a **Team Swarms** sidebar and team-scoped context to the existing OpenMaguro🐟 dashboard. The dashboard already has Agents, Skills, Execution Logs, and Kanban Board views. This prompt adds team management and team-scoped filtering as a core navigation concept.

The app talks to a REST API at `http://localhost:8080`. All data comes from this API — there is no local database or auth.

---

## Design & Style

Continue the existing dark navy theme — deep navy (`#1a2332`), slate blue (`#334155`), muted teal accents (`#2dd4bf`), soft steel grays. Dark background, lighter dark panels for cards, crisp white/light gray text.

The sidebar should feel **minimal, slick, and lightweight** — inspired by Linear's sidebar. Thin, compact, not overwhelming. It sits on the left edge and acts as the primary navigation context switcher.

---

## Sidebar — Team Swarms

### Layout

Add a **slim left sidebar** (approximately 220px wide, collapsible to icon-only ~48px). The sidebar contains:

1. **App logo/name** at the top — "OpenMaguro🐟" or a small fish icon
2. **Team Swarms section** — a list of teams, each showing:
   - A small **colored circle/dot** using the team's `color` hex value
   - The team **title** (truncated if long)
3. **"All Agents"** entry at the top of the team list (default, no team filter) — use a neutral icon/dot
4. **"+ New Swarm"** button at the bottom of the team list — opens create team modal/inline form
5. **Main navigation links** below the team list: Agents, Board, Skills, Logs (move these from the top header into the sidebar)

### Team Selection Behavior

- **Default state:** "All Agents" is selected — no team filter applied. All agents and kanban tasks are shown across all views.
- **When a team is selected:** The team's colored dot gets a highlight/active indicator. All subsequent views (Agents, Board) are filtered by that team's ID:
  - Agents view: `GET /api/v1/agent-tasks?team_id={uuid}`
  - Board view: `GET /api/v1/kanban-tasks?team_id={uuid}`
- **Skills view is NOT filtered by team** — skills are global and always show all skills regardless of team selection.
- **Logs view is NOT filtered by team** — execution logs are global.
- The selected team should persist in URL state or local state so it survives page navigation between Agents/Board/Skills/Logs.

### Team List Item — Context Menu

Right-click (or click a small `...` icon) on a team in the sidebar to show:
- **Edit** — opens edit modal (title, description, color)
- **Delete** — confirmation dialog, then deletes. Explain that agents won't be deleted, just unassigned.

### Team Indicator on Agents

In the Agents view, each agent card/row should show:
- A small **colored dot** matching the agent's team color (if assigned to a team)
- No dot if the agent has no team (`team_id: null`)

### Agent Team Assignment

In the agent create/edit form, add a **"Team Swarm"** dropdown:
- Lists all teams (fetched from `GET /api/v1/teams`)
- "No Team" option (sends `team_id: null` on update, omits `team_id` on create)
- Shows the team's colored dot next to each option
- On create: include `team_id` in `POST /api/v1/agent-tasks` body
- On update: `PATCH /api/v1/agent-tasks/{id}` with `{"team_id": "uuid"}` or `{"team_id": null}` to unassign

---

## Create Team Swarm Modal

Triggered by the "+ New Swarm" button in the sidebar. A clean modal with:

- `title` (text input, required, max 255 chars)
- `description` (textarea, optional)
- `color` (color picker, defaults to `#6366f1` — show a few preset swatches: `#6366f1`, `#2dd4bf`, `#f59e0b`, `#ef4444`, `#8b5cf6`, `#ec4899`, `#10b981`, `#3b82f6` plus a custom hex input)

On submit: `POST /api/v1/teams`. The new team appears in the sidebar immediately.

## Edit Team Swarm Modal

Same form as create, pre-filled with existing values. Save via `PATCH /api/v1/teams/{id}`.

## Delete Team Swarm

Confirmation dialog: "Delete **{team title}**? Agents in this swarm will be unassigned but not deleted."

On confirm: `DELETE /api/v1/teams/{id}`. Remove from sidebar. If the deleted team was currently selected, switch to "All Agents".

---

## Navigation Changes

Move the main page links (Agents, Board, Skills, Logs) from the **top header** into the **sidebar**, below the team list. The top header can remain for the app title or be removed/simplified.

The sidebar navigation should look like:

```
🐟 OpenMaguro
─────────────────
ALL AGENTS          ← default (no team filter)
● Data Team         ← team with colored dot
● DevOps            ← team with colored dot
● Marketing         ← team with colored dot
+ New Swarm
─────────────────
Agents
Board
Skills
Logs
```

The active page link should have a subtle highlight. The active team should have a stronger highlight (background tint using the team's color at low opacity, or a left border accent).

---

## API Reference — Teams

**Base URL:** `http://localhost:8080`

All endpoints return JSON. Errors return `{"error": "message"}`.

### Create Team

```
POST /api/v1/teams
Content-Type: application/json

{
  "title": "Data Team",
  "description": "Agents that handle data processing",
  "color": "#6366f1"
}
```

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `title` | string | Yes | — | Team name (1–255 chars) |
| `description` | string | No | `""` | Team description |
| `color` | string | No | `#6366f1` | Hex color code (validated as `#RRGGBB`) |

**Response `201`:** Created team object.

**Response `422`:** Validation error.

### List Teams

```
GET /api/v1/teams
```

**Response `200`:** Array of team objects (ordered by created_at DESC).

### Get Team

```
GET /api/v1/teams/{id}
```

**Response `200`:** Team object.

**Response `404`:** Team not found.

### Update Team

```
PATCH /api/v1/teams/{id}
Content-Type: application/json

{
  "title": "Updated name",
  "color": "#ef4444"
}
```

All fields optional. Only provided fields are updated.

| Field | Type | Description |
|---|---|---|
| `title` | string | New title (1–255 chars) |
| `description` | string | New description |
| `color` | string | New hex color |

**Response `200`:** Updated team object.

**Response `404`:** Team not found.

### Delete Team

```
DELETE /api/v1/teams/{id}
```

Deleting a team does **not** delete its agents. Agents are unassigned (their `team_id` becomes `null`) via `ON DELETE SET NULL` in the database.

**Response `204`:** No content.

### Team Response Shape

```json
{
  "id": "uuid",
  "title": "Data Team",
  "description": "Agents that handle data processing",
  "color": "#6366f1",
  "created_at": "2026-03-14T10:00:00Z",
  "updated_at": "2026-03-14T10:00:00Z"
}
```

| Field | Type | Description |
|---|---|---|
| `id` | uuid | Unique team ID |
| `title` | string | Team name |
| `description` | string | Team description (may be empty `""`) |
| `color` | string | Hex color code (`#RRGGBB`) |
| `created_at` | ISO 8601 | Creation timestamp |
| `updated_at` | ISO 8601 | Last update timestamp |

---

## API Reference — Agent Tasks (updated for teams)

Agent tasks now include a `team_id` field.

**Agent task response shape (relevant fields):**
```json
{
  "id": "uuid",
  "name": "Daily report agent",
  "cron_expression": "0 6 * * *",
  "prompt": "Generate the daily report...",
  "enabled": true,
  "system_agent": false,
  "on_success_task_id": "uuid or null",
  "on_failure_task_id": "uuid or null",
  "team_id": "uuid or null",
  "created_at": "2026-03-14T10:00:00Z",
  "updated_at": "2026-03-14T10:00:00Z"
}
```

**List agents (with team filter):**
```
GET /api/v1/agent-tasks
GET /api/v1/agent-tasks?team_id={uuid}
```

**Create agent (with team assignment):**
```
POST /api/v1/agent-tasks
Content-Type: application/json

{
  "name": "Data Cruncher",
  "prompt": "Process daily data...",
  "cron_expression": "0 6 * * *",
  "team_id": "uuid-of-team"
}
```

**Update agent team assignment:**
```
PATCH /api/v1/agent-tasks/{id}
Content-Type: application/json

{"team_id": "uuid-of-team"}
```

**Remove agent from team (unassign):**
```
PATCH /api/v1/agent-tasks/{id}
Content-Type: application/json

{"team_id": null}
```

Sending `"team_id": null` explicitly sets the team to null (unassigns). Omitting `team_id` from the PATCH body leaves it unchanged.

---

## API Reference — Kanban Tasks (updated for teams)

Kanban tasks can now be filtered by team. The team filter works through the agent — it returns kanban tasks whose assigned agent belongs to the specified team.

```
GET /api/v1/kanban-tasks?team_id={uuid}
GET /api/v1/kanban-tasks?team_id={uuid}&status={status}
```

| Param | Type | Description |
|---|---|---|
| `team_id` | uuid | Filter kanban tasks by the assigned agent's team |
| `agent_id` | uuid | Filter by assigned agent (existing) |
| `status` | string | Filter by status: `todo`, `progress`, `done`, `failed` (existing) |

When a team is selected in the sidebar, the Board view should pass `?team_id={uuid}` to the kanban list endpoint. When "All Agents" is selected, omit the `team_id` param to get all tasks.

---

## API Reference — Open Agent Workspace

The API can open an agent's workspace directory in the system file explorer. Use this to add an "Open Workspace" button on agent cards/rows.

```
POST /api/v1/agent-tasks/{id}/open-workspace
```

**Response `200`:**
```json
{"path": "/Users/you/.maguro/workspaces/550e8400-..."}
```

**Response `404`:** Agent not found or workspace directory doesn't exist.

This endpoint opens Finder/Explorer on the server machine (works because the dashboard and API run on the same local machine). Add a small **folder icon button** on each agent card/row that calls this endpoint. Display the returned `path` in a tooltip or toast so the user knows where the directory is.

---

## Skills — Environment Secrets

Skills now support **encrypted environment secrets** — API keys and tokens that get securely injected as environment variables at agent runtime. The API never returns secret values, only key names.

### Skills UI Changes

#### Create/Edit Skill Form

Add an **"Environment Secrets"** section below the existing `content` textarea:

- A **key-value pair editor** — a list of rows, each with:
  - `Key` text input (the env var name, e.g. `LINEAR_API_KEY`) — uppercase, no spaces
  - `Value` password input (the secret value, e.g. `lin_api_...`) — masked by default with a toggle-visibility eye icon
  - A **remove row** button (trash/x icon)
- A **"+ Add Secret"** button to add a new key-value row
- The section should feel optional — collapsed or minimal when no secrets are set. Don't clutter the form for skills that are just instructions.

**On create:** Send the key-value pairs as `environment_secrets` in the POST body:
```json
{
  "title": "Linear API",
  "content": "Use the Linear GraphQL API...",
  "environment_secrets": {
    "LINEAR_API_KEY": "lin_api_...",
    "LINEAR_WEBHOOK_SECRET": "whsec_..."
  }
}
```

**On edit:** The API does NOT return secret values — only key names via `secret_keys`. So when editing:
- Show existing secret key names as read-only chips/badges (e.g. `LINEAR_API_KEY ●●●●`)
- To **update** a secret, user must re-enter the value (show an "Update" button or inline edit per key)
- To **remove** a secret, user clicks remove — on save, send the updated `environment_secrets` map (without the removed key)
- To **add** a new secret, use the same "+ Add Secret" row
- When saving, send `environment_secrets` with all current key-value pairs. Omit `environment_secrets` from the PATCH body entirely if the user didn't touch the secrets section (this preserves existing secrets).

**Important:** Sending `"environment_secrets": {}` clears all secrets. Only send it when the user explicitly removes all secrets.

#### Skill List/Card View

On each skill card/row, show a small **lock icon** or **key count badge** if the skill has secrets:
- e.g. "🔑 2" or a small lock icon with count
- Nothing shown if `secret_keys` is empty
- This helps users quickly identify which skills carry credentials

#### Skill Detail View

When viewing a skill's details, show the secret key names in a list:
- `LINEAR_API_KEY` ●●●●
- `LINEAR_WEBHOOK_SECRET` ●●●●

Never show values. The dots/bullets make it clear values exist but are hidden.

### Skills API Reference (updated)

**Create skill:**
```
POST /api/v1/skills
Content-Type: application/json

{
  "title": "Linear API",
  "content": "Use the Linear GraphQL API. Your API key is in $LINEAR_API_KEY.",
  "environment_secrets": {
    "LINEAR_API_KEY": "lin_api_..."
  }
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `title` | string | Yes | Skill name (1–255 chars) |
| `content` | string | Yes | Instructions (markdown) |
| `environment_secrets` | object | No | Key-value map of secrets (encrypted at rest) |

**Update skill:**
```
PATCH /api/v1/skills/{id}
Content-Type: application/json

{"environment_secrets": {"NEW_KEY": "new_value"}}
```

- Omitting `environment_secrets` preserves existing secrets
- Sending `{}` clears all secrets
- Sending a new map fully replaces existing secrets

**Skill response shape:**
```json
{
  "id": "uuid",
  "title": "Linear API",
  "content": "Use the Linear GraphQL API...",
  "secret_keys": ["LINEAR_API_KEY"],
  "created_at": "2026-03-14T10:00:00Z",
  "updated_at": "2026-03-14T10:00:00Z"
}
```

| Field | Type | Description |
|---|---|---|
| `secret_keys` | string[] | Names of stored secrets (values never returned) |

The `environment_secrets` field is **write-only** — it never appears in any API response. Only `secret_keys` (the key names) are returned.

---

## Interaction Details

- **Team list polling:** Fetch `GET /api/v1/teams` on app load and cache it. Refresh when teams are created/updated/deleted. No need for continuous polling — teams change rarely.
- **Team context state:** Store the selected team ID in React state (or URL query param). Pass it to all data-fetching hooks for agents and kanban tasks.
- **Sidebar collapse:** On narrow screens or via a toggle, collapse the sidebar to show only icons/dots (team colored dots + nav icons). Expand on hover or click.
- **Color picker:** Use a simple grid of preset color swatches plus a hex input field. Validate hex format client-side before submit.
- **Team deletion flow:** Show confirmation → DELETE → if deleted team was active, switch to "All Agents" → refresh agent list (some agents will now have `team_id: null`).
- **Empty team state:** When a team is selected but has no agents, show a friendly message: "No agents in this swarm yet. Assign agents from the Agents view."
- **Toast notifications:** Show toasts for: swarm created, swarm updated, swarm deleted, errors.

---

# Maguro Chat UI 🐟

Add a new "Chat" view to the OpenMaguro dashboard — a borderless chat interface for talking to the Maguro meta-agent.

## 1. API Client (`src/lib/api.ts`)

Add a new interface and function:

```ts
export interface ChatResponse {
  reply: string;
  session_id?: string;
}

export function sendChatMessage(message: string): Promise<ChatResponse> {
  return request<ChatResponse>("/chat", {
    method: "POST",
    body: JSON.stringify({ message }),
  });
}

export function resetChatSession(): Promise<void> {
  return request<void>("/chat/reset", { method: "POST" });
}
```

## 2. New Component: `src/components/MaguroChatView.tsx`

Create a chat interface component:

- **Layout**: Full-height flex column. No max-width container — the chat should stretch borderless from the sidebar edge to the right side of the screen. No header bar when in chat tab (the chat IS the view). Messages area takes all available space, input pinned to bottom.

- **Messages area** (`flex-1 overflow-y-auto`):
  - User messages: right-aligned, teal/primary background bubble, rounded corners
  - Maguro replies: left-aligned, subtle dark background (`bg-secondary/30`), rounded corners
  - Maguro replies should render with a **typewriter effect** — characters appear one by one with a ~15ms interval. Use a `useEffect` with `setInterval` that incrementally reveals the text. Once complete, store the full text so re-renders don't re-animate.
  - Support markdown in replies — use a simple approach: split by `\n` for line breaks, wrap `` `code` `` in `<code>` tags, and `**bold**` in `<strong>` tags. No need for a full markdown library.

- **Loading indicator**: When waiting for a response, show animated fish swimming from right to left:
  ```
  🐟  🐠  🐡
  ```
  Use CSS animation: three fish emojis spaced apart, each translating from right to left with staggered `animation-delay` (0s, 0.3s, 0.6s). Use `@keyframes swimLeft` with `translateX`. Loop infinitely. Small text size (~16px). Left-aligned like a Maguro message bubble.

- **Input area** (bottom, sticky):
  - Single-line input or auto-expanding textarea (max 4 rows)
  - Send button with a fish icon or arrow
  - `Enter` to send, `Shift+Enter` for newline
  - Disabled while waiting for response
  - Placeholder: `"Message Maguro 🐟..."`

- **Session reset**: Small subtle button in the top-right corner of the chat area — "New conversation" with a refresh icon. Calls `resetChatSession()` and clears the local messages array.

- **State management**:
  - `messages: Array<{role: 'user' | 'maguro', text: string}>` in local state
  - `isLoading: boolean`
  - Use `useMutation` from TanStack Query for the chat call
  - Auto-scroll to bottom on new messages

## 3. Sidebar (`src/components/AppSidebar.tsx`)

- Add `"chat"` to the `Tab` type
- Add a new nav item **above all others** in `NAV_ITEMS`:
  ```ts
  { tab: "chat", label: "Chat", icon: MessageCircle }
  ```
  Import `MessageCircle` from `lucide-react`.

## 4. Index page (`src/pages/Index.tsx`)

- Add `"chat"` to the `Tab` type
- Import and render `MaguroChatView` when `activeTab === "chat"`
- **Important**: When the chat tab is active, hide the header bar and remove the `max-w-5xl mx-auto px-6 py-6` container — the chat view should be completely borderless (flush with the sidebar). Render it directly inside the `flex-1` main area div, bypassing the `<main>` wrapper.
- Set `"chat"` as the **default active tab** (`useState<Tab>("chat")`)

## 5. Header title

Add to the header title conditionals:
```tsx
{activeTab === "chat" && "Maguro Chat 🐟"}
```
(Though the header is hidden for chat, keep this for completeness.)

## Design notes

- Match the existing dark theme (navy background `bg-background`, teal accent `text-primary`)
- Chat bubbles should feel lightweight, not boxy — small padding, rounded-2xl
- The fish loading animation should be playful but not distracting — small, subtle, smooth
- The typewriter effect should be fast enough to feel responsive (~15ms per character) but slow enough to notice
