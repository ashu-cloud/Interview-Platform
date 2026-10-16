"use client";

import React from "react";
import Link from "next/link";
import Logo from "@/components/ui/Logo";

interface ConsoleSidebarProps {
  activeTab: string;
  setActiveTab: (tab: string) => void;
}

export default function ConsoleSidebar({ activeTab, setActiveTab }: ConsoleSidebarProps) {
  const navItems = [
    {
      id: "command",
      label: "Command",
      icon: (
        <svg className="w-4.5 h-4.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <rect x="3" y="3" width="7" height="7" rx="1" />
          <rect x="14" y="3" width="7" height="7" rx="1" />
          <rect x="14" y="14" width="7" height="7" rx="1" />
          <rect x="3" y="14" width="7" height="7" rx="1" />
        </svg>
      ),
    },
    {
      id: "dossiers",
      label: "Dossiers",
      icon: (
        <svg className="w-4.5 h-4.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z" />
        </svg>
      ),
    },
    {
      id: "live-stage",
      label: "Live Stage",
      icon: (
        <svg className="w-4.5 h-4.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M12 2v20M17 5v14M7 8v8M2 11v2M22 11v2" strokeLinecap="round" />
        </svg>
      ),
    },
    {
      id: "blueprints",
      label: "Blueprints",
      icon: (
        <svg className="w-4.5 h-4.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <rect x="3" y="3" width="18" height="18" rx="2" />
          <path d="M3 9h18M9 21V9" />
        </svg>
      ),
    },
    {
      id: "spectrograms",
      label: "Spectrograms",
      icon: (
        <svg className="w-4.5 h-4.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M3 3v18h18" strokeLinecap="round" />
          <path d="m19 9-5 5-4-4-3 3" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      ),
    },
  ];

  return (
    <aside className="w-72 bg-[#f5f3ef] border-r border-[#e5e2da] h-full flex flex-col justify-between py-6 px-6 shrink-0 shadow-[1px_0_4px_rgba(20,20,19,0.02)]">
      <div className="flex flex-col gap-8">
        {/* Brand Header */}
        <div className="flex items-center gap-3 pb-6 border-b border-[#e5e2da]">
          <Link href="/" className="group flex items-center transition-transform hover:scale-105" aria-label="Home Monograph">
            <Logo className="w-10 h-10 text-sumi" size={40} />
          </Link>
          <div className="flex flex-col">
            <span className="font-mono text-[11px] font-bold uppercase text-sumi tracking-widest leading-tight">
              Atelier Eval
            </span>
            <span className="font-mono text-[10px] text-sumi-muted uppercase tracking-wider">
              Archival Console // 01
            </span>
          </div>
        </div>

        {/* Navigation Categories */}
        <div className="flex flex-col gap-2">
          <span className="font-mono text-[10px] font-semibold text-sumi-muted uppercase tracking-widest px-3 mb-1">
            Folio Categories
          </span>
          <nav className="flex flex-col gap-1.5" aria-label="Console Navigation">
            {navItems.map((item) => {
              const isActive = activeTab === item.id;
              return (
                <button
                  key={item.id}
                  onClick={() => setActiveTab(item.id)}
                  className={`w-full flex items-center gap-3 px-3.5 py-2.5 rounded text-xs font-mono tracking-wide transition-all ${
                    isActive
                      ? "bg-sumi text-[#fbf9f5] font-semibold shadow-xs"
                      : "text-sumi-secondary hover:text-sumi hover:bg-[#eae8e4]"
                  }`}
                >
                  <span className={isActive ? "text-[#fbf9f5]" : "text-sumi-muted"}>{item.icon}</span>
                  <span>{item.label}</span>
                </button>
              );
            })}
          </nav>
        </div>
      </div>

      {/* Stream State Telemetry Badge & Return Link */}
      <div className="flex flex-col gap-3">
        <div className="flex flex-col gap-1.5 bg-[#eae8e4]/60 p-3.5 rounded border border-[#e5e2da]">
          <div className="flex items-center justify-between">
            <span className="font-mono text-[10px] font-semibold text-sumi tracking-widest">STREAM STATE</span>
            <span className="w-2 h-2 rounded-full bg-ultramarine animate-pulse"></span>
          </div>
          <span className="font-mono text-[10px] text-sumi-muted">48kHz • 24-bit Flac • Direct Well</span>
          <span className="font-mono text-[10px] font-medium text-sumi">LATENCY 4.2ms</span>
        </div>

        <Link
          href="/"
          className="flex items-center justify-center gap-2 py-2 px-3 text-[11px] font-mono uppercase tracking-wider text-sumi-muted hover:text-sumi transition-colors border border-transparent hover:border-hairline rounded"
        >
          <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="m15 18-6-6 6-6" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
          <span>Return to Monograph</span>
        </Link>
      </div>
    </aside>
  );
}
