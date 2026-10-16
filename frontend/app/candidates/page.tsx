"use client";

import React, { useState, useEffect, useMemo, Suspense } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import DashboardSidebar from "@/components/layout/DashboardSidebar";
import { api, Track, Blueprint, SessionItem } from "@/lib/api";

function CandidatesContent() {
  const searchParams = useSearchParams();
  const trackIdFilter = searchParams.get("trackId") || "all";

  const [sessions, setSessions] = useState<SessionItem[]>([]);
  const [tracks, setTracks] = useState<Track[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [selectedTrackFilter, setSelectedTrackFilter] = useState(trackIdFilter);

  // Invite Modal
  const [isInviteModalOpen, setIsInviteModalOpen] = useState(false);
  const [inviteName, setInviteName] = useState("");
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteTrackId, setInviteTrackId] = useState("");
  const [availableBlueprints, setAvailableBlueprints] = useState<Blueprint[]>([]);
  const [inviteBlueprintId, setInviteBlueprintId] = useState("");
  const [submittingInvite, setSubmittingInvite] = useState(false);
  const [newlyCreatedLink, setNewlyCreatedLink] = useState<{
    link: string;
    token: string;
    candidateName: string;
  } | null>(null);

  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3500);
  };

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
      const [fetchedSessions, fetchedTracks] = await Promise.all([
        api.getSessions(),
        api.getTracks(),
      ]);
      setSessions(fetchedSessions);
      setTracks(fetchedTracks);
      if (fetchedTracks.length > 0 && !inviteTrackId) {
        setInviteTrackId(fetchedTracks[0].id);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to load candidate pipeline";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  // When invite track changes, fetch its blueprints
  useEffect(() => {
    if (!inviteTrackId) return;
    async function loadBlueprints() {
      try {
        const bps = await api.getBlueprints(inviteTrackId);
        setAvailableBlueprints(bps);
        if (bps.length > 0) {
          setInviteBlueprintId(bps[0].id);
        } else {
          setInviteBlueprintId("");
        }
      } catch (err) {
        console.error("Failed to load blueprints for track:", err);
      }
    }
    loadBlueprints();
  }, [inviteTrackId]);

  const filteredSessions = useMemo(() => {
    return sessions.filter((s) => {
      if (statusFilter !== "all" && s.status !== statusFilter) return false;
      if (selectedTrackFilter !== "all" && s.track_id !== selectedTrackFilter) return false;
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        return (
          (s.participant_name && s.participant_name.toLowerCase().includes(q)) ||
          (s.participant_email && s.participant_email.toLowerCase().includes(q)) ||
          (s.track_title && s.track_title.toLowerCase().includes(q))
        );
      }
      return true;
    });
  }, [sessions, statusFilter, selectedTrackFilter, searchQuery]);

  const handleSendInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!inviteName.trim() || !inviteEmail.trim() || !inviteTrackId) return;
    if (!inviteBlueprintId) {
      showToast("Please ensure an assessment blueprint exists for this track first.");
      return;
    }

    setSubmittingInvite(true);
    try {
      // 1. Create participant in track
      const participant = await api.createParticipant(inviteTrackId, {
        full_name: inviteName.trim(),
        email: inviteEmail.trim(),
      });

      // 2. Issue invite token
      const invite = await api.createInvite(participant.id, {
        docket_id: inviteBlueprintId || "",
        expires_in_days: 7,
      });

      const fullLink = `${window.location.origin}/interview/${invite.token}`;
      setNewlyCreatedLink({
        link: fullLink,
        token: invite.token,
        candidateName: participant.full_name,
      });

      showToast(`Accredited invite issued for ${participant.full_name}`);
      setInviteName("");
      setInviteEmail("");
      loadData();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to issue invite";
      showToast(`Error: ${msg}`);
    } finally {
      setSubmittingInvite(false);
    }
  };

  const copyLink = (link: string) => {
    navigator.clipboard?.writeText(link);
    showToast("Invite link copied to clipboard!");
  };

  return (
    <div className="min-h-screen bg-[#fbf9f5] flex text-sumi antialiased selection:bg-sumi selection:text-paper font-sans">
      <DashboardSidebar />

      <div className="flex-1 lg:pl-64 flex flex-col min-w-0">
        <header className="sticky top-0 h-20 bg-[#fbf9f5]/90 backdrop-blur-md z-40 flex items-center justify-between px-6 lg:px-12 border-b border-[#e5e2da]">
          <div className="flex items-center gap-4">
            <span className="font-mono text-xs uppercase tracking-wider text-sumi font-bold">
              Candidate Pipeline
            </span>
          </div>

          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => {
                setNewlyCreatedLink(null);
                setIsInviteModalOpen(true);
              }}
              className="text-xs font-mono bg-sumi text-white hover:bg-neutral-800 transition-colors px-4 py-2 rounded-xs font-semibold shadow-xs flex items-center gap-1.5 cursor-pointer"
            >
              <span className="text-emerald-400 font-bold text-sm">+</span>
              <span>Invite Candidate</span>
            </button>
          </div>
        </header>

        <main className="w-full px-6 lg:px-12 py-8 max-w-6xl mx-auto flex flex-col pb-24 gap-8">
          <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 pb-2 border-b border-[#e5e2da]">
            <div>
              <div className="flex items-center gap-2 font-mono text-xs text-sumi-muted uppercase tracking-wider mb-1">
                <span>Evaluation Pipeline</span>
                <span className="text-[#c7c7c0]">/</span>
                <span className="text-ultramarine font-semibold">Live Audit Ledger</span>
              </div>
              <h1 className="font-sans text-2xl sm:text-3xl font-medium text-sumi tracking-tight">
                Candidate Assessment Directory
              </h1>
              <p className="font-sans text-sm text-sumi-secondary font-light mt-1">
                Verifiable candidate transcripts and asynchronous evaluation dossiers produced by Voxi.
              </p>
            </div>

            <button
              type="button"
              onClick={loadData}
              className="px-3.5 py-2 bg-white border border-[#e5e2da] hover:border-sumi rounded-xs font-mono text-xs uppercase tracking-wider text-sumi transition-colors shadow-xs"
            >
              ↻ Refresh Ledger
            </button>
          </div>

          {error && (
            <div className="p-4 bg-red-50 border border-red-300 rounded font-mono text-xs text-red-800 flex items-center justify-between">
              <span>Error fetching candidate sessions: {error}</span>
              <button
                onClick={loadData}
                className="underline uppercase font-bold text-[11px] ml-4 cursor-pointer"
              >
                Retry
              </button>
            </div>
          )}

          {/* Filters Bar */}
          <div className="flex flex-wrap items-center justify-between gap-4 p-4 bg-white border border-[#e5e2da] rounded shadow-xs font-mono text-xs">
            <div className="flex-1 min-w-[200px]">
              <input
                type="text"
                placeholder="Search candidates by name, email, or role..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full px-3 py-1.5 bg-[#fbf9f5] border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans"
              />
            </div>

            <div className="flex items-center gap-3">
              <select
                value={selectedTrackFilter}
                onChange={(e) => setSelectedTrackFilter(e.target.value)}
                className="px-3 py-1.5 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans"
              >
                <option value="all">All Role Tracks</option>
                {tracks.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.title}
                  </option>
                ))}
              </select>

              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="px-3 py-1.5 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans"
              >
                <option value="all">All Statuses</option>
                <option value="invited">Invited</option>
                <option value="in_progress">In Progress</option>
                <option value="completed">Completed</option>
                <option value="feedback_ready">Evaluation Ready</option>
              </select>
            </div>
          </div>

          {/* Candidates Table */}
          {loading ? (
            <div className="bg-white border border-[#e5e2da] rounded p-12 text-center font-mono text-xs text-sumi-muted animate-pulse">
              Querying platform-service session records...
            </div>
          ) : filteredSessions.length === 0 ? (
            <div className="p-12 bg-white border border-dashed border-[#d5d2c8] rounded text-center flex flex-col items-center justify-center gap-4">
              <div className="w-12 h-12 rounded-full bg-[#fbf9f5] border border-[#e5e2da] flex items-center justify-center font-mono text-sm font-bold text-sumi-muted">
                00
              </div>
              <div>
                <h3 className="font-sans text-lg font-medium text-sumi">
                  No Candidates In Pipeline
                </h3>
                <p className="font-sans text-xs text-sumi-secondary max-w-md mx-auto mt-1">
                  Issue an accredited assessment token to invite your first candidate to Voxi&apos;s autonomous voice stage.
                </p>
              </div>
              <button
                type="button"
                onClick={() => {
                  setNewlyCreatedLink(null);
                  setIsInviteModalOpen(true);
                }}
                className="px-5 py-2.5 bg-sumi text-white hover:bg-neutral-800 font-mono text-xs uppercase tracking-widest font-semibold rounded-xs transition-colors shadow-xs"
              >
                + Issue Candidate Invite
              </button>
            </div>
          ) : (
            <div className="bg-white border border-[#e5e2da] rounded shadow-xs overflow-hidden">
              <table className="w-full text-left font-sans text-xs">
                <thead>
                  <tr className="bg-[#fbf9f5] border-b border-[#e5e2da] font-mono text-[10px] text-sumi-muted uppercase tracking-wider">
                    <th className="py-3 px-4">Participant</th>
                    <th className="py-3 px-4">Role Track</th>
                    <th className="py-3 px-4">Status</th>
                    <th className="py-3 px-4">Turns</th>
                    <th className="py-3 px-4">Score</th>
                    <th className="py-3 px-4 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#e5e2da]">
                  {filteredSessions.map((s) => (
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
                      <td className="py-3.5 px-4 font-mono text-[11px]">
                        {s.turn_count || 0} turns
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
        </main>
      </div>

      {/* Modal: Invite Candidate */}
      {isInviteModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white border border-[#e5e2da] rounded-sm max-w-lg w-full p-6 shadow-2xl animate-in fade-in zoom-in-95">
            <div className="flex justify-between items-center pb-3 border-b border-[#e5e2da] mb-4">
              <h3 className="font-mono text-sm font-bold uppercase tracking-wider text-sumi">
                Issue Accredited Assessment Invite
              </h3>
              <button
                type="button"
                onClick={() => setIsInviteModalOpen(false)}
                className="text-sumi-muted hover:text-sumi font-mono text-sm"
              >
                ✕
              </button>
            </div>

            {newlyCreatedLink ? (
              <div className="flex flex-col gap-4 font-mono text-xs">
                <div className="p-4 bg-emerald-50 border border-emerald-300 rounded text-emerald-900 space-y-2">
                  <div className="font-bold flex items-center gap-1.5">
                    <span>✓ Single-Use Assessment Token Issued</span>
                  </div>
                  <p className="font-sans text-xs text-emerald-800">
                    Candidate: <strong>{newlyCreatedLink.candidateName}</strong>
                  </p>
                  <div className="pt-2">
                    <span className="text-[10px] uppercase font-bold text-emerald-700 block mb-1">
                      Shareable Voice Stage URL:
                    </span>
                    <div className="p-2.5 bg-white border border-emerald-400 rounded text-[11px] font-mono select-all break-all text-sumi">
                      {newlyCreatedLink.link}
                    </div>
                  </div>
                </div>

                <div className="flex items-center justify-end gap-3 pt-2">
                  <button
                    type="button"
                    onClick={() => copyLink(newlyCreatedLink.link)}
                    className="px-4 py-2 bg-emerald-600 text-white rounded-xs uppercase tracking-wider font-semibold cursor-pointer shadow-xs"
                  >
                    Copy Link
                  </button>
                  <button
                    type="button"
                    onClick={() => setNewlyCreatedLink(null)}
                    className="px-4 py-2 border border-[#e5e2da] hover:border-sumi bg-white rounded-xs uppercase tracking-wider cursor-pointer"
                  >
                    Invite Another
                  </button>
                </div>
              </div>
            ) : (
              <form onSubmit={handleSendInvite} className="flex flex-col gap-4 font-mono text-xs">
                <div className="flex flex-col gap-1.5">
                  <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                    Target Role Track
                  </label>
                  {tracks.length === 0 ? (
                    <div className="p-2.5 bg-amber-50 border border-amber-300 rounded text-amber-900 text-[11px]">
                      No tracks found. <Link href="/tracks" className="underline font-bold">Create a track first</Link>.
                    </div>
                  ) : (
                    <select
                      value={inviteTrackId}
                      onChange={(e) => setInviteTrackId(e.target.value)}
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

                {availableBlueprints.length > 0 && (
                  <div className="flex flex-col gap-1.5">
                    <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                      Assessment Blueprint
                    </label>
                    <select
                      value={inviteBlueprintId}
                      onChange={(e) => setInviteBlueprintId(e.target.value)}
                      className="px-3.5 py-2.5 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-sm"
                    >
                      {availableBlueprints.map((bp) => (
                        <option key={bp.id} value={bp.id}>
                          {bp.title} ({bp.duration_mins}m)
                        </option>
                      ))}
                    </select>
                  </div>
                )}

                <div className="flex flex-col gap-1.5">
                  <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                    Candidate Full Name
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="e.g. Elena Rostova"
                    value={inviteName}
                    onChange={(e) => setInviteName(e.target.value)}
                    className="px-3.5 py-2.5 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-sm"
                  />
                </div>

                <div className="flex flex-col gap-1.5">
                  <label className="text-sumi-muted uppercase tracking-wider font-semibold">
                    Candidate Email
                  </label>
                  <input
                    type="email"
                    required
                    placeholder="elena.rostova@infrastructure.dev"
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
                    className="px-3.5 py-2.5 bg-white border border-[#e5e2da] rounded-xs focus:border-sumi focus:outline-none text-sumi font-sans text-sm"
                  />
                </div>

                <div className="flex items-center justify-end gap-3 pt-3 border-t border-[#e5e2da] mt-2">
                  <button
                    type="button"
                    onClick={() => setIsInviteModalOpen(false)}
                    className="px-4 py-2 border border-[#e5e2da] hover:border-sumi bg-white rounded-xs uppercase tracking-wider cursor-pointer"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={submittingInvite || !inviteTrackId}
                    className="px-5 py-2 bg-sumi text-white hover:bg-neutral-800 rounded-xs uppercase tracking-wider font-semibold shadow-xs disabled:opacity-50 cursor-pointer"
                  >
                    {submittingInvite ? "Generating Token..." : "Issue Invite Link"}
                  </button>
                </div>
              </form>
            )}
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

export default function CandidatesPage() {
  return (
    <Suspense fallback={<div className="p-12 text-center font-mono text-xs text-sumi-muted">Loading Candidates...</div>}>
      <CandidatesContent />
    </Suspense>
  );
}
