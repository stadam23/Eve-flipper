import { useI18n } from "@/lib/i18n";
import {
  formatUtcShort,
  industryPlannerWarningSourceClass,
  type IndustryPlannerWarningEvent,
  type IndustryPlannerWarningSource,
} from "./industryHelpers";

interface Props {
  warnings: IndustryPlannerWarningEvent[];
  onClear: () => void;
  sourceLabel: (source: IndustryPlannerWarningSource) => string;
}

export function IndustryPlannerWarningLog({ warnings, onClear, sourceLabel }: Props) {
  const { t } = useI18n();

  return (
    <div className="mt-2 border border-yellow-500/30 rounded-sm p-2 bg-yellow-500/5">
      <div className="flex items-center justify-between gap-2 mb-1">
        <div className="text-[10px] uppercase tracking-wider text-yellow-300">
          {t("industryLedgerWarningLog")} ({warnings.length})
        </div>
        <button
          type="button"
          onClick={onClear}
          disabled={warnings.length === 0}
          className="px-1.5 py-0.5 text-[10px] border border-eve-border rounded-sm text-eve-dim hover:text-eve-accent hover:border-eve-accent/40 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {t("industryLedgerClearWarnings")}
        </button>
      </div>
      {warnings.length === 0 ? (
        <div className="text-[11px] text-eve-dim">{t("industryLedgerWarningLogEmpty")}</div>
      ) : (
        <div className="max-h-[150px] overflow-auto space-y-1">
          {warnings.map((warning) => (
            <div
              key={warning.id}
              className="grid grid-cols-[auto_minmax(0,1fr)_auto] gap-2 items-start border border-eve-border/40 rounded-sm px-1.5 py-1 bg-eve-dark/30"
            >
              <span className={`px-1 py-0.5 rounded-sm border text-[10px] uppercase tracking-wide ${industryPlannerWarningSourceClass(warning.source)}`}>
                {sourceLabel(warning.source)}
              </span>
              <span className="text-[11px] text-eve-text break-words">{warning.message}</span>
              <span className="text-[10px] text-eve-dim whitespace-nowrap">{formatUtcShort(warning.created_at)}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
