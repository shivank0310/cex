import Link from "next/link";

export function Footer() {
  return (
    <footer className="border-t border-white/10 bg-[#080810] py-12">
      <div className="mx-auto max-w-7xl px-4 sm:px-6">
        <div className="grid gap-8 md:grid-cols-4">
          <div>
            <p className="text-lg font-bold text-white">CEX Platform</p>
            <p className="mt-2 text-sm text-slate-400">
              Centralized exchange with off-chain matching, on-chain custody, and enterprise-grade architecture.
            </p>
          </div>
          <div>
            <p className="font-semibold text-white">Products</p>
            <ul className="mt-3 space-y-2 text-sm text-slate-400">
              <li><Link href="/trade" className="hover:text-cyan-400">Spot Trading</Link></li>
              <li><Link href="/markets" className="hover:text-cyan-400">Market Data</Link></li>
              <li><Link href="/wallet" className="hover:text-cyan-400">Crypto Wallet</Link></li>
            </ul>
          </div>
          <div>
            <p className="font-semibold text-white">Architecture</p>
            <ul className="mt-3 space-y-2 text-sm text-slate-400">
              <li>Matching Engine (off-chain)</li>
              <li>Ledger & Settlement</li>
              <li>Blockchain Service</li>
            </ul>
          </div>
          <div>
            <p className="font-semibold text-white">Stack</p>
            <ul className="mt-3 space-y-2 text-sm text-slate-400">
              <li>Go Microservices</li>
              <li>Kafka + Redis</li>
              <li>Next.js Frontend</li>
            </ul>
          </div>
        </div>
        <p className="mt-10 text-center text-xs text-slate-500">
          © {new Date().getFullYear()} CEX. Off-chain matching · On-chain deposits & withdrawals.
        </p>
      </div>
    </footer>
  );
}
