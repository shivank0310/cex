import { Hero } from "@/components/landing/Hero";
import { Features } from "@/components/landing/Features";
import { Architecture } from "@/components/landing/Architecture";
import Link from "next/link";
import { Button } from "@/components/ui/Button";

export default function HomePage() {
  return (
    <>
      <Hero />
      <Features />
      <Architecture />
      <section className="px-4 py-20 text-center sm:px-6">
        <div className="mx-auto max-w-2xl rounded-3xl border border-violet-500/30 bg-gradient-to-br from-violet-950/50 to-cyan-950/30 p-12">
          <h2 className="text-2xl font-bold text-white">Ready to Launch Your Exchange?</h2>
          <p className="mt-4 text-slate-400">
            Explore the live trading terminal connected to your Go microservices backend.
          </p>
          <Link href="/trade" className="mt-8 inline-block">
            <Button size="lg">Open Trading Terminal</Button>
          </Link>
        </div>
      </section>
    </>
  );
}
