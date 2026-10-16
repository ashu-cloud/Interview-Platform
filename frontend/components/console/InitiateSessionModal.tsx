"use client";

import React, { useState } from "react";

interface InitiateSessionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreated: (candidateName: string, track: string) => void;
}

export default function InitiateSessionModal({ isOpen, onClose, onCreated }: InitiateSessionModalProps) {
  const [candidateName, setCandidateName] = useState("");
  const [track, setTrack] = useState("Distributed Systems");
  const [protocol, setProtocol] = useState("WebRTC (Opus 48k)");

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!candidateName.trim()) return;
    onCreated(candidateName, track);
    onClose();
  };

  return (
    <div
      className="fixed inset-0 z-50 bg-sumi/40 backdrop-blur-xs flex items-center justify-center p-4 animate-in fade-in"
      onClick={onClose}
    >
      <div
        className="bg-[#fbf9f5] border border-hairline w-full max-w-lg rounded p-6 shadow-monograph relative flex flex-col gap-5"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-start justify-between pb-3.5 border-b border-[#e5e2da]">
          <div>
            <span className="font-mono text-[10px] font-bold text-sumi-muted tracking-widest uppercase">
              ORCHESTRATION PIPELINE // DISPATCH
            </span>
            <h3 className="font-sans text-xl font-semibold text-sumi mt-0.5">
              Initiate Synthetic Voice Stage
            </h3>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded text-sumi-muted hover:text-sumi transition-colors"
            aria-label="Close modal"
          >
            <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M18 6 6 18M6 6l12 12" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="font-mono text-xs text-sumi font-semibold tracking-wide uppercase">
              Candidate Designation / ID
            </label>
            <input
              type="text"
              required
              placeholder="e.g. Elena Rostova (#ARC-88194)"
              value={candidateName}
              onChange={(e) => setCandidateName(e.target.value)}
              className="px-3.5 py-2 text-sm font-sans bg-white border border-[#e5e2da] rounded focus:border-sumi focus:outline-none transition-colors"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="font-mono text-xs text-sumi font-semibold tracking-wide uppercase">
              Evaluation Track Rubric
            </label>
            <select
              value={track}
              onChange={(e) => setTrack(e.target.value)}
              className="px-3.5 py-2 text-sm font-sans bg-white border border-[#e5e2da] rounded focus:border-sumi focus:outline-none transition-colors"
            >
              <option value="Distributed Systems">Folio 01 // Distributed Systems (Paxos & Raft)</option>
              <option value="Rust Systems Architecture">Folio 02 // Rust Systems Architecture (Zero-Copy)</option>
              <option value="AI Infrastructure & Compilers">Folio 03 // AI Infrastructure & Compilers (Triton/CUDA)</option>
              <option value="Kernel Memory Subsystems">Folio 04 // Linux Kernel Memory Subsystems</option>
            </select>
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="font-mono text-xs text-sumi font-semibold tracking-wide uppercase">
              Acoustic Protocol
            </label>
            <select
              value={protocol}
              onChange={(e) => setProtocol(e.target.value)}
              className="px-3.5 py-2 text-sm font-sans bg-white border border-[#e5e2da] rounded focus:border-sumi focus:outline-none transition-colors"
            >
              <option value="WebRTC (Opus 48k)">WebRTC Dual-Track (Opus 48kHz, sub-800ms)</option>
              <option value="Direct WebSocket PCM">Direct Bidirectional WebSocket (16-bit PCM)</option>
            </select>
          </div>

          <div className="bg-[#f5f3ef] p-3 rounded border border-[#e5e2da] text-xs font-mono text-sumi-secondary flex items-center justify-between">
            <span>TARGET LATENCY CEILING:</span>
            <span className="font-bold text-sumi">&lt; 800 MS FIRST-TOKEN</span>
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end gap-3 pt-3 border-t border-[#e5e2da]">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 font-mono text-xs uppercase bg-[#eae8e4] text-sumi rounded hover:bg-[#e4e2de] transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-5 py-2 font-mono text-xs uppercase bg-sumi text-white rounded hover:bg-neutral-800 transition-colors flex items-center gap-2 font-semibold shadow-xs"
            >
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
              <span>Launch Live Session</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
