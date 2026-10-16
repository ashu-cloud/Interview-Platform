"use client";

import React, { useEffect, useRef, useState } from "react";
import OldLogo from "@/components/ui/OldLogo";

export default function CustomCursor() {
  const cursorRef = useRef<HTMLDivElement>(null);
  const [isPointer, setIsPointer] = useState(false);
  const [isClicking, setIsClicking] = useState(false);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    // Only run on desktop / fine pointers
    if (typeof window === "undefined" || window.matchMedia("(pointer: coarse)").matches) {
      return;
    }

    const cursorEl = cursorRef.current;
    if (!cursorEl) return;

    let posX = -100;
    let posY = -100;

    const handlePointerMove = (e: PointerEvent) => {
      posX = e.clientX;
      posY = e.clientY;
      cursorEl.style.transform = `translate3d(${posX}px, ${posY}px, 0)`;

      if (!isVisible) {
        setIsVisible(true);
      }

      // Check if hovering interactive target
      const target = e.target as HTMLElement | null;
      if (target) {
        const interactive = Boolean(
          target.closest("a, button, [role='button'], input, select, textarea, [tabindex]:not([tabindex='-1'])")
        );
        setIsPointer(interactive);
      }
    };

    const handlePointerDown = () => setIsClicking(true);
    const handlePointerUp = () => setIsClicking(false);
    const handleMouseLeave = () => setIsVisible(false);
    const handleMouseEnter = () => setIsVisible(true);

    window.addEventListener("pointermove", handlePointerMove, { passive: true });
    window.addEventListener("pointerdown", handlePointerDown);
    window.addEventListener("pointerup", handlePointerUp);
    document.addEventListener("mouseleave", handleMouseLeave);
    document.addEventListener("mouseenter", handleMouseEnter);

    return () => {
      window.removeEventListener("pointermove", handlePointerMove);
      window.removeEventListener("pointerdown", handlePointerDown);
      window.removeEventListener("pointerup", handlePointerUp);
      document.removeEventListener("mouseleave", handleMouseLeave);
      document.removeEventListener("mouseenter", handleMouseEnter);
    };
  }, [isVisible]);

  return (
    <div
      ref={cursorRef}
      className={`fixed top-0 left-0 pointer-events-none z-[9999] will-change-transform transition-opacity duration-150 ${
        isVisible ? "opacity-100" : "opacity-0"
      }`}
      aria-hidden="true"
    >
      <div
        className={`-translate-x-1/2 -translate-y-1/2 transition-transform duration-100 ease-out flex items-center justify-center ${
          isClicking ? "scale-90" : isPointer ? "scale-125" : "scale-100"
        }`}
      >
        <OldLogo
          className={`w-10 h-10 sm:w-11 sm:h-11 transition-colors duration-150 ${
            isPointer ? "text-neutral-950 drop-shadow-[0_4px_12px_rgba(20,20,19,0.28)]" : "text-sumi drop-shadow-[0_2px_6px_rgba(20,20,19,0.18)]"
          }`}
          size={42}
        />
      </div>
    </div>
  );
}
