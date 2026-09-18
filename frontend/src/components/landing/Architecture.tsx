"use client";

import { motion } from "framer-motion";

const FLOW = [
  { step: "Order", desc: "User places limit/market order", color: "from-violet-500 to-purple-600" },
  { step: "Matching", desc: "Off-chain engine (price-time FIFO)", color: "from-purple-500 to-fuchsia-600" },
  { step: "Trade", desc: "Kafka event published", color: "from-fuchsia-500 to-pink-600" },
  { step: "Settlement", desc: "Ledger journal posted", color: "from-pink-500 to-rose-600" },
  { step: "Balance", desc: "User available balance updated", color: "from-rose-500 to-orange-500" },
];

const ONCHAIN = [
  "Deposits → CEXVault → wallet-service → ledger",
  "Withdrawals → ledger reserve → blockchain-service",
  "Treasury fee sweeps on-chain",
];

export function Architecture() {
  return (
    <section className="px-4 py-20 sm:px-6 bg-gradient-to-b from-transparent to-violet-950/20">
      <div className="mx-auto max-w-7xl">
        <h2 className="text-3xl font-bold text-white text-center">Off-Chain vs On-Chain</h2>
        <p className="mt-4 text-center text-slate-400">
          Orders never hit a smart contract. Blockchain is for custody only.
        </p>

        <div className="mt-12 flex flex-wrap justify-center gap-3">
          {FLOW.map((item, i) => (
            <motion.div
              key={item.step}
              initial={{ opacity: 0, x: -10 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.1 }}
              className="flex items-center gap-3"
            >
              <div className={`rounded-xl bg-gradient-to-r ${item.color} px-4 py-3 text-center min-w-[120px]`}>
                <p className="font-bold text-white text-sm">{item.step}</p>
                <p className="text-xs text-white/80 mt-0.5">{item.desc}</p>
              </div>
              {i < FLOW.length - 1 && <span className="text-slate-600 hidden sm:inline">→</span>}
            </motion.div>
          ))}
        </div>

        <div className="mt-14 grid gap-6 md:grid-cols-2">
          <div className="rounded-2xl border border-emerald-500/20 bg-emerald-500/5 p-6">
            <h3 className="font-semibold text-emerald-300">Off-Chain (Go Services)</h3>
            <ul className="mt-4 space-y-2 text-sm text-slate-300">
              <li>• Order placement & validation</li>
              <li>• In-memory matching engine</li>
              <li>• Double-entry ledger</li>
              <li>• Trade settlement</li>
              <li>• Real-time market data</li>
            </ul>
          </div>
          <div className="rounded-2xl border border-cyan-500/20 bg-cyan-500/5 p-6">
            <h3 className="font-semibold text-cyan-300">On-Chain (Solidity)</h3>
            <ul className="mt-4 space-y-2 text-sm text-slate-300">
              {ONCHAIN.map((line) => (
                <li key={line}>• {line}</li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    </section>
  );
}
