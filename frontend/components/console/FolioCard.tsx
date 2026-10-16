"use client";

import React, { useState } from "react";

export interface WaveformBar {
  height: number; // percentage 15 - 95
  type?: "blue" | "dark" | "muted";
}

interface FolioCardProps {
  folioNumber: string;
  code: string;
  activeCount: number;
  title: string;
  description: string;
  badges: string[];
  rubricId: string;
  waveformBars?: WaveformBar[];
  waveformPattern?: number[];
}

export default function FolioCard({
  folioNumber,
  code,
  activeCount,
  title,
  description,
  badges,
  rubricId,
  waveformBars,
  waveformPattern,
}: FolioCardProps) {
  const [isHovered, setIsHovered] = useState(false);
  const [isExplored, setIsExplored] = useState(false);

  // Normalize bars
  const bars: WaveformBar[] = waveformBars || (
    waveformPattern
      ? waveformPattern.slice(0, 8).map((h, i) => ({
          height: h,
          type: i % 3 === 0 ? "blue" : i % 2 === 1 ? "dark" : "muted",
        }))
      : [
          { height: 35, type: "muted" },
          { height: 60, type: "muted" },
          { height: 95, type: "blue" },
          { height: 70, type: "dark" },
          { height: 50, type: "dark" },
          { height: 95, type: "blue" },
          { height: 65, type: "muted" },
          { height: 40, type: "muted" },
        ]
  );

  return (
    <div
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      className="bg-[#f5f3ef] p-6 rounded border border-[#e5e2da] shadow-plate flex flex-col justify-between hover:bg-[#efeeea] hover:border-sumi transition-all group duration-300"
    >
      <div>
        {/* Folio Header */}
        <div className="flex items-center justify-between pb-3.5 border-b border-[#e5e2da]">
          <span className="font-mono text-[11px] font-bold text-sumi-muted tracking-widest uppercase">
            {folioNumber} / {code}
          </span>
          <span className="font-mono text-xs text-ultramarine font-semibold tracking-wider">
            {activeCount} ACTIVE
          </span>
        </div>

        {/* Slender Soundwave Needle Silhouette (Picture 2 aesthetic) */}
        <div className="h-24 my-4 flex items-center justify-center gap-3 sm:gap-3.5 px-4 bg-white/70 rounded-xs border border-[#e5e2da]/80 overflow-hidden">
          {bars.map((bar, idx) => {
            const barColor =
              bar.type === "blue"
                ? "bg-ultramarine"
                : bar.type === "dark"
                ? "bg-sumi"
                : "bg-[#d5d2c9]";

            const dynamicHeight = isHovered
              ? Math.min(95, Math.max(18, bar.height * (1 + (idx % 2 === 0 ? 0.22 : -0.18))))
              : bar.height;

            return (
              <div
                key={idx}
                className={`w-[2.5px] rounded-full transition-all duration-300 ease-out ${barColor}`}
                style={{ height: `${dynamicHeight}%` }}
              ></div>
            );
          })}
        </div>

        {/* Title & Technical Spec */}
        <h3 className="font-sans text-lg font-semibold text-sumi tracking-tight group-hover:text-black transition-colors">
          {title}
        </h3>
        <p className="font-sans text-xs text-sumi-secondary mt-2 leading-relaxed font-normal">
          {description}
        </p>

        {/* Architectural Badges */}
        <div className="flex flex-wrap gap-1.5 mt-4">
          {badges.map((badge, idx) => (
            <span
              key={idx}
              className="bg-[#fbf9f5] px-2 py-0.5 rounded font-mono text-[11px] text-sumi-secondary border border-[#e5e2da]"
            >
              {badge}
            </span>
          ))}
        </div>
      </div>

      {/* Footer Rubric & Explore Action */}
      <div className="pt-4 mt-5 border-t border-[#e5e2da] flex items-center justify-between">
        <span className="font-mono text-[11px] text-sumi-muted tracking-wider">{rubricId}</span>
        <button
          onClick={() => setIsExplored(!isExplored)}
          className="font-mono text-xs text-sumi hover:text-ultramarine uppercase font-medium transition-colors flex items-center gap-1.5 group-hover:translate-x-0.5 transition-transform"
        >
          <span>{isExplored ? "Active Track" : "Explore Track"}</span>
          <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M5 12h14M12 5l7 7-7 7" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        </button>
      </div>

      {isExplored && (
        <div className="mt-3 p-3 bg-white rounded border border-ultramarine text-xs font-mono text-sumi animate-in fade-in">
          <div className="font-semibold text-ultramarine flex items-center gap-1.5 mb-1">
            <span className="w-1.5 h-1.5 rounded-full bg-ultramarine animate-ping"></span>
            <span>Rubric Matrix Calibrated</span>
          </div>
          <p className="text-[11px] text-sumi-secondary">
            Deterministic AST grading criteria deployed to WebSocket cluster for {title}.
          </p>
        </div>
      )}
    </div>
  );
}
