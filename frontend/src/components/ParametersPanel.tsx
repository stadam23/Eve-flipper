import { useState, useCallback } from "react";
import { SystemAutocomplete } from "./SystemAutocomplete";
import { useI18n } from "@/lib/i18n";
import { useGlobalToast } from "./Toast";
import { TabHelp } from "./TabHelp";
import type { ScanParams } from "@/lib/types";
import {
  BUILTIN_PRESETS,
  loadCustomPresets,
  saveCustomPreset,
  applyPreset,
  nextPresetId,
  type SavedPreset,
} from "@/lib/presets";

type TabForParams = "radius" | "region" | "contracts" | "route";

interface Props {
  params: ScanParams;
  onChange: (params: ScanParams) => void;
  isLoggedIn?: boolean;
  tab?: TabForParams;
}

const HELP_STEPS: Record<TabForParams, { steps: string[]; wiki: string }> = {
  radius: { steps: ["helpFlipperStep1", "helpFlipperStep2", "helpFlipperStep3"], wiki: "Getting-Started" },
  region: { steps: ["helpFlipperStep1", "helpFlipperStep2", "helpFlipperStep3"], wiki: "Getting-Started" },
  contracts: { steps: ["helpContractsStep1", "helpContractsStep2", "helpContractsStep3"], wiki: "Contract-Arbitrage" },
  route: { steps: ["helpRouteStep1", "helpRouteStep2", "helpRouteStep3"], wiki: "Route-Builder" },
};

export function ParametersPanel({ params, onChange, isLoggedIn = false, tab = "radius" }: Props) {
  const { t } = useI18n();
  const { addToast } = useGlobalToast();
  const [customPresets, setCustomPresets] = useState<SavedPreset[]>(() => loadCustomPresets());
  const help = HELP_STEPS[tab];

  const set = <K extends keyof ScanParams>(key: K, value: ScanParams[K]) => {
    onChange({ ...params, [key]: value });
  };

  const applyPresetParams = useCallback(
    (presetParams: Partial<ScanParams>) => {
      onChange(applyPreset(params, presetParams));
    },
    [params, onChange]
  );

  const handlePresetChange = (value: string) => {
    if (value === "" || value === "save") return;
    const builtin = BUILTIN_PRESETS.find((p) => p.id === value);
    if (builtin) {
      applyPresetParams(builtin.params);
      return;
    }
    const custom = customPresets.find((p) => p.id === value);
    if (custom) applyPresetParams(custom.params);
  };

  const handleSavePreset = () => {
    const name = window.prompt(t("presetSave"), "My preset");
    if (!name?.trim()) return;
    const preset: SavedPreset = {
      id: nextPresetId(),
      name: name.trim(),
      params: {
        min_margin: params.min_margin,
        min_contract_price: params.min_contract_price,
        max_contract_margin: params.max_contract_margin,
        min_priced_ratio: params.min_priced_ratio,
        min_daily_volume: params.min_daily_volume,
        max_results: params.max_results,
      },
    };
    saveCustomPreset(preset);
    setCustomPresets(loadCustomPresets());
    applyPresetParams(preset.params);
    addToast(t("presetSaved"), "success", 2000);
  };

  return (
    <div className="bg-eve-panel border border-eve-border rounded-sm p-3 sm:p-4">
      {help && (
        <div className="flex justify-end mb-2">
          <TabHelp stepKeys={help.steps} wikiSlug={help.wiki} />
        </div>
      )}
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-x-3 gap-y-3 items-end">
        <Field label={t("presetLabel")}>
          <div className="flex gap-1">
            <select
              value=""
              onChange={(e) => handlePresetChange(e.target.value)}
              className="flex-1 px-3 py-1.5 bg-eve-input border border-eve-border rounded-sm text-eve-text text-sm font-mono
                         focus:outline-none focus:border-eve-accent focus:ring-1 focus:ring-eve-accent/30"
            >
              <option value="">—</option>
              {BUILTIN_PRESETS.map((p) => (
                <option key={p.id} value={p.id}>
                  {t(p.nameKey as "presetConservative")}
                </option>
              ))}
              {customPresets.length > 0 &&
                customPresets.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
            </select>
            <button
              type="button"
              onClick={handleSavePreset}
              className="px-2 py-1.5 bg-eve-input border border-eve-border rounded-sm text-eve-dim hover:text-eve-accent text-xs whitespace-nowrap"
              title={t("presetSave")}
            >
              +
            </button>
          </div>
        </Field>

        <Field label={t("system")}>
          <SystemAutocomplete
            value={params.system_name}
            onChange={(v) => set("system_name", v)}
            isLoggedIn={isLoggedIn}
          />
        </Field>

        <Field label={t("cargoCapacity")}>
          <NumberInput
            value={params.cargo_capacity}
            onChange={(v) => set("cargo_capacity", v)}
            min={1}
            max={1000000}
          />
        </Field>

        <Field label={t("buyRadius")}>
          <NumberInput
            value={params.buy_radius}
            onChange={(v) => set("buy_radius", v)}
            min={1}
            max={50}
          />
        </Field>

        <Field label={t("sellRadius")}>
          <NumberInput
            value={params.sell_radius}
            onChange={(v) => set("sell_radius", v)}
            min={1}
            max={50}
          />
        </Field>

        <Field label={t("minMargin")}>
          <NumberInput
            value={params.min_margin}
            onChange={(v) => set("min_margin", v)}
            min={0.1}
            max={1000}
            step={0.1}
          />
        </Field>

        <Field label={t("salesTax")}>
          <NumberInput
            value={params.sales_tax_percent}
            onChange={(v) => set("sales_tax_percent", v)}
            min={0}
            max={100}
            step={0.1}
          />
        </Field>

        <Field label={t("minDailyVolume")}>
          <NumberInput
            value={params.min_daily_volume ?? 0}
            onChange={(v) => set("min_daily_volume", v)}
            min={0}
            max={999999999}
          />
        </Field>

        <Field label={t("maxInvestment")}>
          <NumberInput
            value={params.max_investment ?? 0}
            onChange={(v) => set("max_investment", v)}
            min={0}
            max={999999999999}
          />
        </Field>

        <Field label={t("maxResults")}>
          <select
            value={params.max_results ?? 100}
            onChange={(e) => set("max_results", parseInt(e.target.value))}
            className="w-full px-3 py-1.5 bg-eve-input border border-eve-border rounded-sm text-eve-text text-sm font-mono
                       focus:outline-none focus:border-eve-accent focus:ring-1 focus:ring-eve-accent/30
                       transition-colors"
          >
            <option value={50}>50</option>
            <option value={100}>100</option>
            <option value={250}>250</option>
            <option value={500}>500</option>
            <option value={1000}>1000</option>
          </select>
        </Field>
      </div>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <label className="text-[11px] uppercase tracking-wider text-eve-dim font-medium">
        {label}
      </label>
      {children}
    </div>
  );
}

function NumberInput({
  value,
  onChange,
  min,
  max,
  step = 1,
}: {
  value: number;
  onChange: (v: number) => void;
  min: number;
  max: number;
  step?: number;
}) {
  return (
    <input
      type="number"
      value={value}
      onChange={(e) => {
        const v = parseFloat(e.target.value);
        if (!isNaN(v) && v >= min && v <= max) onChange(v);
      }}
      min={min}
      max={max}
      step={step}
      className="w-full px-3 py-1.5 bg-eve-input border border-eve-border rounded-sm text-eve-text text-sm font-mono
                 focus:outline-none focus:border-eve-accent focus:ring-1 focus:ring-eve-accent/30
                 transition-colors
                 [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
    />
  );
}
