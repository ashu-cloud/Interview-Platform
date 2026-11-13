"use client";

import React from "react";
import Link from "next/link";
import Logo from "@/components/ui/Logo";
import {
  BlurReveal,
  SectionRackFocusStage,
} from "@/components/motion/ScrollChoreography";

export default function FinalCTA() {
  return (
    <SectionRackFocusStage maxBlur={10}>
      <section id="cta" className="py-24 lg:py-32 bg-paper-card hairline-b" data-purpose="final-call-to-action">
        <div className="max-w-7xl 2xl:max-w-[1440px] mx-auto px-6 lg:px-12">
          <BlurReveal blurPx={9} yOffset={24}>
            <div className="p-10 sm:p-16 bg-white border border-hairline rounded-sm flex flex-col lg:flex-row items-start lg:items-center justify-between gap-10 crosshair-corner shadow-plate">
          <div className="max-w-2xl space-y-4">
            <div className="flex items-center gap-2 font-mono text-xs text-sumi-muted uppercase tracking-widest">
              <Logo className="w-4 h-4 text-sumi" size={16} />
              <span>PRODUCTION VOICE PLATFORM // DEPLOYED LIVE</span>
            </div>
            <h2 className="text-3xl sm:text-4xl font-light text-sumi tracking-tight">
              Experience Autonomous Technical Assessment
            </h2>
            <p className="text-sm text-sumi-secondary leading-relaxed font-sans">
              Sign in to author multi-dimensional rubrics in the Hiring Manager Console, inspect verbatim-grounded evaluation reports, or test the sub-800ms full-duplex conversational audio stage.
            </p>
          </div>

          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 w-full lg:w-auto font-mono text-xs">
            <Link
              className="px-6 py-4 bg-sumi text-paper font-mono uppercase tracking-widest font-semibold rounded-xs hover:bg-neutral-800 transition-colors text-center shadow-xs flex items-center justify-center gap-2"
              href="/auth"
            >
              <span>Launch Hiring Console</span>
              <span>→</span>
            </Link>
            <a
              className="px-6 py-4 bg-white border border-hairline text-sumi font-mono uppercase tracking-widest rounded-xs text-center hover:border-sumi transition-colors shadow-xs"
              href="#pipeline"
            >
              Interactive Pipeline
            </a>
            <a
              className="px-6 py-4 bg-white border border-hairline text-sumi font-mono uppercase tracking-widest rounded-xs text-center hover:border-sumi transition-colors shadow-xs"
              href="https://github.com/ashu-cloud/Interview-Platform"
              rel="noopener noreferrer"
              target="_blank"
            >
              GitHub Repo ↗
            </a>
          </div>
        </div>
          </BlurReveal>
        </div>
      </section>
    </SectionRackFocusStage>
  );
}
