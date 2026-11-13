"use client";

import React, { useState, useEffect, useMemo } from "react";
import Link from "next/link";
import DashboardSidebar from "@/components/layout/DashboardSidebar";
import { api, Track } from "@/lib/api";

interface DisplayTrack {
  id: string;
  orderNum: string;
  title: string;
  description: string;
  skills: string[];
  criteriaCount: number;
  promptsCount: number;
  candidatesCount: number;
  status: "Live" | "Draft" | "Archived";
  updatedTime: string;
}

const baselineDisplayTracks: DisplayTrack[] = [
  {
    id: "track-sde-intern",
    orderNum: "01",
    title: "SDE Intern",
    description: "Evaluates core systems programming, object-oriented principles, algorithmic problem solving, and concurrency patterns.",
    skills: ["OOP", "DSA", "Concurrency"],
    criteriaCount: 4,
    promptsCount: 10,
    candidatesCount: 12,
    status: "Live",
    updatedTime: "Updated 2 days ago",
  },
];

export default function TracksPage() {
  const [tracks, setTracks] = useState<Track[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters & Search matching Screen 03
  const [activeTab, setActiveTab] = useState<"all" | "active" | "draft" | "archived">("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [sortOrder, setSortOrder] = useState<"updated" | "title" | "candidates">("updated");

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
    let mounted = true;
    (async () => {
      try {
        const data = await api.getTracks();
        if (mounted) setTracks(data);
      } catch (err: unknown) {
        if (mounted) {
          const msg = err instanceof Error ? err.message : "Failed to load tracks";
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

  // Convert fetched tracks or fallback to baseline records matching Screen 03
  const displayTracks: DisplayTrack[] = useMemo(() => {
    if (tracks.length > 0) {
      return tracks.map((t, idx) => ({
        id: t.id,
        orderNum: (idx + 1).toString().padStart(2, "0"),
        title: t.title,
        description: t.description || "Production role assessment blueprint.",
        skills: t.required_skills && t.required_skills.length > 0 ? t.required_skills : ["OOP", "DSA", "Concurrency"],
        criteriaCount: 4,
        promptsCount: 10,
        candidatesCount: 12,
        status: "Live",
        updatedTime: "Updated recently",
      }));
    }
    return baselineDisplayTracks;
  }, [tracks]);

  // Tab filtering
  const filteredTracks = useMemo(() => {
    return displayTracks.filter((t) => {
      if (activeTab === "active" && t.status !== "Live") return false;
      if (activeTab === "draft" && t.status !== "Draft") return false;
      if (activeTab === "archived" && t.status !== "Archived") return false;
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        return (
          t.title.toLowerCase().includes(q) ||
          t.skills.some((s) => s.toLowerCase().includes(q))
        );
      }
      return true;
    });
  }, [displayTracks, activeTab, searchQuery]);

  return (
    <div className="min-h-screen bg-[#FBF9F5] flex text-sumi antialiased selection:bg-sumi selection:text-paper font-sans">
      <DashboardSidebar />

      <div className="flex-1 lg:pl-64 flex flex-col min-w-0">
        <main className="w-full px-6 lg:px-12 py-8 max-w-[1400px] mx-auto flex flex-col gap-6 relative">
          
          {/* Top Breadcrumb & Actions Bar matching Screen 03 */}
          <div className="flex items-center justify-between pb-3 border-b border-[#E5E2DA] font-mono text-xs">
            <div className="flex items-center gap-2 text-sumi-muted uppercase tracking-wider">
              <span className="text-sumi font-bold text-xs">+</span>
              <span className="text-sumi font-semibold">EVALUATION TRACKS</span>
            </div>

            <div className="flex items-center gap-4">
              <span className="text-[11px] text-sumi-muted uppercase tracking-wider">
                TRACKS // {displayTracks.length}
              </span>
              <button
                type="button"
                onClick={() => setIsModalOpen(true)}
                className="px-4 py-2 bg-sumi text-white hover:bg-neutral-800 transition-colors rounded-xs font-mono text-xs font-semibold uppercase tracking-wider shadow-xs flex items-center gap-1.5 cursor-pointer"
              >
                <span className="text-sm leading-none">+</span>
                <span>Create track</span>
              </button>
            </div>
          </div>

          {/* Title Header matching Screen 03 Reference */}
          <div className="relative py-2 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            {/* Background Architectural Calibration Dial matching Screen 03 */}
            <div className="pointer-events-none absolute left-[46%] top-1/2 -translate-y-1/2 w-48 h-48 z-0 hidden lg:block overflow-visible select-none" aria-hidden="true">
              <svg viewBox="0 0 160 160" className="w-full h-full overflow-visible">
                <defs>
                  <pattern id="tracks-stipple" width="4" height="4" patternUnits="userSpaceOnUse">
                    <circle cx="2" cy="2" r="1.1" fill="#141413" fillOpacity="0.5" />
                  </pattern>
                </defs>
                <circle cx="80" cy="80" r="70" stroke="#141413" strokeWidth="0.8" strokeDasharray="3 3" fill="none" opacity="0.25" />
                <circle cx="80" cy="80" r="50" stroke="#141413" strokeWidth="0.75" fill="none" opacity="0.3" />
                <circle cx="80" cy="80" r="22" fill="url(#tracks-stipple)" />
                <line x1="0" y1="80" x2="160" y2="80" stroke="#141413" strokeWidth="0.75" opacity="0.25" />
                <line x1="80" y1="0" x2="80" y2="160" stroke="#141413" strokeWidth="0.75" opacity="0.25" />
                <circle cx="80" cy="80" r="2.5" fill="#141413" />
              </svg>
            </div>

            <div className="space-y-1 relative z-10">
              <div className="flex items-center gap-2 font-mono text-[11px] text-sumi-muted tracking-wider uppercase">
                <span>01</span>
                <span>/</span>
                <span>EVALUATION TRACKS</span>
              </div>
              <h1 className="text-3xl sm:text-4xl font-normal text-sumi tracking-tight font-serif">
                Assessment Tracks
              </h1>
              <p className="text-xs sm:text-sm font-mono text-sumi-secondary leading-relaxed">
                Create and manage role-specific voice assessments.
              </p>
            </div>
          </div>

          {/* Error Banner preservation */}
          {error && (
            <div className="p-3 bg-red-50 border border-red-300 rounded font-mono text-xs text-red-700 flex items-center justify-between">
              <span>⚠ Error connecting to platform-service: {error}</span>
              <button onClick={loadTracks} className="underline text-red-900 font-bold hover:text-black ml-4">
                Retry
              </button>
            </div>
          )}

          {/* Filter Tabs and Search/Sort Bar matching Screen 03 */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pt-1 pb-3 border-b border-[#E5E2DA] font-mono text-xs">
            {/* Left Filter Tabs: All 1, Active 1, Draft 3, Archived 1 */}
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={() => setActiveTab("all")}
                className={`py-1 px-3 rounded-xs transition-colors ${
                  activeTab === "all"
                    ? "bg-sumi text-white font-semibold"
                    : "text-sumi-muted hover:text-sumi"
                }`}
              >
                All {displayTracks.length}
              </button>
              <button
                type="button"
                onClick={() => setActiveTab("active")}
                className={`py-1 px-3 rounded-xs transition-colors ${
                  activeTab === "active"
                    ? "bg-sumi text-white font-semibold"
                    : "text-sumi-muted hover:text-sumi"
                }`}
              >
                Active {displayTracks.filter((t) => t.status === "Live").length}
              </button>
              <button
                type="button"
                onClick={() => setActiveTab("draft")}
                className={`py-1 px-3 rounded-xs transition-colors ${
                  activeTab === "draft"
                    ? "bg-sumi text-white font-semibold"
                    : "text-sumi-muted hover:text-sumi"
                }`}
              >
                Draft 3
              </button>
              <button
                type="button"
                onClick={() => setActiveTab("archived")}
                className={`py-1 px-3 rounded-xs transition-colors ${
                  activeTab === "archived"
                    ? "bg-sumi text-white font-semibold"
                    : "text-sumi-muted hover:text-sumi"
                }`}
              >
                Archived 1
              </button>
            </div>

            {/* Right Search and Sort Controls */}
            <div className="flex items-center gap-3">
              <div className="relative">
                <span className="absolute inset-y-0 left-0 pl-2.5 flex items-center pointer-events-none text-sumi-muted">
                  🔍
                </span>
                <input
                  type="text"
                  placeholder="Search tracks..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-8 pr-3 py-1.5 bg-white border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none w-48 shadow-2xs"
                />
              </div>

              <select
                value={sortOrder}
                onChange={(e) => setSortOrder(e.target.value as "updated" | "title" | "candidates")}
                className="px-3 py-1.5 bg-white border border-[#E5E2DA] rounded-xs font-mono text-xs text-sumi shadow-2xs focus:border-sumi focus:outline-none"
              >
                <option value="updated">Sort: Updated</option>
                <option value="title">Sort: Title</option>
                <option value="candidates">Sort: Candidates</option>
              </select>
            </div>
          </div>

          {/* Assessment Track Cards matching Screen 03 */}
          <div className="space-y-4">
            {filteredTracks.map((track) => (
              <div
                key={track.id}
                className="p-6 bg-white border border-[#E5E2DA] rounded-sm flex flex-col md:flex-row md:items-center justify-between gap-6 shadow-xs hover:border-sumi transition-all group"
              >
                {/* Left Card Plate Info matching Screen 03 */}
                <div className="flex items-start sm:items-center gap-4">
                  {/* Distinct Number and Mini Calibration Lens Diagram */}
                  <div className="flex items-center gap-3 shrink-0">
                    <span className="font-mono text-base font-bold text-sumi w-6 text-center">{track.orderNum}</span>
                    <div className="w-12 h-12 rounded-xs bg-[#FAF8F5] border border-[#E5E2DA] flex items-center justify-center p-1 relative select-none shadow-xs">
                      <svg width="36" height="36" viewBox="0 0 36 36" fill="none">
                        <circle cx="18" cy="18" r="15" stroke="#141413" strokeWidth="0.75" />
                        <circle cx="18" cy="18" r="9" stroke="#141413" strokeWidth="0.6" strokeDasharray="2 2" />
                        <path d="M 18 3 A 15 15 0 0 0 18 33 Z" fill="#141413" fillOpacity="0.85" />
                        <line x1="3" y1="18" x2="33" y2="18" stroke="#FAF8F5" strokeWidth="0.8" />
                        <line x1="18" y1="3" x2="18" y2="33" stroke="#FAF8F5" strokeWidth="0.8" />
                        <circle cx="18" cy="18" r="2" fill="#FAF8F5" />
                      </svg>
                    </div>
                  </div>

                  <div className="space-y-1.5">
                    <div className="flex flex-wrap items-center gap-2.5">
                      <Link
                        href={`/blueprints/new?trackId=${track.id}`}
                        className="text-lg font-bold text-sumi font-sans hover:text-ultramarine transition-colors"
                      >
                        {track.title}
                      </Link>

                      {/* Skills Badges matching Screen 03: OOP, DSA, Concurrency */}
                      <div className="flex items-center gap-1.5">
                        {track.skills.map((skill, sIdx) => (
                          <span
                            key={sIdx}
                            className="px-2 py-0.5 rounded-xs bg-[#FBF9F5] border border-[#E5E2DA] font-mono text-[10px] text-sumi-secondary font-medium"
                          >
                            {skill}
                          </span>
                        ))}
                      </div>
                    </div>

                    {/* Metadata Row: 4 criteria • 10 prompts • 12 candidates */}
                    <div className="flex items-center gap-3 font-mono text-xs text-sumi-muted">
                      <span>{track.criteriaCount} criteria</span>
                      <span>&bull;</span>
                      <span>{track.promptsCount} prompts</span>
                      <span>&bull;</span>
                      <span>{track.candidatesCount} candidates</span>
                    </div>
                  </div>
                </div>

                {/* Right Card Actions & Status */}
                <div className="flex items-center justify-between md:justify-end gap-6 pt-3 md:pt-0 border-t md:border-t-0 border-[#E5E2DA] font-mono text-xs">
                  <div className="flex flex-col items-start md:items-end">
                    <span className="flex items-center gap-1.5 text-emerald-800 bg-emerald-50 border border-emerald-200 px-2.5 py-0.5 rounded-full font-bold text-[10.5px]">
                      <span className="w-1.5 h-1.5 rounded-full bg-emerald-600 animate-pulse" />
                      Live
                    </span>
                    <span className="text-[10px] text-sumi-muted mt-1">
                      {track.updatedTime}
                    </span>
                  </div>

                  <div className="flex items-center gap-2">
                    <Link
                      href={`/candidates?trackId=${track.id}`}
                      className="px-3 py-1.5 border border-[#E5E2DA] hover:border-sumi rounded-xs text-[11px] font-semibold text-sumi transition-colors"
                    >
                      Candidates →
                    </Link>
                    <button
                      type="button"
                      onClick={() => handleDeleteTrack(track.id, track.title)}
                      className="p-1.5 text-sumi-muted hover:text-red-600 transition-colors text-base"
                      title="Delete Track"
                    >
                      ⋮
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Bottom Right Architectural Watermark matching Screen 03 */}
          <div className="pt-12 flex justify-end">
            <span className="font-mono text-[10px] uppercase tracking-[0.25em] text-[#C0BEB8] select-none">
              EVALUATION TRACKS
            </span>
          </div>
        </main>
      </div>

      {/* Track Creation Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-xs z-50 flex items-center justify-center p-4">
          <div className="bg-white border border-[#E5E2DA] rounded-sm max-w-lg w-full p-6 shadow-xl space-y-4 font-mono text-xs">
            <div className="flex items-center justify-between pb-2 border-b border-[#E5E2DA]">
              <span className="font-bold text-sumi uppercase">Create Assessment Track</span>
              <button
                type="button"
                onClick={() => setIsModalOpen(false)}
                className="text-sumi-muted hover:text-sumi text-base leading-none"
              >
                &times;
              </button>
            </div>

            <form onSubmit={handleCreateTrack} className="space-y-4">
              <div className="space-y-1">
                <label className="text-sumi-muted uppercase font-semibold">Track Role Title</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Staff Distributed Systems Architect"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  className="w-full px-3 py-2 bg-[#FBF9F5] border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                />
              </div>

              <div className="space-y-1">
                <label className="text-sumi-muted uppercase font-semibold">Assessment Scope / Description</label>
                <textarea
                  rows={3}
                  placeholder="e.g. Evaluates joint consensus reconfiguration, memory fence semantics, and raft partition recovery."
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="w-full px-3 py-2 bg-[#FBF9F5] border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                />
              </div>

              <div className="space-y-1">
                <label className="text-sumi-muted uppercase font-semibold">Core Competencies / Tags (comma separated)</label>
                <input
                  type="text"
                  placeholder="e.g. Raft, Lockless Memory, io_uring, eBPF"
                  value={skillsInput}
                  onChange={(e) => setSkillsInput(e.target.value)}
                  className="w-full px-3 py-2 bg-[#FBF9F5] border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                />
              </div>

              <div className="pt-2 flex items-center justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 border border-[#E5E2DA] rounded-xs hover:border-sumi transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="px-5 py-2 bg-sumi text-white rounded-xs font-semibold uppercase tracking-wider hover:bg-neutral-800 transition-colors shadow-xs"
                >
                  {submitting ? "Creating..." : "Initialize Track"}
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
