"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import Logo from "@/components/ui/Logo";
import { api, User } from "@/lib/api";

interface DashboardSidebarProps {
  className?: string;
}

export default function DashboardSidebar({ className = "" }: DashboardSidebarProps) {
  const pathname = usePathname();
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    queueMicrotask(() => {
      setUser(api.getUser());
    });
  }, []);

  const navItems = [
    {
      title: "Roles & Orchestration",
      href: "/console",
      alternateHref: "/dashboard",
      number: "01",
    },
    {
      title: "Candidate Dossiers",
      href: "/candidates",
      number: "02",
    },
    {
      title: "Evaluation Tracks",
      href: "/tracks",
      number: "03",
    },
    {
      title: "Blueprint Studio",
      href: "/blueprints/new",
      alternateHref: "/blueprints",
      number: "04",
    },
  ];

  return (
    <aside
      className={`fixed left-0 top-0 h-full w-64 bg-[#131313] text-[#E5E2DE] z-50 flex flex-col justify-between border-r border-[#262626] select-none ${className}`}
    >
      <div className="flex flex-col">
        {/* Brand Emblem & Logo - Original Logo Preserved without decorative containers */}
        <div className="h-18 px-5 flex items-center gap-3 border-b border-[#262626]">
          <Link href="/" className="flex items-center gap-3 group" title="VoxAssess Home">
            <Logo className="w-8 h-8 text-white group-hover:text-neutral-200 transition-colors" size={32} />
            <div className="flex flex-col">
              <span className="font-sans text-sm font-semibold tracking-tight text-white leading-tight">
                VoxAssess
              </span>
              <span className="font-mono text-[9px] text-[#8E8D8A] uppercase tracking-widest font-semibold">
                MANAGER HUB
              </span>
            </div>
          </Link>
        </div>

        {/* Navigation Index Header */}
        <div className="px-5 py-3.5 flex items-center justify-between text-[#8E8D8A]">
          <span className="font-mono text-[9.5px] uppercase tracking-wider font-semibold">
            [ WORKSPACE MATRIX 1 ]
          </span>
          <span className="font-mono text-[9px] px-1.5 py-0.5 rounded-xs bg-[#222224] text-[#A09E9B] border border-[#2A2A2C]">
            v4.2
          </span>
        </div>

        {/* Navigation Links */}
        <nav className="flex flex-col px-3 gap-1.5">
          {navItems.map((item) => {
            const isActive =
              pathname === item.href ||
              (item.alternateHref && pathname === item.alternateHref) ||
              (item.href === "/blueprints/new" && pathname.startsWith("/blueprints")) ||
              (item.href !== "/console" && pathname.startsWith(item.href));

            return (
              <Link
                key={item.href}
                href={item.href}
                className={`flex items-center justify-between px-3.5 py-2.5 rounded-xs font-mono text-xs transition-all ${
                  isActive
                    ? "bg-[#1C2333] border border-[#3B82F6]/60 text-white font-semibold shadow-xs"
                    : "text-[#A3A09B] hover:text-white hover:bg-[#1C1C1E] border border-transparent"
                }`}
              >
                <div className="flex items-center gap-2.5 truncate">
                  <span
                    className={`w-1.5 h-1.5 rounded-full ${
                      isActive ? "bg-[#3B82F6] animate-pulse" : "bg-transparent"
                    }`}
                  />
                  <span className="truncate">{item.title}</span>
                </div>
                <span
                  className={`font-mono text-[10px] ml-2 ${
                    isActive ? "text-[#60A5FA] font-bold" : "text-[#6E6C68]"
                  }`}
                >
                  {item.number}
                </span>
              </Link>
            );
          })}
        </nav>
      </div>

      {/* Bottom Tenant & User Info */}
      <div className="p-4 border-t border-[#262626] flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <span className="w-8 h-8 rounded-full border border-[#3A3A3D] bg-[#222226] text-white flex items-center justify-center font-mono text-xs font-bold shadow-xs">
              {user?.full_name
                ? user.full_name
                    .split(" ")
                    .map((n) => n[0])
                    .join("")
                    .toUpperCase()
                    .slice(0, 2)
                : "T"}
            </span>
            <div className="flex flex-col">
              <span className="font-mono text-xs text-white font-semibold tracking-tight leading-tight truncate max-w-[100px]">
                {user?.full_name || "testashu"}
              </span>
              <span className="font-mono text-[9px] text-[#7A7875] tracking-wide truncate max-w-[100px]">
                {user?.email || "recruiter hub"}
              </span>
            </div>
          </div>
          <span className="flex items-center gap-1 font-mono text-[9px] text-[#4EDEA3] bg-[#0E2E1E] px-2 py-0.5 rounded-full border border-[#1B5738] font-semibold">
            <span className="w-1.5 h-1.5 rounded-full bg-[#10B981] animate-pulse" />
            + LIVE
          </span>
        </div>

        {/* Log Out Button */}
        <button
          type="button"
          onClick={() => {
            api.clearAuth();
            router.push("/auth");
          }}
          className="w-full py-1.5 px-3 border border-[#2E2E32] hover:border-red-500 hover:text-red-400 hover:bg-red-950/20 rounded-xs font-mono text-[11px] uppercase tracking-wider text-[#CCC] flex items-center justify-center gap-2 transition-colors font-medium bg-[#1A1A1C] shadow-xs cursor-pointer"
          title="Sign out of VoxAssess"
        >
          <svg
            className="w-3.5 h-3.5 text-[#888]"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
            <polyline points="16 17 21 12 16 7" />
            <line x1="21" y1="12" x2="9" y2="12" />
          </svg>
          <span>LOG OUT</span>
        </button>

        <Link
          href="/"
          className="text-center text-[10px] font-mono uppercase tracking-wider text-[#7A7875] hover:text-[#E5E2DE] transition-colors"
        >
          ← BACK TO LANDING
        </Link>
      </div>
    </aside>
  );
}
