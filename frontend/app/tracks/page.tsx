"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import DashboardSidebar from "@/components/layout/DashboardSidebar";
import { api, Track } from "@/lib/api";

export default function TracksPage() {
  const [tracks, setTracks] = useState<Track[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Track Creation Modal
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [skillsInput, setSkillsInput] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3500);
  };

  const loadTracks = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.getTracks();
      setTracks(data);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to load tracks";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTracks();
  }, []);

  const handleCreateTrack = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;

    setSubmitting(true);
    try {
      const skills = skillsInput
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean);

      const created = await api.createTrack({
        title: title.trim(),
        description: description.trim(),
        required_skills: skills.length > 0 ? skills : ["General Systems Architecture"],
      });

      showToast(`Track '${created.title}' created successfully.`);
      setTitle("");
      setDescription("");
      setSkillsInput("");
      setIsModalOpen(false);
      loadTracks();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to create track";
      showToast(`Error: ${msg}`);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteTrack = async (id: string, trackTitle: string) => {
    if (!confirm(`Are you sure you want to delete track '${trackTitle}'?`)) return;
    try {
      await api.deleteTrack(id);
      showToast(`Track '${trackTitle}' deleted.`);
      loadTracks();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to delete track";
      showToast(`Error: ${msg}`);
    }
  };

  return (
    <div className="min-h-screen bg-[#fbf9f5] flex text-sumi antialiased selection:bg-sumi selection:text-paper font-sans">
      <DashboardSidebar />

      <div className="flex-1 lg:pl-64 flex flex-col min-w-0">
        <header className="sticky top-0 h-20 bg-[#fbf9f5]/90 backdrop-blur-md z-40 flex items-center justify-between px-6 lg:px-12 border-b border-[#e5e2da]">
          <div className="flex items-center gap-4">
            <span className="font-mono text-xs uppercase tracking-wider text-sumi font-bold">
              Evaluation Tracks
            </span>
          </div>

          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => setIsModalOpen(true)}
              className="text-xs font-mono bg-sumi text-white hover:bg-neutral-800 transition-colors px-4 py-2 rounded-xs font-semibold shadow-xs flex items-center gap-1.5 cursor-pointer"
            >
              <span className="text-emerald-400 font-bold text-sm">+</span>
              <span>New Role Track</span>
            </button>
          </div>
        </header>

        <main className="w-full px-6 lg:px-12 py-8 max-w-6xl mx-auto flex flex-col pb-24 gap-8">
          <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 pb-2 border-b border-[#e5e2da]">
            <div>
              <div className="flex items-center gap-2 font-mono text-xs text-sumi-muted uppercase tracking-wider mb-1">
                <span>Rubric Specifications</span>
                <span className="text-[#c7c7c0]">/</span>
                <span className="text-ultramarine font-semibold">Active Mesh</span>
              </div>
              <h1 className="font-sans text-2xl sm:text-3xl font-medium text-sumi tracking-tight">
                Evaluation Role Tracks &amp; Blueprints
              </h1>
              <p className="font-sans text-sm text-sumi-secondary font-light mt-1">
                Live assessment roles configured directly in PostgreSQL and wired to the autonomous evaluation runtime.
              </p>
            </div>

            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={loadTracks}
                className="px-3.5 py-2 bg-white border border-[#e5e2da] hover:border-sumi rounded-xs font-mono text-xs uppercase tracking-wider text-sumi transition-colors shadow-xs"
              >
                ↻ Refresh
              </button>
              <Link
                href="/blueprints/new"
                className="px-4 py-2 bg-sumi text-white hover:bg-neutral-800 rounded-xs font-mono text-xs uppercase tracking-wider transition-colors shadow-xs flex items-center gap-2"
              >
                <span>Blueprint Studio</span>
                <span>→</span>
              </Link>
            </div>
          </div>

          {error && (
            <div className="p-4 bg-red-50 border border-red-300 rounded font-mono text-xs text-red-800 flex items-center justify-between">
              <span>Error connecting to platform-service: {error}</span>
              <button
                onClick={loadTracks}
                className="underline uppercase font-bold text-[11px] ml-4 cursor-pointer"
              >
                Retry
              </button>
            </div>
          )}

          {loading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {[1, 2].map((i) => (
                <div
                  key={i}
                  className="bg-white border border-[#e5e2da] p-6 rounded shadow-sm animate-pulse flex flex-col gap-4"
                >
                  <div className="h-4 bg-[#eae8e4] rounded w-1/3"></div>
                  <div className="h-6 bg-[#eae8e4] rounded w-3/4"></div>
                  <div className="h-12 bg-[#eae8e4] rounded w-full"></div>
                </div>
              ))}
            </div>
          ) : tracks.length === 0 ? (
            <div className="p-12 bg-white border border-dashed border-[#d5d2c8] rounded text-center flex flex-col items-center justify-center gap-4">
              <div className="w-12 h-12 rounded-full bg-[#fbf9f5] border border-[#e5e2da] flex items-center justify-center font-mono text-sm font-bold text-sumi-muted">
                00
              </div>
              <div>
                <h3 className="font-sans text-lg font-medium text-sumi">
                  No Assessment Tracks Provisioned Yet
                </h3>
                <p className="font-sans text-xs text-sumi-secondary max-w-md mx-auto mt-1">
                  Create your first role track (e.g. Senior Distributed Systems Specialist) to begin configuring interview prompts and inviting candidates.
                </p>
              </div>
              <button
                type="button"
                onClick={() => setIsModalOpen(true)}
                className="px-5 py-2.5 bg-sumi text-white hover:bg-neutral-800 font-mono text-xs uppercase tracking-widest font-semibold rounded-xs transition-colors shadow-xs"
              >
                + Create First Role Track
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {tracks.map((t) => (
                <div
                  key={t.id}
                  className="bg-white border border-[#e5e2da] p-6 rounded shadow-sm hover:border-sumi transition-all flex flex-col justify-between gap-6"
                >
                  <div className="flex flex-col gap-3">
                    <div className="flex items-center justify-between font-mono text-xs">
                      <span className="text-sumi-muted tracking-wider font-semibold">
                        ID: {t.id.slice(0, 8)}...
                      </span>
                      <span
                        className={`px-2 py-0.5 rounded-xs font-bold uppercase text-[9px] tracking-wider flex items-center gap-1 ${
                          t.status === "archived"
                            ? "bg-neutral-100 text-neutral-800 border border-neutral-300"
                            : "bg-emerald-100 text-emerald-800 border border-emerald-300"
                        }`}
                      >
                        <span
                          className={`w-1.5 h-1.5 rounded-full ${
                            t.status === "archived" ? "bg-neutral-500" : "bg-emerald-600 animate-pulse"
                          }`}
                        ></span>
                        {t.status === "archived" ? "ARCHIVED" : "ACTIVE"}
                      </span>
                    </div>

                    <div>
                      <h2 className="font-sans text-xl font-medium text-sumi leading-snug">
                        {t.title}
                      </h2>
                      <p className="font-sans text-xs text-sumi-secondary mt-1 line-clamp-2">
                        {t.description || "Structured technical interview track."}
                      </p>
                    </div>

                    {t.required_skills && t.required_skills.length > 0 && (
                      <div className="flex flex-wrap gap-1.5 pt-1">
                        {t.required_skills.map((s, i) => (
                          <span
                            key={i}
                            className="px-2 py-0.5 rounded-xs bg-[#fbf9f5] border border-[#e5e2da] font-mono text-[10px] text-sumi-secondary"
                          >
                            {s}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>

                  <div className="pt-4 border-t border-[#e5e2da] flex flex-col gap-3">
                    <div className="flex items-center gap-2 pt-1 font-mono text-xs">
                      <Link
                        href={`/blueprints/new?trackId=${t.id}`}
                        className="flex-1 py-2 text-center bg-white border border-[#e5e2da] hover:border-sumi rounded-xs text-sumi uppercase tracking-wider text-[11px] font-semibold transition-colors shadow-xs"
                      >
                        Blueprint
                      </Link>
                      <Link
                        href={`/candidates?trackId=${t.id}`}
                        className="flex-1 py-2 text-center bg-sumi text-white hover:bg-neutral-800 rounded-xs uppercase tracking-wider text-[11px] font-semibold transition-colors shadow-xs"
                      >
                        Candidates
                      </Link>
                      <button
                        type="button"
                        onClick={() => {
                          const link = `${window.location.origin}/apply/${t.id}`;
                          navigator.clipboard?.writeText(link);
                          showToast("Public Apply link copied! Candidates can self-register using this URL.");
                        }}
                        className="px-2.5 py-2 text-center bg-white border border-[#e5e2da] hover:border-sumi text-sumi rounded-xs text-[11px] font-semibold transition-colors"
                        title="Copy Public Self-Registration Link"
                      >
                        🔗 Apply Link
                      </button>
                      <button
                        type="button"
                        onClick={() => handleDeleteTrack(t.id, t.title)}
                        className="px-2.5 py-2 text-center bg-white border border-red-200 hover:border-red-500 text-red-600 rounded-xs text-[11px] font-semibold transition-colors"
                        title="Delete Track"
                      >
                        ✕
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </main>
      </div>

      {/* Modal: Create Track */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white border border-[#e5e2da] rounded-sm max-w-lg w-full p-6 shadow-2xl animate-in fade-in zoom-in-95">
            <div className="flex justify-between items-center pb-3 border-b border-[#e5e2da] mb-4">
              <h3 className="font-mono text-sm font-bold uppercase tracking-wider text-sumi">
                Create New Role Track
              </h3>
              <button
                type="button"
                onClick={() => setIsModalOpen(false)}
                className="text-sumi-muted hover:text-sumi font-mono text-sm"
              >
                ✕
              </button>
            </div>

            <form onSubmit={handleCreateTrack} className="flex flex-col gap-4 font-mono text-xs">
              <div className="flex flex-col gap-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                  Track Title
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Senior Distributed Systems Specialist"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  className="px-3.5 py-2.5 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-sm"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                  Track Description
                </label>
                <textarea
                  rows={3}
                  placeholder="Describe evaluation scope, target seniority, and focus areas..."
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="px-3.5 py-2.5 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-xs"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                  Required Skills (Comma separated)
                </label>
                <input
                  type="text"
                  placeholder="Distributed Consensus, Raft, Go, Concurrency, Kafka"
                  value={skillsInput}
                  onChange={(e) => setSkillsInput(e.target.value)}
                  className="px-3.5 py-2 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-xs"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-3 border-t border-[#e5e2da] mt-2">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 border border-[#e5e2da] hover:border-sumi bg-white rounded-xs uppercase tracking-wider cursor-pointer"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="px-5 py-2 bg-sumi text-white hover:bg-neutral-800 rounded-xs uppercase tracking-wider font-semibold shadow-xs disabled:opacity-50 cursor-pointer"
                >
                  {submitting ? "Creating..." : "Provision Track"}
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
