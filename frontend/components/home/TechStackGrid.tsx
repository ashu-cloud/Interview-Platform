import React from "react";

export default function TechStackGrid() {
  const stack = [
    {
      num: "01 / RUNTIME & ENGINE",
      title: "Go 1.25 Runtime Engine",
      desc: "Gorilla WebSockets and sync.Pool memory buffer recycling handling streaming audio frames with sub-millisecond dispatch latency.",
      bullets: ["Gin HTTP REST APIs", "gorilla/websocket WSS Streaming", "sync.Pool Zero-Alloc Buffers"],
      tag: "ZERO-ALLOC POOLS",
    },
    {
      num: "02 / STORAGE & VECTOR",
      title: "PostgreSQL 16 + pgvector",
      desc: "ACID transactional guarantees for session state and turn embeddings indexed via HNSW cosine distance.",
      bullets: ["Upstash Serverless Redis (TLS) / Redis 7.2", "Cloudflare R2 / MinIO Audio Lake", "SQL Row-Level Security (RLS)"],
      tag: "HNSW COSINE",
    },
    {
      num: "03 / STREAMING & BROKER",
      title: "Dual-Mode Event Architecture",
      desc: "Zero-cost transactional outbox for single-node deployments with seamless failover to Apache Kafka KRaft clusters.",
      bullets: ["Event: session.turns", "Event: session.completed", "Postgres Outbox (SKIP LOCKED) Fallback"],
      tag: "ACID OUTBOX",
    },
    {
      num: "04 / NEURAL CRITICAL PATH",
      title: "Multi-Model Neural Pipeline",
      desc: "Multi-vendor latency optimization selecting the lowest-TTFB provider for each discrete stage with prompt grounding.",
      bullets: ["Deepgram Nova-2 (16kHz Realtime STT)", "Groq Llama 3.3 70B & GPT-OSS 120B", "ElevenLabs Turbo v2 (Chunked TTS)"],
      tag: "SUB-800MS TTFB",
    },
  ];

  return (
    <section className="py-28 lg:py-40 hairline-b" data-purpose="tech-stack-matrix">
      <div className="max-w-7xl mx-auto px-6 lg:px-12">
        <div className="max-w-2xl mb-16 pb-6 hairline-b">
          <div className="flex items-center gap-3 font-mono text-xs text-sumi-muted uppercase tracking-widest mb-3">
            <span className="folio-stamp">PL. 06 / FOLIO VI</span>
            <span>SYSTEM ARCHITECTURE &amp; COMPONENTS</span>
          </div>
          <h2 className="text-3xl sm:text-5xl font-light text-sumi tracking-tight">
            Production Technology Stack Grid
          </h2>
          <p className="text-sm text-sumi-secondary mt-3 font-mono leading-relaxed">
            Selected for deterministic sub-millisecond execution, memory safety, and relational integrity.
          </p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
          {stack.map((item, idx) => (
            <div
              key={idx}
              className="p-8 bg-white border border-hairline rounded flex flex-col justify-between hover:border-sumi transition-all crosshair-corner interactive-plate group"
            >
              <div>
                <div className="font-mono text-xs text-sumi-muted uppercase tracking-widest mb-4 flex justify-between items-center">
                  <span>{item.num}</span>
                  <span className="text-[9px] font-bold bg-paper-card px-2 py-0.5 rounded text-sumi-secondary border border-hairline group-hover:bg-sumi group-hover:text-white transition-colors">
                    {item.tag}
                  </span>
                </div>
                <div className="text-lg font-semibold text-sumi mb-2 group-hover:translate-x-0.5 transition-transform">
                  {item.title}
                </div>
                <p className="text-xs text-sumi-secondary leading-relaxed mb-6 font-sans">
                  {item.desc}
                </p>
              </div>
              <div className="font-mono text-[11px] text-sumi-muted space-y-1.5 pt-4 hairline-t">
                {item.bullets.map((b, i) => (
                  <div key={i} className="flex items-center gap-1.5 hover:text-sumi transition-colors">
                    <span className="text-emerald-600">•</span>
                    <span>{b}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
