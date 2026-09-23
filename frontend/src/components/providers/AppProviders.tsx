"use client";

import { useEffect, useRef } from "react";
import { useAuthStore } from "@/store/auth-store";
import { useTradingStore } from "@/store/trading-store";

/** Syncs auth session into the trading store and restores session after refresh. */
export function AppProviders({ children }: { children: React.ReactNode }) {
  const hydrated = useAuthStore((s) => s.hydrated);
  const user = useAuthStore((s) => s.user);
  const accessToken = useAuthStore((s) => s.accessToken);
  const refreshToken = useAuthStore((s) => s.refreshToken);
  const refreshSession = useAuthStore((s) => s.refreshSession);
  const setUserId = useTradingStore((s) => s.setUserId);
  const setAuthToken = useTradingStore((s) => s.setAuthToken);
  const bootstrapped = useRef(false);

  useEffect(() => {
    void useAuthStore.persist.rehydrate();
  }, []);

  useEffect(() => {
    if (!hydrated || bootstrapped.current) return;
    bootstrapped.current = true;
    if (refreshToken) {
      void refreshSession();
    }
  }, [hydrated, refreshToken, refreshSession]);

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
