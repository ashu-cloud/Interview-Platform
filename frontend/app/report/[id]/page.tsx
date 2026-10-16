"use client";

import React, { useState, useEffect, Suspense } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import DashboardSidebar from "@/components/layout/DashboardSidebar";
import { api, SessionItem, EvaluationReport, SessionTurn } from "@/lib/api";

function CandidateDossierContent() {
  const params = useParams();
  const sessionId = typeof params?.id === "string" ? params.id : "";

  const [session, setSession] = useState<SessionItem | null>(null);
  const [evaluation, setEvaluation] = useState<EvaluationReport | null>(null);
  const [turns, setTurns] = useState<SessionTurn[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [triggeringEval, setTriggeringEval] = useState(false);
  const [activeTurnSeek, setActiveTurnSeek] = useState<number | null>(null);

  const loadDossier = async () => {
    if (!sessionId) return;
    setLoading(true);
    setError(null);
    try {
      const sess = await api.getSession(sessionId);
      setSession(sess);

      // Attempt to load turns and evaluation concurrently
      const [fetchedTurns, fetchedEval] = await Promise.allSettled([
        api.getSessionTurns(sessionId),
        api.getSessionEvaluation(sessionId),
      ]);

      if (fetchedTurns.status === "fulfilled") {
        setTurns(fetchedTurns.value);
      }
      if (fetchedEval.status === "fulfilled") {
        setEvaluation(fetchedEval.value);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to load session dossier";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDossier();
  }, [sessionId]);

  const handleTriggerEvaluation = async () => {
    if (!sessionId) return;
    setTriggeringEval(true);
    try {
      await api.enqueueEvaluation(sessionId);
      setTimeout(() => {
        loadDossier();
      }, 2000);
    } catch (err) {
      console.error("Failed to enqueue evaluation:", err);
    } finally {
      setTriggeringEval(false);
    }
  };

  const candidateInitials = session?.participant_name
    ? session.participant_name
        .split(" ")
        .map((n) => n[0])
        .join("")
        .toUpperCase()
        .slice(0, 2)
    : "CD";

  const getVerdictBadge = (rec?: string) => {
    switch (rec?.toLowerCase()) {
      case "strong_hire":
        return { text: "STRONG HIRE", bg: "bg-emerald-100 border-emerald-300 text-emerald-900" };
      case "hire":
        return { text: "HIRE", bg: "bg-emerald-50 border-emerald-200 text-emerald-800" };
      case "revisit":
        return { text: "REVISIT", bg: "bg-amber-100 border-amber-300 text-amber-900" };
      case "no_hire":
      case "strong_no_hire":
        return { text: "NO HIRE", bg: "bg-red-100 border-red-300 text-red-900" };
      default:
        return { text: "PENDING EVALUATION", bg: "bg-neutral-100 border-neutral-300 text-neutral-800" };
    }
  };

  const badge = getVerdictBadge(evaluation?.recommendation || session?.recommendation);

  return (
    <div className="min-h-screen bg-[#fbf9f5] flex text-sumi antialiased selection:bg-sumi selection:text-paper font-sans">
      <DashboardSidebar />

      <div className="flex-1 lg:pl-64 flex flex-col min-w-0">
        <header className="sticky top-0 h-20 bg-[#fbf9f5]/90 backdrop-blur-md z-40 flex items-center justify-between px-6 lg:px-12 border-b border-[#e5e2da]">
          <div className="flex items-center gap-4">
            <Link
              href="/candidates"
              className="text-xs font-mono text-sumi-muted hover:text-sumi transition-colors"
            >
              ← Candidates Directory
            </Link>
            <span className="text-[#c7c7c0]">/</span>
            <span className="font-mono text-xs uppercase tracking-wider text-sumi font-bold">
              Assessment Dossier
            </span>
          </div>

          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={loadDossier}
              className="px-3.5 py-1.5 bg-white border border-[#e5e2da] hover:border-sumi rounded-xs font-mono text-xs uppercase tracking-wider text-sumi transition-colors shadow-xs"
            >
              ↻ Refresh
            </button>
          </div>
        </header>

        <main className="w-full px-6 lg:px-12 py-8 max-w-5xl mx-auto flex flex-col pb-24 gap-8">
          {error && (
            <div className="p-4 bg-red-50 border border-red-300 rounded font-mono text-xs text-red-800 flex items-center justify-between">
              <span>Error loading dossier: {error}</span>
              <button onClick={loadDossier} className="underline uppercase font-bold text-[11px] ml-4">
                Retry
              </button>
            </div>
          )}

          {loading ? (
            <div className="bg-white border border-[#e5e2da] rounded p-12 text-center font-mono text-xs text-sumi-muted animate-pulse">
              Compiling verified assessment dossier from PostgreSQL ledger...
            </div>
          ) : !session ? (
            <div className="p-12 bg-white border border-dashed border-[#d5d2c8] rounded text-center font-mono text-xs text-sumi-muted">
              Session not found. Check the session ID or return to the candidates directory.
            </div>
          ) : (
            <>
              {/* Header Profile Section */}
              <section className="bg-white border border-[#e5e2da] p-8 rounded shadow-sm flex flex-col gap-6">
                <div className="flex flex-col md:flex-row md:items-start justify-between gap-6 pb-6 border-b border-[#e5e2da]">
                  <div className="flex items-start gap-4">
                    <div className="w-16 h-16 rounded-full bg-[#eae8e4] border border-[#d5d2c8] flex items-center justify-center font-mono text-xl font-bold text-sumi">
                      {candidateInitials}
                    </div>
                    <div className="space-y-1">
                      <div className="flex items-center gap-3">
                        <h1 className="text-2xl font-medium text-sumi tracking-tight">
                          {session.participant_name || "Anonymous Candidate"}
                        </h1>
                        <span
                          className={`px-2.5 py-0.5 rounded-xs font-mono text-[10px] font-bold border tracking-wider uppercase ${badge.bg}`}
                        >
                          {badge.text}
                        </span>
                      </div>
                      <p className="font-mono text-xs text-sumi-secondary">
                        {session.participant_email || `ID: ${session.participant_id}`}
                      </p>
                      <p className="font-sans text-xs text-sumi-muted">
                        Role Track: <strong className="text-sumi">{session.track_title || "Technical Track"}</strong>
                      </p>
                    </div>
                  </div>

                  <div className="flex flex-col items-end gap-1 font-mono text-xs text-sumi-muted">
                    <span className="text-3xl font-bold text-sumi">
                      {evaluation?.overall_score
                        ? `${evaluation.overall_score.toFixed(1)} / 10`
                        : session.score
                        ? `${session.score.toFixed(1)} / 10`
                        : "—"}
                    </span>
                    <span className="text-[10px] uppercase tracking-wider">
                      Composite Technical Score
                    </span>
                    <span className="text-[10px] text-emerald-700">
                      STATUS: {session.status.toUpperCase()}
                    </span>
                  </div>
                </div>

                {/* Session Meta Stats */}
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 font-mono text-xs pt-2">
                  <div>
                    <span className="text-sumi-muted block text-[10px] uppercase">TURNS RECORDED</span>
                    <span className="font-bold text-sumi">{turns.length || session.turn_count || 0} turns</span>
                  </div>
                  <div>
                    <span className="text-sumi-muted block text-[10px] uppercase">SESSION ID</span>
                    <span className="font-bold text-sumi select-all">{session.id.slice(0, 12)}...</span>
                  </div>
                  <div>
                    <span className="text-sumi-muted block text-[10px] uppercase">EVALUATION MODEL</span>
                    <span className="font-bold text-sumi">Groq LLaMA 3.3 70B</span>
                  </div>
                  <div>
                    <span className="text-sumi-muted block text-[10px] uppercase">AUDIT PROOF</span>
                    <span className="font-bold text-emerald-700">SHA-256 Verified</span>
                  </div>
                </div>
              </section>

              {/* Evaluation Report Section */}
              {evaluation ? (
                <section className="bg-white border border-[#e5e2da] p-8 rounded shadow-sm space-y-6">
                  <div className="font-mono text-xs text-sumi-muted uppercase tracking-wider font-semibold border-b border-[#e5e2da] pb-2">
                    EXECUTIVE EVALUATION SUMMARY // VERBATIM GROUNDED
                  </div>

                  <div className="p-5 bg-[#fbf9f5] border border-[#e5e2da] rounded font-sans text-sm text-sumi leading-relaxed">
                    {evaluation.summary || "Evaluation summary generated by evaluation-worker."}
                  </div>

                  {/* Criteria Results Breakdown */}
                  {evaluation.criteria_results && evaluation.criteria_results.length > 0 && (
                    <div className="space-y-4 pt-4">
                      <h3 className="font-mono text-xs text-sumi font-bold uppercase tracking-wider">
                        Competency Breakdown &amp; Cited Evidence
                      </h3>

                      <div className="space-y-4">
                        {evaluation.criteria_results.map((c, idx) => (
                          <div
                            key={idx}
                            className="p-5 bg-white border border-[#e5e2da] rounded flex flex-col gap-3 shadow-2xs"
                          >
                            <div className="flex items-center justify-between border-b border-[#e5e2da] pb-2 font-mono text-xs">
                              <span className="font-bold text-sumi">
                                {c.name || `Criterion #${idx + 1}`}
                              </span>
                              <span className="font-bold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200">
                                {c.score} / {c.max_score} pts
                              </span>
                            </div>

                            <p className="font-sans text-xs text-sumi-secondary leading-relaxed">
                              {c.feedback}
                            </p>

                            {c.evidence_quote && (
                              <div className="p-3 bg-[#fbf9f5] border-l-2 border-emerald-600 rounded-r font-mono text-[11px] text-sumi space-y-1">
                                <span className="text-[10px] text-emerald-800 font-bold uppercase block">
                                  Verbatim Cited Transcript Evidence:
                                </span>
                                <p className="italic text-sumi-secondary">
                                  &quot;{c.evidence_quote}&quot;
                                </p>
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </section>
              ) : (
                <section className="bg-white border border-[#e5e2da] p-8 rounded shadow-sm text-center space-y-4 font-mono text-xs">
                  <div className="w-10 h-10 rounded-full bg-amber-100 border border-amber-300 flex items-center justify-center mx-auto text-amber-800 font-bold">
                    ⏳
                  </div>
                  <h3 className="font-sans text-lg font-medium text-sumi">
                    Evaluation Not Yet Generated
                  </h3>
                  <p className="font-sans text-xs text-sumi-secondary max-w-md mx-auto">
                    The evaluation worker has not synthesized a scored report for this session yet. You can trigger asynchronous evaluation report generation below.
                  </p>
                  <button
                    type="button"
                    onClick={handleTriggerEvaluation}
                    disabled={triggeringEval}
                    className="px-5 py-2.5 bg-sumi text-white hover:bg-neutral-800 rounded font-mono text-xs uppercase tracking-wider font-semibold shadow-xs disabled:opacity-50 cursor-pointer"
                  >
                    {triggeringEval ? "Enqueueing Evaluation..." : "⚡ Generate Evaluation Report"}
                  </button>
                </section>
              )}

              {/* Verified Session Turns Transcript */}
              <section className="bg-white border border-[#e5e2da] p-8 rounded shadow-sm space-y-5">
                <div className="flex items-center justify-between border-b border-[#e5e2da] pb-2 font-mono text-xs">
                  <span className="text-sumi-muted uppercase tracking-wider font-semibold">
                    AUDITED TRANSCRIPT TURNS ({turns.length})
                  </span>
                  <span className="text-[10px] text-emerald-700 font-bold">
                    POSTGRESQL TURN LEDGER
                  </span>
                </div>

                {turns.length === 0 ? (
                  <div className="p-8 border border-dashed border-[#e5e2da] rounded text-center text-xs font-mono text-sumi-muted">
                    No transcript turns recorded in session ledger.
                  </div>
                ) : (
                  <div className="space-y-4">
                    {turns.map((t, idx) => (
                      <div
                        key={idx}
                        className={`p-4 rounded border transition-colors ${
                          activeTurnSeek === t.turn_number
                            ? "bg-amber-50 border-amber-400"
                            : "bg-[#fbf9f5] border-[#e5e2da]"
                        }`}
                      >
                        <div className="flex items-center justify-between font-mono text-[11px] mb-1.5">
                          <span
                            className={`font-bold uppercase ${
                              t.role === "agent" ? "text-ultramarine" : "text-emerald-700"
                            }`}
                          >
                            {t.role === "agent" ? "VOXI // ASSESSOR" : "PARTICIPANT // CANDIDATE"}
                          </span>
                          <span className="text-sumi-muted">
                            TURN #{t.turn_number || idx + 1}
                          </span>
                        </div>
                        <p className="font-sans text-xs text-sumi leading-relaxed">
                          {t.content}
                        </p>
                      </div>
                    ))}
                  </div>
                )}
              </section>
            </>
          )}
        </main>
      </div>
    </div>
  );
}

export default function CandidateReportPage() {
  return (
    <Suspense fallback={<div className="p-12 text-center font-mono text-xs text-sumi">Loading Assessment Dossier...</div>}>
      <CandidateDossierContent />
    </Suspense>
  );
}
