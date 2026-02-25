import { useMemo } from "react";
import { useI18n } from "@/lib/i18n";

interface Props {
  selectedProjectId: number;
  hasPlanSeedSource: boolean;
  hasVisualPlanRows: boolean;
  hasPreview: boolean;
  previewStale: boolean;
  strictBlueprintApplyBlocked: boolean;
  missingBindings: number;
  previewing: boolean;
  applying: boolean;
  lastPreviewPatchExists: boolean;
  onGenerateDraft: () => void;
  onPreview: () => void;
  onApplyPreview: () => void;
  onApplyCurrent: () => void;
  onOpenPlanner: () => void;
  onOpenOperations: () => void;
}

function stepClass(done: boolean): string {
  return done
    ? "border-emerald-500/40 bg-emerald-500/10 text-emerald-300"
    : "border-eve-border bg-eve-dark/30 text-eve-dim";
}

export function IndustryJobsGuidePanel({
  selectedProjectId,
  hasPlanSeedSource,
  hasVisualPlanRows,
  hasPreview,
  previewStale,
  strictBlueprintApplyBlocked,
  missingBindings,
  previewing,
  applying,
  lastPreviewPatchExists,
  onGenerateDraft,
  onPreview,
  onApplyPreview,
  onApplyCurrent,
  onOpenPlanner,
  onOpenOperations,
}: Props) {
  const { t } = useI18n();

  const stepState = useMemo(
    () => ({
      project: selectedProjectId > 0,
      draft: hasVisualPlanRows || hasPlanSeedSource,
      preview: hasPreview,
      apply: hasPreview && !previewStale,
    }),
    [selectedProjectId, hasVisualPlanRows, hasPlanSeedSource, hasPreview, previewStale]
  );

  return (
    <div className="mt-2 border border-emerald-500/30 rounded-sm p-2 bg-emerald-500/5">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <div className="text-[10px] uppercase tracking-wider text-emerald-300">{t("industryJobsGuideTitle")}</div>
          <div className="text-[11px] text-eve-dim">{t("industryJobsGuideIntro")}</div>
        </div>
        <div className="inline-flex gap-1">
          <button
            type="button"
            onClick={onOpenPlanner}
            className="px-2 py-1 rounded-sm text-[11px] font-semibold border border-cyan-500/40 text-cyan-300 hover:bg-cyan-500/10"
          >
            {t("industryJobsGuideOpenPlanner")}
          </button>
          <button
            type="button"
            onClick={onOpenOperations}
            className="px-2 py-1 rounded-sm text-[11px] font-semibold border border-fuchsia-500/40 text-fuchsia-300 hover:bg-fuchsia-500/10"
          >
            {t("industryJobsGuideOpenOps")}
          </button>
        </div>
      </div>

      <div className="mt-2 grid grid-cols-2 lg:grid-cols-4 gap-1.5 text-[11px]">
        <div className={`px-2 py-1 rounded-sm border ${stepClass(stepState.project)}`}>
          1. {t("industryJobsGuideStepProject")}
        </div>
        <div className={`px-2 py-1 rounded-sm border ${stepClass(stepState.draft)}`}>
          2. {t("industryJobsGuideStepPlan")}
        </div>
        <div className={`px-2 py-1 rounded-sm border ${stepClass(stepState.preview)}`}>
          3. {t("industryJobsGuideStepPreview")}
        </div>
        <div className={`px-2 py-1 rounded-sm border ${stepClass(stepState.apply)}`}>
          4. {t("industryJobsGuideStepApply")}
        </div>
      </div>

      <div className="mt-2 flex flex-wrap items-center gap-2">
        <button
          type="button"
          onClick={onGenerateDraft}
          disabled={!hasPlanSeedSource}
          className="px-3 py-1.5 rounded-sm text-xs font-semibold border border-eve-border text-eve-dim hover:text-eve-accent hover:border-eve-accent/40 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {t("industryJobsGuideCreatePlan")}
        </button>
        <button
          type="button"
          onClick={onPreview}
          disabled={previewing || selectedProjectId <= 0 || (!hasVisualPlanRows && !hasPlanSeedSource)}
          className="px-3 py-1.5 rounded-sm text-xs font-semibold border border-cyan-500/40 text-cyan-300 hover:bg-cyan-500/10 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {previewing ? t("industryLedgerPreviewing") : t("industryJobsGuideRunPreview")}
        </button>
        <button
          type="button"
          onClick={onApplyPreview}
          disabled={applying || !lastPreviewPatchExists || strictBlueprintApplyBlocked}
          className="px-3 py-1.5 rounded-sm text-xs font-semibold border border-emerald-500/40 text-emerald-300 hover:bg-emerald-500/10 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {t("industryJobsGuideApplyPreview")}
        </button>
        <button
          type="button"
          onClick={onApplyCurrent}
          disabled={applying || selectedProjectId <= 0 || (!hasVisualPlanRows && !hasPlanSeedSource) || strictBlueprintApplyBlocked}
          className="px-3 py-1.5 rounded-sm text-xs font-semibold border border-eve-accent/40 text-eve-accent hover:bg-eve-accent/10 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {t("industryJobsGuideApplyCurrent")}
        </button>
      </div>

      {strictBlueprintApplyBlocked && (
        <div className="mt-2 text-[11px] text-red-300">
          {t("industryJobsGuideFixBindings", { count: missingBindings })}
        </div>
      )}
      {hasPreview && previewStale && (
        <div className="mt-1 text-[11px] text-yellow-300">{t("industryJobsGuidePreviewStale")}</div>
      )}
    </div>
  );
}
