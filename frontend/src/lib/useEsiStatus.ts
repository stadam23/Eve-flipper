import { useEffect, useState } from "react";
import { getStatus } from "./api";

interface UseEsiStatusReturn {
  /** `true` = ESI reachable, `false` = down, `null` = still loading initial check */
  esiAvailable: boolean | null;
}

const ESI_POLL_INTERVAL_MS = 5000;
const ESI_FAILS_FOR_DOWN = 3;
const ESI_LAST_OK_GRACE_SEC = 60;

/**
 * Polls the backend `/api/status` endpoint every 5 seconds to track
 * whether the EVE ESI is reachable.  Returns `null` while the first
 * check is in-flight.
 */
export function useEsiStatus(): UseEsiStatusReturn {
  const [esiAvailable, setEsiAvailable] = useState<boolean | null>(null);

  useEffect(() => {
    let mounted = true;
    let failedChecks = 0;
    const checkEsi = async () => {
      try {
        const status = await getStatus();
        if (!mounted) return;

        if (status.esi_ok) {
          failedChecks = 0;
          setEsiAvailable(true);
          return;
        }

        const nowSec = Math.floor(Date.now() / 1000);
        const lastOK = status.esi_last_ok ?? 0;
        const recentlyHealthy = lastOK > 0 && nowSec-lastOK <= ESI_LAST_OK_GRACE_SEC;
        if (recentlyHealthy) {
          failedChecks = 0;
          setEsiAvailable(true);
          return;
        }

        failedChecks += 1;
        if (failedChecks >= ESI_FAILS_FOR_DOWN) {
          setEsiAvailable(false);
          return;
        }

        // Keep current state during short transient ESI flaps.
        setEsiAvailable((prev) => (prev === null ? null : true));
      } catch {
        if (!mounted) return;
        failedChecks += 1;
        if (failedChecks >= ESI_FAILS_FOR_DOWN) {
          setEsiAvailable(false);
        }
      }
    };
    checkEsi();
    const interval = setInterval(checkEsi, ESI_POLL_INTERVAL_MS);
    return () => {
      mounted = false;
      clearInterval(interval);
    };
  }, []);

  return { esiAvailable };
}
