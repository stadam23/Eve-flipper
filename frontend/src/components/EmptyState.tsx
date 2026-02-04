import { useI18n } from "@/lib/i18n";

export type EmptyReason =
  | "no_scan_yet"
  | "no_results"
  | "esi_offline"
  | "filters_too_strict"
  | "no_stations"
  | "no_item_selected"
  | "loading";

interface EmptyStateProps {
  reason: EmptyReason;
  /** Extra hint lines (e.g. which filters applied) */
  hints?: string[];
  /** Optional wiki page path (e.g. "Station-Trading") */
  wikiSlug?: string;
}

const WIKI_BASE = "https://github.com/ilyaux/Eve-flipper/wiki";

const TITLE_KEYS: Record<EmptyReason, string> = {
  no_scan_yet: "emptyTitle_no_scan_yet",
  no_results: "emptyTitle_no_results",
  esi_offline: "emptyTitle_esi_offline",
  filters_too_strict: "emptyTitle_filters_too_strict",
  no_stations: "emptyTitle_no_stations",
  no_item_selected: "emptyTitle_no_item_selected",
  loading: "emptyTitle_loading",
};

const DESC_KEYS: Record<EmptyReason, string> = {
  no_scan_yet: "emptyDesc_no_scan_yet",
  no_results: "emptyDesc_no_results",
  esi_offline: "emptyDesc_esi_offline",
  filters_too_strict: "emptyDesc_filters_too_strict",
  no_stations: "emptyDesc_no_stations",
  no_item_selected: "emptyDesc_no_item_selected",
  loading: "emptyDesc_loading",
};

export function EmptyState({ reason, hints = [], wikiSlug }: EmptyStateProps) {
  const { t } = useI18n();
  const title = t(TITLE_KEYS[reason] as "emptyTitle_no_scan_yet");
  const desc = t(DESC_KEYS[reason] as "emptyDesc_no_scan_yet");

  const wikiUrl = wikiSlug ? `${WIKI_BASE}/${wikiSlug}` : `${WIKI_BASE}`;

  return (
    <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
      <div className="w-14 h-14 rounded-full bg-eve-panel border border-eve-border flex items-center justify-center mb-4 text-eve-dim">
        {reason === "no_scan_yet" || reason === "no_item_selected" ? (
          <svg className="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        ) : reason === "no_results" || reason === "filters_too_strict" ? (
          <svg className="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        ) : reason === "esi_offline" ? (
          <svg className="w-7 h-7 text-eve-error" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M18.364 5.636a9 9 0 010 12.728m0 0l-2.829-2.829m2.829 2.829L21 21M15.536 8.464a5 5 0 010 7.072m0 0l-2.829-2.829m-4.243 2.829a4.978 4.978 0 01-1.414-2.83m-1.414 5.658a9 9 0 01-2.167-9.238m7.824 2.167a1 1 0 111.414 1.414m-1.414-1.414L3 3m8.293 8.293l1.414 1.414" />
          </svg>
        ) : reason === "loading" ? (
          <svg className="w-7 h-7 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
        ) : (
          <svg className="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.121 2.121a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.121-2.121A1 1 0 006.586 13H4" />
          </svg>
        )}
      </div>
      <h3 className="text-sm font-semibold text-eve-text mb-1">{title}</h3>
      <p className="text-xs text-eve-dim max-w-sm mb-3">{desc}</p>
      {hints.length > 0 && (
        <ul className="text-xs text-eve-dim text-left list-disc list-inside mb-3 space-y-0.5">
          {hints.map((h, i) => (
            <li key={i}>{h}</li>
          ))}
        </ul>
      )}
      {wikiSlug && (
        <a
          href={wikiUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="text-xs text-eve-accent hover:text-eve-accent-hover transition-colors"
        >
          {t("emptyReadWiki")} →
        </a>
      )}
    </div>
  );
}
