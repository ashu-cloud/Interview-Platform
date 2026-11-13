"use client";

import React from "react";

// Vertical architectural typography matching Screen 01 and Screen 04
export function VerticalArchitecturalType({
  lines = ["REAL", "INTERVIEWS", "CLEAR", "DECISIONS"],
  className = "",
}: {
  lines?: string[];
  className?: string;
}) {
  return (
    <div
      className={`flex flex-col justify-between font-mono font-bold tracking-[0.22em] text-sumi select-none uppercase ${className}`}
      aria-hidden="true"
    >
      {lines.map((line, idx) => (
        <span
          key={idx}
          className="text-xs sm:text-sm lg:text-[13px] leading-tight text-sumi/90 hover:text-sumi transition-colors"
        >
          {line}
        </span>
      ))}
    </div>
  );
}

// Precision Architectural Corner Registration Crosshair
export function RegistrationMark({
  coord = "LOC 01 // REF 48k",
  className = "",
}: {
  coord?: string;
  className?: string;
}) {
  return (
    <div className={`inline-flex items-center gap-1.5 font-mono text-[9px] text-sumi-muted tracking-widest ${className}`}>
      <span className="text-sumi font-bold text-[11px] leading-none">+</span>
      <span>{coord}</span>
    </div>
  );
}

// Architectural Dimension Annotation Line
export function DimensionLine({
  label = "1440.0 MM",
  className = "w-full",
}: {
  label?: string;
  className?: string;
}) {
  return (
    <div className={`flex items-center gap-2 font-mono text-[9px] text-sumi-muted select-none ${className}`}>
      <span className="text-[10px] leading-none text-sumi-muted">|</span>
      <div className="h-px flex-1 bg-hairline relative">
        <span className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-paper px-2 text-[8.5px] uppercase tracking-wider text-sumi-secondary">
          {label}
        </span>
      </div>
      <span className="text-[10px] leading-none text-sumi-muted">|</span>
    </div>
  );
}
