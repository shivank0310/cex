import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { AuthResponse, AuthUser } from "@/types";
import * as authApi from "@/lib/api/auth";

interface AuthState {
  user: AuthUser | null;
  accessToken: string | null;
  refreshToken: string | null;
  hydrated: boolean;
  setHydrated: () => void;
  setSession: (response: AuthResponse) => void;
  clearSession: () => void;
  logout: () => Promise<void>;
  refreshSession: () => Promise<boolean>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      hydrated: false,
      setHydrated: () => set({ hydrated: true }),
      setSession: (response) =>
        set({
          user: response.user,
          accessToken: response.token.access_token,
          refreshToken: response.token.refresh_token,
        }),
      clearSession: () =>
        set({ user: null, accessToken: null, refreshToken: null }),
      logout: async () => {
        const { refreshToken } = get();
        if (refreshToken) {
          try {
            await authApi.logout(refreshToken);
          } catch {
            /* ignore */
          }
        }
        set({ user: null, accessToken: null, refreshToken: null });
      },
      refreshSession: async () => {
        const { refreshToken } = get();
        if (!refreshToken) return false;
        try {
          const response = await authApi.refresh(refreshToken);
          get().setSession(response);
          return true;
        } catch {
          get().clearSession();
          return false;
        }
      },
    }),
    {
      name: "cex-auth",
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
      }),
      onRehydrateStorage: () => (state) => {
        state?.setHydrated();
      },
    }
  )
);

export function useIsAuthenticated(): boolean {
  const user = useAuthStore((s) => s.user);
  const token = useAuthStore((s) => s.accessToken);
  return Boolean(user && token);
}
