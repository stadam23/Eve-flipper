import { useState } from "react";
import { useI18n } from "@/lib/i18n";

const WIKI_BASE = "https://github.com/ilyaux/Eve-flipper/wiki";

interface TabHelpProps {
  /** Translation key for step 1, 2, 3... (e.g. "helpFlipperStep1") */
  stepKeys: string[];
  wikiSlug: string;
}

export function TabHelp({ stepKeys, wikiSlug }: TabHelpProps) {
  const { t } = useI18n();
  const [open, setOpen] = useState(false);

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="w-6 h-6 rounded-full border border-eve-border bg-eve-panel text-eve-dim hover:text-eve-accent hover:border-eve-accent/50 flex items-center justify-center text-xs font-bold transition-colors"
        title={t("helpTitle")}
        aria-label={t("helpTitle")}
      >
        ?
      </button>
      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} aria-hidden="true" />
          <div className="absolute right-0 top-full mt-1 z-50 w-72 bg-eve-panel border border-eve-border rounded-sm shadow-eve-glow p-3">
            <div className="text-[11px] uppercase tracking-wider text-eve-dim mb-2">{t("helpTitle")}</div>
            <ol className="list-decimal list-inside space-y-1 text-sm text-eve-text mb-3">
              {stepKeys.map((key, i) => (
                <li key={i}>{t(key as "helpFlipperStep1")}</li>
              ))}
            </ol>
            <a
              href={`${WIKI_BASE}/${wikiSlug}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-xs text-eve-accent hover:text-eve-accent-hover"
            >
              {t("emptyReadWiki")} →
            </a>
          </div>
        </>
      )}
    </div>
  );
}
