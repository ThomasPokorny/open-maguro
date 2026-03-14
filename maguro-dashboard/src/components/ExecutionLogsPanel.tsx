import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { listExecutions, Execution } from "@/lib/api";
import { relativeTime, formatDuration } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { ChevronDown, ChevronRight, ScrollText, X } from "lucide-react";
import { Button } from "@/components/ui/button";

const STATUS_CONFIG: Record<
  Execution["status"],
  { label: string; cls: string }
> = {
  success: { label: "success", cls: "badge-success" },
  failure: { label: "failure", cls: "badge-failure" },
  running: { label: "running", cls: "badge-running" },
  pending: { label: "pending", cls: "badge-pending" },
};

interface ExecutionRowProps {
  execution: Execution;
}

function ExecutionRow({ execution }: ExecutionRowProps) {
  const [open, setOpen] = useState(false);
  const cfg = STATUS_CONFIG[execution.status] ?? STATUS_CONFIG.pending;
  const duration = formatDuration(execution.started_at, execution.finished_at);

  return (
    <div className="border-b border-border last:border-b-0">
      <div
        className="flex items-center gap-3 px-4 py-2.5 cursor-pointer hover:bg-muted/20 transition-colors"
        onClick={() => setOpen(!open)}
      >
        <button className="text-muted-foreground flex-shrink-0">
          {open ? <ChevronDown className="w-3.5 h-3.5" /> : <ChevronRight className="w-3.5 h-3.5" />}
        </button>
        <span
          className={cn(
            "text-xs font-medium px-2 py-0.5 rounded-full flex-shrink-0",
            cfg.cls
          )}
        >
          {cfg.label}
        </span>
        <span className="text-sm text-foreground truncate flex-1">
          {execution.task_name ?? "Unknown task"}
        </span>
        <span className="text-xs text-muted-foreground font-mono flex-shrink-0 hidden sm:block">
          {duration}
        </span>
        <span className="text-xs text-muted-foreground flex-shrink-0">
          {relativeTime(execution.started_at)}
        </span>
      </div>
      {open && (
        <div className="px-10 pb-3 space-y-2 bg-muted/10">
          {execution.summary && (
            <div>
              <p className="text-xs text-muted-foreground mb-1">Summary</p>
              <p className="text-sm text-foreground bg-muted/30 rounded p-2 font-mono text-xs leading-relaxed whitespace-pre-wrap">
                {execution.summary}
              </p>
            </div>
          )}
          {execution.error && (
            <div>
              <p className="text-xs text-muted-foreground mb-1">Error</p>
              <p className="text-sm text-destructive bg-destructive/10 rounded p-2 font-mono text-xs leading-relaxed whitespace-pre-wrap">
                {execution.error}
              </p>
            </div>
          )}
          {!execution.summary && !execution.error && (
            <p className="text-xs text-muted-foreground italic">No additional details</p>
          )}
        </div>
      )}
    </div>
  );
}

// --- Logs Panel ---

interface ExecutionLogsPanelProps {
  open: boolean;
  onClose: () => void;
}

export function ExecutionLogsPanel({ open, onClose }: ExecutionLogsPanelProps) {
  const [filter, setFilter] = useState<Execution["status"] | "all">("all");

  const { data: executions = [], isLoading } = useQuery({
    queryKey: ["executions"],
    queryFn: listExecutions,
    enabled: open,
    refetchInterval: open ? 10_000 : false,
  });

  const filtered =
    filter === "all" ? executions : executions.filter((e) => e.status === filter);

  if (!open) return null;

  return (
    <div className="fixed inset-x-0 bottom-0 z-50 max-h-[50vh] flex flex-col bg-card border-t border-border shadow-2xl">
      {/* Panel header */}
      <div className="flex items-center gap-3 px-4 py-2.5 border-b border-border flex-shrink-0">
        <ScrollText className="w-4 h-4 text-primary" />
        <span className="text-sm font-semibold text-foreground">Execution Logs</span>
        <span className="text-xs text-muted-foreground">{executions.length} entries</span>

        {/* Filter chips */}
        <div className="flex gap-1 ml-3">
          {(["all", "success", "failure", "running", "pending"] as const).map((s) => (
            <button
              key={s}
              onClick={() => setFilter(s)}
              className={cn(
                "text-xs px-2 py-0.5 rounded-full transition-colors",
                filter === s
                  ? s === "all"
                    ? "bg-secondary text-foreground"
                    : cn(STATUS_CONFIG[s as Execution["status"]]?.cls)
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              {s}
            </button>
          ))}
        </div>

        <Button
          size="sm"
          variant="ghost"
          onClick={onClose}
          className="ml-auto h-6 w-6 p-0 text-muted-foreground hover:text-foreground"
        >
          <X className="w-4 h-4" />
        </Button>
      </div>

      {/* Scrollable list */}
      <div className="overflow-y-auto flex-1">
        {isLoading && (
          <div className="space-y-px">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="h-10 bg-muted/20 animate-pulse" />
            ))}
          </div>
        )}
        {!isLoading && filtered.length === 0 && (
          <p className="text-center text-muted-foreground text-sm py-8">No executions found</p>
        )}
        {!isLoading && filtered.map((exec) => (
          <ExecutionRow key={exec.id} execution={exec} />
        ))}
      </div>
    </div>
  );
}
