import { useState, useCallback } from "react";
import { SystemAutocomplete } from "./SystemAutocomplete";
import { useI18n, type TranslationKey } from "@/lib/i18n";
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

const inputClass =
  "w-full px-2.5 py-1.5 bg-eve-input border border-eve-border rounded text-eve-text text-sm font-mono " +
  "focus:outline-none focus:border-eve-accent focus:ring-1 focus:ring-eve-accent/30 transition-colors " +
  "[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none";

export function ParametersPanel({ params, onChange, isLoggedIn = false, tab = "radius" }: Props) {
  const { t } = useI18n();
  const { addToast } = useGlobalToast();
  const [customPresets, setCustomPresets] = useState<SavedPreset[]>(() => loadCustomPresets());
  const [showAdvanced, setShowAdvanced] = useState(false);
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
        min_route_security: params.min_route_security,
      },
    };
    saveCustomPreset(preset);
    setCustomPresets(loadCustomPresets());
    applyPresetParams(preset.params);
    addToast(t("presetSaved"), "success", 2000);
  };

  return (
    <div className="bg-eve-panel border border-eve-border rounded-sm overflow-hidden">
      {/* Header: help + preset row */}
      <div className="flex items-center justify-between gap-3 px-3 py-2 border-b border-eve-border/60 bg-eve-panel/80">
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <span className="text-[10px] uppercase tracking-wider text-eve-dim font-medium shrink-0">
            {t("presetLabel")}
          </span>
          <select
            value=""
            onChange={(e) => handlePresetChange(e.target.value)}
            className={`flex-1 min-w-0 max-w-[140px] ${inputClass} py-1`}
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
            className="shrink-0 w-7 h-7 flex items-center justify-center bg-eve-input border border-eve-border rounded text-eve-dim hover:text-eve-accent hover:border-eve-accent/50 text-sm"
            title={t("presetSave")}
          >
            +
          </button>
        </div>
        {help && <TabHelp stepKeys={help.steps} wikiSlug={help.wiki} />}
      </div>

      <div className="p-4 w-full">
        {/* Full-width grid: 4 equal columns, labels stay on one line */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-x-6 gap-y-5 w-full">
          <Field label={t("system")} title={t("system")}>
            <SystemAutocomplete
              value={params.system_name}
              onChange={(v) => set("system_name", v)}
              isLoggedIn={isLoggedIn}
            />
          </Field>
          <Field label={t("paramsCargo")} title={t("cargoCapacity")}>
            <NumberInput
              value={params.cargo_capacity}
              onChange={(v) => set("cargo_capacity", v)}
              min={1}
              max={1000000}
              className={inputClass}
            />
          </Field>
          <Field label={t("paramsBuy")} title={t("buyRadius")}>
            <NumberInput
              value={params.buy_radius}
              onChange={(v) => set("buy_radius", v)}
              min={1}
              max={50}
              className={inputClass}
            />
          </Field>
          <Field label={t("paramsSell")} title={t("sellRadius")}>
            <NumberInput
              value={params.sell_radius}
              onChange={(v) => set("sell_radius", v)}
              min={1}
              max={50}
              className={inputClass}
            />
          </Field>
          <Field label={t("paramsMargin")} title={t("minMargin")}>
            <NumberInput
              value={params.min_margin}
              onChange={(v) => set("min_margin", v)}
              min={0.1}
              max={1000}
              step={0.1}
              className={inputClass}
            />
          </Field>
          <Field label={t("paramsTax")} title={t("salesTax")}>
            <NumberInput
              value={params.sales_tax_percent}
              onChange={(v) => set("sales_tax_percent", v)}
              min={0}
              max={100}
              step={0.1}
              className={inputClass}
            />
          </Field>
          <Field label={t("paramsResults")} title={t("maxResults")}>
            <select
              value={params.max_results ?? 100}
              onChange={(e) => set("max_results", parseInt(e.target.value))}
              className={inputClass}
            >
              <option value={50}>50</option>
              <option value={100}>100</option>
              <option value={250}>250</option>
              <option value={500}>500</option>
              <option value={1000}>1000</option>
            </select>
          </Field>
          <RouteSecurityField
            value={params.min_route_security ?? 0}
            onChange={(v) => set("min_route_security", v)}
            inputClass={inputClass}
            t={t}
          />
        </div>

        {/* Advanced: collapsible */}
        <div className="border-t border-eve-border/50 mt-4 pt-3">
          <button
            type="button"
            onClick={() => setShowAdvanced((a) => !a)}
            className="flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-eve-dim hover:text-eve-accent font-medium transition-colors"
          >
            <span className={`transition-transform ${showAdvanced ? "rotate-90" : ""}`}>▸</span>
            {t("advancedFilters")}
          </button>
          {showAdvanced && (
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-x-6 gap-y-5 mt-3 w-full">
              <Field label={t("minDailyVolume")}>
                <NumberInput
                  value={params.min_daily_volume ?? 0}
                  onChange={(v) => set("min_daily_volume", v)}
                  min={0}
                  max={999999999}
                  className={inputClass}
                />
              </Field>
              <Field label={t("maxInvestment")}>
                <NumberInput
                  value={params.max_investment ?? 0}
                  onChange={(v) => set("max_investment", v)}
                  min={0}
                  max={999999999999}
                  className={inputClass}
                />
              </Field>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

const ROUTE_SECURITY_PRESETS = [0, 0.45, 0.5, 0.6, 0.7, 0.8, 0.9] as const;

function RouteSecurityField({
  value,
  onChange,
  inputClass,
  t,
}: {
  value: number;
  onChange: (v: number) => void;
  inputClass: string;
  t: (key: TranslationKey) => string;
}) {
  const isPreset = ROUTE_SECURITY_PRESETS.some((p) => Math.abs(p - value) < 1e-6);
  const selectValue = isPreset ? String(value) : "custom";

  return (
    <Field label={t("paramsSecurity")} title={t("routeSecurityHint")}>
      <div className="flex gap-2 items-stretch">
        <select
          value={selectValue}
          onChange={(e) => {
            const v = e.target.value;
            if (v === "custom") {
              // set to a value not in presets so that selectValue becomes "custom" and input appears
              onChange(0.35);
              return;
            }
            onChange(parseFloat(v));
          }}
          className={`flex-1 min-w-0 ${inputClass}`}
        >
          <option value="0">{t("routeSecurityAll")}</option>
          <option value="0.45">{t("routeSecurityHighsec")}</option>
          <option value="0.5">{t("routeSecurityMin05")}</option>
          <option value="0.6">{t("routeSecurityMin06")}</option>
          <option value="0.7">{t("routeSecurityMin07")}</option>
          <option value="0.8">{t("routeSecurityMin08")}</option>
          <option value="0.9">{t("routeSecurityMin09")}</option>
          <option value="custom">{t("routeSecurityCustom")}</option>
        </select>
        {selectValue === "custom" && (
          <input
            type="number"
            min={-1}
            max={1}
            step={0.1}
            value={value}
            onChange={(e) => {
              const v = parseFloat(e.target.value);
              if (!Number.isNaN(v)) onChange(Math.max(-1, Math.min(1, v)));
            }}
            className={`w-16 ${inputClass}`}
            title={t("routeSecurityHint")}
          />
        )}
      </div>
    </Field>
  );
}

function Field({
  label,
  children,
  className = "",
  title,
}: {
  label: string;
  children: React.ReactNode;
  className?: string;
  title?: string;
}) {
  return (
    <div className={`flex flex-col gap-1.5 min-w-0 ${className}`} title={title}>
      <label className="text-[10px] uppercase tracking-wider text-eve-dim font-medium whitespace-nowrap truncate">
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
  className = "",
}: {
  value: number;
  onChange: (v: number) => void;
  min: number;
  max: number;
  step?: number;
  className?: string;
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
      className={className || inputClass}
    />
  );
}
