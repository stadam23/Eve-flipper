import type { Dispatch, SetStateAction } from "react";
import { useI18n } from "@/lib/i18n";
import { formatISK } from "@/lib/format";
import { useGlobalToast } from "../Toast";
import type { FlatMaterial, IndustryAnalysis } from "@/lib/types";
import { formatDuration } from "./industryHelpers";
import { IndustryMaterialTree } from "./IndustryMaterialTree";
import { IndustryShoppingList } from "./IndustryShoppingList";
import { IndustrySummaryCard } from "./IndustrySummaryCard";

interface IndustryAnalysisResultsPanelProps {
  result: IndustryAnalysis;
  viewMode: "tree" | "shopping";
  setViewMode: Dispatch<SetStateAction<"tree" | "shopping">>;
  salesTaxPercent: number;
  brokerFee: number;
  onOpenExecutionPlan: (material: FlatMaterial) => void;
}

export function IndustryAnalysisResultsPanel({
  result,
  viewMode,
  setViewMode,
  salesTaxPercent,
  brokerFee,
  onOpenExecutionPlan,
}: IndustryAnalysisResultsPanelProps) {
  const { t } = useI18n();
  const { addToast } = useGlobalToast();

  return (
    <div className="flex-1 min-h-0 m-2 mt-0 flex flex-col">
      <div className="shrink-0 grid grid-cols-2 md:grid-cols-4 gap-2 mb-2">
        <IndustrySummaryCard
          label={t("industryMarketPrice")}
          value={formatISK(result.market_buy_price ?? 0)}
          subtext={`${(result.total_quantity ?? 0).toLocaleString()} ${t("industryUnits")}`}
          color="text-eve-dim"
        />
        <IndustrySummaryCard
          label={t("industryBuildCost")}
          value={formatISK(result.optimal_build_cost ?? 0)}
          subtext={result.blueprint_cost_included > 0
            ? `${t("industryJobCost")}: ${formatISK(result.total_job_cost ?? 0)} · ${t("industryBPCostIncluded")}: ${formatISK(result.blueprint_cost_included)}`
            : `${t("industryJobCost")}: ${formatISK(result.total_job_cost ?? 0)}`}
          color="text-eve-accent"
        />
        <IndustrySummaryCard
          label={t("industrySavings")}
          value={formatISK(result.savings ?? 0)}
          subtext={`${(result.savings_percent ?? 0).toFixed(1)}%`}
          color={(result.savings ?? 0) > 0 ? "text-green-400" : "text-red-400"}
        />
        <IndustrySummaryCard
          label={t("industryProfit")}
          value={formatISK(result.profit ?? 0)}
          subtext={`${(result.profit_percent ?? 0).toFixed(1)}% ROI`}
          color={(result.profit ?? 0) > 0 ? "text-green-400" : "text-red-400"}
        />
      </div>

      <div className="shrink-0 grid grid-cols-2 md:grid-cols-4 gap-2 mb-2">
        <IndustrySummaryCard
          label={t("industryISKPerHour")}
          value={formatISK(result.isk_per_hour ?? 0)}
          color={(result.isk_per_hour ?? 0) > 0 ? "text-yellow-400" : "text-red-400"}
        />
        <IndustrySummaryCard
          label={t("industryMfgTime")}
          value={formatDuration(result.manufacturing_time ?? 0)}
          color="text-eve-dim"
        />
        <IndustrySummaryCard
          label={t("industrySellRevenue")}
          value={formatISK(result.sell_revenue ?? 0)}
          subtext={`-${salesTaxPercent}% tax -${brokerFee}% broker`}
          color="text-eve-dim"
        />
        <IndustrySummaryCard
          label={t("industryJobCost")}
          value={formatISK(result.total_job_cost ?? 0)}
          subtext={`SCI: ${((result.system_cost_index ?? 0) * 100).toFixed(2)}%`}
          color="text-eve-dim"
        />
      </div>

      <div className="shrink-0 flex items-center gap-2 mb-2 flex-wrap">
        <button
          onClick={() => setViewMode("tree")}
          className={`px-3 py-1 text-xs rounded-sm transition-colors ${
            viewMode === "tree"
              ? "bg-eve-accent/20 text-eve-accent border border-eve-accent/30"
              : "text-eve-dim hover:text-eve-text border border-eve-border"
          }`}
        >
          {t("industryTreeView")}
        </button>
        <button
          onClick={() => setViewMode("shopping")}
          className={`px-3 py-1 text-xs rounded-sm transition-colors ${
            viewMode === "shopping"
              ? "bg-eve-accent/20 text-eve-accent border border-eve-accent/30"
              : "text-eve-dim hover:text-eve-text border border-eve-border"
          }`}
        >
          {t("industryShoppingList")}
        </button>
        {viewMode === "shopping" && result.flat_materials.length > 0 && (
          <>
            <button
              onClick={() => {
                const header = "Item\tQuantity\tUnit Price\tTotal\tVolume (m³)";
                const rows = result.flat_materials.map(
                  (m) => `${m.type_name}\t${m.quantity}\t${m.unit_price}\t${m.total_price}\t${m.volume}`
                );
                navigator.clipboard.writeText([header, ...rows].join("\n"));
                addToast(t("copied"), "success", 2000);
              }}
              className="px-3 py-1 text-xs rounded-sm text-eve-dim hover:text-eve-accent border border-eve-border hover:border-eve-accent/30 transition-colors"
            >
              {t("industryExportClipboard")}
            </button>
            <button
              onClick={() => {
                const header = "Item,Quantity,Unit Price,Total,Volume (m³)";
                const rows = result.flat_materials.map(
                  (m) => `"${(m.type_name || "").replace(/"/g, "\"\"")}",${m.quantity},${m.unit_price},${m.total_price},${m.volume}`
                );
                const csv = "\uFEFF" + [header, ...rows].join("\n");
                const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
                const url = URL.createObjectURL(blob);
                const a = document.createElement("a");
                a.href = url;
                a.download = `industry-shopping-list-${new Date().toISOString().slice(0, 10)}.csv`;
                a.click();
                URL.revokeObjectURL(url);
                addToast(t("industryExportCSV"), "success", 2000);
              }}
              className="px-3 py-1 text-xs rounded-sm text-eve-dim hover:text-eve-accent border border-eve-border hover:border-eve-accent/30 transition-colors"
            >
              {t("industryExportCSV")}
            </button>
          </>
        )}
      </div>

      <div className="flex-1 min-h-0 overflow-auto border border-eve-border rounded-sm bg-eve-panel">
        {viewMode === "tree" ? (
          <IndustryMaterialTree node={result.material_tree} />
        ) : (
          <IndustryShoppingList
            materials={result.flat_materials}
            regionId={result.region_id ?? 0}
            onOpenExecutionPlan={onOpenExecutionPlan}
          />
        )}
      </div>
    </div>
  );
}
