"use client";

import { useEffect } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useTradingStore } from "@/store/trading-store";

/** Syncs auth session into the trading store for order/wallet API calls. */
export function AppProviders({ children }: { children: React.ReactNode }) {
  const hydrated = useAuthStore((s) => s.hydrated);
  const user = useAuthStore((s) => s.user);
  const accessToken = useAuthStore((s) => s.accessToken);
  const setUserId = useTradingStore((s) => s.setUserId);
  const setAuthToken = useTradingStore((s) => s.setAuthToken);

  useEffect(() => {
    if (!hydrated) return;
    if (user && accessToken) {
      setUserId(user.id);
      setAuthToken(accessToken);
    } else {
      setUserId("");
      setAuthToken("");
    }
  }, [hydrated, user, accessToken, setUserId, setAuthToken]);

  return <>{children}</>;
}
