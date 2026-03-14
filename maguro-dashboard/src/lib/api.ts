const BASE_URL = "http://localhost:8080/api/v1";

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { "Content-Type": "application/json", ...options?.headers },
    ...options,
  });
  if (res.status === 204) return undefined as T;
  const data = await res.json();
  if (!res.ok) throw new Error(data?.error ?? `HTTP ${res.status}`);
  return data as T;
}

// --- Types ---

export interface AgentTask {
  id: string;
  name: string;
  task_type: string;
  cron_expression: string | null;
  prompt: string;
  run_at: string | null;
  mcp_config: string | null;
  allowed_tools: string | null;
  enabled: boolean;
  system_agent: boolean;
  global_skill_access: boolean;
  on_success_task_id: string | null;
  on_failure_task_id: string | null;
  color: string | null;
  team_id: string | null;
  created_at: string;
  updated_at: string;
}

export interface KanbanTask {
  id: string;
  title: string;
  description: string;
  agent_task_id: string;
  status: "todo" | "progress" | "done" | "failed";
  result: string | null;
  created_at: string;
  updated_at: string;
}

export interface Skill {
  id: string;
  title: string;
  content: string;
  secret_keys: string[];
  created_at: string;
  updated_at: string;
}

export interface Execution {
  id: string;
  agent_task_id: string | null;
  task_name: string | null;
  status: "pending" | "running" | "success" | "failure";
  started_at: string;
  finished_at: string | null;
  summary: string | null;
  error: string | null;
  created_at: string;
}

export interface Team {
  id: string;
  title: string;
  description: string;
  color: string;
  created_at: string;
  updated_at: string;
}

// --- Agent Tasks ---

export const listAgents = (params?: { team_id?: string }) => {
  const query = new URLSearchParams();
  if (params?.team_id) query.set("team_id", params.team_id);
  const qs = query.toString();
  return request<AgentTask[]>(`/agent-tasks${qs ? `?${qs}` : ""}`);
};
export const getAgent = (id: string) => request<AgentTask>(`/agent-tasks/${id}`);
export const createAgent = (body: Partial<AgentTask>) =>
    request<AgentTask>("/agent-tasks", { method: "POST", body: JSON.stringify(body) });
export const updateAgent = (id: string, body: Partial<AgentTask>) =>
    request<AgentTask>(`/agent-tasks/${id}`, { method: "PATCH", body: JSON.stringify(body) });
export const deleteAgent = (id: string) =>
    request<void>(`/agent-tasks/${id}`, { method: "DELETE" });
export const runAgent = (id: string) =>
    request<{ status: string }>(`/agent-tasks/${id}/run`, { method: "POST" });

// --- Skills ---

export const listSkills = () => request<Skill[]>("/skills");
export const createSkill = (body: { title: string; content: string; environment_secrets?: Record<string, string> }) =>
    request<Skill>("/skills", { method: "POST", body: JSON.stringify(body) });
export const updateSkill = (id: string, body: Partial<Omit<Skill, "secret_keys">> & { environment_secrets?: Record<string, string> }) =>
    request<Skill>(`/skills/${id}`, { method: "PATCH", body: JSON.stringify(body) });
export const deleteSkill = (id: string) =>
    request<void>(`/skills/${id}`, { method: "DELETE" });

// --- Agent ↔ Skill Associations ---

export const listAgentSkills = (agentId: string) =>
    request<Skill[]>(`/agent-tasks/${agentId}/skills`);
export const attachSkill = (agentId: string, skillId: string) =>
    request<void>(`/agent-tasks/${agentId}/skills/${skillId}`, { method: "POST" });
export const detachSkill = (agentId: string, skillId: string) =>
    request<void>(`/agent-tasks/${agentId}/skills/${skillId}`, { method: "DELETE" });

// --- Executions ---

export const listExecutions = () => request<Execution[]>("/executions");
export const listAgentExecutions = (agentId: string) =>
    request<Execution[]>(`/agent-tasks/${agentId}/executions`);

// --- Kanban Tasks ---

export const listKanbanTasks = (params?: { agent_id?: string; status?: string; team_id?: string }) => {
  const query = new URLSearchParams();
  if (params?.agent_id) query.set("agent_id", params.agent_id);
  if (params?.status) query.set("status", params.status);
  if (params?.team_id) query.set("team_id", params.team_id);
  const qs = query.toString();
  return request<KanbanTask[]>(`/kanban-tasks${qs ? `?${qs}` : ""}`);
};
export const createKanbanTask = (body: { title: string; description?: string; agent_task_id: string }) =>
    request<KanbanTask>("/kanban-tasks", { method: "POST", body: JSON.stringify(body) });
export const updateKanbanTask = (id: string, body: { title?: string; description?: string; agent_task_id?: string }) =>
    request<KanbanTask>(`/kanban-tasks/${id}`, { method: "PATCH", body: JSON.stringify(body) });
export const deleteKanbanTask = (id: string) =>
    request<void>(`/kanban-tasks/${id}`, { method: "DELETE" });

// --- Teams ---

export const listTeams = () => request<Team[]>("/teams");
export const createTeam = (body: { title: string; description?: string; color?: string }) =>
    request<Team>("/teams", { method: "POST", body: JSON.stringify(body) });
export const updateTeam = (id: string, body: Partial<Pick<Team, "title" | "description" | "color">>) =>
    request<Team>(`/teams/${id}`, { method: "PATCH", body: JSON.stringify(body) });
export const deleteTeam = (id: string) =>
    request<void>(`/teams/${id}`, { method: "DELETE" });

// --- Workspace ---

export const openWorkspace = (id: string) =>
    request<{ path: string }>(`/agent-tasks/${id}/open-workspace`, { method: "POST" });

// --- Chat ---

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