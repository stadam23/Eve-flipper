import type { Dispatch, SetStateAction } from "react";
import { useI18n } from "@/lib/i18n";
import type { IndustryProjectSnapshot, IndustryTaskStatus } from "@/lib/types";
import type { IndustryTaskDependencyBoard } from "./industryHelpers";
import type { IndustryJobsWorkspaceTab } from "./IndustryJobsWorkspaceNav";

interface IndustryTaskBoardPanelProps {
  jobsWorkspaceTab: IndustryJobsWorkspaceTab;
  ledgerSnapshot: IndustryProjectSnapshot | null;
  selectedLedgerTaskIDs: number[];
  bulkLedgerTaskPriority: number;
  setBulkLedgerTaskPriority: Dispatch<SetStateAction<number>>;
  handleBulkSetLedgerTaskPriority: (priority: number) => Promise<void>;
  updatingLedgerTasksBulk: boolean;
  handleBulkSetLedgerTaskStatus: (status: IndustryTaskStatus) => Promise<void>;
  setSelectedLedgerTaskIDs: Dispatch<SetStateAction<number[]>>;
  allVisibleLedgerTasksSelected: boolean;
  handleSelectAllVisibleLedgerTasks: (selected: boolean) => void;
  selectedLedgerTaskIDSet: Set<number>;
  toggleLedgerTaskSelection: (taskId: number, selected: boolean) => void;
  industryTaskStatusClass: (status: string) => string;
  formatUtcShort: (value: string) => string;
  handleSetLedgerTaskPriority: (taskId: number, priority: number) => Promise<void>;
  updatingLedgerTaskId: number;
  handleSetLedgerTaskStatus: (taskId: number, status: IndustryTaskStatus) => Promise<void>;
  taskDependencyBoard: IndustryTaskDependencyBoard;
}

export function IndustryTaskBoardPanel({
  jobsWorkspaceTab,
  ledgerSnapshot,
  selectedLedgerTaskIDs,
  bulkLedgerTaskPriority,
  setBulkLedgerTaskPriority,
  handleBulkSetLedgerTaskPriority,
  updatingLedgerTasksBulk,
  handleBulkSetLedgerTaskStatus,
  setSelectedLedgerTaskIDs,
  allVisibleLedgerTasksSelected,
  handleSelectAllVisibleLedgerTasks,
  selectedLedgerTaskIDSet,
  toggleLedgerTaskSelection,
  industryTaskStatusClass,
  formatUtcShort,
  handleSetLedgerTaskPriority,
  updatingLedgerTaskId,
  handleSetLedgerTaskStatus,
  taskDependencyBoard,
}: IndustryTaskBoardPanelProps) {
  const { t } = useI18n();
  const parentMissingCount = Object.values(taskDependencyBoard.parent_missing_by_task).reduce(
    (acc, isMissing) => (isMissing ? acc + 1 : acc),
    0
  );

  if (jobsWorkspaceTab !== "operations" || !ledgerSnapshot) {
    return null;
  }

  return (
    <div className="mt-2 border border-eve-border/40 rounded-sm p-2 bg-eve-dark/20">
      <div className="flex items-center justify-between gap-2 mb-1">
        <div className="text-[10px] uppercase tracking-wider text-eve-dim">
          {t("industryLedgerTaskBoardTitle", { count: ledgerSnapshot.tasks.length })}
        </div>
        <div className="inline-flex items-center gap-1 text-[11px]">
          {parentMissingCount > 0 && (
            <span className="px-1.5 py-0.5 rounded-sm border border-yellow-500/40 text-yellow-300 bg-yellow-500/10">
              {t("industryLedgerTaskBoardParentMissing", { count: parentMissingCount })}
            </span>
          )}
          {taskDependencyBoard.orphans > 0 && (
            <span className="px-1.5 py-0.5 rounded-sm border border-amber-500/40 text-amber-300 bg-amber-500/10">
              {t("industryLedgerTaskBoardOrphans", { count: taskDependencyBoard.orphans })}
            </span>
          )}
          {taskDependencyBoard.cycles > 0 && (
            <span className="px-1.5 py-0.5 rounded-sm border border-red-500/40 text-red-300 bg-red-500/10">
              {t("industryLedgerTaskBoardCycles", { count: taskDependencyBoard.cycles })}
            </span>
          )}
          <span className="text-eve-dim">{t("industryLedgerSelected")}: {selectedLedgerTaskIDs.length}</span>
          <input
            type="number"
            value={bulkLedgerTaskPriority}
            onChange={(e) => setBulkLedgerTaskPriority(Math.round(Number(e.target.value) || 0))}
            className="w-16 px-1.5 py-0.5 bg-eve-input border border-eve-border rounded-sm text-[11px] text-eve-text font-mono"
            title={t("industryLedgerTaskBoardBulkPriorityTitle")}
          />
          <button
            type="button"
            onClick={() => { void handleBulkSetLedgerTaskPriority(bulkLedgerTaskPriority); }}
            disabled={updatingLedgerTasksBulk || selectedLedgerTaskIDs.length === 0}
            className="px-1.5 py-0.5 border border-fuchsia-500/40 text-fuchsia-300 rounded-sm hover:bg-fuchsia-500/10 disabled:opacity-50"
          >
            {t("industryLedgerTaskBoardPriorityShort")}
          </button>
          <button
            type="button"
            onClick={() => { void handleBulkSetLedgerTaskStatus("ready"); }}
            disabled={updatingLedgerTasksBulk || selectedLedgerTaskIDs.length === 0}
            className="px-1.5 py-0.5 border border-amber-500/40 text-amber-300 rounded-sm hover:bg-amber-500/10 disabled:opacity-50"
          >
            {t("industryLedgerTaskBoardReady")}
          </button>
          <button
            type="button"
            onClick={() => { void handleBulkSetLedgerTaskStatus("active"); }}
            disabled={updatingLedgerTasksBulk || selectedLedgerTaskIDs.length === 0}
            className="px-1.5 py-0.5 border border-blue-500/40 text-blue-300 rounded-sm hover:bg-blue-500/10 disabled:opacity-50"
          >
            {t("industryLedgerSetActive")}
          </button>
          <button
            type="button"
            onClick={() => { void handleBulkSetLedgerTaskStatus("paused"); }}
            disabled={updatingLedgerTasksBulk || selectedLedgerTaskIDs.length === 0}
            className="px-1.5 py-0.5 border border-indigo-500/40 text-indigo-300 rounded-sm hover:bg-indigo-500/10 disabled:opacity-50"
          >
            {t("industryLedgerTaskBoardFreeze")}
          </button>
          <button
            type="button"
            onClick={() => { void handleBulkSetLedgerTaskStatus("ready"); }}
            disabled={updatingLedgerTasksBulk || selectedLedgerTaskIDs.length === 0}
            className="px-1.5 py-0.5 border border-cyan-500/40 text-cyan-300 rounded-sm hover:bg-cyan-500/10 disabled:opacity-50"
          >
            {t("industryLedgerTaskBoardUnfreeze")}
          </button>
          <button
            type="button"
            onClick={() => { void handleBulkSetLedgerTaskStatus("completed"); }}
            disabled={updatingLedgerTasksBulk || selectedLedgerTaskIDs.length === 0}
            className="px-1.5 py-0.5 border border-emerald-500/40 text-emerald-300 rounded-sm hover:bg-emerald-500/10 disabled:opacity-50"
          >
            {t("industryLedgerTaskBoardComplete")}
          </button>
          <button
            type="button"
            onClick={() => setSelectedLedgerTaskIDs([])}
            disabled={selectedLedgerTaskIDs.length === 0}
            className="px-1.5 py-0.5 border border-eve-border text-eve-dim rounded-sm hover:text-eve-accent hover:border-eve-accent/40 disabled:opacity-50"
          >
            {t("industryLedgerClearSelection")}
          </button>
        </div>
      </div>
      <div className="border border-eve-border rounded-sm max-h-[220px] overflow-auto">
        <table className="w-full text-[11px]">
          <thead className="sticky top-0 bg-eve-dark z-10">
            <tr className="text-eve-dim uppercase tracking-wider border-b border-eve-border/60">
              <th className="px-1.5 py-1 text-left">
                <input
                  type="checkbox"
                  checked={allVisibleLedgerTasksSelected}
                  onChange={(e) => handleSelectAllVisibleLedgerTasks(e.target.checked)}
                  className="accent-eve-accent"
                />
              </th>
              <th className="px-1.5 py-1 text-left">{t("industryLedgerTask")}</th>
              <th className="px-1.5 py-1 text-left">{t("industryLedgerActivity")}</th>
              <th className="px-1.5 py-1 text-right">{t("industryLedgerRuns")}</th>
              <th className="px-1.5 py-1 text-right">{t("industryLedgerTaskBoardPriority")}</th>
              <th className="px-1.5 py-1 text-left">{t("industryLedgerTaskBoardParent")}</th>
              <th className="px-1.5 py-1 text-right">{t("industryLedgerTaskBoardDepth")}</th>
              <th className="px-1.5 py-1 text-left">{t("industryLedgerStatus")}</th>
              <th className="px-1.5 py-1 text-left">{t("industryLedgerTaskBoardWindow")}</th>
              <th className="px-1.5 py-1 text-right">{t("industryLedgerActions")}</th>
            </tr>
          </thead>
          <tbody>
            {ledgerSnapshot.tasks.map((task) => (
              <tr key={`task-board-${task.id}`} className="border-b border-eve-border/30">
                <td className="px-1.5 py-1 text-eve-dim">
                  <input
                    type="checkbox"
                    checked={selectedLedgerTaskIDSet.has(task.id)}
                    onChange={(e) => toggleLedgerTaskSelection(task.id, e.target.checked)}
                    className="accent-eve-accent"
                  />
                </td>
                <td className="px-1.5 py-1 text-eve-text">
                  <div className="truncate">{task.name}</div>
                  <div className="text-[10px] text-eve-dim">#{task.id}</div>
                </td>
                <td className="px-1.5 py-1 text-eve-dim">{task.activity}</td>
                <td className="px-1.5 py-1 text-right text-eve-accent font-mono">{task.target_runs || 0}</td>
                <td className="px-1.5 py-1 text-right text-eve-dim font-mono">{task.priority || 0}</td>
                <td className="px-1.5 py-1 text-eve-dim">
                  {taskDependencyBoard.parent_by_task[task.id] ? (
                    <span className={taskDependencyBoard.parent_missing_by_task[task.id] ? "text-yellow-300" : "text-eve-dim"}>
                      #{taskDependencyBoard.parent_by_task[task.id]}
                    </span>
                  ) : (
                    "—"
                  )}
                </td>
                <td className="px-1.5 py-1 text-right">
                  <div className="inline-flex items-center gap-1">
                    <span className="text-eve-dim font-mono">{taskDependencyBoard.depth_by_task[task.id] || 1}</span>
                    {taskDependencyBoard.critical_task_ids.has(task.id) && (
                      <span className="px-1 py-0.5 text-[9px] uppercase rounded-sm border border-fuchsia-500/40 text-fuchsia-300 bg-fuchsia-500/10">
                        CP
                      </span>
                    )}
                  </div>
                </td>
                <td className="px-1.5 py-1">
                  <span className={`px-1.5 py-0.5 text-[10px] uppercase rounded-sm border ${industryTaskStatusClass(task.status)}`}>
                    {task.status}
                  </span>
                </td>
                <td className="px-1.5 py-1 text-eve-dim whitespace-nowrap">
                  {formatUtcShort(task.planned_start)} - {formatUtcShort(task.planned_end)}
                </td>
                <td className="px-1.5 py-1 text-right">
                  <div className="inline-flex gap-1">
                    <button
                      type="button"
                      onClick={() => { void handleSetLedgerTaskPriority(task.id, (task.priority || 0) + 10); }}
                      disabled={updatingLedgerTaskId === task.id || updatingLedgerTasksBulk}
                      className="px-1 py-0.5 text-[10px] border border-fuchsia-500/40 text-fuchsia-300 rounded-sm hover:bg-fuchsia-500/10 disabled:opacity-50"
                      title={t("industryLedgerTaskBoardPriorityUpTitle")}
                    >
                      +P
                    </button>
                    <button
                      type="button"
                      onClick={() => { void handleSetLedgerTaskPriority(task.id, (task.priority || 0) - 10); }}
                      disabled={updatingLedgerTaskId === task.id || updatingLedgerTasksBulk}
                      className="px-1 py-0.5 text-[10px] border border-fuchsia-500/40 text-fuchsia-300 rounded-sm hover:bg-fuchsia-500/10 disabled:opacity-50"
                      title={t("industryLedgerTaskBoardPriorityDownTitle")}
                    >
                      -P
                    </button>
                    {task.status !== "ready" && task.status !== "completed" && task.status !== "cancelled" && (
                      <button
                        type="button"
                        onClick={() => { void handleSetLedgerTaskStatus(task.id, "ready"); }}
                        disabled={updatingLedgerTaskId === task.id || updatingLedgerTasksBulk}
                        className="px-1 py-0.5 text-[10px] border border-amber-500/40 text-amber-300 rounded-sm hover:bg-amber-500/10 disabled:opacity-50"
                      >
                        {t("industryLedgerTaskBoardReady")}
                      </button>
                    )}
                    {task.status !== "active" && task.status !== "completed" && task.status !== "cancelled" && (
                      <button
                        type="button"
                        onClick={() => { void handleSetLedgerTaskStatus(task.id, "active"); }}
                        disabled={updatingLedgerTaskId === task.id || updatingLedgerTasksBulk}
                        className="px-1 py-0.5 text-[10px] border border-blue-500/40 text-blue-300 rounded-sm hover:bg-blue-500/10 disabled:opacity-50"
                      >
                        {t("industryLedgerSetActive")}
                      </button>
                    )}
                    {task.status !== "completed" && task.status !== "cancelled" && (
                      <button
                        type="button"
                        onClick={() => { void handleSetLedgerTaskStatus(task.id, "completed"); }}
                        disabled={updatingLedgerTaskId === task.id || updatingLedgerTasksBulk}
                        className="px-1 py-0.5 text-[10px] border border-emerald-500/40 text-emerald-300 rounded-sm hover:bg-emerald-500/10 disabled:opacity-50"
                      >
                        {t("industryLedgerTaskBoardComplete")}
                      </button>
                    )}
                    {task.status === "paused" ? (
                      <button
                        type="button"
                        onClick={() => { void handleSetLedgerTaskStatus(task.id, "ready"); }}
                        disabled={updatingLedgerTaskId === task.id || updatingLedgerTasksBulk}
                        className="px-1 py-0.5 text-[10px] border border-cyan-500/40 text-cyan-300 rounded-sm hover:bg-cyan-500/10 disabled:opacity-50"
                      >
                        {t("industryLedgerTaskBoardUnfreeze")}
                      </button>
                    ) : (
                      task.status !== "completed" && task.status !== "cancelled" && (
                        <button
                          type="button"
                          onClick={() => { void handleSetLedgerTaskStatus(task.id, "paused"); }}
                          disabled={updatingLedgerTaskId === task.id || updatingLedgerTasksBulk}
                          className="px-1 py-0.5 text-[10px] border border-indigo-500/40 text-indigo-300 rounded-sm hover:bg-indigo-500/10 disabled:opacity-50"
                        >
                          {t("industryLedgerTaskBoardFreeze")}
                        </button>
                      )
                    )}
                  </div>
                </td>
              </tr>
            ))}
            {ledgerSnapshot.tasks.length === 0 && (
              <tr>
                <td colSpan={10} className="px-2 py-2 text-center text-eve-dim">
                  {t("industryLedgerTaskBoardNoTasks")}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
