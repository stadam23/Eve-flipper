import { useEffect, useRef, useState } from "react";
import { getStatus } from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import type { AppStatus } from "@/lib/types";

function formatTimeAgo(timestamp: number): string {
  const seconds = Math.floor(Date.now() / 1000 - timestamp);
  if (seconds < 60) return `${seconds}s ago`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  return `${Math.floor(seconds / 86400)}d ago`;
}

export function StatusBar() {
  const { t } = useI18n();
  const [status, setStatus] = useState<AppStatus | null>(null);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    
    const poll = async () => {
      try {
        const data = await getStatus();
        if (mountedRef.current) {
          setStatus(data);
        }
      } catch {
        // If we can't reach our own backend, show as offline
        if (mountedRef.current) {
          setStatus(prev => prev ? { ...prev, esi_ok: false } : null);
        }
      }
    };
    
    poll();
    const id = setInterval(poll, 5000);
    
    return () => {
      mountedRef.current = false;
      clearInterval(id);
    };
  }, []);

  // Build ESI label with more info when unavailable
  const getEsiLabel = () => {
    if (status === null) return t("esiApi");
    if (status.esi_ok) return t("esiApi");
    
    // ESI is down - show when it was last working
    if (status.esi_last_ok) {
      return `${t("esiUnavailable")} (${formatTimeAgo(status.esi_last_ok)})`;
    }
    return t("esiUnavailable");
  };

  const sdeOk =
    (status?.sde_loaded ?? false) ||
    ((status?.sde_systems ?? 0) > 0 && (status?.sde_types ?? 0) > 0);

  return (
    <div className="flex items-center gap-4 h-[34px] px-4 bg-eve-panel border border-eve-border rounded-sm">
      <StatusDot
        ok={sdeOk}
        loading={status === null}
        label={
          sdeOk
            ? `SDE: ${status?.sde_systems ?? 0} ${t("sdeSystems")}, ${status?.sde_types ?? 0} ${t("sdeTypes")}`
            : t("sdeLoading")
        }
      />
      <div className="w-px h-4 bg-eve-border" />
      <StatusDot
        ok={status?.esi_ok ?? false}
        loading={status === null}
        label={getEsiLabel()}
        warning={!status?.esi_ok && status !== null}
      />
    </div>
  );
}

function StatusDot({ ok, loading, label, warning }: { ok: boolean; loading: boolean; label: string; warning?: boolean }) {
  return (
    <div className="flex items-center gap-2 text-xs">
      <div
        className={`w-2 h-2 rounded-full ${
          loading
            ? "bg-eve-accent animate-pulse"
            : ok
              ? "bg-eve-success"
              : warning
                ? "bg-eve-error animate-pulse"
                : "bg-eve-error"
        }`}
      />
      <span className={ok ? "text-eve-text" : warning ? "text-eve-error" : "text-eve-dim"}>{label}</span>
    </div>
  );
}
