"use client";

import React, { useState, useEffect, useMemo, Suspense } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import DashboardSidebar from "@/components/layout/DashboardSidebar";
import { api, Track, Blueprint, SessionItem } from "@/lib/api";

interface CandidateRecord {
  id: string;
  orderNumber: string;
  name: string;
  email: string;
  status: "Completed" | "In progress" | "Invited" | "Failed";
  invitedDate: string;
  startedDate: string;
  completedDate: string;
  reportId?: string;
}

const baselineCandidateRows: CandidateRecord[] = [
  {
    id: "cand-01",
    orderNumber: "01",
    name: "Priya Shah",
    email: "priya@example.com",
    status: "Completed",
    invitedDate: "Nov 10",
    startedDate: "Nov 11",
    completedDate: "Nov 11",
    reportId: "rep-01",
  },
  {
    id: "cand-02",
    orderNumber: "02",
    name: "Daniel Kim",
    email: "daniel@example.com",
    status: "In progress",
    invitedDate: "Nov 11",
    startedDate: "Nov 11",
    completedDate: "—",
  },
  {
    id: "cand-03",
    orderNumber: "03",
    name: "Maria Garcia",
    email: "maria@example.com",
    status: "Invited",
    invitedDate: "Nov 12",
    startedDate: "—",
    completedDate: "—",
  },
  {
    id: "cand-04",
    orderNumber: "04",
    name: "James Wilson",
    email: "james@example.com",
    status: "Completed",
    invitedDate: "Nov 9",
    startedDate: "Nov 10",
    completedDate: "Nov 10",
    reportId: "rep-04",
  },
  {
    id: "cand-05",
    orderNumber: "05",
    name: "Aisha Khan",
    email: "aisha@example.com",
    status: "In progress",
    invitedDate: "Nov 11",
    startedDate: "Nov 12",
    completedDate: "—",
  },
];

function CandidatesContent() {
  const searchParams = useSearchParams();
  const trackIdFilter = searchParams.get("trackId") || "all";

  const [sessions, setSessions] = useState<SessionItem[]>([]);
  const [tracks, setTracks] = useState<Track[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [selectedTrackFilter, setSelectedTrackFilter] = useState(trackIdFilter);

  // Direct Inline Invite Form
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
    let mounted = true;
    (async () => {
      try {
        const [fetchedSessions, fetchedTracks] = await Promise.all([
          api.getSessions(),
          api.getTracks(),
        ]);
        if (!mounted) return;
        setSessions(fetchedSessions);
        setTracks(fetchedTracks);
        if (fetchedTracks.length > 0 && !inviteTrackId) {
          setInviteTrackId(fetchedTracks[0].id);
        }
      } catch (err: unknown) {
        if (!mounted) return;
        const msg = err instanceof Error ? err.message : "Failed to load candidate pipeline";
        setError(msg);
      } finally {
        if (mounted) setLoading(false);
      }
    })();
    return () => {
      mounted = false;
    };
  }, [inviteTrackId]);

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

  const handleSendInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!inviteName.trim() || !inviteEmail.trim()) {
      showToast("Please enter candidate name and email.");
      return;
    }

    const effectiveTrackId = inviteTrackId || (tracks[0] ? tracks[0].id : "");
    if (!effectiveTrackId) {
      showToast("Please create an assessment track first.");
      return;
    }

    setSubmittingInvite(true);
    try {
      const participant = await api.createParticipant(effectiveTrackId, {
        full_name: inviteName.trim(),
        email: inviteEmail.trim(),
      });

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

      showToast(`Accredited invite link generated for ${participant.full_name}`);
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

  // Compute records to display (live sessions mapped, or fallback to baseline records matching Screen 05)
  const displayRows: CandidateRecord[] = useMemo(() => {
    if (sessions.length > 0) {
      return sessions.map((s, idx) => {
        let statusLabel: CandidateRecord["status"] = "Invited";
        if (s.status === "completed" || s.status === "feedback_ready" || s.status === "evaluation_ready") {
          statusLabel = "Completed";
        } else if (s.status === "in_progress") {
          statusLabel = "In progress";
        }

        const dateStr = s.created_at ? new Date(s.created_at).toLocaleDateString("en-US", { month: "short", day: "numeric" }) : "Nov 12";

        return {
          id: s.id,
          orderNumber: (idx + 1).toString().padStart(2, "0"),
          name: s.participant_name || "Elena Rostova",
          email: s.participant_email || "candidate@systems.ai",
          status: statusLabel,
          invitedDate: dateStr,
          startedDate: s.started_at ? new Date(s.started_at).toLocaleDateString("en-US", { month: "short", day: "numeric" }) : "—",
          completedDate: s.completed_at ? new Date(s.completed_at).toLocaleDateString("en-US", { month: "short", day: "numeric" }) : "—",
          reportId: s.id,
        };
      });
    }
    return baselineCandidateRows;
  }, [sessions]);

  // Filter based on status tabs
  const filteredRows = useMemo(() => {
    if (statusFilter === "all") return displayRows;
    return displayRows.filter((r) => r.status.toLowerCase() === statusFilter.toLowerCase());
  }, [displayRows, statusFilter]);

  // Pipeline metrics
  const totalCount = displayRows.length;
  const invitedCount = displayRows.filter((r) => r.status === "Invited").length;
  const inProgressCount = displayRows.filter((r) => r.status === "In progress").length;
  const completedCount = displayRows.filter((r) => r.status === "Completed").length;

  return (
    <div className="min-h-screen bg-[#FBF9F5] flex text-sumi antialiased selection:bg-sumi selection:text-paper font-sans">
      <DashboardSidebar />

      <div className="flex-1 lg:pl-64 flex flex-col min-w-0">
        <main className="w-full px-6 lg:px-12 py-8 max-w-[1400px] mx-auto flex flex-col gap-6">
          
          {/* Top Title Bar & Track Selector matching Screen 05 */}
          <div className="relative flex flex-col sm:flex-row sm:items-end justify-between gap-4 pb-4 border-b border-[#E5E2DA]">
            {/* Background Architectural Drafting Guide & Parabola Arc matching Screen 05 */}
            <div className="pointer-events-none absolute left-[45%] top-0 w-80 h-32 z-0 hidden lg:block overflow-visible select-none" aria-hidden="true">
              <svg viewBox="0 0 320 120" className="w-full h-full overflow-visible">
                <path d="M 10 15 Q 120 105 240 25" stroke="#141413" strokeWidth="0.8" strokeDasharray="3 3" fill="none" opacity="0.25" />
                <line x1="120" y1="5" x2="120" y2="115" stroke="#141413" strokeWidth="0.75" opacity="0.2" />
                <circle cx="120" cy="60" r="2" fill="#141413" opacity="0.4" />
                <circle cx="240" cy="25" r="2" fill="#141413" opacity="0.4" />
              </svg>
            </div>

            <div className="space-y-1 relative z-10">
              <div className="flex items-center gap-2 font-mono text-[11px] text-sumi-muted tracking-wider uppercase">
                <span className="text-sumi font-bold text-xs">+</span>
                <span>CANDIDATES</span>
              </div>
              <div className="flex items-start gap-3">
                <svg className="w-5 h-5 text-[#767872] shrink-0 mt-1 select-none" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <circle cx="12" cy="12" r="7" />
                  <circle cx="12" cy="12" r="2" />
                  <line x1="12" y1="2" x2="12" y2="7" />
                  <line x1="12" y1="17" x2="12" y2="22" />
                  <line x1="2" y1="12" x2="7" y2="12" />
                  <line x1="17" y1="12" x2="22" y2="12" />
                </svg>
                <h1 className="text-3xl sm:text-4xl font-normal text-sumi tracking-tight font-serif">
                  Candidates
                </h1>
              </div>
              <p className="text-xs sm:text-sm font-mono text-sumi-secondary leading-relaxed">
                Manage invitations and track assessment progress.
              </p>
            </div>

            {/* Role Track Dropdown on Top Right */}
            <div className="flex items-center gap-3">
              <select
                value={selectedTrackFilter}
                onChange={(e) => {
                  setSelectedTrackFilter(e.target.value);
                  setInviteTrackId(e.target.value);
                }}
                className="px-3.5 py-1.5 bg-white border border-[#E5E2DA] rounded-xs font-mono text-xs text-sumi shadow-xs focus:outline-none focus:border-sumi"
              >
                {tracks.length > 0 ? (
                  tracks.map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.title}
                    </option>
                  ))
                ) : (
                  <option value="all">SDE Intern</option>
                )}
              </select>
            </div>
          </div>

          {/* Error Banner preservation */}
          {error && (
            <div className="p-3 bg-red-50 border border-red-300 rounded font-mono text-xs text-red-700 flex items-center justify-between">
              <span>⚠ Error fetching candidate pipeline: {error}</span>
              <button onClick={loadData} className="underline text-red-900 font-bold hover:text-black ml-4">
                Retry
              </button>
            </div>
          )}

          {/* Filter Tabs matching Screen 05: All 12, Invited 4, In progress 3, Completed 5, Failed 0 */}
          <div className="flex items-center gap-2 border-b border-[#E5E2DA] pb-2 font-mono text-xs">
            <button
              type="button"
              onClick={() => setStatusFilter("all")}
              className={`px-3 py-1 rounded-xs transition-colors ${
                statusFilter === "all"
                  ? "bg-sumi text-white font-semibold"
                  : "text-sumi-muted hover:text-sumi"
              }`}
            >
              All {totalCount}
            </button>
            <button
              type="button"
              onClick={() => setStatusFilter("Invited")}
              className={`px-3 py-1 rounded-xs transition-colors ${
                statusFilter === "Invited"
                  ? "bg-sumi text-white font-semibold"
                  : "text-sumi-muted hover:text-sumi"
              }`}
            >
              Invited {invitedCount}
            </button>
            <button
              type="button"
              onClick={() => setStatusFilter("In progress")}
              className={`px-3 py-1 rounded-xs transition-colors ${
                statusFilter === "In progress"
                  ? "bg-sumi text-white font-semibold"
                  : "text-sumi-muted hover:text-sumi"
              }`}
            >
              In progress {inProgressCount}
            </button>
            <button
              type="button"
              onClick={() => setStatusFilter("Completed")}
              className={`px-3 py-1 rounded-xs transition-colors ${
                statusFilter === "Completed"
                  ? "bg-sumi text-white font-semibold"
                  : "text-sumi-muted hover:text-sumi"
              }`}
            >
              Completed {completedCount}
            </button>
            <button
              type="button"
              onClick={() => setStatusFilter("Failed")}
              className={`px-3 py-1 rounded-xs transition-colors ${
                statusFilter === "Failed"
                  ? "bg-sumi text-white font-semibold"
                  : "text-sumi-muted hover:text-sumi"
              }`}
            >
              Failed 0
            </button>
          </div>

          {/* Upper 2-Box Grid Matching Reference Screen 05 */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
            
            {/* Box A (7 cols): Invite Candidate Form */}
            <div className="lg:col-span-7 p-6 bg-white border border-[#E5E2DA] rounded-sm shadow-xs flex flex-col justify-between space-y-4">
              <div>
                <div className="flex items-center gap-2 font-mono text-xs font-bold text-sumi">
                  <span>Invite candidate</span>
                </div>
                <p className="font-mono text-[11px] text-sumi-secondary mt-0.5">
                  Generate a unique link for this assessment track.
                </p>
              </div>

              <form onSubmit={handleSendInvite} className="space-y-3 font-mono text-xs">
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-sumi-muted">
                    👤
                  </span>
                  <input
                    type="text"
                    required
                    placeholder="e.g. Priya Shah"
                    value={inviteName}
                    onChange={(e) => setInviteName(e.target.value)}
                    className="w-full pl-9 pr-3 py-2 bg-[#FBF9F5] border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                  />
                </div>

                <div className="flex flex-col sm:flex-row gap-3">
                  <div className="relative flex-1">
                    <span className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-sumi-muted font-mono">
                      @
                    </span>
                    <input
                      type="email"
                      required
                      placeholder="e.g. priya@example.com"
                      value={inviteEmail}
                      onChange={(e) => setInviteEmail(e.target.value)}
                      className="w-full pl-9 pr-3 py-2 bg-[#FBF9F5] border border-[#E5E2DA] rounded-xs font-sans text-xs focus:border-sumi focus:outline-none"
                    />
                  </div>

                  <button
                    type="submit"
                    disabled={submittingInvite}
                    className="px-5 py-2 bg-sumi text-white hover:bg-neutral-800 rounded-xs font-mono text-xs font-semibold uppercase tracking-wider transition-all whitespace-nowrap shadow-xs disabled:opacity-50 cursor-pointer"
                  >
                    {submittingInvite ? "Generating..." : "Generate link"}
                  </button>
                </div>
              </form>

              {/* Newly generated invite token banner */}
              {newlyCreatedLink && (
                <div className="p-3 bg-emerald-50 border border-emerald-300 rounded-xs font-mono text-xs text-emerald-900 space-y-1">
                  <div className="font-bold flex justify-between">
                    <span>Single-use link ready for {newlyCreatedLink.candidateName}:</span>
                    <button
                      type="button"
                      onClick={() => copyLink(newlyCreatedLink.link)}
                      className="underline font-bold text-emerald-800 hover:text-black cursor-pointer"
                    >
                      Copy Link
                    </button>
                  </div>
                  <div className="text-[10.5px] truncate font-mono text-emerald-700 bg-white p-1 rounded border border-emerald-200">
                    {newlyCreatedLink.link}
                  </div>
                </div>
              )}
            </div>

            {/* Box B (5 cols): Track Pipeline Metrics with Histogram Chart matching Screen 05 */}
            <div className="lg:col-span-5 p-6 bg-white border border-[#E5E2DA] rounded-sm shadow-xs flex flex-col justify-between space-y-4">
              <div>
                <div className="font-mono text-xs font-bold text-sumi">
                  Track pipeline
                </div>
                <div className="grid grid-cols-4 gap-2 mt-3 text-left font-mono">
                  <div>
                    <div className="text-2xl font-bold text-sumi">{totalCount}</div>
                    <div className="text-[10px] text-sumi-muted uppercase tracking-wider">Total</div>
                  </div>
                  <div>
                    <div className="text-2xl font-bold text-sumi">{invitedCount}</div>
                    <div className="text-[10px] text-sumi-muted uppercase tracking-wider">Invited</div>
                  </div>
                  <div>
                    <div className="text-2xl font-bold text-amber-700">{inProgressCount}</div>
                    <div className="text-[10px] text-sumi-muted uppercase tracking-wider">In progress</div>
                  </div>
                  <div>
                    <div className="text-2xl font-bold text-emerald-700">{completedCount}</div>
                    <div className="text-[10px] text-sumi-muted uppercase tracking-wider">Completed</div>
                  </div>
                </div>
              </div>

              {/* Vertical Bars Histogram Chart Matching Screen 05 */}
              <div className="pt-2 flex items-end justify-between gap-1.5 h-12 px-1">
                <div className="w-2.5 bg-[#0284C7] rounded-t-xs h-7" />
                <div className="w-2.5 bg-[#0284C7] rounded-t-xs h-3" />
                <div className="w-2.5 bg-[#0284C7] rounded-t-xs h-5" />
                <div className="w-2.5 bg-[#0284C7] rounded-t-xs h-10" />
                <div className="w-2.5 bg-[#38BDF8] rounded-t-xs h-4" />
                <div className="w-2.5 bg-[#38BDF8] rounded-t-xs h-6" />
                <div className="w-2.5 bg-[#38BDF8] rounded-t-xs h-5" />
                <div className="w-2.5 bg-[#38BDF8] rounded-t-xs h-6" />
                <div className="w-2.5 bg-[#0284C7] rounded-t-xs h-9" />
                <div className="w-2.5 bg-[#7DD3FC] rounded-t-xs h-5" />
              </div>
            </div>
          </div>

          {/* Lower Table Section Matching Screen 05 */}
          <div className="bg-white border border-[#E5E2DA] rounded-sm overflow-hidden shadow-xs">
            <div className="overflow-x-auto">
              <table className="w-full text-left font-mono text-xs">
                <thead>
                  <tr className="bg-[#F8F5EE] border-b border-[#E5E2DA] text-sumi-muted text-[10.5px] uppercase tracking-wider">
                    <th className="py-3 px-4 w-12">#</th>
                    <th className="py-3 px-4">Candidate</th>
                    <th className="py-3 px-4">Status</th>
                    <th className="py-3 px-4">Invited</th>
                    <th className="py-3 px-4">Started</th>
                    <th className="py-3 px-4">Completed</th>
                    <th className="py-3 px-4 text-right">Report</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#E5E2DA]">
                  {filteredRows.map((cand) => (
                    <tr key={cand.id} className="hover:bg-[#FBF9F5] transition-colors">
                      <td className="py-3 px-4 text-sumi-muted">{cand.orderNumber}</td>
                      <td className="py-3 px-4">
                        <div className="font-semibold text-sumi font-sans text-xs">{cand.name}</div>
                        <div className="text-[10px] text-sumi-muted font-mono">{cand.email}</div>
                      </td>
                      <td className="py-3 px-4">
                        <span
                          className={`px-2 py-0.5 rounded-full text-[10px] font-semibold border ${
                            cand.status === "Completed"
                              ? "bg-blue-50 border-blue-200 text-blue-800"
                              : cand.status === "In progress"
                              ? "bg-amber-50 border-amber-200 text-amber-800"
                              : "bg-[#F4F1EA] border-[#E5E2DA] text-sumi-secondary"
                          }`}
                        >
                          {cand.status}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-sumi-secondary text-[11px]">{cand.invitedDate}</td>
                      <td className="py-3 px-4 text-sumi-secondary text-[11px]">{cand.startedDate}</td>
                      <td className="py-3 px-4 text-sumi-secondary text-[11px]">{cand.completedDate}</td>
                      <td className="py-3 px-4 text-right">
                        {cand.reportId ? (
                          <Link
                            href={`/reports/${cand.reportId}`}
                            className="text-ultramarine font-semibold hover:underline text-[11px] inline-flex items-center gap-1"
                          >
                            <span>View</span>
                            <span>→</span>
                          </Link>
                        ) : (
                          <span className="text-sumi-muted">—</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </main>
      </div>

      {toastMessage && (
        <div className="fixed bottom-6 right-6 z-50 p-3 bg-sumi text-white font-mono text-xs rounded shadow-lg animate-in fade-in">
          {toastMessage}
        </div>
      )}
    </div>
  );
}

export default function CandidatesPage() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-[#FBF9F5] flex items-center justify-center font-mono text-xs">Loading candidates ledger...</div>}>
      <CandidatesContent />
    </Suspense>
  );
}
