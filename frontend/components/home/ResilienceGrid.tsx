import React from "react";

export default function ResilienceGrid() {
  const cards = [
    {
      track: "// TRACK 01",
      title: "Dual-Track Critical Path",
      desc: "In-memory Upstash / Redis 7.2 state and Redlock distributed mutexes handle the audio loop in under 5ms, decoupling conversational speed entirely from the heavier analytical evaluation engine.",
      label: "STORAGE:",
      val: "Upstash / Redis 7.2 (TLS)",
      badge: "LATENCY < 5MS",
    },
    {
      track: "// TRACK 02",
      title: "Dual-Mode Event Delivery",
      desc: "Dispatches to PostgreSQL Transactional Outbox (with FOR UPDATE SKIP LOCKED batching for 0-cost single-node) or Apache Kafka KRaft topics (session.turns, session.completed) for distributed clusters.",
      label: "BROKER:",
      val: "Postgres Outbox / Kafka",
      badge: "ZERO-DATA-LOSS",
    },
    {
      track: "// TRACK 03",
      title: "Concurrency & Session Locks",
      desc: "Strict mutex architecture prevents 'ghost tab' audio packet collisions, duplicate submissions from dropped connections, and dual-evaluator race conditions during evaluation synthesis.",
      label: "LOCKING:",
      val: "Redlock Mutex",
      badge: "IDEMPOTENT WSS",
    },
    {
      track: "// TRACK 04",
      title: "Multi-Tenant SQL Isolation",
      desc: "PostgreSQL Row-Level Security (RLS) ensures absolute enterprise data separation. One-way bcrypt password hashing and 128-bit cryptographically secure invite tokens.",
      label: "SECURITY:",
      val: "Postgres RLS + AES-256",
      badge: "TENANT ISOLATED",
    },
  ];

  return (
    <section className="py-28 lg:py-40 hairline-b" data-purpose="resilience-architecture" id="architecture">
      <div className="max-w-7xl mx-auto px-6 lg:px-12">
        <div className="flex flex-col md:flex-row md:items-end justify-between mb-16 pb-6 hairline-b">
          <div>
            <div className="flex items-center gap-3 font-mono text-xs text-sumi-muted uppercase tracking-widest mb-3">
              <span className="folio-stamp">PL. 04 / FOLIO IV</span>
              <span>DISTRIBUTED RESILIENCE</span>
            </div>
            <h2 className="text-3xl sm:text-5xl font-light text-sumi tracking-tight">
              Enterprise Infrastructure &amp; Concurrency Substrates
            </h2>
          </div>
          <p className="text-xs font-mono text-sumi-muted mt-4 md:mt-0 uppercase tracking-widest">
            SYSTEM DISCIPLINE: STRICT SWISS GRID • ZERO DATA LOSS GUARANTEE
          </p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
          {cards.map((card, i) => (
            <div
              key={i}
              className="p-8 bg-white border border-hairline rounded flex flex-col justify-between hover:border-sumi transition-all crosshair-corner interactive-plate group"
            >
              <div>
                <div className="font-mono text-xs text-sumi-muted mb-4 flex justify-between items-center">
                  <span>{card.track}</span>
                  <span className="text-[9px] bg-paper-card px-2 py-0.5 rounded text-sumi-secondary font-semibold border border-hairline group-hover:bg-sumi group-hover:text-white transition-colors">
                    {card.badge}
                  </span>
                </div>
                <h3 className="text-lg font-semibold text-sumi mb-3 font-sans group-hover:translate-x-0.5 transition-transform">
                  {card.title}
                </h3>
                <p className="text-xs text-sumi-secondary leading-relaxed font-sans">
                  {card.desc}
                </p>
              </div>
              <div className="mt-8 pt-4 hairline-t font-mono text-[11px] text-sumi flex justify-between items-center">
                <span className="text-sumi-muted">{card.label}</span>
                <span className="font-semibold text-emerald-700">{card.val}</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
