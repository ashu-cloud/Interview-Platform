"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import Logo from "@/components/ui/Logo";
import { api, User } from "@/lib/api";

interface DashboardSidebarProps {
  className?: string;
}

export default function DashboardSidebar({ className = "" }: DashboardSidebarProps) {
  const pathname = usePathname();
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    setUser(api.getUser());
  }, []);

  const navItems = [
    {
      title: "Roles & Orchestration",
      href: "/dashboard",
      alternateHref: "/console",
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
      alternateHref: "/roles/new",
      number: "04",
    },
  ];

  return (
    <aside
      className={`fixed left-0 top-0 h-full w-64 bg-[#fbf9f5] z-50 flex flex-col justify-between border-r border-[#e5e2da] ${className}`}
    >
      <div className="flex flex-col">
        {/* Brand Emblem & Logo */}
        <div className="h-24 px-6 flex items-center gap-3.5 border-b border-[#e5e2da]">
          <Link href="/" className="flex items-center gap-3.5 group" title="VoxAssess Home">
            <Logo className="w-12 h-12 text-sumi group-hover:scale-105 transition-transform shrink-0" size={48} />
            <div className="flex flex-col">
              <span className="font-sans text-base font-semibold tracking-tight text-sumi">
                VoxAssess
              </span>
              <span className="font-mono text-[10px] text-sumi-muted uppercase tracking-widest font-semibold">
                Manager Hub
              </span>
            </div>
          </Link>
        </div>

        {/* Navigation Index Header */}
        <div className="px-6 py-4 flex items-center justify-between">
          <span className="font-mono text-[10px] uppercase tracking-widest text-sumi-muted font-semibold">
            [ Workspace Matrix ]
          </span>
          <span className="font-mono text-[9px] px-1.5 py-0.5 rounded-xs bg-[#efeeea] text-sumi-muted">
            v4.2
          </span>
        </div>

        {/* Navigation Links */}
        <nav className="flex flex-col px-3 gap-1">
          {navItems.map((item) => {
            const isActive =
              pathname === item.href ||
              (item.alternateHref && pathname === item.alternateHref) ||
              (item.href !== "/dashboard" && pathname.startsWith(item.href));

            return (
              <Link
                key={item.href}
                href={item.href}
                className={`flex items-center justify-between px-3.5 py-2.5 rounded-xs font-mono text-xs transition-colors ${
                  isActive
                    ? "bg-[#141413]/5 text-sumi font-semibold shadow-2xs border-l-2 border-sumi"
                    : "text-sumi-secondary hover:text-sumi hover:bg-[#efeeea]"
                }`}
              >
                <span className="flex items-center gap-2.5">
                  <span
                    className={`w-1.5 h-1.5 rounded-full ${
                      isActive ? "bg-ultramarine animate-pulse" : "bg-transparent"
                    }`}
                  ></span>
                  <span>{item.title}</span>
                </span>
                <span
                  className={`font-mono text-[10px] ${
                    isActive ? "text-ultramarine font-bold" : "text-sumi-muted"
                  }`}
                >
                  {item.number}
                </span>
              </Link>
            );
          })}
        </nav>
      </div>

      {/* Bottom Tenant & Cluster Info */}
      <div className="p-5 border-t border-[#e5e2da] flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="w-7 h-7 rounded-full border border-[#e5e2da] bg-[#eae8e4] flex items-center justify-center font-mono text-[10px] text-sumi font-bold shadow-xs">
              {user?.full_name
                ? user.full_name
                    .split(" ")
                    .map((n) => n[0])
                    .join("")
                    .toUpperCase()
                    .slice(0, 2)
                : "AU"}
            </span>
            <div className="flex flex-col">
              <span className="font-mono text-xs text-sumi font-semibold tracking-tight leading-none">
                {user?.full_name || "Assessment Author"}
              </span>
              <span className="font-mono text-[9px] text-sumi-muted tracking-wider">
                {user?.email || "Author Console"}
              </span>
            </div>
          </div>
          <span className="flex items-center gap-1 font-mono text-[9px] text-emerald-800 bg-emerald-100 px-1.5 py-0.5 rounded-full font-semibold">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-600 animate-pulse"></span>
            LIVE
          </span>
        </div>

        {/* Log Out Button */}
        <button
          type="button"
          onClick={() => {
            api.clearAuth();
            window.location.href = "/auth";
          }}
          className="w-full py-2 px-3 border border-[#e5e2da] hover:border-red-500 hover:text-red-600 hover:bg-red-50/50 rounded-xs font-mono text-xs uppercase tracking-wider text-sumi flex items-center justify-center gap-2 transition-colors font-semibold bg-white shadow-xs group"
          title="Sign out of VoxAssess"
        >
          <svg
            className="w-3.5 h-3.5 text-sumi-muted group-hover:text-red-500 transition-colors"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
            <polyline points="16 17 21 12 16 7" />
            <line x1="21" y1="12" x2="9" y2="12" />
          </svg>
          <span>Log Out</span>
        </button>

        <Link
          href="/"
          className="text-center text-[10px] font-mono uppercase tracking-wider text-sumi-muted hover:text-sumi transition-colors"
        >
          ← Back to Landing
        </Link>
      </div>
    </aside>
  );
}
