import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { findRoutes, setWaypointInGame } from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import type { RouteResult, RouteHop, ScanParams } from "@/lib/types";
import { ExecutionPlannerPopup } from "./ExecutionPlannerPopup";
import { useGlobalToast } from "./Toast";
import { handleEveUIError } from "@/lib/handleEveUIError";
import {
  TabSettingsPanel,
  SettingsCheckbox,
  SettingsField,
  SettingsGrid,
  SettingsNumberInput,
} from "./TabSettingsPanel";

type SortKey = "hops" | "profit" | "jumps" | "ppj";
type SortDir = "asc" | "desc";

interface Props {
  params: ScanParams;
  onChange?: (params: ScanParams) => void;
  /** Results loaded externally (e.g. from history) */
  loadedResults?: RouteResult[] | null;
  isLoggedIn?: boolean;
}

function formatISK(v: number): string {
  if (v >= 1e9) return (v / 1e9).toFixed(1) + "B";
  if (v >= 1e6) return (v / 1e6).toFixed(1) + "M";
  if (v >= 1e3) return (v / 1e3).toFixed(1) + "K";
  return v.toFixed(0);
}

function formatISKFull(v: number): string {
  return v.toLocaleString("en-US", { maximumFractionDigits: 0 });
}

export function RouteBuilder({ params, onChange, loadedResults, isLoggedIn = false }: Props) {
  const { t } = useI18n();
  const [minHops, setMinHops] = useState<number | "">(params.route_min_hops ?? 2);
  const [maxHops, setMaxHops] = useState<number | "">(params.route_max_hops ?? 5);
  const [targetSystemName, setTargetSystemName] = useState(params.route_target_system_name ?? "");
  const [minISKPerJump, setMinISKPerJump] = useState<number | "">(params.route_min_isk_per_jump ?? 0);
  const [allowEmptyHops, setAllowEmptyHops] = useState<boolean>(params.route_allow_empty_hops ?? false);
  const [results, setResults] = useState<RouteResult[]>([]);

  // Accept externally loaded results (from history)
  useEffect(() => {
    if (loadedResults && loadedResults.length > 0) {
      setResults(loadedResults);
    }
  }, [loadedResults]);

  useEffect(() => {
    setMinHops(params.route_min_hops ?? 2);
  }, [params.route_min_hops]);

  useEffect(() => {
    setMaxHops(params.route_max_hops ?? 5);
  }, [params.route_max_hops]);
  useEffect(() => {
    setTargetSystemName(params.route_target_system_name ?? "");
  }, [params.route_target_system_name]);
  useEffect(() => {
    setMinISKPerJump(params.route_min_isk_per_jump ?? 0);
  }, [params.route_min_isk_per_jump]);
  useEffect(() => {
    setAllowEmptyHops(params.route_allow_empty_hops ?? false);
  }, [params.route_allow_empty_hops]);
  const [scanning, setScanning] = useState(false);
  const [progress, setProgress] = useState("");
  const [selectedRoute, setSelectedRoute] = useState<RouteResult | null>(null);
  const [sortKey, setSortKey] = useState<SortKey>("profit");
  const [sortDir, setSortDir] = useState<SortDir>("desc");
  const abortRef = useRef<AbortController | null>(null);

  const applyRouteParams = useCallback(
    (patch: Partial<ScanParams>) => {
      if (!onChange) return;
      onChange({
        ...params,
        ...patch,
      });
    },
    [onChange, params],
  );

  const handleMinHopsChange = useCallback(
    (value: number) => {
      const boundedMin = Math.max(1, Math.min(25, value));
      const currentMax = typeof maxHops === "number" ? maxHops : 5;
      const boundedMax = Math.max(boundedMin, Math.min(25, currentMax));
      setMinHops(boundedMin);
      setMaxHops(boundedMax);
      applyRouteParams({
        route_min_hops: boundedMin,
        route_max_hops: boundedMax,
      });
    },
    [maxHops, applyRouteParams],
  );

  const handleMaxHopsChange = useCallback(
    (value: number) => {
      const currentMin = typeof minHops === "number" ? minHops : 2;
      const boundedMax = Math.max(currentMin, Math.min(25, value));
      setMaxHops(boundedMax);
      applyRouteParams({
        route_min_hops: currentMin,
        route_max_hops: boundedMax,
      });
    },
    [minHops, applyRouteParams],
  );

  const handleTargetSystemChange = useCallback(
    (value: string) => {
      setTargetSystemName(value);
      applyRouteParams({ route_target_system_name: value });
    },
    [applyRouteParams],
  );

  const handleMinISKPerJumpChange = useCallback(
    (value: number) => {
      const bounded = Math.max(0, value);
      setMinISKPerJump(bounded);
      applyRouteParams({ route_min_isk_per_jump: bounded });
    },
    [applyRouteParams],
  );

  const handleAllowEmptyHopsChange = useCallback(
    (enabled: boolean) => {
      setAllowEmptyHops(enabled);
      applyRouteParams({ route_allow_empty_hops: enabled });
    },
    [applyRouteParams],
  );

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortKey(key);
      setSortDir("desc");
    }
  };

  const sortedResults = useMemo(() => {
    if (results.length === 0) return results;
    const getter: Record<SortKey, (r: RouteResult) => number> = {
      hops: (r) => r.HopCount,
      profit: (r) => r.TotalProfit,
      jumps: (r) => r.TotalJumps,
      ppj: (r) => r.ProfitPerJump,
    };
    const get = getter[sortKey];
    const mul = sortDir === "asc" ? 1 : -1;
    return [...results].sort((a, b) => (get(a) - get(b)) * mul);
  }, [results, sortKey, sortDir]);

  const handleSearch = useCallback(async () => {
    if (scanning) {
      abortRef.current?.abort();
      return;
    }
    const controller = new AbortController();
    abortRef.current = controller;
    setScanning(true);
    setProgress(t("scanStarting"));
    setResults([]);
    setSelectedRoute(null);

    try {
      const searchMinHops = typeof minHops === "number" ? minHops : 2;
      const searchMaxHops = typeof maxHops === "number" ? maxHops : 5;
      const searchMinISK = typeof minISKPerJump === "number" ? Math.max(0, minISKPerJump) : 0;
      const searchParams: ScanParams = {
        ...params,
        route_target_system_name: targetSystemName.trim(),
        route_min_isk_per_jump: searchMinISK,
        route_allow_empty_hops: allowEmptyHops,
        route_min_hops: searchMinHops,
        route_max_hops: searchMaxHops,
      };
      const res = await findRoutes(searchParams, searchMinHops, searchMaxHops, setProgress, controller.signal);
      setResults(res);
    } catch (e: unknown) {
      if (e instanceof Error && e.name !== "AbortError") {
        setProgress(t("errorPrefix") + e.message);
      }
    } finally {
      setScanning(false);
    }
  }, [scanning, params, minHops, maxHops, minISKPerJump, targetSystemName, allowEmptyHops, t]);

  const routeSummary = (route: RouteResult) =>
    route.Hops
      .map((h) => h.SystemName)
      .concat([route.Hops[route.Hops.length - 1]?.DestSystemName ?? ""])
      .concat(route.TargetSystemName ? [route.TargetSystemName] : [])
      .filter(Boolean)
      .join(" → ");
  const copyRouteSystems = async (route: RouteResult) => {
    await navigator.clipboard.writeText(routeSummary(route));
  };

  return (
    <div className="flex flex-col h-full">
      {/* Settings Panel - unified design */}
      <div className="shrink-0 m-2">
        <TabSettingsPanel
          title={t("routeSettings")}
          hint={t("routeSettingsHint")}
          icon="🗺"
          defaultExpanded={true}
          persistKey="route"
          help={{ stepKeys: ["helpRouteStep1", "helpRouteStep2", "helpRouteStep3"], wikiSlug: "Route-Builder" }}
        >
          <div className="flex items-center gap-4 flex-wrap">
            <SettingsGrid cols={4}>
              <SettingsField label={t("routeMinHops")}>
                <SettingsNumberInput
                  value={typeof minHops === "number" ? minHops : 2}
                  onChange={handleMinHopsChange}
                  min={1}
                  max={25}
                />
              </SettingsField>
              <SettingsField label={t("routeMaxHops")}>
                <SettingsNumberInput
                  value={typeof maxHops === "number" ? maxHops : 5}
                  onChange={handleMaxHopsChange}
                  min={typeof minHops === "number" ? minHops : 1}
                  max={25}
                />
              </SettingsField>
              <SettingsField label={t("routeMinISKPerJump")}>
                <SettingsNumberInput
                  value={typeof minISKPerJump === "number" ? minISKPerJump : 0}
                  onChange={handleMinISKPerJumpChange}
                  min={0}
                  step={1000}
                />
              </SettingsField>
              <SettingsField label={t("routeTargetSystem")}>
                <input
                  type="text"
                  value={targetSystemName}
                  onChange={(e) => handleTargetSystemChange(e.target.value)}
                  placeholder={t("routeTargetSystemPlaceholder")}
                  className="w-full px-3 py-1.5 bg-eve-input border border-eve-border rounded-sm text-eve-text text-sm
                             focus:outline-none focus:border-eve-accent focus:ring-1 focus:ring-eve-accent/30
                             transition-colors"
                />
              </SettingsField>
              <SettingsField label={t("routeAllowEmptyHops")}>
                <SettingsCheckbox
                  checked={allowEmptyHops}
                  onChange={handleAllowEmptyHopsChange}
                />
              </SettingsField>
            </SettingsGrid>

            <div className="flex items-center gap-3 ml-auto">
              <button
                onClick={handleSearch}
                disabled={!params.system_name}
                className={`px-5 py-1.5 rounded-sm text-xs font-semibold uppercase tracking-wider transition-all
                  ${scanning
                    ? "bg-eve-error/80 text-white hover:bg-eve-error"
                    : "bg-eve-accent text-eve-dark hover:bg-eve-accent-hover shadow-eve-glow"
                  }
                  disabled:bg-eve-input disabled:text-eve-dim disabled:cursor-not-allowed disabled:shadow-none`}
              >
                {scanning ? t("stop") : t("routeFind")}
              </button>
              {progress && <span className="text-[10px] text-eve-dim">{progress}</span>}
            </div>
          </div>
          {results.length > 0 && (
            <div className="mt-2 text-xs text-eve-dim">
              {t("routeFound", { count: results.length })}
            </div>
          )}
        </TabSettingsPanel>
      </div>

      {/* Results table */}
      <div className="flex-1 min-h-0 overflow-auto">
        {results.length > 0 ? (
          <table className="w-full text-xs">
            <thead className="sticky top-0 bg-eve-panel z-10">
              <tr className="text-eve-dim text-[10px] uppercase tracking-wider border-b border-eve-border">
                <th className="px-3 py-2 text-left font-medium">#</th>
                <th className="px-3 py-2 text-left font-medium">{t("routeColumn")}</th>
                <SortTh k="hops" cur={sortKey} dir={sortDir} onClick={toggleSort} align="right" label={t("routeHopsCol")} />
                <SortTh k="profit" cur={sortKey} dir={sortDir} onClick={toggleSort} align="right" label={t("colProfit")} />
                <SortTh k="jumps" cur={sortKey} dir={sortDir} onClick={toggleSort} align="right" label={t("colJumps")} />
                <SortTh k="ppj" cur={sortKey} dir={sortDir} onClick={toggleSort} align="right" label={t("colProfitPerJump")} />
              </tr>
            </thead>
            <tbody>
              {sortedResults.map((route, i) => (
                <tr
                  key={i}
                  onDoubleClick={() => setSelectedRoute(route)}
                  className="cursor-pointer hover:bg-eve-accent/10 border-b border-eve-border/30 transition-colors"
                >
                  <td className="px-3 py-2 text-eve-dim font-mono">{i + 1}</td>
                  <td className="px-3 py-2 text-eve-text max-w-[400px] truncate" title={routeSummary(route)}>
                    {routeSummary(route)}
                  </td>
                  <td className="px-3 py-2 text-right font-mono text-eve-dim">{route.HopCount}</td>
                  <td className="px-3 py-2 text-right font-mono text-green-400">{formatISK(route.TotalProfit)}</td>
                  <td className="px-3 py-2 text-right font-mono text-eve-dim">{route.TotalJumps}</td>
                  <td className="px-3 py-2 text-right font-mono text-yellow-400">{formatISK(route.ProfitPerJump)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : !scanning ? (
          <div className="flex items-center justify-center h-full text-eve-dim text-xs">
            {progress || t("routePrompt")}
          </div>
        ) : null}
      </div>

      {/* Detail popup */}
      {selectedRoute && (
          <RouteDetailPopup
          route={selectedRoute}
          onClose={() => setSelectedRoute(null)}
          onCopySystems={copyRouteSystems}
          salesTaxPercent={params.sales_tax_percent ?? 0}
          brokerFeePercent={params.broker_fee_percent ?? 0}
          splitTradeFees={params.split_trade_fees ?? false}
          buyBrokerFeePercent={params.buy_broker_fee_percent}
          sellBrokerFeePercent={params.sell_broker_fee_percent}
          buySalesTaxPercent={params.buy_sales_tax_percent}
          sellSalesTaxPercent={params.sell_sales_tax_percent}
          isLoggedIn={isLoggedIn}
        />
      )}
    </div>
  );
}

function SortTh({
  k,
  cur,
  dir,
  onClick,
  align,
  label,
}: {
  k: SortKey;
  cur: SortKey;
  dir: SortDir;
  onClick: (k: SortKey) => void;
  align: "left" | "right";
  label: string;
}) {
  const active = cur === k;
  return (
    <th
      className={`px-3 py-2 font-medium cursor-pointer select-none hover:text-eve-accent transition-colors ${
        align === "right" ? "text-right" : "text-left"
      } ${active ? "text-eve-accent" : ""}`}
      onClick={() => onClick(k)}
    >
      {label}
      {active && (
        <span className="ml-1 text-[9px]">{dir === "asc" ? "\u25B2" : "\u25BC"}</span>
      )}
    </th>
  );
}

function RouteDetailPopup({
  route,
  onClose,
  onCopySystems,
  salesTaxPercent = 0,
  brokerFeePercent = 0,
  splitTradeFees = false,
  buyBrokerFeePercent,
  sellBrokerFeePercent,
  buySalesTaxPercent,
  sellSalesTaxPercent,
  isLoggedIn = false,
}: {
  route: RouteResult;
  onClose: () => void;
  onCopySystems: (route: RouteResult) => Promise<void>;
  salesTaxPercent?: number;
  brokerFeePercent?: number;
  splitTradeFees?: boolean;
  buyBrokerFeePercent?: number;
  sellBrokerFeePercent?: number;
  buySalesTaxPercent?: number;
  sellSalesTaxPercent?: number;
  isLoggedIn?: boolean;
}) {
  const { t } = useI18n();
  const { addToast } = useGlobalToast();
  const [execPlanHop, setExecPlanHop] = useState<RouteHop | null>(null);

  const handleSetWaypoint = async (systemID: number) => {
    try {
      await setWaypointInGame(systemID);
      addToast(t("actionSuccess"), "success", 2000);
    } catch (err: any) {
      const { messageKey, duration } = handleEveUIError(err);
      addToast(t(messageKey), "error", duration);
    }
  };

  const handleCopySystems = async () => {
    try {
      await onCopySystems(route);
      addToast(t("copied"), "success", 1400);
    } catch {
      addToast(t("errorSomethingWentWrong"), "error", 2200);
    }
  };

  const handleCopyRoute = async () => {
    const lines = ["=== EVE Flipper Route ==="];
    route.Hops.forEach((hop, i) => {
      const emptyJumps = hop.EmptyJumps ?? 0;
      const totalHopJumps = hop.Jumps + emptyJumps;
      lines.push(`[${i + 1}] ${hop.StationName || hop.SystemName}`);
      lines.push(`    Buy: ${hop.TypeName} x${hop.Units} @ ${formatISKFull(hop.BuyPrice)} ISK`);
      if (emptyJumps > 0) {
        lines.push(`    Empty move: ${emptyJumps} jumps`);
      }
      lines.push(`    → ${hop.DestSystemName} (${totalHopJumps} jumps, trade ${hop.Jumps})`);
      lines.push(`    Sell: @ ${formatISKFull(hop.SellPrice)} ISK → Profit: ${formatISK(hop.Profit)}`);
      lines.push("");
    });
    if (route.TargetSystemName) {
      lines.push(`Target: ${route.TargetSystemName} (${route.TargetJumps ?? 0} jumps)`);
    }
    lines.push(`Total: ${formatISKFull(route.TotalProfit)} ISK / ${route.TotalJumps} jumps / ${formatISK(route.ProfitPerJump)} ISK/jump`);
    try {
      await navigator.clipboard.writeText(lines.join("\n"));
      addToast(t("copied"), "success", 1400);
    } catch {
      addToast(t("errorSomethingWentWrong"), "error", 2200);
    }
  };

  return (
    <>
    <div
      className="fixed inset-0 bg-black/60 flex items-center justify-center z-50"
      onClick={onClose}
    >
      <div
        className="bg-eve-panel border border-eve-border rounded-sm max-w-2xl w-full mx-2 sm:mx-4 max-h-[90vh] sm:max-h-[80vh] flex flex-col shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-eve-border">
          <h2 className="text-sm font-semibold text-eve-accent uppercase tracking-wider">
            {t("routeDetails")}
          </h2>
          <button
            onClick={onClose}
            className="text-eve-dim hover:text-eve-text text-lg leading-none"
          >
            ✕
          </button>
        </div>

        {/* Hops */}
        <div className="flex-1 overflow-y-auto p-4 space-y-0">
          {route.Hops.map((hop, i) => (
            <div key={i}>
              {/* Hop card */}
              <div className="bg-eve-dark/50 border border-eve-border/50 rounded-sm p-3">
                <div className="flex items-center gap-2 mb-2">
                  <span className="w-6 h-6 flex items-center justify-center rounded-full bg-eve-accent/20 text-eve-accent text-[11px] font-bold">
                    {i + 1}
                  </span>
                  <span className="text-xs font-medium text-eve-text">
                    {hop.StationName || hop.SystemName}
                  </span>
                  <div className="ml-auto flex items-center gap-1.5">
                    {isLoggedIn && hop.SystemID && (
                      <RouteDetailActionButton onClick={() => handleSetWaypoint(hop.SystemID)} title={t("setDestination")} tone="neutral">
                        <span className="text-[11px] leading-none">⌖</span>
                        <span>{t("routeBuy")}</span>
                      </RouteDetailActionButton>
                    )}
                    {isLoggedIn && hop.DestSystemID && (
                      <RouteDetailActionButton onClick={() => handleSetWaypoint(hop.DestSystemID)} title={t("setDestination")} tone="neutral">
                        <span className="text-[11px] leading-none">⌖</span>
                        <span>{t("routeSell")}</span>
                      </RouteDetailActionButton>
                    )}
                    {hop.RegionID != null && hop.RegionID > 0 && (
                      <RouteDetailActionButton onClick={() => setExecPlanHop(hop)} title={t("execPlanTitle")} tone="accent">
                        <span className="text-[11px] leading-none">▦</span>
                        <span className="hidden sm:inline">{t("execPlanTitle")}</span>
                      </RouteDetailActionButton>
                    )}
                  </div>
                </div>

                <div className="ml-8 space-y-1 text-xs">
                  <div className="flex items-center gap-2">
                    <span className="text-eve-dim">{t("routeBuy")}:</span>
                    <span className="text-eve-text font-medium">{hop.TypeName}</span>
                    <span className="text-eve-dim">×{hop.Units}</span>
                    <span className="text-eve-dim">@</span>
                    <span className="font-mono text-eve-text">{formatISKFull(hop.BuyPrice)} ISK</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-eve-dim">→ {t("routeDeliverTo")}:</span>
                    <span className="text-eve-text">{hop.DestStationName || hop.DestSystemName}</span>
                    <span className="text-eve-dim font-mono">
                      ({hop.Jumps + (hop.EmptyJumps ?? 0)} {t("routeJumpsUnit")})
                    </span>
                    {(hop.EmptyJumps ?? 0) > 0 && (
                      <span className="text-eve-dim text-[11px]">
                        {t("routeEmptyLeg", { count: hop.EmptyJumps ?? 0 })}
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-eve-dim">{t("routeSell")}:</span>
                    <span className="font-mono text-eve-text">@ {formatISKFull(hop.SellPrice)} ISK</span>
                    <span className="text-eve-dim">→</span>
                    <span className="font-mono text-green-400">+{formatISKFull(hop.Profit)} ISK</span>
                  </div>
                </div>
              </div>

              {/* Connector */}
              {i < route.Hops.length - 1 && (
                <div className="flex justify-center py-1">
                  <div className="flex flex-col items-center">
                    <div className="w-px h-2 bg-eve-border" />
                    <svg width="10" height="6" viewBox="0 0 10 6" className="text-eve-accent">
                      <path d="M5 6L0 0h10z" fill="currentColor" />
                    </svg>
                    <div className="w-px h-2 bg-eve-border" />
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Summary + actions footer */}
        <div className="px-4 py-3 border-t border-eve-border bg-eve-dark/30 space-y-3">
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
            <RouteMetricChip label={t("routeTotalProfit")} value={`${formatISKFull(route.TotalProfit)} ISK`} tone="profit" />
            <RouteMetricChip label={t("routeTotalJumps")} value={String(route.TotalJumps)} tone="dim" />
            <RouteMetricChip label={`ISK/${t("routeJumpsUnit")}`} value={formatISK(route.ProfitPerJump)} tone="ppj" />
            <RouteMetricChip label={t("routeHopsCol")} value={String(route.HopCount)} tone="dim" />
            {route.TargetSystemName && (
              <RouteMetricChip
                label={t("routeTargetTail")}
                value={`${route.TargetSystemName} (+${route.TargetJumps ?? 0})`}
                tone="dim"
              />
            )}
          </div>
          <div className="flex flex-wrap justify-end gap-2">
            <button
              onClick={handleCopySystems}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-sm text-[11px] font-semibold uppercase tracking-wider text-eve-dim border border-eve-border bg-eve-dark/60 hover:text-eve-text hover:border-eve-accent/30 hover:bg-eve-dark transition-all"
            >
              <span className="text-[11px] leading-none">⎘</span>
              <span>{t("copyRouteSystems")}</span>
            </button>
            <button
              onClick={handleCopyRoute}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-sm text-[11px] font-semibold uppercase tracking-wider text-eve-dark bg-eve-accent border border-eve-accent hover:bg-eve-accent-hover shadow-eve-glow transition-all"
            >
              <span className="text-[11px] leading-none">⎘</span>
              <span>{t("copyRoute")}</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <ExecutionPlannerPopup
      open={execPlanHop !== null}
      onClose={() => setExecPlanHop(null)}
      typeID={execPlanHop?.TypeID ?? 0}
      typeName={execPlanHop?.TypeName ?? ""}
      regionID={execPlanHop?.RegionID ?? 0}
      defaultQuantity={execPlanHop?.Units ?? 100}
      isBuy={true}
      brokerFeePercent={brokerFeePercent}
      salesTaxPercent={salesTaxPercent}
      splitTradeFees={splitTradeFees}
      buyBrokerFeePercent={buyBrokerFeePercent}
      sellBrokerFeePercent={sellBrokerFeePercent}
      buySalesTaxPercent={buySalesTaxPercent}
      sellSalesTaxPercent={sellSalesTaxPercent}
    />
    </>
  );
}

function RouteDetailActionButton({
  onClick,
  title,
  children,
  tone,
}: {
  onClick: () => void;
  title: string;
  children: ReactNode;
  tone: "neutral" | "accent";
}) {
  const styleByTone =
    tone === "accent"
      ? "text-eve-accent border-eve-accent/40 bg-eve-accent/10 hover:bg-eve-accent/20 hover:border-eve-accent/60"
      : "text-eve-dim border-eve-border bg-eve-dark/40 hover:text-eve-text hover:border-eve-accent/30 hover:bg-eve-dark/70";
  return (
    <button
      type="button"
      onClick={onClick}
      title={title}
      className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-sm text-[10px] font-semibold uppercase tracking-wide border transition-all ${styleByTone}`}
    >
      {children}
    </button>
  );
}

function RouteMetricChip({
  label,
  value,
  tone,
}: {
  label: string;
  value: string;
  tone: "profit" | "ppj" | "dim";
}) {
  const valueClass = tone === "profit" ? "text-green-400" : tone === "ppj" ? "text-yellow-400" : "text-eve-text";
  return (
    <div className="border border-eve-border/60 bg-eve-dark/70 px-2 py-1.5 rounded-sm">
      <div className="text-[10px] uppercase tracking-wide text-eve-dim">{label}</div>
      <div className={`text-xs font-mono font-semibold ${valueClass}`}>{value}</div>
    </div>
  );
}
