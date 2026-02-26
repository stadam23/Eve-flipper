import { useEffect, useMemo, useState, type Dispatch, type SetStateAction } from "react";
import { useI18n } from "@/lib/i18n";
import type { IndustryMaterialDiff } from "@/lib/types";
import type { IndustryJobsWorkspaceTab } from "./IndustryJobsWorkspaceNav";

interface IndustryMaterialDiffPanelProps {
  jobsWorkspaceTab: IndustryJobsWorkspaceTab;
  materialRows: IndustryMaterialDiff[];
  rebalanceInventoryScope: "single" | "all";
  setRebalanceInventoryScope: Dispatch<SetStateAction<"single" | "all">>;
  rebalanceLookbackDays: number;
  setRebalanceLookbackDays: Dispatch<SetStateAction<number>>;
  rebalanceStrategy: "preserve" | "buy" | "build";
  setRebalanceStrategy: Dispatch<SetStateAction<"preserve" | "buy" | "build">>;
  rebalanceWarehouseScope: "global" | "location_first" | "strict_location";
  setRebalanceWarehouseScope: Dispatch<SetStateAction<"global" | "location_first" | "strict_location">>;
  blueprintSyncDefaultBPCRuns: number;
  setBlueprintSyncDefaultBPCRuns: Dispatch<SetStateAction<number>>;
  syncingLedgerBlueprintPool: boolean;
  handleSyncLedgerBlueprintPoolFromAssets: () => Promise<void>;
  rebalanceUseSelectedStation: boolean;
  setRebalanceUseSelectedStation: Dispatch<SetStateAction<boolean>>;
  handleRebalanceLedgerMaterialsFromInventory: () => Promise<void>;
  rebalancingLedgerMaterials: boolean;
}

export function IndustryMaterialDiffPanel({
  jobsWorkspaceTab,
  materialRows,
  rebalanceInventoryScope,
  setRebalanceInventoryScope,
  rebalanceLookbackDays,
  setRebalanceLookbackDays,
  rebalanceStrategy,
  setRebalanceStrategy,
  rebalanceWarehouseScope,
  setRebalanceWarehouseScope,
  blueprintSyncDefaultBPCRuns,
  setBlueprintSyncDefaultBPCRuns,
  syncingLedgerBlueprintPool,
  handleSyncLedgerBlueprintPoolFromAssets,
  rebalanceUseSelectedStation,
  setRebalanceUseSelectedStation,
  handleRebalanceLedgerMaterialsFromInventory,
  rebalancingLedgerMaterials,
}: IndustryMaterialDiffPanelProps) {
  const { t } = useI18n();
  const [materialFilterMode, setMaterialFilterMode] = useState<"all" | "stock" | "buy" | "build" | "missing">("all");
  const [materialSearch, setMaterialSearch] = useState("");
  const [materialExportStatus, setMaterialExportStatus] = useState("");
  const [materialRowsPerPage, setMaterialRowsPerPage] = useState(120);
  const [materialPage, setMaterialPage] = useState(1);

  const resolveMaterialAction = (row: IndustryMaterialDiff): "stock" | "buy" | "build" | "buy_build" | "missing" => {
    const missingQty = row.missing_qty || 0;
    const buyQty = row.buy_qty || 0;
    const buildQty = row.build_qty || 0;
    if (missingQty > 0) return "missing";
    if (buyQty > 0 && buildQty > 0) return "buy_build";
    if (buyQty > 0) return "buy";
    if (buildQty > 0) return "build";
    return "stock";
  };

  const materialLaneCounts = useMemo(() => {
    return materialRows.reduce(
      (acc, row) => {
        const action = resolveMaterialAction(row);
        acc[action] += 1;
        return acc;
      },
      { stock: 0, buy: 0, build: 0, buy_build: 0, missing: 0 }
    );
  }, [materialRows]);

  const filteredMaterialRows = useMemo(() => {
    const query = materialSearch.trim().toLowerCase();
    return materialRows
      .filter((row) => {
        const action = resolveMaterialAction(row);
        if (materialFilterMode === "stock" && action !== "stock") return false;
        if (materialFilterMode === "buy" && action !== "buy" && action !== "buy_build") return false;
        if (materialFilterMode === "build" && action !== "build" && action !== "buy_build") return false;
        if (materialFilterMode === "missing" && action !== "missing") return false;
        if (!query) return true;
        const name = String(row.type_name || "").toLowerCase();
        const id = String(row.type_id || "");
        return name.includes(query) || id.includes(query);
      })
      .sort((a, b) => {
        const missingDelta = (b.missing_qty || 0) - (a.missing_qty || 0);
        if (missingDelta !== 0) return missingDelta;
        const buyDelta = (b.buy_qty || 0) - (a.buy_qty || 0);
        if (buyDelta !== 0) return buyDelta;
        const buildDelta = (b.build_qty || 0) - (a.build_qty || 0);
        if (buildDelta !== 0) return buildDelta;
        return (b.required_qty || 0) - (a.required_qty || 0);
      });
  }, [materialRows, materialFilterMode, materialSearch]);

  const filteredMaterialTotals = useMemo(() => {
    return filteredMaterialRows.reduce(
      (acc, row) => {
        acc.required += row.required_qty || 0;
        acc.stock += row.available_qty || 0;
        acc.buy += row.buy_qty || 0;
        acc.build += row.build_qty || 0;
        acc.missing += row.missing_qty || 0;
        return acc;
      },
      { required: 0, stock: 0, buy: 0, build: 0, missing: 0 }
    );
  }, [filteredMaterialRows]);

  const materialTotalPages = useMemo(
    () => Math.max(1, Math.ceil(filteredMaterialRows.length / Math.max(1, materialRowsPerPage))),
    [filteredMaterialRows.length, materialRowsPerPage]
  );

  useEffect(() => {
    setMaterialPage((prev) => Math.min(Math.max(1, prev), materialTotalPages));
  }, [materialTotalPages]);

  useEffect(() => {
    setMaterialPage(1);
  }, [materialFilterMode, materialSearch, materialRowsPerPage]);

  const visibleMaterialRows = useMemo(() => {
    const pageSize = Math.max(1, materialRowsPerPage);
    const start = (materialPage - 1) * pageSize;
    return filteredMaterialRows.slice(start, start + pageSize);
  }, [filteredMaterialRows, materialPage, materialRowsPerPage]);

  const csvEscape = (value: string): string => {
    const needsQuotes = value.includes(",") || value.includes("\"") || value.includes("\n");
    if (!needsQuotes) return value;
    return `"${value.replace(/"/g, "\"\"")}"`;
  };

  const downloadTextFile = (filename: string, content: string): void => {
    const blob = new Blob([content], { type: "text/plain;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    try {
      a.click();
    } finally {
      // remove() is safe even if node was detached by browser/runtime internals
      a.remove();
      URL.revokeObjectURL(url);
    }
  };

  const handleExportMaterialCSV = () => {
    if (filteredMaterialRows.length === 0) {
      setMaterialExportStatus(t("industryLedgerMaterialNoRowsExport"));
      return;
    }
    const header = ["type_id", "type_name", "required_qty", "stock_qty", "buy_qty", "build_qty", "missing_qty"];
    const rows = filteredMaterialRows.map((row) => [
      String(row.type_id || 0),
      csvEscape(String(row.type_name || `Type ${row.type_id}`)),
      String(row.required_qty || 0),
      String(row.available_qty || 0),
      String(row.buy_qty || 0),
      String(row.build_qty || 0),
      String(row.missing_qty || 0),
    ]);
    const content = [header.join(","), ...rows.map((r) => r.join(","))].join("\n");
    downloadTextFile(`industry-shopping-diff-${Date.now()}.csv`, content);
    setMaterialExportStatus(t("industryLedgerMaterialCSVExported", { count: filteredMaterialRows.length }));
  };

  const handleExportMaterialMultibuy = async () => {
    const lines = filteredMaterialRows
      .map((row) => {
        const qty = (row.buy_qty || 0) + (row.missing_qty || 0);
        if (qty <= 0) return "";
        const name = String(row.type_name || `Type ${row.type_id}`).trim();
        if (!name) return "";
        return `${name}\t${qty}`;
      })
      .filter((line) => line.length > 0);
    if (lines.length === 0) {
      setMaterialExportStatus(t("industryLedgerMaterialNoRowsMultibuy"));
      return;
    }
    const payload = lines.join("\n");
    downloadTextFile(`industry-multibuy-${Date.now()}.txt`, payload);
    try {
      if (navigator.clipboard && typeof navigator.clipboard.writeText === "function") {
        await navigator.clipboard.writeText(payload);
        setMaterialExportStatus(t("industryLedgerMaterialMultibuyCopied", { count: lines.length }));
        return;
      }
    } catch {
      // keep silent and fall back to export-only signal
    }
    setMaterialExportStatus(t("industryLedgerMaterialMultibuyExported", { count: lines.length }));
  };

  if (jobsWorkspaceTab !== "operations") {
    return null;
  }

  const actionHint = filteredMaterialTotals.missing > 0
    ? t("industryLedgerMaterialHintMissing")
    : filteredMaterialTotals.buy > 0
      ? t("industryLedgerMaterialHintBuy")
      : filteredMaterialTotals.build > 0
        ? t("industryLedgerMaterialHintBuild")
        : t("industryLedgerMaterialHintCovered");

  const strictLocationWarning = rebalanceWarehouseScope === "strict_location" && filteredMaterialTotals.missing > 0;
  const stationOnlyWarning = rebalanceUseSelectedStation && filteredMaterialTotals.missing > 0;
  const blueprintSyncHint = materialRows.length === 0;

  return (
    <div className="mt-2 border border-eve-border/40 rounded-sm p-2 bg-eve-dark/20">
      <div className="flex items-center justify-between gap-2 mb-1">
        <div className="text-[10px] uppercase tracking-wider text-eve-dim">
          {t("industryLedgerMaterialCoverageTitle", { count: materialRows.length })}
        </div>
        <div className="inline-flex items-center gap-1 text-[11px]">
          <span className="text-eve-dim">{t("industryLedgerMaterialScope")}</span>
          <select
            value={rebalanceInventoryScope}
            onChange={(e) => setRebalanceInventoryScope(e.target.value as "single" | "all")}
            className="px-1.5 py-0.5 bg-eve-input border border-eve-border rounded-sm text-[11px] text-eve-text"
          >
            <option value="single">{t("industryLedgerMaterialScopeSingle")}</option>
            <option value="all">{t("industryLedgerMaterialScopeAllChars")}</option>
          </select>
          <span className="text-eve-dim">{t("industryLedgerMaterialDays")}</span>
          <input
            type="number"
            min={1}
            max={365}
            value={rebalanceLookbackDays}
            onChange={(e) => setRebalanceLookbackDays(Math.max(1, Math.min(365, Math.round(Number(e.target.value) || 180))))}
            className="w-16 px-1.5 py-0.5 bg-eve-input border border-eve-border rounded-sm text-[11px] text-eve-text font-mono"
          />
          <span className="text-eve-dim">{t("industryLedgerMaterialFallback")}</span>
          <select
            value={rebalanceStrategy}
            onChange={(e) => setRebalanceStrategy(e.target.value as "preserve" | "buy" | "build")}
            className="px-1.5 py-0.5 bg-eve-input border border-eve-border rounded-sm text-[11px] text-eve-text"
          >
            <option value="preserve">{t("industryLedgerMaterialFallbackPreserve")}</option>
            <option value="buy">{t("industryLedgerMaterialFallbackBuy")}</option>
            <option value="build">{t("industryLedgerMaterialFallbackBuild")}</option>
          </select>
          <span className="text-eve-dim">{t("industryLedgerMaterialWarehouse")}</span>
          <select
            value={rebalanceWarehouseScope}
            onChange={(e) => setRebalanceWarehouseScope(e.target.value as "global" | "location_first" | "strict_location")}
            className="px-1.5 py-0.5 bg-eve-input border border-eve-border rounded-sm text-[11px] text-eve-text"
          >
            <option value="location_first">{t("industryLedgerMaterialWarehouseLocationFirst")}</option>
            <option value="strict_location">{t("industryLedgerMaterialWarehouseStrictLocation")}</option>
            <option value="global">{t("industryLedgerMaterialWarehouseGlobalPool")}</option>
          </select>
          <span className="text-eve-dim">{t("industryLedgerMaterialBPCRuns")}</span>
          <input
            type="number"
            min={1}
            max={1000}
            value={blueprintSyncDefaultBPCRuns}
            onChange={(e) => setBlueprintSyncDefaultBPCRuns(Math.max(1, Math.min(1000, Math.round(Number(e.target.value) || 1))))}
            className="w-16 px-1.5 py-0.5 bg-eve-input border border-eve-border rounded-sm text-[11px] text-eve-text font-mono"
          />
          <button
            type="button"
            onClick={() => { void handleSyncLedgerBlueprintPoolFromAssets(); }}
            disabled={syncingLedgerBlueprintPool}
            className="px-1.5 py-0.5 border border-fuchsia-500/40 text-fuchsia-300 rounded-sm hover:bg-fuchsia-500/10 disabled:opacity-50"
          >
            {syncingLedgerBlueprintPool ? t("industryLedgerMaterialSyncingBlueprints") : t("industryLedgerMaterialSyncBlueprints")}
          </button>
          <label className="inline-flex items-center gap-1 text-eve-dim select-none">
            <input
              type="checkbox"
              checked={rebalanceUseSelectedStation}
              onChange={(e) => setRebalanceUseSelectedStation(e.target.checked)}
              className="accent-eve-accent"
            />
            {t("industryLedgerMaterialStationOnly")}
          </label>
          <button
            type="button"
            onClick={() => { void handleRebalanceLedgerMaterialsFromInventory(); }}
            disabled={rebalancingLedgerMaterials || materialRows.length === 0}
            className="px-1.5 py-0.5 border border-cyan-500/40 text-cyan-300 rounded-sm hover:bg-cyan-500/10 disabled:opacity-50"
          >
            {rebalancingLedgerMaterials ? t("industryLedgerMaterialRebalancing") : t("industryLedgerMaterialRebalanceFromInventory")}
          </button>
        </div>
      </div>
      {blueprintSyncHint && (
        <div className="mb-1 text-[11px] text-yellow-300">{t("industryLedgerMaterialHintSyncBlueprints")}</div>
      )}
      {strictLocationWarning && (
        <div className="mb-1 text-[11px] text-red-300">{t("industryLedgerMaterialHintStrictLocation")}</div>
      )}
      {stationOnlyWarning && (
        <div className="mb-1 text-[11px] text-amber-300">{t("industryLedgerMaterialHintStationOnly")}</div>
      )}
      <div className="mb-1 grid grid-cols-2 md:grid-cols-5 gap-1 text-[11px]">
        <button
          type="button"
          onClick={() => setMaterialFilterMode("all")}
          className={`px-2 py-1 rounded-sm border text-left ${
            materialFilterMode === "all"
              ? "border-eve-accent/50 bg-eve-accent/10 text-eve-accent"
              : "border-eve-border text-eve-dim hover:text-eve-text"
          }`}
        >
          <div className="uppercase text-[10px] tracking-wider">{t("industryLedgerMaterialLaneAll")}</div>
          <div className="font-mono">{materialRows.length}</div>
        </button>
        <button
          type="button"
          onClick={() => setMaterialFilterMode("stock")}
          className={`px-2 py-1 rounded-sm border text-left ${
            materialFilterMode === "stock"
              ? "border-cyan-500/50 bg-cyan-500/10 text-cyan-300"
              : "border-cyan-500/30 text-cyan-200/90 hover:bg-cyan-500/5"
          }`}
        >
          <div className="uppercase text-[10px] tracking-wider">{t("industryLedgerMaterialLaneInStock")}</div>
          <div className="font-mono">{materialLaneCounts.stock.toLocaleString()}</div>
        </button>
        <button
          type="button"
          onClick={() => setMaterialFilterMode("buy")}
          className={`px-2 py-1 rounded-sm border text-left ${
            materialFilterMode === "buy"
              ? "border-amber-500/50 bg-amber-500/10 text-amber-300"
              : "border-amber-500/30 text-amber-200/90 hover:bg-amber-500/5"
          }`}
        >
          <div className="uppercase text-[10px] tracking-wider">{t("industryLedgerMaterialLaneNeedBuy")}</div>
          <div className="font-mono">{(materialLaneCounts.buy + materialLaneCounts.buy_build).toLocaleString()}</div>
        </button>
        <button
          type="button"
          onClick={() => setMaterialFilterMode("build")}
          className={`px-2 py-1 rounded-sm border text-left ${
            materialFilterMode === "build"
              ? "border-fuchsia-500/50 bg-fuchsia-500/10 text-fuchsia-300"
              : "border-fuchsia-500/30 text-fuchsia-200/90 hover:bg-fuchsia-500/5"
          }`}
        >
          <div className="uppercase text-[10px] tracking-wider">{t("industryLedgerMaterialLaneNeedBuild")}</div>
          <div className="font-mono">{(materialLaneCounts.build + materialLaneCounts.buy_build).toLocaleString()}</div>
        </button>
        <button
          type="button"
          onClick={() => setMaterialFilterMode("missing")}
          className={`px-2 py-1 rounded-sm border text-left ${
            materialFilterMode === "missing"
              ? "border-red-500/50 bg-red-500/10 text-red-300"
              : "border-red-500/30 text-red-200/90 hover:bg-red-500/5"
          }`}
        >
          <div className="uppercase text-[10px] tracking-wider">{t("industryLedgerMaterialLaneMissing")}</div>
          <div className="font-mono">{materialLaneCounts.missing.toLocaleString()}</div>
        </button>
      </div>
      <div className="mb-1 flex flex-wrap items-center gap-1 text-[11px]">
        <span className="text-eve-dim">{t("industryLedgerMaterialView")}</span>
        <select
          value={materialFilterMode}
          onChange={(e) => setMaterialFilterMode(e.target.value as "all" | "stock" | "buy" | "build" | "missing")}
          className="px-1.5 py-0.5 bg-eve-input border border-eve-border rounded-sm text-[11px] text-eve-text"
        >
          <option value="all">{t("industryLedgerMaterialViewAll")}</option>
          <option value="stock">{t("industryLedgerMaterialViewStockOnly")}</option>
          <option value="buy">{t("industryLedgerMaterialViewBuyOnly")}</option>
          <option value="build">{t("industryLedgerMaterialViewBuildOnly")}</option>
          <option value="missing">{t("industryLedgerMaterialViewMissingOnly")}</option>
        </select>
        <input
          type="text"
          value={materialSearch}
          onChange={(e) => setMaterialSearch(e.target.value)}
          placeholder={t("industryLedgerMaterialSearchPlaceholder")}
          className="px-1.5 py-0.5 bg-eve-input border border-eve-border rounded-sm text-[11px] text-eve-text min-w-[190px]"
        />
        <button
          type="button"
          onClick={handleExportMaterialCSV}
          disabled={filteredMaterialRows.length === 0}
          className="px-1.5 py-0.5 border border-eve-border text-eve-dim rounded-sm hover:text-eve-accent hover:border-eve-accent/40 disabled:opacity-50"
        >
          {t("industryExportCSV")}
        </button>
        <button
          type="button"
          onClick={() => { void handleExportMaterialMultibuy(); }}
          disabled={!filteredMaterialRows.some((row) => ((row.buy_qty || 0) + (row.missing_qty || 0)) > 0)}
          className="px-1.5 py-0.5 border border-emerald-500/40 text-emerald-300 rounded-sm hover:bg-emerald-500/10 disabled:opacity-50"
        >
          {t("industryLedgerMaterialExportMultibuy")}
        </button>
        <span className="text-eve-dim">
          {t("industryLedgerMaterialRowsCount", { visible: filteredMaterialRows.length, total: materialRows.length })}
        </span>
        <span className="text-eve-dim">{t("industryLedgerMaterialRowsPerPage")}</span>
        <select
          value={materialRowsPerPage}
          onChange={(e) => setMaterialRowsPerPage(Math.max(20, Math.min(500, Number(e.target.value) || 120)))}
          className="px-1.5 py-0.5 bg-eve-input border border-eve-border rounded-sm text-[11px] text-eve-text"
        >
          <option value={50}>50</option>
          <option value={120}>120</option>
          <option value={250}>250</option>
          <option value={500}>500</option>
        </select>
        <span className="text-eve-dim">
          {t("industryLedgerMaterialPage")} {materialPage}/{materialTotalPages}
        </span>
        <button
          type="button"
          onClick={() => setMaterialPage((prev) => Math.max(1, prev - 1))}
          disabled={materialPage <= 1}
          className="px-1 py-0.5 border border-eve-border rounded-sm text-eve-dim hover:text-eve-accent disabled:opacity-40"
        >
          {"<"}
        </button>
        <button
          type="button"
          onClick={() => setMaterialPage((prev) => Math.min(materialTotalPages, prev + 1))}
          disabled={materialPage >= materialTotalPages}
          className="px-1 py-0.5 border border-eve-border rounded-sm text-eve-dim hover:text-eve-accent disabled:opacity-40"
        >
          {">"}
        </button>
        <span className="text-eve-dim">
          {t("industryLedgerMaterialTotals", {
            buy: filteredMaterialTotals.buy.toLocaleString(),
            build: filteredMaterialTotals.build.toLocaleString(),
            missing: filteredMaterialTotals.missing.toLocaleString(),
          })}
        </span>
        <span className={filteredMaterialTotals.missing > 0 ? "text-red-300" : "text-cyan-300"}>
          {actionHint}
        </span>
        {materialExportStatus && (
          <span className="text-cyan-300">{materialExportStatus}</span>
        )}
      </div>
      <div className="border border-eve-border rounded-sm max-h-[180px] overflow-auto">
        <table className="w-full text-[11px]">
          <thead className="sticky top-0 bg-eve-dark z-10">
            <tr className="text-eve-dim uppercase tracking-wider border-b border-eve-border/60">
              <th className="px-1.5 py-1 text-left">{t("industryLedgerMaterialColumnMaterial")}</th>
              <th className="px-1.5 py-1 text-right">{t("industryLedgerMaterialColumnRequired")}</th>
              <th className="px-1.5 py-1 text-right">{t("industryLedgerMaterialColumnStock")}</th>
              <th className="px-1.5 py-1 text-right">{t("industryLedgerMaterialColumnBuy")}</th>
              <th className="px-1.5 py-1 text-right">{t("industryLedgerMaterialColumnBuild")}</th>
              <th className="px-1.5 py-1 text-right">{t("industryLedgerMaterialColumnMissing")}</th>
              <th className="px-1.5 py-1 text-left">{t("industryLedgerMaterialColumnAction")}</th>
            </tr>
          </thead>
          <tbody>
            {visibleMaterialRows.map((row: IndustryMaterialDiff) => (
              <tr key={`mat-diff-${row.type_id}`} className="border-b border-eve-border/30">
                <td className="px-1.5 py-1 text-eve-text">
                  <div className="truncate">{row.type_name || `Type ${row.type_id}`}</div>
                  <div className="text-[10px] text-eve-dim">#{row.type_id}</div>
                </td>
                <td className="px-1.5 py-1 text-right font-mono text-eve-accent">{(row.required_qty || 0).toLocaleString()}</td>
                <td className="px-1.5 py-1 text-right font-mono text-cyan-300">{(row.available_qty || 0).toLocaleString()}</td>
                <td className="px-1.5 py-1 text-right font-mono text-eve-dim">{(row.buy_qty || 0).toLocaleString()}</td>
                <td className="px-1.5 py-1 text-right font-mono text-fuchsia-300">{(row.build_qty || 0).toLocaleString()}</td>
                <td className="px-1.5 py-1 text-right font-mono text-red-300">{(row.missing_qty || 0).toLocaleString()}</td>
                <td className="px-1.5 py-1">
                  {(() => {
                    const action = resolveMaterialAction(row);
                    if (action === "missing") {
                      return <span className="px-1.5 py-0.5 text-[10px] uppercase rounded-sm border border-red-500/40 text-red-300 bg-red-500/10">{t("industryLedgerMaterialActionFixMissing")}</span>;
                    }
                    if (action === "buy_build") {
                      return <span className="px-1.5 py-0.5 text-[10px] uppercase rounded-sm border border-amber-500/40 text-amber-300 bg-amber-500/10">{t("industryLedgerMaterialActionBuyBuild")}</span>;
                    }
                    if (action === "buy") {
                      return <span className="px-1.5 py-0.5 text-[10px] uppercase rounded-sm border border-amber-500/40 text-amber-300 bg-amber-500/10">{t("industryLedgerMaterialActionBuy")}</span>;
                    }
                    if (action === "build") {
                      return <span className="px-1.5 py-0.5 text-[10px] uppercase rounded-sm border border-fuchsia-500/40 text-fuchsia-300 bg-fuchsia-500/10">{t("industryLedgerMaterialActionBuild")}</span>;
                    }
                    return <span className="px-1.5 py-0.5 text-[10px] uppercase rounded-sm border border-cyan-500/40 text-cyan-300 bg-cyan-500/10">{t("industryLedgerMaterialActionUseStock")}</span>;
                  })()}
                </td>
              </tr>
            ))}
            {filteredMaterialRows.length === 0 && (
              <tr>
                <td colSpan={7} className="px-2 py-2 text-center text-eve-dim">
                  {t("industryLedgerMaterialNoMatch")}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
