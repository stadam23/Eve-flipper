import { useState, useMemo, useCallback, useEffect, useRef } from "react";
import type { FlipResult, StationCacheMeta, WatchlistItem } from "@/lib/types";
import { formatISK, formatMargin } from "@/lib/format";
import { useI18n, type TranslationKey } from "@/lib/i18n";
import {
  addToWatchlist,
  clearStationTradeStates,
  deleteStationTradeStates,
  getStationTradeStates,
  getWatchlist,
  openMarketInGame,
  rebootStationCache,
  removeFromWatchlist,
  setStationTradeState,
  setWaypointInGame,
} from "@/lib/api";
import { useGlobalToast } from "./Toast";
import { EmptyState, type EmptyReason } from "./EmptyState";
import { ExecutionPlannerPopup } from "./ExecutionPlannerPopup";
import { handleEveUIError } from "@/lib/handleEveUIError";

const PAGE_SIZE = 100;
const CACHE_TTL_FALLBACK_MS = 20 * 60 * 1000;

type SortKey = keyof FlipResult;
type SortDir = "asc" | "desc";
type HiddenMode = "done" | "ignored";
type HiddenFilterTab = "all" | "done" | "ignored";

type HiddenFlipEntry = {
  key: string;
  mode: HiddenMode;
  updatedAt: string;
  typeName: string;
  buyStation: string;
  sellStation: string;
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

interface Props {
  results: FlipResult[];
  scanning: boolean;
  progress: string;
  cacheMeta?: StationCacheMeta | null;
  tradeStateTab?: "radius" | "region";
  scanCompletedWithZero?: boolean;
  salesTaxPercent?: number;
  brokerFeePercent?: number;
  splitTradeFees?: boolean;
  buyBrokerFeePercent?: number;
  sellBrokerFeePercent?: number;
  buySalesTaxPercent?: number;
  sellSalesTaxPercent?: number;
  showRegions?: boolean;
  isLoggedIn?: boolean;
}

type ColumnDef = {
  key: SortKey;
  labelKey: TranslationKey;
  width: string;
  numeric: boolean;
};

/* ─── Column definitions ─── */

const baseColumnDefs: ColumnDef[] = [
  {
    key: "TypeName",
    labelKey: "colItem",
    width: "min-w-[180px]",
    numeric: false,
  },
  {
    key: "BuyPrice",
    labelKey: "colBuyPrice",
    width: "min-w-[120px]",
    numeric: true,
  },
  {
    key: "BestAskQty",
    labelKey: "colBestAskQty",
    width: "min-w-[90px]",
    numeric: true,
  },
  {
    key: "ExpectedBuyPrice",
    labelKey: "colExpectedBuyPrice",
    width: "min-w-[120px]",
    numeric: true,
  },
  {
    key: "BuyStation",
    labelKey: "colBuyStation",
    width: "min-w-[150px]",
    numeric: false,
  },
  {
    key: "SellPrice",
    labelKey: "colSellPrice",
    width: "min-w-[120px]",
    numeric: true,
  },
  {
    key: "BestBidQty",
    labelKey: "colBestBidQty",
    width: "min-w-[90px]",
    numeric: true,
  },
  {
    key: "ExpectedSellPrice",
    labelKey: "colExpectedSellPrice",
    width: "min-w-[120px]",
    numeric: true,
  },
  {
    key: "SellStation",
    labelKey: "colSellStation",
    width: "min-w-[150px]",
    numeric: false,
  },
  {
    key: "MarginPercent",
    labelKey: "colMargin",
    width: "min-w-[80px]",
    numeric: true,
  },
  {
    key: "IskPerM3",
    labelKey: "colIskPerM3",
    width: "min-w-[90px]",
    numeric: true,
  },
  {
    key: "UnitsToBuy",
    labelKey: "colUnitsToBuy",
    width: "min-w-[80px]",
    numeric: true,
  },
  {
    key: "FilledQty",
    labelKey: "colFilledQty",
    width: "min-w-[80px]",
    numeric: true,
  },
  {
    key: "CanFill",
    labelKey: "colCanFill",
    width: "min-w-[70px]",
    numeric: false,
  },
  {
    key: "BuyOrderRemain",
    labelKey: "colAcceptQty",
    width: "min-w-[80px]",
    numeric: true,
  },
  {
    key: "RealProfit",
    labelKey: "colRealProfit",
    width: "min-w-[120px]",
    numeric: true,
  },
  {
    key: "TotalProfit",
    labelKey: "colProfit",
    width: "min-w-[120px]",
    numeric: true,
  },
  {
    key: "ExpectedProfit",
    labelKey: "colExpectedProfit",
    width: "min-w-[100px]",
    numeric: true,
  },
  {
    key: "ProfitPerJump",
    labelKey: "colProfitPerJump",
    width: "min-w-[110px]",
    numeric: true,
  },
  {
    key: "TotalJumps",
    labelKey: "colJumps",
    width: "min-w-[60px]",
    numeric: true,
  },
  {
    key: "DailyVolume",
    labelKey: "colDailyVolume",
    width: "min-w-[80px]",
    numeric: true,
  },
  {
    key: "S2BPerDay",
    labelKey: "colS2BPerDay",
    width: "min-w-[90px]",
    numeric: true,
  },
  {
    key: "BfSPerDay",
    labelKey: "colBfSPerDay",
    width: "min-w-[90px]",
    numeric: true,
  },
  {
    key: "S2BBfSRatio",
    labelKey: "colS2BBfSRatio",
    width: "min-w-[90px]",
    numeric: true,
  },
  {
    key: "DailyProfit",
    labelKey: "colDailyProfit",
    width: "min-w-[110px]",
    numeric: true,
  },
  {
    key: "PriceTrend",
    labelKey: "colPriceTrend",
    width: "min-w-[70px]",
    numeric: true,
  },
  {
    key: "BuyCompetitors",
    labelKey: "colBuyCompetitors",
    width: "min-w-[70px]",
    numeric: true,
  },
  {
    key: "SellCompetitors",
    labelKey: "colSellCompetitors",
    width: "min-w-[70px]",
    numeric: true,
  },
];

const regionColumnDefs: ColumnDef[] = [
  {
    key: "BuyRegionName" as SortKey,
    labelKey: "colBuyRegion" as TranslationKey,
    width: "min-w-[120px]",
    numeric: false,
  },
  {
    key: "SellRegionName" as SortKey,
    labelKey: "colSellRegion" as TranslationKey,
    width: "min-w-[120px]",
    numeric: false,
  },
];

function buildColumnDefs(showRegions: boolean): ColumnDef[] {
  if (!showRegions) return baseColumnDefs;
  const cols = [...baseColumnDefs];
  const sellIdx = cols.findIndex((c) => c.key === "SellStation");
  if (sellIdx >= 0) cols.splice(sellIdx + 1, 0, regionColumnDefs[1]);
  const buyIdx = cols.findIndex((c) => c.key === "BuyStation");
  if (buyIdx >= 0) cols.splice(buyIdx + 1, 0, regionColumnDefs[0]);
  return cols;
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

function flipStateKey(row: FlipResult): string {
  return [
    row.TypeID ?? 0,
    row.BuyLocationID ?? 0,
    row.SellLocationID ?? 0,
    row.BuySystemID ?? 0,
    row.SellSystemID ?? 0,
  ].join(":");
}

function flipStateIDs(row: FlipResult): {
  typeID: number;
  stationID: number;
  regionID: number;
} {
  const typeID =
    row.TypeID > 0 && row.TypeID < 2_147_483_647
      ? row.TypeID
      : (hash53(flipStateKey(row)) % 2_147_483_000) + 1;
  const stationID = hash53(`flip:${flipStateKey(row)}`) || 1;
  const regionID = row.BuyRegionID || row.SellRegionID || 0;
  return { typeID, stationID, regionID };
}

function tradeStateIndexKey(typeID: number, stationID: number, regionID: number): string {
  return `${typeID}:${stationID}:${regionID}`;
}

/* ─── Row identity ───
 * Stable per-row object id to avoid duplicate keys when data has collisions.
 */
let _nextRowId = 1;
const _rowIdMap = new WeakMap<FlipResult, number>();
function getRowId(row: FlipResult): number {
  let id = _rowIdMap.get(row);
  if (id == null) {
    id = _nextRowId++;
    _rowIdMap.set(row, id);
  }
  return id;
}

/* ─── IndexedRow: carries stable identity for rows ─── */
interface IndexedRow {
  id: number; // stable id from WeakMap
  row: FlipResult;
}

/* ─── Filter helpers ─── */

function passesNumericFilter(num: number, fval: string): boolean {
  const trimmed = fval.trim();
  if (!trimmed) return true;
  // Range: "100-500"
  if (trimmed.includes("-") && !trimmed.startsWith("-")) {
    const [minS, maxS] = trimmed.split("-");
    const mn = parseFloat(minS);
    const mx = parseFloat(maxS);
    if (!isNaN(mn) && !isNaN(mx) && (num < mn || num > mx)) return false;
    return true;
  }
  if (trimmed.startsWith(">=")) {
    const v = parseFloat(trimmed.slice(2));
    return isNaN(v) || num >= v;
  }
  if (trimmed.startsWith(">")) {
    const v = parseFloat(trimmed.slice(1));
    return isNaN(v) || num > v;
  }
  if (trimmed.startsWith("<=")) {
    const v = parseFloat(trimmed.slice(2));
    return isNaN(v) || num <= v;
  }
  if (trimmed.startsWith("<")) {
    const v = parseFloat(trimmed.slice(1));
    return isNaN(v) || num < v;
  }
  if (trimmed.startsWith("=")) {
    const v = parseFloat(trimmed.slice(1));
    return isNaN(v) || num === v;
  }
  // Plain number: >= threshold
  const mn = parseFloat(trimmed);
  return isNaN(mn) || num >= mn;
}

function passesTextFilter(val: unknown, fval: string): boolean {
  return String(val ?? "")
    .toLowerCase()
    .includes(fval.toLowerCase());
}

function rowProfitPerUnit(row: FlipResult): number {
  if (row.RealProfit != null && row.FilledQty != null && row.FilledQty > 0) {
    const realPerUnit = row.RealProfit / row.FilledQty;
    if (Number.isFinite(realPerUnit)) return realPerUnit;
  }
  const fallback = row.ProfitPerUnit;
  return Number.isFinite(fallback) ? fallback : 0;
}

function rowIskPerM3(row: FlipResult): number {
  const volume = Number(row.Volume);
  if (!Number.isFinite(volume) || volume <= 0) return 0;
  return rowProfitPerUnit(row) / volume;
}

function rowS2BPerDay(row: FlipResult): number {
  if (row.S2BPerDay != null && Number.isFinite(row.S2BPerDay)) {
    return row.S2BPerDay;
  }
  const total = Number(row.DailyVolume);
  if (!Number.isFinite(total) || total <= 0) return 0;
  const buyDepth = Number(row.BuyOrderRemain);
  const sellDepth = Number(row.SellOrderRemain);
  if (buyDepth <= 0 && sellDepth <= 0) return total / 2;
  if (buyDepth <= 0) return 0;
  if (sellDepth <= 0) return total;
  return (total * buyDepth) / (buyDepth + sellDepth);
}

function rowBfSPerDay(row: FlipResult): number {
  if (row.BfSPerDay != null && Number.isFinite(row.BfSPerDay)) {
    return row.BfSPerDay;
  }
  const total = Number(row.DailyVolume);
  if (!Number.isFinite(total) || total <= 0) return 0;
  const s2b = rowS2BPerDay(row);
  const bfs = total - s2b;
  return bfs > 0 ? bfs : 0;
}

function rowS2BBfSRatio(row: FlipResult): number {
  if (row.S2BBfSRatio != null && Number.isFinite(row.S2BBfSRatio)) {
    return row.S2BBfSRatio;
  }
  const bfs = rowBfSPerDay(row);
  if (bfs <= 0) return 0;
  return rowS2BPerDay(row) / bfs;
}

function getCellValue(row: FlipResult, key: SortKey): unknown {
  if (key === "IskPerM3") {
    if (row.IskPerM3 != null && Number.isFinite(row.IskPerM3)) {
      return row.IskPerM3;
    }
    return rowIskPerM3(row);
  }
  if (key === "S2BPerDay") return rowS2BPerDay(row);
  if (key === "BfSPerDay") return rowBfSPerDay(row);
  if (key === "S2BBfSRatio") return rowS2BBfSRatio(row);
  return row[key];
}

function passesFilter(row: FlipResult, col: ColumnDef, fval: string): boolean {
  if (!fval) return true;
  const cellVal = getCellValue(row, col.key);
  return col.numeric
    ? passesNumericFilter(cellVal as number, fval)
    : passesTextFilter(cellVal, fval);
}

/* ─── Cell formatting ─── */

function fmtCell(col: ColumnDef, row: FlipResult): string {
  const val = getCellValue(row, col.key);
  if (
    col.key === "ExpectedProfit" ||
    col.key === "RealProfit" ||
    col.key === "ExpectedBuyPrice" ||
    col.key === "ExpectedSellPrice"
  ) {
    if (val == null || Number.isNaN(val)) return "\u2014";
    if (Number(val) <= 0) return "\u2014";
    return formatISK(val as number);
  }
  if (col.key === "BestAskQty" || col.key === "BestBidQty") {
    if (val == null || Number(val) <= 0) return "\u2014";
    return Number(val).toLocaleString();
  }
  if (col.key === "CanFill") {
    if (val == null) return "\u2014";
    return val ? "✓" : "✕";
  }
  if (
    col.key === "BuyPrice" ||
    col.key === "SellPrice" ||
    col.key === "TotalProfit" ||
    col.key === "ProfitPerJump" ||
    col.key === "DailyProfit" ||
    col.key === "IskPerM3"
  ) {
    return formatISK(val as number);
  }
  if (col.key === "MarginPercent") return formatMargin(val as number);
  if (col.key === "S2BBfSRatio") {
    const ratio = Number(val);
    return Number.isFinite(ratio) ? ratio.toFixed(2) : "\u2014";
  }
  if (col.key === "PriceTrend") {
    const v = val as number;
    return (v >= 0 ? "+" : "") + v.toFixed(1) + "%";
  }
  if (typeof val === "number") return val.toLocaleString();
  return String(val ?? "");
}

/* ═══════════════════════════════════════════════════════════════════════
 * COMPONENT
 * ═══════════════════════════════════════════════════════════════════════ */

export function ScanResultsTable({
  results,
  scanning,
  progress,
  cacheMeta,
  tradeStateTab = "radius",
  scanCompletedWithZero,
  salesTaxPercent,
  brokerFeePercent,
  splitTradeFees,
  buyBrokerFeePercent,
  sellBrokerFeePercent,
  buySalesTaxPercent,
  sellSalesTaxPercent,
  showRegions = false,
  isLoggedIn = false,
}: Props) {
  const { t } = useI18n();
  const emptyReason: EmptyReason = scanCompletedWithZero
    ? "no_results"
    : "no_scan_yet";
  const { addToast } = useGlobalToast();

  const columnDefs = useMemo(() => buildColumnDefs(showRegions), [showRegions]);

  // ── State ──
  const [sortKey, setSortKey] = useState<SortKey>("RealProfit");
  const [sortDir, setSortDir] = useState<SortDir>("desc");
  const [filters, setFilters] = useState<Record<string, string>>({});
  const [showFilters, setShowFilters] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [pinnedIds, setPinnedIds] = useState<Set<number>>(new Set());
  const [page, setPage] = useState(0);
  const [compactMode, setCompactMode] = useState(false);
  const [showHiddenRows, setShowHiddenRows] = useState(false);
  const [hiddenMap, setHiddenMap] = useState<Record<string, HiddenFlipEntry>>({});
  const [ignoredModalOpen, setIgnoredModalOpen] = useState(false);
  const [ignoredSearch, setIgnoredSearch] = useState("");
  const [ignoredTab, setIgnoredTab] = useState<HiddenFilterTab>("all");
  const [ignoredSelectedKeys, setIgnoredSelectedKeys] = useState<Set<string>>(new Set());
  const [cacheNowTs, setCacheNowTs] = useState<number>(Date.now());
  const [lastScanTs, setLastScanTs] = useState<number>(Date.now());
  const [cacheRebooting, setCacheRebooting] = useState(false);

  // Watchlist
  const [watchlist, setWatchlist] = useState<WatchlistItem[]>([]);
  useEffect(() => {
    getWatchlist()
      .then(setWatchlist)
      .catch(() => {});
  }, []);
  const watchlistIds = useMemo(
    () => new Set(watchlist.map((w) => w.type_id)),
    [watchlist],
  );

  // Context menu
  const [contextMenu, setContextMenu] = useState<{
    x: number;
    y: number;
    id: number;
    row: FlipResult;
  } | null>(null);
  const contextMenuRef = useRef<HTMLDivElement>(null);
  const [execPlanRow, setExecPlanRow] = useState<FlipResult | null>(null);

  useEffect(() => {
    if (contextMenu && contextMenuRef.current) {
      const menu = contextMenuRef.current;
      const rect = menu.getBoundingClientRect();
      const pad = 10;
      let x = contextMenu.x,
        y = contextMenu.y;
      if (x + rect.width > window.innerWidth - pad)
        x = window.innerWidth - rect.width - pad;
      if (y + rect.height > window.innerHeight - pad)
        y = window.innerHeight - rect.height - pad;
      menu.style.left = `${Math.max(pad, x)}px`;
      menu.style.top = `${Math.max(pad, y)}px`;
    }
  }, [contextMenu]);

  // ── Data pipeline: index → filter → sort ──
  const { indexed, filtered, sorted, variantByRowId } = useMemo(() => {
    const indexed: IndexedRow[] = results.map((row) => ({
      id: getRowId(row),
      row,
    }));

    const hasFilters = Object.values(filters).some((v) => !!v);
    const filtered = hasFilters
      ? indexed.filter((ir) => {
          for (const col of columnDefs) {
            const fval = filters[col.key];
            if (!fval) continue;
            if (!passesFilter(ir.row, col, fval)) return false;
          }
          return true;
        })
      : indexed;

    const sorted = filtered.slice();
    sorted.sort((a, b) => {
      const aPin = pinnedIds.has(a.id);
      const bPin = pinnedIds.has(b.id);
      if (aPin !== bPin) return aPin ? -1 : 1;

      const av = getCellValue(a.row, sortKey);
      const bv = getCellValue(b.row, sortKey);
      if (typeof av === "number" || typeof bv === "number") {
        if (av == null && bv == null) return 0;
        if (av == null) return 1;
        if (bv == null) return -1;
        const diff = (av as number) - (bv as number);
        return sortDir === "asc" ? diff : -diff;
      }
      const cmp = String(av ?? "").localeCompare(String(bv ?? ""));
      return sortDir === "asc" ? cmp : -cmp;
    });

    const totalByType = new Map<number, number>();
    for (const ir of sorted) {
      totalByType.set(ir.row.TypeID, (totalByType.get(ir.row.TypeID) ?? 0) + 1);
    }
    const seenByType = new Map<number, number>();
    const variantByRowId = new Map<number, { index: number; total: number }>();
    for (const ir of sorted) {
      const total = totalByType.get(ir.row.TypeID) ?? 0;
      const index = (seenByType.get(ir.row.TypeID) ?? 0) + 1;
      seenByType.set(ir.row.TypeID, index);
      if (total > 1) {
        variantByRowId.set(ir.id, { index, total });
      }
    }

    return { indexed, filtered, sorted, variantByRowId };
  }, [results, filters, columnDefs, sortKey, sortDir, pinnedIds]);

  const displaySorted = useMemo(() => {
    if (showHiddenRows) return sorted;
    return sorted.filter((ir) => !hiddenMap[flipStateKey(ir.row)]);
  }, [sorted, showHiddenRows, hiddenMap]);

  const { pageRows, totalPages, safePage } = useMemo(() => {
    const totalPages = Math.max(1, Math.ceil(displaySorted.length / PAGE_SIZE));
    const safePage = Math.min(page, totalPages - 1);
    const pageRows = displaySorted.slice(
      safePage * PAGE_SIZE,
      (safePage + 1) * PAGE_SIZE,
    );
    return { pageRows, totalPages, safePage };
  }, [displaySorted, page]);

  // Reset page when data/filters/sort change
  useEffect(() => {
    setPage(0);
  }, [results, filters, sortKey, sortDir, showHiddenRows, hiddenMap]);

  // Reset selection/pins/context menu when results change
  useEffect(() => {
    setSelectedIds(new Set());
    setPinnedIds(new Set());
    setContextMenu(null);
    if (!scanning && results.length > 0) {
      setLastScanTs(Date.now());
    }
  }, [results]);

  // Drop filters for columns that are no longer visible
  useEffect(() => {
    const allowed = new Set(columnDefs.map((col) => col.key));
    setFilters((prev) => {
      let changed = false;
      const next: Record<string, string> = {};
      for (const [key, value] of Object.entries(prev)) {
        if (allowed.has(key as SortKey)) {
          next[key] = value;
        } else {
          changed = true;
        }
      }
      return changed ? next : prev;
    });
  }, [columnDefs]);

  // Prune selected rows that are hidden by filters
  useEffect(() => {
    if (selectedIds.size === 0) return;
    const visibleIds = new Set(displaySorted.map((ir) => ir.id));
    setSelectedIds((prev) => {
      if (prev.size === 0) return prev;
      const next = new Set([...prev].filter((id) => visibleIds.has(id)));
      return next.size === prev.size ? prev : next;
    });
  }, [displaySorted, selectedIds.size]);

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

  useEffect(() => {
    const timer = window.setInterval(() => {
      setCacheNowTs(Date.now());
    }, 1000);
    return () => window.clearInterval(timer);
  }, []);

  const cacheView = useMemo(() => {
    const scopeLabel =
      tradeStateTab === "region"
        ? t("hiddenScopeRegionScan")
        : t("hiddenScopeRadiusScan");
    const fallbackRegionCount = showRegions ? 2 : 1;
    return mapServerCacheMeta(cacheMeta, scopeLabel, fallbackRegionCount, lastScanTs);
  }, [cacheMeta, lastScanTs, showRegions, t, tradeStateTab]);

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
        const byStateKey = new Map<string, IndexedRow>();
        for (const ir of indexed) {
          const ids = flipStateIDs(ir.row);
          byStateKey.set(tradeStateIndexKey(ids.typeID, ids.stationID, ids.regionID), ir);
        }
        setHiddenMap((prev) => {
          const next: Record<string, HiddenFlipEntry> = {};
          for (const s of states) {
            const stateKey = tradeStateIndexKey(s.type_id, s.station_id, s.region_id);
            const ir = byStateKey.get(stateKey);
            const key = ir ? flipStateKey(ir.row) : stateKey;
            const prevEntry = prev[key];
            next[key] = {
              key,
              mode: s.mode,
              updatedAt: s.updated_at,
              typeName: ir?.row.TypeName ?? prevEntry?.typeName ?? t("hiddenTypeFallback", { id: s.type_id }),
              buyStation: ir?.row.BuyStation ?? prevEntry?.buyStation ?? t("hiddenUnknown"),
              sellStation: ir?.row.SellStation ?? prevEntry?.sellStation ?? t("hiddenUnknown"),
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
    [indexed, t, tradeStateTab],
  );

  useEffect(() => {
    if (scanning) return;
    void refreshHiddenStates(cacheView.currentRevision);
  }, [cacheView.currentRevision, refreshHiddenStates, scanning, results]);

  // ── Summary stats ──
  const summary = useMemo(() => {
    const rows =
      selectedIds.size > 0
        ? displaySorted.filter((ir) => selectedIds.has(ir.id))
        : displaySorted;
    if (rows.length === 0) return null;
    const totalProfit = rows.reduce(
      (s, ir) => s + (ir.row.RealProfit ?? ir.row.ExpectedProfit ?? ir.row.TotalProfit),
      0,
    );
    const avgMargin =
      rows.reduce((s, ir) => s + ir.row.MarginPercent, 0) / rows.length;
    return { totalProfit, avgMargin, count: rows.length };
  }, [displaySorted, selectedIds]);

  // ── Callbacks ──
  const toggleSort = useCallback(
    (key: SortKey) => {
      if (key === sortKey) {
        setSortDir((d) => (d === "asc" ? "desc" : "asc"));
      } else {
        setSortKey(key);
        setSortDir("desc");
      }
    },
    [sortKey],
  );

  const setFilter = useCallback((key: string, value: string) => {
    setFilters((f) => ({ ...f, [key]: value }));
  }, []);

  const clearFilters = useCallback(() => {
    setFilters({});
  }, []);
  const hasActiveFilters = Object.values(filters).some((v) => !!v);

  const toggleSelect = useCallback((id: number) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const toggleSelectAll = useCallback(() => {
    setSelectedIds((prev) => {
      if (prev.size === displaySorted.length) return new Set();
      return new Set(displaySorted.map((ir) => ir.id));
    });
  }, [displaySorted]);

  const togglePin = useCallback((id: number) => {
    setPinnedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const handleContextMenu = useCallback(
    (e: React.MouseEvent, id: number, row: FlipResult) => {
      e.preventDefault();
      setContextMenu({ x: e.clientX, y: e.clientY, id, row });
    },
    [],
  );

  const copyText = useCallback(
    (text: string) => {
      navigator.clipboard.writeText(text);
      addToast(t("copied"), "success", 2000);
      setContextMenu(null);
    },
    [addToast, t],
  );

  const setRowHiddenState = useCallback(
    async (row: FlipResult, mode: HiddenMode) => {
      const key = flipStateKey(row);
      const ids = flipStateIDs(row);
      const entry: HiddenFlipEntry = {
        key,
        mode,
        updatedAt: new Date().toISOString(),
        typeName: row.TypeName,
        buyStation: row.BuyStation,
        sellStation: row.SellStation,
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
      const next: Record<string, HiddenFlipEntry> = {};
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
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : t("cacheRebootFailed");
      addToast(msg, "error", 2800);
    } finally {
      setCacheRebooting(false);
    }
  }, [addToast, cacheRebooting, t]);

  const exportCSV = useCallback(() => {
    const rows =
      selectedIds.size > 0
        ? displaySorted.filter((ir) => selectedIds.has(ir.id))
        : displaySorted;
    const header = columnDefs.map((c) => t(c.labelKey)).join(",");
    const csvRows = rows.map((ir) =>
      columnDefs
        .map((col) => {
          const str = String(getCellValue(ir.row, col.key) ?? "");
          return str.includes(",") ? `"${str}"` : str;
        })
        .join(","),
    );
    const csv = [header, ...csvRows].join("\n");
    const blob = new Blob(["\uFEFF" + csv], { type: "text/csv;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `eve-flipper-${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    URL.revokeObjectURL(url);
    addToast(`${t("exportCSV")}: ${rows.length} rows`, "success", 2000);
  }, [displaySorted, selectedIds, columnDefs, addToast, t]);

  const copyTable = useCallback(() => {
    const rows =
      selectedIds.size > 0
        ? displaySorted.filter((ir) => selectedIds.has(ir.id))
        : displaySorted;
    const header = columnDefs.map((c) => t(c.labelKey)).join("\t");
    const tsv = rows.map((ir) =>
      columnDefs.map((col) => fmtCell(col, ir.row)).join("\t"),
    );
    navigator.clipboard.writeText([header, ...tsv].join("\n"));
    addToast(t("copied"), "success", 2000);
  }, [displaySorted, selectedIds, columnDefs, addToast, t]);

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
        entry.typeName.toLowerCase().includes(q) ||
        entry.buyStation.toLowerCase().includes(q) ||
        entry.sellStation.toLowerCase().includes(q)
      );
    });
  }, [hiddenEntries, ignoredSearch, ignoredTab]);

  const contextHiddenEntry = contextMenu
    ? hiddenMap[flipStateKey(contextMenu.row)]
    : undefined;

  // ── Render ──
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
            filtered.length !== indexed.length ? (
              t("showing", { shown: filtered.length, total: indexed.length })
            ) : (
              t("foundDeals", { count: indexed.length })
            )
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
          {pinnedIds.size > 0 && (
            <span className="text-eve-accent">
              📌 {t("pinned", { count: pinnedIds.size })}
            </span>
          )}
          {selectedIds.size > 0 && (
            <span className="text-eve-accent">
              {t("selected", { count: selectedIds.size })}
            </span>
          )}
        </div>

        <div className="flex-1" />

        {/* Pagination */}
        {displaySorted.length > PAGE_SIZE && (
          <div className="flex items-center gap-1 text-eve-dim">
            <button
              onClick={() => setPage(0)}
              disabled={safePage === 0}
              className="px-1.5 py-0.5 rounded-sm hover:text-eve-text disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
            >
              «
            </button>
            <button
              onClick={() => setPage((p) => Math.max(0, p - 1))}
              disabled={safePage === 0}
              className="px-1.5 py-0.5 rounded-sm hover:text-eve-text disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
            >
              ‹
            </button>
            <span className="px-2 text-eve-text font-mono tabular-nums">
              {safePage + 1} / {totalPages}
            </span>
            <button
              onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
              disabled={safePage >= totalPages - 1}
              className="px-1.5 py-0.5 rounded-sm hover:text-eve-text disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
            >
              ›
            </button>
            <button
              onClick={() => setPage(totalPages - 1)}
              disabled={safePage >= totalPages - 1}
              className="px-1.5 py-0.5 rounded-sm hover:text-eve-text disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
            >
              »
            </button>
          </div>
        )}

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
              className="px-2 py-0.5 rounded-sm border border-eve-border/60 bg-eve-dark/40 text-[11px] text-eve-dim hover:border-eve-accent/50 hover:text-eve-accent transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
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

        {/* Action buttons */}
        <ToolbarBtn
          label="⊞"
          title={showFilters ? t("clearFilters") : t("filterPlaceholder")}
          active={showFilters}
          onClick={() => setShowFilters((v) => !v)}
        />
        {hasActiveFilters && (
          <ToolbarBtn
            label="✕"
            title={t("clearFilters")}
            onClick={clearFilters}
          />
        )}
        {results.length > 0 && (
          <>
            <ToolbarBtn
              label={compactMode ? "⊞" : "⊟"}
              title={compactMode ? t("comfyRows") : t("compactRows")}
              active={compactMode}
              onClick={() => setCompactMode((v) => !v)}
            />
            <ToolbarBtn
              label="CSV"
              title={t("exportCSV")}
              onClick={exportCSV}
            />
            <ToolbarBtn
              label="⎘"
              title={t("copyTable")}
              onClick={copyTable}
            />
          </>
        )}
      </div>

      {/* Table */}
      <div className="flex-1 min-h-0 flex flex-col border border-eve-border rounded-sm overflow-auto table-scroll-container">
        <table className="w-full text-sm">
          <thead className="sticky top-0 z-10">
            <tr className="bg-eve-dark border-b border-eve-border">
              <th className="w-8 px-1 py-2 text-center">
                <input
                  type="checkbox"
                  checked={
                    displaySorted.length > 0 &&
                    selectedIds.size === displaySorted.length
                  }
                  onChange={toggleSelectAll}
                  className="accent-eve-accent cursor-pointer"
                />
              </th>
              <th className="w-8 px-1 py-2" />
              {columnDefs.map((col) => (
                <th
                  key={col.key}
                  onClick={() => toggleSort(col.key)}
                  className={`${col.width} px-3 py-2 text-left text-[11px] uppercase tracking-wider text-eve-dim font-medium cursor-pointer select-none hover:text-eve-accent transition-colors ${sortKey === col.key ? "text-eve-accent" : ""}`}
                >
                  {t(col.labelKey)}
                  {sortKey === col.key && (
                    <span className="ml-1">
                      {sortDir === "asc" ? "▲" : "▼"}
                    </span>
                  )}
                </th>
              ))}
            </tr>
            {showFilters && (
              <tr className="bg-eve-dark/80 border-b border-eve-border">
                <th className="w-8" />
                <th className="w-8" />
                {columnDefs.map((col) => (
                  <th key={col.key} className={`${col.width} px-1 py-1`}>
                    <input
                      type="text"
                      value={filters[col.key] ?? ""}
                      onChange={(e) => setFilter(col.key, e.target.value)}
                      placeholder={
                        col.numeric ? "e.g. >100" : t("filterPlaceholder")
                      }
                      className="w-full px-2 py-0.5 bg-eve-input border border-eve-border rounded-sm text-eve-text text-xs font-mono placeholder:text-eve-dim/50 focus:outline-none focus:border-eve-accent/50 transition-colors"
                    />
                  </th>
                ))}
              </tr>
            )}
          </thead>
          <tbody>
            {pageRows.map((ir, i) => {
              const isPinned = pinnedIds.has(ir.id);
              const isSelected = selectedIds.has(ir.id);
              const globalIdx = safePage * PAGE_SIZE + i;
              const variant = variantByRowId.get(ir.id);
              const rowHidden = hiddenMap[flipStateKey(ir.row)];
              return (
                <tr
                  key={ir.id}
                  onContextMenu={(e) => handleContextMenu(e, ir.id, ir.row)}
                  className={`border-b border-eve-border/50 hover:bg-eve-accent/5 transition-colors ${compactMode ? "text-xs" : ""} ${
                    isPinned
                      ? "bg-eve-accent/10 border-l-2 border-l-eve-accent"
                      : isSelected
                        ? "bg-eve-accent/5"
                        : globalIdx % 2 === 0
                          ? "bg-eve-panel"
                          : "bg-eve-dark"
                  } ${rowHidden ? "opacity-60" : ""}`}
                >
                  <td
                    className={`w-8 px-1 text-center ${compactMode ? "py-1" : "py-1.5"}`}
                  >
                    <input
                      type="checkbox"
                      checked={isSelected}
                      onChange={() => toggleSelect(ir.id)}
                      className="accent-eve-accent cursor-pointer"
                    />
                  </td>
                  <td
                    className={`w-8 px-1 text-center ${compactMode ? "py-1" : "py-1.5"}`}
                  >
                    <button
                      onClick={() => togglePin(ir.id)}
                      className={`text-xs cursor-pointer transition-opacity ${isPinned ? "opacity-100" : "opacity-30 hover:opacity-70"}`}
                      title={isPinned ? t("unpinRow") : t("pinRow")}
                    >
                      {"📌"}
                    </button>
                  </td>
                  {columnDefs.map((col) => (
                    <td
                      key={col.key}
                      className={`px-3 ${compactMode ? "py-1" : "py-1.5"} ${col.width} ${col.key === "TypeName" ? "" : "truncate"} ${col.numeric ? "text-eve-accent font-mono" : "text-eve-text"}`}
                    >
                      {col.key === "TypeName" ? (
                        <div className="flex items-center gap-1.5 min-w-0">
                          <span className="truncate">{ir.row.TypeName}</span>
                          {variant && (
                            <span
                              title={t("variantChipHint")}
                              className="shrink-0 inline-flex items-center px-1 py-px rounded-[2px] border border-eve-accent/35 bg-eve-accent/10 text-eve-accent text-[9px] leading-none font-medium uppercase tracking-normal"
                            >
                              {t("variantChip", {
                                index: variant.index,
                                total: variant.total,
                              })}
                            </span>
                          )}
                        </div>
                      ) : (
                        fmtCell(col, ir.row)
                      )}
                    </td>
                  ))}
                </tr>
              );
            })}
            {displaySorted.length === 0 && !scanning && (
              <tr>
                <td colSpan={columnDefs.length + 2} className="p-0">
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
                    <EmptyState reason={emptyReason} wikiSlug="Getting-Started" />
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
            {t("colRealProfit")}:{" "}
            <span className="text-eve-accent font-mono font-semibold">
              {formatISK(summary.totalProfit)}
            </span>
          </span>
          <span className="text-eve-dim">
            {t("avgMargin")}:{" "}
            <span className="text-eve-accent font-mono font-semibold">
              {formatMargin(summary.avgMargin)}
            </span>
          </span>
          {selectedIds.size > 0 && (
            <span className="text-eve-dim italic">
              ({t("selected", { count: selectedIds.size })})
            </span>
          )}
        </div>
      )}

      {/* Context menu */}
      {contextMenu && (
        <>
          <div
            className="fixed inset-0 z-50"
            onClick={() => setContextMenu(null)}
          />
          <div
            ref={contextMenuRef}
            className="fixed z-50 bg-eve-panel border border-eve-border rounded-sm shadow-eve-glow-strong py-1 min-w-[200px]"
            style={{ left: contextMenu.x, top: contextMenu.y }}
          >
            <ContextItem
              label={t("copyItem")}
              onClick={() => copyText(contextMenu.row.TypeName ?? "")}
            />
            <ContextItem
              label={t("copyBuyStation")}
              onClick={() => copyText(contextMenu.row.BuyStation ?? "")}
            />
            <ContextItem
              label={t("copySellStation")}
              onClick={() => copyText(contextMenu.row.SellStation ?? "")}
            />
            <ContextItem
              label={t("copyTradeRoute")}
              onClick={() =>
                copyText(
                  `Buy: ${contextMenu.row.TypeName} x${contextMenu.row.UnitsToBuy} @ ${contextMenu.row.BuyStation} \u2192 Sell: @ ${contextMenu.row.SellStation}`,
                )
              }
            />
            <ContextItem
              label={t("copySystemAutopilot")}
              onClick={() => copyText(contextMenu.row.BuySystemName)}
            />
            <div className="h-px bg-eve-border my-1" />
            <ContextItem
              label={t("openInEveref")}
              onClick={() => {
                window.open(
                  `https://everef.net/type/${contextMenu.row.TypeID}`,
                  "_blank",
                );
                setContextMenu(null);
              }}
            />
            <ContextItem
              label={t("openInJitaSpace")}
              onClick={() => {
                window.open(
                  `https://www.jita.space/market/${contextMenu.row.TypeID}`,
                  "_blank",
                );
                setContextMenu(null);
              }}
            />
            <div className="h-px bg-eve-border my-1" />
            <ContextItem
              label={
                watchlistIds.has(contextMenu.row.TypeID)
                  ? t("untrackItem")
                  : `\u2B50 ${t("trackItem")}`
              }
              onClick={() => {
                const row = contextMenu.row;
                if (watchlistIds.has(row.TypeID)) {
                  removeFromWatchlist(row.TypeID)
                    .then(setWatchlist)
                    .then(() =>
                      addToast(t("watchlistRemoved"), "success", 2000),
                    )
                    .catch(() => addToast(t("watchlistError"), "error", 3000));
                } else {
                  addToWatchlist(row.TypeID, row.TypeName)
                    .then((r) => {
                      setWatchlist(r.items);
                      addToast(
                        r.inserted
                          ? t("watchlistItemAdded")
                          : t("watchlistAlready"),
                        r.inserted ? "success" : "info",
                        2000,
                      );
                    })
                    .catch(() => addToast(t("watchlistError"), "error", 3000));
                }
                setContextMenu(null);
              }}
            />
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
            {(contextMenu.row.BuyRegionID != null ||
              contextMenu.row.SellRegionID != null) && (
              <ContextItem
                label={t("placeDraft")}
                onClick={() => {
                  setExecPlanRow(contextMenu.row);
                  setContextMenu(null);
                }}
              />
            )}
            {/* EVE UI actions */}
            {isLoggedIn && (
              <>
                <div className="h-px bg-eve-border my-1" />
                <ContextItem
                  label={`🎮 ${t("openMarket")}`}
                  onClick={async () => {
                    try {
                      await openMarketInGame(contextMenu.row.TypeID);
                      addToast(t("actionSuccess"), "success", 2000);
                    } catch (err: any) {
                      const { messageKey, duration } = handleEveUIError(err);
                      addToast(t(messageKey), "error", duration);
                    }
                    setContextMenu(null);
                  }}
                />
                <ContextItem
                  label={`🎯 ${t("setDestination")} (Buy)`}
                  onClick={async () => {
                    try {
                      await setWaypointInGame(contextMenu.row.BuySystemID);
                      addToast(t("actionSuccess"), "success", 2000);
                    } catch (err: any) {
                      const { messageKey, duration } = handleEveUIError(err);
                      addToast(t(messageKey), "error", duration);
                    }
                    setContextMenu(null);
                  }}
                />
                {contextMenu.row.SellSystemID !== contextMenu.row.BuySystemID && (
                  <ContextItem
                    label={`🎯 ${t("setDestination")} (Sell)`}
                    onClick={async () => {
                      try {
                        await setWaypointInGame(contextMenu.row.SellSystemID);
                        addToast(t("actionSuccess"), "success", 2000);
                      } catch (err: any) {
                        addToast(t("actionFailed").replace("{error}", err.message), "error", 3000);
                      }
                      setContextMenu(null);
                    }}
                  />
                )}
              </>
            )}
            <div className="h-px bg-eve-border my-1" />
            <ContextItem
              label={
                pinnedIds.has(contextMenu.id) ? t("unpinRow") : t("pinRow")
              }
              onClick={() => {
                togglePin(contextMenu.id);
                setContextMenu(null);
              }}
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
                  {t("hiddenDealsTitle")}
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
                placeholder={t("hiddenSearchItemOrStation")}
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
                        {t("colItem")}
                      </th>
                      <th className="px-2 py-1 text-left text-eve-dim uppercase tracking-wide">
                        {t("hiddenColRoute")}
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
                        <td className="px-2 py-1 text-eve-text">{entry.typeName}</td>
                        <td className="px-2 py-1 text-eve-dim truncate">
                          {`${entry.buyStation} -> ${entry.sellStation}`}
                        </td>
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

      <ExecutionPlannerPopup
        open={execPlanRow !== null}
        onClose={() => setExecPlanRow(null)}
        typeID={execPlanRow?.TypeID ?? 0}
        typeName={execPlanRow?.TypeName ?? ""}
        regionID={execPlanRow?.BuyRegionID ?? 0}
        locationID={execPlanRow?.BuyLocationID ?? 0}
        sellRegionID={execPlanRow?.SellRegionID}
        sellLocationID={execPlanRow?.SellLocationID ?? 0}
        defaultQuantity={execPlanRow?.UnitsToBuy ?? 100}
        brokerFeePercent={brokerFeePercent}
        salesTaxPercent={salesTaxPercent}
        splitTradeFees={splitTradeFees}
        buyBrokerFeePercent={buyBrokerFeePercent}
        sellBrokerFeePercent={sellBrokerFeePercent}
        buySalesTaxPercent={buySalesTaxPercent}
        sellSalesTaxPercent={sellSalesTaxPercent}
      />
    </div>
  );
}

/* ─── Small reusable pieces ─── */

function ToolbarBtn({
  label,
  title,
  active,
  onClick,
}: {
  label: string;
  title: string;
  active?: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      title={title}
      className={`px-2 py-0.5 rounded-sm text-xs font-medium transition-colors cursor-pointer ${
        active
          ? "bg-eve-accent/20 text-eve-accent border border-eve-accent/30"
          : "text-eve-dim hover:text-eve-text border border-eve-border hover:border-eve-border-light"
      }`}
    >
      {label}
    </button>
  );
}

function ContextItem({
  label,
  onClick,
}: {
  label: string;
  onClick: () => void;
}) {
  return (
    <div
      onClick={onClick}
      className="px-4 py-1.5 text-sm text-eve-text hover:bg-eve-accent/20 hover:text-eve-accent cursor-pointer transition-colors"
    >
      {label}
    </div>
  );
}
