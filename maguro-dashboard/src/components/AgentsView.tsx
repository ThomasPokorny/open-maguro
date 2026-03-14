import { useState, useEffect, useCallback, useRef } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  AgentTask,
  Skill,
  Team,
  createAgent,
  updateAgent,
  deleteAgent,
  runAgent,
  listAgentSkills,
  attachSkill,
  detachSkill,
  listAgentExecutions,
  listTeams,
  openWorkspace,
} from "@/lib/api";
import { relativeTime } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { toast } from "@/hooks/use-toast";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import {
  ChevronDown,
  ChevronRight,
  Folder,
  Play,
  Plus,
  Trash2,
  X,
  Save,
  Link2,
} from "lucide-react";

// --- Color palette (macOS-inspired) ---

const AGENT_COLORS: { label: string; value: string; bg: string; ring: string }[] = [
  { label: "Red",    value: "red",    bg: "bg-[hsl(0,72%,51%)]",   ring: "ring-[hsl(0,72%,51%)]" },
  { label: "Orange", value: "orange", bg: "bg-[hsl(25,95%,53%)]",  ring: "ring-[hsl(25,95%,53%)]" },
  { label: "Yellow", value: "yellow", bg: "bg-[hsl(48,96%,53%)]",  ring: "ring-[hsl(48,96%,53%)]" },
  { label: "Green",  value: "green",  bg: "bg-[hsl(142,71%,45%)]", ring: "ring-[hsl(142,71%,45%)]" },
  { label: "Blue",   value: "blue",   bg: "bg-[hsl(217,91%,60%)]", ring: "ring-[hsl(217,91%,60%)]" },
  { label: "Purple", value: "purple", bg: "bg-[hsl(271,81%,56%)]", ring: "ring-[hsl(271,81%,56%)]" },
  { label: "Gray",   value: "gray",   bg: "bg-[hsl(220,9%,46%)]",  ring: "ring-[hsl(220,9%,46%)]" },
];

function ColorPicker({ value, onChange }: { value: string | null | undefined; onChange: (c: string | null) => void }) {
  return (
    <div className="flex items-center gap-1.5 flex-wrap">
      <button
        type="button"
        title="No Color"
        onClick={() => onChange(null)}
        className={cn(
          "w-5 h-5 rounded-full border-2 border-border bg-background transition-all",
          !value && "ring-2 ring-offset-1 ring-ring ring-offset-background",
        )}
      />
      {AGENT_COLORS.map((c) => (
        <button
          key={c.value}
          type="button"
          title={c.label}
          onClick={() => onChange(c.value)}
          className={cn(
            "w-5 h-5 rounded-full transition-all",
            c.bg,
            value === c.value && `ring-2 ring-offset-1 ${c.ring} ring-offset-background`,
          )}
        />
      ))}
    </div>
  );
}

// --- Agent Row ---

interface AgentRowProps {
  agent: AgentTask;
  allAgents: AgentTask[];
  allSkills: Skill[];
  allTeams: Team[];
  isOpen: boolean;
  onToggle: () => void;
  onDeleted: () => void;
}

function AgentRow({ agent, allAgents, allSkills, allTeams, isOpen, onToggle, onDeleted }: AgentRowProps) {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<Partial<AgentTask>>({});
  const [confirmDelete, setConfirmDelete] = useState(false);

  // Last run via executions
  const { data: executions } = useQuery({
    queryKey: ["executions", agent.id],
    queryFn: () => listAgentExecutions(agent.id),
    staleTime: 30_000,
  });

  // Agent skills
  const { data: agentSkills = [], refetch: refetchSkills } = useQuery({
    queryKey: ["agent-skills", agent.id],
    queryFn: () => listAgentSkills(agent.id),
    enabled: isOpen,
  });

  // Only initialise the form when the row transitions from closed → open.
  const prevOpenRef = useRef(false);
  useEffect(() => {
    if (isOpen && !prevOpenRef.current) {
      setForm({
        name: agent.name,
        cron_expression: agent.cron_expression ?? "",
        prompt: agent.prompt,
        enabled: agent.enabled,
        mcp_config: agent.mcp_config ?? "",
        allowed_tools: agent.allowed_tools ?? "",
        system_agent: agent.system_agent,
        global_skill_access: agent.global_skill_access,
        on_success_task_id: agent.on_success_task_id ?? "",
        on_failure_task_id: agent.on_failure_task_id ?? "",
        team_id: agent.team_id ?? "",
      });
    }
    prevOpenRef.current = isOpen;
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen]);

  const toggleMutation = useMutation({
    mutationFn: (enabled: boolean) => updateAgent(agent.id, { enabled }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["agents"] }),
    onError: (e: Error) => toast({ title: "Error", description: e.message, variant: "destructive" }),
  });

  const toNullable = (v: unknown) => (v == null || v === "" ? null : (v as string));

  const saveMutation = useMutation({
    mutationFn: () => updateAgent(agent.id, {
      name: form.name,
      cron_expression: (form.cron_expression ?? "").trim() === "" ? null : form.cron_expression,
      prompt: form.prompt,
      enabled: form.enabled,
      mcp_config: toNullable(form.mcp_config),
      allowed_tools: toNullable(form.allowed_tools),
      system_agent: form.system_agent,
      global_skill_access: form.global_skill_access,
      on_success_task_id: toNullable(form.on_success_task_id),
      on_failure_task_id: toNullable(form.on_failure_task_id),
      team_id: toNullable(form.team_id as string | undefined),
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["agents"] });
      toast({ title: "Agent saved" });
    },
    onError: (e: Error) => {
      const msg = e.message;
      if (msg.toLowerCase().includes("circular chain")) {
        toast({ title: "Circular chain detected", description: msg, variant: "destructive" });
      } else {
        toast({ title: "Error", description: msg, variant: "destructive" });
      }
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteAgent(agent.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["agents"] });
      toast({ title: "Agent deleted" });
      onDeleted();
    },
    onError: (e: Error) => toast({ title: "Error", description: e.message, variant: "destructive" }),
  });

  const runMutation = useMutation({
    mutationFn: () => runAgent(agent.id),
    onSuccess: () => toast({ title: "▶️ Execution started", description: agent.name }),
    onError: (e: Error) => toast({ title: "Error", description: e.message, variant: "destructive" }),
  });

  const workspaceMutation = useMutation({
    mutationFn: () => openWorkspace(agent.id),
    onSuccess: () => toast({ title: "Opened workspace" }),
    onError: (e: Error) => {
      if (e.message.includes("404") || e.message.includes("does not exist")) {
        toast({ title: "Workspace not found", variant: "destructive" });
      } else {
        toast({ title: "Error", description: e.message, variant: "destructive" });
      }
    },
  });

  const attachMutation = useMutation({
    mutationFn: (skillId: string) => attachSkill(agent.id, skillId),
    onSuccess: () => refetchSkills(),
  });

  const detachMutation = useMutation({
    mutationFn: (skillId: string) => detachSkill(agent.id, skillId),
    onSuccess: () => refetchSkills(),
  });

  const lastRun = executions?.[0]?.started_at;
  const assignedSkillIds = new Set(agentSkills.map((s) => s.id));
  const unassignedSkills = allSkills.filter((s) => !assignedSkillIds.has(s.id));
  const agentColor = AGENT_COLORS.find(c => c.value === agent.color);
  const agentTeam = allTeams.find((t) => t.id === agent.team_id);

  return (
    <>
      <div className="border border-border rounded-lg overflow-hidden mb-2 bg-card">
        {/* Collapsed row */}
        <div className="flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-muted/30 transition-colors select-none">
          <button
            onClick={onToggle}
            className="text-muted-foreground hover:text-foreground transition-colors flex-shrink-0"
          >
            {isOpen ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
          </button>

          <div className="flex-1 flex items-center gap-3 min-w-0" onClick={onToggle}>
            {agentColor && (
              <span className={cn("w-2.5 h-2.5 rounded-full flex-shrink-0", agentColor.bg)} />
            )}
            {agentTeam && !agentColor && (
              <span
                className="w-2.5 h-2.5 rounded-full flex-shrink-0"
                style={{ backgroundColor: agentTeam.color }}
                title={agentTeam.title}
              />
            )}
            <span className="font-semibold text-foreground truncate">{agent.name}</span>
            {agent.cron_expression && (
              <span className="font-mono text-xs bg-muted text-muted-foreground px-2 py-0.5 rounded hidden sm:block">
                {agent.cron_expression}
              </span>
            )}
            {(agent.on_success_task_id || agent.on_failure_task_id) && (
              <span className="hidden sm:inline-flex items-center gap-1 text-[10px] text-primary bg-primary/10 border border-primary/20 rounded px-1.5 py-0.5">
                <Link2 className="w-3 h-3" />
                {[
                  agent.on_success_task_id && allAgents.find(a => a.id === agent.on_success_task_id)?.name,
                  agent.on_failure_task_id && allAgents.find(a => a.id === agent.on_failure_task_id)?.name,
                ].filter(Boolean).join(" / ")}
              </span>
            )}
          </div>

          <div className="flex items-center gap-3 flex-shrink-0">
            <span className="text-xs text-muted-foreground hidden md:block">
              {relativeTime(lastRun)}
            </span>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={(e) => { e.stopPropagation(); workspaceMutation.mutate(); }}
                    disabled={workspaceMutation.isPending}
                    className="text-muted-foreground hover:text-foreground hover:bg-muted h-7 px-2"
                  >
                    <Folder className="w-3.5 h-3.5" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Open workspace</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
            <Switch
              checked={agent.enabled}
              onCheckedChange={(v) => toggleMutation.mutate(v)}
              onClick={(e) => e.stopPropagation()}
              className="data-[state=checked]:bg-primary"
            />
            <Button
              size="sm"
              variant="ghost"
              onClick={(e) => { e.stopPropagation(); runMutation.mutate(); }}
              disabled={runMutation.isPending}
              className="text-primary hover:text-primary hover:bg-primary/10 h-7 px-2"
            >
              <Play className="w-3.5 h-3.5" />
            </Button>
          </div>
        </div>

        {/* Expanded edit form */}
        {isOpen && (
          <div className="border-t border-border px-4 py-4 space-y-4 bg-muted/10">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label className="text-muted-foreground text-xs">Name</Label>
                <Input
                  value={form.name ?? ""}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  className="bg-input border-border text-foreground"
                />
              </div>
              <div className="space-y-1.5">
                <Label className="text-muted-foreground text-xs">Cron Expression</Label>
                <Input
                  value={form.cron_expression ?? ""}
                  onChange={(e) => setForm({ ...form, cron_expression: e.target.value })}
                  placeholder="0 6 * * *"
                  className="bg-input border-border text-foreground font-mono text-sm"
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <Label className="text-muted-foreground text-xs">Team Swarm</Label>
              <select
                value={(form.team_id as string) ?? ""}
                onChange={(e) => setForm({ ...form, team_id: e.target.value || null })}
                className="bg-input border border-border text-foreground text-sm rounded-md px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="">No Team</option>
                {allTeams.map((t) => (
                  <option key={t.id} value={t.id}>{t.title}</option>
                ))}
              </select>
            </div>

            <div className="space-y-1.5">
              <Label className="text-muted-foreground text-xs">Prompt</Label>
              <Textarea
                value={form.prompt ?? ""}
                onChange={(e) => setForm({ ...form, prompt: e.target.value })}
                rows={6}
                className="bg-input border-border text-foreground resize-y min-h-[120px]"
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label className="text-muted-foreground text-xs">MCP Config</Label>
                <Input
                  value={form.mcp_config ?? ""}
                  onChange={(e) => setForm({ ...form, mcp_config: e.target.value })}
                  placeholder="Optional"
                  className="bg-input border-border text-foreground font-mono text-sm"
                />
              </div>
              <div className="space-y-1.5">
                <Label className="text-muted-foreground text-xs">Allowed Tools (comma-separated)</Label>
                <Input
                  value={form.allowed_tools ?? ""}
                  onChange={(e) => setForm({ ...form, allowed_tools: e.target.value })}
                  placeholder="tool1, tool2"
                  className="bg-input border-border text-foreground font-mono text-sm"
                />
              </div>
            </div>

            {/* Chaining fields */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label className="text-muted-foreground text-xs flex items-center gap-1">
                  <Link2 className="w-3 h-3" /> On Success → Chain to
                </Label>
                <select
                  value={(form.on_success_task_id as string) ?? ""}
                  onChange={(e) => setForm({ ...form, on_success_task_id: e.target.value || null })}
                  className="bg-input border border-border text-foreground text-sm rounded-md px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-ring"
                >
                  <option value="">None</option>
                  {allAgents.filter(a => a.id !== agent.id).map((a) => (
                    <option key={a.id} value={a.id}>{a.name}</option>
                  ))}
                </select>
              </div>
              <div className="space-y-1.5">
                <Label className="text-muted-foreground text-xs flex items-center gap-1">
                  <Link2 className="w-3 h-3 text-destructive" /> On Failure → Chain to
                </Label>
                <select
                  value={(form.on_failure_task_id as string) ?? ""}
                  onChange={(e) => setForm({ ...form, on_failure_task_id: e.target.value || null })}
                  className="bg-input border border-border text-foreground text-sm rounded-md px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-ring"
                >
                  <option value="">None</option>
                  {allAgents.filter(a => a.id !== agent.id).map((a) => (
                    <option key={a.id} value={a.id}>{a.name}</option>
                  ))}
                </select>
              </div>
            </div>

            <div className="flex flex-wrap gap-6">
              {[
                { key: "enabled", label: "Enabled" },
                { key: "system_agent", label: "System Agent" },
                { key: "global_skill_access", label: "Global Skill Access" },
              ].map(({ key, label }) => (
                <div key={key} className="flex items-center gap-2">
                   <Switch
                    id={`${agent.id}-${key}`}
                    checked={(form as unknown as Record<string, boolean>)[key] ?? false}
                    onCheckedChange={(v) => setForm({ ...form, [key]: v })}
                    className="data-[state=checked]:bg-primary"
                  />
                  <Label htmlFor={`${agent.id}-${key}`} className="text-sm text-muted-foreground cursor-pointer">
                    {label}
                  </Label>
                </div>
              ))}
            </div>

            {/* Skills Assignment */}
            <div className="space-y-2">
              <Label className="text-muted-foreground text-xs">Skills</Label>
              {form.global_skill_access ? (
                <p className="text-xs text-primary bg-primary/10 border border-primary/20 rounded px-3 py-2">
                  ✦ This agent has access to all skills via Global Skill Access
                </p>
              ) : (
                <div className="space-y-2">
                  <div className="flex flex-wrap gap-2">
                    {agentSkills.map((s) => (
                      <span
                        key={s.id}
                        className="inline-flex items-center gap-1 text-xs bg-secondary text-foreground px-2 py-1 rounded-full border border-border"
                      >
                        {s.title}
                        <button
                          onClick={() => detachMutation.mutate(s.id)}
                          className="text-muted-foreground hover:text-destructive ml-1"
                        >
                          <X className="w-3 h-3" />
                        </button>
                      </span>
                    ))}
                    {agentSkills.length === 0 && (
                      <span className="text-xs text-muted-foreground">No skills assigned</span>
                    )}
                  </div>
                  {unassignedSkills.length > 0 && (
                    <select
                      className="bg-input border border-border text-foreground text-sm rounded px-2 py-1.5 w-full md:w-64"
                      value=""
                      onChange={(e) => { if (e.target.value) attachMutation.mutate(e.target.value); }}
                    >
                      <option value="">+ Add skill…</option>
                      {unassignedSkills.map((s) => (
                        <option key={s.id} value={s.id}>{s.title}</option>
                      ))}
                    </select>
                  )}
                </div>
              )}
            </div>

            {/* Actions */}
            <div className="flex items-center gap-2 pt-2 border-t border-border">
              <Button
                onClick={() => saveMutation.mutate()}
                disabled={saveMutation.isPending}
                size="sm"
                className="bg-primary text-primary-foreground hover:bg-primary/90"
              >
                <Save className="w-3.5 h-3.5 mr-1" />
                Save
              </Button>
              <Button
                onClick={() => setConfirmDelete(true)}
                size="sm"
                variant="ghost"
                className="text-destructive hover:text-destructive hover:bg-destructive/10 ml-auto"
              >
                <Trash2 className="w-3.5 h-3.5 mr-1" />
                Delete
              </Button>
            </div>
          </div>
        )}
      </div>

      <ConfirmDialog
        open={confirmDelete}
        onOpenChange={setConfirmDelete}
        title="Delete Agent"
        description={`Are you sure you want to delete "${agent.name}"? This action cannot be undone.`}
        confirmLabel="Delete"
        destructive
        onConfirm={() => deleteMutation.mutate()}
      />
    </>
  );
}

// --- Create Agent Form ---

interface CreateAgentFormProps {
  onCreated: () => void;
  onCancel: () => void;
  allTeams: Team[];
}

function CreateAgentForm({ onCreated, onCancel, allTeams }: CreateAgentFormProps) {
  const [form, setForm] = useState({
    name: "",
    cron_expression: "",
    prompt: "",
    enabled: true,
    mcp_config: "",
    allowed_tools: "",
    system_agent: false,
    global_skill_access: false,
    color: null as string | null,
    team_id: "" as string,
  });

  const mutation = useMutation({
    mutationFn: () =>
      createAgent({
        name: form.name,
        cron_expression: form.cron_expression || null,
        prompt: form.prompt,
        enabled: form.enabled,
        mcp_config: form.mcp_config || null,
        allowed_tools: form.allowed_tools || null,
        system_agent: form.system_agent,
        global_skill_access: form.global_skill_access,
        color: form.color || null,
        team_id: form.team_id || null,
      }),
    onSuccess: () => {
      toast({ title: "Agent created" });
      onCreated();
    },
    onError: (e: Error) => toast({ title: "Error", description: e.message, variant: "destructive" }),
  });

  return (
    <div className="border border-primary/30 rounded-lg bg-card mb-4 p-4 space-y-4">
      <h3 className="text-sm font-semibold text-primary">New Agent</h3>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <Label className="text-muted-foreground text-xs">Name *</Label>
          <Input
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            placeholder="My agent"
            className="bg-input border-border text-foreground"
          />
        </div>
        <div className="space-y-1.5">
          <Label className="text-muted-foreground text-xs">Cron Expression</Label>
          <Input
            value={form.cron_expression}
            onChange={(e) => setForm({ ...form, cron_expression: e.target.value })}
            placeholder="0 6 * * *"
            className="bg-input border-border text-foreground font-mono text-sm"
          />
        </div>
      </div>

      <div className="space-y-1.5">
        <Label className="text-muted-foreground text-xs">Color</Label>
        <ColorPicker value={form.color} onChange={(c) => setForm({ ...form, color: c })} />
      </div>

      <div className="space-y-1.5">
        <Label className="text-muted-foreground text-xs">Team Swarm</Label>
        <select
          value={form.team_id}
          onChange={(e) => setForm({ ...form, team_id: e.target.value })}
          className="bg-input border border-border text-foreground text-sm rounded-md px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-ring"
        >
          <option value="">No Team</option>
          {allTeams.map((t) => (
            <option key={t.id} value={t.id}>{t.title}</option>
          ))}
        </select>
      </div>

      <div className="space-y-1.5">
        <Label className="text-muted-foreground text-xs">Prompt *</Label>
        <Textarea
          value={form.prompt}
          onChange={(e) => setForm({ ...form, prompt: e.target.value })}
          rows={5}
          placeholder="Describe what this agent should do…"
          className="bg-input border-border text-foreground resize-y min-h-[100px]"
        />
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <Label className="text-muted-foreground text-xs">MCP Config</Label>
          <Input
            value={form.mcp_config}
            onChange={(e) => setForm({ ...form, mcp_config: e.target.value })}
            placeholder="Optional"
            className="bg-input border-border text-foreground font-mono text-sm"
          />
        </div>
        <div className="space-y-1.5">
          <Label className="text-muted-foreground text-xs">Allowed Tools</Label>
          <Input
            value={form.allowed_tools}
            onChange={(e) => setForm({ ...form, allowed_tools: e.target.value })}
            placeholder="tool1, tool2"
            className="bg-input border-border text-foreground font-mono text-sm"
          />
        </div>
      </div>
      <div className="flex flex-wrap gap-6">
        {[
          { key: "enabled", label: "Enabled" },
          { key: "system_agent", label: "System Agent" },
          { key: "global_skill_access", label: "Global Skill Access" },
        ].map(({ key, label }) => (
          <div key={key} className="flex items-center gap-2">
              <Switch
              id={`new-${key}`}
              checked={(form as unknown as Record<string, boolean>)[key] ?? false}
              onCheckedChange={(v) => setForm({ ...form, [key]: v })}
              className="data-[state=checked]:bg-primary"
            />
            <Label htmlFor={`new-${key}`} className="text-sm text-muted-foreground cursor-pointer">
              {label}
            </Label>
          </div>
        ))}
      </div>
      <div className="flex gap-2 pt-2 border-t border-border">
        <Button
          onClick={() => mutation.mutate()}
          disabled={mutation.isPending || !form.name || !form.prompt}
          size="sm"
          className="bg-primary text-primary-foreground hover:bg-primary/90"
        >
          Create Agent
        </Button>
        <Button onClick={onCancel} size="sm" variant="ghost" className="text-muted-foreground">
          Cancel
        </Button>
      </div>
    </div>
  );
}

// --- Agents View ---

interface AgentsViewProps {
  agents: AgentTask[];
  allSkills: Skill[];
  isLoading: boolean;
  onRefetch: () => void;
}

export function AgentsView({ agents, allSkills, isLoading, onRefetch }: AgentsViewProps) {
  const [openId, setOpenId] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);

  const { data: allTeams = [] } = useQuery({
    queryKey: ["teams"],
    queryFn: listTeams,
    staleTime: 60_000,
  });

  const handleToggle = useCallback((id: string) => {
    setOpenId((prev) => (prev === id ? null : id));
  }, []);

  if (isLoading) {
    return (
      <div className="space-y-2">
        {[...Array(3)].map((_, i) => (
          <div key={i} className="h-14 rounded-lg bg-card border border-border animate-pulse" />
        ))}
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <p className="text-sm text-muted-foreground">{agents.length} agent{agents.length !== 1 ? "s" : ""}</p>
        <Button
          size="sm"
          onClick={() => setShowCreate(true)}
          className="bg-primary text-primary-foreground hover:bg-primary/90"
        >
          <Plus className="w-3.5 h-3.5 mr-1" />
          New Agent
        </Button>
      </div>

      {showCreate && (
        <CreateAgentForm
          onCreated={() => { setShowCreate(false); onRefetch(); }}
          onCancel={() => setShowCreate(false)}
          allTeams={allTeams}
        />
      )}

      {agents.length === 0 && !showCreate && (
        <div className="text-center py-16 text-muted-foreground">
          <p className="text-lg mb-1">No agents yet</p>
          <p className="text-sm">Create your first agent to get started</p>
        </div>
      )}

      {agents.map((agent) => (
        <AgentRow
          key={agent.id}
          agent={agent}
          allAgents={agents}
          allSkills={allSkills}
          allTeams={allTeams}
          isOpen={openId === agent.id}
          onToggle={() => handleToggle(agent.id)}
          onDeleted={onRefetch}
        />
      ))}
    </div>
  );
}
