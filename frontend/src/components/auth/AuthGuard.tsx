"use client";

import Link from "next/link";
import { useAuthStore, useIsAuthenticated } from "@/store/auth-store";
import { Button } from "@/components/ui/Button";

interface AuthGuardProps {
  children: React.ReactNode;
  title?: string;
}

export function AuthGuard({ children, title = "Sign in required" }: AuthGuardProps) {
  const hydrated = useAuthStore((s) => s.hydrated);
  const isAuthenticated = useIsAuthenticated();

  if (!hydrated) {
    return (
      <div className="flex min-h-[40vh] items-center justify-center text-slate-400">
        Loading session...
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="mx-auto max-w-md rounded-2xl border border-white/10 bg-[#12121a] p-8 text-center">
        <h2 className="text-xl font-semibold text-white">{title}</h2>
        <p className="mt-2 text-sm text-slate-400">
          Connect to auth-service via the API gateway to trade and manage your wallet.
        </p>
        <div className="mt-6 flex justify-center gap-3">
          <Link href="/login">
            <Button>Sign In</Button>
          </Link>
          <Link href="/register">
            <Button variant="secondary">Register</Button>
          </Link>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
