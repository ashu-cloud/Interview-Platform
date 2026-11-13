"use client";

import React from "react";
import Logo from "@/components/ui/Logo";

export default function Footer() {
  return (
    <footer className="bg-paper hairline-t py-16 text-xs font-mono text-sumi-secondary" data-purpose="site-footer">
      <div className="max-w-7xl 2xl:max-w-[1440px] mx-auto px-6 lg:px-12">
        {/* Upper Colophon Block */}
        <div className="flex flex-col lg:flex-row items-start lg:items-center justify-between gap-8 pb-10 hairline-b">
          <div className="flex items-center gap-5">
            <Logo className="w-12 h-12 text-sumi shrink-0" size={48} />
            <div>
              <div className="flex flex-wrap items-center gap-3 text-sumi font-bold mb-1">
                <span className="folio-stamp bg-white">COLOPHON</span>
                <span className="text-sm tracking-wider">
                  ASSESSMENT ENGINE // ARCHIVAL CONSOLE v4.2
                </span>
              </div>
              <p className="text-[11px] text-sumi-muted font-sans leading-relaxed">
                Autonomous Technical Voice &amp; Evaluation Platform &bull; Portfolio Work by Senior AI &amp; Systems Engineer
              </p>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-4 sm:gap-6 text-[11px]">
            <span className="inline-flex items-center gap-2 text-emerald-700 font-semibold bg-emerald-50 px-2.5 py-1 rounded border border-emerald-200">
              <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse" />
              LATENCY DRIFT: ±0.8 MS
            </span>
            <span className="text-hairline">/</span>
            <span className="text-sumi-muted font-medium">ATTESTATION: SUMI-V3 / CALIBRATED</span>
          </div>
        </div>

        {/* Lower Links & Copyright */}
        <div className="mt-10 flex flex-col md:flex-row items-start md:items-center justify-between gap-6 text-[11px] text-sumi-muted">
          <div>
            &copy; 2025 Assessment Engine. Engineered with Swiss-grid precision and Japanese Sumi carbon aesthetic.
          </div>

          <nav className="flex flex-wrap items-center gap-5 sm:gap-7 text-xs font-mono uppercase tracking-wider">
            <a className="hover:text-sumi transition-colors" href="#pipeline">
              Streaming Voice
            </a>
            <a className="hover:text-sumi transition-colors" href="#coverage">
              Adaptive Coverage
            </a>
            <a className="hover:text-sumi transition-colors" href="#speculative-rag">
              Predictive RAG
            </a>
            <a className="hover:text-sumi transition-colors" href="#scatter-gather">
              Scatter-Gather
            </a>
            <a className="hover:text-sumi transition-colors" href="#architecture">
              Resilience Grid
            </a>
            <a className="hover:text-sumi transition-colors" href="#showcase">
              Interface Studio
            </a>
          </nav>
        </div>
      </div>
    </footer>
  );
}
