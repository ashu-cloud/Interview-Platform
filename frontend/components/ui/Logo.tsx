"use client";

import React, { useEffect, useRef } from "react";

interface LogoProps {
  className?: string;
  size?: number;
  alt?: string;
  interactive?: boolean;
}

export default function Logo({
  className = "w-8 h-8",
  size = 32,
  interactive = true,
}: LogoProps) {
  const svgRef = useRef<SVGSVGElement>(null);
  const pupilLeftRef = useRef<SVGGElement>(null);
  const pupilRightRef = useRef<SVGGElement>(null);
  const pupilCenterRef = useRef<SVGGElement>(null);

  useEffect(() => {
    if (!interactive || typeof window === "undefined") return;

    let rafId: number | null = null;
    let targetX = 0;
    let targetY = 0;
    let currentX = 0;
    let currentY = 0;

    const handlePointerMove = (e: MouseEvent) => {
      const svg = svgRef.current;
      if (!svg) return;

      const rect = svg.getBoundingClientRect();
      const centerX = rect.left + rect.width / 2;
      const centerY = rect.top + rect.height / 2;

      const deltaX = e.clientX - centerX;
      const deltaY = e.clientY - centerY;
      const distance = Math.hypot(deltaX, deltaY);

      if (distance < 2) {
        targetX = 0;
        targetY = 0;
        return;
      }

      // Max eye displacement in SVG viewBox units (12 viewBox units inside the 64x64 eye socket)
      const maxOffset = 12;
      const influenceRadius = 400;
      const magnitude = Math.min(distance / influenceRadius, 1.0) * maxOffset;
      const angle = Math.atan2(deltaY, deltaX);

      targetX = Math.cos(angle) * magnitude;
      targetY = Math.sin(angle) * magnitude;
    };

    const handleMouseLeave = () => {
      targetX = 0;
      targetY = 0;
    };

    // Organic interpolation (lerp) for lifelike gaze tracking
    const updateLoop = () => {
      currentX += (targetX - currentX) * 0.18;
      currentY += (targetY - currentY) * 0.18;

      const moveX = Math.round(currentX * 100) / 100;
      const moveY = Math.round(currentY * 100) / 100;

      if (pupilLeftRef.current) {
        pupilLeftRef.current.setAttribute("transform", `translate(${moveX}, ${moveY})`);
      }
      if (pupilRightRef.current) {
        pupilRightRef.current.setAttribute("transform", `translate(${moveX}, ${moveY})`);
      }
      if (pupilCenterRef.current) {
        pupilCenterRef.current.setAttribute("transform", `translate(${moveX * 0.4}, ${moveY * 0.4})`);
      }

      rafId = requestAnimationFrame(updateLoop);
    };

    window.addEventListener("mousemove", handlePointerMove, { passive: true });
    document.addEventListener("mouseleave", handleMouseLeave);
    rafId = requestAnimationFrame(updateLoop);

    return () => {
      window.removeEventListener("mousemove", handlePointerMove);
      document.removeEventListener("mouseleave", handleMouseLeave);
      if (rafId) cancelAnimationFrame(rafId);
    };
  }, [interactive]);

  return (
    <svg
      ref={svgRef}
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 512 512"
      width={size}
      height={size}
      fill="none"
      className={`shrink-0 transition-transform duration-200 group-hover:scale-105 will-change-transform ${className}`}
      aria-label="Assessment Engine Agent Entity Logo"
    >
      {/* Exterior Archival Framing / Corner Marks */}
      <path d="M 64 104 L 64 64 L 104 64" stroke="currentColor" strokeWidth="4" strokeLinecap="square" />
      <path d="M 448 104 L 448 64 L 408 64" stroke="currentColor" strokeWidth="4" strokeLinecap="square" />
      <path d="M 64 408 L 64 448 L 104 448" stroke="currentColor" strokeWidth="4" strokeLinecap="square" />
      <path d="M 448 408 L 448 448 L 408 448" stroke="currentColor" strokeWidth="4" strokeLinecap="square" />

      {/* Subtle coordinate registration points */}
      <line x1="256" y1="44" x2="256" y2="56" stroke="currentColor" strokeWidth="2.5" />
      <line x1="256" y1="456" x2="256" y2="468" stroke="currentColor" strokeWidth="2.5" />
      <line x1="44" y1="256" x2="56" y2="256" stroke="currentColor" strokeWidth="2.5" />
      <line x1="456" y1="256" x2="468" y2="256" stroke="currentColor" strokeWidth="2.5" />

      {/* Outer Modular Container Grid (Rounded Square / Architectural Chassis) */}
      <rect x="96" y="96" width="320" height="320" rx="36" fill="currentColor" />

      {/* Inner Negative-Space Modular Entity Geometry (Negative Space Apertures) */}
      {/* Forehead / Brow band negative space cutouts */}
      <rect x="136" y="136" width="48" height="48" rx="8" fill="#FBF9F5" />
      <rect x="232" y="136" width="48" height="48" rx="8" fill="#FBF9F5" />
      <rect x="328" y="136" width="48" height="48" rx="8" fill="#FBF9F5" />

      {/* Eyes (Negative Space Sockets) */}
      <rect x="136" y="216" width="64" height="64" rx="10" fill="#FBF9F5" />
      <rect x="312" y="216" width="64" height="64" rx="10" fill="#FBF9F5" />

      {/* Central Bridge / Muzzle / Vertical Node */}
      <rect x="232" y="216" width="48" height="96" rx="8" fill="#FBF9F5" />

      {/* Mouth / Lower Jaw */}
      <rect x="184" y="280" width="48" height="80" rx="8" fill="#FBF9F5" />
      <rect x="280" y="280" width="48" height="80" rx="8" fill="#FBF9F5" />
      <rect x="152" y="328" width="48" height="48" rx="8" fill="#FBF9F5" />
      <rect x="312" y="328" width="48" height="48" rx="8" fill="#FBF9F5" />

      {/* Dynamic Cursor-Tracking Pupils */}
      {/* Left Eye Pupil (Centered baseline: cx=168, cy=248) */}
      <g ref={pupilLeftRef}>
        <rect x="156" y="236" width="24" height="24" rx="4" fill="currentColor" />
      </g>

      {/* Right Eye Pupil (Centered baseline: cx=344, cy=248) */}
      <g ref={pupilRightRef}>
        <circle cx="344" cy="248" r="10" fill="currentColor" />
      </g>

      {/* Center Nose / Aperture micro-node (Subtle 3D parallax) */}
      <g ref={pupilCenterRef}>
        <circle cx="256" cy="248" r="6" fill="currentColor" />
      </g>
    </svg>
  );
}
