import { useEffect, useMemo, useState } from "react";
import { useI18n } from "@/lib/i18n";
import type { IndustryPlanPreview } from "@/lib/types";
import { formatUtcShort } from "./industryHelpers";

interface IndustryPlanPreviewPanelProps {
  jobsWorkspaceTab: "guide" | "planning" | "operations";
  lastLedgerPlanPreview: IndustryPlanPreview | null;
  isLastLedgerPreviewStale: boolean;
  clearPlanPreview: () => void;
}

export function IndustryPlanPreviewPanel({
  jobsWorkspaceTab,
  lastLedgerPlanPreview,
  isLastLedgerPreviewStale,
  clearPlanPreview,
}: IndustryPlanPreviewPanelProps) {
  const { t } = useI18n();
  const [rowsPerPage, setRowsPerPage] = useState(40);
  const [page, setPage] = useState(1);
  const previewJobs = lastLedgerPlanPreview?.jobs ?? [];
  const totalPages = Math.max(1, Math.ceil(previewJobs.length / Math.max(1, rowsPerPage)));

  useEffect(() => {
    setPage((prev) => Math.min(Math.max(1, prev), totalPages));
  }, [totalPages]);

  const visibleJobs = useMemo(() => {
    const safeSize = Math.max(1, rowsPerPage);
    const start = (page - 1) * safeSize;
    return previewJobs.slice(start, start + safeSize).map((job, idx) => ({
      job,
      rowKey: `preview-job-${start + idx}-${job.task_id ?? 0}-${job.activity}-${job.runs ?? 0}-${job.external_job_id ?? 0}-${job.started_at ?? ""}-${job.finished_at ?? ""}`,
    }));
  }, [previewJobs, page, rowsPerPage]);

  if (jobsWorkspaceTab !== "planning" || !lastLedgerPlanPreview) {
    return null;
  }

  return (
    <div className="mt-2 border border-eve-border/40 rounded-sm p-2 bg-eve-dark/20">
      <div className="flex items-center justify-between gap-2 mb-1">
        <div className="text-[10px] uppercase tracking-wider text-eve-dim">{t("industryLedgerPlanPreview")}</div>
        <button
          type="button"
          onClick={clearPlanPreview}
          className="px-1.5 py-0.5 text-[10px] border border-eve-border rounded-sm text-eve-dim hover:text-eve-accent hover:border-eve-accent/40"
        >
          {t("clear")}
        </button>
      </div>
      <div className="text-[10px] text-eve-dim mb-1">
        {t("industryLedgerApplyPreviewHint")}
      </div>
      {isLastLedgerPreviewStale && (
        <div className="text-[11px] text-yellow-400 mb-1">
          {t("industryLedgerPreviewChangedHint")}
        </div>
      )}
      <div className="text-[11px] text-eve-dim">
        tasks:{lastLedgerPlanPreview.summary.tasks_inserted} jobs:{lastLedgerPlanPreview.summary.jobs_inserted} mats:{lastLedgerPlanPreview.summary.materials_upserted} bp:{lastLedgerPlanPreview.summary.blueprints_upserted}
        {lastLedgerPlanPreview.summary.scheduler_applied && (
          <> split:{lastLedgerPlanPreview.summary.jobs_split_from ?? 0}{"->"}{lastLedgerPlanPreview.summary.jobs_planned_total ?? lastLedgerPlanPreview.summary.jobs_inserted}</>
        )}
      </div>
      {lastLedgerPlanPreview.warnings.length > 0 && (
        <div className="mt-1 text-[11px] text-yellow-400">
          {lastLedgerPlanPreview.warnings.join(" | ")}
        </div>
      )}
      <div className="mt-1 inline-flex items-center gap-1 text-[10px] text-eve-dim">
        <span>rows/page</span>
        <select
          value={rowsPerPage}
          onChange={(e) => setRowsPerPage(Math.max(10, Math.min(400, Number(e.target.value) || 40)))}
          className="px-1 py-0.5 bg-eve-input border border-eve-border rounded-sm text-[10px] text-eve-text"
        >
          <option value={20}>20</option>
          <option value={40}>40</option>
          <option value={80}>80</option>
          <option value={160}>160</option>
        </select>
        <span>{page}/{totalPages}</span>
        <button
          type="button"
          onClick={() => setPage((prev) => Math.max(1, prev - 1))}
          disabled={page <= 1}
          className="px-1 border border-eve-border rounded-sm disabled:opacity-40"
        >
          {"<"}
        </button>
        <button
          type="button"
          onClick={() => setPage((prev) => Math.min(totalPages, prev + 1))}
          disabled={page >= totalPages}
          className="px-1 border border-eve-border rounded-sm disabled:opacity-40"
        >
          {">"}
        </button>
      </div>
      <div className="mt-2 border border-eve-border rounded-sm max-h-[180px] overflow-auto">
        <table className="w-full text-[11px]">
          <thead className="sticky top-0 bg-eve-dark z-10">
            <tr className="text-eve-dim uppercase tracking-wider border-b border-eve-border/60">
              <th className="px-1.5 py-1 text-left">{t("industryLedgerTask")}</th>
              <th className="px-1.5 py-1 text-left">{t("industryLedgerActivity")}</th>
              <th className="px-1.5 py-1 text-right">{t("industryLedgerRuns")}</th>
              <th className="px-1.5 py-1 text-left">{t("industryLedgerStart")}</th>
              <th className="px-1.5 py-1 text-left">{t("industryLedgerFinish")}</th>
            </tr>
          </thead>
          <tbody>
            {visibleJobs.map(({ job, rowKey }) => (
              <tr key={rowKey} className="border-b border-eve-border/30">
                <td className="px-1.5 py-1 text-eve-dim font-mono">{job.task_id ?? 0}</td>
                <td className="px-1.5 py-1 text-eve-text">{job.activity}</td>
                <td className="px-1.5 py-1 text-right text-eve-accent font-mono">{job.runs ?? 0}</td>
                <td className="px-1.5 py-1 text-eve-dim whitespace-nowrap">{formatUtcShort(job.started_at || "")}</td>
                <td className="px-1.5 py-1 text-eve-dim whitespace-nowrap">{formatUtcShort(job.finished_at || "")}</td>
              </tr>
            ))}
            {lastLedgerPlanPreview.jobs.length === 0 && (
              <tr>
                <td colSpan={5} className="px-2 py-2 text-center text-eve-dim">{t("industryLedgerNoPreviewJobs")}</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
