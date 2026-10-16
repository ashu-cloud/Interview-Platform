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

  // Questions and Rubric lists (no seeded mock records)
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
    showToast("Rubric evaluation criterion added.");
  };

  const handleRemoveRubric = (index: number) => {
    setRubrics(rubrics.filter((_, i) => i !== index));
  };

  const handlePublishBlueprint = async () => {
    if (!selectedTrackId) {
      showToast("Please select a target Role Track first.");
      return;
    }
    if (!blueprintTitle.trim()) {
      showToast("Please provide a title for the blueprint.");
      return;
    }
    if (questions.length === 0) {
      showToast("Please add at least one question prompt.");
      return;
    }
    if (rubrics.length === 0) {
      showToast("Please add at least one rubric evaluation criterion.");
      return;
    }

    setSubmitting(true);
    try {
      const created = await api.createBlueprint(selectedTrackId, {
        title: blueprintTitle.trim(),
        instructions: instructions.trim(),
        duration_mins: targetDuration,
        questions: questions.map((q, idx) => ({ ...q, order_index: idx + 1 })),
        rubric: rubrics.map((r, idx) => ({ ...r, order_index: idx + 1 })),
      });

      showToast(`Blueprint '${created.title}' deployed successfully!`);
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
    <div className="min-h-screen bg-[#fbf9f5] flex text-sumi antialiased selection:bg-sumi selection:text-paper font-sans">
      <DashboardSidebar />

      <div className="flex-1 lg:pl-64 flex flex-col min-w-0">
        <header className="sticky top-0 h-20 bg-[#fbf9f5]/90 backdrop-blur-md z-40 flex items-center justify-between px-6 lg:px-12 border-b border-[#e5e2da]">
          <div className="flex items-center gap-4">
            <Link
              href="/tracks"
              className="text-xs font-mono text-sumi-muted hover:text-sumi transition-colors"
            >
              ← Tracks
            </Link>
            <span className="text-[#c7c7c0]">/</span>
            <span className="font-mono text-xs uppercase tracking-wider text-sumi font-bold">
              Blueprint Studio
            </span>
          </div>

          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={handlePublishBlueprint}
              disabled={submitting}
              className="text-xs font-mono bg-sumi text-white hover:bg-neutral-800 transition-colors px-4 py-2 rounded-xs font-semibold shadow-xs flex items-center gap-2 cursor-pointer disabled:opacity-50"
            >
              <span>{submitting ? "Compiling..." : "Deploy Blueprint & Activate Track"}</span>
              <span className="text-emerald-400">→</span>
            </button>
          </div>
        </header>

        <main className="w-full px-6 lg:px-12 py-8 max-w-5xl mx-auto flex flex-col pb-24 gap-8">
          <div className="border-b border-[#e5e2da] pb-4">
            <h1 className="font-sans text-2xl sm:text-3xl font-medium text-sumi tracking-tight">
              Assessment Blueprint Studio
            </h1>
            <p className="font-sans text-xs text-sumi-secondary mt-1">
              Configure deterministic prompts, evaluation rubrics, and durations stored directly in PostgreSQL with pgvector embeddings.
            </p>
          </div>

          {/* Module 1: Track Selection & Blueprint Meta */}
          <section className="bg-white border border-[#e5e2da] p-6 rounded shadow-sm flex flex-col gap-5">
            <div className="font-mono text-xs text-sumi-muted uppercase tracking-wider font-semibold border-b border-[#e5e2da] pb-2">
              01 // ROLE TRACK &amp; BLUEPRINT SCOPE
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 font-mono text-xs">
              <div className="flex flex-col gap-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                  Target Role Track
                </label>
                {loadingTracks ? (
                  <div className="h-10 bg-[#eae8e4] rounded animate-pulse"></div>
                ) : tracks.length === 0 ? (
                  <div className="p-2.5 bg-amber-50 border border-amber-300 rounded text-amber-900 text-[11px]">
                    No tracks found. <Link href="/tracks" className="underline font-bold">Create a track first</Link>.
                  </div>
                ) : (
                  <select
                    value={selectedTrackId}
                    onChange={(e) => setSelectedTrackId(e.target.value)}
                    className="px-3.5 py-2.5 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-sm"
                  >
                    {tracks.map((t) => (
                      <option key={t.id} value={t.id}>
                        {t.title}
                      </option>
                    ))}
                  </select>
                )}
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                  Blueprint Title
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Distributed Consensus & Invariant Verification"
                  value={blueprintTitle}
                  onChange={(e) => setBlueprintTitle(e.target.value)}
                  className="px-3.5 py-2 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-sm"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                  Duration (Minutes)
                </label>
                <input
                  type="number"
                  min={10}
                  max={120}
                  value={targetDuration}
                  onChange={(e) => setTargetDuration(parseInt(e.target.value) || 30)}
                  className="px-3.5 py-2 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-sm"
                />
              </div>

              <div className="flex flex-col gap-1.5 sm:col-span-2">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                  Interview Instructions for Assessor (Voxi)
                </label>
                <textarea
                  rows={2}
                  value={instructions}
                  onChange={(e) => setInstructions(e.target.value)}
                  className="px-3.5 py-2 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-xs leading-relaxed"
                />
              </div>
            </div>
          </section>

          {/* Module 2: Question Prompts */}
          <section className="bg-white border border-[#e5e2da] p-6 rounded shadow-sm flex flex-col gap-5">
            <div className="flex items-center justify-between border-b border-[#e5e2da] pb-2 font-mono text-xs">
              <span className="text-sumi-muted uppercase tracking-wider font-semibold">
                02 // ASSESSMENT QUESTION PROMPTS ({questions.length})
              </span>
              <button
                type="button"
                onClick={() => setIsAddQModalOpen(true)}
                className="px-3 py-1 bg-sumi text-white hover:bg-neutral-800 rounded-xs uppercase tracking-wider text-[11px] font-semibold transition-colors cursor-pointer"
              >
                + Add Question
              </button>
            </div>

            {questions.length === 0 ? (
              <div className="p-8 border border-dashed border-[#e5e2da] rounded text-center text-xs font-mono text-sumi-muted">
                No question prompts added yet. Click &quot;+ Add Question&quot; to define questions for Voxi to ask the participant.
              </div>
            ) : (
              <div className="flex flex-col gap-3">
                {questions.map((q, idx) => (
                  <div
                    key={idx}
                    className="p-4 bg-[#fbf9f5] border border-[#e5e2da] rounded flex items-start justify-between gap-4"
                  >
                    <div className="flex items-start gap-3">
                      <span className="font-mono text-xs font-bold text-sumi-muted px-2 py-1 bg-white border border-[#e5e2da] rounded">
                        #{idx + 1}
                      </span>
                      <div className="flex flex-col gap-1">
                        <p className="font-sans text-sm font-medium text-sumi">
                          {q.question_text}
                        </p>
                        {q.follow_up_hint && (
                          <p className="font-mono text-[11px] text-sumi-muted italic">
                            Hint: {q.follow_up_hint}
                          </p>
                        )}
                      </div>
                    </div>
                    <button
                      type="button"
                      onClick={() => handleRemoveQuestion(idx)}
                      className="text-red-500 hover:text-red-700 font-mono text-xs cursor-pointer"
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>
            )}
          </section>

          {/* Module 3: Rubric Evaluation Criteria */}
          <section className="bg-white border border-[#e5e2da] p-6 rounded shadow-sm flex flex-col gap-5">
            <div className="flex items-center justify-between border-b border-[#e5e2da] pb-2 font-mono text-xs">
              <span className="text-sumi-muted uppercase tracking-wider font-semibold">
                03 // EVALUATION RUBRIC CRITERIA ({rubrics.length})
              </span>
              <button
                type="button"
                onClick={() => setIsAddRubricModalOpen(true)}
                className="px-3 py-1 bg-sumi text-white hover:bg-neutral-800 rounded-xs uppercase tracking-wider text-[11px] font-semibold transition-colors cursor-pointer"
              >
                + Add Rubric
              </button>
            </div>

            {rubrics.length === 0 ? (
              <div className="p-8 border border-dashed border-[#e5e2da] rounded text-center text-xs font-mono text-sumi-muted">
                No evaluation criteria defined yet. Click &quot;+ Add Rubric&quot; to specify competencies for the asynchronous feedback worker to score.
              </div>
            ) : (
              <div className="flex flex-col gap-3">
                {rubrics.map((r, idx) => (
                  <div
                    key={idx}
                    className="p-4 bg-[#fbf9f5] border border-[#e5e2da] rounded flex items-start justify-between gap-4"
                  >
                    <div className="flex items-start gap-3">
                      <span className="font-mono text-xs font-bold text-sumi-muted px-2 py-1 bg-white border border-[#e5e2da] rounded">
                        Max {r.max_score}pts
                      </span>
                      <div className="flex flex-col gap-1">
                        <p className="font-sans text-sm font-medium text-sumi">
                          {r.name}
                        </p>
                        <p className="font-sans text-xs text-sumi-secondary">
                          {r.description}
                        </p>
                      </div>
                    </div>
                    <button
                      type="button"
                      onClick={() => handleRemoveRubric(idx)}
                      className="text-red-500 hover:text-red-700 font-mono text-xs cursor-pointer"
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>
            )}
          </section>
        </main>
      </div>

      {/* Modal: Add Question */}
      {isAddQModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white border border-[#e5e2da] rounded-sm max-w-lg w-full p-6 shadow-2xl animate-in fade-in zoom-in-95">
            <div className="flex justify-between items-center pb-3 border-b border-[#e5e2da] mb-4">
              <h3 className="font-mono text-sm font-bold uppercase tracking-wider text-sumi">
                Add Assessment Question Prompt
              </h3>
              <button
                type="button"
                onClick={() => setIsAddQModalOpen(false)}
                className="text-sumi-muted hover:text-sumi font-mono text-sm"
              >
                ✕
              </button>
            </div>

            <form onSubmit={handleAddQuestion} className="flex flex-col gap-4 font-mono text-xs">
              <div className="flex flex-col gap-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                  Question Text
                </label>
                <textarea
                  rows={3}
                  required
                  placeholder="e.g. When scaling Raft with heavy pipelining over high-latency links, how do you prevent follower memory exhaust?"
                  value={newQText}
                  onChange={(e) => setNewQText(e.target.value)}
                  className="px-3.5 py-2.5 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-sm"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                  Follow-Up Hint / Directive
                </label>
                <input
                  type="text"
                  placeholder="e.g. Probe sliding window buffer limits and uncommitted batch surges."
                  value={newQHint}
                  onChange={(e) => setNewQHint(e.target.value)}
                  className="px-3.5 py-2 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-xs"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-[#e5e2da] mt-2">
                <button
                  type="button"
                  onClick={() => setIsAddQModalOpen(false)}
                  className="px-4 py-2 border border-[#e5e2da] hover:border-sumi bg-white rounded-xs uppercase tracking-wider cursor-pointer"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 bg-sumi text-white hover:bg-neutral-800 rounded-xs uppercase tracking-wider font-semibold shadow-xs cursor-pointer"
                >
                  Add Prompt
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: Add Rubric */}
      {isAddRubricModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white border border-[#e5e2da] rounded-sm max-w-lg w-full p-6 shadow-2xl animate-in fade-in zoom-in-95">
            <div className="flex justify-between items-center pb-3 border-b border-[#e5e2da] mb-4">
              <h3 className="font-mono text-sm font-bold uppercase tracking-wider text-sumi">
                Add Rubric Evaluation Criterion
              </h3>
              <button
                type="button"
                onClick={() => setIsAddRubricModalOpen(false)}
                className="text-sumi-muted hover:text-sumi font-mono text-sm"
              >
                ✕
              </button>
            </div>

            <form onSubmit={handleAddRubric} className="flex flex-col gap-4 font-mono text-xs">
              <div className="flex flex-col gap-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                  Criterion Name
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Distributed Consensus & Invariant Verification"
                  value={newRubricName}
                  onChange={(e) => setNewRubricName(e.target.value)}
                  className="px-3.5 py-2.5 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-sm"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                  Evaluation Guidance / Description
                </label>
                <textarea
                  rows={2}
                  placeholder="e.g. Evaluates candidate understanding of majority quorums, fencing tokens, and split-brain resolution."
                  value={newRubricDesc}
                  onChange={(e) => setNewRubricDesc(e.target.value)}
                  className="px-3.5 py-2 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-xs"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                  Max Score
                </label>
                <input
                  type="number"
                  min={1}
                  max={100}
                  value={newRubricMax}
                  onChange={(e) => setNewRubricMax(parseInt(e.target.value) || 10)}
                  className="px-3.5 py-2 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-sm"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-[#e5e2da] mt-2">
                <button
                  type="button"
                  onClick={() => setIsAddRubricModalOpen(false)}
                  className="px-4 py-2 border border-[#e5e2da] hover:border-sumi bg-white rounded-xs uppercase tracking-wider cursor-pointer"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 bg-sumi text-white hover:bg-neutral-800 rounded-xs uppercase tracking-wider font-semibold shadow-xs cursor-pointer"
                >
                  Add Criterion
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {toastMessage && (
        <div className="fixed bottom-6 right-6 z-50 bg-sumi text-white px-4 py-2.5 rounded shadow-xl font-mono text-xs flex items-center gap-2 animate-in fade-in slide-in-from-bottom-2">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
          <span>{toastMessage}</span>
        </div>
      )}
    </div>
  );
}

export default function BlueprintBuilderPage() {
  return (
    <Suspense fallback={<div className="p-12 text-center font-mono text-xs text-sumi-muted">Loading Blueprint Studio...</div>}>
      <BlueprintBuilderContent />
    </Suspense>
  );
}
