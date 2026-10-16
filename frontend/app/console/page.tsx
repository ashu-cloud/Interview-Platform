"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import DashboardSidebar from "@/components/layout/DashboardSidebar";
import KineticSculpture from "@/components/ui/KineticSculpture";
import { api, Track, SessionItem } from "@/lib/api";

export default function ConsolePage() {
  const [tracks, setTracks] = useState<Track[]>([]);
  const [sessions, setSessions] = useState<SessionItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3000);
  };

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
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
    loadData();
  }, []);

  // Compute real metrics
  const completedSessions = sessions.filter(
    (s) => s.status === "completed" || s.status === "feedback_ready"
  );
  const inProgressSessions = sessions.filter((s) => s.status === "in_progress");

  const averageScore =
    completedSessions.length > 0
      ? (
          completedSessions.reduce((acc, s) => acc + (s.score || 0), 0) /
          completedSessions.length
        ).toFixed(1)
      : "—";

  return (
    <div className="min-h-screen bg-[#fbf9f5] flex text-sumi antialiased selection:bg-sumi selection:text-paper font-sans">
      <DashboardSidebar />

      <div className="flex-1 lg:pl-64 flex flex-col min-w-0">
        <header className="sticky top-0 h-20 bg-[#fbf9f5]/90 backdrop-blur-md z-40 flex items-center justify-between px-6 lg:px-12 border-b border-[#e5e2da]">
          <div className="flex items-center gap-4">
            <span className="font-mono text-xs uppercase tracking-wider text-sumi font-bold">
              Orchestration Console
            </span>
          </div>

          <div className="flex items-center gap-3">
            <Link
              href="/tracks"
              className="text-xs font-mono bg-sumi text-white hover:bg-neutral-800 transition-colors px-4 py-2 rounded-xs font-semibold shadow-xs flex items-center gap-1.5"
            >
              <span className="text-emerald-400 font-bold text-sm">+</span>
              <span>New Role Track</span>
            </Link>
          </div>
        </header>

        <main className="w-full px-6 lg:px-12 py-8 max-w-6xl mx-auto flex flex-col pb-24 gap-8">
          {/* Header Title Bar */}
          <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 pb-2 border-b border-[#e5e2da]">
            <div>
              <div className="flex items-center gap-2 font-mono text-xs text-sumi-muted uppercase tracking-wider mb-1">
                <span>Evaluation Engine Core</span>
                <span className="text-[#c7c7c0]">/</span>
                <span className="text-emerald-700 font-semibold">Live Operational Ledger</span>
              </div>
              <h1 className="font-sans text-2xl sm:text-3xl font-medium text-sumi tracking-tight">
                Talent Operations Console
              </h1>
              <p className="font-sans text-sm text-sumi-secondary font-light mt-1">
                Autonomous voice interrogation metrics, active track profiles, and verifiable evidence dossiers.
              </p>
            </div>

            <button
              type="button"
              onClick={loadData}
              className="px-3.5 py-2 bg-white border border-[#e5e2da] hover:border-sumi rounded-xs font-mono text-xs uppercase tracking-wider text-sumi transition-colors shadow-xs"
            >
              ↻ Refresh
            </button>
          </div>

          {error && (
            <div className="p-4 bg-red-50 border border-red-300 rounded font-mono text-xs text-red-800 flex items-center justify-between">
              <span>Error fetching platform data: {error}</span>
              <button onClick={loadData} className="underline uppercase font-bold text-[11px] ml-4 cursor-pointer">
                Retry
              </button>
            </div>
          )}

          {/* Metric KPI Plates */}
          <section className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="bg-white border border-[#e5e2da] p-5 rounded shadow-xs flex flex-col justify-between">
              <span className="font-mono text-[10px] text-sumi-muted uppercase tracking-wider font-semibold">
                ACTIVE ROLE TRACKS
              </span>
              <div className="text-3xl font-mono font-bold text-sumi mt-2">
                {tracks.length}
              </div>
              <span className="font-mono text-[10px] text-sumi-secondary mt-1">
                Configured in PostgreSQL
              </span>
            </div>

            <div className="bg-white border border-[#e5e2da] p-5 rounded shadow-xs flex flex-col justify-between">
              <span className="font-mono text-[10px] text-sumi-muted uppercase tracking-wider font-semibold">
                SESSIONS IN PROGRESS
              </span>
              <div className="text-3xl font-mono font-bold text-ultramarine mt-2">
                {inProgressSessions.length}
              </div>
              <span className="font-mono text-[10px] text-sumi-secondary mt-1">
                Live voice stages active
              </span>
            </div>

            <div className="bg-white border border-[#e5e2da] p-5 rounded shadow-xs flex flex-col justify-between">
              <span className="font-mono text-[10px] text-sumi-muted uppercase tracking-wider font-semibold">
                AUDITED EVALUATIONS
              </span>
              <div className="text-3xl font-mono font-bold text-emerald-700 mt-2">
                {completedSessions.length}
              </div>
              <span className="font-mono text-[10px] text-sumi-secondary mt-1">
                Scored by evaluation-worker
              </span>
            </div>

            <div className="bg-white border border-[#e5e2da] p-5 rounded shadow-xs flex flex-col justify-between">
              <span className="font-mono text-[10px] text-sumi-muted uppercase tracking-wider font-semibold">
                MEAN COMPOSITE SCORE
              </span>
              <div className="text-3xl font-mono font-bold text-sumi mt-2">
                {averageScore}
              </div>
              <span className="font-mono text-[10px] text-sumi-secondary mt-1">
                Across completed sessions
              </span>
            </div>
          </section>

          {/* Active Tracks Section */}
          <section className="space-y-4">
            <div className="flex items-center justify-between border-b border-[#e5e2da] pb-2 font-mono text-xs">
              <span className="text-sumi-muted uppercase tracking-wider font-semibold">
                ASSESSMENT ROLE TRACKS ({tracks.length})
              </span>
              <Link href="/tracks" className="text-ultramarine hover:underline">
                View All Tracks →
              </Link>
            </div>

            {loading ? (
              <div className="p-8 bg-white border border-[#e5e2da] rounded text-center font-mono text-xs text-sumi-muted animate-pulse">
                Querying platform-service tracks...
              </div>
            ) : tracks.length === 0 ? (
              <div className="p-8 bg-white border border-dashed border-[#e5e2da] rounded text-center font-mono text-xs text-sumi-muted">
                No role tracks configured yet.{" "}
                <Link href="/tracks" className="underline font-bold text-sumi">
                  Create your first track
                </Link>{" "}
                to start conducting assessments.
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {tracks.slice(0, 4).map((t) => (
                  <div
                    key={t.id}
                    className="p-5 bg-white border border-[#e5e2da] rounded shadow-xs flex flex-col justify-between gap-4 hover:border-sumi transition-colors"
                  >
                    <div>
                      <div className="flex justify-between items-center font-mono text-[10px] text-sumi-muted mb-1">
                        <span>ID: {t.id.slice(0, 8)}...</span>
                        <span className="text-emerald-700 font-bold uppercase">ACTIVE</span>
                      </div>
                      <h3 className="font-sans text-base font-medium text-sumi">
                        {t.title}
                      </h3>
                      <p className="font-sans text-xs text-sumi-secondary mt-1 line-clamp-2">
                        {t.description || "Configured technical assessment track."}
                      </p>
                    </div>

                    <div className="pt-3 border-t border-[#e5e2da] flex items-center justify-between font-mono text-xs">
                      <Link
                        href={`/blueprints/new?trackId=${t.id}`}
                        className="text-sumi hover:underline"
                      >
                        Configure Blueprint →
                      </Link>
                      <Link
                        href={`/candidates?trackId=${t.id}`}
                        className="px-3 py-1 bg-sumi text-white hover:bg-neutral-800 rounded-xs uppercase text-[10px]"
                      >
                        Candidates
                      </Link>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>

          {/* Recent Candidate Sessions Section */}
          <section className="space-y-4">
            <div className="flex items-center justify-between border-b border-[#e5e2da] pb-2 font-mono text-xs">
              <span className="text-sumi-muted uppercase tracking-wider font-semibold">
                RECENT SESSIONS &amp; EVALUATION LEDGER ({sessions.length})
              </span>
              <Link href="/candidates" className="text-ultramarine hover:underline">
                View Full Pipeline →
              </Link>
            </div>

            {loading ? (
              <div className="p-8 bg-white border border-[#e5e2da] rounded text-center font-mono text-xs text-sumi-muted animate-pulse">
                Querying recent sessions...
              </div>
            ) : sessions.length === 0 ? (
              <div className="p-8 bg-white border border-dashed border-[#e5e2da] rounded text-center font-mono text-xs text-sumi-muted">
                No assessment sessions recorded in the database yet.{" "}
                <Link href="/candidates" className="underline font-bold text-sumi">
                  Issue an invite link
                </Link>{" "}
                to start an interview.
              </div>
            ) : (
              <div className="bg-white border border-[#e5e2da] rounded shadow-xs overflow-hidden">
                <table className="w-full text-left font-sans text-xs">
                  <thead>
                    <tr className="bg-[#fbf9f5] border-b border-[#e5e2da] font-mono text-[10px] text-sumi-muted uppercase tracking-wider">
                      <th className="py-3 px-4">Candidate</th>
                      <th className="py-3 px-4">Track</th>
                      <th className="py-3 px-4">Status</th>
                      <th className="py-3 px-4">Score</th>
                      <th className="py-3 px-4 text-right">Action</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[#e5e2da]">
                    {sessions.slice(0, 8).map((s) => (
                      <tr key={s.id} className="hover:bg-[#fbf9f5]/60 transition-colors">
                        <td className="py-3.5 px-4">
                          <div className="font-medium text-sumi">
                            {s.participant_name || "Anonymous Candidate"}
                          </div>
                          <div className="font-mono text-[11px] text-sumi-muted">
                            {s.participant_email || `ID: ${s.participant_id.slice(0, 8)}...`}
                          </div>
                        </td>
                        <td className="py-3.5 px-4 font-mono text-[11px] text-sumi-secondary">
                          {s.track_title || "General Architecture"}
                        </td>
                        <td className="py-3.5 px-4 font-mono text-[10px] uppercase">
                          <span
                            className={`px-2 py-0.5 rounded-xs font-bold inline-flex items-center gap-1 ${
                              s.status === "completed" || s.status === "feedback_ready"
                                ? "bg-emerald-100 text-emerald-800"
                                : s.status === "in_progress"
                                ? "bg-blue-100 text-blue-800 animate-pulse"
                                : "bg-amber-100 text-amber-800"
                            }`}
                          >
                            {s.status}
                          </span>
                        </td>
                        <td className="py-3.5 px-4 font-mono font-bold">
                          {s.score ? `${s.score.toFixed(1)} / 10` : "—"}
                        </td>
                        <td className="py-3.5 px-4 text-right font-mono text-[11px]">
                          <Link
                            href={`/report/${s.id}`}
                            className="px-3 py-1.5 bg-white border border-[#e5e2da] hover:border-sumi rounded-xs uppercase tracking-wider font-semibold transition-colors"
                          >
                            Dossier →
                          </Link>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>
        </main>
      </div>

      {toastMessage && (
        <div className="fixed bottom-6 right-6 z-50 bg-sumi text-white px-4 py-2.5 rounded shadow-xl font-mono text-xs flex items-center gap-2 animate-in fade-in slide-in-from-bottom-2">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
          <span>{toastMessage}</span>
        </div>
      )}
    </div>
  );
}
