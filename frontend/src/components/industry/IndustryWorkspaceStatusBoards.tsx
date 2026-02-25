import type { IndustryProjectSnapshot } from "@/lib/types";
import { industryJobStatusClass, industryTaskStatusClass } from "./industryHelpers";

interface IndustryWorkspaceStatusBoardsProps {
  jobsWorkspaceTab: "guide" | "planning" | "operations";
  taskStatusDone: number;
  taskStatusTotal: number;
  taskStatusDonePct: number;
  taskStatusBoard: Record<string, number>;
  jobStatusDone: number;
  jobStatusTotal: number;
  jobStatusDonePct: number;
  jobStatusBoard: Record<string, number>;
  ledgerSnapshot: IndustryProjectSnapshot | null;
  materialCoverageTotals: {
    required: number;
    stock: number;
    buy: number;
    build: number;
    missing: number;
  };
}

export function IndustryWorkspaceStatusBoards({
  jobsWorkspaceTab,
  taskStatusDone,
  taskStatusTotal,
  taskStatusDonePct,
  taskStatusBoard,
  jobStatusDone,
  jobStatusTotal,
  jobStatusDonePct,
  jobStatusBoard,
  ledgerSnapshot,
  materialCoverageTotals,
}: IndustryWorkspaceStatusBoardsProps) {
  if (jobsWorkspaceTab !== "guide" && jobsWorkspaceTab !== "operations") {
    return null;
  }

  return (
    <div className="mt-2 grid grid-cols-1 xl:grid-cols-3 gap-2">
      <div className="border border-eve-border/40 rounded-sm p-2 bg-eve-dark/20">
        <div className="flex items-center justify-between gap-2 mb-1">
          <div className="text-[10px] uppercase tracking-wider text-eve-dim">Task Status Board</div>
          <div className="text-[10px] text-eve-dim">
            {taskStatusDone}/{taskStatusTotal} ({taskStatusDonePct}%)
          </div>
        </div>
        <div className="h-1.5 rounded bg-eve-dark/50 overflow-hidden mb-2">
          <div
            className="h-full bg-emerald-500/70"
            style={{ width: `${taskStatusDonePct}%` }}
          />
        </div>
        <div className="flex flex-wrap gap-1">
          {["planned", "ready", "active", "paused", "blocked", "completed", "cancelled"].map((status) => (
            <span
              key={`task-status-chip-${status}`}
              className={`px-1.5 py-0.5 text-[10px] uppercase rounded-sm border ${industryTaskStatusClass(status)}`}
            >
              {status}:{taskStatusBoard[status] || 0}
            </span>
          ))}
        </div>
      </div>

      <div className="border border-eve-border/40 rounded-sm p-2 bg-eve-dark/20">
        <div className="flex items-center justify-between gap-2 mb-1">
          <div className="text-[10px] uppercase tracking-wider text-eve-dim">Job Status Board</div>
          <div className="text-[10px] text-eve-dim">
            {jobStatusDone}/{jobStatusTotal} ({jobStatusDonePct}%)
          </div>
        </div>
        <div className="h-1.5 rounded bg-eve-dark/50 overflow-hidden mb-2">
          <div
            className="h-full bg-blue-500/70"
            style={{ width: `${jobStatusDonePct}%` }}
          />
        </div>
        <div className="flex flex-wrap gap-1">
          {["planned", "queued", "active", "paused", "completed", "failed", "cancelled"].map((status) => (
            <span
              key={`job-status-chip-${status}`}
              className={`px-1.5 py-0.5 text-[10px] uppercase rounded-sm border ${industryJobStatusClass(status)}`}
            >
              {status}:{jobStatusBoard[status] || 0}
            </span>
          ))}
        </div>
      </div>

      <div className="border border-eve-border/40 rounded-sm p-2 bg-eve-dark/20">
        <div className="text-[10px] uppercase tracking-wider text-eve-dim mb-1">Material Coverage Board</div>
        <div className="grid grid-cols-2 gap-1 text-[11px] text-eve-dim">
          <span>rows</span>
          <span className="text-right font-mono text-eve-text">{ledgerSnapshot?.material_diff.length ?? 0}</span>
          <span>required</span>
          <span className="text-right font-mono text-eve-accent">{materialCoverageTotals.required.toLocaleString()}</span>
          <span>stock</span>
          <span className="text-right font-mono text-cyan-300">{materialCoverageTotals.stock.toLocaleString()}</span>
          <span>buy</span>
          <span className="text-right font-mono text-eve-dim">{materialCoverageTotals.buy.toLocaleString()}</span>
          <span>build</span>
          <span className="text-right font-mono text-fuchsia-300">{materialCoverageTotals.build.toLocaleString()}</span>
          <span>missing</span>
          <span className="text-right font-mono text-red-300">{materialCoverageTotals.missing.toLocaleString()}</span>
        </div>
      </div>
    </div>
  );
}
