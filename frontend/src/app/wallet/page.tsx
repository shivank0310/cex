"use client";

import { AuthGuard } from "@/components/auth/AuthGuard";
import { WalletPanel } from "@/components/wallet/WalletPanel";

export default function WalletPage() {
  return (
    <AuthGuard title="Sign in to access wallet">
      <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6">
        <h1 className="text-2xl font-bold text-white">Wallet</h1>
        <p className="mt-2 text-slate-400">
          Ledger balances (off-chain) · Deposits & withdrawals via blockchain-service
        </p>
        <div className="mt-8">
          <WalletPanel />
        </div>
      </div>
    </AuthGuard>
  );
}
