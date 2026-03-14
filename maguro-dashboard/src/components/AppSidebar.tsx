import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Team,
  listTeams,
  createTeam,
  updateTeam,
  deleteTeam,
} from "@/lib/api";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { toast } from "@/hooks/use-toast";
import {
  ChevronLeft,
  ChevronRight,
  Plus,
  MoreHorizontal,
  Cpu,
  LayoutGrid,
  Users,
  BookOpen,
  ScrollText,
  MessageCircle,
  X,
  Pencil,
  Trash2,
} from "lucide-react";

// ── Swarm Color Swatches ─────────────────────────────────────────────────────

const SWARM_SWATCHES = [
  "#6366f1", "#2dd4bf", "#f59e0b", "#ef4444",
  "#8b5cf6", "#ec4899", "#10b981", "#3b82f6",
];

function SwarmColorPicker({
                            value,
                            onChange,
                          }: {
  value: string;
  onChange: (c: string) => void;
}) {
  const [customHex, setCustomHex] = useState(
      SWARM_SWATCHES.includes(value) ? "" : value
  );

  const isValidHex = (h: string) => /^#[0-9A-Fa-f]{6}$/.test(h);

  return (
      <div className="space-y-2">
        <div className="flex flex-wrap gap-2">
          {SWARM_SWATCHES.map((swatch) => (
              <button
                  key={swatch}
                  type="button"
                  onClick={() => { onChange(swatch); setCustomHex(""); }}
                  className={cn(
                      "w-6 h-6 rounded-full border-2 transition-all",
                      value === swatch ? "border-foreground scale-110" : "border-transparent"
                  )}
                  style={{ backgroundColor: swatch }}
                  title={swatch}
              />
          ))}
        </div>
        <div className="flex items-center gap-2">
          <div
              className="w-6 h-6 rounded-full border border-border flex-shrink-0"
              style={{ backgroundColor: isValidHex(customHex) ? customHex : "transparent" }}
          />
          <Input
              value={customHex}
              onChange={(e) => {
                setCustomHex(e.target.value);
                if (isValidHex(e.target.value)) onChange(e.target.value);
              }}
              placeholder="#RRGGBB"
              className="bg-input border-border text-foreground font-mono text-xs h-7 w-28"
              maxLength={7}
          />
        </div>
      </div>
  );
}

// ── Team Form Modal ───────────────────────────────────────────────────────────

interface TeamFormModalProps {
  team?: Team;
  onClose: () => void;
}

function TeamFormModal({ team, onClose }: TeamFormModalProps) {
  const queryClient = useQueryClient();
  const isEdit = !!team;

  const [form, setForm] = useState({
    title: team?.title ?? "",
    description: team?.description ?? "",
    color: team?.color ?? "#6366f1",
  });

  const createMutation = useMutation({
    mutationFn: () => createTeam(form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["teams"] });
      toast({ title: "Swarm created" });
      onClose();
    },
    onError: (e: Error) => toast({ title: "Error", description: e.message, variant: "destructive" }),
  });

  const updateMutation = useMutation({
    mutationFn: () => updateTeam(team!.id, form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["teams"] });
      toast({ title: "Swarm updated" });
      onClose();
    },
    onError: (e: Error) => toast({ title: "Error", description: e.message, variant: "destructive" }),
  });

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
      <div
          className="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 backdrop-blur-sm"
          onClick={onClose}
      >
        <div
            className="bg-card border border-border rounded-xl shadow-2xl w-full max-w-md mx-4 p-5 space-y-4"
            onClick={(e) => e.stopPropagation()}
        >
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-foreground">
              {isEdit ? "Edit Swarm" : "New Swarm"}
            </h3>
            <button onClick={onClose} className="text-muted-foreground hover:text-foreground">
              <X className="w-4 h-4" />
            </button>
          </div>

          <div className="space-y-1.5">
            <Label className="text-xs text-muted-foreground">Name *</Label>
            <Input
                value={form.title}
                onChange={(e) => setForm({ ...form, title: e.target.value })}
                placeholder="e.g. Data Team"
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
                placeholder="What does this swarm do?"
                rows={2}
                className="bg-input border-border text-foreground resize-none text-sm"
            />
          </div>

          <div className="space-y-1.5">
            <Label className="text-xs text-muted-foreground">Color</Label>
            <SwarmColorPicker
                value={form.color}
                onChange={(c) => setForm({ ...form, color: c })}
            />
          </div>

          <div className="flex gap-2 pt-1 border-t border-border">
            <Button
                onClick={() => isEdit ? updateMutation.mutate() : createMutation.mutate()}
                disabled={isPending || !form.title.trim()}
                size="sm"
                className="bg-primary text-primary-foreground hover:bg-primary/90"
            >
              {isEdit ? "Save Changes" : "Create Swarm"}
            </Button>
            <Button onClick={onClose} size="sm" variant="ghost" className="text-muted-foreground">
              Cancel
            </Button>
          </div>
        </div>
      </div>
  );
}

// ── Team Row ──────────────────────────────────────────────────────────────────

interface TeamRowProps {
  team: Team;
  isActive: boolean;
  collapsed: boolean;
  onSelect: () => void;
  onEdit: () => void;
  onDelete: () => void;
}

function TeamRow({ team, isActive, collapsed, onSelect, onEdit, onDelete }: TeamRowProps) {
  const [menuOpen, setMenuOpen] = useState(false);

  return (
      <div className="relative group">
        <button
            onClick={onSelect}
            className={cn(
                "w-full flex items-center gap-2.5 px-2 py-1.5 rounded-md text-sm transition-colors text-left",
                isActive
                    ? "bg-secondary/80 text-foreground font-medium"
                    : "text-muted-foreground hover:text-foreground hover:bg-secondary/40"
            )}
            style={
              isActive
                  ? { boxShadow: `inset 2px 0 0 ${team.color}` }
                  : undefined
            }
            title={collapsed ? team.title : undefined}
        >
        <span
            className="w-2.5 h-2.5 rounded-full flex-shrink-0"
            style={{ backgroundColor: team.color }}
        />
          {!collapsed && (
              <span className="flex-1 truncate">{team.title}</span>
          )}
        </button>

        {/* Context menu trigger — only show when not collapsed */}
        {!collapsed && (
            <button
                onClick={(e) => { e.stopPropagation(); setMenuOpen((v) => !v); }}
                className={cn(
                    "absolute right-1 top-1/2 -translate-y-1/2 p-0.5 rounded text-muted-foreground hover:text-foreground transition-opacity",
                    menuOpen ? "opacity-100" : "opacity-0 group-hover:opacity-100"
                )}
            >
              <MoreHorizontal className="w-3.5 h-3.5" />
            </button>
        )}

        {/* Dropdown menu */}
        {menuOpen && (
            <>
              <div className="fixed inset-0 z-[55]" onClick={() => setMenuOpen(false)} />
              <div className="absolute right-0 top-7 z-[56] bg-popover border border-border rounded-lg shadow-xl py-1 min-w-[140px]">
                <button
                    onClick={() => { setMenuOpen(false); onEdit(); }}
                    className="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-foreground hover:bg-secondary/60 transition-colors"
                >
                  <Pencil className="w-3.5 h-3.5 text-muted-foreground" />
                  Edit
                </button>
                <button
                    onClick={() => { setMenuOpen(false); onDelete(); }}
                    className="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-destructive hover:bg-destructive/10 transition-colors"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                  Delete
                </button>
              </div>
            </>
        )}
      </div>
  );
}

// ── Nav Links ─────────────────────────────────────────────────────────────────

type Tab = "chat" | "agents" | "skills" | "board" | "logs";

const NAV_ITEMS: { tab: Tab; label: string; icon: React.FC<{ className?: string }> }[] = [
  { tab: "chat", label: "Chat", icon: MessageCircle },
  { tab: "agents", label: "Agents", icon: Cpu },
  { tab: "board", label: "Board", icon: LayoutGrid },
  { tab: "skills", label: "Skills", icon: BookOpen },
  { tab: "logs", label: "Logs", icon: ScrollText },
];

// ── AppSidebar ─────────────────────────────────────────────────────────────────

interface AppSidebarProps {
  activeTab: Tab;
  onTabChange: (tab: Tab) => void;
  selectedTeamId: string | null;
  onTeamChange: (id: string | null) => void;
}

export function AppSidebar({
                             activeTab,
                             onTabChange,
                             selectedTeamId,
                             onTeamChange,
                           }: AppSidebarProps) {
  const queryClient = useQueryClient();
  const [collapsed, setCollapsed] = useState(false);
  const [showCreate, setShowCreate] = useState(false);
  const [editingTeam, setEditingTeam] = useState<Team | null>(null);
  const [deletingTeam, setDeletingTeam] = useState<Team | null>(null);

  const { data: teams = [] } = useQuery({
    queryKey: ["teams"],
    queryFn: listTeams,
    staleTime: 60_000,
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteTeam(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ["teams"] });
      queryClient.invalidateQueries({ queryKey: ["agents"] });
      if (selectedTeamId === id) onTeamChange(null);
      toast({ title: "Swarm deleted", description: "Agents have been unassigned." });
      setDeletingTeam(null);
    },
    onError: (e: Error) => toast({ title: "Error", description: e.message, variant: "destructive" }),
  });

  const sidebarWidth = collapsed ? "w-12" : "w-[220px]";

  return (
      <>
        <aside
            className={cn(
                "flex flex-col h-screen sticky top-0 flex-shrink-0 border-r border-border bg-card transition-all duration-200 overflow-hidden",
                sidebarWidth
            )}
        >
          {/* Logo */}
          <div className={cn(
              "flex items-center gap-2 px-3 h-14 border-b border-border flex-shrink-0",
              collapsed ? "justify-center" : ""
          )}>
            <Cpu className="w-5 h-5 text-primary flex-shrink-0" />
            {!collapsed && (
                <span className="text-sm font-bold tracking-tight text-foreground truncate">
              OpenMaguro🐟
            </span>
            )}
          </div>

          {/* Scrollable body */}
          <div className="flex-1 overflow-y-auto overflow-x-hidden py-3 space-y-1 px-2">

            {/* Team Swarms section */}
            {!collapsed && (
                <p className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground px-1 mb-1">
                  Swarms
                </p>
            )}

            {/* All Agents */}
            <button
                onClick={() => onTeamChange(null)}
                title={collapsed ? "All Agents" : undefined}
                className={cn(
                    "w-full flex items-center gap-2.5 px-2 py-1.5 rounded-md text-sm transition-colors",
                    selectedTeamId === null
                        ? "bg-secondary/80 text-foreground font-medium"
                        : "text-muted-foreground hover:text-foreground hover:bg-secondary/40"
                )}
            >
              <Users className="w-3 h-3 flex-shrink-0 text-muted-foreground" />
              {!collapsed && <span className="truncate">All Agents</span>}
            </button>

            {/* Team list */}
            {teams.map((team) => (
                <TeamRow
                    key={team.id}
                    team={team}
                    isActive={selectedTeamId === team.id}
                    collapsed={collapsed}
                    onSelect={() => onTeamChange(team.id)}
                    onEdit={() => setEditingTeam(team)}
                    onDelete={() => setDeletingTeam(team)}
                />
            ))}

            {/* New Swarm */}
            <button
                onClick={() => setShowCreate(true)}
                title={collapsed ? "New Swarm" : undefined}
                className="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-xs text-muted-foreground hover:text-primary hover:bg-primary/10 transition-colors mt-1"
            >
              <Plus className="w-3.5 h-3.5 flex-shrink-0" />
              {!collapsed && <span>New Swarm</span>}
            </button>

            {/* Divider */}
            <div className="border-t border-border my-2" />

            {/* Nav links */}
            {!collapsed && (
                <p className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground px-1 mb-1">
                  Views
                </p>
            )}

            {NAV_ITEMS.map(({ tab, label, icon: Icon }) => (
                <button
                    key={tab}
                    onClick={() => onTabChange(tab)}
                    title={collapsed ? label : undefined}
                    className={cn(
                        "w-full flex items-center gap-2.5 px-2 py-1.5 rounded-md text-sm transition-colors",
                        activeTab === tab
                            ? "bg-primary/15 text-primary font-medium"
                            : "text-muted-foreground hover:text-foreground hover:bg-secondary/40"
                    )}
                >
                  <Icon className="w-4 h-4 flex-shrink-0" />
                  {!collapsed && <span>{label}</span>}
                </button>
            ))}
          </div>

          {/* Collapse toggle */}
          <div className="border-t border-border px-2 py-2 flex-shrink-0">
            <button
                onClick={() => setCollapsed((v) => !v)}
                className="w-full flex items-center justify-center gap-1.5 text-xs text-muted-foreground hover:text-foreground px-2 py-1.5 rounded-md hover:bg-secondary/40 transition-colors"
            >
              {collapsed
                  ? <ChevronRight className="w-3.5 h-3.5" />
                  : <><ChevronLeft className="w-3.5 h-3.5" /><span>Collapse</span></>
              }
            </button>
          </div>
        </aside>

        {/* Modals */}
        {showCreate && <TeamFormModal onClose={() => setShowCreate(false)} />}
        {editingTeam && (
            <TeamFormModal
                team={editingTeam}
                onClose={() => setEditingTeam(null)}
            />
        )}
        {deletingTeam && (
            <ConfirmDialog
                open
                onOpenChange={(o) => { if (!o) setDeletingTeam(null); }}
                title={`Delete "${deletingTeam.title}"?`}
                description="Agents in this swarm will be unassigned but not deleted."
                confirmLabel="Delete Swarm"
                destructive
                onConfirm={() => deleteMutation.mutate(deletingTeam.id)}
            />
        )}
      </>
  );
}