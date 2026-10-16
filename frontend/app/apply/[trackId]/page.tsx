"use client";

import React, { useState, useEffect, Suspense } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import Logo from "@/components/ui/Logo";
import { api, Track, Blueprint } from "@/lib/api";

function CandidateApplyContent() {
  const params = useParams();
  const router = useRouter();
  const trackId = (params?.trackId as string) || "";

  const [track, setTrack] = useState<Track | null>(null);
  const [blueprints, setBlueprints] = useState<Blueprint[]>([]);
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!trackId) return;

    async function fetchTrackData() {
      setLoading(true);
      setError(null);
      try {
        const [trackRes, bpRes] = await Promise.all([
          api.getTrack(trackId),
          api.getBlueprints(trackId),
        ]);
        setTrack(trackRes);
        setBlueprints(bpRes);
      } catch (err: unknown) {
        const msg = err instanceof Error ? err.message : "Failed to load track information";
        setError(msg);
      } finally {
        setLoading(false);
      }
    }

    fetchTrackData();
  }, [trackId]);

  const handleApply = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!fullName.trim() || !email.trim()) {
      setError("Please provide both your full name and email address.");
      return;
    }

    if (blueprints.length === 0) {
      setError("No assessment blueprint is currently active for this track. Please contact the hiring team.");
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      // 1. Create candidate participant record
      const participant = await api.createParticipant(trackId, {
        full_name: fullName.trim(),
        email: email.trim(),
      });

      // 2. Generate invite for first active blueprint
      const blueprintId = blueprints[0].id;
      const invite = await api.createInvite(participant.id, {
        docket_id: blueprintId,
        expires_in_days: 7,
      });

      // 3. Immediately route candidate into their personalized interview stage
      router.push(`/interview/${invite.token}`);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Application submission failed. Please try again.";
      setError(msg);
      setSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#fbf9f5] flex flex-col justify-between text-sumi antialiased selection:bg-sumi selection:text-paper font-sans">
      <header className="h-16 px-6 lg:px-12 flex items-center justify-between border-b border-[#e5e2da]">
        <Link href="/" className="flex items-center gap-2.5">
          <Logo className="w-5 h-5 text-sumi" />
          <span className="font-mono text-xs uppercase tracking-widest font-bold text-sumi">
            VoxAssess
          </span>
        </Link>
        <div className="flex items-center gap-2 font-mono text-xs text-sumi-muted">
          <span>Application Portal</span>
        </div>
      </header>

      <main className="flex-1 flex items-center justify-center p-6">
        <div className="max-w-lg w-full bg-white border border-[#e5e2da] rounded-xs shadow-xs p-8 sm:p-10">
          {loading ? (
            <div className="py-12 text-center font-mono text-xs text-sumi-muted animate-pulse">
              Loading track specification...
            </div>
          ) : error && !track ? (
            <div className="space-y-4">
              <div className="p-3 bg-red-50 border border-red-200 text-red-700 font-mono text-xs rounded-xs">
                {error}
              </div>
              <Link
                href="/"
                className="inline-block font-mono text-xs text-sumi underline hover:text-black"
              >
                ← Return to Platform Home
              </Link>
            </div>
          ) : (
            <>
              <div className="flex items-center gap-2 font-mono text-[11px] text-sumi-muted uppercase tracking-wider mb-2">
                <span>Role Application</span>
                <span className="text-[#c7c7c0]">/</span>
                <span className="text-emerald-700 font-semibold">Self-Registration</span>
              </div>

              <h1 className="font-sans text-2xl font-medium text-sumi tracking-tight mb-2">
                {track?.title}
              </h1>
              <p className="font-sans text-xs text-sumi-secondary leading-relaxed mb-6">
                {track?.description || "Autonomous technical dialogue assessment. Complete registration to launch your session immediately."}
              </p>

              {track?.required_skills && track.required_skills.length > 0 && (
                <div className="flex flex-wrap gap-1.5 mb-6">
                  {track.required_skills.map((skill, i) => (
                    <span
                      key={i}
                      className="px-2.5 py-0.5 bg-[#f5f3ef] border border-[#e5e2da] rounded-xs font-mono text-[10px] text-sumi"
                    >
                      {skill}
                    </span>
                  ))}
                </div>
              )}

              <form onSubmit={handleApply} className="space-y-4">
                <div>
                  <label className="block font-mono text-[11px] uppercase tracking-wider text-sumi mb-1.5 font-medium">
                    Your Full Name
                  </label>
                  <input
                    type="text"
                    required
                    value={fullName}
                    onChange={(e) => setFullName(e.target.value)}
                    placeholder="e.g. Jordan Mercer"
                    className="w-full px-3.5 py-2.5 bg-[#fbf9f5] border border-[#d5d2c8] focus:border-sumi focus:outline-none rounded-xs font-mono text-xs text-sumi"
                  />
                </div>

                <div>
                  <label className="block font-mono text-[11px] uppercase tracking-wider text-sumi mb-1.5 font-medium">
                    Email Address
                  </label>
                  <input
                    type="email"
                    required
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="jordan.mercer@systems.io"
                    className="w-full px-3.5 py-2.5 bg-[#fbf9f5] border border-[#d5d2c8] focus:border-sumi focus:outline-none rounded-xs font-mono text-xs text-sumi"
                  />
                </div>

                {error && (
                  <div className="p-3 bg-red-50 border border-red-200 text-red-700 font-mono text-xs rounded-xs">
                    {error}
                  </div>
                )}

                <button
                  type="submit"
                  disabled={submitting}
                  className="w-full py-2.5 bg-sumi text-white hover:bg-neutral-800 disabled:opacity-50 transition-colors font-mono text-xs font-medium rounded-xs shadow-xs cursor-pointer flex items-center justify-center gap-2"
                >
                  {submitting ? (
                    <>
                      <span className="w-3 h-3 border border-white border-t-transparent rounded-full animate-spin"></span>
                      <span>Generating Session Gateway...</span>
                    </>
                  ) : (
                    <span>Begin Assessment Stage →</span>
                  )}
                </button>
              </form>

              <div className="mt-8 pt-6 border-t border-[#e5e2da] flex items-center justify-between text-xs font-mono text-sumi-muted">
                <span>Direct Pipeline Wire</span>
                <span className="text-emerald-600">Zero Mock Data</span>
              </div>
            </>
          )}
        </div>
      </main>

      <footer className="h-12 px-6 flex items-center justify-center border-t border-[#e5e2da] text-xs font-mono text-sumi-muted">
        <span>Autonomous Technical Assessment Infrastructure</span>
      </footer>
    </div>
  );
}

export default function CandidateApplyPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-[#fbf9f5] flex items-center justify-center font-mono text-xs text-sumi-muted">
          Loading application portal...
        </div>
      }
    >
      <CandidateApplyContent />
    </Suspense>
  );
}
