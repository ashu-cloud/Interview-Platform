"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import Logo from "@/components/ui/Logo";
import { api } from "@/lib/api";

export default function Header() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  useEffect(() => {
    queueMicrotask(() => {
      setIsLoggedIn(api.isAuthenticated());
    });
  }, []);

  const navLinks = [
    { label: "Streaming Voice", href: "#pipeline" },
    { label: "Adaptive Coverage", href: "#coverage" },
    { label: "Predictive RAG", href: "#speculative-rag" },
    { label: "Scatter-Gather", href: "#scatter-gather" },
    { label: "Resilience Grid", href: "#architecture" },
    { label: "Interface Studio", href: "#showcase" },
  ];

  return (
    <header className="sticky top-0 z-50 bg-[#fbf9f5]/92 backdrop-blur-md hairline-b" data-purpose="navigation-header">
      {/* Main Navigation Bar */}
      <div className="max-w-7xl 2xl:max-w-[1440px] mx-auto px-6 lg:px-12 h-20 flex items-center justify-between">
        {/* Brand Logo - Original Logo Preserved without decorative containers */}
        <div className="flex items-center gap-3.5 shrink-0">
          <Link
            className="group flex items-center transition-transform hover:scale-102"
            href="/"
            aria-label="VoxAssess Homepage"
          >
            <Logo
              className="w-9 h-9 sm:w-10 sm:h-10 text-sumi transition-colors"
              size={40}
            />
          </Link>
          <div className="flex flex-col border-l border-hairline pl-3 py-0.5 font-mono text-[10px] text-sumi-muted leading-tight select-none">
            <span className="font-semibold text-sumi uppercase tracking-widest text-[10px]">VOXASSESS</span>
            <span className="tracking-wider text-[9px] text-sumi-secondary">SYSTEM MONOGRAPH v4.2</span>
          </div>
        </div>

        {/* Desktop Navigation Links */}
        <nav
          className="hidden xl:flex items-center gap-5 2xl:gap-7 text-xs font-mono uppercase tracking-widest text-sumi-secondary ml-10 lg:ml-16 mr-auto"
          data-purpose="primary-links"
        >
          {navLinks.map((link) => (
            <a
              key={link.href}
              className="hover:text-sumi transition-colors py-2 relative group focus:outline-hidden focus:text-sumi"
              href={link.href}
            >
              {link.label}
              <span className="absolute bottom-1 left-0 w-0 h-0.5 bg-sumi transition-all group-hover:w-full" />
            </a>
          ))}
        </nav>

        {/* Action Cluster */}
        <div className="flex items-center gap-3.5 shrink-0 ml-4">
          <a
            className="hidden sm:inline-flex items-center gap-2 text-xs font-mono px-3.5 py-2 rounded-xs border border-hairline hover:border-sumi bg-white text-sumi transition-all shadow-xs"
            href="https://github.com/ashu-cloud/Interview-Platform"
            target="_blank"
            rel="noopener noreferrer"
            title="GitHub Repository"
          >
            <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
              <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
              />
            </svg>
            <span>GitHub</span>
          </a>

          {isLoggedIn ? (
            <Link
              className="inline-flex items-center gap-2 text-xs font-mono font-medium px-4 py-2 rounded-xs bg-sumi text-paper hover:bg-neutral-800 transition-colors shadow-xs"
              href="/console"
            >
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
              <span>Console</span>
              <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                <path d="M5 12h14M12 5l7 7-7 7" strokeLinecap="round" strokeLinejoin="round" />
              </svg>
            </Link>
          ) : (
            <Link
              className="inline-flex items-center gap-2 text-xs font-mono font-medium px-4 py-2 rounded-xs bg-sumi text-paper hover:bg-neutral-800 transition-colors shadow-xs"
              href="/auth"
            >
              <span>Login</span>
              <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                <path d="M5 12h14M12 5l7 7-7 7" strokeLinecap="round" strokeLinejoin="round" />
              </svg>
            </Link>
          )}

          {/* Accessible Mobile Menu Button */}
          <button
            type="button"
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            className="xl:hidden p-2 rounded-xs border border-hairline bg-white text-sumi hover:bg-paper-card focus:outline-hidden"
            aria-label="Toggle Navigation Drawer"
            aria-expanded={isMobileMenuOpen}
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="1.75">
              {isMobileMenuOpen ? (
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              ) : (
                <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />
              )}
            </svg>
          </button>
        </div>
      </div>

      {/* Responsive Mobile Drawer */}
      {isMobileMenuOpen && (
        <div className="xl:hidden hairline-t bg-paper/98 border-b border-hairline px-6 py-6 font-mono text-xs shadow-lg animate-in fade-in duration-150">
          <div className="flex flex-col space-y-4">
            {navLinks.map((link) => (
              <a
                key={link.href}
                href={link.href}
                onClick={() => setIsMobileMenuOpen(false)}
                className="py-2 text-sumi-secondary hover:text-sumi uppercase tracking-wider flex items-center justify-between border-b border-hairline/60"
              >
                <span>{link.label}</span>
                <span className="text-[10px] text-sumi-muted">→</span>
              </a>
            ))}
            <div className="pt-2 flex items-center justify-between">
              <a
                href="https://github.com/ashu-cloud/Interview-Platform"
                target="_blank"
                rel="noopener noreferrer"
                className="text-sumi hover:text-sumi-secondary inline-flex items-center gap-1.5"
              >
                <span>GitHub Repository</span>
                <span>↗</span>
              </a>
              <Link
                href="/auth"
                onClick={() => setIsMobileMenuOpen(false)}
                className="px-3 py-1.5 bg-sumi text-white text-[11px] rounded-xs"
              >
                Candidate Stage Access →
              </Link>
            </div>
          </div>
        </div>
      )}
    </header>
  );
}
