"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Activity, BarChart3, LayoutDashboard, LogOut, User, Wallet } from "lucide-react";
import clsx from "clsx";
import { Button } from "@/components/ui/Button";
import { useAuthStore, useIsAuthenticated } from "@/store/auth-store";

const NAV = [
  { href: "/", label: "Home", icon: LayoutDashboard },
  { href: "/markets", label: "Markets", icon: BarChart3 },
  { href: "/trade", label: "Trade", icon: Activity },
  { href: "/wallet", label: "Wallet", icon: Wallet },
];

export function Header() {
  const pathname = usePathname();
  const router = useRouter();
  const isAuthenticated = useIsAuthenticated();
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);

  async function handleLogout() {
    await logout();
    router.replace("/");
    router.refresh();
  }

  return (
    <header className="sticky top-0 z-50 border-b border-white/10 bg-[#0a0a12]/80 backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6">
        <Link href="/" className="flex items-center gap-2">
          <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-violet-600 to-cyan-500 font-bold text-white">
            C
          </div>
          <span className="text-lg font-bold text-white">CEX</span>
        </Link>

        <nav className="hidden items-center gap-1 md:flex">
          {NAV.map(({ href, label, icon: Icon }) => (
            <Link
              key={href}
              href={href}
              className={clsx(
                "flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors",
                pathname === href
                  ? "bg-white/10 text-white"
                  : "text-slate-400 hover:text-white hover:bg-white/5"
              )}
            >
              <Icon className="h-4 w-4" />
              {label}
            </Link>
          ))}
        </nav>

        <div className="flex items-center gap-3">
          {isAuthenticated && user ? (
            <>
              <span className="hidden items-center gap-2 text-sm text-slate-300 sm:flex">
                <User className="h-4 w-4" />
                {user.username || user.email}
              </span>
              <Button type="button" size="sm" variant="ghost" onClick={handleLogout} aria-label="Sign out">
                <LogOut className="h-4 w-4" />
              </Button>
            </>
          ) : (
            <>
              <Link href="/login">
                <Button size="sm" variant="ghost">Sign In</Button>
              </Link>
              <Link href="/register">
                <Button size="sm">Register</Button>
              </Link>
            </>
          )}
          <Link href="/trade">
            <Button size="sm" variant="secondary">Trade</Button>
          </Link>
        </div>
      </div>
    </header>
  );
}
