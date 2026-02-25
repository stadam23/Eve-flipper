import { useState, useMemo, useCallback, useRef, useEffect } from "react";
import type { ContractResult, StationCacheMeta } from "@/lib/types";
import { formatISK, formatMargin } from "@/lib/format";
import { useI18n, type TranslationKey } from "@/lib/i18n";
import { useGlobalToast } from "./Toast";
import { EmptyState, type EmptyReason } from "./EmptyState";
import {
  clearStationTradeStates,
  deleteStationTradeStates,
  getStationTradeStates,
  openContractInGame,
  rebootStationCache,
  setStationTradeState,
} from "@/lib/api";
import { handleEveUIError } from "@/lib/handleEveUIError";
import { ContractDetailsPopup } from "./ContractDetailsPopup";

type SortKey = keyof ContractResult;
type SortDir = "asc" | "desc";
type HiddenMode = "done" | "ignored";
type HiddenFilterTab = "all" | "done" | "ignored";

type HiddenContractEntry = {
  key: string;
  mode: HiddenMode;
  updatedAt: string;
  title: string;
  stationName: string;
  stateTypeID: number;
  stateStationID: number;
  stateRegionID: number;
};

type CacheMetaView = {
  currentRevision: number;
  lastRefreshAt: number;
  nextExpiryAt: number;
  scopeLabel: string;
  regionCount: number;
};

const CACHE_TTL_FALLBACK_MS = 20 * 60 * 1000;

interface Props {
  results: ContractResult[];
  scanning: boolean;
  progress: string;
  cacheMeta?: StationCacheMeta | null;
  tradeStateTab?: "contracts";
  excludeRigPriceIfShip?: boolean;
  /** When 0 results, show these filter hints (e.g. "Min price: 10M", "Max margin: 100%") */
  filterHints?: string[];
  isLoggedIn?: boolean;
}

const columnDefs: { key: SortKey; labelKey: TranslationKey; width: string; numeric: boolean }[] = [
  { key: "Title", labelKey: "colTitle", width: "min-w-[200px]", numeric: false },
  { key: "Price", labelKey: "colContractPrice", width: "min-w-[120px]", numeric: true },
  { key: "MarketValue", labelKey: "colMarketValue", width: "min-w-[120px]", numeric: true },
  { key: "ExpectedProfit", labelKey: "colContractExpectedProfit", width: "min-w-[120px]", numeric: true },
  { key: "Profit", labelKey: "colContractProfit", width: "min-w-[120px]", numeric: true },
  { key: "SellConfidence", labelKey: "colContractConfidence", width: "min-w-[95px]", numeric: true },
  { key: "EstLiquidationDays", labelKey: "colContractLiqDays", width: "min-w-[85px]", numeric: true },
  { key: "MarginPercent", labelKey: "colContractMargin", width: "min-w-[80px]", numeric: true },
  { key: "Volume", labelKey: "colVolume", width: "min-w-[80px]", numeric: true },
  { key: "StationName", labelKey: "colStation", width: "min-w-[180px]", numeric: false },
  { key: "SystemName", labelKey: "colContractSystem", width: "min-w-[120px]", numeric: false },
  { key: "RegionName", labelKey: "colContractRegion", width: "min-w-[120px]", numeric: false },
  { key: "LiquidationSystemName", labelKey: "colContractLiqSystem", width: "min-w-[140px]", numeric: false },
  { key: "ItemCount", labelKey: "colItems", width: "min-w-[70px]", numeric: true },
  { key: "ProfitPerJump", labelKey: "colContractPPJ", width: "min-w-[110px]", numeric: true },
  { key: "Jumps", labelKey: "colContractJumps", width: "min-w-[60px]", numeric: true },
];

function rowKey(row: ContractResult) {
  return `contract-${row.ContractID}`;
}

function hash53(input: string): number {
  let h1 = 0xdeadbeef ^ input.length;
  let h2 = 0x41c6ce57 ^ input.length;
  for (let i = 0; i < input.length; i++) {
    const ch = input.charCodeAt(i);
    h1 = Math.imul(h1 ^ ch, 2654435761);
    h2 = Math.imul(h2 ^ ch, 1597334677);
  }
  h1 = Math.imul(h1 ^ (h1 >>> 16), 2246822507) ^ Math.imul(h2 ^ (h2 >>> 13), 3266489909);
  h2 = Math.imul(h2 ^ (h2 >>> 16), 2246822507) ^ Math.imul(h1 ^ (h1 >>> 13), 3266489909);
  return 4294967296 * (2097151 & h2) + (h1 >>> 0);
}

function contractStateIDs(row: ContractResult): {
  typeID: number;
  stationID: number;
  regionID: number;
} {
  const base = rowKey(row);
  const h = hash53(base) || 1;
  return {
    typeID: Number((h % 2_147_483_000) + 1),
    stationID: h,
    regionID: 0,
  };
}

function tradeStateIndexKey(typeID: number, stationID: number, regionID: number): string {
  return `${typeID}:${stationID}:${regionID}`;
}

function formatCountdown(totalSec: number): string {
  const sec = Math.max(0, Math.floor(totalSec));
  const mm = Math.floor(sec / 60)
    .toString()
    .padStart(2, "0");
  const ss = (sec % 60).toString().padStart(2, "0");
  return `${mm}:${ss}`;
}

function mapServerCacheMeta(
  meta: StationCacheMeta | null | undefined,
  fallbackScope: string,
  fallbackRegionCount: number,
  fallbackBaseTs: number,
): CacheMetaView {
  if (!meta) {
    return {
      currentRevision: Math.floor(fallbackBaseTs / 1000),
      lastRefreshAt: fallbackBaseTs,
      nextExpiryAt: fallbackBaseTs + CACHE_TTL_FALLBACK_MS,
      scopeLabel: fallbackScope,
      regionCount: fallbackRegionCount,
    };
  }
  const lastRefreshTs = meta.last_refresh_at
    ? Date.parse(meta.last_refresh_at)
    : fallbackBaseTs;
  const nextExpiryTs = meta.next_expiry_at
    ? Date.parse(meta.next_expiry_at)
    : fallbackBaseTs + Math.max(60, meta.min_ttl_sec || 60) * 1000;
  return {
    currentRevision:
      meta.current_revision && Number.isFinite(meta.current_revision)
        ? meta.current_revision
        : Math.floor(nextExpiryTs / 1000),
    lastRefreshAt: Number.isFinite(lastRefreshTs) ? lastRefreshTs : fallbackBaseTs,
    nextExpiryAt: Number.isFinite(nextExpiryTs)
      ? nextExpiryTs
      : fallbackBaseTs + CACHE_TTL_FALLBACK_MS,
    scopeLabel: fallbackScope,
    regionCount: Math.max(1, fallbackRegionCount),
  };
}

function numericCellValue(row: ContractResult, key: SortKey): number {
  if (key === "MarginPercent") return row.ExpectedMarginPercent ?? row.MarginPercent ?? 0;
  if (key === "ExpectedProfit") return row.ExpectedProfit ?? row.Profit ?? 0;
  const val = row[key];
  return typeof val === "number" ? val : 0;
}

export function ContractResultsTable({
  results,
  scanning,
  progress,
  cacheMeta,
  tradeStateTab = "contracts",
  excludeRigPriceIfShip = true,
  filterHints,
  isLoggedIn = false,
}: Props) {
  const { t } = useI18n();
  const { addToast } = useGlobalToast();
  const emptyReason: EmptyReason = (results.length === 0 && filterHints && filterHints.length > 0)
    ? "filters_too_strict"
    : "no_scan_yet";

  const [sortKey, setSortKey] = useState<SortKey>("ExpectedProfit");
  const [sortDir, setSortDir] = useState<SortDir>("desc");
  const [filters, setFilters] = useState<Record<string, string>>({});
  const [showFilters, setShowFilters] = useState(false);
  const [showHiddenRows, setShowHiddenRows] = useState(false);
  const [hiddenMap, setHiddenMap] = useState<Record<string, HiddenContractEntry>>({});
  const [ignoredModalOpen, setIgnoredModalOpen] = useState(false);
  const [ignoredSearch, setIgnoredSearch] = useState("");
  const [ignoredTab, setIgnoredTab] = useState<HiddenFilterTab>("all");
  const [ignoredSelectedKeys, setIgnoredSelectedKeys] = useState<Set<string>>(new Set());
  const [cacheNowTs, setCacheNowTs] = useState<number>(Date.now());
  const [lastScanTs, setLastScanTs] = useState<number>(Date.now());
  const [cacheRebooting, setCacheRebooting] = useState(false);

  // Contract details popup
  const [selectedContract, setSelectedContract] = useState<ContractResult | null>(null);

  // Context menu
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; row: ContractResult } | null>(null);
  const contextMenuRef = useRef<HTMLDivElement>(null);

  const handleContextMenu = useCallback((e: React.MouseEvent, row: ContractResult) => {
    e.preventDefault();
    setContextMenu({ x: e.clientX, y: e.clientY, row });
  }, []);

  const copyText = (text: string) => {
    navigator.clipboard.writeText(text);
    addToast(t("copied"), "success", 2000);
    setContextMenu(null);
  };

  // Adjust context menu position
  useEffect(() => {
    if (contextMenu && contextMenuRef.current) {
      const menu = contextMenuRef.current;
      const rect = menu.getBoundingClientRect();
      const padding = 10;
      let x = contextMenu.x;
      let y = contextMenu.y;
      if (x + rect.width > window.innerWidth - padding) x = window.innerWidth - rect.width - padding;
      if (y + rect.height > window.innerHeight - padding) y = window.innerHeight - rect.height - padding;
      x = Math.max(padding, x);
      y = Math.max(padding, y);
      menu.style.left = `${x}px`;
      menu.style.top = `${y}px`;
    }
  }, [contextMenu]);

  useEffect(() => {
    if (!scanning && results.length > 0) {
      setLastScanTs(Date.now());
    }
  }, [results, scanning]);

  useEffect(() => {
    const timer = window.setInterval(() => {
      setCacheNowTs(Date.now());
    }, 1000);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    if (!ignoredModalOpen) {
      setIgnoredSearch("");
      setIgnoredTab("all");
      setIgnoredSelectedKeys(new Set());
    }
  }, [ignoredModalOpen]);

  useEffect(() => {
    setIgnoredSelectedKeys((prev) => {
      if (prev.size === 0) return prev;
      const next = new Set<string>();
      for (const key of prev) {
        if (hiddenMap[key]) next.add(key);
      }
      return next.size === prev.size ? prev : next;
    });
  }, [hiddenMap]);

  const filtered = useMemo(() => {
    if (Object.values(filters).every((v) => !v)) return results;
    return results.filter((row) => {
      for (const col of columnDefs) {
        const fval = filters[col.key];
        if (!fval) continue;
        const cellVal = row[col.key];
        if (col.numeric) {
          // Support filters: "100-500" (range), ">100", ">=100", "<500", "<=500", "=100" (exact), or plain number (>= threshold)
          const num = numericCellValue(row, col.key);
          const trimmed = fval.trim();
          if (trimmed.includes("-") && !trimmed.startsWith("-")) {
            // Range: "100-500"
            const [minS, maxS] = trimmed.split("-");
            const min = parseFloat(minS);
            const max = parseFloat(maxS);
            if (!isNaN(min) && !isNaN(max) && (num < min || num > max)) return false;
          } else if (trimmed.startsWith(">=")) {
            const min = parseFloat(trimmed.slice(2));
            if (!isNaN(min) && num < min) return false;
          } else if (trimmed.startsWith(">")) {
            const min = parseFloat(trimmed.slice(1));
            if (!isNaN(min) && num <= min) return false;
          } else if (trimmed.startsWith("<=")) {
            const max = parseFloat(trimmed.slice(2));
            if (!isNaN(max) && num > max) return false;
          } else if (trimmed.startsWith("<")) {
            const max = parseFloat(trimmed.slice(1));
            if (!isNaN(max) && num >= max) return false;
          } else if (trimmed.startsWith("=")) {
            // Exact match
            const target = parseFloat(trimmed.slice(1));
            if (!isNaN(target) && num !== target) return false;
          } else {
            // Plain number: treat as >= (minimum threshold)
            const min = parseFloat(trimmed);
            if (!isNaN(min) && num < min) return false;
          }
        } else {
          if (!String(cellVal).toLowerCase().includes(fval.toLowerCase())) return false;
        }
      }
      return true;
    });
  }, [results, filters]);

  const sorted = useMemo(() => {
    const copy = [...filtered];
    const currentCol = columnDefs.find((c) => c.key === sortKey);
    const numericSort = !!currentCol?.numeric;
    copy.sort((a, b) => {
      const av = a[sortKey];
      const bv = b[sortKey];
      if (numericSort) {
        const an = numericCellValue(a, sortKey);
        const bn = numericCellValue(b, sortKey);
        return sortDir === "asc" ? an - bn : bn - an;
      }
      return sortDir === "asc"
        ? String(av).localeCompare(String(bv))
        : String(bv).localeCompare(String(av));
    });
    return copy;
  }, [filtered, sortKey, sortDir]);

  const displaySorted = useMemo(() => {
    if (showHiddenRows) return sorted;
    return sorted.filter((row) => !hiddenMap[rowKey(row)]);
  }, [hiddenMap, showHiddenRows, sorted]);

  const cacheView = useMemo(
    () => mapServerCacheMeta(cacheMeta, t("hiddenScopeContractScan"), 1, lastScanTs),
    [cacheMeta, lastScanTs, t],
  );
  const cacheSecondsLeft = useMemo(
    () => Math.floor((cacheView.nextExpiryAt - cacheNowTs) / 1000),
    [cacheNowTs, cacheView.nextExpiryAt],
  );
  const cacheBadgeText = useMemo(() => {
    if (cacheSecondsLeft <= 0) return t("cacheStale");
    return t("cacheLabel", { time: formatCountdown(cacheSecondsLeft) });
  }, [cacheSecondsLeft, t]);

  const refreshHiddenStates = useCallback(
    async (currentRevision?: number) => {
      try {
        const resp = await getStationTradeStates({
          tab: tradeStateTab,
          currentRevision,
        });
        const states = Array.isArray(resp.states) ? resp.states : [];
        const byStateKey = new Map<string, ContractResult>();
        for (const row of results) {
          const ids = contractStateIDs(row);
          byStateKey.set(tradeStateIndexKey(ids.typeID, ids.stationID, ids.regionID), row);
        }
        setHiddenMap((prev) => {
          const next: Record<string, HiddenContractEntry> = {};
          for (const s of states) {
            const stateKey = tradeStateIndexKey(s.type_id, s.station_id, s.region_id);
            const row = byStateKey.get(stateKey);
            const key = row ? rowKey(row) : stateKey;
            const prevEntry = prev[key];
            next[key] = {
              key,
              mode: s.mode,
              updatedAt: s.updated_at,
              title: row?.Title ?? prevEntry?.title ?? t("hiddenContractFallback", { id: s.type_id }),
              stationName: row?.StationName ?? prevEntry?.stationName ?? t("hiddenUnknown"),
              stateTypeID: s.type_id,
              stateStationID: s.station_id,
              stateRegionID: s.region_id,
            };
          }
          return next;
        });
      } catch {
        // best effort
      }
    },
    [results, t, tradeStateTab],
  );

  useEffect(() => {
    if (scanning) return;
    void refreshHiddenStates(cacheView.currentRevision);
  }, [cacheView.currentRevision, refreshHiddenStates, scanning, results]);

  const summary = useMemo(() => {
    if (displaySorted.length === 0) return null;
    const totalProfit = displaySorted.reduce((sum, r) => sum + r.Profit, 0);
    const totalExpected = displaySorted.reduce((sum, r) => sum + (r.ExpectedProfit ?? r.Profit), 0);
    const avgMargin =
      displaySorted.reduce((sum, r) => sum + (r.ExpectedMarginPercent ?? r.MarginPercent), 0) /
      displaySorted.length;
    return { totalProfit, totalExpected, avgMargin, count: displaySorted.length };
  }, [displaySorted]);

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortKey(key);
      setSortDir("desc");
    }
  };

  const hasActiveFilters = Object.values(filters).some((v) => !!v);

  const formatCell = (col: (typeof columnDefs)[number], row: ContractResult): string => {
    const val = row[col.key];
    if (val == null || val === "") return "\u2014";
    if (
      col.key === "Price" ||
      col.key === "MarketValue" ||
      col.key === "Profit" ||
      col.key === "ExpectedProfit" ||
      col.key === "ProfitPerJump"
    ) {
      return formatISK(val as number);
    }
    if (col.key === "MarginPercent") return formatMargin((row.ExpectedMarginPercent ?? row.MarginPercent) as number);
    if (col.key === "SellConfidence") return `${(val as number).toFixed(1)}%`;
    if (col.key === "Volume") return (val as number).toFixed(1);
    if (col.key === "EstLiquidationDays") return (val as number).toFixed(1);
    if (typeof val === "number") return val.toLocaleString("ru-RU");
    return String(val);
  };

  const exportCSV = () => {
    const header = columnDefs.map((c) => t(c.labelKey)).join(",");
    const csvRows = displaySorted.map((row) =>
      columnDefs.map((col) => {
        const val = col.numeric ? numericCellValue(row, col.key) : row[col.key];
        const str = String(val);
        return str.includes(",") ? `"${str}"` : str;
      }).join(",")
    );
    const csv = [header, ...csvRows].join("\n");
    const blob = new Blob(["\uFEFF" + csv], { type: "text/csv;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `eve-contracts-${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const setRowHiddenState = useCallback(
    async (row: ContractResult, mode: HiddenMode) => {
      const key = rowKey(row);
      const ids = contractStateIDs(row);
      const entry: HiddenContractEntry = {
        key,
        mode,
        updatedAt: new Date().toISOString(),
        title: row.Title,
        stationName: row.StationName,
        stateTypeID: ids.typeID,
        stateStationID: ids.stationID,
        stateRegionID: ids.regionID,
      };
      setHiddenMap((prev) => ({ ...prev, [key]: entry }));
      setContextMenu(null);
      try {
        await setStationTradeState({
          tab: tradeStateTab,
          type_id: ids.typeID,
          station_id: ids.stationID,
          region_id: ids.regionID,
          mode,
          until_revision: mode === "done" ? cacheView.currentRevision : 0,
        });
      } catch {
        addToast(t("hiddenStateSaveFailed"), "error", 2600);
        void refreshHiddenStates(cacheView.currentRevision);
      }
    },
    [addToast, cacheView.currentRevision, refreshHiddenStates, t, tradeStateTab],
  );

  const unhideRowsByKeys = useCallback(
    async (keys: string[]) => {
      if (keys.length === 0) return;
      const unique = [...new Set(keys)];
      const payload = unique
        .map((k) => hiddenMap[k])
        .filter(Boolean)
        .map((e) => ({
          type_id: e.stateTypeID,
          station_id: e.stateStationID,
          region_id: e.stateRegionID,
        }));
      setHiddenMap((prev) => {
        const next = { ...prev };
        let changed = false;
        for (const key of unique) {
          if (next[key]) {
            delete next[key];
            changed = true;
          }
        }
        return changed ? next : prev;
      });
      setIgnoredSelectedKeys((prev) => {
        const next = new Set(prev);
        for (const key of unique) next.delete(key);
        return next;
      });
      try {
        if (payload.length > 0) {
          await deleteStationTradeStates({ tab: tradeStateTab, keys: payload });
        }
      } catch {
        addToast(t("hiddenStateUnhideFailed"), "error", 2600);
        void refreshHiddenStates(cacheView.currentRevision);
      }
    },
    [addToast, cacheView.currentRevision, hiddenMap, refreshHiddenStates, t, tradeStateTab],
  );

  const clearDoneHiddenRows = useCallback(async () => {
    const hasDone = Object.values(hiddenMap).some((h) => h.mode === "done");
    if (!hasDone) return;
    setHiddenMap((prev) => {
      const next: Record<string, HiddenContractEntry> = {};
      for (const [key, entry] of Object.entries(prev)) {
        if (entry.mode !== "done") next[key] = entry;
      }
      return next;
    });
    try {
      await clearStationTradeStates({ tab: tradeStateTab, mode: "done" });
    } catch {
      addToast(t("hiddenStateClearDoneFailed"), "error", 2600);
      void refreshHiddenStates(cacheView.currentRevision);
    }
  }, [addToast, cacheView.currentRevision, hiddenMap, refreshHiddenStates, t, tradeStateTab]);

  const clearAllHiddenRows = useCallback(async () => {
    if (Object.keys(hiddenMap).length === 0) return;
    setHiddenMap({});
    setIgnoredSelectedKeys(new Set());
    try {
      await clearStationTradeStates({ tab: tradeStateTab });
    } catch {
      addToast(t("hiddenStateClearAllFailed"), "error", 2600);
      void refreshHiddenStates(cacheView.currentRevision);
    }
  }, [addToast, cacheView.currentRevision, hiddenMap, refreshHiddenStates, t, tradeStateTab]);

  const handleRebootCache = useCallback(async () => {
    if (cacheRebooting) return;
    setCacheRebooting(true);
    try {
      const res = await rebootStationCache();
      setLastScanTs(Date.now());
      addToast(t("cacheRebooted", { count: res.cleared }), "success", 2400);
      addToast(t("cacheRebootRescanHint"), "info", 2600);
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : t("cacheRebootFailed");
      addToast(msg, "error", 2800);
    } finally {
      setCacheRebooting(false);
    }
  }, [addToast, cacheRebooting, t]);

  const hiddenEntries = useMemo(
    () =>
      Object.values(hiddenMap).sort((a, b) =>
        b.updatedAt.localeCompare(a.updatedAt),
      ),
    [hiddenMap],
  );
  const hiddenCounts = useMemo(() => {
    let done = 0;
    let ignored = 0;
    for (const row of hiddenEntries) {
      if (row.mode === "done") done++;
      if (row.mode === "ignored") ignored++;
    }
    return { total: hiddenEntries.length, done, ignored };
  }, [hiddenEntries]);
  const filteredHiddenEntries = useMemo(() => {
    const q = ignoredSearch.trim().toLowerCase();
    return hiddenEntries.filter((entry) => {
      if (ignoredTab !== "all" && entry.mode !== ignoredTab) return false;
      if (!q) return true;
      return (
        entry.title.toLowerCase().includes(q) ||
        entry.stationName.toLowerCase().includes(q)
      );
    });
  }, [hiddenEntries, ignoredSearch, ignoredTab]);
  const contextHiddenEntry = contextMenu
    ? hiddenMap[rowKey(contextMenu.row)]
    : undefined;

  return (
    <div className="flex-1 flex flex-col min-h-0">
      {/* Toolbar */}
      <div className="shrink-0 flex items-center gap-2 px-2 py-1.5 text-xs">
        <div className="flex items-center gap-2 text-eve-dim">
          {scanning ? (
            <span className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-eve-accent animate-pulse" />
              {progress}
            </span>
          ) : results.length > 0 ? (
            filtered.length !== results.length
              ? t("showing", { shown: filtered.length, total: results.length })
              : t("foundContracts", { count: results.length })
          ) : null}
          {!scanning && results.length > 0 && hiddenCounts.total > 0 && (
            <span className="text-eve-dim">
              |{" "}
              {t("hiddenVisibleSummary", {
                visible: displaySorted.length,
                hidden: hiddenCounts.total,
              })}
            </span>
          )}
          {!scanning && results.length > 0 && cacheSecondsLeft <= 0 && (
            <span className="text-red-300">| {t("cacheStaleHint")}</span>
          )}
        </div>
        <div className="flex-1" />
        {results.length > 0 && !scanning && (
          <>
            <label className="inline-flex items-center gap-1 px-2 py-0.5 rounded-sm border border-eve-border/60 bg-eve-dark/40 text-[11px] cursor-pointer">
              <input
                type="checkbox"
                checked={showHiddenRows}
                onChange={(e) => setShowHiddenRows(e.target.checked)}
                className="accent-eve-accent"
              />
              <span>{t("showHidden")}</span>
            </label>
            <button
              type="button"
              onClick={() => setIgnoredModalOpen(true)}
              className="px-2 py-0.5 rounded-sm border border-eve-border/60 bg-eve-dark/40 text-[11px] hover:border-eve-accent/50 hover:text-eve-accent transition-colors"
              title={t("hiddenOpenManagerTitle")}
            >
              {t("hiddenButton", { count: hiddenCounts.total })}
            </button>
            <button
              type="button"
              onClick={() => {
                void handleRebootCache();
              }}
              disabled={cacheRebooting}
              className={`px-2 py-0.5 rounded-sm border bg-eve-dark/40 text-[11px] transition-colors disabled:opacity-40 disabled:cursor-not-allowed ${
                cacheSecondsLeft <= 0
                  ? "border-red-500/60 text-red-300 hover:bg-red-900/20"
                  : "border-eve-border/60 text-eve-dim hover:border-eve-accent/50 hover:text-eve-accent"
              }`}
              title={t("cacheHardResetTitle")}
            >
              {cacheRebooting ? t("cacheRebooting") : t("cacheReboot")}
            </button>
            <button
              type="button"
              className={`px-2 py-0.5 rounded-sm border text-[11px] font-mono transition-colors ${
                cacheSecondsLeft <= 0
                  ? "border-red-500/50 text-red-300 bg-red-950/30"
                  : "border-eve-border/60 text-eve-accent bg-eve-dark/40 hover:border-eve-accent/50"
              }`}
              title={`${t("cacheTooltipScope")}: ${cacheView.scopeLabel}\n${t("cacheTooltipRegions")}: ${cacheView.regionCount}\n${t("cacheTooltipLastRefresh")}: ${new Date(cacheView.lastRefreshAt).toLocaleTimeString()}\n${t("cacheTooltipNextExpiry")}: ${new Date(cacheView.nextExpiryAt).toLocaleTimeString()}`}
            >
              {cacheBadgeText}
            </button>
          </>
        )}
        <button
          onClick={() => setShowFilters((v) => !v)}
          className={`px-2 py-0.5 rounded-sm text-xs font-medium transition-colors cursor-pointer
            ${showFilters ? "bg-eve-accent/20 text-eve-accent border border-eve-accent/30" : "text-eve-dim hover:text-eve-text border border-eve-border hover:border-eve-border-light"}`}
        >
          ⊞
        </button>
        {hasActiveFilters && (
          <button
            onClick={() => setFilters({})}
            className="px-2 py-0.5 rounded-sm text-xs font-medium text-eve-dim hover:text-eve-text border border-eve-border cursor-pointer"
          >
            ✕
          </button>
        )}
        {results.length > 0 && (
          <button
            onClick={exportCSV}
            className="px-2 py-0.5 rounded-sm text-xs font-medium text-eve-dim hover:text-eve-text border border-eve-border cursor-pointer"
          >
            CSV
          </button>
        )}
      </div>

      {/* Table */}
      <div className="flex-1 min-h-0 overflow-auto border border-eve-border rounded-sm table-scroll-wrapper table-scroll-container">
        <table className="w-full text-sm">
          <thead className="sticky top-0 z-10">
            <tr className="bg-eve-dark border-b border-eve-border">
              {columnDefs.map((col) => (
                <th
                  key={col.key}
                  onClick={() => toggleSort(col.key)}
                  className={`${col.width} px-3 py-2 text-left text-[11px] uppercase tracking-wider
                             text-eve-dim font-medium cursor-pointer select-none
                             hover:text-eve-accent transition-colors ${
                               sortKey === col.key ? "text-eve-accent" : ""
                             }`}
                >
                  {t(col.labelKey)}
                  {sortKey === col.key && (
                    <span className="ml-1">{sortDir === "asc" ? "▲" : "▼"}</span>
                  )}
                </th>
              ))}
            </tr>
            {showFilters && (
              <tr className="bg-eve-dark/80 border-b border-eve-border">
                {columnDefs.map((col) => (
                  <th key={col.key} className={`${col.width} px-1 py-1`}>
                    <input
                      type="text"
                      value={filters[col.key] ?? ""}
                      onChange={(e) => setFilters((f) => ({ ...f, [col.key]: e.target.value }))}
                      placeholder={col.numeric ? "e.g. >100" : t("filterPlaceholder")}
                      className="w-full px-2 py-0.5 bg-eve-input border border-eve-border rounded-sm
                                 text-eve-text text-xs font-mono placeholder:text-eve-dim/50
                                 focus:outline-none focus:border-eve-accent/50 transition-colors"
                    />
                  </th>
                ))}
              </tr>
            )}
          </thead>
          <tbody>
            {displaySorted.map((row, i) => (
              <tr
                key={rowKey(row)}
                onClick={() => setSelectedContract(row)}
                onContextMenu={(e) => handleContextMenu(e, row)}
                className={`border-b border-eve-border/50 hover:bg-eve-accent/5 transition-colors cursor-pointer ${
                  i % 2 === 0 ? "bg-eve-panel" : "bg-eve-dark"
                } ${hiddenMap[rowKey(row)] ? "opacity-60" : ""}`}
              >
                {columnDefs.map((col) => (
                  <td
                    key={col.key}
                    className={`px-3 py-1.5 ${col.width} truncate ${
                      col.numeric ? "text-eve-accent font-mono" : "text-eve-text"
                    }`}
                  >
                    {formatCell(col, row)}
                  </td>
                ))}
              </tr>
            ))}
            {displaySorted.length === 0 && !scanning && (
              <tr>
                <td colSpan={columnDefs.length} className="p-0">
                  {results.length > 0 && hiddenCounts.total > 0 && !showHiddenRows ? (
                    <div className="p-6 text-center text-sm text-eve-dim">
                      {t("hiddenAllRowsPrefix")}{" "}
                      <span className="text-eve-accent">{t("showHidden")}</span>{" "}
                      {t("hiddenAllRowsOrOpen")}{" "}
                      <span className="text-eve-accent">
                        {t("hiddenButton", { count: hiddenCounts.total })}
                      </span>
                      .
                    </div>
                  ) : (
                    <EmptyState
                      reason={emptyReason}
                      hints={filterHints}
                      wikiSlug="Contract-Arbitrage"
                    />
                  )}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Summary footer */}
      {summary && displaySorted.length > 0 && (
        <div className="shrink-0 flex items-center gap-6 px-3 py-1.5 border-t border-eve-border text-xs">
          <span className="text-eve-dim">
            {t("totalProfit")}:{" "}
            <span className="text-eve-accent font-mono font-semibold">{formatISK(summary.totalProfit)}</span>
          </span>
          <span className="text-eve-dim">
            {t("colContractExpectedProfit")}:{" "}
            <span className="text-eve-success font-mono font-semibold">{formatISK(summary.totalExpected)}</span>
          </span>
          <span className="text-eve-dim">
            {t("avgMargin")}:{" "}
            <span className="text-eve-accent font-mono font-semibold">{formatMargin(summary.avgMargin)}</span>
          </span>
        </div>
      )}

      {/* Context menu */}
      {contextMenu && (
        <>
          <div className="fixed inset-0 z-50" onClick={() => setContextMenu(null)} />
          <div
            ref={contextMenuRef}
            className="fixed z-50 bg-eve-panel border border-eve-border rounded-sm shadow-eve-glow-strong py-1 min-w-[200px]"
            style={{ left: contextMenu.x, top: contextMenu.y }}
          >
            <ContextItem label={t("copyItem")} onClick={() => copyText(contextMenu.row.Title)} />
            <ContextItem label={t("copyStation")} onClick={() => copyText(contextMenu.row.StationName)} />
            <ContextItem label={t("copyContractID")} onClick={() => copyText(String(contextMenu.row.ContractID))} />
            <div className="h-px bg-eve-border my-1" />
            {contextHiddenEntry ? (
              <ContextItem
                label={t("hiddenContextUnhide")}
                onClick={() => {
                  void unhideRowsByKeys([contextHiddenEntry.key]);
                  setContextMenu(null);
                }}
              />
            ) : (
              <>
                <ContextItem
                  label={t("hiddenContextMarkDone")}
                  onClick={() => {
                    void setRowHiddenState(contextMenu.row, "done");
                  }}
                />
                <ContextItem
                  label={t("hiddenContextIgnore")}
                  onClick={() => {
                    void setRowHiddenState(contextMenu.row, "ignored");
                  }}
                />
              </>
            )}
            <div className="h-px bg-eve-border my-1" />
            {/* EVE UI actions */}
            {isLoggedIn && (
              <>
                <ContextItem
                  label={`🎮 ${t("openContract")}`}
                  onClick={async () => {
                    try {
                      await openContractInGame(contextMenu.row.ContractID);
                      addToast(t("actionSuccess"), "success", 2000);
                    } catch (err: any) {
                      const { messageKey, duration } = handleEveUIError(err);
                      addToast(t(messageKey), "error", duration);
                    }
                    setContextMenu(null);
                  }}
                />
                <div className="h-px bg-eve-border my-1" />
              </>
            )}
            <ContextItem
              label={t("openInEveref")}
              onClick={() => { window.open(`https://everef.net/contract/${contextMenu.row.ContractID}`, "_blank"); setContextMenu(null); }}
            />
          </div>
        </>
      )}

      {ignoredModalOpen && (
        <>
          <div
            className="fixed inset-0 z-[60] bg-black/70"
            onClick={() => setIgnoredModalOpen(false)}
          />
          <div className="fixed z-[61] left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-[min(980px,92vw)] h-[min(680px,88vh)] rounded-sm border border-eve-border bg-eve-panel shadow-eve-glow-strong p-3 flex flex-col">
            <div className="flex items-center justify-between gap-3">
              <div>
                <h3 className="text-sm uppercase tracking-wider text-eve-text font-semibold">
                  {t("hiddenContractsTitle")}
                </h3>
                <p className="text-[11px] text-eve-dim mt-0.5">
                  {t("hiddenSummary", {
                    done: hiddenCounts.done,
                    ignored: hiddenCounts.ignored,
                    total: hiddenCounts.total,
                  })}
                </p>
              </div>
              <button
                type="button"
                onClick={() => setIgnoredModalOpen(false)}
                className="px-2 py-1 rounded-sm border border-eve-border/60 text-eve-dim hover:text-eve-accent hover:border-eve-accent/50 transition-colors text-xs"
              >
                {t("close")}
              </button>
            </div>

            <div className="mt-3 flex flex-wrap items-center gap-2">
              <input
                value={ignoredSearch}
                onChange={(e) => setIgnoredSearch(e.target.value)}
                placeholder={t("hiddenSearchTitleOrStation")}
                className="h-8 px-2 min-w-[240px] rounded-sm border border-eve-border bg-eve-input text-eve-text text-xs"
              />
              <div className="flex items-center gap-1">
                {(["all", "done", "ignored"] as HiddenFilterTab[]).map((tab) => (
                  <button
                    key={tab}
                    type="button"
                    onClick={() => setIgnoredTab(tab)}
                    className={`px-2 py-1 rounded-sm border text-xs uppercase tracking-wide transition-colors ${
                      ignoredTab === tab
                        ? "border-eve-accent text-eve-accent bg-eve-accent/10"
                        : "border-eve-border/60 text-eve-dim hover:border-eve-accent/40 hover:text-eve-text"
                    }`}
                  >
                    {tab === "all"
                      ? t("hiddenFilterAll")
                      : tab === "done"
                        ? t("hiddenFilterDone")
                        : t("hiddenFilterIgnored")}
                  </button>
                ))}
              </div>
              <div className="flex-1" />
              <button
                type="button"
                onClick={() => {
                  void unhideRowsByKeys([...ignoredSelectedKeys]);
                }}
                disabled={ignoredSelectedKeys.size === 0}
                className="px-2 py-1 rounded-sm border border-eve-accent/60 text-eve-accent hover:bg-eve-accent/10 transition-colors text-xs disabled:opacity-40 disabled:cursor-not-allowed"
              >
                {t("hiddenUnignoreSelected")}
              </button>
              <button
                type="button"
                onClick={() => {
                  void clearDoneHiddenRows();
                }}
                disabled={hiddenCounts.done === 0}
                className="px-2 py-1 rounded-sm border border-eve-border/60 text-eve-text hover:border-eve-accent/40 hover:text-eve-accent transition-colors text-xs disabled:opacity-40 disabled:cursor-not-allowed"
              >
                {t("hiddenClearDone")}
              </button>
              <button
                type="button"
                onClick={() => {
                  void clearAllHiddenRows();
                }}
                disabled={hiddenCounts.total === 0}
                className="px-2 py-1 rounded-sm border border-red-500/50 text-red-300 hover:bg-red-500/10 transition-colors text-xs disabled:opacity-40 disabled:cursor-not-allowed"
              >
                {t("hiddenClearAll")}
              </button>
            </div>

            <div className="mt-3 flex-1 min-h-0 border border-eve-border/60 rounded-sm overflow-auto eve-scrollbar">
              {filteredHiddenEntries.length > 0 ? (
                <table className="w-full text-xs">
                  <thead className="sticky top-0 bg-eve-dark/95 border-b border-eve-border/60">
                    <tr>
                      <th className="w-8 px-2 py-1 text-center">
                        <input
                          type="checkbox"
                          checked={
                            filteredHiddenEntries.length > 0 &&
                            filteredHiddenEntries.every((entry) =>
                              ignoredSelectedKeys.has(entry.key),
                            )
                          }
                          onChange={(e) => {
                            if (!e.target.checked) {
                              setIgnoredSelectedKeys(new Set());
                              return;
                            }
                            setIgnoredSelectedKeys(
                              new Set(filteredHiddenEntries.map((entry) => entry.key)),
                            );
                          }}
                          className="accent-eve-accent"
                        />
                      </th>
                      <th className="px-2 py-1 text-left text-eve-dim uppercase tracking-wide">
                        {t("colTitle")}
                      </th>
                      <th className="px-2 py-1 text-left text-eve-dim uppercase tracking-wide">
                        {t("colStation")}
                      </th>
                      <th className="px-2 py-1 text-left text-eve-dim uppercase tracking-wide">
                        {t("colType")}
                      </th>
                      <th className="px-2 py-1 text-left text-eve-dim uppercase tracking-wide">
                        {t("updated")}
                      </th>
                      <th className="px-2 py-1 text-right text-eve-dim uppercase tracking-wide">
                        {t("orderDeskAction")}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredHiddenEntries.map((entry, idx) => (
                      <tr
                        key={entry.key}
                        className={`border-b border-eve-border/30 ${
                          idx % 2 === 0 ? "bg-eve-panel" : "bg-eve-dark"
                        }`}
                      >
                        <td className="px-2 py-1 text-center">
                          <input
                            type="checkbox"
                            checked={ignoredSelectedKeys.has(entry.key)}
                            onChange={(e) => {
                              setIgnoredSelectedKeys((prev) => {
                                const next = new Set(prev);
                                if (e.target.checked) next.add(entry.key);
                                else next.delete(entry.key);
                                return next;
                              });
                            }}
                            className="accent-eve-accent"
                          />
                        </td>
                        <td className="px-2 py-1 text-eve-text truncate">{entry.title}</td>
                        <td className="px-2 py-1 text-eve-dim">{entry.stationName}</td>
                        <td className="px-2 py-1">
                          <span
                            className={`inline-flex items-center px-1.5 py-0.5 rounded-sm border text-[10px] uppercase tracking-wide ${
                              entry.mode === "ignored"
                                ? "border-red-500/40 text-red-300 bg-red-950/30"
                                : "border-eve-accent/40 text-eve-accent bg-eve-accent/10"
                            }`}
                          >
                            {entry.mode === "ignored"
                              ? t("hiddenFilterIgnored")
                              : t("hiddenFilterDone")}
                          </span>
                        </td>
                        <td className="px-2 py-1 text-eve-dim font-mono">
                          {new Date(entry.updatedAt).toLocaleString()}
                        </td>
                        <td className="px-2 py-1 text-right">
                          <button
                            type="button"
                            onClick={() => {
                              void unhideRowsByKeys([entry.key]);
                            }}
                            className="px-2 py-0.5 rounded-sm border border-eve-accent/60 text-eve-accent hover:bg-eve-accent/10 transition-colors text-[11px]"
                          >
                            {t("hiddenUnignore")}
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <div className="h-full flex items-center justify-center text-eve-dim text-xs">
                  {t("hiddenNoRowsForFilter")}
                </div>
              )}
            </div>
          </div>
        </>
      )}

      {/* Contract details popup */}
      <ContractDetailsPopup
        open={!!selectedContract}
        contractID={selectedContract?.ContractID ?? 0}
        contractTitle={selectedContract?.Title ?? ""}
        contractPrice={selectedContract?.Price ?? 0}
        contractMarketValue={selectedContract?.MarketValue}
        contractProfit={selectedContract?.Profit}
        excludeRigPriceIfShip={excludeRigPriceIfShip}
        pickupStationName={selectedContract?.StationName ?? ""}
        pickupSystemName={selectedContract?.SystemName ?? ""}
        pickupRegionName={selectedContract?.RegionName ?? ""}
        liquidationSystemName={selectedContract?.LiquidationSystemName ?? ""}
        liquidationRegionName={selectedContract?.LiquidationRegionName ?? ""}
        liquidationJumps={selectedContract?.LiquidationJumps}
        totalJumps={selectedContract?.Jumps}
        isLoggedIn={isLoggedIn}
        onClose={() => setSelectedContract(null)}
      />
    </div>
  );
}

function ContextItem({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <div
      onClick={onClick}
      className="px-4 py-1.5 text-sm text-eve-text hover:bg-eve-accent/20 hover:text-eve-accent cursor-pointer transition-colors"
    >
      {label}
    </div>
  );
}
