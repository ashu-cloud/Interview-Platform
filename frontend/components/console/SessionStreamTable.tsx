"use client";

import React, { useState } from "react";

interface Session {
  id: string;
  name: string;
  track: string;
  latency: string;
  protocol: string;
  confidence: number;
  state: "EVALUATION ACTIVE" | "DIALECTIC INTERCHANGE" | "FLAGGED FOR REVIEW" | "SYNTHESIS COMPLETE";
  stateColor: string;
  transcriptSample: string;
}

const initialSessions: Session[] = [
  {
    id: "#ARC-9821",
    name: "Kaelen Vance",
    track: "Distributed Consensus & Raft",
    latency: "184ms",
    protocol: "WebRTC / Opus 48k",
    confidence: 99.1,
    state: "EVALUATION ACTIVE",
    stateColor: "bg-ultramarine text-white",
    transcriptSample:
      "When evaluating Raft leadership election during a transient asymmetric network partition, the term counter strictly prevents split-brain writes via quorums...",
  },
  {
    id: "#ARC-9822",
    name: "Mina Thorne",
    track: "Rust Core Architecture",
    latency: "212ms",
    protocol: "WebRTC / Direct Well",
    confidence: 97.8,
    state: "DIALECTIC INTERCHANGE",
    stateColor: "bg-[#eae8e4] text-sumi",
    transcriptSample:
      "Pinning futures to heap memory via Box::pin allows intrusive self-referential generator states to safely traverse await boundaries without memory invalidation...",
  },
  {
    id: "#ARC-9825",
    name: "Dr. Julian Aris",
    track: "AI Inference Compilers",
    latency: "488ms",
    protocol: "Jitter Spike • Relayed",
    confidence: 92.4,
    state: "FLAGGED FOR REVIEW",
    stateColor: "bg-[#ffdcd7] text-[#93000a] border border-[#ffdad6]",
    transcriptSample:
      "Triton block pointers enable coalesced tensor load instructions across shared memory banks; however, our warp scheduler reported a 42ms stall...",
  },
  {
    id: "#ARC-9828",
    name: "Sora Takahashi",
    track: "Kernel Memory Subsystems",
    latency: "190ms",
    protocol: "WebRTC / Direct Well",
    confidence: 99.6,
    state: "SYNTHESIS COMPLETE",
    stateColor: "bg-[#e4e2de] text-sumi",
    transcriptSample:
      "Zero-copy io_uring submission ring batching minimizes context switches during ring buffer drain sequences across NUMA node boundaries...",
  },
];

export default function SessionStreamTable() {
  const [sessions] = useState<Session[]>(initialSessions);
  const [selectedSession, setSelectedSession] = useState<Session | null>(null);
  const [filterQuery, setFilterQuery] = useState("");

  const filteredSessions = sessions.filter(
    (s) =>
      s.name.toLowerCase().includes(filterQuery.toLowerCase()) ||
      s.track.toLowerCase().includes(filterQuery.toLowerCase()) ||
      s.id.toLowerCase().includes(filterQuery.toLowerCase())
  );

  return (
    <section className="flex flex-col gap-4">
      {/* Table Header Section */}
      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-3 pb-3 border-b border-[#e5e2da]">
        <div>
          <span className="font-mono text-[11px] font-bold uppercase text-sumi-muted tracking-widest">
            FIG. 02 // REAL-TIME CONVERSATIONAL INGEST
          </span>
          <h2 className="font-sans text-2xl font-medium text-sumi tracking-tight mt-1">Active Session Stream</h2>
        </div>

        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 font-mono text-xs text-sumi-secondary">
            <span className="flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-ultramarine animate-pulse"></span>
              <span>Synchronous</span>
            </span>
            <span className="text-[#c7c7c0]">•</span>
            <span>BUFFER: 16-BIT PCM</span>
          </div>

          <div className="relative">
            <input
              type="text"
              placeholder="Search stream..."
              value={filterQuery}
              onChange={(e) => setFilterQuery(e.target.value)}
              className="px-3 py-1 text-xs font-mono bg-white border border-[#e5e2da] rounded focus:border-sumi focus:outline-none transition-colors w-40 sm:w-48"
            />
          </div>
        </div>
      </div>

      {/* Editorial Table */}
      <div className="w-full overflow-x-auto bg-white rounded border border-[#e5e2da] shadow-plate">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-[#f5f3ef]/80 border-b border-[#e5e2da] text-sumi-muted font-mono text-[11px] tracking-wider uppercase">
              <th className="py-3.5 px-6 font-semibold">IDENTIFIER</th>
              <th className="py-3.5 px-4 font-semibold">DISCIPLINE TRACK</th>
              <th className="py-3.5 px-4 font-semibold">SPECTRAL LATENCY</th>
              <th className="py-3.5 px-4 font-semibold">PHONETIC CONFIDENCE</th>
              <th className="py-3.5 px-4 font-semibold">STREAM STATE</th>
              <th className="py-3.5 px-6 text-right font-semibold">INSPECT</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-[#eeeae1] font-sans text-sm">
            {filteredSessions.map((session) => {
              const isFlagged = session.state === "FLAGGED FOR REVIEW";
              return (
                <tr
                  key={session.id}
                  className="hover:bg-[#f5f3ef]/50 transition-colors group"
                >
                  {/* Candidate Name & UID */}
                  <td className="py-4 px-6">
                    <div className="flex items-center gap-3">
                      <span
                        className={`w-2 h-2 rounded-full shrink-0 ${
                          isFlagged ? "bg-vermilion animate-pulse" : "bg-ultramarine"
                        }`}
                      ></span>
                      <div>
                        <div className="font-medium text-sumi tracking-tight">{session.name}</div>
                        <div className="font-mono text-xs text-sumi-muted">{session.id}</div>
                      </div>
                    </div>
                  </td>

                  {/* Discipline Track */}
                  <td className="py-4 px-4">
                    <span className="bg-[#efeeea] px-2.5 py-1 rounded font-mono text-xs text-sumi inline-block">
                      {session.track}
                    </span>
                  </td>

                  {/* Latency & Protocol */}
                  <td className="py-4 px-4">
                    <div
                      className={`font-mono text-xs font-semibold ${
                        isFlagged ? "text-vermilion" : "text-sumi"
                      }`}
                    >
                      {session.latency}
                    </div>
                    <div
                      className={`font-mono text-[11px] ${
                        isFlagged ? "text-vermilion font-medium" : "text-sumi-muted"
                      }`}
                    >
                      {session.protocol}
                    </div>
                  </td>

                  {/* Confidence Bar */}
                  <td className="py-4 px-4">
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-xs text-sumi font-medium">
                        {session.confidence.toFixed(1)}%
                      </span>
                      <div className="w-20 bg-[#eae8e4] h-1.5 rounded overflow-hidden hidden sm:block">
                        <div
                          className={`h-full ${isFlagged ? "bg-vermilion" : "bg-sumi"}`}
                          style={{ width: `${session.confidence}%` }}
                        ></div>
                      </div>
                    </div>
                  </td>

                  {/* Stream State Badge */}
                  <td className="py-4 px-4">
                    <span
                      className={`inline-flex items-center px-2 py-0.5 rounded font-mono text-[10px] tracking-widest font-semibold uppercase ${
                        session.stateColor
                      }`}
                    >
                      {session.state}
                    </span>
                  </td>

                  {/* Inspect Stream Action */}
                  <td className="py-4 px-6 text-right">
                    <button
                      onClick={() => setSelectedSession(session)}
                      aria-label={`Inspect stream for ${session.name}`}
                      className="bg-[#f5f3ef] hover:bg-sumi hover:text-white transition-all p-2 rounded text-sumi border border-[#e5e2da] group-hover:border-sumi"
                    >
                      <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M12 2v20M17 5v14M7 8v8M2 11v2M22 11v2" strokeLinecap="round" />
                      </svg>
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {/* Live Stream Telemetry Modal / Drawer */}
      {selectedSession && (
        <div
          className="fixed inset-0 z-50 bg-sumi/40 backdrop-blur-xs flex items-center justify-center p-4 animate-in fade-in"
          onClick={() => setSelectedSession(null)}
        >
          <div
            className="bg-[#fbf9f5] border border-hairline w-full max-w-2xl rounded p-6 shadow-monograph relative flex flex-col gap-5"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Modal Header */}
            <div className="flex items-start justify-between pb-4 border-b border-[#e5e2da]">
              <div>
                <span className="font-mono text-[10px] font-bold text-sumi-muted tracking-widest uppercase">
                  ACTIVE VOICE STAGE TELEMETRY // {selectedSession.id}
                </span>
                <h3 className="font-sans text-xl font-semibold text-sumi mt-0.5">
                  {selectedSession.name} • {selectedSession.track}
                </h3>
              </div>
              <button
                onClick={() => setSelectedSession(null)}
                className="p-1 rounded text-sumi-muted hover:text-sumi transition-colors"
                aria-label="Close telemetry modal"
              >
                <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M18 6 6 18M6 6l12 12" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </button>
            </div>

            {/* Simulated Live Audio Spectrum Visualizer */}
            <div className="flex flex-col gap-2 bg-[#141413] text-[#fbf9f5] p-4 rounded">
              <div className="flex items-center justify-between text-xs font-mono text-[#a1a1aa]">
                <span className="flex items-center gap-1.5">
                  <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping"></span>
                  <span>OPUS 48KHZ STEREO INGEST</span>
                </span>
                <span>LATENCY: {selectedSession.latency}</span>
              </div>

              {/* Animated Audio Equalizer Waveform Bars */}
              <div className="h-16 flex items-center justify-between gap-1 pt-3">
                {[45, 60, 30, 85, 95, 40, 70, 55, 90, 65, 35, 80, 50, 75, 90, 60, 40, 85, 30, 95].map(
                  (val, idx) => (
                    <div
                      key={idx}
                      className="flex-1 bg-ultramarine rounded-xs transition-all duration-300"
                      style={{
                        height: `${Math.max(12, (val * (idx % 2 === 0 ? 0.9 : 1.1)) % 100)}%`,
                      }}
                    ></div>
                  )
                )}
              </div>
            </div>

            {/* Transcript Stream Sample */}
            <div className="flex flex-col gap-1.5">
              <span className="font-mono text-[10px] uppercase tracking-wider text-sumi-muted font-semibold">
                Turn Attribution Citations (AST Verified)
              </span>
              <p className="font-sans text-sm text-sumi bg-[#f5f3ef] p-3.5 rounded border border-[#e5e2da] leading-relaxed italic">
                "{selectedSession.transcriptSample}"
              </p>
            </div>

            {/* Metrics Breakdown Grid */}
            <div className="grid grid-cols-3 gap-3 pt-2 font-mono text-xs">
              <div className="bg-[#f5f3ef] p-2.5 rounded border border-[#e5e2da]">
                <div className="text-[10px] text-sumi-muted">CONFIDENCE</div>
                <div className="font-bold text-sumi text-sm mt-0.5">{selectedSession.confidence}%</div>
              </div>
              <div className="bg-[#f5f3ef] p-2.5 rounded border border-[#e5e2da]">
                <div className="text-[10px] text-sumi-muted">PROTOCOL</div>
                <div className="font-bold text-sumi text-sm mt-0.5">{selectedSession.protocol.split(" ")[0]}</div>
              </div>
              <div className="bg-[#f5f3ef] p-2.5 rounded border border-[#e5e2da]">
                <div className="text-[10px] text-sumi-muted">JURY STATUS</div>
                <div className="font-bold text-emerald-700 text-sm mt-0.5">GROUNDED</div>
              </div>
            </div>

            {/* Modal Actions */}
            <div className="flex justify-end gap-3 pt-2 border-t border-[#e5e2da]">
              <button
                onClick={() => setSelectedSession(null)}
                className="px-4 py-2 font-mono text-xs uppercase bg-[#eae8e4] text-sumi rounded hover:bg-[#e4e2de] transition-colors"
              >
                Close View
              </button>
              <a
                href="#live-stage"
                className="px-4 py-2 font-mono text-xs uppercase bg-sumi text-white rounded hover:bg-neutral-800 transition-colors flex items-center gap-1.5"
              >
                <span>Open Live Stage Monitor</span>
                <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M5 12h14M12 5l7 7-7 7" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </a>
            </div>
          </div>
        </div>
      )}
    </section>
  );
}
