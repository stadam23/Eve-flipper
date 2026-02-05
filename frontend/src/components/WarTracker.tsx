import { useEffect, useState, useCallback } from "react";
import { getHotZones, refreshDemandData, getRegionOpportunities } from "../lib/api";
import { useI18n, type TranslationKey } from "../lib/i18n";
import type { DemandRegion, RegionOpportunities, TradeOpportunity } from "../lib/types";

interface WarTrackerProps {
  onError?: (msg: string) => void;
  onOpenRegionArbitrage?: (regionName: string) => void;
}

export function WarTracker({ onError, onOpenRegionArbitrage }: WarTrackerProps) {
  const { t } = useI18n();
  const [hotZones, setHotZones] = useState<DemandRegion[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [fromCache, setFromCache] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedRegion, setSelectedRegion] = useState<DemandRegion | null>(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await getHotZones(50);
      setHotZones(resp.hot_zones || []);
      setFromCache(resp.from_cache);
    } catch (e) {
      const msg = e instanceof Error ? e.message : "Failed to load data";
      setError(msg);
      onError?.(msg);
    } finally {
      setLoading(false);
    }
  }, [onError]);

  const handleRefresh = async () => {
    setRefreshing(true);
    setError(null);
    try {
      // Refresh is now synchronous - will wait for all regions to be fetched
      await refreshDemandData();
      // Reload data after successful refresh
      await loadData();
    } catch (e) {
      const msg = e instanceof Error ? e.message : "Failed to refresh";
      setError(msg);
      onError?.(msg);
    } finally {
      setRefreshing(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [loadData]);

  const formatISK = (value: number) => {
    if (value >= 1e15) return `${(value / 1e15).toFixed(1)}Q`;
    if (value >= 1e12) return `${(value / 1e12).toFixed(1)}T`;
    if (value >= 1e9) return `${(value / 1e9).toFixed(1)}B`;
    if (value >= 1e6) return `${(value / 1e6).toFixed(1)}M`;
    if (value >= 1e3) return `${(value / 1e3).toFixed(1)}K`;
    return value.toFixed(0);
  };

  const warZones = hotZones.filter(z => z.status === "war");
  const conflictZones = hotZones.filter(z => z.status === "conflict");
  const elevatedZones = hotZones.filter(z => z.status === "elevated");
  const normalZones = hotZones.filter(z => z.status === "normal");

  return (
    <div className="flex-1 flex flex-col gap-4 p-4 overflow-auto">
      {/* Header */}
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div>
          <h2 className="text-lg font-semibold text-eve-accent flex items-center gap-2">
            🗺️ {t("warTrackerTitle") || "War Tracker"}
          </h2>
          <p className="text-xs text-eve-dim mt-1">
            {t("warTrackerDesc") || "Monitor kill activity across EVE regions to find war profit opportunities"}
          </p>
        </div>
        <div className="flex items-center gap-2">
          {fromCache && (
            <span className="text-xs text-eve-dim px-2 py-1 bg-eve-dark rounded-sm">
              📦 {t("fromCache") || "From cache"}
            </span>
          )}
          <button
            onClick={handleRefresh}
            disabled={refreshing || loading}
            className="px-3 py-1.5 text-xs bg-eve-accent text-eve-dark rounded-sm hover:bg-eve-accent-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {refreshing ? t("refreshing") || "Refreshing..." : t("refresh") || "Refresh Data"}
          </button>
          <button
            onClick={loadData}
            disabled={loading}
            className="px-3 py-1.5 text-xs bg-eve-panel border border-eve-border text-eve-text rounded-sm hover:border-eve-accent/50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {loading ? "..." : "↻"}
          </button>
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="p-3 bg-red-500/10 border border-red-500/30 rounded-sm text-red-400 text-sm">
          {error}
        </div>
      )}

      {/* Loading */}
      {loading && hotZones.length === 0 && (
        <div className="flex-1 flex items-center justify-center">
          <div className="text-eve-dim text-sm animate-pulse">
            {t("loadingRegions") || "Loading region data..."}
          </div>
        </div>
      )}

      {/* Summary Stats */}
      {hotZones.length > 0 && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
          <div className="p-3 bg-red-500/10 border border-red-500/30 rounded-sm">
            <div className="text-2xl font-bold text-red-500">{warZones.length}</div>
            <div className="text-xs text-red-400">🔥 {t("warZones") || "War Zones"}</div>
          </div>
          <div className="p-3 bg-orange-500/10 border border-orange-500/30 rounded-sm">
            <div className="text-2xl font-bold text-orange-500">{conflictZones.length}</div>
            <div className="text-xs text-orange-400">⚠️ {t("conflicts") || "Conflicts"}</div>
          </div>
          <div className="p-3 bg-yellow-500/10 border border-yellow-500/30 rounded-sm">
            <div className="text-2xl font-bold text-yellow-500">{elevatedZones.length}</div>
            <div className="text-xs text-yellow-400">📈 {t("elevated") || "Elevated"}</div>
          </div>
          <div className="p-3 bg-green-500/10 border border-green-500/30 rounded-sm">
            <div className="text-2xl font-bold text-green-500">{normalZones.length}</div>
            <div className="text-xs text-green-400">✅ {t("normal") || "Normal"}</div>
          </div>
        </div>
      )}

      {/* Hot Zones List */}
      {hotZones.length > 0 && (
        <div className="flex-1 flex flex-col gap-2 min-h-0 overflow-auto">
          {/* War Zones */}
          {warZones.length > 0 && (
            <div className="space-y-2">
              <h3 className="text-sm font-semibold text-red-500 flex items-center gap-2 sticky top-0 bg-eve-panel py-1">
                🔥 {t("warZonesTitle") || "Active War Zones"} ({warZones.length})
              </h3>
              <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
                {warZones.map(zone => (
                  <RegionCard key={zone.region_id} zone={zone} formatISK={formatISK} onClick={() => setSelectedRegion(zone)} />
                ))}
              </div>
            </div>
          )}

          {/* Conflicts */}
          {conflictZones.length > 0 && (
            <div className="space-y-2">
              <h3 className="text-sm font-semibold text-orange-500 flex items-center gap-2 sticky top-0 bg-eve-panel py-1">
                ⚠️ {t("conflictsTitle") || "Active Conflicts"} ({conflictZones.length})
              </h3>
              <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
                {conflictZones.map(zone => (
                  <RegionCard key={zone.region_id} zone={zone} formatISK={formatISK} onClick={() => setSelectedRegion(zone)} />
                ))}
              </div>
            </div>
          )}

          {/* Elevated */}
          {elevatedZones.length > 0 && (
            <div className="space-y-2">
              <h3 className="text-sm font-semibold text-yellow-500 flex items-center gap-2 sticky top-0 bg-eve-panel py-1">
                📈 {t("elevatedTitle") || "Elevated Activity"} ({elevatedZones.length})
              </h3>
              <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
                {elevatedZones.map(zone => (
                  <RegionCard key={zone.region_id} zone={zone} formatISK={formatISK} onClick={() => setSelectedRegion(zone)} />
                ))}
              </div>
            </div>
          )}

          {/* Normal - collapsed by default */}
          {normalZones.length > 0 && (
            <details className="space-y-2">
              <summary className="text-sm font-semibold text-green-500 flex items-center gap-2 cursor-pointer sticky top-0 bg-eve-panel py-1">
                ✅ {t("normalTitle") || "Normal Activity"} ({normalZones.length})
              </summary>
              <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3 pt-2">
                {normalZones.map(zone => (
                  <RegionCard key={zone.region_id} zone={zone} formatISK={formatISK} onClick={() => setSelectedRegion(zone)} />
                ))}
              </div>
            </details>
          )}
        </div>
      )}

      {/* Region Detail Popup */}
      {selectedRegion && (
        <RegionDetailPopup
          region={selectedRegion}
          onClose={() => setSelectedRegion(null)}
          onOpenArbitrage={() => {
            onOpenRegionArbitrage?.(selectedRegion.region_name);
            setSelectedRegion(null);
          }}
          formatISK={formatISK}
        />
      )}

      {/* Empty state */}
      {!loading && hotZones.length === 0 && !error && (
        <div className="flex-1 flex flex-col items-center justify-center text-eve-dim">
          <div className="text-4xl mb-4">🌌</div>
          <div className="text-sm">{t("noDataYet") || "No data yet"}</div>
          <button
            onClick={handleRefresh}
            className="mt-4 px-4 py-2 text-sm bg-eve-accent text-eve-dark rounded-sm hover:bg-eve-accent-hover transition-colors"
          >
            {t("loadRegionData") || "Load Region Data"}
          </button>
        </div>
      )}
    </div>
  );
}

// Region Card Component
interface RegionCardProps {
  zone: DemandRegion;
  formatISK: (value: number) => string;
  onClick?: () => void;
}

function RegionCard({ zone, formatISK, onClick }: RegionCardProps) {
  const getStatusBorder = (status: string) => {
    switch (status) {
      case "war": return "border-red-500/50 hover:border-red-500";
      case "conflict": return "border-orange-500/50 hover:border-orange-500";
      case "elevated": return "border-yellow-500/50 hover:border-yellow-500";
      default: return "border-eve-border hover:border-eve-accent/50";
    }
  };

  const hotScorePercent = Math.round((zone.hot_score - 1) * 100);
  const hotScoreDisplay = hotScorePercent >= 0 ? `+${hotScorePercent}%` : `${hotScorePercent}%`;

  return (
    <div 
      className={`p-3 bg-eve-dark/50 border rounded-sm transition-colors cursor-pointer ${getStatusBorder(zone.status)}`}
      onClick={onClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => e.key === "Enter" && onClick?.()}
    >
      <div className="flex items-start justify-between gap-2">
        <div>
          <div className="font-semibold text-eve-text">{zone.region_name}</div>
          <div className="text-xs text-eve-dim mt-0.5">
            {zone.kills_today.toLocaleString()} kills/day
            <span className={`ml-2 ${zone.hot_score >= 1.2 ? 'text-red-400' : 'text-eve-dim'}`}>
              ({hotScoreDisplay} vs avg)
            </span>
          </div>
        </div>
        <div className={`text-lg font-bold ${
          zone.status === "war" ? "text-red-500" :
          zone.status === "conflict" ? "text-orange-500" :
          zone.status === "elevated" ? "text-yellow-500" : "text-green-500"
        }`}>
          {zone.hot_score.toFixed(1)}x
        </div>
      </div>
      
      <div className="mt-2 flex flex-wrap gap-2 text-xs">
        <div className="px-2 py-0.5 bg-eve-panel rounded-sm">
          👥 {zone.active_players.toLocaleString()} active
        </div>
        <div className="px-2 py-0.5 bg-eve-panel rounded-sm">
          💀 {formatISK(zone.isk_destroyed)} ISK lost
        </div>
      </div>

      {zone.top_ships && zone.top_ships.length > 0 && (
        <div className="mt-2 text-xs text-eve-dim">
          <span className="text-eve-text">Top losses:</span>{" "}
          {zone.top_ships.slice(0, 3).join(", ")}
        </div>
      )}
      
      <div className="mt-2 text-xs text-eve-accent opacity-60">
        Click for details →
      </div>
    </div>
  );
}

// Region Detail Popup Component
interface RegionDetailPopupProps {
  region: DemandRegion;
  onClose: () => void;
  onOpenArbitrage: () => void;
  formatISK: (value: number) => string;
}

function RegionDetailPopup({ region, onClose, onOpenArbitrage, formatISK }: RegionDetailPopupProps) {
  const { t } = useI18n();
  const [opportunities, setOpportunities] = useState<RegionOpportunities | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  useEffect(() => {
    const loadOpportunities = async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await getRegionOpportunities(region.region_id);
        setOpportunities(data);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load opportunities");
      } finally {
        setLoading(false);
      }
    };
    loadOpportunities();
  }, [region.region_id]);

  const hotScorePercent = Math.round((region.hot_score - 1) * 100);
  const hotScoreDisplay = hotScorePercent >= 0 ? `+${hotScorePercent}%` : `${hotScorePercent}%`;

  const getStatusInfo = (status: string) => {
    switch (status) {
      case "war": return { icon: "🔥", label: t("statusWar") || "Active War", color: "text-red-500", bg: "bg-red-500/10" };
      case "conflict": return { icon: "⚠️", label: t("statusConflict") || "Conflict Zone", color: "text-orange-500", bg: "bg-orange-500/10" };
      case "elevated": return { icon: "📈", label: t("statusElevated") || "Elevated Activity", color: "text-yellow-500", bg: "bg-yellow-500/10" };
      default: return { icon: "✅", label: t("statusNormal") || "Normal Activity", color: "text-green-500", bg: "bg-green-500/10" };
    }
  };

  const statusInfo = getStatusInfo(region.status);

  const formatPrice = (value: number) => {
    if (value >= 1e9) return `${(value / 1e9).toFixed(1)}B`;
    if (value >= 1e6) return `${(value / 1e6).toFixed(1)}M`;
    if (value >= 1e3) return `${(value / 1e3).toFixed(1)}K`;
    return value.toFixed(0);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70" onClick={onClose}>
      <div 
        className="bg-eve-panel border border-eve-border rounded-sm max-w-3xl w-full max-h-[90vh] flex flex-col shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header - fixed at top */}
        <div className="shrink-0 bg-eve-panel border-b border-eve-border p-4 flex items-start justify-between">
          <div>
            <div className="flex items-center gap-2">
              <span className="text-2xl">{statusInfo.icon}</span>
              <h2 className="text-xl font-bold text-eve-text">{region.region_name}</h2>
            </div>
            <div className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-sm text-xs mt-2 ${statusInfo.bg} ${statusInfo.color}`}>
              {statusInfo.label}
            </div>
            {/* Security and distance info */}
            {opportunities && (
              <div className="flex items-center gap-3 mt-2 text-xs">
                {/* Security blocks */}
                <div className="flex items-center gap-1">
                  {opportunities.security_blocks?.map((block, i) => (
                    <span
                      key={i}
                      className={`w-3 h-3 rounded-sm ${
                        block === "high" ? "bg-green-500" :
                        block === "low" ? "bg-yellow-500" : "bg-red-500"
                      }`}
                      title={block === "high" ? "Highsec" : block === "low" ? "Lowsec" : "Nullsec"}
                    />
                  ))}
                  <span className="text-eve-dim ml-1">
                    {opportunities.security_class === "highsec" ? "Highsec" :
                     opportunities.security_class === "lowsec" ? "Lowsec" : "Nullsec"}
                  </span>
                </div>
                {/* Distance from Jita */}
                {opportunities.jumps_from_jita > 0 && (
                  <div className="flex items-center gap-1 text-eve-dim">
                    <span>•</span>
                    <span className="text-eve-accent font-medium">{opportunities.jumps_from_jita}</span>
                    <span>{t("jumpsFromJita") || "jumps from Jita"}</span>
                    {opportunities.main_system && (
                      <span className="text-eve-text">→ {opportunities.main_system}</span>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-eve-dark/50 rounded-sm transition-colors text-eve-dim hover:text-eve-text"
          >
            ✕
          </button>
        </div>

        {/* Content - scrollable */}
        <div className="flex-1 overflow-auto p-4 space-y-4">
          {/* Stats Grid */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <div className="p-3 bg-eve-dark/50 rounded-sm">
              <div className={`text-2xl font-bold ${statusInfo.color}`}>{region.hot_score.toFixed(1)}x</div>
              <div className="text-xs text-eve-dim">{t("activityIndex") || "Activity Index"}</div>
            </div>
            <div className="p-3 bg-eve-dark/50 rounded-sm">
              <div className="text-2xl font-bold text-eve-text">{region.kills_today.toLocaleString()}</div>
              <div className="text-xs text-eve-dim">{t("killsPerDay") || "Kills/Day"}</div>
            </div>
            <div className="p-3 bg-eve-dark/50 rounded-sm">
              <div className="text-2xl font-bold text-eve-text">{region.active_players.toLocaleString()}</div>
              <div className="text-xs text-eve-dim">{t("activePlayers") || "Active Players"}</div>
            </div>
            <div className="p-3 bg-eve-dark/50 rounded-sm">
              <div className="text-2xl font-bold text-eve-accent">{hotScoreDisplay}</div>
              <div className="text-xs text-eve-dim">{t("vsBaseline") || "vs Average"}</div>
            </div>
          </div>

          {/* Loading/Error States */}
          {loading && (
            <div className="text-center p-8 text-eve-dim">
              <div className="animate-pulse">{t("loadingOpportunities") || "Loading trade opportunities..."}</div>
            </div>
          )}

          {error && (
            <div className="text-center p-4 bg-red-500/10 border border-red-500/30 rounded-sm text-red-400 text-sm">
              {error}
            </div>
          )}

          {/* Trade Opportunities */}
          {opportunities && !loading && (
            <>
              {/* Total Potential */}
              {opportunities.total_potential > 0 && (
                <div className="p-4 bg-green-500/10 border border-green-500/30 rounded-sm">
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="text-sm font-semibold text-green-400 flex items-center gap-2">
                        💰 {t("dailyProfitPotential") || "Daily Profit Potential"}
                      </h3>
                      <p className="text-xs text-eve-dim mt-1">{t("basedOnDestroyedShips") || "Based on destroyed ships and demand"}</p>
                    </div>
                    <div className="text-2xl font-bold text-green-400">
                      {formatISK(opportunities.total_potential)} ISK
                    </div>
                  </div>
                </div>
              )}

              {/* Ships Opportunities */}
              {opportunities.ships && opportunities.ships.length > 0 && (
                <div className="space-y-2">
                  <h3 className="text-sm font-semibold text-eve-text flex items-center gap-2">
                    🚀 {t("shipOpportunities") || "Ship Opportunities"}
                    <span className="text-xs text-eve-dim font-normal">({opportunities.ships.length})</span>
                  </h3>
                  <div className="space-y-2">
                    {opportunities.ships.slice(0, 5).map((opp, i) => (
                      <OpportunityCard key={i} opportunity={opp} formatPrice={formatPrice} t={t} />
                    ))}
                  </div>
                </div>
              )}

              {/* Modules Opportunities */}
              {opportunities.modules && opportunities.modules.length > 0 && (
                <div className="space-y-2">
                  <h3 className="text-sm font-semibold text-eve-text flex items-center gap-2">
                    🔧 {t("moduleOpportunities") || "Module Opportunities"}
                    <span className="text-xs text-eve-dim font-normal">({opportunities.modules.length})</span>
                  </h3>
                  <div className="space-y-2">
                    {opportunities.modules.slice(0, 5).map((opp, i) => (
                      <OpportunityCard key={i} opportunity={opp} formatPrice={formatPrice} t={t} />
                    ))}
                  </div>
                </div>
              )}

              {/* Ammo Opportunities */}
              {opportunities.ammo && opportunities.ammo.length > 0 && (
                <div className="space-y-2">
                  <h3 className="text-sm font-semibold text-eve-text flex items-center gap-2">
                    📦 {t("ammoOpportunities") || "Ammo & Consumables"}
                    <span className="text-xs text-eve-dim font-normal">({opportunities.ammo.length})</span>
                  </h3>
                  <div className="space-y-2">
                    {opportunities.ammo.slice(0, 5).map((opp, i) => (
                      <OpportunityCard key={i} opportunity={opp} formatPrice={formatPrice} t={t} />
                    ))}
                  </div>
                </div>
              )}

              {/* No opportunities found */}
              {(!opportunities.ships || opportunities.ships.length === 0) &&
               (!opportunities.modules || opportunities.modules.length === 0) &&
               (!opportunities.ammo || opportunities.ammo.length === 0) && (
                <div className="text-center p-4 bg-eve-dark/30 rounded-sm text-eve-dim text-sm">
                  {t("noOpportunities") || "No significant trade opportunities found. Prices may be similar to Jita."}
                </div>
              )}
            </>
          )}

        </div>

        {/* Footer - fixed at bottom with stats and actions */}
        <div className="shrink-0 bg-eve-panel border-t border-eve-border">
          {/* Top Destroyed Ships + Total ISK in one row */}
          <div className="p-3 border-b border-eve-border/50 flex items-center justify-between gap-4">
            {/* Top Ships */}
            {region.top_ships && region.top_ships.length > 0 && (
              <div className="flex items-center gap-2 flex-1 min-w-0">
                <span className="text-xs text-eve-dim shrink-0">💀</span>
                <div className="flex flex-wrap gap-1">
                  {region.top_ships.slice(0, 5).map((ship, i) => (
                    <span key={i} className="px-1.5 py-0.5 bg-eve-dark/50 border border-eve-border/50 rounded-sm text-[10px] text-eve-text">
                      {ship}
                    </span>
                  ))}
                </div>
              </div>
            )}
            {/* Total ISK */}
            <div className="text-right shrink-0">
              <div className="text-[10px] text-eve-dim">{t("totalIskDestroyed") || "ISK Destroyed"}</div>
              <div className="text-sm font-bold text-eve-accent">{formatISK(region.isk_destroyed)}</div>
            </div>
          </div>
          
          {/* Action buttons */}
          <div className="p-3 flex flex-wrap gap-2">
            <button
              onClick={onOpenArbitrage}
              className="flex-1 px-4 py-2 bg-eve-accent text-eve-dark font-semibold rounded-sm hover:bg-eve-accent-hover transition-colors text-sm"
            >
              🔍 {t("findArbitrageOpportunities") || "Find Arbitrage in Region"}
            </button>
            <a
              href={`https://zkillboard.com/region/${region.region_id}/`}
              target="_blank"
              rel="noopener noreferrer"
              className="px-4 py-2 bg-eve-panel border border-eve-border text-eve-text rounded-sm hover:border-eve-accent/50 transition-colors text-sm"
            >
              📊 Zkillboard
            </a>
            <a
              href={`https://evemaps.dotlan.net/map/${region.region_name.replace(/ /g, "_")}`}
              target="_blank"
              rel="noopener noreferrer"
              className="px-4 py-2 bg-eve-panel border border-eve-border text-eve-text rounded-sm hover:border-eve-accent/50 transition-colors text-sm"
            >
              🗺️ Dotlan
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}

// Individual opportunity card component
interface OpportunityCardProps {
  opportunity: TradeOpportunity;
  formatPrice: (value: number) => string;
  t: (key: TranslationKey) => string;
}

function OpportunityCard({ opportunity, formatPrice, t }: OpportunityCardProps) {
  const noSupply = opportunity.region_price <= 0;
  const profitColor = opportunity.profit_percent >= 50 ? "text-green-400" : 
                      opportunity.profit_percent >= 20 ? "text-yellow-400" : "text-eve-accent";
  
  // Рекомендуемая цена если нет предложения (Jita + 30%)
  const suggestedPrice = noSupply ? opportunity.jita_price * 1.3 : opportunity.region_price;
  
  return (
    <div className={`p-3 border rounded-sm hover:border-eve-accent/30 transition-colors ${
      noSupply ? "bg-green-500/10 border-green-500/30" : "bg-eve-dark/50 border-eve-border/50"
    }`}>
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium text-eve-text truncate" title={opportunity.type_name}>
              {opportunity.type_name || `Type #${opportunity.type_id}`}
            </span>
            {noSupply && (
              <span className="px-1.5 py-0.5 bg-green-500/20 text-green-400 text-[10px] rounded-sm font-medium shrink-0">
                {t("noCompetition") || "NO COMPETITION"}
              </span>
            )}
          </div>
          <div className="flex items-center gap-3 mt-1 text-xs">
            <span className="text-eve-dim">
              Jita: <span className="text-eve-text">{formatPrice(opportunity.jita_price)}</span>
            </span>
            <span className="text-eve-dim">→</span>
            {noSupply ? (
              <span className="text-green-400">
                {t("sellFor") || "Sell for"}: <span className="font-medium">{formatPrice(suggestedPrice)}+</span>
              </span>
            ) : (
              <span className="text-eve-dim">
                {t("region") || "Region"}: <span className="text-eve-text">{formatPrice(opportunity.region_price)}</span>
              </span>
            )}
          </div>
        </div>
        <div className="text-right shrink-0">
          <div className={`font-bold ${profitColor}`}>
            +{opportunity.profit_percent.toFixed(0)}%
          </div>
          <div className="text-xs text-eve-dim">
            +{formatPrice(opportunity.profit_per_unit)}/unit
          </div>
        </div>
      </div>
      <div className="flex items-center justify-between mt-2 pt-2 border-t border-eve-border/30 text-xs">
        <span className="text-eve-dim">
          {t("demand") || "Demand"}: <span className="text-eve-text">{opportunity.kills_per_day}/day</span>
        </span>
        <span className="text-green-400 font-medium">
          ≈ {formatPrice(opportunity.daily_profit)} ISK/day
        </span>
      </div>
    </div>
  );
}
