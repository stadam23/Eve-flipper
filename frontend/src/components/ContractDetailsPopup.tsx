import { useEffect, useMemo, useState } from "react";
import { Modal } from "./Modal";
import { getContractDetails, openContractInGame } from "../lib/api";
import type { ContractDetails, ContractItem } from "../lib/types";
import { useI18n } from "../lib/i18n";
import { formatISK } from "../lib/format";
import { useGlobalToast } from "./Toast";
import { handleEveUIError } from "../lib/handleEveUIError";

interface ContractDetailsPopupProps {
  open: boolean;
  contractID: number;
  contractTitle: string;
  contractPrice: number;
  contractMarketValue?: number;
  contractProfit?: number;
  excludeRigPriceIfShip?: boolean;
  pickupStationName?: string;
  pickupSystemName?: string;
  pickupRegionName?: string;
  liquidationSystemName?: string;
  liquidationRegionName?: string;
  liquidationJumps?: number;
  totalJumps?: number;
  isLoggedIn?: boolean;
  onClose: () => void;
}

function formatLocation(parts: Array<string | undefined>): string {
  const normalized = parts
    .map((part) => (part ?? "").trim())
    .filter((part) => part.length > 0);
  return normalized.length > 0 ? normalized.join(" • ") : "\u2014";
}

function isShipItem(item: ContractItem): boolean {
  return item.is_ship === true || item.category_id === 6;
}

function isRigItem(item: ContractItem): boolean {
  if (item.is_rig === true) return true;
  const groupName = (item.group_name ?? "").trim().toLowerCase();
  return item.category_id === 7 && groupName.startsWith("rig");
}

function formatSignedISK(value: number): string {
  if (value > 0) return `+${formatISK(value)}`;
  return formatISK(value);
}

export function ContractDetailsPopup({
  open,
  contractID,
  contractTitle,
  contractPrice,
  contractMarketValue,
  contractProfit,
  excludeRigPriceIfShip = true,
  pickupStationName,
  pickupSystemName,
  pickupRegionName,
  liquidationSystemName,
  liquidationRegionName,
  liquidationJumps,
  totalJumps,
  isLoggedIn = false,
  onClose,
}: ContractDetailsPopupProps) {
  const { t } = useI18n();
  const { addToast } = useGlobalToast();
  const [details, setDetails] = useState<ContractDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const pickupLocation = formatLocation([pickupStationName, pickupSystemName, pickupRegionName]);
  const liquidationLocation = formatLocation([liquidationSystemName, liquidationRegionName]);
  const hasLiquidationPoint = liquidationLocation !== "\u2014";

  useEffect(() => {
    if (!open) return;
    setLoading(true);
    setError(null);
    getContractDetails(contractID)
      .then((data) => {
        setDetails(data);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message || "Failed to load contract details");
        setLoading(false);
      });
  }, [contractID, open]);

  const displayedTitle = useMemo(() => {
    if (!details || !Array.isArray(details.items) || details.items.length === 0) {
      return contractTitle;
    }
    const includedNames = details.items
      .filter((item) => item.is_included && item.quantity > 0)
      .map((item) => item.type_name?.trim() || `Type ${item.type_id}`)
      .filter((name) => name.length > 0);
    if (includedNames.length === 0) {
      return contractTitle;
    }
    if (includedNames.length <= 3) {
      return includedNames.join(", ");
    }
    return `${includedNames.slice(0, 2).join(", ")} + ${includedNames.length - 2} more`;
  }, [contractTitle, details]);

  // Keep raw rows to preserve risk signals (damage/fitted-like markers/BP params).
  const includedItems = details?.items.filter((item) => item.is_included) || [];
  const requestedItems = details?.items.filter((item) => !item.is_included) || [];
  const hasShipInContract = includedItems.some(isShipItem);
  const rigItemsInContract = includedItems.filter(isRigItem);
  const rigItemsTotalQty = rigItemsInContract.reduce((sum, item) => sum + item.quantity, 0);
  const rigExclusionApplied = excludeRigPriceIfShip && hasShipInContract && rigItemsInContract.length > 0;
  const scanSpread = typeof contractMarketValue === "number" ? contractMarketValue - contractPrice : null;
  const handleOpenContract = async () => {
    try {
      await openContractInGame(contractID);
      addToast(t("actionSuccess"), "success", 2000);
    } catch (err: any) {
      const { messageKey, duration } = handleEveUIError(err);
      addToast(t(messageKey), "error", duration);
    }
  };

  return (
    <Modal open={open} onClose={onClose} title={`${t("contractDetails")} #${contractID}`}>
      <div className="p-4 flex flex-col gap-4">
        {/* Contract info */}
        <div className="border border-eve-border rounded-sm p-3 bg-eve-panel">
          <div className="flex items-start justify-between gap-3">
            <div className="text-sm text-eve-text flex-1">
              <span className="text-eve-dim">{t("colTitle")}:</span> {displayedTitle}
            </div>
            {isLoggedIn && (
              <button
                onClick={handleOpenContract}
                className="px-2.5 py-1 rounded-sm border border-eve-border text-eve-dim hover:text-eve-accent hover:border-eve-accent/40 transition-colors text-xs whitespace-nowrap"
              >
                🎮 {t("openContract")}
              </button>
            )}
          </div>
          <div className="text-sm text-eve-accent font-mono mt-1">
            <span className="text-eve-dim">{t("iskPrice")}:</span> {formatISK(contractPrice)}
          </div>
          <div className="mt-3 grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
            <div>
              <div className="text-eve-dim uppercase tracking-wider">{t("contractPickupPoint")}</div>
              <div className="text-eve-text mt-1">{pickupLocation}</div>
            </div>
            <div>
              <div className="text-eve-dim uppercase tracking-wider">{t("contractLiquidationPoint")}</div>
              <div className={`mt-1 ${hasLiquidationPoint ? "text-eve-success" : "text-eve-dim"}`}>
                {liquidationLocation}
              </div>
            </div>
          </div>
          {hasLiquidationPoint && (
            <div className="mt-3 flex flex-wrap items-center gap-2 text-xs">
              <span className="px-2 py-0.5 rounded-sm border border-eve-border text-eve-dim">
                {t("contractRouteLabel")}: {(pickupSystemName || pickupStationName || "\u2014")} {"\u2192"} {liquidationSystemName}
              </span>
              <span className="px-2 py-0.5 rounded-sm border border-eve-border text-eve-dim">
                {t("contractLiqJumps")}: {typeof liquidationJumps === "number" ? liquidationJumps : "\u2014"}
              </span>
              <span className="px-2 py-0.5 rounded-sm border border-eve-border text-eve-dim">
                {t("colContractJumps")}: {typeof totalJumps === "number" ? totalJumps : "\u2014"}
              </span>
            </div>
          )}
        </div>

        {/* Warning about damage and fitted items */}
        <div className="border border-yellow-700/50 bg-yellow-900/20 rounded-sm p-3">
          <div className="flex items-start gap-2">
            <span className="text-yellow-400 text-lg">⚠</span>
            <div className="flex-1 text-xs text-yellow-200">
              <div className="font-semibold mb-1">{t("contractDetailsWarningTitle")}</div>
              <div className="text-yellow-300/90">
                • {t("contractDetailsWarningDamage")}
                <br />
                • {t("contractDetailsWarningFitted")}
              </div>
            </div>
          </div>
        </div>

        {loading && (
          <div className="flex items-center justify-center h-40">
            <div className="text-eve-dim">{t("loading")}...</div>
          </div>
        )}

        {error && (
          <div className="text-eve-error bg-red-900/20 border border-red-700 rounded-sm p-3 text-sm">
            {error}
          </div>
        )}

        {!loading && !error && details && (
          <>
            {excludeRigPriceIfShip && (
              <div className="border border-yellow-700/50 bg-yellow-950/20 rounded-sm p-3">
                <div className="text-xs font-semibold uppercase tracking-wider text-yellow-300">
                  {t("contractRigCheckoutTitle")}
                </div>
                <div className="text-xs text-yellow-200/90 mt-1">
                  {rigExclusionApplied ? t("contractRigCheckoutAppliedHint") : t("contractRigCheckoutNoRigHint")}
                </div>
                <div className="mt-3 grid grid-cols-1 md:grid-cols-3 gap-3">
                  <div>
                    <div className="text-[11px] uppercase tracking-wider text-eve-dim">{t("iskPrice")}</div>
                    <div className="mt-1 text-sm font-mono text-eve-accent">{formatISK(contractPrice)}</div>
                  </div>
                  <div>
                    <div className="text-[11px] uppercase tracking-wider text-eve-dim">
                      {t("contractRigCheckoutValueLabel")}
                    </div>
                    <div className="mt-1 text-sm font-mono text-eve-success">
                      {typeof contractMarketValue === "number" ? formatISK(contractMarketValue) : "\u2014"}
                    </div>
                  </div>
                  <div>
                    <div className="text-[11px] uppercase tracking-wider text-eve-dim">
                      {t("contractRigCheckoutSpreadLabel")}
                    </div>
                    <div className={`mt-1 text-sm font-mono ${scanSpread != null && scanSpread >= 0 ? "text-eve-success" : "text-eve-error"}`}>
                      {scanSpread != null ? formatSignedISK(scanSpread) : "\u2014"}
                    </div>
                  </div>
                </div>
                <div className="mt-2 text-[11px] text-eve-dim">
                  {t("contractRigDetectedRows")}: {rigItemsInContract.length} · {t("contractRigDetectedQty")}: {rigItemsTotalQty.toLocaleString()}
                  {typeof contractProfit === "number" && (
                    <> · {t("colContractProfit")}: {formatSignedISK(contractProfit)}</>
                  )}
                </div>
              </div>
            )}

            {/* Items included (seller provides) */}
            {includedItems.length > 0 && (
              <div className="border border-eve-border rounded-sm overflow-hidden">
                <div className="px-3 py-2 bg-eve-panel border-b border-eve-border text-xs font-semibold text-green-400 uppercase tracking-wider">
                  ✓ {t("itemsIncluded")} ({includedItems.length})
                </div>
                <table className="w-full text-sm">
                  <thead className="bg-eve-panel border-b border-eve-border">
                    <tr>
                      <th className="text-left px-3 py-1.5 text-xs text-eve-dim uppercase tracking-wider">{t("colItem")}</th>
                      <th className="text-right px-3 py-1.5 text-xs text-eve-dim uppercase tracking-wider">{t("execPlanQuantity")}</th>
                      <th className="text-left px-3 py-1.5 text-xs text-eve-dim uppercase tracking-wider">{t("colType")}</th>
                    </tr>
                  </thead>
                  <tbody className="text-eve-text">
                    {includedItems.map((item, idx) => (
                      <ItemRow
                        key={`${item.record_id}-${item.item_id}-${idx}`}
                        item={item}
                        highlightRig={rigExclusionApplied && isRigItem(item)}
                      />
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            {/* Items requested (buyer must provide) */}
            {requestedItems.length > 0 && (
              <div className="border border-eve-border rounded-sm overflow-hidden">
                <div className="px-3 py-2 bg-eve-panel border-b border-eve-border text-xs font-semibold text-yellow-400 uppercase tracking-wider">
                  ⚠ {t("itemsRequested")} ({requestedItems.length})
                </div>
                <table className="w-full text-sm">
                  <thead className="bg-eve-panel border-b border-eve-border">
                    <tr>
                      <th className="text-left px-3 py-1.5 text-xs text-eve-dim uppercase tracking-wider">{t("colItem")}</th>
                      <th className="text-right px-3 py-1.5 text-xs text-eve-dim uppercase tracking-wider">{t("execPlanQuantity")}</th>
                      <th className="text-left px-3 py-1.5 text-xs text-eve-dim uppercase tracking-wider">{t("colType")}</th>
                    </tr>
                  </thead>
                  <tbody className="text-eve-text">
                    {requestedItems.map((item, idx) => (
                      <ItemRow key={`${item.record_id}-${item.item_id}-${idx}`} item={item} />
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </>
        )}
      </div>
    </Modal>
  );
}

function ItemRow({ item, highlightRig = false }: { item: ContractItem; highlightRig?: boolean }) {
  const { t } = useI18n();

  let typeLabel = t("colItem");
  if (item.is_blueprint_copy) {
    typeLabel = `${t("blueprintCopy")} (${item.runs || 0} ${t("runs")})`;
  } else if (item.material_efficiency !== undefined || item.time_efficiency !== undefined) {
    typeLabel = `${t("blueprint")} (ME: ${item.material_efficiency || 0}, TE: ${item.time_efficiency || 0})`;
  } else if (item.flag !== undefined && item.flag >= 46 && item.flag <= 53) {
    typeLabel = `Fitted (Rig Slot ${item.flag - 46})`;
  } else if (item.singleton) {
    typeLabel = "Likely fitted/singleton";
  }

  const damagePercent = item.damage ? Math.round(item.damage * 100) : 0;

  return (
    <tr className={`border-b border-eve-border last:border-b-0 ${highlightRig ? "bg-yellow-900/20" : ""}`}>
      <td className="px-3 py-1.5">
        <div className="flex items-center gap-2">
          <img
            src={`https://images.evetech.net/types/${item.type_id}/icon?size=32`}
            alt={item.type_name}
            className="w-8 h-8 flex-shrink-0"
            onError={(e) => {
              // Fallback if icon fails to load
              e.currentTarget.style.display = 'none';
            }}
          />
          <div className="flex-1">
            <div className="text-eve-text flex items-center gap-2">
              <span>{item.type_name || `Type ${item.type_id}`}</span>
              {highlightRig && (
                <span className="px-1.5 py-0.5 rounded-sm border border-yellow-600/70 text-[10px] uppercase tracking-wider text-yellow-300">
                  {t("contractRigExcludedTag")}
                </span>
              )}
            </div>
            <div className="text-[10px] text-eve-dim/80 font-mono">type_id: {item.type_id}</div>
            {damagePercent > 0 && (
              <div className="text-xs text-red-400">⚠ Damaged {damagePercent}%</div>
            )}
          </div>
        </div>
      </td>
      <td className="px-3 py-1.5 text-right font-mono text-eve-accent">
        {item.quantity.toLocaleString()}
      </td>
      <td className="px-3 py-1.5 text-xs text-eve-dim">{typeLabel}</td>
    </tr>
  );
}
