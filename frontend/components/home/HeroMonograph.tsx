"use client";

import React, { useRef } from "react";
import Link from "next/link";
import { motion, useScroll, useSpring, useTransform } from "motion/react";
import HeroVoiceTelemetry from "./HeroVoiceTelemetry";
import HeroArchitecturalCircle from "@/components/artwork/HeroArchitecturalCircle";
import { VerticalArchitecturalType } from "@/components/artwork/DraftingAccents";
import {
  BlurReveal,
  HeroRackFocusStage,
  ScrubParallaxLayer,
} from "@/components/motion/ScrollChoreography";

export default function HeroMonograph() {
  const sectionRef = useRef<HTMLElement>(null);
  const { scrollYProgress } = useScroll({
    target: sectionRef,
    offset: ["start start", "end start"],
  });
  const smooth = useSpring(scrollYProgress, { stiffness: 220, damping: 32 });
  const leftCircleY = useTransform(smooth, [0, 1], [0, -70]);
  const leftCircleRotate = useTransform(smooth, [0, 1], [0, 38]);

  return (
    <HeroRackFocusStage>
      <section
        id="hero"
        ref={sectionRef}
        className="relative pt-10 pb-20 lg:pt-14 lg:pb-28 hairline-b overflow-x-clip bg-[#FAF8F5]"
        data-purpose="hero-monograph"
      >
      {/* Background Architectural Drafting Guide Lines */}
      <div className="pointer-events-none absolute inset-0 z-0">
        <svg className="w-full h-full stroke-[#141413]/[0.05]" fill="none">
          <defs>
            <pattern id="hero-draft-grid" width="80" height="80" patternUnits="userSpaceOnUse">
              <path d="M 80 0 L 0 0 0 80" strokeWidth="0.75" />
            </pattern>
          </defs>
          <rect width="100%" height="100%" fill="url(#hero-draft-grid)" />
        </svg>

        {/* Global architectural hairline guides */}
        <div className="absolute top-14 left-0 right-0 h-px bg-[#141413]/[0.08]" />
        <div className="absolute top-0 bottom-0 left-[22%] w-px bg-[#141413]/[0.05] hidden xl:block" />
        <div className="absolute top-0 bottom-0 right-[25%] w-px bg-[#141413]/[0.05] hidden xl:block" />
      </div>

      {/* Cropped Half Circle on Far Left Edge (scrubs & rotates on scroll) */}
      <motion.div
        style={{ y: leftCircleY, rotate: leftCircleRotate, willChange: "transform" }}
        className="pointer-events-none absolute left-0 top-1/4 -translate-y-1/2 -translate-x-[55%] w-80 h-80 z-0 hidden lg:block select-none"
        aria-hidden="true"
      >
        <svg viewBox="0 0 320 320" className="w-full h-full overflow-visible">
          <defs>
            <pattern id="hero-left-hatch" width="6" height="6" patternUnits="userSpaceOnUse">
              <path d="M 0 3 L 6 3 M 3 0 L 3 6" stroke="#141413" strokeWidth="0.85" strokeOpacity="0.85" />
            </pattern>
          </defs>
          {/* Main circle rim */}
          <circle cx="160" cy="160" r="140" stroke="#141413" strokeWidth="1.5" fill="#FAF8F5" />
          {/* Upper right quadrant: dense crosshatch */}
          <path d="M 160 160 L 160 20 A 140 140 0 0 1 300 160 Z" fill="url(#hero-left-hatch)" />
          {/* Lower right quadrant: cobalt blue */}
          <path d="M 160 160 L 300 160 A 140 140 0 0 1 160 300 Z" fill="#1D4ED8" />
          {/* Outer guide ring & crosshair axes */}
          <circle cx="160" cy="160" r="152" stroke="#E5E2DA" strokeWidth="0.8" strokeDasharray="3 3" fill="none" />
          <line x1="20" y1="160" x2="300" y2="160" stroke="#141413" strokeWidth="1" strokeDasharray="4 4" opacity="0.35" />
          <line x1="160" y1="20" x2="160" y2="300" stroke="#141413" strokeWidth="1" strokeDasharray="4 4" opacity="0.35" />
        </svg>
      </motion.div>

      <div className="relative z-10 max-w-7xl 2xl:max-w-[1440px] mx-auto px-6 lg:px-12">
        {/* Archival Folio Stamp & Spec Status Rail */}
        <div className="flex flex-wrap items-center justify-between gap-4 font-mono text-xs text-sumi-muted pb-4 mb-8 hairline-b">
          <div className="flex flex-wrap items-center gap-3">
            <span className="folio-stamp flex items-center gap-1.5 bg-white shadow-xs px-2.5 py-1 border border-hairline rounded-xs">
              <span className="font-bold text-sumi">01</span>
              <span className="text-hairline">/</span>
              <span>PL. 01 / FOLIO I</span>
            </span>
            <span className="flex items-center gap-2 text-sumi font-semibold tracking-wider text-[11px]">
              <span className="w-1.5 h-1.5 rounded-full bg-[#1D4ED8]" />
              <span>ORCHESTRATION CADENCE</span>
            </span>
            <span className="text-hairline">—</span>
            <span className="tracking-wide uppercase text-[11px] text-sumi-secondary">
              AUTONOMOUS SYSTEM ENGINEERING EVALUATION
            </span>
          </div>

          <div className="flex items-center gap-3 text-[11px]">
            <span className="px-2.5 py-1 rounded bg-white border border-hairline text-sumi-secondary">
              DETERMINISTIC RUBRIC TOPOLOGY
            </span>
            <span className="px-2.5 py-1 rounded bg-emerald-50 border border-emerald-200 text-emerald-800 font-medium">
              VERIFIED TRANSCRIPT ATTRIBUTION
            </span>
          </div>
        </div>

        {/* Hero Headline & Editorial Context */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-10 items-start mb-12">
          {/* Main Display Headline (7 cols) */}
          <div className="lg:col-span-7 space-y-5">
            <BlurReveal immediate delay={0.02} yOffset={10} blurPx={4}>
              <div className="font-mono text-[10.5px] uppercase tracking-widest text-sumi-muted flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-[#1D4ED8]" />
                <span>SYSTEM SPECIFICATION // RFC-6455 STREAMING WSS</span>
              </div>
            </BlurReveal>

            <BlurReveal immediate delay={0.08} yOffset={16} blurPx={6}>
              <h1 className="text-4xl sm:text-5xl xl:text-[4rem] font-light tracking-tight text-sumi leading-[1.05] font-sans">
                Technical Voice Interviews Grounded in{" "}
                <span className="font-serif italic font-normal text-sumi">Deterministic Truth</span>.
              </h1>
            </BlurReveal>

            <BlurReveal immediate delay={0.14} yOffset={14} blurPx={5}>
              <p className="text-base sm:text-lg text-sumi-secondary font-normal leading-relaxed max-w-xl">
                Sub-800ms bidirectional conversational voice powered by sentence-boundary streaming,
                mathematical cosine centroid coverage tracking, and tamper-proof verbatim substring-verified transcript attribution.
              </p>
            </BlurReveal>

            {/* Target Assessment Domains */}
            <div className="pt-2 flex flex-wrap items-center gap-2 font-mono text-[11px] text-sumi-muted">
              <span className="text-sumi font-semibold">TARGET DOMAINS:</span>
              <span className="px-2 py-0.5 rounded bg-white border border-hairline text-sumi">
                Senior Infrastructure
              </span>
              <span className="px-2 py-0.5 rounded bg-white border border-hairline text-sumi">
                Distributed Systems
              </span>
              <span className="px-2 py-0.5 rounded bg-white border border-hairline text-sumi">
                AI Platform Engineering
              </span>
            </div>

            {/* Call-to-Action Controls */}
            <div className="pt-3 flex flex-wrap items-center gap-4">
              <Link
                className="py-3 px-6 bg-sumi text-paper font-mono text-xs uppercase tracking-widest font-semibold rounded-xs hover:bg-neutral-800 transition-all shadow-xs flex items-center gap-2"
                href="/auth"
              >
                <span>Launch Hiring Console</span>
                <span>→</span>
              </Link>
              <a
                className="py-3 px-6 border border-hairline bg-white hover:border-sumi font-mono text-xs uppercase tracking-widest rounded-xs transition-colors text-sumi shadow-xs"
                href="#pipeline"
              >
                Live Telemetry
              </a>
              <div className="flex items-center gap-2 text-xs font-mono text-sumi-muted pl-1">
                <span>Candidate token?</span>
                <Link href="/auth" className="text-sumi font-semibold underline hover:text-ultramarine transition-colors">
                  Enter Stage Access →
                </Link>
              </div>
            </div>
          </div>

          {/* Right Column: Signature Geometric Artwork & Mandate Box (5 cols) */}
          <ScrubParallaxLayer yRange={[0, -28]} className="lg:col-span-5 flex flex-col gap-6">
            {/* Signature Circular Drafting Diagram matching Screen 01 */}
            <BlurReveal immediate delay={0.1} yOffset={16} blurPx={6}>
              <div className="relative flex items-center justify-between gap-4">
                <div className="w-[280px] sm:w-[320px]">
                  <HeroArchitecturalCircle size={320} />
                </div>
                <VerticalArchitecturalType
                  lines={["REAL", "INTERVIEWS", "CLEAR", "DECISIONS"]}
                  className="shrink-0 h-[220px]"
                />
              </div>
            </BlurReveal>

            {/* Mandate Monograph Box matching Screen 01 */}
            <BlurReveal immediate delay={0.18} yOffset={14} blurPx={5}>
              <div className="bg-[#FAF8F5] border border-hairline p-5 rounded-xs relative shadow-xs">
                <p className="font-mono text-[10px] sm:text-[10.5px] text-sumi-secondary leading-relaxed uppercase tracking-wider">
                  ENGINEERED FOR SENIOR INFRASTRUCTURE, DISTRIBUTED SYSTEMS, AND AI PLATFORM ASSESSMENT WHERE QUALITATIVE SMALL-TALK IS REPLACED WITH HIGH-LOW LATENCY CROSS-EXAMINATION, AND MATHEMATICAL VERIFICATION.
                </p>
                <div className="mt-4 pt-3 border-t border-hairline flex items-center justify-between font-mono text-[9.5px] text-sumi-muted tracking-widest">
                  <span>AUTHOR: SENIOR SYSTEMS //</span>
                  <span className="font-bold text-sumi">PORTFOLIO NOW</span>
                </div>
              </div>
            </BlurReveal>
          </ScrubParallaxLayer>
        </div>

        {/* Live Audio Telemetry Instrument Showcase */}
        <div className="mt-6">
          <HeroVoiceTelemetry />
        </div>

        {/* Protocol Attestation Bar */}
        <div className="mt-6 flex flex-wrap items-center justify-between text-xs font-mono text-sumi-muted gap-4">
          <div className="flex items-center gap-2">
            <span className="text-sumi font-bold">FIG. 01</span>
            <span>//</span>
            <span>REAL-TIME DUPLEX AUDIO CONSOLE &bull; BIDIRECTIONAL WSS TELEMETRY</span>
          </div>
          <div className="flex items-center gap-2 text-sumi text-[11px]">
            <span className="w-1.5 h-1.5 bg-emerald-500 rounded-full animate-pulse" />
            <span>SYNCHRONOUS PROTOCOL: RFC-6455 STREAMING WSS</span>
            <span className="text-hairline">|</span>
            <span className="text-emerald-700 font-semibold">FAST AUDIO PATH &lt;800MS</span>
          </div>
        </div>
      </div>
      </section>
    </HeroRackFocusStage>
  );
}
