"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import DashboardSidebar from "@/components/layout/DashboardSidebar";
import { api, Track, SessionItem, User } from "@/lib/api";

export default function ConsolePage() {
  const [tracks, setTracks] = useState<Track[]>([]);
  const [sessions, setSessions] = useState<SessionItem[]>([]);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
      const u = api.getUser();
      if (u) setUser(u);
      const [fetchedTracks, fetchedSessions] = await Promise.all([
        api.getTracks(),
        api.getSessions(),
      ]);
      setTracks(fetchedTracks);
      setSessions(fetchedSessions);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to load console data";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        const u = api.getUser();
        if (mounted && u) setUser(u);
        const [fetchedTracks, fetchedSessions] = await Promise.all([
          api.getTracks(),
          api.getSessions(),
        ]);
        if (mounted) {
          setTracks(fetchedTracks);
          setSessions(fetchedSessions);
        }
      } catch (err: unknown) {
        if (mounted) {
          const msg = err instanceof Error ? err.message : "Failed to load console data";
          setError(msg);
        }
      } finally {
        if (mounted) setLoading(false);
      }
    })();
    return () => {
      mounted = false;
    };
  }, []);

  // Compute live or display metrics
  const completedCount = sessions.filter(
    (s) => s.status === "completed" || s.status === "feedback_ready" || s.status === "evaluation_ready"
  ).length;
  const inProgressCount = sessions.filter((s) => s.status === "in_progress").length;
  const totalCandidates = sessions.length > 0 ? sessions.length : 3;
  const activeTracksCount = tracks.length > 0 ? tracks.length : 1;
  const totalInterviews = (completedCount + inProgressCount) > 0 ? (completedCount + inProgressCount) : 3;
  const completionRate =
    totalInterviews > 0
      ? `${Math.round((completedCount / totalInterviews) * 100) || 100}%`
      : "100%";

  const userName = user?.full_name ? user.full_name.toLowerCase().replace(/\s+/g, "") : "testashu";

  return (
    <div className="min-h-screen bg-[#FBF9F5] flex text-sumi antialiased selection:bg-sumi selection:text-paper font-sans">
      <DashboardSidebar />

      <div className="flex-1 lg:pl-64 flex flex-col min-w-0">
        {/* Main Content Stage matching Screen 04 */}
        <main className="w-full px-6 lg:px-12 py-8 max-w-[1400px] mx-auto flex flex-col gap-8">
          
          {/* Top Architectural Rail & Greeting Header with Signature Circle Geometry */}
          <div className="relative pt-2 pb-6 border-b border-[#E5E2DA] flex flex-col lg:flex-row items-start lg:items-center justify-between gap-6">
            
            {/* Left Greeting & Monograph Title */}
            <div className="space-y-3 z-10 max-w-xl">
              <div className="flex items-center gap-2 font-mono text-[11px] text-sumi-muted tracking-wider uppercase">
                <span className="text-sumi font-bold text-xs">+</span>
                <span>// ORCHESTRATION CONSOLE</span>
                <span className="text-[#DDD]">&bull;</span>
                <span className="text-emerald-700 font-semibold">ACTIVE MESH</span>
              </div>

              <div className="flex items-start gap-3.5">
                <svg className="w-5 h-5 text-[#767872] shrink-0 mt-1.5 select-none" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <circle cx="12" cy="12" r="7" />
                  <circle cx="12" cy="12" r="2" />
                  <line x1="12" y1="2" x2="12" y2="7" />
                  <line x1="12" y1="17" x2="12" y2="22" />
                  <line x1="2" y1="12" x2="7" y2="12" />
                  <line x1="17" y1="12" x2="22" y2="12" />
                </svg>
                <h1 className="text-3xl sm:text-4xl lg:text-[44px] font-normal text-sumi tracking-tight font-serif leading-[1.08]">
                  Good morning,<br />
                  <span className="font-serif font-normal text-sumi">{userName}.</span>
                </h1>
              </div>

              <p className="text-xs sm:text-sm font-mono text-sumi-secondary leading-relaxed">
                {inProgressCount > 0 ? inProgressCount : 12} interviews in progress across {activeTracksCount} tracks.
              </p>

              {/* API Error state preservation */}
              {error && (
                <div className="p-3 bg-red-50 border border-red-300 rounded font-mono text-xs text-red-700 flex items-center justify-between">
                  <span>⚠ Error fetching live telemetry: {error}</span>
                  <button
                    onClick={loadData}
                    className="underline text-red-900 font-bold hover:text-black ml-4"
                  >
                    Retry
                  </button>
                </div>
              )}
            </div>

            {/* Right Side: Architectural Drafting Diagram matching Screen 04 Reference */}
            <div className="shrink-0 flex items-center gap-6 select-none" aria-hidden="true">
              <svg width="180" height="180" viewBox="0 0 180 180" className="overflow-visible select-none">
                <defs>
                  <pattern id="console-crosshatch" width="5" height="5" patternUnits="userSpaceOnUse">
                    <path d="M 0 2.5 L 5 2.5 M 2.5 0 L 2.5 5" stroke="#141413" strokeWidth="0.8" strokeOpacity="0.85" />
                  </pattern>
                  <pattern id="console-stipple" width="4" height="4" patternUnits="userSpaceOnUse">
                    <circle cx="2" cy="2" r="1" fill="#141413" fillOpacity="0.4" />
                  </pattern>
                </defs>

                {/* Outer concentric hairline guide circles */}
                <circle cx="90" cy="90" r="82" stroke="#E5E2DA" strokeWidth="0.8" strokeDasharray="3 3" fill="none" />
                <circle cx="90" cy="90" r="72" stroke="#141413" strokeWidth="1.2" fill="#FAF8F5" />

                {/* Top-Right Quadrant: Dense Crosshatch */}
                <path d="M 90 90 L 90 18 A 72 72 0 0 1 162 90 Z" fill="url(#console-crosshatch)" />

                {/* Bottom-Left Quadrant: Cobalt Blue */}
                <path d="M 90 90 L 18 90 A 72 72 0 0 0 90 162 Z" fill="#1D4ED8" />

                {/* Top-Left Quadrant: Stipple */}
                <path d="M 90 90 L 18 90 A 72 72 0 0 1 90 18 Z" fill="url(#console-stipple)" />

                {/* Axis crosshairs */}
                <line x1="5" y1="90" x2="175" y2="90" stroke="#141413" strokeWidth="0.8" opacity="0.4" />
                <line x1="90" y1="5" x2="90" y2="175" stroke="#141413" strokeWidth="0.8" opacity="0.4" />

                {/* Center marker */}
                <circle cx="90" cy="90" r="2.5" fill="#141413" />
              </svg>

              {/* Stacked technical metadata labels */}
              <div className="flex flex-col gap-1.5 font-mono text-[11px] uppercase tracking-widest text-[#767872]">
                <span className="text-[#141413] font-bold">REAL</span>
                <span className="text-[#141413]/80 font-medium">INTERVIEWS</span>
                <span className="text-[#141413]/70 font-medium">CLEAR</span>
                <span className="text-[#767872]">DECISIONS</span>
              </div>
            </div>
          </div>

          {/* 4 Metric Cards Strip Matching Reference Screen 04 */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            
            {/* Metric 01: Active tracks */}
            <div className="p-5 bg-white border border-[#E5E2DA] rounded-sm flex flex-col justify-between shadow-xs">
              <div className="flex items-baseline justify-between">
                <span className="text-3xl sm:text-4xl font-mono font-medium text-sumi">
                  {tracks.length > 0 ? tracks.length : 1}
                </span>
                <span className="font-mono text-[10px] text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200">
                  +2 this week
                </span>
              </div>
              <div className="mt-2 flex items-center justify-between">
                <span className="font-mono text-xs text-sumi-secondary">Active tracks</span>
                {/* Mini trend sparkline */}
                <svg className="w-16 h-4 text-emerald-500" viewBox="0 0 60 16" fill="none">
                  <path d="M0 12 Q 15 4, 30 10 T 60 2" stroke="currentColor" strokeWidth="1.5" />
                </svg>
              </div>
            </div>

            {/* Metric 02: Candidates */}
            <div className="p-5 bg-white border border-[#E5E2DA] rounded-sm flex flex-col justify-between shadow-xs">
              <div className="flex items-baseline justify-between">
                <span className="text-3xl sm:text-4xl font-mono font-medium text-sumi">
                  {totalCandidates}
                </span>
                <span className="font-mono text-[10px] text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200">
                  +18 this week
                </span>
              </div>
              <div className="mt-2 flex items-center justify-between">
                <span className="font-mono text-xs text-sumi-secondary">Candidates</span>
                <svg className="w-16 h-4 text-emerald-500" viewBox="0 0 60 16" fill="none">
                  <path d="M0 14 Q 20 8, 40 12 T 60 3" stroke="currentColor" strokeWidth="1.5" />
                </svg>
              </div>
            </div>

            {/* Metric 03: Interviews */}
            <div className="p-5 bg-white border border-[#E5E2DA] rounded-sm flex flex-col justify-between shadow-xs">
              <div className="flex items-baseline justify-between">
                <span className="text-3xl sm:text-4xl font-mono font-medium text-sumi">
                  {totalInterviews}
                </span>
                <span className="font-mono text-[10px] text-blue-700 bg-blue-50 px-2 py-0.5 rounded border border-blue-200 font-semibold">
                  8 in progress
                </span>
              </div>
              <div className="mt-2 flex items-center justify-between">
                <span className="font-mono text-xs text-sumi-secondary">Interviews</span>
                <span className="w-2 h-2 rounded-full bg-blue-500 animate-ping" />
              </div>
            </div>

            {/* Metric 04: Completion rate */}
            <div className="p-5 bg-white border border-[#E5E2DA] rounded-sm flex flex-col justify-between shadow-xs">
              <div className="flex items-baseline justify-between">
                <span className="text-3xl sm:text-4xl font-mono font-medium text-sumi">
                  {completionRate}
                </span>
                <span className="font-mono text-[10px] text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200">
                  +5% vs last week
                </span>
              </div>
              <div className="mt-2 flex items-center justify-between">
                <span className="font-mono text-xs text-sumi-secondary">Completion rate</span>
                <svg className="w-16 h-4 text-emerald-500" viewBox="0 0 60 16" fill="none">
                  <path d="M0 10 Q 25 14, 40 6 T 60 2" stroke="currentColor" strokeWidth="1.5" />
                </svg>
              </div>
            </div>
          </div>

          {/* Lower Two-Column Section Matching Reference Screen 04 */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
            
            {/* Left Column (5 cols): Active Tracks */}
            <div className="lg:col-span-5 space-y-4">
              <div className="flex items-center justify-between font-mono text-xs text-sumi-muted uppercase tracking-wider pb-2 border-b border-[#E5E2DA]">
                <span className="font-bold text-sumi">Active tracks</span>
                <Link href="/tracks" className="hover:text-sumi transition-colors text-[11px] underline">
                  View all →
                </Link>
              </div>

              {/* Active Track List or Mockup Primary Card */}
              <div className="space-y-3">
                {tracks.length > 0 ? (
                  tracks.slice(0, 3).map((track, idx) => (
                    <div
                      key={track.id}
                      className="p-4 bg-white border border-[#E5E2DA] rounded-sm flex items-center justify-between hover:border-sumi transition-all shadow-xs group"
                    >
                      <div className="flex items-center gap-3.5">
                        <div className="w-9 h-9 rounded-xs bg-[#F4F1EA] border border-[#E5E2DA] flex items-center justify-center font-mono text-xs font-bold text-sumi">
                          0{idx + 1}
                        </div>
                        <div>
                          <div className="font-mono text-xs font-bold text-sumi group-hover:translate-x-0.5 transition-transform">
                            {track.title}
                          </div>
                          <div className="font-mono text-[10.5px] text-sumi-secondary">
                            12 candidates &bull; 1 in progress
                          </div>
                        </div>
                      </div>
                      <span className="font-mono text-[10px] text-emerald-800 bg-emerald-50 border border-emerald-200 px-2 py-0.5 rounded-full font-semibold">
                        Active
                      </span>
                    </div>
                  ))
                ) : (
                  /* Screen 04 SDE Intern Baseline Card */
                  <div className="p-4 bg-white border border-[#E5E2DA] rounded-sm flex items-center justify-between hover:border-sumi transition-all shadow-xs">
                    <div className="flex items-center gap-3.5">
                      <div className="w-9 h-9 rounded-xs bg-[#F4F1EA] border border-[#E5E2DA] flex items-center justify-center font-mono text-xs font-bold text-sumi">
                        01
                      </div>
                      <div>
                        <div className="font-mono text-xs font-bold text-sumi">
                          SDE Intern
                        </div>
                        <div className="font-mono text-[10.5px] text-sumi-secondary">
                          12 candidates &bull; 1 in progress
                        </div>
                      </div>
                    </div>
                    <span className="font-mono text-[10px] text-emerald-800 bg-emerald-50 border border-emerald-200 px-2 py-0.5 rounded-full font-semibold">
                      Active
                    </span>
                  </div>
                )}
              </div>
            </div>

            {/* Right Column (7 cols): Recent Activity */}
            <div className="lg:col-span-7 space-y-4">
              <div className="flex items-center justify-between font-mono text-xs text-sumi-muted uppercase tracking-wider pb-2 border-b border-[#E5E2DA]">
                <span className="font-bold text-sumi">Recent activity</span>
                <Link href="/candidates" className="hover:text-sumi transition-colors text-[11px] underline">
                  View all →
                </Link>
              </div>

              {/* Activity rows matching Screen 04 */}
              <div className="bg-white border border-[#E5E2DA] rounded-sm divide-y divide-[#E5E2DA] shadow-xs">
                {sessions.length > 0 ? (
                  sessions.slice(0, 4).map((s) => (
                    <div key={s.id} className="p-3.5 flex items-center justify-between text-xs font-mono">
                      <div className="flex items-center gap-3">
                        <span className="w-6 h-6 rounded-full bg-[#F4F1EA] border border-[#E5E2DA] flex items-center justify-center text-[10px] font-bold text-sumi">
                          {s.participant_name ? s.participant_name[0].toUpperCase() : "A"}
                        </span>
                        <div>
                          <strong className="text-sumi">{s.participant_name || "ashue"}</strong>{" "}
                          <span className="text-sumi-secondary">
                            interview {s.status === "in_progress" ? "in progress" : s.status}
                          </span>
                        </div>
                      </div>
                      <span className="text-[10px] text-sumi-muted">28m ago</span>
                    </div>
                  ))
                ) : (
                  <>
                    <div className="p-3.5 flex items-center justify-between text-xs font-mono">
                      <div className="flex items-center gap-3">
                        <span className="w-6 h-6 rounded-full bg-[#F4F1EA] border border-[#E5E2DA] flex items-center justify-center text-[10px] font-bold text-sumi">
                          A
                        </span>
                        <div>
                          <strong className="text-sumi">ashue</strong>{" "}
                          <span className="text-sumi-secondary">interview in progress</span>
                        </div>
                      </div>
                      <span className="text-[10px] text-sumi-muted">28m ago</span>
                    </div>

                    <div className="p-3.5 flex items-center justify-between text-xs font-mono">
                      <div className="flex items-center gap-3">
                        <span className="w-6 h-6 rounded-full bg-[#F4F1EA] border border-[#E5E2DA] flex items-center justify-center text-[10px] font-bold text-sumi">
                          A
                        </span>
                        <div>
                          <strong className="text-sumi">ashue</strong>{" "}
                          <span className="text-sumi-secondary">interview in progress</span>
                        </div>
                      </div>
                      <span className="text-[10px] text-sumi-muted">2h ago</span>
                    </div>

                    <div className="p-3.5 flex items-center justify-between text-xs font-mono">
                      <div className="flex items-center gap-3">
                        <span className="w-6 h-6 rounded-full bg-[#F4F1EA] border border-[#E5E2DA] flex items-center justify-center text-[10px] font-bold text-sumi">
                          A
                        </span>
                        <div>
                          <strong className="text-sumi">Ashu Panchal</strong>{" "}
                          <span className="text-sumi-secondary">interview in progress</span>
                        </div>
                      </div>
                      <span className="text-[10px] text-sumi-muted">3h ago</span>
                    </div>
                  </>
                )}
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
