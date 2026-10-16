import React from "react";

export default function FinalCTA() {
  return (
    <section className="py-28 lg:py-36 bg-white hairline-b" data-purpose="final-call-to-action">
      <div className="max-w-7xl mx-auto px-6 lg:px-12">
        <div className="p-10 sm:p-16 bg-paper-card border border-hairline rounded-sm flex flex-col md:flex-row items-center justify-between gap-10 crosshair-corner shadow-plate">
          <div className="max-w-2xl">
            <div className="font-mono text-xs text-sumi-muted uppercase tracking-widest mb-3">
              DEPLOYMENT EVALUATION &amp; BENCHMARK
            </div>
            <h2 className="text-3xl sm:text-4xl font-light text-sumi tracking-tight mb-4">
              Experience the Voice Critical Path Live
            </h2>
            <p className="text-sm text-sumi-secondary leading-relaxed font-sans">
              Launch an interactive evaluation session or inspect the architectural whitepaper and code repository. Engineered as a demonstrative benchmark for senior AI &amp; systems engineering capabilities.
            </p>
          </div>
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-4 w-full md:w-auto">
            <a
              className="px-7 py-4 bg-sumi text-paper font-mono text-xs uppercase tracking-widest font-semibold rounded hover:bg-neutral-800 transition-colors text-center"
              href="#pipeline-demo"
            >
              Launch Voice Stage Demo
            </a>
            <a
              className="px-7 py-4 bg-white border border-hairline text-sumi font-mono text-xs uppercase tracking-widest rounded text-center hover:border-sumi transition-colors"
              href="https://github.com/ashu-cloud/Interview-Platform"
              rel="noopener noreferrer"
              target="_blank"
            >
              Inspect Repository →
            </a>
          </div>
        </div>
      </div>
    </section>
  );
}
