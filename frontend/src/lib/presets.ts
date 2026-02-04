import type { ScanParams } from "./types";

const STORAGE_KEY = "eve-flipper-presets";

export interface SavedPreset {
  id: string;
  name: string;
  params: Partial<ScanParams>;
}

export const BUILTIN_PRESETS: { id: string; nameKey: string; params: Partial<ScanParams> }[] = [
  {
    id: "conservative",
    nameKey: "presetConservative",
    params: {
      min_margin: 10,
      min_contract_price: 50_000_000,
      max_contract_margin: 80,
      min_priced_ratio: 0.9,
      min_daily_volume: 10,
      max_results: 50,
    },
  },
  {
    id: "normal",
    nameKey: "presetNormal",
    params: {
      min_margin: 5,
      min_contract_price: 10_000_000,
      max_contract_margin: 100,
      min_priced_ratio: 0.8,
      min_daily_volume: 0,
      max_results: 100,
    },
  },
  {
    id: "aggressive",
    nameKey: "presetAggressive",
    params: {
      min_margin: 2,
      min_contract_price: 1_000_000,
      max_contract_margin: 150,
      min_priced_ratio: 0.7,
      min_daily_volume: 0,
      max_results: 250,
    },
  },
];

export function loadCustomPresets(): SavedPreset[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw) as SavedPreset[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

export function saveCustomPreset(preset: SavedPreset): void {
  const list = loadCustomPresets();
  const idx = list.findIndex((p) => p.id === preset.id);
  if (idx >= 0) list[idx] = preset;
  else list.push(preset);
  localStorage.setItem(STORAGE_KEY, JSON.stringify(list));
}

export function deleteCustomPreset(id: string): void {
  const list = loadCustomPresets().filter((p) => p.id !== id);
  localStorage.setItem(STORAGE_KEY, JSON.stringify(list));
}

export function applyPreset(current: ScanParams, presetParams: Partial<ScanParams>): ScanParams {
  return { ...current, ...presetParams };
}

export function nextPresetId(): string {
  return `custom-${Date.now()}`;
}
