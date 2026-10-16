import React from "react";
import Logo from "@/components/ui/Logo";

export default function Footer() {
  return (
    <footer className="bg-paper hairline-t py-16 text-xs font-mono text-sumi-secondary" data-purpose="site-footer">
      <div className="max-w-7xl mx-auto px-6 lg:px-12">
        <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-8 pb-10 hairline-b">
          <div className="flex items-center gap-4">
            <Logo className="w-10 h-10 text-sumi" size={40} />
            <div>
              <div className="flex items-center gap-3 text-sumi font-bold mb-1">
                <span className="folio-stamp">COLOPHON</span>
                <span className="text-sm tracking-wider">
                  ASSESSMENT ENGINE // ARCHIVAL CONSOLE v4.2
                </span>
              </div>
              <p className="text-[11px] text-sumi-muted font-sans">
                Autonomous Technical Voice &amp; Evaluation Platform • Portfolio Work by Senior AI &amp; Systems Engineer
              </p>
            </div>
          </div>
          <div className="flex items-center gap-6 text-[11px]">
            <span className="inline-flex items-center gap-2 text-emerald-600">
              <span className="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
              LATENCY DRIFT: ±0.8 MS
            </span>
            <span className="text-hairline">/</span>
            <span className="text-sumi-muted">ATTESTATION: SUMI-V3 / CALIBRATED</span>
          </div>
        </div>
        <div className="mt-10 flex flex-col sm:flex-row items-center justify-between gap-6 text-[11px] text-sumi-muted">
          <div>
            © 2025 Assessment Engine. Engineered with Swiss-grid precision and Sumi carbon aesthetic.
          </div>
          <div className="flex items-center gap-6">
            <a className="hover:text-sumi transition-colors" href="#pipeline">
              Voice Pipeline
            </a>
            <a className="hover:text-sumi transition-colors" href="#coverage">
              Adaptive Coverage
            </a>
            <a className="hover:text-sumi transition-colors" href="#scatter-gather">
              Scatter-Gather Worker
            </a>
            <a className="hover:text-sumi transition-colors" href="#architecture">
              Architecture Spec
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}
