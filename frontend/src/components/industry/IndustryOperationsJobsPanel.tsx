import { useI18n } from "@/lib/i18n";
import type { Dispatch, SetStateAction } from "react";
import type { IndustryJobStatus, IndustryLedger, IndustryLedgerEntry } from "@/lib/types";
import type { IndustryJobsWorkspaceTab } from "./IndustryJobsWorkspaceNav";
import { IndustrySummaryCard } from "./IndustrySummaryCard";

interface IndustryOperationsJobsContext {
  jobsWorkspaceTab: IndustryJobsWorkspaceTab;
  ledgerData: IndustryLedger | null;
  formatISK: (value: number) => string;
  industryJobStatusClass: (status: string) => string;
  formatUtcShort: (value: string) => string;
  selectedLedgerJobIDs: number[];
  updatingLedgerJobsBulk: boolean;
  handleBulkSetLedgerJobStatus: (status: IndustryJobStatus) => Promise<void>;
  setSelectedLedgerJobIDs: Dispatch<SetStateAction<number[]>>;
  allVisibleLedgerJobsSelected: boolean;
  handleSelectAllVisibleLedgerJobs: (selected: boolean) => void;
  selectedLedgerJobIDSet: Set<number>;
  toggleLedgerJobSelection: (jobId: number, selected: boolean) => void;
  handleSetLedgerJobStatus: (jobId: number, status: IndustryJobStatus) => Promise<void>;
  updatingLedgerJobId: number;
}

interface IndustryOperationsJobsPanelProps {
  ctx: IndustryOperationsJobsContext;
}

export function IndustryOperationsJobsPanel({ ctx }: IndustryOperationsJobsPanelProps) {
  const { t } = useI18n();
  const {
    jobsWorkspaceTab,
    ledgerData,
    formatISK,
    industryJobStatusClass,
    formatUtcShort,
    selectedLedgerJobIDs,
    updatingLedgerJobsBulk,
    handleBulkSetLedgerJobStatus,
    setSelectedLedgerJobIDs,
    allVisibleLedgerJobsSelected,
    handleSelectAllVisibleLedgerJobs,
    selectedLedgerJobIDSet,
    toggleLedgerJobSelection,
    handleSetLedgerJobStatus,
    updatingLedgerJobId,
  } = ctx;

  return (
    <>
      {jobsWorkspaceTab === "operations" && ledgerData && (
        <div className="mt-2 grid grid-cols-2 md:grid-cols-6 gap-2">
          <IndustrySummaryCard label={t("industryLedgerJobs")} value={String(ledgerData.total)} color="text-eve-accent" />
          <IndustrySummaryCard label={t("industryLedgerPlanned")} value={String(ledgerData.planned)} color="text-eve-dim" />
          <IndustrySummaryCard label={t("industryLedgerActive")} value={String(ledgerData.active)} color="text-blue-400" />
          <IndustrySummaryCard label={t("industryLedgerCompleted")} value={String(ledgerData.completed)} color="text-green-400" />
          <IndustrySummaryCard label={t("industryLedgerFailed")} value={String(ledgerData.failed)} color="text-red-400" />
          <IndustrySummaryCard label={t("industryLedgerCost")} value={formatISK(ledgerData.total_cost_isk || 0)} color="text-eve-accent" />
        </div>
      )}

      {jobsWorkspaceTab === "operations" && (
        <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
          <span className="text-eve-dim">{t("industryLedgerSelected")}: {selectedLedgerJobIDs.length}</span>
          <button
            type="button"
            onClick={() => { void handleBulkSetLedgerJobStatus("active"); }}
            disabled={updatingLedgerJobsBulk || selectedLedgerJobIDs.length === 0}
            className="px-2 py-1 border border-blue-500/40 text-blue-300 rounded-sm hover:bg-blue-500/10 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {t("industryLedgerSetActive")}
          </button>
          <button
            type="button"
            onClick={() => { void handleBulkSetLedgerJobStatus("paused"); }}
            disabled={updatingLedgerJobsBulk || selectedLedgerJobIDs.length === 0}
            className="px-2 py-1 border border-indigo-500/40 text-indigo-300 rounded-sm hover:bg-indigo-500/10 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {t("industryLedgerPause")}
          </button>
          <button
            type="button"
            onClick={() => { void handleBulkSetLedgerJobStatus("queued"); }}
            disabled={updatingLedgerJobsBulk || selectedLedgerJobIDs.length === 0}
            className="px-2 py-1 border border-cyan-500/40 text-cyan-300 rounded-sm hover:bg-cyan-500/10 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {t("industryLedgerResume")}
          </button>
          <button
            type="button"
            onClick={() => { void handleBulkSetLedgerJobStatus("completed"); }}
            disabled={updatingLedgerJobsBulk || selectedLedgerJobIDs.length === 0}
            className="px-2 py-1 border border-emerald-500/40 text-emerald-300 rounded-sm hover:bg-emerald-500/10 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {t("industryLedgerSetCompleted")}
          </button>
          <button
            type="button"
            onClick={() => setSelectedLedgerJobIDs([])}
            disabled={selectedLedgerJobIDs.length === 0}
            className="px-2 py-1 border border-eve-border text-eve-dim rounded-sm hover:text-eve-accent hover:border-eve-accent/40 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {t("industryLedgerClearSelection")}
          </button>
        </div>
      )}

      {jobsWorkspaceTab === "operations" && (
        <div className="mt-2 border border-eve-border rounded-sm bg-eve-dark/30 max-h-[240px] overflow-auto">
          <table className="w-full text-xs">
            <thead className="sticky top-0 bg-eve-dark z-10">
              <tr className="text-eve-dim text-[10px] uppercase tracking-wider border-b border-eve-border/60">
                <th className="px-2 py-1.5 text-left">
                  <input
                    type="checkbox"
                    checked={allVisibleLedgerJobsSelected}
                    onChange={(e) => handleSelectAllVisibleLedgerJobs(e.target.checked)}
                    className="accent-eve-accent"
                  />
                </th>
                <th className="px-2 py-1.5 text-left">{t("industryLedgerJob")}</th>
                <th className="px-2 py-1.5 text-left">{t("industryLedgerTaskActivity")}</th>
                <th className="px-2 py-1.5 text-right">{t("industryLedgerRuns")}</th>
                <th className="px-2 py-1.5 text-right">{t("industryLedgerCost")}</th>
                <th className="px-2 py-1.5 text-left">{t("industryLedgerStatus")}</th>
                <th className="px-2 py-1.5 text-left">{t("industryLedgerUpdated")}</th>
                <th className="px-2 py-1.5 text-right">{t("industryLedgerActions")}</th>
              </tr>
            </thead>
            <tbody>
              {(ledgerData?.entries ?? []).map((entry: IndustryLedgerEntry) => (
                <tr key={entry.job_id} className="border-b border-eve-border/40 hover:bg-eve-accent/5">
                  <td className="px-2 py-1.5 text-eve-dim">
                    <input
                      type="checkbox"
                      checked={selectedLedgerJobIDSet.has(entry.job_id)}
                      onChange={(e) => toggleLedgerJobSelection(entry.job_id, e.target.checked)}
                      className="accent-eve-accent"
                    />
                  </td>
                  <td className="px-2 py-1.5 text-eve-dim">#{entry.job_id}</td>
                  <td className="px-2 py-1.5 text-eve-text">
                    <div className="truncate">{entry.task_name || `${t("industryLedgerTask")} ${entry.task_id || "n/a"}`}</div>
                    <div className="text-[10px] text-eve-dim">{entry.activity}</div>
                  </td>
                  <td className="px-2 py-1.5 text-right text-eve-accent font-mono">{entry.runs}</td>
                  <td className="px-2 py-1.5 text-right text-eve-dim font-mono">{formatISK(entry.cost_isk || 0)}</td>
                  <td className="px-2 py-1.5">
                    <span className={`px-1.5 py-0.5 text-[10px] uppercase rounded-sm border ${industryJobStatusClass(entry.status)}`}>
                      {entry.status}
                    </span>
                  </td>
                  <td className="px-2 py-1.5 text-eve-dim whitespace-nowrap">{formatUtcShort(entry.updated_at)}</td>
                  <td className="px-2 py-1.5 text-right">
                    <div className="inline-flex gap-1">
                      {entry.status !== "active" && entry.status !== "completed" && entry.status !== "cancelled" && (
                        <button
                          type="button"
                          onClick={() => { void handleSetLedgerJobStatus(entry.job_id, "active"); }}
                          disabled={updatingLedgerJobId === entry.job_id || updatingLedgerJobsBulk}
                          className="px-1.5 py-0.5 text-[10px] border border-blue-500/40 text-blue-300 rounded-sm hover:bg-blue-500/10 disabled:opacity-50"
                        >
                          {t("industryLedgerSetActive")}
                        </button>
                      )}
                      {entry.status === "paused" ? (
                        <button
                          type="button"
                          onClick={() => { void handleSetLedgerJobStatus(entry.job_id, "queued"); }}
                          disabled={updatingLedgerJobId === entry.job_id || updatingLedgerJobsBulk}
                          className="px-1.5 py-0.5 text-[10px] border border-cyan-500/40 text-cyan-300 rounded-sm hover:bg-cyan-500/10 disabled:opacity-50"
                        >
                          {t("industryLedgerResume")}
                        </button>
                      ) : (
                        entry.status !== "completed" && entry.status !== "cancelled" && (
                          <button
                            type="button"
                            onClick={() => { void handleSetLedgerJobStatus(entry.job_id, "paused"); }}
                            disabled={updatingLedgerJobId === entry.job_id || updatingLedgerJobsBulk}
                            className="px-1.5 py-0.5 text-[10px] border border-indigo-500/40 text-indigo-300 rounded-sm hover:bg-indigo-500/10 disabled:opacity-50"
                          >
                            {t("industryLedgerPause")}
                          </button>
                        )
                      )}
                      {entry.status !== "completed" && entry.status !== "cancelled" && (
                        <button
                          type="button"
                          onClick={() => { void handleSetLedgerJobStatus(entry.job_id, "completed"); }}
                          disabled={updatingLedgerJobId === entry.job_id || updatingLedgerJobsBulk}
                          className="px-1.5 py-0.5 text-[10px] border border-emerald-500/40 text-emerald-300 rounded-sm hover:bg-emerald-500/10 disabled:opacity-50"
                        >
                          {t("industryLedgerSetCompleted")}
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
              {(!ledgerData || ledgerData.entries.length === 0) && (
                <tr>
                  <td colSpan={8} className="px-2 py-4 text-center text-eve-dim">
                    {t("industryLedgerNoJobs")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </>
  );
}
