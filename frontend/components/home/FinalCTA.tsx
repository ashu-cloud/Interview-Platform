import React from "react";
import Link from "next/link";

export default function FinalCTA() {
  return (
    <section className="py-28 lg:py-36 bg-white hairline-b" data-purpose="final-call-to-action">
      <div className="max-w-7xl mx-auto px-6 lg:px-12">
        <div className="p-10 sm:p-16 bg-paper-card border border-hairline rounded-sm flex flex-col md:flex-row items-center justify-between gap-10 crosshair-corner shadow-plate">
          <div className="max-w-2xl">
            <div className="font-mono text-xs text-sumi-muted uppercase tracking-widest mb-3">
              PRODUCTION VOICE PLATFORM // DEPLOYED LIVE
            </div>
            <h2 className="text-3xl sm:text-4xl font-light text-sumi tracking-tight mb-4">
              Experience Autonomous Technical Assessment
            </h2>
            <p className="text-sm text-sumi-secondary leading-relaxed font-sans">
              Sign in to author multi-dimensional rubrics in the Hiring Manager Console, inspect verbatim-grounded evaluation reports, or test the sub-800ms full-duplex conversational audio stage.
            </p>
          </div>
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-4 w-full md:w-auto">
            <Link
              className="px-7 py-4 bg-sumi text-paper font-mono text-xs uppercase tracking-widest font-semibold rounded hover:bg-neutral-800 transition-colors text-center shadow-xs"
              href="/auth"
            >
              Launch Hiring Console →
            </Link>
            <a
              className="px-7 py-4 bg-white border border-hairline text-sumi font-mono text-xs uppercase tracking-widest rounded text-center hover:border-sumi transition-colors"
              href="#pipeline-demo"
            >
              Interactive Telemetry
            </a>
            <a
              className="px-7 py-4 bg-white border border-hairline text-sumi font-mono text-xs uppercase tracking-widest rounded text-center hover:border-sumi transition-colors"
              href="https://github.com/ashu-cloud/Interview-Platform"
              rel="noopener noreferrer"
              target="_blank"
            >
              GitHub Repo →
            </a>
          </div>
        </div>
      </div>
    </section>
  );
}
