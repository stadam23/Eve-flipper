import { useI18n } from "@/lib/i18n";

export type IndustryJobsWorkspaceTab = "guide" | "planning" | "operations";

interface Props {
  activeTab: IndustryJobsWorkspaceTab;
  onChange: (tab: IndustryJobsWorkspaceTab) => void;
  warningsCount: number;
  missingBindings: number;
  activeJobs: number;
}

const tabTone: Record<IndustryJobsWorkspaceTab, string> = {
  guide: "border-emerald-500/40",
  planning: "border-cyan-500/40",
  operations: "border-fuchsia-500/40",
};

export function IndustryJobsWorkspaceNav({
  activeTab,
  onChange,
  warningsCount,
  missingBindings,
  activeJobs,
}: Props) {
  const { t } = useI18n();

  const tabs: Array<{ id: IndustryJobsWorkspaceTab; label: string; hint: string }> = [
    {
      id: "guide",
      label: t("industryJobsWorkspaceGuide"),
      hint: t("industryJobsWorkspaceGuideHint"),
    },
    {
      id: "planning",
      label: t("industryJobsWorkspacePlanning"),
      hint: t("industryJobsWorkspacePlanningHint"),
    },
    {
      id: "operations",
      label: t("industryJobsWorkspaceOps"),
      hint: t("industryJobsWorkspaceOpsHint"),
    },
  ];

  return (
    <div className="mt-2 border border-eve-border/40 rounded-sm p-2 bg-eve-dark/20">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="inline-flex rounded-sm border border-eve-border overflow-hidden">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              type="button"
              onClick={() => onChange(tab.id)}
              className={`px-3 py-1.5 text-xs font-semibold uppercase tracking-wide transition-colors ${
                activeTab === tab.id
                  ? `bg-eve-accent/20 text-eve-accent ${tabTone[tab.id]}`
                  : "bg-eve-panel text-eve-dim hover:text-eve-text"
              }`}
              title={tab.hint}
            >
              {tab.label}
            </button>
          ))}
        </div>
        <div className="flex flex-wrap items-center gap-1.5 text-[10px]">
          <span className="px-1.5 py-0.5 rounded-sm border border-yellow-500/40 text-yellow-300 bg-yellow-500/10">
            warnings:{warningsCount}
          </span>
          <span className="px-1.5 py-0.5 rounded-sm border border-red-500/40 text-red-300 bg-red-500/10">
            missing bp:{missingBindings}
          </span>
          <span className="px-1.5 py-0.5 rounded-sm border border-blue-500/40 text-blue-300 bg-blue-500/10">
            active jobs:{activeJobs}
          </span>
        </div>
      </div>
      <div className="mt-1 text-[10px] text-eve-dim">{t("industryJobsWorkspaceHint")}</div>
    </div>
  );
}
