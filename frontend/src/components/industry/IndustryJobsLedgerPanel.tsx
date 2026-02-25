import type { ComponentProps } from "react";
import { useI18n } from "@/lib/i18n";
import type { IndustryJobsWorkspaceTab } from "./IndustryJobsWorkspaceNav";
import { IndustryJobsWorkspaceNav } from "./IndustryJobsWorkspaceNav";
import { IndustryJobsGuidePanel } from "./IndustryJobsGuidePanel";
import { IndustryPlannerWarningLog } from "./IndustryPlannerWarningLog";
import { IndustryDependencyBoard } from "./IndustryDependencyBoard";
import { IndustryOperationsBoards } from "./IndustryOperationsBoards";
import { IndustryPlannerBuilderPanel } from "./IndustryPlannerBuilderPanel";
import { IndustryOperationsJobsPanel } from "./IndustryOperationsJobsPanel";
import { IndustryJobsProjectHeader } from "./IndustryJobsProjectHeader";
import { IndustryJobsPlanningActions } from "./IndustryJobsPlanningActions";
import { IndustryPlannerSchedulerPanel } from "./IndustryPlannerSchedulerPanel";
import { IndustryPlanPreviewPanel } from "./IndustryPlanPreviewPanel";
import { IndustryWorkspaceStatusBoards } from "./IndustryWorkspaceStatusBoards";

interface IndustryJobsLedgerPanelProps {
  isLoggedIn: boolean;
  ledgerProjectsLoading: boolean;
  jobsWorkspaceTab: IndustryJobsWorkspaceTab;
  setJobsWorkspaceTab: (tab: IndustryJobsWorkspaceTab) => void;
  warningsCount: number;
  missingBindings: number;
  activeJobs: number;
  projectHeaderProps: ComponentProps<typeof IndustryJobsProjectHeader>;
  guidePanelProps: ComponentProps<typeof IndustryJobsGuidePanel>;
  planningActionsProps: ComponentProps<typeof IndustryJobsPlanningActions>;
  warningLogProps: ComponentProps<typeof IndustryPlannerWarningLog>;
  workspaceStatusBoardsProps: ComponentProps<typeof IndustryWorkspaceStatusBoards>;
  dependencyBoardProps: ComponentProps<typeof IndustryDependencyBoard>;
  schedulerPanelProps: ComponentProps<typeof IndustryPlannerSchedulerPanel>;
  planPreviewPanelProps: ComponentProps<typeof IndustryPlanPreviewPanel>;
  operationsBoardsProps: ComponentProps<typeof IndustryOperationsBoards>;
  plannerBuilderProps: ComponentProps<typeof IndustryPlannerBuilderPanel>;
  operationsJobsProps: ComponentProps<typeof IndustryOperationsJobsPanel>;
}

export function IndustryJobsLedgerPanel({
  isLoggedIn,
  ledgerProjectsLoading,
  jobsWorkspaceTab,
  setJobsWorkspaceTab,
  warningsCount,
  missingBindings,
  activeJobs,
  projectHeaderProps,
  guidePanelProps,
  planningActionsProps,
  warningLogProps,
  workspaceStatusBoardsProps,
  dependencyBoardProps,
  schedulerPanelProps,
  planPreviewPanelProps,
  operationsBoardsProps,
  plannerBuilderProps,
  operationsJobsProps,
}: IndustryJobsLedgerPanelProps) {
  const { t } = useI18n();

  return (
    <div className="shrink-0 m-2 mt-0 pb-2">
      <div className="bg-eve-panel border border-eve-border rounded-sm p-3">
        <div className="flex items-center justify-between gap-2 mb-2">
          <div className="text-[10px] uppercase tracking-wider text-eve-dim">{t("industryLedgerTitle")}</div>
          {ledgerProjectsLoading && <span className="text-[10px] text-eve-dim">{t("industryLedgerSyncingProjects")}</span>}
        </div>
        {!isLoggedIn ? (
          <div className="text-xs text-eve-dim">
            {t("industryLedgerLoginRequired")}
          </div>
        ) : (
          <>
            <IndustryJobsProjectHeader {...projectHeaderProps} />

            <IndustryJobsWorkspaceNav
              activeTab={jobsWorkspaceTab}
              onChange={setJobsWorkspaceTab}
              warningsCount={warningsCount}
              missingBindings={missingBindings}
              activeJobs={activeJobs}
            />

            {jobsWorkspaceTab === "guide" && (
              <IndustryJobsGuidePanel {...guidePanelProps} />
            )}

            <IndustryJobsPlanningActions {...planningActionsProps} />

            {(jobsWorkspaceTab === "guide" || jobsWorkspaceTab === "planning") && (
              <IndustryPlannerWarningLog {...warningLogProps} />
            )}

            <IndustryWorkspaceStatusBoards {...workspaceStatusBoardsProps} />

            {jobsWorkspaceTab === "planning" && (
              <IndustryDependencyBoard {...dependencyBoardProps} />
            )}

            <IndustryPlannerSchedulerPanel {...schedulerPanelProps} />

            <IndustryPlanPreviewPanel {...planPreviewPanelProps} />

            <IndustryOperationsBoards {...operationsBoardsProps} />

            <IndustryPlannerBuilderPanel {...plannerBuilderProps} />

            <IndustryOperationsJobsPanel {...operationsJobsProps} />
          </>
        )}
      </div>
    </div>
  );
}
