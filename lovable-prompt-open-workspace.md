# OpenMaguro🐟 Dashboard — Open Workspace Button

Small UI addition: add a subtle workspace button to each agent row/card in the Agents list view.

The app talks to a REST API at `http://localhost:8080`.

---

## What to Add

On each **agent card/row in the Agents list**, add a small **📁** icon button:

- Place it inline with the existing action icons (edit, delete, etc.) — it should blend in, not stand out
- Just the 📁 emoji or a small folder icon, no text label. Use a tooltip on hover: "Open workspace"
- Keep it visually quiet — same opacity/color as other secondary action icons. It should not compete with the agent name or status
- On click: `POST /api/v1/agent-tasks/{id}/open-workspace`
  - On success (`200`): brief toast, e.g. "Opened workspace"
  - On error (`404`): toast "Workspace not found"

---

## API Reference

```
POST /api/v1/agent-tasks/{id}/open-workspace
```

No request body.

**Response `200`:**
```json
{"path": "/Users/you/.maguro/workspaces/550e8400-..."}
```

**Response `404`:**
```json
{"error": "workspace directory does not exist"}
```

The server opens Finder/Explorer on the local machine.
