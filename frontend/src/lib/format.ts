import type { Locale } from "./i18n";

// Get browser locale or use provided locale
function getLocaleString(locale?: Locale): string {
  if (locale === "ru") return "ru-RU";
  if (locale === "en") return "en-US";
  // Fallback to browser locale
  return navigator.language || "en-US";
}

export function formatISK(value: number, locale?: Locale): string {
  if (value == null || isNaN(value)) return "0";
  const localeStr = getLocaleString(locale);
  const abs = Math.abs(value);
  const sign = value < 0 ? "-" : "";
  if (abs >= 1_000_000_000) {
    return sign + (abs / 1_000_000_000).toLocaleString(localeStr, { maximumFractionDigits: 2 }) + " B";
  }
  if (abs >= 1_000_000) {
    return sign + (abs / 1_000_000).toLocaleString(localeStr, { maximumFractionDigits: 2 }) + " M";
  }
  if (abs >= 1_000) {
    return sign + (abs / 1_000).toLocaleString(localeStr, { maximumFractionDigits: 1 }) + " K";
  }
  return value.toLocaleString(localeStr, { maximumFractionDigits: 1 });
}

export function formatMargin(value: number, locale?: Locale): string {
  const localeStr = getLocaleString(locale);
  return value.toLocaleString(localeStr, { minimumFractionDigits: 1, maximumFractionDigits: 1 }) + "%";
}

export function formatNumber(value: number, locale?: Locale): string {
  const localeStr = getLocaleString(locale);
  return value.toLocaleString(localeStr);
}

// Format ISK with full precision (no abbreviations)
export function formatISKFull(value: number, locale?: Locale): string {
  const localeStr = getLocaleString(locale);
  return value.toLocaleString(localeStr, { maximumFractionDigits: 0 });
}
