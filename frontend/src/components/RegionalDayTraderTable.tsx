import { Fragment, useCallback, useMemo, useState } from "react";
import type { RegionalDayTradeHub, RegionalDayTradeItem, StationCacheMeta } from "@/lib/types";
import { formatISK, formatISKFull, formatMargin } from "@/lib/format";
import { EmptyState } from "./EmptyState";

interface Props {
  hubs: RegionalDayTradeHub[];
  scanning: boolean;
  progress: string;
  cacheMeta?: StationCacheMeta | null;
  totalItems?: number;
  targetRegionName?: string;
  periodDays?: number;
}

// ─── Sanity signals ──────────────────────────────────────────────────────────

interface Signal {
  label: string;
  title: string;
  severity: "warn" | "info";
}

function itemSignals(item: RegionalDayTradeItem): Signal[] {
  const signals: Signal[] = [];
  if (item.target_dos > 90)
    signals.push({ label: "SAT", title: `Market saturated: ${item.target_dos.toFixed(0)} days of supply`, severity: "warn" });
  if (item.target_demand_per_day > 0 && item.target_demand_per_day < 1)
    signals.push({ label: "LOW", title: `Very low demand: ${item.target_demand_per_day.toFixed(2)} units/day`, severity: "warn" });
  if (item.source_avg_price > 0 && item.target_now_price > 0) {
    const spread = (item.target_now_price - item.source_avg_price) / item.source_avg_price * 100;
    if (spread > 200)
      signals.push({ label: "ODD", title: `Unusual spread: ${spread.toFixed(0)}% — verify prices`, severity: "warn" });
  }
  return signals;
}

// ─── Sorting ─────────────────────────────────────────────────────────────────

type SortKey =
  | "name"
  | "purchase_units"
  | "source_units"
  | "demand_per_day"
  | "supply_units"
  | "dos"
  | "now_profit"
  | "period_profit"
  | "roi_now"
  | "roi_period"
  | "margin_now"
  | "margin_period"
  | "capital"
  | "shipping"
  | "source_price"
  | "target_now_price"
  | "target_period_price"
  | "jumps"
  | "item_volume"
  | "item_count";

type SortDir = "asc" | "desc";

function hubSortVal(hub: RegionalDayTradeHub, key: SortKey): number | string {
  switch (key) {
    case "name": return hub.source_system_name;
    case "purchase_units": return hub.purchase_units;
    case "source_units": return hub.source_units;
    case "demand_per_day": return hub.target_demand_per_day;
    case "supply_units": return hub.target_supply_units;
    case "dos": return hub.target_dos;
    case "now_profit": return hub.target_now_profit;
    case "period_profit": return hub.target_period_profit;
    case "roi_now": return hub.target_now_profit / Math.max(hub.capital_required + hub.shipping_cost, 1) * 100;
    case "roi_period": return hub.target_period_profit / Math.max(hub.capital_required + hub.shipping_cost, 1) * 100;
    case "capital": return hub.capital_required;
    case "shipping": return hub.shipping_cost;
    case "item_count": return hub.item_count;
    default: return 0;
  }
}

function itemSortVal(item: RegionalDayTradeItem, key: SortKey): number | string {
  switch (key) {
    case "name": return item.type_name;
    case "purchase_units": return item.purchase_units;
    case "source_units": return item.source_units;
    case "demand_per_day": return item.target_demand_per_day;
    case "supply_units": return item.target_supply_units;
    case "dos": return item.target_dos;
    case "now_profit": return item.target_now_profit;
    case "period_profit": return item.target_period_profit;
    case "roi_now": return item.roi_now;
    case "roi_period": return item.roi_period;
    case "margin_now": return item.margin_now;
    case "margin_period": return item.margin_period;
    case "capital": return item.capital_required;
    case "shipping": return item.shipping_cost;
    case "source_price": return item.source_avg_price;
    case "target_now_price": return item.target_now_price;
    case "target_period_price": return item.target_period_price;
    case "jumps": return item.jumps;
    case "item_volume": return item.item_volume;
    default: return 0;
  }
}

function cmp(a: number | string, b: number | string, dir: SortDir): number {
  if (typeof a === "string" && typeof b === "string")
    return dir === "asc" ? a.localeCompare(b) : b.localeCompare(a);
  return dir === "asc" ? (a as number) - (b as number) : (b as number) - (a as number);
}

// ─── CSV export ──────────────────────────────────────────────────────────────

function exportCSV(hubs: RegionalDayTradeHub[]) {
  const header = [
    "Source Hub", "Source Region", "Item", "Target", "Target Region",
    "Buy Units", "Source Units", "Demand/Day", "Supply Units", "DOS",
    "Source Price", "Target Now Price", "Target Period Price",
    "Now Profit", "Period Profit", "ROI Now %", "ROI Period %",
    "Margin Now %", "Margin Period %", "Capital", "Shipping", "Jumps", "Vol m3",
  ];
  const rows: string[][] = [header];
  for (const hub of hubs) {
    for (const item of hub.items) {
      rows.push([
        hub.source_system_name,
        hub.source_region_name,
        item.type_name,
        item.target_system_name,
        item.target_region_name,
        String(item.purchase_units),
        String(item.source_units),
        item.target_demand_per_day.toFixed(2),
        String(item.target_supply_units),
        item.target_dos.toFixed(2),
        item.source_avg_price.toFixed(2),
        item.target_now_price.toFixed(2),
        item.target_period_price.toFixed(2),
        item.target_now_profit.toFixed(0),
        item.target_period_profit.toFixed(0),
        item.roi_now.toFixed(2),
        item.roi_period.toFixed(2),
        item.margin_now.toFixed(2),
        item.margin_period.toFixed(2),
        item.capital_required.toFixed(0),
        item.shipping_cost.toFixed(0),
        String(item.jumps),
        item.item_volume.toFixed(4),
      ]);
    }
  }
  const csv = rows.map((r) => r.map((c) => `"${c.replace(/"/g, '""')}"`).join(",")).join("\n");
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `regional_day_trader_${new Date().toISOString().slice(0, 10)}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}

// ─── Detail panel ─────────────────────────────────────────────────────────────

function DetailPanel({
  item,
  hubSecurity,
  onClose,
}: {
  item: RegionalDayTradeItem;
  hubSecurity: number;
  onClose: () => void;
}) {
  const signals = itemSignals(item);
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={onClose}>
      <div
        className="bg-eve-panel border border-eve-border rounded-sm shadow-2xl w-full max-w-lg mx-4 font-mono text-xs text-eve-text"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-3 py-2 border-b border-eve-border bg-eve-dark/60">
          <span className="font-semibold text-sm text-eve-accent truncate">{item.type_name}</span>
          <button onClick={onClose} className="text-eve-dim hover:text-eve-text ml-3 shrink-0">✕</button>
        </div>

        <div className="p-3 space-y-3">
          {/* Signals */}
          {signals.length > 0 && (
            <div className="flex flex-wrap gap-1.5">
              {signals.map((s, i) => (
                <span
                  key={i}
                  title={s.title}
                  className={`px-1.5 py-0.5 rounded-sm border text-[10px] font-bold cursor-help ${
                    s.severity === "warn"
                      ? "border-yellow-500/60 text-yellow-400 bg-yellow-900/20"
                      : "border-blue-500/60 text-blue-400 bg-blue-900/20"
                  }`}
                >
                  ⚠ {s.label}
                </span>
              ))}
            </div>
          )}

          {/* Route */}
          <div className="flex items-center gap-2 text-eve-dim">
            <span className="text-eve-text">{item.source_system_name}</span>
            <span>({item.source_region_name})</span>
            <span className="text-eve-accent">→ {item.jumps}j →</span>
            <span className="text-eve-text">{item.target_system_name}</span>
            <span>({item.target_region_name})</span>
          </div>

          {/* Prices */}
          <div className="grid grid-cols-2 gap-3">
            <div className="rounded-sm border border-eve-border/50 bg-eve-dark/30 p-2 space-y-1">
              <div className="text-[10px] uppercase tracking-wider text-eve-dim font-semibold">Source (Buy)</div>
              <Row label="Price" value={formatISKFull(item.source_avg_price)} />
              <Row label="Available" value={item.source_units.toLocaleString() + " units"} />
              <Row label="Sec" value={hubSecurity.toFixed(1)} />
            </div>
            <div className="rounded-sm border border-eve-border/50 bg-eve-dark/30 p-2 space-y-1">
              <div className="text-[10px] uppercase tracking-wider text-eve-dim font-semibold">Target (Sell)</div>
              <Row label="Now price" value={formatISKFull(item.target_now_price)} />
              <Row label="Period avg price" value={formatISKFull(item.target_period_price)} />
              <Row label="Demand/Day" value={item.target_demand_per_day.toFixed(2)} />
              <Row label="Supply" value={item.target_supply_units.toLocaleString()} />
              <Row label="DOS" value={item.target_dos.toFixed(2) + " days"} dim={item.target_dos > 30} />
            </div>
          </div>

          {/* Position */}
          <div className="rounded-sm border border-eve-border/50 bg-eve-dark/30 p-2 space-y-1">
            <div className="text-[10px] uppercase tracking-wider text-eve-dim font-semibold">
              Position ({item.purchase_units.toLocaleString()} units · {item.item_volume.toFixed(2)} m³/unit)
            </div>
            <div className="grid grid-cols-2 gap-x-4">
              <Row label="Capital" value={formatISKFull(item.capital_required)} />
              <Row label="Shipping" value={formatISKFull(item.shipping_cost)} />
              <Row label="Now Profit" value={formatISK(item.target_now_profit)} accent={item.target_now_profit > 0} />
              <Row label="Period Profit" value={formatISK(item.target_period_profit)} accent={item.target_period_profit > 0} />
              <Row label="ROI Now" value={formatMargin(item.roi_now)} accent={item.roi_now > 0} />
              <Row label="ROI Period" value={formatMargin(item.roi_period)} accent={item.roi_period > 0} />
              <Row label="Margin Now" value={formatMargin(item.margin_now)} />
              <Row label="Margin Period" value={formatMargin(item.margin_period)} />
            </div>
          </div>

          {/* Coverage */}
          {(item.assets > 0 || item.active_orders > 0) && (
            <div className="rounded-sm border border-eve-border/50 bg-eve-dark/30 p-2 space-y-1">
              <div className="text-[10px] uppercase tracking-wider text-eve-dim font-semibold">Inventory Coverage</div>
              {item.assets > 0 && <Row label="Assets in target" value={item.assets.toLocaleString() + " units"} />}
              {item.active_orders > 0 && <Row label="Active sell orders" value={item.active_orders.toLocaleString() + " units"} />}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function Row({
  label,
  value,
  accent,
  dim,
}: {
  label: string;
  value: string;
  accent?: boolean;
  dim?: boolean;
}) {
  return (
    <div className="flex items-center justify-between gap-2">
      <span className="text-eve-dim">{label}</span>
      <span className={accent ? "text-eve-accent font-semibold" : dim ? "text-yellow-400" : "text-eve-text"}>
        {value}
      </span>
    </div>
  );
}

// ─── Countdown helper ─────────────────────────────────────────────────────────

function formatCountdown(totalSeconds: number): string {
  const sec = Math.max(0, Math.floor(totalSeconds));
  const mm = Math.floor(sec / 60).toString().padStart(2, "0");
  const ss = (sec % 60).toString().padStart(2, "0");
  return `${mm}:${ss}`;
}

function cacheBadge(meta?: StationCacheMeta | null): string {
  if (!meta?.next_expiry_at) return "Cache n/a";
  const expiry = Date.parse(meta.next_expiry_at);
  if (!Number.isFinite(expiry)) return "Cache n/a";
  const secondsLeft = Math.floor((expiry - Date.now()) / 1000);
  if (secondsLeft <= 0) return "Cache stale";
  return `Cache ${formatCountdown(secondsLeft)}`;
}

// ─── Sort header ──────────────────────────────────────────────────────────────

function SortTh({
  label,
  sortKey,
  current,
  dir,
  onSort,
  className = "",
}: {
  label: string;
  sortKey: SortKey;
  current: SortKey;
  dir: SortDir;
  onSort: (k: SortKey) => void;
  className?: string;
}) {
  const active = current === sortKey;
  return (
    <th
      className={`text-right px-2 py-1.5 cursor-pointer select-none hover:text-eve-accent transition-colors ${
        active ? "text-eve-accent" : ""
      } ${className}`}
      onClick={() => onSort(sortKey)}
    >
      {label}
      {active && <span className="ml-0.5 text-[9px]">{dir === "desc" ? "▼" : "▲"}</span>}
    </th>
  );
}

// ─── Main component ───────────────────────────────────────────────────────────

export function RegionalDayTraderTable({
  hubs,
  scanning,
  progress,
  cacheMeta,
  totalItems = 0,
  targetRegionName = "",
  periodDays = 14,
}: Props) {
  const [expanded, setExpanded] = useState<Set<number>>(new Set());
  const [sortKey, setSortKey] = useState<SortKey>("period_profit");
  const [sortDir, setSortDir] = useState<SortDir>("desc");
  const [detail, setDetail] = useState<{ item: RegionalDayTradeItem; hubSecurity: number } | null>(null);

  const handleSort = useCallback((key: SortKey) => {
    setSortKey((prev) => {
      if (prev === key) setSortDir((d) => (d === "desc" ? "asc" : "desc"));
      else setSortDir("desc");
      return key;
    });
  }, []);

  const sortedHubs = useMemo(() => {
    return [...hubs]
      .map((hub) => ({
        ...hub,
        items: [...hub.items].sort((a, b) =>
          cmp(itemSortVal(a, sortKey), itemSortVal(b, sortKey), sortDir)
        ),
      }))
      .sort((a, b) => cmp(hubSortVal(a, sortKey), hubSortVal(b, sortKey), sortDir));
  }, [hubs, sortKey, sortDir]);

  const totals = useMemo(() => {
    let nowProfit = 0, periodProfit = 0, capital = 0;
    for (const hub of hubs) {
      nowProfit += hub.target_now_profit;
      periodProfit += hub.target_period_profit;
      capital += hub.capital_required;
    }
    return { nowProfit, periodProfit, capital };
  }, [hubs]);

  if (scanning && hubs.length === 0)
    return <EmptyState reason="loading" hints={progress ? [progress] : []} />;
  if (!scanning && hubs.length === 0)
    return <EmptyState reason={progress ? "no_results" : "no_scan_yet"} />;

  const thProps = { current: sortKey, dir: sortDir, onSort: handleSort };

  return (
    <div className="h-full min-h-0 flex flex-col">
      {/* Stats bar */}
      <div className="shrink-0 px-2 py-1.5 border-b border-eve-border bg-eve-dark/30 text-[11px] font-mono text-eve-dim flex items-center justify-between gap-2">
        <div className="flex items-center gap-3 min-w-0">
          <span className="text-eve-text">HUBS <span className="text-eve-accent">{hubs.length}</span></span>
          <span>ITEMS <span className="text-eve-accent">{totalItems}</span></span>
          <span>PERIOD <span className="text-eve-accent">{periodDays}d</span></span>
          {targetRegionName && (
            <span className="truncate">TARGET <span className="text-eve-accent">{targetRegionName}</span></span>
          )}
        </div>
        <div className="flex items-center gap-3">
          <span>NOW {formatISK(totals.nowProfit)}</span>
          <span>PERIOD {formatISK(totals.periodProfit)}</span>
          <span>CAP {formatISK(totals.capital)}</span>
          <button
            onClick={() => exportCSV(hubs)}
            className="px-2 py-0.5 rounded-sm border border-eve-border/60 text-eve-dim hover:text-eve-accent hover:border-eve-accent/40 transition-colors text-[10px]"
            title="Export to CSV"
          >
            CSV
          </button>
          <span className="text-eve-accent">{cacheBadge(cacheMeta)}</span>
        </div>
      </div>

      <div className="flex-1 min-h-0 overflow-auto">
        <table className="w-full min-w-[2800px] text-[12px] font-mono table-fixed">
          <thead className="sticky top-0 z-10 bg-eve-dark/95 border-b border-eve-border text-eve-dim uppercase tracking-wider text-[10px]">
            <tr>
              <th className="text-left px-2 py-1.5 w-[200px] cursor-pointer select-none hover:text-eve-accent" onClick={() => handleSort("name")}>
                Source Hub / Item {sortKey === "name" && <span className="text-[9px]">{sortDir === "desc" ? "▼" : "▲"}</span>}
              </th>
              <th className="text-right px-2 py-1.5 w-[52px]">Sec</th>
              <SortTh label="Buy Units"    sortKey="purchase_units"    className="w-[90px]"  {...thProps} />
              <SortTh label="Src Units"    sortKey="source_units"      className="w-[90px]"  {...thProps} />
              <SortTh label="Demand/Day"   sortKey="demand_per_day"    className="w-[100px]" {...thProps} />
              <SortTh label="Supply"       sortKey="supply_units"      className="w-[80px]"  {...thProps} />
              <SortTh label="DOS"          sortKey="dos"               className="w-[65px]"  {...thProps} />
              <SortTh label="Src Price"    sortKey="source_price"      className="w-[100px]" {...thProps} />
              <SortTh label="Tgt Now"      sortKey="target_now_price"  className="w-[100px]" {...thProps} />
              <SortTh label={`Tgt ${periodDays}d`} sortKey="target_period_price" className="w-[100px]" {...thProps} />
              <SortTh label="Now Profit"   sortKey="now_profit"        className="w-[110px]" {...thProps} />
              <SortTh label="Period Profit" sortKey="period_profit"    className="w-[115px]" {...thProps} />
              <SortTh label="ROI Now"      sortKey="roi_now"           className="w-[85px]"  {...thProps} />
              <SortTh label="ROI Period"   sortKey="roi_period"        className="w-[90px]"  {...thProps} />
              <SortTh label="Margin Now"   sortKey="margin_now"        className="w-[95px]"  {...thProps} />
              <SortTh label="Margin Period" sortKey="margin_period"    className="w-[100px]" {...thProps} />
              <SortTh label="Capital"      sortKey="capital"           className="w-[105px]" {...thProps} />
              <SortTh label="Shipping"     sortKey="shipping"          className="w-[95px]"  {...thProps} />
              <SortTh label="Jumps"        sortKey="jumps"             className="w-[60px]"  {...thProps} />
              <SortTh label="Vol m³"       sortKey="item_volume"       className="w-[70px]"  {...thProps} />
              <SortTh label="Items"        sortKey="item_count"        className="w-[60px]"  {...thProps} />
              <th className="text-left px-2 py-1.5 w-[150px]">Target</th>
            </tr>
          </thead>
          <tbody>
            {sortedHubs.map((hub) => {
              const isOpen = expanded.has(hub.source_system_id);
              const hubROINow = hub.capital_required + hub.shipping_cost > 0
                ? hub.target_now_profit / (hub.capital_required + hub.shipping_cost) * 100 : 0;
              const hubROIPeriod = hub.capital_required + hub.shipping_cost > 0
                ? hub.target_period_profit / (hub.capital_required + hub.shipping_cost) * 100 : 0;

              return (
                <Fragment key={`hub-block-${hub.source_system_id}`}>
                  <tr
                    className="border-b border-eve-border/60 bg-eve-panel/30 hover:bg-eve-panel/55 cursor-pointer"
                    onClick={() =>
                      setExpanded((prev) => {
                        const next = new Set(prev);
                        if (next.has(hub.source_system_id)) next.delete(hub.source_system_id);
                        else next.add(hub.source_system_id);
                        return next;
                      })
                    }
                  >
                    <td className="px-2 py-1.5 text-eve-text truncate">
                      <span className="text-eve-accent mr-1">{isOpen ? "▼" : "►"}</span>
                      {hub.source_system_name}
                    </td>
                    <td className="px-2 py-1.5 text-right">{hub.security.toFixed(1)}</td>
                    <td className="px-2 py-1.5 text-right">{hub.purchase_units.toLocaleString()}</td>
                    <td className="px-2 py-1.5 text-right">{hub.source_units.toLocaleString()}</td>
                    <td className="px-2 py-1.5 text-right">{Math.round(hub.target_demand_per_day).toLocaleString()}</td>
                    <td className="px-2 py-1.5 text-right">{hub.target_supply_units.toLocaleString()}</td>
                    <td className="px-2 py-1.5 text-right">{hub.target_dos.toFixed(2)}</td>
                    <td className="px-2 py-1.5 text-right text-eve-dim">—</td>
                    <td className="px-2 py-1.5 text-right text-eve-dim">—</td>
                    <td className="px-2 py-1.5 text-right text-eve-dim">—</td>
                    <td className="px-2 py-1.5 text-right text-eve-accent">{formatISK(hub.target_now_profit)}</td>
                    <td className="px-2 py-1.5 text-right text-eve-accent">{formatISK(hub.target_period_profit)}</td>
                    <td className="px-2 py-1.5 text-right">{formatMargin(hubROINow)}</td>
                    <td className="px-2 py-1.5 text-right">{formatMargin(hubROIPeriod)}</td>
                    <td className="px-2 py-1.5 text-right text-eve-dim">—</td>
                    <td className="px-2 py-1.5 text-right text-eve-dim">—</td>
                    <td className="px-2 py-1.5 text-right">{formatISK(hub.capital_required)}</td>
                    <td className="px-2 py-1.5 text-right">{formatISK(hub.shipping_cost)}</td>
                    <td className="px-2 py-1.5 text-right text-eve-dim">—</td>
                    <td className="px-2 py-1.5 text-right text-eve-dim">—</td>
                    <td className="px-2 py-1.5 text-right">{hub.item_count.toLocaleString()}</td>
                    <td className="px-2 py-1.5 text-eve-dim">Mixed</td>
                  </tr>

                  {isOpen &&
                    hub.items.map((item) => {
                      const signals = itemSignals(item);
                      return (
                        <tr
                          key={`item-${hub.source_system_id}-${item.type_id}-${item.target_system_id}`}
                          className="border-b border-eve-border/20 bg-eve-dark/30 hover:bg-eve-dark/55 cursor-pointer"
                          onClick={() => setDetail({ item, hubSecurity: hub.security })}
                        >
                          <td className="px-2 py-1.5 pl-6 text-eve-text truncate">
                            <span className="flex items-center gap-1 min-w-0">
                              <span className="truncate">{item.type_name}</span>
                              {signals.map((s, i) => (
                                <span
                                  key={i}
                                  title={s.title}
                                  className="shrink-0 px-1 py-0 rounded-sm border border-yellow-500/50 text-yellow-400 text-[9px] font-bold cursor-help"
                                >
                                  {s.label}
                                </span>
                              ))}
                            </span>
                          </td>
                          <td className="px-2 py-1.5 text-right">{hub.security.toFixed(1)}</td>
                          <td className="px-2 py-1.5 text-right">{item.purchase_units.toLocaleString()}</td>
                          <td className="px-2 py-1.5 text-right">{item.source_units.toLocaleString()}</td>
                          <td className="px-2 py-1.5 text-right">{item.target_demand_per_day.toFixed(1)}</td>
                          <td className="px-2 py-1.5 text-right">{item.target_supply_units.toLocaleString()}</td>
                          <td className={`px-2 py-1.5 text-right ${item.target_dos > 30 ? "text-yellow-400" : ""}`}>
                            {item.target_dos.toFixed(2)}
                          </td>
                          <td className="px-2 py-1.5 text-right">{formatISK(item.source_avg_price)}</td>
                          <td className="px-2 py-1.5 text-right">{formatISK(item.target_now_price)}</td>
                          <td className="px-2 py-1.5 text-right">{formatISK(item.target_period_price)}</td>
                          <td className="px-2 py-1.5 text-right">{formatISK(item.target_now_profit)}</td>
                          <td className="px-2 py-1.5 text-right text-eve-accent">{formatISK(item.target_period_profit)}</td>
                          <td className="px-2 py-1.5 text-right">{formatMargin(item.roi_now)}</td>
                          <td className="px-2 py-1.5 text-right text-eve-accent">{formatMargin(item.roi_period)}</td>
                          <td className="px-2 py-1.5 text-right">{formatMargin(item.margin_now)}</td>
                          <td className="px-2 py-1.5 text-right">{formatMargin(item.margin_period)}</td>
                          <td className="px-2 py-1.5 text-right">{formatISK(item.capital_required)}</td>
                          <td className="px-2 py-1.5 text-right">{formatISK(item.shipping_cost)}</td>
                          <td className="px-2 py-1.5 text-right">{item.jumps}</td>
                          <td className="px-2 py-1.5 text-right">{item.item_volume.toFixed(2)}</td>
                          <td className="px-2 py-1.5 text-right">1</td>
                          <td className="px-2 py-1.5 text-eve-dim truncate">{item.target_system_name}</td>
                        </tr>
                      );
                    })}
                </Fragment>
              );
            })}
          </tbody>
        </table>
      </div>

      {detail && (
        <DetailPanel
          item={detail.item}
          hubSecurity={detail.hubSecurity}
          onClose={() => setDetail(null)}
        />
      )}
    </div>
  );
}
