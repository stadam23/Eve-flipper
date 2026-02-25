import { useMemo } from "react";
import { useI18n } from "@/lib/i18n";
import { formatISK } from "@/lib/format";
import type { FlatMaterial } from "@/lib/types";

interface IndustryShoppingListProps {
  materials: FlatMaterial[];
  regionId: number;
  onOpenExecutionPlan: (material: FlatMaterial) => void;
}

export function IndustryShoppingList({
  materials,
  regionId,
  onOpenExecutionPlan,
}: IndustryShoppingListProps) {
  const { t } = useI18n();

  const totalCost = useMemo(
    () => materials.reduce((sum, material) => sum + material.total_price, 0),
    [materials]
  );

  const totalVolume = useMemo(
    () => materials.reduce((sum, material) => sum + material.volume, 0),
    [materials]
  );

  return (
    <div>
      <table className="w-full text-sm">
        <thead className="sticky top-0 bg-eve-dark z-10">
          <tr className="text-eve-dim text-[10px] uppercase tracking-wider border-b border-eve-border">
            <th className="min-w-[32px] px-1 py-2" />
            <th className="px-3 py-2 text-left font-medium">Item</th>
            <th className="px-3 py-2 text-right font-medium">Quantity</th>
            <th className="px-3 py-2 text-right font-medium">Unit Price</th>
            <th className="px-3 py-2 text-right font-medium">Total</th>
            <th className="px-3 py-2 text-right font-medium">Volume</th>
          </tr>
        </thead>
        <tbody>
          {materials.map((material, index) => (
            <tr
              key={material.type_id}
              className={`border-b border-eve-border/50 hover:bg-eve-accent/5 ${
                index % 2 === 0 ? "bg-eve-panel" : "bg-eve-dark"
              }`}
            >
              <td className="px-1 py-1.5 text-center">
                {regionId > 0 && (
                  <button
                    type="button"
                    onClick={() => onOpenExecutionPlan(material)}
                    className="text-eve-dim hover:text-eve-accent transition-colors text-sm"
                    title={t("execPlanTitle")}
                  >
                    📊
                  </button>
                )}
              </td>
              <td className="px-3 py-1.5 text-eve-text">{material.type_name}</td>
              <td className="px-3 py-1.5 text-right font-mono text-eve-accent">
                {material.quantity.toLocaleString()}
              </td>
              <td className="px-3 py-1.5 text-right font-mono text-eve-dim">
                {formatISK(material.unit_price)}
              </td>
              <td className="px-3 py-1.5 text-right font-mono text-eve-accent">
                {formatISK(material.total_price)}
              </td>
              <td className="px-3 py-1.5 text-right font-mono text-eve-dim">
                {material.volume.toLocaleString(undefined, { maximumFractionDigits: 1 })} m3
              </td>
            </tr>
          ))}
        </tbody>
        <tfoot className="bg-eve-dark border-t border-eve-border">
          <tr>
            <td className="px-1 py-2" />
            <td className="px-3 py-2 text-eve-dim font-medium">Total</td>
            <td className="px-3 py-2 text-right font-mono text-eve-accent font-semibold">
              {materials.length} items
            </td>
            <td className="px-3 py-2" />
            <td className="px-3 py-2 text-right font-mono text-eve-accent font-semibold">
              {formatISK(totalCost)}
            </td>
            <td className="px-3 py-2 text-right font-mono text-eve-dim">
              {totalVolume.toLocaleString(undefined, { maximumFractionDigits: 1 })} m3
            </td>
          </tr>
        </tfoot>
      </table>
    </div>
  );
}
