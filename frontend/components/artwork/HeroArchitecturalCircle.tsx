"use client";

import React, { useRef, useState } from "react";
import { motion, useMotionValue, useScroll, useSpring, useTransform } from "motion/react";

interface HeroArchitecturalCircleProps {
  className?: string;
  size?: number;
}

const CALIBRATION_DETENTS = [0, 32, -24, 68, -48];

export default function HeroArchitecturalCircle({
  className = "",
  size = 420,
}: HeroArchitecturalCircleProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [detentIndex, setDetentIndex] = useState(0);

  const { scrollYProgress } = useScroll();
  const smooth = useSpring(scrollYProgress, { stiffness: 200, damping: 32 });
  const outerRingRotate = useTransform(smooth, [0, 0.35, 1], [0, 28, 72]);
  const orbitalCounterRotate = useTransform(smooth, [0, 0.35, 1], [0, -18, -44]);
  const vectorScrollRotate = useTransform(smooth, [0, 0.3], [0, 26]);

  // Interactive pointer 3D parallax & azimuth nudge (0 React re-renders per mouse move)
  const ptrX = useMotionValue(0);
  const ptrY = useMotionValue(0);
  const tiltY = useSpring(useTransform(ptrX, [-0.5, 0.5], [-6, 6]), { stiffness: 320, damping: 28 });
  const tiltX = useSpring(useTransform(ptrY, [-0.5, 0.5], [6, -6]), { stiffness: 320, damping: 28 });
  const manualAngle = useSpring(CALIBRATION_DETENTS[detentIndex], { stiffness: 280, damping: 24 });
  const combinedSweep = useTransform(
    [vectorScrollRotate, manualAngle, ptrX],
    ([s, m, px]) => (s as number) + (m as number) + (px as number) * 28
  );

  const activeTheta = (45 + CALIBRATION_DETENTS[detentIndex] + 360) % 360;
  const cosineValue = Math.abs(Math.cos((activeTheta * Math.PI) / 180)).toFixed(3);

  // SVG coordinates: viewBox 0 0 420 420, center at (210, 210), radius R = 150
  const cx = 210;
  const cy = 210;
  const r = 150;
  const r2 = (n: number) => Math.round(n * 100) / 100;

  return (
    <motion.div
      ref={containerRef}
      data-magnetic="true"
      onPointerMove={(e) => {
        if (e.pointerType !== "mouse") return;
        const rect = containerRef.current?.getBoundingClientRect();
        if (!rect) return;
        ptrX.set((e.clientX - rect.left) / rect.width - 0.5);
        ptrY.set((e.clientY - rect.top) / rect.height - 0.5);
      }}
      onPointerLeave={() => {
        ptrX.set(0);
        ptrY.set(0);
      }}
      onClick={() => setDetentIndex((idx) => (idx + 1) % CALIBRATION_DETENTS.length)}
      style={{ perspective: 800, rotateX: tiltX, rotateY: tiltY }}
      whileTap={{ scale: 0.97 }}
      title="Click or sweep cursor to calibrate cosine azimuth angle"
      className={`relative select-none pointer-events-auto cursor-crosshair ${className}`}
      role="button"
      tabIndex={0}
      aria-label={`Interactive Cosine Calibration Instrument (theta ${activeTheta} degrees, cos ${cosineValue})`}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          setDetentIndex((idx) => (idx + 1) % CALIBRATION_DETENTS.length);
        }
      }}
    >
      <svg
        viewBox="0 0 420 420"
        width={size}
        height={size}
        className="w-full h-auto overflow-visible"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          {/* Dense architectural crosshatch pattern matching Screen 01 / Screen 04 */}
          <pattern
            id="drafting-crosshatch-dense"
            width="6"
            height="6"
            patternUnits="userSpaceOnUse"
          >
            <path
              d="M 0 3 L 6 3 M 3 0 L 3 6"
              stroke="#141413"
              strokeWidth="0.85"
              strokeOpacity="0.88"
            />
          </pattern>

          {/* Stipple / fine dot pattern for subtle quadrant texture */}
          <pattern
            id="drafting-stipple-dots"
            width="8"
            height="8"
            patternUnits="userSpaceOnUse"
          >
            <circle cx="4" cy="4" r="0.75" fill="#141413" fillOpacity="0.4" />
          </pattern>

          {/* Diagonal fine hatching */}
          <pattern
            id="drafting-diagonal-hatch"
            width="8"
            height="8"
            patternUnits="userSpaceOnUse"
            patternTransform="rotate(45)"
          >
            <line x1="0" y1="0" x2="0" y2="8" stroke="#141413" strokeWidth="0.75" strokeOpacity="0.6" />
          </pattern>

          {/* Mask for top-right quadrant */}
          <clipPath id="top-right-quadrant-clip">
            <rect x={cx} y={cy - r} width={r} height={r} />
          </clipPath>

          {/* Mask for bottom-left quadrant / arc slice */}
          <clipPath id="bottom-left-quadrant-clip">
            <rect x={cx - r} y={cy} width={r} height={r} />
          </clipPath>

          {/* Mask for main circle */}
          <clipPath id="main-circle-clip">
            <circle cx={cx} cy={cy} r={r} />
          </clipPath>
        </defs>

        {/* --- Background Drafting Coordinate Lines & Tangent Arcs --- */}
        {/* Full-bleed horizontal drafting guide line */}
        <line
          x1="-40"
          y1={cy}
          x2="460"
          y2={cy}
          stroke="#E5E2DA"
          strokeWidth="1"
          strokeDasharray="4 4"
        />
        {/* Full-bleed vertical drafting guide line */}
        <line
          x1={cx}
          y1="-40"
          x2={cx}
          y2="460"
          stroke="#E5E2DA"
          strokeWidth="1"
          strokeDasharray="4 4"
        />

        {/* Large intersecting offset circle (creates elegant lens intersection on the left) */}
        <circle
          cx={cx - 100}
          cy={cy + 70}
          r={r * 1.25}
          stroke="#DDD8CE"
          strokeWidth="0.8"
          strokeDasharray="3 3"
        />

        {/* Secondary intersecting orbital arc on the right (counter-rotates on scroll) */}
        <motion.g
          initial={{ transformOrigin: "210px 210px" }}
          style={{ rotate: orbitalCounterRotate, transformBox: "view-box" }}
        >
          <circle
            cx={cx + 90}
            cy={cy - 60}
            r={r * 1.15}
            stroke="#E5E2DA"
            strokeWidth="0.75"
          />

          {/* Concentric Outer Graduation Rings */}
          <circle
            cx={cx}
            cy={cy}
            r={r + 34}
            stroke="#DDD8CE"
            strokeWidth="0.75"
            strokeDasharray="2 4"
          />
        </motion.g>

        <motion.g
          initial={{ transformOrigin: "210px 210px" }}
          style={{ rotate: outerRingRotate, transformBox: "view-box" }}
        >
          <circle
            cx={cx}
            cy={cy}
            r={r + 20}
            stroke="#141413"
            strokeWidth="0.8"
            strokeOpacity="0.35"
          />
          <circle
            cx={cx}
            cy={cy}
            r={r + 10}
            stroke="#E5E2DA"
            strokeWidth="0.6"
          />

          {/* Outer Calibrated Ticks around (r + 20) */}
          {Array.from({ length: 48 }).map((_, i) => {
            const angle = (i * 360) / 48;
            const rad = (angle * Math.PI) / 180;
            const isMajor = i % 6 === 0;
            const rInner = isMajor ? r + 13 : r + 16;
            const rOuter = r + 20;
            const x1 = r2(cx + Math.cos(rad) * rInner);
            const y1 = r2(cy + Math.sin(rad) * rInner);
            const x2 = r2(cx + Math.cos(rad) * rOuter);
            const y2 = r2(cy + Math.sin(rad) * rOuter);

            return (
              <line
                key={i}
                x1={x1}
                y1={y1}
                x2={x2}
                y2={y2}
                stroke="#141413"
                strokeWidth={isMajor ? "1.25" : "0.6"}
                strokeOpacity={isMajor ? "0.85" : "0.4"}
              />
            );
          })}
        </motion.g>

        {/* Cardinal Degree & Formula Annotations */}
        <text
          x={cx + r + 26}
          y={cy + 3}
          fill="#141413"
          fontSize="8"
          fontFamily="var(--font-jetbrains-mono), monospace"
          letterSpacing="0.1em"
        >
          0° / X
        </text>
        <text
          x={cx - 10}
          y={cy - r - 26}
          fill="#141413"
          fontSize="8"
          fontFamily="var(--font-jetbrains-mono), monospace"
          letterSpacing="0.1em"
        >
          90° / Y
        </text>
        <text
          x={cx - r - 48}
          y={cy + 3}
          fill="#141413"
          fontSize="8"
          fontFamily="var(--font-jetbrains-mono), monospace"
          letterSpacing="0.1em"
        >
          180°
        </text>
        <text
          x={cx - 18}
          y={cy + r + 30}
          fill="#141413"
          fontSize="8"
          fontFamily="var(--font-jetbrains-mono), monospace"
          letterSpacing="0.1em"
        >
          270° / τ
        </text>

        {/* --- MAIN CIRCLE INTERNAL COMPOSITION --- */}
        <g clipPath="url(#main-circle-clip)">
          {/* Base Paper Fill */}
          <circle cx={cx} cy={cy} r={r} fill="#FBF9F5" />

          {/* QUADRANT 1 (Top-Right): DENSE ARCHITECTURAL CROSSHATCH (Signature motif) */}
          <rect
            x={cx}
            y={cy - r}
            width={r}
            height={r}
            fill="url(#drafting-crosshatch-dense)"
          />

          {/* QUADRANT 3 (Bottom-Left): RICH ULTRAMARINE BLUE WEDGE (#1D4ED8) */}
          <path
            d={`M ${cx} ${cy} L ${cx - r} ${cy} A ${r} ${r} 0 0 0 ${cx} ${cy + r} Z`}
            fill="#1D4ED8"
          />

          {/* Sub-slice with fine stipple inside bottom-left for depth */}
          <path
            d={`M ${cx} ${cy} L ${cx - r * 0.7} ${cy} A ${r * 0.7} ${r * 0.7} 0 0 0 ${cx} ${cy + r * 0.7} Z`}
            fill="#172554"
            fillOpacity="0.45"
          />

          {/* QUADRANT 2 (Top-Left): CONCENTRIC CONTOUR ARCS & DIAGONAL TEXTURE */}
          <rect
            x={cx - r}
            y={cy - r}
            width={r}
            height={r}
            fill="url(#drafting-diagonal-hatch)"
          />
          <circle cx={cx} cy={cy} r={r * 0.75} stroke="#141413" strokeWidth="0.8" strokeOpacity="0.35" fill="none" />
          <circle cx={cx} cy={cy} r={r * 0.5} stroke="#141413" strokeWidth="0.8" strokeOpacity="0.45" fill="none" />
          <circle cx={cx} cy={cy} r={r * 0.25} stroke="#141413" strokeWidth="0.8" strokeOpacity="0.35" fill="none" />

          {/* QUADRANT 4 (Bottom-Right): CLEAN ARCHIVAL LINEN WITH INTERACTIVE ANGLE RAY & RULER */}
          <motion.g
            initial={{ transformOrigin: "210px 210px" }}
            style={{ rotate: combinedSweep, transformBox: "view-box" }}
          >
            <line
              x1={cx}
              y1={cy}
              x2={r2(cx + Math.cos(Math.PI / 4) * r)}
              y2={r2(cy + Math.sin(Math.PI / 4) * r)}
              stroke="#1D4ED8"
              strokeWidth="1.5"
            />
            <circle
              cx={r2(cx + Math.cos(Math.PI / 4) * (r * 0.65))}
              cy={r2(cy + Math.sin(Math.PI / 4) * (r * 0.65))}
              r="3.5"
              fill="#1D4ED8"
            />
          </motion.g>
          <text
            x={cx + 36}
            y={cy + 52}
            fill="#57534E"
            fontSize="7"
            fontFamily="var(--font-jetbrains-mono), monospace"
          >
            cos θ = {cosineValue}
          </text>
          <text
            x={cx + 36}
            y={cy + 64}
            fill="#A8A29E"
            fontSize="6.5"
            fontFamily="var(--font-jetbrains-mono), monospace"
          >
            CLICK // {activeTheta}° AZIMUTH
          </text>

          {/* Primary Axis Hairlines Inside Circle */}
          <line x1={cx - r} y1={cy} x2={cx + r} y2={cy} stroke="#141413" strokeWidth="1.2" />
          <line x1={cx} y1={cy - r} x2={cx} y2={cy + r} stroke="#141413" strokeWidth="1.2" />

          {/* 45-degree Diagonal Lines */}
          <line
            x1={cx - r * 0.707}
            y1={cy - r * 0.707}
            x2={cx + r * 0.707}
            y2={cy + r * 0.707}
            stroke="#141413"
            strokeWidth="0.75"
            strokeDasharray="3 3"
            strokeOpacity="0.5"
          />
          <line
            x1={cx - r * 0.707}
            y1={cy + r * 0.707}
            x2={cx + r * 0.707}
            y2={cy - r * 0.707}
            stroke="#141413"
            strokeWidth="0.75"
            strokeDasharray="3 3"
            strokeOpacity="0.5"
          />
        </g>

        {/* Main Outer Rim Border */}
        <circle
          cx={cx}
          cy={cy}
          r={r}
          stroke="#141413"
          strokeWidth="1.75"
          fill="none"
        />

        {/* Registration Marks and Crosshair Accents */}
        {/* Center Pivot Point */}
        <circle cx={cx} cy={cy} r="3" fill="#141413" />
        <circle cx={cx} cy={cy} r="6" stroke="#FBF9F5" strokeWidth="1.2" fill="none" />
        <circle cx={cx} cy={cy} r="8" stroke="#141413" strokeWidth="0.8" fill="none" />

        {/* Precision Registration Crosshairs at Edge Intersections */}
        {/* Top */}
        <line x1={cx - 7} y1={cy - r} x2={cx + 7} y2={cy - r} stroke="#141413" strokeWidth="1.5" />
        <line x1={cx} y1={cy - r - 7} x2={cx} y2={cy - r + 7} stroke="#141413" strokeWidth="1.5" />

        {/* Bottom */}
        <line x1={cx - 7} y1={cy + r} x2={cx + 7} y2={cy + r} stroke="#141413" strokeWidth="1.5" />
        <line x1={cx} y1={cy + r - 7} x2={cx} y2={cy + r + 7} stroke="#141413" strokeWidth="1.5" />

        {/* Left */}
        <line x1={cx - r - 7} y1={cy} x2={cx - r + 7} y2={cy} stroke="#141413" strokeWidth="1.5" />
        <line x1={cx - r} y1={cy - 7} x2={cx - r} y2={cy + 7} stroke="#141413" strokeWidth="1.5" />

        {/* Right */}
        <line x1={cx + r - 7} y1={cy} x2={cx + r + 7} y2={cy} stroke="#141413" strokeWidth="1.5" />
        <line x1={cx + r} y1={cy - 7} x2={cx + r} y2={cy + 7} stroke="#141413" strokeWidth="1.5" />

        {/* Tangent Annotation Box / Dimension Leader */}
        <path
          d={`M ${cx + r * 0.707} ${cy - r * 0.707} L ${cx + r + 30} ${cy - r * 0.707} L ${cx + r + 50} ${cy - r * 0.707 - 14}`}
          stroke="#141413"
          strokeWidth="0.8"
          fill="none"
        />
        <text
          x={cx + r + 54}
          y={cy - r * 0.707 - 12}
          fill="#141413"
          fontSize="7.5"
          fontFamily="var(--font-jetbrains-mono), monospace"
          fontWeight="600"
        >
          CADENCE // 12ms
        </text>
        <text
          x={cx + r + 54}
          y={cy - r * 0.707 - 3}
          fill="#57534E"
          fontSize="6.5"
          fontFamily="var(--font-jetbrains-mono), monospace"
        >
          WSS DUPLEX 48kHz
        </text>
      </svg>
    </motion.div>
  );
}
