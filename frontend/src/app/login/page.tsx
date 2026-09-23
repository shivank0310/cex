"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";
import { Button } from "@/components/ui/Button";
import { login } from "@/lib/api/auth";
import { useAuthStore } from "@/store/auth-store";

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const setSession = useAuthStore((s) => s.setSession);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const response = await login(email, password);
      setSession(response);
      const redirect = searchParams.get("redirect") || "/trade";
      router.replace(redirect);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="mt-8 space-y-4 rounded-2xl border border-white/10 bg-[#12121a] p-6">
      <label className="block">
        <span className="text-xs text-slate-500">Email</span>
        <input
          type="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="mt-1 w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-white"
        />
      </label>
      <label className="block">
        <span className="text-xs text-slate-500">Password</span>
        <input
          type="password"
          required
          minLength={8}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="mt-1 w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-white"
        />
      </label>
      {error && <p className="text-sm text-rose-400">{error}</p>}
      <Button type="submit" className="w-full" disabled={loading}>
        {loading ? "Signing in..." : "Sign In"}
      </Button>
    </form>
  );
}

export default function LoginPage() {
  return (
    <div className="mx-auto flex min-h-[70vh] max-w-md flex-col justify-center px-4 py-12">
      <h1 className="text-2xl font-bold text-white">Sign in</h1>
      <p className="mt-2 text-sm text-slate-400">
        Authenticate via auth-service · JWT for protected API routes
      </p>

      <Suspense fallback={<p className="mt-8 text-slate-500">Loading...</p>}>
        <LoginForm />
      </Suspense>

      <p className="mt-4 text-center text-sm text-slate-400">
        No account?{" "}
        <Link href="/register" className="text-cyan-400 hover:underline">
          Register
        </Link>
      </p>
    </div>
  );
}
