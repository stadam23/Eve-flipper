import { useI18n } from "@/lib/i18n";
import type { Dispatch, SetStateAction } from "react";
import type { IndustryLedger, IndustryProject, IndustryProjectSnapshot } from "@/lib/types";

interface IndustryJobsProjectHeaderProps {
  newLedgerProjectName: string;
  setNewLedgerProjectName: Dispatch<SetStateAction<string>>;
  newLedgerProjectStrategy: "conservative" | "balanced" | "aggressive";
  setNewLedgerProjectStrategy: Dispatch<SetStateAction<"conservative" | "balanced" | "aggressive">>;
  creatingLedgerProject: boolean;
  handleCreateLedgerProject: () => Promise<void>;
  refreshLedgerProjects: (preferredProjectId?: number) => Promise<void>;
  selectedLedgerProjectId: number;
  setSelectedLedgerProjectId: Dispatch<SetStateAction<number>>;
  ledgerProjects: IndustryProject[];
  ledgerLoading: boolean;
  ledgerData: IndustryLedger | null;
  ledgerSnapshotLoading: boolean;
  ledgerSnapshot: IndustryProjectSnapshot | null;
  handleLoadLedgerSnapshotToBuilder: () => void;
}

export function IndustryJobsProjectHeader({
  newLedgerProjectName,
  setNewLedgerProjectName,
  newLedgerProjectStrategy,
  setNewLedgerProjectStrategy,
  creatingLedgerProject,
  handleCreateLedgerProject,
  refreshLedgerProjects,
  selectedLedgerProjectId,
  setSelectedLedgerProjectId,
  ledgerProjects,
  ledgerLoading,
  ledgerData,
  ledgerSnapshotLoading,
  ledgerSnapshot,
  handleLoadLedgerSnapshotToBuilder,
}: IndustryJobsProjectHeaderProps) {
  const { t } = useI18n();

  return (
    <>
      <div className="grid grid-cols-1 md:grid-cols-[1fr_180px_auto_auto] gap-2">
        <input
          type="text"
          value={newLedgerProjectName}
          onChange={(e) => setNewLedgerProjectName(e.target.value)}
          placeholder={t("industryLedgerNewProjectPlaceholder")}
          className="w-full px-2 py-1.5 bg-eve-input border border-eve-border rounded-sm text-xs text-eve-text
                   focus:outline-none focus:border-eve-accent focus:ring-1 focus:ring-eve-accent/30 transition-colors"
        />
        <select
          value={newLedgerProjectStrategy}
          onChange={(e) => setNewLedgerProjectStrategy(e.target.value as "conservative" | "balanced" | "aggressive")}
          className="px-2 py-1.5 bg-eve-input border border-eve-border rounded-sm text-xs text-eve-text focus:outline-none focus:border-eve-accent"
        >
          <option value="conservative">{t("industryLedgerStrategyConservative")}</option>
          <option value="balanced">{t("industryLedgerStrategyBalanced")}</option>
          <option value="aggressive">{t("industryLedgerStrategyAggressive")}</option>
        </select>
        <button
          type="button"
          onClick={() => { void handleCreateLedgerProject(); }}
          disabled={creatingLedgerProject || !newLedgerProjectName.trim()}
          className="px-3 py-1.5 rounded-sm text-xs font-semibold border border-eve-border text-eve-accent hover:border-eve-accent/40 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {creatingLedgerProject ? t("industryLedgerCreatingProject") : t("industryLedgerCreateProject")}
        </button>
        <button
          type="button"
          onClick={() => { void refreshLedgerProjects(selectedLedgerProjectId || undefined); }}
          className="px-3 py-1.5 rounded-sm text-xs font-semibold border border-eve-border text-eve-dim hover:text-eve-accent hover:border-eve-accent/40"
        >
          {t("refresh")}
        </button>
      </div>

      <div className="mt-2 grid grid-cols-1 md:grid-cols-[1fr_auto] gap-2 items-center">
        <select
          value={selectedLedgerProjectId}
          onChange={(e) => setSelectedLedgerProjectId(Number(e.target.value))}
          className="w-full px-2 py-1.5 bg-eve-input border border-eve-border rounded-sm text-xs text-eve-text focus:outline-none focus:border-eve-accent"
        >
          {ledgerProjects.length === 0 && <option value={0}>{t("industryLedgerNoProjects")}</option>}
          {ledgerProjects.map((project) => (
            <option key={project.id} value={project.id}>
              {project.name} [{project.status}]
            </option>
          ))}
        </select>
        <div className="text-[10px] text-eve-dim">
          {ledgerLoading
            ? t("industryLedgerLoading")
            : ledgerData
              ? `${ledgerData.entries.length} ${t("industryLedgerRows")}`
              : t("industryLedgerNoData")}
        </div>
      </div>

      <div className="mt-1 flex flex-wrap items-center gap-2 text-[10px] text-eve-dim">
        <span>
          {ledgerSnapshotLoading
            ? t("industryLedgerSnapshotLoading")
            : ledgerSnapshot
              ? `snapshot T${ledgerSnapshot.tasks.length} J${ledgerSnapshot.jobs.length} M${ledgerSnapshot.materials.length} B${ledgerSnapshot.blueprints.length}`
              : t("industryLedgerNoSnapshotShort")}
        </span>
        <button
          type="button"
          onClick={handleLoadLedgerSnapshotToBuilder}
          disabled={!ledgerSnapshot}
          className="px-2 py-0.5 rounded-sm border border-eve-border text-eve-dim hover:text-eve-accent hover:border-eve-accent/40 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {t("industryLedgerLoadSnapshotToBuilder")}
        </button>
      </div>
    </>
  );
}
