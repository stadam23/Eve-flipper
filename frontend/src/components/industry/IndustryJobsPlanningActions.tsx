import { useI18n } from "@/lib/i18n";
import type { Dispatch, SetStateAction } from "react";
import type { BuildableItem, IndustryAnalysis, IndustryPlanPreview } from "@/lib/types";

interface IndustryJobsPlanningActionsProps {
  jobsWorkspaceTab: "guide" | "planning" | "operations";
  handleGeneratePlanDraft: () => void;
  previewingLedgerPlan: boolean;
  selectedLedgerProjectId: number;
  hasVisualPlanRows: boolean;
  result: IndustryAnalysis | null;
  selectedItem: BuildableItem | null;
  handlePreviewCurrentAnalysisToLedgerPlan: () => Promise<void>;
  applyingLedgerPlan: boolean;
  lastLedgerPlanPreviewPatch: unknown;
  strictBlueprintApplyBlocked: boolean;
  isLastLedgerPreviewStale: boolean;
  handleApplyLastPreviewToLedgerPlan: (bypassStrictGate?: boolean) => Promise<void>;
  handleApplyCurrentAnalysisToLedgerPlan: (bypassStrictGate?: boolean) => Promise<void>;
  addVisualTaskRow: () => void;
  addVisualJobRow: () => void;
  addVisualMaterialRow: () => void;
  addVisualBlueprintRow: () => void;
  autoFixVisualTaskBlueprintBindings: () => void;
  planDraftTasksLength: number;
  planTaskBlueprintOptionsLength: number;
  visualTaskBlueprintBindingStatsMissing: number;
  visualTaskBlueprintBindingStatsFallback: number;
  clearVisualPlanBuilder: () => void;
  replaceLedgerPlanOnApply: boolean;
  setReplaceLedgerPlanOnApply: Dispatch<SetStateAction<boolean>>;
  useVisualPlanBuilder: boolean;
  setUseVisualPlanBuilder: Dispatch<SetStateAction<boolean>>;
  strictBlueprintBindingMode: boolean;
  setStrictBlueprintBindingMode: Dispatch<SetStateAction<boolean>>;
  planDraftJobsLength: number;
  planDraftMaterialsLength: number;
  planDraftBlueprintsLength: number;
  lastLedgerPlanSummary: string;
  lastLedgerPlanPreview: IndustryPlanPreview | null;
}

export function IndustryJobsPlanningActions({
  jobsWorkspaceTab,
  handleGeneratePlanDraft,
  previewingLedgerPlan,
  selectedLedgerProjectId,
  hasVisualPlanRows,
  result,
  selectedItem,
  handlePreviewCurrentAnalysisToLedgerPlan,
  applyingLedgerPlan,
  lastLedgerPlanPreviewPatch,
  strictBlueprintApplyBlocked,
  isLastLedgerPreviewStale,
  handleApplyLastPreviewToLedgerPlan,
  handleApplyCurrentAnalysisToLedgerPlan,
  addVisualTaskRow,
  addVisualJobRow,
  addVisualMaterialRow,
  addVisualBlueprintRow,
  autoFixVisualTaskBlueprintBindings,
  planDraftTasksLength,
  planTaskBlueprintOptionsLength,
  visualTaskBlueprintBindingStatsMissing,
  visualTaskBlueprintBindingStatsFallback,
  clearVisualPlanBuilder,
  replaceLedgerPlanOnApply,
  setReplaceLedgerPlanOnApply,
  useVisualPlanBuilder,
  setUseVisualPlanBuilder,
  strictBlueprintBindingMode,
  setStrictBlueprintBindingMode,
  planDraftJobsLength,
  planDraftMaterialsLength,
  planDraftBlueprintsLength,
  lastLedgerPlanSummary,
  lastLedgerPlanPreview,
}: IndustryJobsPlanningActionsProps) {
  const { t } = useI18n();

  if (jobsWorkspaceTab !== "planning") {
    return null;
  }

  return (
    <div className="mt-2 flex flex-wrap items-center gap-2">
      <button
        type="button"
        onClick={handleGeneratePlanDraft}
        disabled={!result || !selectedItem}
        className="px-3 py-1.5 rounded-sm text-xs font-semibold border border-eve-border text-eve-dim hover:text-eve-accent hover:border-eve-accent/40 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {t("industryLedgerSeedBuilder")}
      </button>
      <button
        type="button"
        onClick={() => { void handlePreviewCurrentAnalysisToLedgerPlan(); }}
        disabled={
          previewingLedgerPlan ||
          selectedLedgerProjectId <= 0 ||
          (!hasVisualPlanRows && (!result || !selectedItem))
        }
        className="px-3 py-1.5 rounded-sm text-xs font-semibold border border-eve-border text-eve-dim hover:text-eve-accent hover:border-eve-accent/40 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {previewingLedgerPlan ? t("industryLedgerPreviewing") : t("industryLedgerPreviewPlan")}
      </button>
      <button
        type="button"
        onClick={() => { void handleApplyLastPreviewToLedgerPlan(); }}
        disabled={
          applyingLedgerPlan ||
          selectedLedgerProjectId <= 0 ||
          !lastLedgerPlanPreviewPatch ||
          strictBlueprintApplyBlocked
        }
        className={`px-3 py-1.5 rounded-sm text-xs font-semibold border disabled:opacity-50 disabled:cursor-not-allowed ${
          isLastLedgerPreviewStale
            ? "border-yellow-500/60 text-yellow-300 hover:bg-yellow-500/10"
            : "border-eve-accent/50 text-eve-accent hover:bg-eve-accent/10"
        }`}
      >
        {applyingLedgerPlan
          ? t("industryLedgerApplying")
          : isLastLedgerPreviewStale
            ? t("industryLedgerApplyPreviewStale")
            : t("industryLedgerApplyPreview")}
      </button>
      <button
        type="button"
        onClick={() => { void handleApplyCurrentAnalysisToLedgerPlan(); }}
        disabled={
          applyingLedgerPlan ||
          selectedLedgerProjectId <= 0 ||
          (!hasVisualPlanRows && (!result || !selectedItem)) ||
          strictBlueprintApplyBlocked
        }
        className="px-3 py-1.5 rounded-sm text-xs font-semibold border border-eve-accent/40 text-eve-accent hover:bg-eve-accent/10 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {applyingLedgerPlan ? t("industryLedgerApplying") : t("industryLedgerApplyCurrentPlan")}
      </button>
      <button
        type="button"
        onClick={addVisualTaskRow}
        className="px-2 py-1 rounded-sm text-[11px] font-semibold border border-eve-border text-eve-dim hover:text-eve-accent hover:border-eve-accent/40"
      >
        {t("industryLedgerAddTask")}
      </button>
      <button
        type="button"
        onClick={addVisualJobRow}
        className="px-2 py-1 rounded-sm text-[11px] font-semibold border border-eve-border text-eve-dim hover:text-eve-accent hover:border-eve-accent/40"
      >
        {t("industryLedgerAddJob")}
      </button>
      <button
        type="button"
        onClick={addVisualMaterialRow}
        className="px-2 py-1 rounded-sm text-[11px] font-semibold border border-eve-border text-eve-dim hover:text-eve-accent hover:border-eve-accent/40"
      >
        {t("industryLedgerAddMaterial")}
      </button>
      <button
        type="button"
        onClick={addVisualBlueprintRow}
        className="px-2 py-1 rounded-sm text-[11px] font-semibold border border-eve-border text-eve-dim hover:text-eve-accent hover:border-eve-accent/40"
      >
        {t("industryLedgerAddBlueprint")}
      </button>
      <button
        type="button"
        onClick={autoFixVisualTaskBlueprintBindings}
        disabled={planDraftTasksLength === 0 || planTaskBlueprintOptionsLength === 0 || (visualTaskBlueprintBindingStatsMissing === 0 && visualTaskBlueprintBindingStatsFallback === 0)}
        className="px-2 py-1 rounded-sm text-[11px] font-semibold border border-amber-500/40 text-amber-300 hover:bg-amber-500/10 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        Auto-fix BP Bindings
      </button>
      <button
        type="button"
        onClick={clearVisualPlanBuilder}
        disabled={!hasVisualPlanRows}
        className="px-2 py-1 rounded-sm text-[11px] font-semibold border border-eve-border text-eve-dim hover:text-eve-accent hover:border-eve-accent/40 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {t("industryLedgerClearBuilder")}
      </button>
      <label className="flex items-center gap-1 text-[11px] text-eve-dim select-none">
        <input
          type="checkbox"
          checked={replaceLedgerPlanOnApply}
          onChange={(e) => setReplaceLedgerPlanOnApply(e.target.checked)}
          className="accent-eve-accent"
        />
        {t("industryLedgerReplacePlanRows")}
      </label>
      <label className="flex items-center gap-1 text-[11px] text-eve-dim select-none">
        <input
          type="checkbox"
          checked={useVisualPlanBuilder}
          onChange={(e) => setUseVisualPlanBuilder(e.target.checked)}
          className="accent-eve-accent"
        />
        {t("industryLedgerUseVisualBuilder")}
      </label>
      <label className="flex items-center gap-1 text-[11px] text-eve-dim select-none">
        <input
          type="checkbox"
          checked={strictBlueprintBindingMode}
          onChange={(e) => setStrictBlueprintBindingMode(e.target.checked)}
          className="accent-eve-accent"
        />
        Strict BP gate
      </label>
      <span className="text-[11px] text-eve-dim">
        {t("industryLedgerRowsPrefix")}: T{planDraftTasksLength} J{planDraftJobsLength} M{planDraftMaterialsLength} B{planDraftBlueprintsLength}
      </span>
      {strictBlueprintApplyBlocked && (
        <>
          <span className="text-[11px] text-red-300">
            Strict BP gate active: fix {visualTaskBlueprintBindingStatsMissing} missing bindings before Apply
          </span>
          {lastLedgerPlanPreviewPatch && (
            <button
              type="button"
              onClick={() => {
                if (!window.confirm("Apply preview anyway and bypass strict BP gate once?")) return;
                void handleApplyLastPreviewToLedgerPlan(true);
              }}
              disabled={applyingLedgerPlan}
              className="px-2 py-1 rounded-sm text-[11px] font-semibold border border-red-500/50 text-red-300 hover:bg-red-500/10 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Apply Preview Anyway
            </button>
          )}
          <button
            type="button"
            onClick={() => {
              if (!window.confirm("Apply current plan anyway and bypass strict BP gate once?")) return;
              void handleApplyCurrentAnalysisToLedgerPlan(true);
            }}
            disabled={applyingLedgerPlan || (!hasVisualPlanRows && (!result || !selectedItem))}
            className="px-2 py-1 rounded-sm text-[11px] font-semibold border border-red-500/50 text-red-300 hover:bg-red-500/10 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Apply Current Anyway
          </button>
        </>
      )}
      {lastLedgerPlanSummary && (
        <span className="text-[11px] text-eve-dim">{lastLedgerPlanSummary}</span>
      )}
      {lastLedgerPlanPreview && (
        <span className={`text-[11px] ${isLastLedgerPreviewStale ? "text-yellow-400" : "text-emerald-400"}`}>
          {isLastLedgerPreviewStale
            ? t("industryLedgerPreviewStale")
            : t("industryLedgerPreviewInSync")}
        </span>
      )}
    </div>
  );
}
