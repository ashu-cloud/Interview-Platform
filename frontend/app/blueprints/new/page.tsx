"use client";

import React, { useState, useEffect, Suspense } from "react";
import Link from "next/link";
import { useSearchParams, useRouter } from "next/navigation";
import DashboardSidebar from "@/components/layout/DashboardSidebar";
import { api, Track, QuestionPrompt, EvaluationCriterion } from "@/lib/api";

function BlueprintBuilderContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const queryTrackId = searchParams.get("trackId") || "";

  // Tracks list for track selector
  const [tracks, setTracks] = useState<Track[]>([]);
  const [selectedTrackId, setSelectedTrackId] = useState(queryTrackId);
  const [loadingTracks, setLoadingTracks] = useState(true);

  // Blueprint Meta
  const [blueprintTitle, setBlueprintTitle] = useState("");
  const [instructions, setInstructions] = useState(
    "Probe architectural trade-offs, concurrency invariants, and fault-domain resilience under scale."
  );
  const [targetDuration, setTargetDuration] = useState(40);

  // Questions and Rubric lists
  const [questions, setQuestions] = useState<QuestionPrompt[]>([]);
  const [rubrics, setRubrics] = useState<EvaluationCriterion[]>([]);

  // Modals for adding items
  const [isAddQModalOpen, setIsAddQModalOpen] = useState(false);
  const [newQText, setNewQText] = useState("");
  const [newQHint, setNewQHint] = useState("");

  const [isAddRubricModalOpen, setIsAddRubricModalOpen] = useState(false);
  const [newRubricName, setNewRubricName] = useState("");
  const [newRubricDesc, setNewRubricDesc] = useState("");
  const [newRubricMax, setNewRubricMax] = useState(10);

  // Submit states
  const [submitting, setSubmitting] = useState(false);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3500);
  };

  useEffect(() => {
    async function fetchTracks() {
      setLoadingTracks(true);
      try {
        const data = await api.getTracks();
        setTracks(data);
        if (data.length > 0 && !selectedTrackId) {
          setSelectedTrackId(data[0].id);
        }
      } catch (err) {
        console.error("Failed to fetch tracks:", err);
      } finally {
        setLoadingTracks(false);
      }
    }
    fetchTracks();
  }, [selectedTrackId]);

  const handleAddQuestion = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newQText.trim()) return;

    const newQ: QuestionPrompt = {
      order_index: questions.length + 1,
      question_text: newQText.trim(),
      follow_up_hint: newQHint.trim() || undefined,
    };

    setQuestions([...questions, newQ]);
    setIsAddQModalOpen(false);
    setNewQText("");
    setNewQHint("");
    showToast("Assessment question prompt added.");
  };

  const handleRemoveQuestion = (index: number) => {
    setQuestions(questions.filter((_, i) => i !== index));
  };

  const handleAddRubric = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newRubricName.trim()) return;

    const newR: EvaluationCriterion = {
      order_index: rubrics.length + 1,
      name: newRubricName.trim(),
      description: newRubricDesc.trim() || "Demonstrates core technical competence in this domain.",
      max_score: Number(newRubricMax) || 10,
    };

    setRubrics([...rubrics, newR]);
    setIsAddRubricModalOpen(false);
    setNewRubricName("");
    setNewRubricDesc("");
    setNewRubricMax(10);
    showToast("Evaluation rubric criterion added.");
  };

  const handleRemoveRubric = (index: number) => {
    setRubrics(rubrics.filter((_, i) => i !== index));
  };

  const handlePublishBlueprint = async () => {
    const effectiveTrackId = selectedTrackId || (tracks[0] ? tracks[0].id : "");
    if (!effectiveTrackId) {
      showToast("Please select or create an assessment track first.");
      return;
    }
    if (!blueprintTitle.trim()) {
      showToast("Please enter a title for this blueprint.");
      return;
    }

    setSubmitting(true);
    try {
      // 1. Create blueprint
      const bp = await api.createBlueprint(effectiveTrackId, {
        title: blueprintTitle.trim(),
        instructions: instructions.trim(),
        duration_mins: targetDuration,
        rubric: rubrics.map((r, idx) => ({
          name: r.name,
          description: r.description,
          max_score: r.max_score,
          order_index: idx + 1,
        })),
        questions: questions.map((q, idx) => ({
          order_index: idx + 1,
          question_text: q.question_text,
          follow_up_hint: q.follow_up_hint,
        })),
      });

      showToast(`Blueprint '${bp.title}' deployed and activated!`);
      setTimeout(() => {
        router.push(`/tracks`);
      }, 1200);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to deploy blueprint";
      showToast(`Error: ${msg}`);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#FBF9F5] flex text-sumi antialiased selection:bg-sumi selection:text-paper font-sans">
      <DashboardSidebar />

      <div className="flex-1 lg:pl-64 flex flex-col min-w-0">
        <main className="w-full px-6 lg:px-12 py-8 max-w-[1400px] mx-auto flex flex-col gap-8">
          
          {/* Top Breadcrumb & Deploy Button */}
          <div className="flex items-center justify-between pb-4 border-b border-[#E5E2DA] font-mono text-xs">
            <div className="flex items-center gap-2">
              <Link href="/tracks" className="text-sumi-muted hover:text-sumi transition-colors">
                Tracks
              </Link>
              <span className="text-[#CCC]">/</span>
              <span className="font-bold text-sumi tracking-wider uppercase">BLUEPRINT STUDIO</span>
            </div>

            <button
              type="button"
              onClick={handlePublishBlueprint}
              disabled={submitting}
              className="px-5 py-2.5 bg-sumi text-white hover:bg-neutral-800 transition-colors rounded-xs font-mono text-xs font-semibold tracking-wider flex items-center gap-2 shadow-xs cursor-pointer disabled:opacity-50"
            >
              <span>{submitting ? "Deploying..." : "Deploy Blueprint & Activate Track"}</span>
              <span className="text-emerald-400 font-bold">→</span>
            </button>
          </div>

          {/* Title Header matching Screen 06 Reference */}
          <div className="relative flex flex-col md:flex-row md:items-start justify-between gap-6 pb-2">
            <div className="space-y-1.5 max-w-2xl">
              <h1 className="text-3xl sm:text-4xl font-normal text-sumi tracking-tight font-serif">
                Assessment Blueprint Studio
              </h1>
              <p className="text-xs sm:text-sm font-mono text-sumi-secondary leading-relaxed">
                Configure deterministic prompts, evaluation rubrics, and durations stored directly in PostgreSQL with pgvector embeddings.
              </p>
            </div>

            {/* Right Architectural Arc Diagram matching Screen 06 Reference */}
            <div className="shrink-0 hidden md:block select-none" aria-hidden="true">
              <svg width="120" height="120" viewBox="0 0 120 120" className="overflow-visible">
                {/* Concentric rings */}
                <circle cx="60" cy="60" r="54" stroke="#E5E2DA" strokeWidth="0.8" strokeDasharray="3 3" fill="none" />
                <circle cx="60" cy="60" r="46" stroke="#141413" strokeWidth="1" fill="#FAF8F5" />
                {/* Cobalt blue quarter arc on top-right */}
                <path d="M 60 60 L 60 14 A 46 46 0 0 1 106 60 Z" fill="#1D4ED8" />
                {/* Bottom right dark quarter slice */}
                <path d="M 60 60 L 106 60 A 46 46 0 0 1 60 106 Z" fill="#141413" fillOpacity="0.85" />
                {/* Crosshairs */}
                <line x1="6" y1="60" x2="114" y2="60" stroke="#141413" strokeWidth="0.75" opacity="0.3" />
                <line x1="60" y1="6" x2="60" y2="114" stroke="#141413" strokeWidth="0.75" opacity="0.3" />
                <circle cx="60" cy="60" r="2" fill="#141413" />
              </svg>
            </div>
          </div>

          {/* Section 01: Role Track & Blueprint Scope matching Screen 06 */}
          <section className="p-6 sm:p-8 bg-white border border-[#E5E2DA] rounded-sm shadow-xs space-y-6">
            <div className="font-mono text-xs text-sumi-muted uppercase tracking-wider font-semibold border-b border-[#E5E2DA] pb-3">
              01 // ROLE TRACK &amp; BLUEPRINT SCOPE
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-12 gap-5 font-mono text-xs">
              {/* Target Role Track */}
              <div className="sm:col-span-6 space-y-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold block">
                  TARGET ROLE TRACK
                </label>
                {loadingTracks ? (
                  <div className="w-full px-3.5 py-2.5 bg-white border border-[#E5E2DA] rounded-xs font-sans text-xs text-sumi-muted">
                    Loading tracks...
                  </div>
                ) : tracks.length > 0 ? (
                  <select
                    value={selectedTrackId}
                    onChange={(e) => setSelectedTrackId(e.target.value)}
                    className="w-full px-3.5 py-2.5 bg-white border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                  >
                    {tracks.map((t) => (
                      <option key={t.id} value={t.id}>
                        {t.title}
                      </option>
                    ))}
                  </select>
                ) : (
                  <select
                    value={selectedTrackId || "sde-intern"}
                    onChange={(e) => setSelectedTrackId(e.target.value)}
                    className="w-full px-3.5 py-2.5 bg-white border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                  >
                    <option value="sde-intern">SDE Intern</option>
                    <option value="staff-architect">Staff Distributed Systems Architect</option>
                  </select>
                )}
              </div>

              {/* Blueprint Title */}
              <div className="sm:col-span-6 space-y-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold block">
                  BLUEPRINT TITLE
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Distributed Consensus & Invariant Verification"
                  value={blueprintTitle}
                  onChange={(e) => setBlueprintTitle(e.target.value)}
                  className="w-full px-3.5 py-2 bg-white border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                />
              </div>

              {/* Duration (Minutes) */}
              <div className="sm:col-span-6 space-y-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold block">
                  DURATION (MINUTES)
                </label>
                <input
                  type="number"
                  min={10}
                  max={120}
                  value={targetDuration}
                  onChange={(e) => setTargetDuration(parseInt(e.target.value) || 30)}
                  className="w-full px-3.5 py-2 bg-white border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                />
              </div>

              {/* Interview Instructions for Assessor (Voxi) */}
              <div className="sm:col-span-12 space-y-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold block">
                  INTERVIEW INSTRUCTIONS FOR ASSESSOR (VOXI)
                </label>
                <textarea
                  rows={3}
                  value={instructions}
                  onChange={(e) => setInstructions(e.target.value)}
                  className="w-full px-3.5 py-2.5 bg-white border border-[#E5E2DA] rounded-xs font-sans text-xs leading-relaxed focus:border-sumi focus:outline-none"
                />
              </div>
            </div>
          </section>

          {/* Section 02: Assessment Question Prompts matching Screen 06 */}
          <section className="p-6 sm:p-8 bg-white border border-[#E5E2DA] rounded-sm shadow-xs space-y-4">
            <div className="flex items-center justify-between border-b border-[#E5E2DA] pb-3">
              <div className="font-mono text-xs text-sumi-muted uppercase tracking-wider font-semibold">
                02 // ASSESSMENT QUESTION PROMPTS ({questions.length})
              </div>
              <button
                type="button"
                onClick={() => setIsAddQModalOpen(true)}
                className="px-3.5 py-1.5 bg-sumi text-white hover:bg-neutral-800 transition-colors rounded-xs font-mono text-xs uppercase tracking-wider font-semibold shadow-xs cursor-pointer flex items-center gap-1.5"
              >
                <span>+</span>
                <span>ADD QUESTION</span>
              </button>
            </div>

            {questions.length === 0 ? (
              <div className="p-8 text-center bg-[#FBF9F5] border border-dashed border-[#E5E2DA] rounded-sm font-mono text-xs text-sumi-secondary">
                No question prompts added yet. Click &ldquo;+ Add Question&rdquo; to define questions for Voxi to ask the participant.
              </div>
            ) : (
              <div className="space-y-3 font-mono text-xs">
                {questions.map((q, idx) => (
                  <div
                    key={idx}
                    className="p-4 bg-[#FBF9F5] border border-[#E5E2DA] rounded-xs flex items-start justify-between gap-4"
                  >
                    <div className="space-y-1 flex-1">
                      <div className="text-[10.5px] font-bold text-sumi uppercase">
                        PROMPT 0{idx + 1}
                      </div>
                      <p className="font-sans text-xs text-sumi leading-relaxed">
                        &ldquo;{q.question_text}&rdquo;
                      </p>
                      {q.follow_up_hint && (
                        <div className="text-[10.5px] text-sumi-muted italic">
                          HINT: {q.follow_up_hint}
                        </div>
                      )}
                    </div>
                    <button
                      type="button"
                      onClick={() => handleRemoveQuestion(idx)}
                      className="text-sumi-muted hover:text-red-600 transition-colors"
                      title="Remove Prompt"
                    >
                      &times;
                    </button>
                  </div>
                ))}
              </div>
            )}
          </section>

          {/* Section 03: Evaluation Rubric Criteria matching Screen 06 */}
          <section className="p-6 sm:p-8 bg-white border border-[#E5E2DA] rounded-sm shadow-xs space-y-4">
            <div className="flex items-center justify-between border-b border-[#E5E2DA] pb-3">
              <div className="font-mono text-xs text-sumi-muted uppercase tracking-wider font-semibold">
                03 // EVALUATION RUBRIC CRITERIA ({rubrics.length})
              </div>
              <button
                type="button"
                onClick={() => setIsAddRubricModalOpen(true)}
                className="px-3.5 py-1.5 bg-sumi text-white hover:bg-neutral-800 transition-colors rounded-xs font-mono text-xs uppercase tracking-wider font-semibold shadow-xs cursor-pointer flex items-center gap-1.5"
              >
                <span>+</span>
                <span>ADD RUBRIC</span>
              </button>
            </div>

            {rubrics.length === 0 ? (
              <div className="p-8 text-center bg-[#FBF9F5] border border-dashed border-[#E5E2DA] rounded-sm font-mono text-xs text-sumi-secondary">
                No evaluation criteria defined yet. Click &ldquo;+ Add Rubric&rdquo; to specify competencies for the asynchronous feedback worker to score.
              </div>
            ) : (
              <div className="space-y-3 font-mono text-xs">
                {rubrics.map((r, idx) => (
                  <div
                    key={idx}
                    className="p-4 bg-[#FBF9F5] border border-[#E5E2DA] rounded-xs flex items-start justify-between gap-4"
                  >
                    <div className="space-y-1 flex-1">
                      <div className="flex items-center justify-between text-[11px] font-bold text-sumi">
                        <span>CRITERION 0{idx + 1}: {r.name}</span>
                        <span className="text-emerald-700">MAX: {r.max_score} PTS</span>
                      </div>
                      <p className="font-sans text-xs text-sumi-secondary leading-relaxed">
                        {r.description}
                      </p>
                    </div>
                    <button
                      type="button"
                      onClick={() => handleRemoveRubric(idx)}
                      className="text-sumi-muted hover:text-red-600 transition-colors"
                      title="Remove Rubric"
                    >
                      &times;
                    </button>
                  </div>
                ))}
              </div>
            )}
          </section>
        </main>
      </div>

      {/* Modal: Add Question Prompt */}
      {isAddQModalOpen && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-xs z-50 flex items-center justify-center p-4">
          <div className="bg-white border border-[#E5E2DA] rounded-sm max-w-lg w-full p-6 shadow-xl space-y-4 font-mono text-xs">
            <div className="flex items-center justify-between pb-2 border-b border-[#E5E2DA]">
              <span className="font-bold text-sumi uppercase">Add Assessment Question Prompt</span>
              <button
                type="button"
                onClick={() => setIsAddQModalOpen(false)}
                className="text-sumi-muted hover:text-sumi text-base leading-none"
              >
                &times;
              </button>
            </div>

            <form onSubmit={handleAddQuestion} className="space-y-4">
              <div className="space-y-1">
                <label className="text-sumi-muted uppercase font-semibold">Question Prompt Text</label>
                <textarea
                  rows={3}
                  required
                  placeholder="e.g. When scaling Raft with pipelining over high-latency WAN links, how do you prevent follower memory exhaustion during batch surges?"
                  value={newQText}
                  onChange={(e) => setNewQText(e.target.value)}
                  className="w-full px-3 py-2 bg-[#FBF9F5] border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                />
              </div>

              <div className="space-y-1">
                <label className="text-sumi-muted uppercase font-semibold">Follow-up Probe Hint (Optional)</label>
                <input
                  type="text"
                  placeholder="e.g. Check for monotonic fencing token awareness and CP over AP."
                  value={newQHint}
                  onChange={(e) => setNewQHint(e.target.value)}
                  className="w-full px-3 py-2 bg-[#FBF9F5] border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                />
              </div>

              <div className="pt-2 flex items-center justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setIsAddQModalOpen(false)}
                  className="px-4 py-2 border border-[#E5E2DA] rounded-xs hover:border-sumi transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 bg-sumi text-white rounded-xs font-semibold uppercase tracking-wider hover:bg-neutral-800 transition-colors shadow-xs"
                >
                  Append Prompt
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: Add Rubric Criterion */}
      {isAddRubricModalOpen && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-xs z-50 flex items-center justify-center p-4">
          <div className="bg-white border border-[#E5E2DA] rounded-sm max-w-lg w-full p-6 shadow-xl space-y-4 font-mono text-xs">
            <div className="flex items-center justify-between pb-2 border-b border-[#E5E2DA]">
              <span className="font-bold text-sumi uppercase">Add Rubric Competency Criterion</span>
              <button
                type="button"
                onClick={() => setIsAddRubricModalOpen(false)}
                className="text-sumi-muted hover:text-sumi text-base leading-none"
              >
                &times;
              </button>
            </div>

            <form onSubmit={handleAddRubric} className="space-y-4">
              <div className="space-y-1">
                <label className="text-sumi-muted uppercase font-semibold">Criterion Name</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Consensus & Raft Architecture"
                  value={newRubricName}
                  onChange={(e) => setNewRubricName(e.target.value)}
                  className="w-full px-3 py-2 bg-[#FBF9F5] border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                />
              </div>

              <div className="space-y-1">
                <label className="text-sumi-muted uppercase font-semibold">Evaluation Rubric Description</label>
                <textarea
                  rows={2}
                  placeholder="e.g. Candidate demonstrates deep understanding of log compaction, quorum thresholds, and split-brain fencing."
                  value={newRubricDesc}
                  onChange={(e) => setNewRubricDesc(e.target.value)}
                  className="w-full px-3 py-2 bg-[#FBF9F5] border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                />
              </div>

              <div className="space-y-1">
                <label className="text-sumi-muted uppercase font-semibold">Max Score (Points)</label>
                <input
                  type="number"
                  min={1}
                  max={100}
                  value={newRubricMax}
                  onChange={(e) => setNewRubricMax(parseInt(e.target.value) || 10)}
                  className="w-full px-3 py-2 bg-[#FBF9F5] border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                />
              </div>

              <div className="pt-2 flex items-center justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setIsAddRubricModalOpen(false)}
                  className="px-4 py-2 border border-[#E5E2DA] rounded-xs hover:border-sumi transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 bg-sumi text-white rounded-xs font-semibold uppercase tracking-wider hover:bg-neutral-800 transition-colors shadow-xs"
                >
                  Append Criterion
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {toastMessage && (
        <div className="fixed bottom-6 right-6 z-50 p-3 bg-sumi text-white font-mono text-xs rounded shadow-lg animate-in fade-in">
          {toastMessage}
        </div>
      )}
    </div>
  );
}

export default function BlueprintBuilderPage() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-[#FBF9F5] flex items-center justify-center font-mono text-xs">Loading blueprint studio...</div>}>
      <BlueprintBuilderContent />
    </Suspense>
  );
}
