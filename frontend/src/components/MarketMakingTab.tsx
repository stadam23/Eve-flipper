import { useCallback, useEffect, useRef, useState } from "react";
import { getPLEXDashboard, type PLEXDashboardParams } from "../lib/api";
import { formatISK } from "../lib/format";
import { useI18n } from "../lib/i18n";
import type { PLEXDashboard, ArbitragePath, MarketDepthInfo, PLEXGlobalPrice } from "../lib/types";

export function MarketMakingTab() {
  const { t } = useI18n();
  const [dashboard, setDashboard] = useState<PLEXDashboard | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [salesTax, setSalesTax] = useState(3.6);
  const [brokerFee, setBrokerFee] = useState(1.0);

  const abortRef = useRef<AbortController | null>(null);

  const fetchData = useCallback(async () => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;

    setLoading(true);
    setError("");
    try {
      const params: PLEXDashboardParams = { salesTax, brokerFee };
      const data = await getPLEXDashboard(params, controller.signal);
      setDashboard(data);
    } catch (e: unknown) {
      if (e instanceof Error && e.name === "AbortError") return;
      setError(e instanceof Error ? e.message : "Failed to load data");
    } finally {
      setLoading(false);
    }
  }, [salesTax, brokerFee]);

  useEffect(() => {
    fetchData();
    return () => abortRef.current?.abort();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const [selectedArb, setSelectedArb] = useState<ArbitragePath | null>(null);

  // Filter only non-NES paths: market_process + spread
  const marketPaths = dashboard?.arbitrage.filter(a => a.type === "market_process") ?? [];
  const spreadPaths = dashboard?.arbitrage.filter(a => a.type === "spread") ?? [];

  return (
    <div className="flex flex-col gap-3 h-full overflow-y-auto pr-1 scrollbar-thin">
      {/* Top bar: controls */}
      <div className="flex items-center gap-3 flex-wrap shrink-0">
        <h2 className="text-sm font-semibold text-eve-accent uppercase tracking-wider">{t("tabMarketMaking")}</h2>
        <div className="flex items-center gap-2 text-xs">
          <label className="text-eve-dim">{t("paramsTax")}</label>
          <input
            type="number"
            step="0.1"
            min="0"
            max="100"
            value={salesTax}
            onChange={(e) => setSalesTax(parseFloat(e.target.value) || 0)}
            className="w-16 px-1.5 py-1 bg-eve-input border border-eve-border rounded-sm text-xs text-eve-text"
          />
          <label className="text-eve-dim">{t("paramsBrokerFee")}</label>
          <input
            type="number"
            step="0.1"
            min="0"
            max="100"
            value={brokerFee}
            onChange={(e) => setBrokerFee(parseFloat(e.target.value) || 0)}
            className="w-16 px-1.5 py-1 bg-eve-input border border-eve-border rounded-sm text-xs text-eve-text"
          />
        </div>
        <button
          onClick={fetchData}
          disabled={loading}
          className="px-3 py-1.5 rounded-sm text-xs font-semibold uppercase tracking-wider bg-eve-accent text-eve-dark hover:bg-eve-accent-hover shadow-eve-glow disabled:opacity-50 disabled:cursor-not-allowed transition-all"
        >
          {loading ? t("plexLoading") : t("plexRefresh")}
        </button>
        {error && <span className="text-xs text-eve-error">{error}</span>}
      </div>

      {!dashboard && !loading && !error && (
        <div className="flex-1 flex items-center justify-center text-eve-dim text-sm">{t("plexEmpty")}</div>
      )}

      {dashboard && (
        <>
          {/* Spread overview banner */}
          <SpreadOverview price={dashboard.plex_price} paths={spreadPaths} />

          {/* Market Process Arbitrage */}
          {marketPaths.length > 0 && (
            <div className="bg-eve-dark border border-eve-border rounded-sm p-3 shrink-0">
              <h3 className="text-xs font-semibold text-eve-dim uppercase tracking-wider mb-2">{t("mmMarketArbitrage")}</h3>
              <p className="text-[10px] text-eve-dim mb-2">{t("mmMarketArbHint")}</p>
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="text-eve-dim border-b border-eve-border">
                      <th className="text-left py-1.5 px-2 font-medium">{t("plexPath")}</th>
                      <th className="text-right py-1.5 px-2 font-medium">{t("plexCost")}</th>
                      <th className="text-right py-1.5 px-2 font-medium">{t("plexRevenue")}</th>
                      <th className="text-right py-1.5 px-2 font-medium">{t("plexProfit")}</th>
                      <th className="text-right py-1.5 px-2 font-medium">ROI</th>
                    </tr>
                  </thead>
                  <tbody>
                    {marketPaths.map((arb, i) => (
                      <ArbRow key={i} arb={arb} onClick={() => setSelectedArb(arb)} />
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Spread Trading Matrix */}
          {spreadPaths.length > 0 && (
            <div className="bg-eve-dark border border-eve-border rounded-sm p-3 shrink-0">
              <h3 className="text-xs font-semibold text-eve-dim uppercase tracking-wider mb-2">{t("plexSpreadSection")}</h3>
              <p className="text-[10px] text-eve-dim mb-2">{t("mmSpreadHint")}</p>
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="text-eve-dim border-b border-eve-border">
                      <th className="text-left py-1.5 px-2 font-medium">{t("mmItem")}</th>
                      <th className="text-right py-1.5 px-2 font-medium">{t("mmBuyOrder")}</th>
                      <th className="text-right py-1.5 px-2 font-medium">{t("mmSellOrder")}</th>
                      <th className="text-right py-1.5 px-2 font-medium">{t("mmSpreadISK")}</th>
                      <th className="text-right py-1.5 px-2 font-medium">{t("plexProfit")}</th>
                      <th className="text-right py-1.5 px-2 font-medium">ROI</th>
                    </tr>
                  </thead>
                  <tbody>
                    {spreadPaths.map((arb, i) => (
                      <SpreadRow key={i} arb={arb} onClick={() => setSelectedArb(arb)} />
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Market Depth */}
          {dashboard.market_depth && (
            <MarketDepthCard depth={dashboard.market_depth} />
          )}

          {/* Tips */}
          <div className="bg-eve-dark border border-eve-border rounded-sm p-3 shrink-0">
            <h3 className="text-xs font-semibold text-eve-dim uppercase tracking-wider mb-2">{t("mmTipsTitle")}</h3>
            <div className="space-y-1.5 text-[11px] text-eve-dim leading-relaxed">
              <p>• {t("plexTipSpread1")}</p>
              <p>• {t("plexTipSpread2")}</p>
              <p>• {t("plexTipSpread3")}</p>
              <p>• {t("mmTip4")}</p>
              <p>• {t("mmTip5")}</p>
              <p>• {t("mmTip6")}</p>
            </div>
          </div>
        </>
      )}

      {/* Detail modal */}
      {selectedArb && (
        <SpreadModal arb={selectedArb} onClose={() => setSelectedArb(null)} />
      )}
    </div>
  );
}

// ===================================================================
// Sub-components
// ===================================================================

function SpreadOverview({ price, paths }: { price: PLEXGlobalPrice; paths: ArbitragePath[] }) {
  const { t } = useI18n();
  const viableCount = paths.filter(p => p.viable).length;
  const bestROI = paths.length > 0 ? Math.max(...paths.map(p => p.roi)) : 0;
  const totalPotential = paths.filter(p => p.viable).reduce((sum, p) => sum + p.profit_isk, 0);

  return (
    <div className="bg-eve-dark border border-eve-accent/30 rounded-sm p-3 shrink-0">
      <div className="grid grid-cols-2 sm:grid-cols-5 gap-4">
        <div>
          <div className="text-[10px] text-eve-dim uppercase tracking-wider mb-0.5">{t("mmPLEXSpread")}</div>
          <div className="text-lg font-mono font-bold text-eve-text">{formatISK(price.spread)}</div>
          <div className="text-[10px] text-eve-dim">{price.spread_pct.toFixed(2)}%</div>
        </div>
        <div>
          <div className="text-[10px] text-eve-dim uppercase tracking-wider mb-0.5">{t("mmBestBid")}</div>
          <div className="text-lg font-mono font-bold text-eve-success">{formatISK(price.buy_price)}</div>
        </div>
        <div>
          <div className="text-[10px] text-eve-dim uppercase tracking-wider mb-0.5">{t("mmBestAsk")}</div>
          <div className="text-lg font-mono font-bold text-eve-error">{formatISK(price.sell_price)}</div>
        </div>
        <div>
          <div className="text-[10px] text-eve-dim uppercase tracking-wider mb-0.5">{t("mmViablePaths")}</div>
          <div className={`text-lg font-mono font-bold ${viableCount > 0 ? "text-eve-success" : "text-eve-error"}`}>
            {viableCount} / {paths.length}
          </div>
        </div>
        <div>
          <div className="text-[10px] text-eve-dim uppercase tracking-wider mb-0.5">{t("mmBestROI")}</div>
          <div className={`text-lg font-mono font-bold ${bestROI > 0 ? "text-eve-success" : "text-eve-error"}`}>
            {bestROI > 0 ? "+" : ""}{bestROI.toFixed(1)}%
          </div>
          {totalPotential > 0 && (
            <div className="text-[10px] text-eve-success">{t("mmTotalPotential")}: {formatISK(totalPotential)}</div>
          )}
        </div>
      </div>
    </div>
  );
}

function ArbRow({ arb, onClick }: { arb: ArbitragePath; onClick: () => void }) {
  return (
    <tr className={`border-b border-eve-border/50 hover:bg-eve-panel/50 transition-colors cursor-pointer ${arb.viable ? "" : arb.no_data ? "opacity-40" : "opacity-50"}`} onClick={onClick}>
      <td className="py-1.5 px-2">
        <div className="flex items-center gap-1.5">
          <span className={`w-1.5 h-1.5 rounded-full shrink-0 ${arb.no_data ? "bg-eve-warning" : arb.viable ? "bg-eve-success" : "bg-eve-error"}`} />
          <span className="text-eve-text hover:text-eve-accent transition-colors">{arb.name}</span>
          {arb.no_data && <span className="text-[9px] text-eve-warning uppercase tracking-wider">no data</span>}
        </div>
      </td>
      <td className="py-1.5 px-2 text-right font-mono text-eve-text">{arb.no_data ? "—" : formatISK(arb.cost_isk)}</td>
      <td className="py-1.5 px-2 text-right font-mono text-eve-text">{arb.no_data ? "—" : formatISK(arb.revenue_isk)}</td>
      <td className={`py-1.5 px-2 text-right font-mono font-semibold ${arb.no_data ? "text-eve-dim" : arb.profit_isk >= 0 ? "text-eve-success" : "text-eve-error"}`}>
        {arb.no_data ? "—" : `${arb.profit_isk >= 0 ? "+" : ""}${formatISK(arb.profit_isk)}`}
      </td>
      <td className={`py-1.5 px-2 text-right font-mono font-semibold ${arb.no_data ? "text-eve-dim" : arb.roi >= 0 ? "text-eve-success" : "text-eve-error"}`}>
        {arb.no_data ? "—" : `${arb.roi >= 0 ? "+" : ""}${arb.roi.toFixed(1)}%`}
      </td>
    </tr>
  );
}

function SpreadRow({ arb, onClick }: { arb: ArbitragePath; onClick: () => void }) {
  return (
    <tr className={`border-b border-eve-border/50 hover:bg-eve-panel/50 transition-colors cursor-pointer ${arb.viable ? "" : arb.no_data ? "opacity-40" : "opacity-50"}`} onClick={onClick}>
      <td className="py-2 px-2">
        <div className="flex items-center gap-1.5">
          <span className={`w-1.5 h-1.5 rounded-full shrink-0 ${arb.no_data ? "bg-eve-warning" : arb.viable ? "bg-eve-success" : "bg-eve-error"}`} />
          <span className="text-eve-text font-medium hover:text-eve-accent transition-colors">{arb.name.replace(" (Market Make)", "")}</span>
        </div>
      </td>
      <td className="py-2 px-2 text-right font-mono text-eve-success">{arb.no_data ? "—" : formatISK(arb.cost_isk)}</td>
      <td className="py-2 px-2 text-right font-mono text-eve-error">{arb.no_data ? "—" : formatISK(arb.revenue_gross)}</td>
      <td className="py-2 px-2 text-right font-mono text-eve-text">{arb.no_data ? "—" : formatISK(arb.revenue_gross - arb.cost_isk)}</td>
      <td className={`py-2 px-2 text-right font-mono font-semibold ${arb.no_data ? "text-eve-dim" : arb.profit_isk >= 0 ? "text-eve-success" : "text-eve-error"}`}>
        {arb.no_data ? "—" : `${arb.profit_isk >= 0 ? "+" : ""}${formatISK(arb.profit_isk)}`}
      </td>
      <td className={`py-2 px-2 text-right font-mono font-semibold ${arb.no_data ? "text-eve-dim" : arb.roi >= 0 ? "text-eve-success" : "text-eve-error"}`}>
        {arb.no_data ? "—" : `${arb.roi >= 0 ? "+" : ""}${arb.roi.toFixed(1)}%`}
      </td>
    </tr>
  );
}

function MarketDepthCard({ depth }: { depth: MarketDepthInfo }) {
  const { t } = useI18n();
  return (
    <div className="bg-eve-dark border border-eve-border rounded-sm p-3 shrink-0">
      <h3 className="text-xs font-semibold text-eve-dim uppercase tracking-wider mb-1">{t("plexMarketDepth")}</h3>
      <p className="text-[10px] text-eve-dim mb-2">{t("mmDepthHint")}</p>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        {/* PLEX */}
        <DepthCard
          label="PLEX"
          sellVol={depth.plex_sell_depth_5.total_volume}
          levels={depth.plex_sell_depth_5.levels}
          bestPrice={depth.plex_sell_depth_5.best_price}
          worstPrice={depth.plex_sell_depth_5.worst_price}
        />
        {/* Extractor */}
        <DepthCard
          label={t("plexDepthExtractor")}
          sellVol={depth.extractor_sell_qty}
          buyVol={depth.extractor_buy_qty}
        />
        {/* Injector */}
        <DepthCard
          label={t("plexDepthInjector")}
          sellVol={depth.injector_sell_qty}
          buyVol={depth.injector_buy_qty}
        />
      </div>
    </div>
  );
}

function DepthCard({ label, sellVol, buyVol, levels, bestPrice, worstPrice }: {
  label: string;
  sellVol: number;
  buyVol?: number;
  levels?: number;
  bestPrice?: number;
  worstPrice?: number;
}) {
  const { t } = useI18n();
  return (
    <div className="border border-eve-border/50 rounded-sm p-2.5">
      <div className="text-[10px] text-eve-accent uppercase tracking-wider font-semibold mb-1.5">{label}</div>
      <div className="space-y-1 text-xs">
        <div className="flex justify-between">
          <span className="text-eve-error">{t("plexDepthSellOrders")}</span>
          <span className="font-mono text-eve-text">{sellVol.toLocaleString()}</span>
        </div>
        {buyVol != null && (
          <div className="flex justify-between">
            <span className="text-eve-success">{t("plexDepthBuyOrders")}</span>
            <span className="font-mono text-eve-text">{buyVol.toLocaleString()}</span>
          </div>
        )}
        {levels != null && (
          <div className="flex justify-between">
            <span className="text-eve-dim">{t("plexDepthLevels")}</span>
            <span className="font-mono text-eve-text">{levels}</span>
          </div>
        )}
        {bestPrice != null && bestPrice > 0 && worstPrice != null && (
          <div className="flex justify-between">
            <span className="text-eve-dim">Range</span>
            <span className="font-mono text-eve-text text-[10px]">{formatISK(bestPrice)} → {formatISK(worstPrice)}</span>
          </div>
        )}
      </div>
    </div>
  );
}

// ===================================================================
// Spread Detail Modal
// ===================================================================

function SpreadModal({ arb, onClose }: { arb: ArbitragePath; onClose: () => void }) {
  const { t } = useI18n();

  useEffect(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === "Escape") onClose(); };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onClose]);

  const steps = getFlowSteps(arb);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={onClose}>
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" />
      <div
        className="relative bg-eve-dark border border-eve-border rounded-sm shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-eve-border">
          <div className="flex items-center gap-2">
            <span className={`w-2.5 h-2.5 rounded-full ${arb.viable ? "bg-eve-success" : "bg-eve-error"}`} />
            <h2 className="text-sm font-semibold text-eve-text uppercase tracking-wider">{arb.name}</h2>
          </div>
          <button onClick={onClose} className="text-eve-dim hover:text-eve-text transition-colors text-lg leading-none px-1">&times;</button>
        </div>

        {/* Flow diagram */}
        <div className="p-4">
          <h3 className="text-[10px] text-eve-dim uppercase tracking-wider font-medium mb-3">{t("plexArbFlow")}</h3>
          <div className="flex items-stretch gap-0 overflow-x-auto pb-2">
            {steps.map((step, i) => (
              <div key={i} className="flex items-center shrink-0">
                <div className={`border rounded-sm px-3 py-2.5 min-w-[110px] text-center ${step.color}`}>
                  <div className="text-xs font-semibold text-eve-text whitespace-nowrap">{step.label}</div>
                  <div className="text-[10px] text-eve-dim mt-0.5 whitespace-nowrap">{step.sub}</div>
                </div>
                {i < steps.length - 1 && (
                  <div className="flex items-center px-1 text-eve-dim shrink-0">
                    <svg width="20" height="12" viewBox="0 0 20 12" fill="none" className="text-eve-dim">
                      <path d="M0 6H16M16 6L11 1M16 6L11 11" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Financial breakdown */}
        <div className="px-4 pb-4">
          <h3 className="text-[10px] text-eve-dim uppercase tracking-wider font-medium mb-3">{t("plexArbBreakdown")}</h3>
          <div className="grid grid-cols-2 gap-3">
            {/* Cost side */}
            <div className="border border-eve-error/20 bg-eve-error/5 rounded-sm p-3">
              <div className="text-[10px] text-eve-error uppercase tracking-wider font-medium mb-2">{t("plexCost")}</div>
              <div className="space-y-1.5 text-xs">
                {arb.type === "spread" ? (
                  <div className="flex justify-between">
                    <span className="text-eve-dim">{t("mmBuyOrderBroker")}</span>
                    <span className="font-mono text-eve-text">{formatISK(arb.cost_isk)}</span>
                  </div>
                ) : (
                  <>
                    <div className="flex justify-between">
                      <span className="text-eve-dim">{t("mmMarketBuyCost")}</span>
                      <span className="font-mono text-eve-text">{formatISK(arb.cost_isk)}</span>
                    </div>
                    {arb.type === "market_process" && (
                      <div className="flex justify-between">
                        <span className="text-eve-dim">{t("mmRequiresSP")}</span>
                        <span className="font-mono text-eve-text">&ge; 5.5M SP</span>
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
            {/* Revenue side */}
            <div className="border border-eve-success/20 bg-eve-success/5 rounded-sm p-3">
              <div className="text-[10px] text-eve-success uppercase tracking-wider font-medium mb-2">{t("plexRevenue")}</div>
              <div className="space-y-1.5 text-xs">
                <div className="flex justify-between">
                  <span className="text-eve-dim">{t("mmSellPrice")}</span>
                  <span className="font-mono text-eve-text">{formatISK(arb.revenue_gross)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-eve-dim">{t("mmAfterFees")}</span>
                  <span className="font-mono text-eve-text">{formatISK(arb.revenue_isk)}</span>
                </div>
              </div>
            </div>
          </div>

          {/* Result */}
          <div className={`mt-3 border rounded-sm p-3 ${arb.viable ? "border-eve-success/30 bg-eve-success/5" : "border-eve-error/30 bg-eve-error/5"}`}>
            <div className="flex items-center justify-between">
              <div>
                <div className="text-[10px] text-eve-dim uppercase tracking-wider font-medium mb-1">{t("plexProfit")}</div>
                <div className={`text-xl font-mono font-bold ${arb.viable ? "text-eve-success" : "text-eve-error"}`}>
                  {arb.profit_isk >= 0 ? "+" : ""}{formatISK(arb.profit_isk)}
                </div>
              </div>
              <div className="text-right">
                <div className="text-[10px] text-eve-dim uppercase tracking-wider font-medium mb-1">ROI</div>
                <div className={`text-xl font-mono font-bold ${arb.viable ? "text-eve-success" : "text-eve-error"}`}>
                  {arb.roi >= 0 ? "+" : ""}{arb.roi.toFixed(1)}%
                </div>
              </div>
            </div>
          </div>

          {/* Tips */}
          <div className="mt-3 text-[11px] text-eve-dim leading-relaxed space-y-1">
            <div className="text-[10px] text-eve-dim uppercase tracking-wider font-medium mb-1">{t("plexArbTips")}</div>
            {arb.type === "spread" && (
              <>
                <p>• {t("plexTipSpread1")}</p>
                <p>• {t("plexTipSpread2")}</p>
                <p>• {t("plexTipSpread3")}</p>
              </>
            )}
            {arb.type === "market_process" && (
              <>
                <p>• {t("plexTipMarket1")}</p>
                <p>• {t("plexTipSPChain1")}</p>
                <p>• {t("plexTipSPChain3")}</p>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function getFlowSteps(arb: ArbitragePath): { label: string; sub: string; color: string }[] {
  const costStr = formatISK(arb.cost_isk);

  if (arb.type === "spread") {
    const itemName = arb.name.split(" ")[0];
    return [
      { label: "Buy Order", sub: `${costStr} ISK`, color: "border-eve-success/40 bg-eve-success/5" },
      { label: itemName, sub: "Wait for fill", color: "border-blue-500/40 bg-blue-500/5" },
      { label: "Sell Order", sub: `${formatISK(arb.revenue_gross)} ISK`, color: "border-eve-error/40 bg-eve-error/5" },
      { label: "Profit", sub: `${formatISK(arb.profit_isk)} ISK`, color: arb.viable ? "border-eve-success/40 bg-eve-success/5" : "border-eve-error/40 bg-eve-error/5" },
    ];
  }

  if (arb.type === "market_process") {
    return [
      { label: "Buy Extractor", sub: `${costStr} ISK`, color: "border-eve-accent/40 bg-eve-accent/5" },
      { label: "Extract SP", sub: "500,000 SP", color: "border-purple-500/40 bg-purple-500/5" },
      { label: "Skill Injector", sub: "Created from SP", color: "border-cyan-500/40 bg-cyan-500/5" },
      { label: "Sell on Market", sub: `${formatISK(arb.revenue_isk)} ISK`, color: "border-eve-success/40 bg-eve-success/5" },
    ];
  }

  return [
    { label: "Cost", sub: costStr, color: "border-eve-error/40 bg-eve-error/5" },
    { label: "Revenue", sub: formatISK(arb.revenue_isk), color: "border-eve-success/40 bg-eve-success/5" },
  ];
}
