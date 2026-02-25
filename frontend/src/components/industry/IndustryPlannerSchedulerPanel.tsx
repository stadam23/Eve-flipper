import { useI18n } from "@/lib/i18n";
import type { IndustryJobStatus } from "@/lib/types";
import type { Dispatch, SetStateAction } from "react";

interface IndustryPlannerSchedulerPanelProps {
  jobsWorkspaceTab: "guide" | "planning" | "operations";
  enablePlanScheduler: boolean;
  setEnablePlanScheduler: Dispatch<SetStateAction<boolean>>;
  schedulerSlotCount: number;
  setSchedulerSlotCount: Dispatch<SetStateAction<number>>;
  schedulerMaxRunsPerJob: number;
  setSchedulerMaxRunsPerJob: Dispatch<SetStateAction<number>>;
  schedulerMaxDurationHours: number;
  setSchedulerMaxDurationHours: Dispatch<SetStateAction<number>>;
  schedulerQueueStatus: IndustryJobStatus;
  setSchedulerQueueStatus: Dispatch<SetStateAction<IndustryJobStatus>>;
}

export function IndustryPlannerSchedulerPanel({
  jobsWorkspaceTab,
  enablePlanScheduler,
  setEnablePlanScheduler,
  schedulerSlotCount,
  setSchedulerSlotCount,
  schedulerMaxRunsPerJob,
  setSchedulerMaxRunsPerJob,
  schedulerMaxDurationHours,
  setSchedulerMaxDurationHours,
  schedulerQueueStatus,
  setSchedulerQueueStatus,
}: IndustryPlannerSchedulerPanelProps) {
  const { t } = useI18n();

  if (jobsWorkspaceTab !== "planning") {
    return null;
  }

  return (
    <div className="mt-2 flex flex-wrap items-center gap-2 border border-eve-border/40 rounded-sm px-2 py-1.5 bg-eve-dark/20">
      <label className="flex items-center gap-1 text-[11px] text-eve-dim select-none">
        <input
          type="checkbox"
          checked={enablePlanScheduler}
          onChange={(e) => setEnablePlanScheduler(e.target.checked)}
          className="accent-eve-accent"
        />
        {t("industryLedgerAutoSplitScheduler")}
      </label>
      {enablePlanScheduler && (
        <>
          <span className="text-[11px] text-eve-dim">{t("industryLedgerSchedulerSlots")}</span>
          <input
            type="number"
            min={1}
            max={64}
            value={schedulerSlotCount}
            onChange={(e) => setSchedulerSlotCount(Math.max(1, Math.min(64, Number(e.target.value) || 1)))}
            className="w-16 px-1.5 py-1 bg-eve-input border border-eve-border rounded-sm text-[11px] text-eve-text font-mono"
          />
          <span className="text-[11px] text-eve-dim">{t("industryLedgerSchedulerMaxRuns")}</span>
          <input
            type="number"
            min={1}
            value={schedulerMaxRunsPerJob}
            onChange={(e) => setSchedulerMaxRunsPerJob(Math.max(1, Number(e.target.value) || 1))}
            className="w-20 px-1.5 py-1 bg-eve-input border border-eve-border rounded-sm text-[11px] text-eve-text font-mono"
          />
          <span className="text-[11px] text-eve-dim">{t("industryLedgerSchedulerMaxHours")}</span>
          <input
            type="number"
            min={1}
            value={schedulerMaxDurationHours}
            onChange={(e) => setSchedulerMaxDurationHours(Math.max(1, Number(e.target.value) || 1))}
            className="w-20 px-1.5 py-1 bg-eve-input border border-eve-border rounded-sm text-[11px] text-eve-text font-mono"
          />
          <span className="text-[11px] text-eve-dim">{t("industryLedgerSchedulerQueueStatus")}</span>
          <select
            value={schedulerQueueStatus}
            onChange={(e) => setSchedulerQueueStatus(e.target.value as IndustryJobStatus)}
            className="px-1.5 py-1 bg-eve-input border border-eve-border rounded-sm text-[11px] text-eve-text"
          >
            <option value="queued">{t("industryLedgerStatusQueued")}</option>
            <option value="planned">{t("industryLedgerStatusPlanned")}</option>
          </select>
        </>
      )}
    </div>
  );
}
