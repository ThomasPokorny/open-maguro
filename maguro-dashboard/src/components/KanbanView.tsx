import { useState, useEffect, useRef } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  AgentTask,
  KanbanTask,
  listKanbanTasks,
  createKanbanTask,
  updateKanbanTask,
  deleteKanbanTask,
} from "@/lib/api";
import { relativeTime } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { toast } from "@/hooks/use-toast";
import {
  Plus,
  Trash2,
  ChevronDown,
  ChevronUp,
  X,
  Save,
  Link2,
} from "lucide-react";

// ── Helpers ──────────────────────────────────────────────────────────────────

// Same palette as swarm preset colors
const AGENT_PALETTE = [
  "#6366f1", "#2dd4bf", "#f59e0b", "#ef4444",
  "#8b5cf6", "#ec4899", "#10b981", "#3b82f6",
];

function agentColor(name: string): string {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = (hash * 31 + name.charCodeAt(i)) >>> 0;
  }
  return AGENT_PALETTE[hash % AGENT_PALETTE.length];
}

const STATUS_META = {
  todo: {
    label: "Todo",
    headerClass: "text-muted-foreground border-border",
    dotClass: "bg-muted-foreground",
    emptyMsg: "No tasks waiting",
  },
  progress: {
    label: "In Progress",
    headerClass: "text-primary border-primary/40",
    dotClass: "bg-primary animate-pulse",
    emptyMsg: "No tasks in progress",
  },
  done: {
    label: "Done",
    headerClass: "text-emerald-400 border-emerald-400/40",
    dotClass: "bg-emerald-400",
    emptyMsg: "No completed tasks",
  },
  failed: {
    label: "Failed",
    headerClass: "text-destructive border-destructive/40",
    dotClass: "bg-destructive",
    emptyMsg: "No failed tasks",
  },
} as const;

// ── Status Badge ─────────────────────────────────────────────────────────────

function StatusBadge({ status }: { status: KanbanTask["status"] }) {
  const colors: Record<KanbanTask["status"], string> = {
    todo: "bg-muted text-muted-foreground border-border",
    progress: "bg-primary/15 text-primary border-primary/30",
    done: "bg-emerald-400/15 text-emerald-400 border-emerald-400/30",
    failed: "bg-destructive/15 text-destructive border-destructive/30",
  };
  return (
    <span className={cn("inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-semibold", colors[status])}>
      {STATUS_META[status].label}
    </span>
  );
}

// ── Agent Badge ───────────────────────────────────────────────────────────────

function AgentBadge({ agentId, agents }: { agentId: string; agents: AgentTask[] }) {
  const agent = agents.find((a) => a.id === agentId);
  if (!agent) return null;
  const color = agentColor(agent.name);
  return (
    <span className="inline-flex items-center gap-1.5 text-[10px] bg-secondary text-foreground px-2 py-0.5 rounded-full border border-border max-w-[140px] truncate">
      <span className="w-1.5 h-1.5 rounded-full flex-shrink-0" style={{ backgroundColor: color }} />
      {agent.name}
    </span>
  );
}

// ── Kanban Card ───────────────────────────────────────────────────────────────

interface KanbanCardProps {
  task: KanbanTask;
  agents: AgentTask[];
  onDeleted: () => void;
  onUpdated: () => void;
}

function KanbanCard({ task, agents, onDeleted, onUpdated }: KanbanCardProps) {
  const queryClient = useQueryClient();
  const [expanded, setExpanded] = useState(false);
  const [editing, setEditing] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const agent = agents.find((a) => a.id === task.agent_task_id);
  const cardAccent = agent ? agentColor(agent.name) : undefined;
  const [editForm, setEditForm] = useState({ title: task.title, description: task.description });

  const isEditable = task.status === "todo";
  const timeLabel = task.status === "todo" ? relativeTime(task.created_at) : relativeTime(task.updated_at);

  const saveMutation = useMutation({
    mutationFn: () => updateKanbanTask(task.id, { title: editForm.title, description: editForm.description }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["kanban-tasks"] });
      setEditing(false);
      toast({ title: "Task updated" });
      onUpdated();
    },
    onError: (e: Error) => toast({ title: "Error", description: e.message, variant: "destructive" }),
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteKanbanTask(task.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["kanban-tasks"] });
      toast({ title: "Task deleted" });
      onDeleted();
    },
    onError: (e: Error) => toast({ title: "Error", description: e.message, variant: "destructive" }),
  });

  if (editing) {
    return (
      <div className="bg-card border border-primary/40 rounded-lg p-3 space-y-2 shadow-md">
        <Input
          value={editForm.title}
          onChange={(e) => setEditForm({ ...editForm, title: e.target.value })}
          className="bg-input border-border text-foreground text-sm h-8"
        />
        <Textarea
          value={editForm.description}
          onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
          rows={3}
          className="bg-input border-border text-foreground text-sm resize-y min-h-[60px]"
        />
        <div className="flex gap-2">
          <Button
            size="sm"
            onClick={() => saveMutation.mutate()}
            disabled={saveMutation.isPending || !editForm.title}
            className="h-7 text-xs bg-primary text-primary-foreground hover:bg-primary/90"
          >
            <Save className="w-3 h-3 mr-1" /> Save
          </Button>
          <Button size="sm" variant="ghost" onClick={() => setEditing(false)} className="h-7 text-xs text-muted-foreground">
            Cancel
          </Button>
        </div>
      </div>
    );
  }

  return (
    <>
      <div
        className={cn(
          "bg-card border border-border rounded-lg p-3 shadow-sm transition-all duration-200 overflow-hidden relative",
          isEditable && "hover:border-border/80 cursor-pointer",
          !isEditable && "cursor-default"
        )}
        style={cardAccent ? { borderLeftColor: cardAccent, borderLeftWidth: "3px" } : undefined}
        onClick={() => {
          if (isEditable) setEditing(true);
          else setExpanded((v) => !v);
        }}
      >
        {/* Title row */}
        <div className="flex items-start gap-2 mb-1.5">
          <p className="font-semibold text-sm text-foreground flex-1 leading-snug">{task.title}</p>
          <div className="flex items-center gap-1 flex-shrink-0 ml-1" onClick={(e) => e.stopPropagation()}>
            {(task.result || task.description) && (
              <button
                onClick={() => setExpanded((v) => !v)}
                className="text-muted-foreground hover:text-foreground p-0.5"
              >
                {expanded ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
              </button>
            )}
            <button
              onClick={() => setConfirmDelete(true)}
              className="text-muted-foreground hover:text-destructive p-0.5 transition-colors"
            >
              <Trash2 className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        {/* Description preview */}
        {task.description && !expanded && (
          <p className="text-xs text-muted-foreground line-clamp-2 mb-1.5">{task.description}</p>
        )}

        {/* Expanded content */}
        {expanded && (
          <div className="mt-1 mb-2 space-y-2">
            {task.description && (
              <p className="text-xs text-muted-foreground whitespace-pre-wrap">{task.description}</p>
            )}
            {task.result && (
              <div className="bg-muted/40 border border-border rounded p-2">
                <p className="text-[10px] font-semibold text-muted-foreground mb-1 uppercase tracking-wide">Result</p>
                <p className="text-xs text-foreground whitespace-pre-wrap">{task.result}</p>
              </div>
            )}
          </div>
        )}

        {/* Footer */}
        <div className="flex items-center gap-2 mt-1.5 flex-wrap">
          <AgentBadge agentId={task.agent_task_id} agents={agents} />
          <span className="text-[10px] text-muted-foreground ml-auto">{timeLabel}</span>
          {task.status === "failed" && <StatusBadge status="failed" />}
        </div>
      </div>

      <ConfirmDialog
        open={confirmDelete}
        onOpenChange={setConfirmDelete}
        title="Delete Task"
        description={`Delete "${task.title}"? This cannot be undone.`}
        confirmLabel="Delete"
        destructive
        onConfirm={() => deleteMutation.mutate()}
      />
    </>
  );
}

// ── Column ────────────────────────────────────────────────────────────────────

interface ColumnProps {
  status: "todo" | "progress" | "done";
  tasks: KanbanTask[];
  failedTasks: KanbanTask[];
  agents: AgentTask[];
  onDeleted: () => void;
  onUpdated: () => void;
  showCreate?: boolean;
  onCreateClick?: () => void;
}

function KanbanColumn({ status, tasks, failedTasks, agents, onDeleted, onUpdated, showCreate, onCreateClick }: ColumnProps) {
  const meta = STATUS_META[status];
  // Show failed in done column
  const displayTasks = status === "done" ? [...tasks, ...failedTasks] : tasks;
  const count = displayTasks.length;

  return (
    <div className="flex flex-col min-w-[280px] flex-1 max-w-sm">
      {/* Column header */}
      <div className={cn("flex items-center gap-2 px-1 mb-3 pb-2 border-b", meta.headerClass)}>
        <span className={cn("w-2 h-2 rounded-full flex-shrink-0", meta.dotClass)} />
        <span className="font-semibold text-sm">{meta.label}</span>
        <span className="text-xs bg-muted text-muted-foreground rounded-full px-1.5 py-0.5 ml-1">{count}</span>
        {showCreate && (
          <button
            onClick={onCreateClick}
            className="ml-auto text-muted-foreground hover:text-primary transition-colors"
          >
            <Plus className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Card list */}
      <div className="flex-1 space-y-2 overflow-y-auto pr-1" style={{ maxHeight: "calc(100vh - 200px)" }}>
        {displayTasks.length === 0 && (
          <p className="text-xs text-muted-foreground text-center py-8 opacity-60">{meta.emptyMsg}</p>
        )}
        {displayTasks.map((task) => (
          <KanbanCard
            key={task.id}
            task={task}
            agents={agents}
            onDeleted={onDeleted}
            onUpdated={onUpdated}
          />
        ))}
      </div>
    </div>
  );
}

// ── Create Task Modal ─────────────────────────────────────────────────────────

interface CreateTaskModalProps {
  agents: AgentTask[];
  onCreated: () => void;
  onClose: () => void;
}

function CreateTaskModal({ agents, onCreated, onClose }: CreateTaskModalProps) {
  const queryClient = useQueryClient();
  const [form, setForm] = useState({ title: "", description: "", agent_task_id: agents[0]?.id ?? "" });

  const mutation = useMutation({
    mutationFn: () => createKanbanTask({ title: form.title, description: form.description, agent_task_id: form.agent_task_id }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["kanban-tasks"] });
      toast({ title: "Task created" });
      onCreated();
      onClose();
    },
    onError: (e: Error) => toast({ title: "Error", description: e.message, variant: "destructive" }),
  });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div
        className="bg-card border border-border rounded-xl shadow-2xl w-full max-w-md mx-4 p-5 space-y-4"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between">
          <h3 className="font-semibold text-foreground">New Task</h3>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground">
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="space-y-1.5">
          <Label className="text-xs text-muted-foreground">Title *</Label>
          <Input
            value={form.title}
            onChange={(e) => setForm({ ...form, title: e.target.value })}
            placeholder="What needs to be done?"
            maxLength={255}
            className="bg-input border-border text-foreground"
            autoFocus
          />
        </div>

        <div className="space-y-1.5">
          <Label className="text-xs text-muted-foreground">Description</Label>
          <Textarea
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
            placeholder="Optional instructions for the agent…"
            rows={3}
            className="bg-input border-border text-foreground resize-y min-h-[80px]"
          />
        </div>

        <div className="space-y-1.5">
          <Label className="text-xs text-muted-foreground">Assign to Agent *</Label>
          <select
            value={form.agent_task_id}
            onChange={(e) => setForm({ ...form, agent_task_id: e.target.value })}
            className="bg-input border border-border text-foreground text-sm rounded-md px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <option value="">Select agent…</option>
            {agents.map((a) => (
              <option key={a.id} value={a.id}>{a.name}</option>
            ))}
          </select>
        </div>

        <div className="flex gap-2 pt-1 border-t border-border">
          <Button
            onClick={() => mutation.mutate()}
            disabled={mutation.isPending || !form.title || !form.agent_task_id}
            size="sm"
            className="bg-primary text-primary-foreground hover:bg-primary/90"
          >
            Create Task
          </Button>
          <Button onClick={onClose} size="sm" variant="ghost" className="text-muted-foreground">
            Cancel
          </Button>
        </div>
      </div>
    </div>
  );
}

// ── Board View ────────────────────────────────────────────────────────────────

interface KanbanViewProps {
  agents: AgentTask[];
  selectedTeamId?: string | null;
}

export function KanbanView({ agents, selectedTeamId }: KanbanViewProps) {
  const queryClient = useQueryClient();
  const [agentFilter, setAgentFilter] = useState<string>("");
  const [showCreate, setShowCreate] = useState(false);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const { data: tasks = [], refetch } = useQuery({
    queryKey: ["kanban-tasks", agentFilter, selectedTeamId],
    queryFn: () => {
      const params: { agent_id?: string; team_id?: string } = {};
      if (agentFilter) params.agent_id = agentFilter;
      if (selectedTeamId) params.team_id = selectedTeamId;
      return listKanbanTasks(Object.keys(params).length ? params : undefined);
    },
    staleTime: 4_000,
  });

  // Poll every 5 seconds
  useEffect(() => {
    pollRef.current = setInterval(() => {
      queryClient.invalidateQueries({ queryKey: ["kanban-tasks"] });
    }, 5_000);
    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, [queryClient]);

  const todoTasks = tasks.filter((t) => t.status === "todo");
  const progressTasks = tasks.filter((t) => t.status === "progress");
  const doneTasks = tasks.filter((t) => t.status === "done");
  const failedTasks = tasks.filter((t) => t.status === "failed");

  const handleRefresh = () => refetch();

  return (
    <div className="flex flex-col gap-4">
      {/* Board toolbar */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="flex items-center gap-2">
          <select
            value={agentFilter}
            onChange={(e) => setAgentFilter(e.target.value)}
            className="bg-input border border-border text-foreground text-sm rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <option value="">All Agents</option>
            {agents.map((a) => (
              <option key={a.id} value={a.id}>{a.name}</option>
            ))}
          </select>
        </div>

        <Button
          size="sm"
          onClick={() => setShowCreate(true)}
          className="ml-auto bg-primary text-primary-foreground hover:bg-primary/90"
        >
          <Plus className="w-3.5 h-3.5 mr-1" />
          New Task
        </Button>
      </div>

      {/* Board columns */}
      <div className="flex gap-5 overflow-x-auto pb-4">
        <KanbanColumn
          status="todo"
          tasks={todoTasks}
          failedTasks={[]}
          agents={agents}
          onDeleted={handleRefresh}
          onUpdated={handleRefresh}
          showCreate
          onCreateClick={() => setShowCreate(true)}
        />
        <KanbanColumn
          status="progress"
          tasks={progressTasks}
          failedTasks={[]}
          agents={agents}
          onDeleted={handleRefresh}
          onUpdated={handleRefresh}
        />
        <KanbanColumn
          status="done"
          tasks={doneTasks}
          failedTasks={failedTasks}
          agents={agents}
          onDeleted={handleRefresh}
          onUpdated={handleRefresh}
        />
      </div>

      {showCreate && (
        <CreateTaskModal
          agents={agents}
          onCreated={handleRefresh}
          onClose={() => setShowCreate(false)}
        />
      )}
    </div>
  );
}
