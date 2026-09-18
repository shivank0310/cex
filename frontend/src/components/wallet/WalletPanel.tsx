"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import { createDepositAddress, getBalance, requestWithdrawal } from "@/lib/api/wallet";
import { formatQuantity, parseQuantity, shortAddress } from "@/lib/format";
import { useAuthStore } from "@/store/auth-store";
import type { Balance, WalletAddress } from "@/types";

export function WalletPanel() {
  const user = useAuthStore((s) => s.user);
  const userId = user?.id ?? "";
  const [asset, setAsset] = useState("USDT");
  const [balance, setBalance] = useState<Balance | null>(null);
  const [address, setAddress] = useState<WalletAddress | null>(null);
  const [withdrawAddr, setWithdrawAddr] = useState("");
  const [withdrawAmt, setWithdrawAmt] = useState("");
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState("");

  async function loadBalance() {
    if (!userId) return;
    try {
      const b = await getBalance(userId, asset);
      setBalance(b);
    } catch {
      setBalance(null);
    }
  }

  useEffect(() => {
    if (userId) loadBalance();
  }, [userId, asset]);

  async function handleDepositAddress() {
    if (!userId) return;
    setLoading(true);
    setMsg("");
    try {
      const w = await createDepositAddress(userId, asset);
      setAddress(w);
      setMsg("Deposit address created");
      await loadBalance();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "Failed");
    } finally {
      setLoading(false);
    }
  }

  async function handleWithdraw() {
    if (!userId) return;
    setLoading(true);
    setMsg("");
    try {
      const amount = parseQuantity(withdrawAmt);
      const res = await requestWithdrawal({
        user_id: userId,
        asset,
        amount,
        to_address: withdrawAddr,
      });
      setMsg(`Withdrawal ${res.status} — ${res.tx_hash || res.id}`);
      await loadBalance();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "Withdrawal failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      <div className="rounded-2xl border border-white/10 bg-[#12121a] p-6">
        <h3 className="text-lg font-semibold text-white">Balances</h3>
        <p className="mt-1 text-xs text-slate-500">User: {user?.email}</p>
        <div className="mt-4 flex gap-2">
          {["USDT", "BTC"].map((a) => (
            <button
              key={a}
              onClick={() => setAsset(a)}
              className={`rounded-lg px-4 py-2 text-sm font-medium ${asset === a ? "bg-violet-600 text-white" : "bg-white/5 text-slate-400"}`}
            >
              {a}
            </button>
          ))}
          <Button size="sm" variant="ghost" onClick={loadBalance}>Refresh</Button>
        </div>
        {balance ? (
          <div className="mt-6 grid grid-cols-3 gap-4">
            <div>
              <p className="text-xs text-slate-500">Available</p>
              <p className="text-xl font-mono text-emerald-400">{formatQuantity(balance.available)}</p>
            </div>
            <div>
              <p className="text-xs text-slate-500">Locked</p>
              <p className="text-xl font-mono text-amber-400">{formatQuantity(balance.locked)}</p>
            </div>
            <div>
              <p className="text-xs text-slate-500">Total</p>
              <p className="text-xl font-mono text-white">{formatQuantity(balance.total)}</p>
            </div>
          </div>
        ) : (
          <p className="mt-6 text-sm text-slate-500">
            No balance yet — deposit funds or contact admin to credit your ledger account.
          </p>
        )}
      </div>

      <div className="rounded-2xl border border-white/10 bg-[#12121a] p-6">
        <h3 className="text-lg font-semibold text-white">Deposit</h3>
        <p className="mt-2 text-sm text-slate-400">
          On-chain deposit → blockchain-service → ledger credit
        </p>
        <Button className="mt-4" onClick={handleDepositAddress} disabled={loading || !userId}>
          Get Deposit Address
        </Button>
        {address && (
          <div className="mt-4 rounded-lg bg-white/5 p-3 font-mono text-sm text-cyan-300 break-all">
            {address.address}
            <p className="mt-1 text-xs text-slate-500">{address.chain} · {shortAddress(address.address, 8)}</p>
          </div>
        )}
      </div>

      <div className="rounded-2xl border border-white/10 bg-[#12121a] p-6 lg:col-span-2">
        <h3 className="text-lg font-semibold text-white">Withdraw</h3>
        <div className="mt-4 grid gap-4 sm:grid-cols-2">
          <input
            placeholder="To address (0x...)"
            value={withdrawAddr}
            onChange={(e) => setWithdrawAddr(e.target.value)}
            className="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-white"
          />
          <input
            placeholder="Amount"
            value={withdrawAmt}
            onChange={(e) => setWithdrawAmt(e.target.value)}
            className="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-white"
          />
        </div>
        <Button className="mt-4" variant="secondary" onClick={handleWithdraw} disabled={loading || !userId}>
          Request Withdrawal
        </Button>
        {msg && <p className="mt-3 text-sm text-cyan-300">{msg}</p>}
      </div>
    </div>
  );
}
