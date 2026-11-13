"use client";

import React, { useEffect, useRef, useState } from "react";
import {
  MotionConfig,
  motion,
  useInView,
  useMotionTemplate,
  useMotionValue,
  useScroll,
  useSpring,
  useTransform,
  useVelocity,
} from "motion/react";

/** Shared architectural easing curves (exponential deceleration, zero bounce) */
export const ARCHITECTURAL_EASE: [number, number, number, number] = [0.22, 1, 0.36, 1];
export const MOVE_EASE: [number, number, number, number] = [0.25, 1, 0.5, 1];
/** Apple critically-damped hardware spring */
export const APPLE_SPRING = { type: "spring" as const, stiffness: 460, damping: 36, mass: 0.8 };
export const APPLE_GENTLE_SPRING = { type: "spring" as const, stiffness: 280, damping: 30 };

interface EditorialMotionProviderProps {
  children: React.ReactNode;
}

export function EditorialMotionProvider({ children }: EditorialMotionProviderProps) {
  return <MotionConfig reducedMotion="user">{children}</MotionConfig>;
}

interface HeroRackFocusStageProps {
  children: React.ReactNode;
  className?: string;
}

/**
 * Hero section wrapper: immediately crisp at y=0 (no hidden above-the-fold state),
 * then smoothly racks out of focus, scales slightly back in Z-depth, and fades
 * as the reader scrolls down into Table 01.
 */
export function HeroRackFocusStage({ children, className = "" }: HeroRackFocusStageProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start start", "end start"],
  });

  const smooth = useSpring(scrollYProgress, {
    stiffness: 260,
    damping: 32,
    restDelta: 0.001,
  });

  const filter = useTransform(smooth, [0, 0.58, 0.98], [
    "blur(0px)",
    "blur(0px)",
    "blur(11px)",
  ]);
  const opacity = useTransform(smooth, [0, 0.58, 0.98], [1, 1, 0.28]);
  const scale = useTransform(smooth, [0, 0.58, 1], [1, 1, 0.968]);
  const y = useTransform(smooth, [0, 1], [0, -36]);

  return (
    <div ref={containerRef} className={`relative overflow-x-clip bg-[#FAF8F5] ${className}`}>
      <motion.div
        style={{ filter, opacity, scale, y, willChange: "transform, filter, opacity" }}
        className="origin-top"
      >
        {children}
      </motion.div>
    </div>
  );
}

interface SectionRackFocusStageProps {
  children: React.ReactNode;
  className?: string;
  /** Optional blur intensity in px at seam boundaries (capped at <= 16px per gpu budget) */
  maxBlur?: number;
}

/**
 * Scrubbed Section-to-Section Rack-Focus Transition Wrapper.
 * Continuously couples scroll position to optical focus (blur), Z-scale, Y-travel,
 * and opacity so moving from one monograph plate to the next creates an editorial
 * camera lens rack-focus seam while staying 100% sharp in the reading sweet spot.
 */
export function SectionRackFocusStage({
  children,
  className = "",
  maxBlur = 10,
}: SectionRackFocusStageProps) {
  const ref = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ["start end", "end start"],
  });

  const smooth = useSpring(scrollYProgress, {
    stiffness: 260,
    damping: 32,
    restDelta: 0.001,
  });

  const filter = useTransform(
    smooth,
    [0, 0.05, 0.13, 0.87, 0.95, 1],
    [
      `blur(${maxBlur}px)`,
      `blur(${Math.round(maxBlur * 0.45)}px)`,
      "blur(0px)",
      "blur(0px)",
      `blur(${Math.round(maxBlur * 0.5)}px)`,
      `blur(${maxBlur}px)`,
    ]
  );

  const opacity = useTransform(smooth, [0, 0.06, 0.14, 0.86, 0.95, 1], [0.26, 0.78, 1, 1, 0.72, 0.24]);
  const scale = useTransform(smooth, [0, 0.14, 0.86, 1], [0.968, 1, 1, 0.972]);
  const y = useTransform(smooth, [0, 0.14, 0.86, 1], [30, 0, 0, -24]);

  return (
    <div ref={ref} className={`relative overflow-x-clip bg-[#FBF9F5] ${className}`}>
      <motion.div
        style={{ filter, opacity, scale, y, willChange: "transform, filter, opacity" }}
        className="origin-center"
      >
        {children}
      </motion.div>
    </div>
  );
}

interface DraftingSeamRuleProps {
  label?: string;
  folioRef?: string;
}

/**
 * Architectural seam divider between sections whose reticle cursor, beam fill,
 * and degree readout scrub smoothly as the seam crosses the viewport.
 */
export function DraftingSeamRule({
  label = "CALIBRATION SEAM",
  folioRef = "ΔZ // FOCUS SHIFT",
}: DraftingSeamRuleProps) {
  const ref = useRef<HTMLDivElement>(null);
  const { scrollY } = useScroll();
  const scrollVelocity = useVelocity(scrollY);
  const smoothVelocity = useSpring(scrollVelocity, { stiffness: 240, damping: 28 });

  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ["start end", "end start"],
  });

  const smooth = useSpring(scrollYProgress, { stiffness: 280, damping: 34 });
  const scaleX = useTransform(smooth, [0.1, 0.65], [0.08, 1]);
  const vernierX = useTransform(smooth, [0, 1], ["6%", "92%"]);
  const markerRotate = useTransform(smooth, [0, 1], [0, 270]);
  const velocityScaleX = useTransform(smoothVelocity, [-2400, 0, 2400], [1.9, 1, 1.9]);
  const velocitySkewX = useTransform(smoothVelocity, [-2200, 0, 2200], [10, 0, -10]);

  return (
    <div
      ref={ref}
      aria-hidden="true"
      className="relative w-full h-10 bg-[#FBF9F5] border-y border-hairline/70 flex items-center overflow-hidden select-none"
    >
      <div className="max-w-7xl 2xl:max-w-[1440px] w-full mx-auto px-6 lg:px-12 relative flex items-center justify-between font-mono text-[9.5px] tracking-widest text-sumi-muted">
        <span className="bg-[#FBF9F5] pr-3 z-10 flex items-center gap-2">
          <motion.span
            style={{ rotate: markerRotate }}
            className="inline-block text-sumi font-bold"
          >
            +
          </motion.span>
          <span>{label}</span>
        </span>

        {/* Scrubbed + velocity-reactive central drafting rule */}
        <div className="absolute left-6 right-6 lg:left-12 lg:right-12 top-1/2 -translate-y-1/2 h-px bg-hairline">
          <motion.div
            style={{ scaleX, transformOrigin: "left center" }}
            className="h-full w-full bg-[#1D4ED8]/55"
          />
          <motion.div
            style={{ left: vernierX, scaleX: velocityScaleX, skewX: velocitySkewX }}
            className="absolute -top-1.5 w-3 h-3 -translate-x-1/2 flex items-center justify-center"
          >
            <span className="w-2.5 h-2.5 border border-[#1D4ED8] bg-[#FBF9F5] rotate-45 block shadow-[0_0_8px_rgba(29,78,216,0.35)]" />
          </motion.div>
        </div>

        <span className="bg-[#FBF9F5] pl-3 z-10 flex items-center gap-2">
          <span>{folioRef}</span>
          <motion.span
            style={{ rotate: markerRotate }}
            className="inline-block text-sumi font-bold"
          >
            +
          </motion.span>
        </span>
      </div>
    </div>
  );
}

interface ChapterBridgeProps {
  title: string;
  folioTag?: string;
}

/**
 * Intelligence-Factory-inspired centered italic serif chapter interstitial
 * paired with a scroll-scrubbed vertical ink stem guiding eye flow into the next plate.
 */
export function ChapterBridge({ title, folioTag }: ChapterBridgeProps) {
  const ref = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ["start 92%", "center 45%"],
  });
  const smooth = useSpring(scrollYProgress, { stiffness: 240, damping: 30 });
  const blur = useTransform(smooth, [0, 0.45, 1], ["blur(5px)", "blur(0px)", "blur(0px)"]);
  const opacity = useTransform(smooth, [0, 0.35], [0.35, 1]);

  return (
    <div
      ref={ref}
      aria-hidden="true"
      className="relative w-full py-10 bg-[#FBF9F5] flex flex-col items-center justify-center gap-2.5 select-none overflow-hidden"
    >
      {folioTag && (
        <span className="font-mono text-[9px] uppercase tracking-[0.26em] text-sumi-muted">
          {folioTag}
        </span>
      )}
      <motion.div
        style={{ filter: blur, opacity }}
        className="font-serif italic text-xl sm:text-2xl text-sumi/85 tracking-wide"
      >
        {title}
      </motion.div>
      <svg width="16" height="36" viewBox="0 0 16 36" fill="none">
        <line x1="8" y1="0" x2="8" y2="32" stroke="#E5E2DA" strokeWidth="1" />
        <motion.path
          d="M 8 0 L 8 32"
          stroke="#141413"
          strokeWidth="1.2"
          style={{ pathLength: smooth }}
        />
        <motion.path
          d="M 4 27 L 8 32 L 12 27"
          stroke="#141413"
          strokeWidth="1.2"
          strokeLinecap="round"
          strokeLinejoin="round"
          style={{ pathLength: smooth }}
        />
      </svg>
    </div>
  );
}

interface SketchyCardFrameProps {
  children: React.ReactNode;
  className?: string;
}

/**
 * Hand-drafted organic rounded card outline (Intelligence Factory Frames 4-7 aesthetic)
 * drawn with SVG pathLength on viewport entry around workflow cards.
 */
export function SketchyCardFrame({ children, className = "" }: SketchyCardFrameProps) {
  const ref = useRef<HTMLDivElement>(null);
  const isInView = useInView(ref, { once: true, amount: "some" });

  return (
    <div ref={ref} className={`relative ${className}`}>
      <svg
        aria-hidden="true"
        viewBox="0 0 1000 400"
        preserveAspectRatio="none"
        className="pointer-events-none absolute inset-0 w-full h-full overflow-visible"
        fill="none"
      >
        <motion.path
          d="M 28 3 C 260 1, 740 4, 972 2 C 990 3, 998 14, 998 32 C 999 130, 997 270, 998 368 C 997 388, 988 397, 970 397 C 720 399, 280 396, 30 398 C 12 397, 3 387, 2 368 C 1 270, 3 130, 2 32 C 3 13, 12 4, 28 3 Z"
          stroke="#141413"
          strokeOpacity="0.45"
          strokeWidth="1.2"
          vectorEffect="non-scaling-stroke"
          initial={{ pathLength: 0 }}
          animate={{ pathLength: isInView ? 1 : 0 }}
          transition={{ duration: 0.9, ease: ARCHITECTURAL_EASE }}
        />
      </svg>
      <div className="relative z-10 h-full flex flex-col">{children}</div>
    </div>
  );
}

interface BlurRevealProps {
  children: React.ReactNode;
  className?: string;
  delay?: number;
  yOffset?: number;
  xOffset?: number;
  blurPx?: number;
  /** If true, animates immediately on mount (for above-the-fold hero typography) */
  immediate?: boolean;
}

/**
 * Blur-before-move editorial entrance: filter: blur() resolves slightly ahead
 * of spatial settle so typography glides into optical sharpness.
 */
export function BlurReveal({
  children,
  className = "",
  delay = 0,
  yOffset = 22,
  xOffset = 0,
  blurPx = 8,
  immediate = false,
}: BlurRevealProps) {
  const initialState = {
    opacity: 0,
    y: yOffset,
    x: xOffset,
    filter: `blur(${blurPx}px)`,
  };
  const visibleState = {
    opacity: 1,
    y: 0,
    x: 0,
    filter: "blur(0px)",
  };

  if (immediate) {
    return (
      <motion.div
        initial={initialState}
        animate={visibleState}
        transition={{
          duration: 0.62,
          delay,
          ease: ARCHITECTURAL_EASE,
          filter: { duration: 0.44, delay, ease: ARCHITECTURAL_EASE },
        }}
        className={className}
      >
        {children}
      </motion.div>
    );
  }

  return (
    <motion.div
      initial={initialState}
      whileInView={visibleState}
      viewport={{ once: true, amount: "some" }}
      transition={{
        duration: 0.65,
        delay,
        ease: ARCHITECTURAL_EASE,
        filter: { duration: 0.46, delay, ease: ARCHITECTURAL_EASE },
      }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

interface StaggerWaveProps {
  children: React.ReactNode;
  className?: string;
  staggerDelay?: number;
}

export function StaggerWave({
  children,
  className = "",
  staggerDelay = 0.08,
}: StaggerWaveProps) {
  const waveRef = useRef<HTMLDivElement>(null);
  const isInView = useInView(waveRef, { once: true, amount: "some" });

  return (
    <motion.div
      ref={waveRef}
      initial="hidden"
      animate={isInView ? "visible" : "hidden"}
      variants={{
        hidden: {},
        visible: {
          transition: {
            staggerChildren: staggerDelay,
          },
        },
      }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

interface StaggerItemProps {
  children: React.ReactNode;
  className?: string;
  yOffset?: number;
  xOffset?: number;
  blurPx?: number;
}

export function StaggerItem({
  children,
  className = "",
  yOffset = 20,
  xOffset = 0,
  blurPx = 7,
}: StaggerItemProps) {
  return (
    <motion.div
      variants={{
        hidden: {
          opacity: 0,
          y: yOffset,
          x: xOffset,
          filter: `blur(${blurPx}px)`,
        },
        visible: {
          opacity: 1,
          y: 0,
          x: 0,
          filter: "blur(0px)",
          transition: {
            duration: 0.58,
            ease: ARCHITECTURAL_EASE,
            filter: { duration: 0.4, ease: ARCHITECTURAL_EASE },
          },
        },
      }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

interface ScrubParallaxLayerProps {
  children: React.ReactNode;
  className?: string;
  yRange?: [number, number];
  rotateRange?: [number, number];
  scaleRange?: [number, number];
}

/**
 * Scrubbed parallax layer anchored to its container's scroll traversal.
 */
export function ScrubParallaxLayer({
  children,
  className = "",
  yRange = [24, -24],
  rotateRange = [0, 0],
  scaleRange = [1, 1],
}: ScrubParallaxLayerProps) {
  const ref = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ["start end", "end start"],
  });
  const smooth = useSpring(scrollYProgress, { stiffness: 220, damping: 30 });
  const y = useTransform(smooth, [0, 1], yRange);
  const rotate = useTransform(smooth, [0, 1], rotateRange);
  const scale = useTransform(smooth, [0, 1], scaleRange);

  return (
    <div ref={ref} className={className}>
      <motion.div style={{ y, rotate, scale, willChange: "transform" }}>
        {children}
      </motion.div>
    </div>
  );
}

const FOLIO_STOPS = [
  { id: "hero", num: "01", label: "MONOGRAPH" },
  { id: "benchmarks", num: "02", label: "BENCHMARKS" },
  { id: "coverage", num: "03", label: "COVERAGE" },
  { id: "pipeline", num: "04", label: "PIPELINE" },
  { id: "architecture", num: "05", label: "RESILIENCE" },
  { id: "showcase", num: "06", label: "INTERFACE" },
  { id: "scatter-gather", num: "07", label: "MATRIX" },
  { id: "cta", num: "08", label: "DEPLOY" },
];

/**
 * Carrier Element unifying the 8-section homepage monograph into a single
 * optical instrument: renders a scrubbed top progress hairline + floating desktop
 * vernier folio index on the right margin.
 */
export function ScrollFolioRail() {
  const { scrollYProgress } = useScroll();
  const smoothProgress = useSpring(scrollYProgress, {
    stiffness: 260,
    damping: 34,
    restDelta: 0.001,
  });

  const [percent, setPercent] = useState(0);
  const [activeId, setActiveId] = useState("hero");

  useEffect(() => {
    const unsubscribe = smoothProgress.on("change", (latest) => {
      setPercent(Math.round(latest * 100));
    });
    return () => unsubscribe();
  }, [smoothProgress]);

  useEffect(() => {
    const elements = FOLIO_STOPS.map((stop) => document.getElementById(stop.id)).filter(
      (el): el is HTMLElement => Boolean(el)
    );
    if (elements.length === 0) return;

    const observer = new IntersectionObserver(
      (entries) => {
        // Pick the visible section nearest the upper-middle of the viewport
        const visible = entries
          .filter((e) => e.isIntersecting)
          .sort((a, b) => Math.abs(a.boundingClientRect.top) - Math.abs(b.boundingClientRect.top));
        if (visible.length > 0) {
          setActiveId(visible[0].target.id);
        }
      },
      { rootMargin: "-25% 0px -45% 0px", threshold: [0, 0.15, 0.4] }
    );

    elements.forEach((el) => observer.observe(el));
    return () => observer.disconnect();
  }, []);

  const handleJump = (id: string) => {
    const el = document.getElementById(id);
    if (el) {
      el.scrollIntoView({ behavior: "smooth", block: "start" });
    }
  };

  return (
    <>
      {/* Continuous Top Calibration Hairline */}
      <div
        aria-hidden="true"
        className="fixed top-0 left-0 right-0 h-[2px] z-50 pointer-events-none bg-transparent"
      >
        <motion.div
          style={{ scaleX: smoothProgress, transformOrigin: "left center" }}
          className="h-full w-full bg-[#1D4ED8]"
        />
      </div>

      {/* Right-Margin Architectural Folio Telemetry Rail (wide desktop only) */}
      <aside
        aria-label="Scroll folio telemetry"
        className="hidden 2xl:flex fixed right-4 top-1/2 -translate-y-1/2 z-40 flex-col items-end gap-2 pointer-events-auto select-none"
      >
        <div className="bg-[#FAF8F5]/90 backdrop-blur-xs border border-hairline rounded-xs px-2.5 py-3 shadow-xs space-y-2.5">
          <div className="font-mono text-[9px] text-sumi-muted tracking-widest border-b border-hairline pb-1.5 flex items-center justify-between gap-3">
            <span>DEPTH</span>
            <span className="font-bold text-sumi tabular-nums">
              {String(percent).padStart(3, "0")}%
            </span>
          </div>

          <nav className="flex flex-col gap-1">
            {FOLIO_STOPS.map((stop) => {
              const isActive = activeId === stop.id;
              return (
                <button
                  key={stop.id}
                  type="button"
                  onClick={() => handleJump(stop.id)}
                  className={`flex items-center justify-end gap-2 font-mono text-[9px] tracking-wider transition-colors cursor-pointer py-0.5 ${
                    isActive
                      ? "text-[#1D4ED8] font-bold"
                      : "text-sumi-muted hover:text-sumi"
                  }`}
                >
                  <span className={isActive ? "opacity-100" : "opacity-65"}>{stop.label}</span>
                  <span className="tabular-nums">{stop.num}</span>
                  <span
                    className={`h-px transition-all ${
                      isActive ? "w-3 bg-[#1D4ED8]" : "w-1.5 bg-hairline"
                    }`}
                  />
                </button>
              );
            })}
          </nav>

          {/* Mini vertical scrubbed bar */}
          <div className="pt-1 border-t border-hairline">
            <div className="w-full h-1 bg-hairline rounded-full overflow-hidden">
              <motion.div
                style={{ scaleX: smoothProgress, transformOrigin: "left center" }}
                className="h-full w-full bg-[#1D4ED8]"
              />
            </div>
          </div>
        </div>
      </aside>
    </>
  );
}

interface MagneticSpotlightPlateProps {
  children: React.ReactNode;
  className?: string;
  tiltStrength?: number;
}

/**
 * Apple / Behance 3D perspective interactive surface with zero-rerender pointer
 * coordinate tracking: drives spring rotateX/rotateY tilt + radial spotlight beam.
 */
export function MagneticSpotlightPlate({
  children,
  className = "",
  tiltStrength = 4.2,
}: MagneticSpotlightPlateProps) {
  const plateRef = useRef<HTMLDivElement>(null);
  const mouseX = useMotionValue(0);
  const mouseY = useMotionValue(0);
  const normX = useMotionValue(0);
  const normY = useMotionValue(0);
  const spotOpacity = useMotionValue(0);

  const smoothRotY = useSpring(
    useTransform(normX, [-0.5, 0.5], [-tiltStrength, tiltStrength]),
    { stiffness: 360, damping: 30 }
  );
  const smoothRotX = useSpring(
    useTransform(normY, [-0.5, 0.5], [tiltStrength, -tiltStrength]),
    { stiffness: 360, damping: 30 }
  );
  const smoothSpotOpacity = useSpring(spotOpacity, { stiffness: 300, damping: 28 });

  const spotlightBg = useMotionTemplate`radial-gradient(320px circle at ${mouseX}px ${mouseY}px, rgba(29, 78, 216, 0.085), transparent 72%)`;

  const handlePointerMove = (e: React.PointerEvent<HTMLDivElement>) => {
    if (e.pointerType !== "mouse") return;
    const rect = plateRef.current?.getBoundingClientRect();
    if (!rect || rect.width === 0 || rect.height === 0) return;
    const lx = e.clientX - rect.left;
    const ly = e.clientY - rect.top;
    mouseX.set(lx);
    mouseY.set(ly);
    normX.set(lx / rect.width - 0.5);
    normY.set(ly / rect.height - 0.5);
    spotOpacity.set(1);
  };

  const handlePointerLeave = () => {
    normX.set(0);
    normY.set(0);
    spotOpacity.set(0);
  };

  return (
    <motion.div
      ref={plateRef}
      data-magnetic="true"
      onPointerMove={handlePointerMove}
      onPointerLeave={handlePointerLeave}
      style={{
        perspective: 950,
        rotateX: smoothRotX,
        rotateY: smoothRotY,
        transformStyle: "preserve-3d",
      }}
      whileHover={{ y: -3 }}
      transition={APPLE_SPRING}
      className={`relative ${className}`}
    >
      <motion.div
        aria-hidden="true"
        className="pointer-events-none absolute inset-0 z-10 transition-none"
        style={{ background: spotlightBg, opacity: smoothSpotOpacity }}
      />
      {children}
    </motion.div>
  );
}

interface ClipCurtainRevealProps {
  children: React.ReactNode;
  className?: string;
  direction?: "up" | "left" | "center";
  delay?: number;
}

/**
 * Behance-style GPU clip-path curtain reveal anchored to an unclipped outer sentinel
 * so IntersectionObserver always observes full bounding height and never locks at 0%.
 */
export function ClipCurtainReveal({
  children,
  className = "",
  direction = "up",
  delay = 0,
}: ClipCurtainRevealProps) {
  const sentinelRef = useRef<HTMLDivElement>(null);
  const isInView = useInView(sentinelRef, { once: true, amount: "some" });

  const hiddenClip =
    direction === "left"
      ? "inset(0% 100% 0% 0%)"
      : direction === "center"
      ? "inset(0% 50% 0% 50%)"
      : "inset(100% 0% 0% 0%)";

  return (
    <div ref={sentinelRef} className={className}>
      <motion.div
        initial={{ clipPath: hiddenClip, opacity: 0.45 }}
        animate={
          isInView
            ? { clipPath: "inset(0% 0% 0% 0%)", opacity: 1 }
            : { clipPath: hiddenClip, opacity: 0.45 }
        }
        transition={{ duration: 0.68, delay, ease: ARCHITECTURAL_EASE }}
      >
        {children}
      </motion.div>
    </div>
  );
}

interface InteractiveSplitCompareProps {
  leftLabel: string;
  leftTitle: string;
  leftItems: { metric: string; note: string }[];
  rightLabel: string;
  rightTitle: string;
  rightItems: { metric: string; note: string }[];
}

/**
 * Behance / Apple Draggable Clip-Path Architecture Comparison Curtain.
 * Users drag the optical divider (or press ArrowLeft/ArrowRight) to wipe between
 * Monolithic LLM execution and VoxAssess Scatter-Gather topology in real time.
 */
export function InteractiveSplitCompare({
  leftLabel,
  leftTitle,
  leftItems,
  rightLabel,
  rightTitle,
  rightItems,
}: InteractiveSplitCompareProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [splitPct, setSplitPct] = useState(52);
  const [isDragging, setIsDragging] = useState(false);

  const smoothPct = useSpring(splitPct, { stiffness: 420, damping: 38 });
  const clipRight = useTransform(smoothPct, (p) => `inset(0 ${100 - p}% 0 0)`);
  const handleLeft = useTransform(smoothPct, (p) => `${p}%`);

  const updateFromClientX = (clientX: number) => {
    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect || rect.width === 0) return;
    const raw = ((clientX - rect.left) / rect.width) * 100;
    setSplitPct(Math.max(12, Math.min(88, Math.round(raw * 10) / 10)));
  };

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center justify-between gap-2 font-mono text-[10px] text-sumi-muted uppercase tracking-wider">
        <span className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-[#1D4ED8]" />
          <span>DRAG OPTICAL CURTAIN TO COMPARE EXECUTION PROFILES</span>
        </span>
        <span className="text-sumi font-semibold tabular-nums">
          SPLIT: {Math.round(splitPct)}% // MESH VS. MONOLITH
        </span>
      </div>

      <div
        ref={containerRef}
        role="slider"
        tabIndex={0}
        aria-label="Drag comparison curtain between Monolithic LLM and VoxAssess Scatter-Gather"
        aria-valuemin={12}
        aria-valuemax={88}
        aria-valuenow={Math.round(splitPct)}
        onKeyDown={(e) => {
          if (e.key === "ArrowLeft") setSplitPct((p) => Math.max(12, p - 4));
          if (e.key === "ArrowRight") setSplitPct((p) => Math.min(88, p + 4));
        }}
        onPointerDown={(e) => {
          setIsDragging(true);
          (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
          updateFromClientX(e.clientX);
        }}
        onPointerMove={(e) => {
          if (isDragging) updateFromClientX(e.clientX);
        }}
        onPointerUp={() => setIsDragging(false)}
        onPointerCancel={() => setIsDragging(false)}
        className="relative h-[230px] sm:h-[210px] border border-hairline rounded-sm overflow-hidden select-none cursor-ew-resize bg-white shadow-xs focus:outline-hidden focus:ring-2 focus:ring-[#1D4ED8]"
      >
        {/* Base Layer (Right Side - Monolithic Blocking Profile) */}
        <div className="absolute inset-0 bg-[#F4F1EA] p-6 sm:p-8 flex flex-col justify-between">
          <div className="flex items-center justify-between font-mono text-[10px] text-neutral-500">
            <span className="uppercase tracking-widest font-bold">{rightLabel}</span>
            <span className="px-2 py-0.5 rounded bg-neutral-200/80 text-neutral-700 font-semibold">
              SERIAL BLOCKING
            </span>
          </div>
          <div className="space-y-3">
            <div className="font-sans text-lg font-medium text-neutral-700">{rightTitle}</div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 font-mono text-xs">
              {rightItems.map((it, idx) => (
                <div key={idx} className="p-2.5 bg-white/75 border border-hairline rounded-xs">
                  <div className="font-bold text-neutral-600">{it.metric}</div>
                  <div className="text-[10.5px] font-sans text-neutral-500 mt-0.5">{it.note}</div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Clipped Overlay Layer (Left Side - VoxAssess Deterministic Mesh) */}
        <motion.div
          style={{ clipPath: clipRight }}
          className="absolute inset-0 bg-[#141413] text-[#FAF8F5] p-6 sm:p-8 flex flex-col justify-between"
        >
          <div className="flex items-center justify-between font-mono text-[10px] text-neutral-400">
            <span className="uppercase tracking-widest font-bold text-emerald-400">{leftLabel}</span>
            <span className="px-2 py-0.5 rounded bg-emerald-500/15 border border-emerald-500/40 text-emerald-300 font-semibold">
              SCATTER-GATHER // VERIFIED
            </span>
          </div>
          <div className="space-y-3">
            <div className="font-sans text-lg font-medium text-white">{leftTitle}</div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 font-mono text-xs">
              {leftItems.map((it, idx) => (
                <div key={idx} className="p-2.5 bg-white/5 border border-white/15 rounded-xs">
                  <div className="font-bold text-emerald-400">{it.metric}</div>
                  <div className="text-[10.5px] font-sans text-neutral-300 mt-0.5">{it.note}</div>
                </div>
              ))}
            </div>
          </div>
        </motion.div>

        {/* Draggable Optical Handle Line */}
        <motion.div
          style={{ left: handleLeft }}
          className="absolute top-0 bottom-0 w-0.5 bg-[#1D4ED8] z-20 -translate-x-1/2 pointer-events-none"
        >
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-8 h-8 rounded-full bg-[#1D4ED8] text-white shadow-md border-2 border-white flex items-center justify-center font-mono text-[10px] font-bold">
            ↔
          </div>
        </motion.div>
      </div>
    </div>
  );
}

/**
 * Behance-caliber desktop optical crosshair cursor follower driven strictly
 * by MotionValues + springs (0 React re-renders per frame; hidden on touch/coarse).
 */
export function DraftingCursorReticle() {
  const cx = useMotionValue(-100);
  const cy = useMotionValue(-100);
  const visible = useMotionValue(0);
  const reticleScale = useMotionValue(1);

  const sx = useSpring(cx, { stiffness: 620, damping: 42 });
  const sy = useSpring(cy, { stiffness: 620, damping: 42 });
  const sOpacity = useSpring(visible, { stiffness: 300, damping: 30 });
  const sScale = useSpring(reticleScale, { stiffness: 420, damping: 30 });

  useEffect(() => {
    if (typeof window === "undefined") return;
    const finePointer = window.matchMedia("(pointer: fine)");
    if (!finePointer.matches) return;

    const onMove = (e: MouseEvent) => {
      cx.set(e.clientX);
      cy.set(e.clientY);
      visible.set(1);

      const target = e.target as HTMLElement | null;
      const isMagnetic = Boolean(
        target?.closest("a, button, [data-magnetic='true'], [role='slider']")
      );
      reticleScale.set(isMagnetic ? 1.65 : 1);
    };

    const onLeave = () => visible.set(0);

    window.addEventListener("mousemove", onMove, { passive: true });
    document.addEventListener("mouseleave", onLeave);
    return () => {
      window.removeEventListener("mousemove", onMove);
      document.removeEventListener("mouseleave", onLeave);
    };
  }, [cx, cy, visible, reticleScale]);

  return (
    <motion.div
      aria-hidden="true"
      style={{ x: sx, y: sy, opacity: sOpacity, scale: sScale }}
      className="fixed top-0 left-0 z-50 pointer-events-none hidden xl:flex -ml-3.5 -mt-3.5 w-7 h-7 items-center justify-center"
    >
      <span className="w-7 h-7 rounded-full border border-[#1D4ED8]/45 flex items-center justify-center bg-[#1D4ED8]/[0.03]">
        <span className="w-1 h-1 rounded-full bg-[#1D4ED8]" />
      </span>
    </motion.div>
  );
}
