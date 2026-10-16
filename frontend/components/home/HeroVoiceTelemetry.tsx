"use client";

import React, { useEffect, useRef, useState } from "react";
import Logo from "@/components/ui/Logo";

const sampleTurns = [
  {
    role: "ELENA",
    time: "00:26:40",
    text: "Tail latency climbs from 2.8ms to roughly 44ms. But our philosophy is categorical: we accept tail latency spikes in exchange for zero stale dirty reads on ledger accounts. Correctness must supersede perceived velocity in financial topologies.",
    prompt:
      "When scaling Raft with heavy pipelining over high-latency transatlantic links, how do you prevent follower memory exhaust during uncommitted batch surges?",
    topic: "Raft Pipelining & Memory Backpressure",
  },
  {
    role: "ELENA",
    time: "00:28:15",
    text: "In Raft, if a split-brain condition emerges across the trans-Atlantic WAN, monotonic fencing tokens guarantee the minority partition fails fast without corrupting state. We prioritize CP over AP strictly under Brewer's theorem.",
    prompt:
      "If the transatlantic fiber drops mid-proposal, how do you guarantee followers don't execute phantom commits before heartbeat quorum timeout?",
    topic: "WAN Partition & Fencing Epochs",
  },
  {
    role: "ELENA",
    time: "00:31:02",
    text: "We bypassed the OS filesystem page cache completely, utilizing direct io_uring with pre-allocated O_DIRECT ringbuffers for constant-time 4KB block writes under sustained heavy traffic. Zero kernel copy overhead.",
    prompt:
      "What kernel synchronization barriers did you introduce to prevent dirty page thrashing during concurrent fsync calls?",
    topic: "io_uring & Zero-Copy Storage",
  },
];

type VisualizerMode = "cymatics" | "oscilloscope" | "spectrogram";

export default function HeroVoiceTelemetry() {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const [turnIndex, setTurnIndex] = useState(0);
  const [isInterrupted, setIsInterrupted] = useState(false);
  const [dbfs, setDbfs] = useState("-14.2 dBFS");
  const [bufferWidth, setBufferWidth] = useState("80%");
  const [visualizerMode, setVisualizerMode] = useState<VisualizerMode>("cymatics");
  const [isLiveMicActive, setIsLiveMicActive] = useState(false);
  const [isPlayingAudio, setIsPlayingAudio] = useState(false);
  const [cursorCoords, setCursorCoords] = useState<{ x: number; freq: number; db: number } | null>(null);
  const [liveStreamingText, setLiveStreamingText] = useState(sampleTurns[0].text);
  const [isSimulatingSpeech, setIsSimulatingSpeech] = useState(false);
  const [modalFrequencies, setModalFrequencies] = useState({ n: 4, m: 3, hz: 1840 });

  const amplitudeMultiplierRef = useRef(1.0);
  const shockwaveRadiusRef = useRef(0);
  const modalParamsRef = useRef({ n: 4, m: 3 });
  const liveMicIntervalRef = useRef<NodeJS.Timeout | null>(null);

  // Canvas visualizer loop
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    let animationFrameId: number;
    let phase = 0;

    const render = () => {
      ctx.fillStyle = "#141413";
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      // Fine Grid drafting lines
      ctx.strokeStyle = "rgba(255, 255, 255, 0.04)";
      ctx.lineWidth = 1;
      ctx.beginPath();
      for (let x = 0; x < canvas.width; x += 32) {
        ctx.moveTo(x, 0);
        ctx.lineTo(x, canvas.height);
      }
      for (let y = 0; y < canvas.height; y += 20) {
        ctx.moveTo(0, y);
        ctx.lineTo(canvas.width, y);
      }
      ctx.stroke();

      const midY = canvas.height / 2;
      const amp = amplitudeMultiplierRef.current * (isLiveMicActive ? 1.45 : 1.0);

      if (visualizerMode === "cymatics") {
        // Harmonic Chladni Cymatics Resonant Manifold
        const { n, m } = modalParamsRef.current;
        const cols = 28;
        const rows = 12;
        const cellW = canvas.width / cols;
        const cellH = canvas.height / rows;

        ctx.lineWidth = 1.2;

        if (amp < 0.1) {
          // Zero-State Equilibrium Line (Post-interruption context_cancel)
          ctx.strokeStyle = "rgba(217, 67, 38, 0.85)";
          ctx.beginPath();
          ctx.moveTo(0, midY);
          ctx.lineTo(canvas.width, midY);
          ctx.stroke();

          ctx.fillStyle = "rgba(217, 67, 38, 0.9)";
          ctx.font = "9px 'JetBrains Mono', monospace";
          ctx.fillText("ZERO-STATE EQUILIBRIUM // CONTEXT_CANCEL TRIGGERED (0ms)", 14, midY - 8);
        } else {
          // Render fine resonant contour field
          for (let r = 0; r <= rows; r++) {
            ctx.beginPath();
            const normY = r / rows;

            for (let c = 0; c <= cols; c++) {
              const normX = c / cols;

              // Chladni 2D nodal standing wave formula
              // w(x,y) = a*sin(n*pi*x)*sin(m*pi*y) - b*cos(m*pi*x)*cos(n*pi*y)
              const w =
                Math.sin(n * Math.PI * normX + phase * 0.4) * Math.sin(m * Math.PI * normY) -
                Math.cos(m * Math.PI * normX) * Math.cos(n * Math.PI * normY - phase * 0.3);

              const ripple = isLiveMicActive ? Math.sin(normX * 30 + phase * 2) * 4 : 0;
              const px = c * cellW;
              const py = r * cellH + w * (14 * amp) + ripple;

              if (c === 0) ctx.moveTo(px, py);
              else ctx.lineTo(px, py);
            }

            const alpha = 0.2 + (r / rows) * 0.55;
            ctx.strokeStyle = isLiveMicActive
              ? `rgba(78, 222, 163, ${alpha})`
              : `rgba(16, 185, 129, ${alpha})`;
            ctx.stroke();
          }

          // Inked nodal crosshairs at resonant focal points
          ctx.fillStyle = "#34d399";
          for (let k = 1; k < n; k++) {
            const focalX = (k / n) * canvas.width;
            ctx.beginPath();
            ctx.arc(focalX, midY + Math.sin(phase + k) * 6, 2, 0, Math.PI * 2);
            ctx.fill();
          }
        }
      } else if (visualizerMode === "oscilloscope") {
        // Primary Waveform
        ctx.lineWidth = 1.5;
        ctx.strokeStyle = isLiveMicActive ? "#4edea3" : "#10b981";
        ctx.beginPath();

        for (let x = 0; x < canvas.width; x++) {
          const normX = x / canvas.width;
          const sine1 = Math.sin(normX * 18 + phase) * 22;
          const sine2 = Math.sin(normX * 36 - phase * 1.5) * 11;
          const sine3 = Math.sin(normX * 8 + phase * 0.5) * 8;
          const noise = isLiveMicActive ? (Math.random() - 0.5) * 6 : 0;
          const envelope = Math.sin(normX * Math.PI);
          const y = midY + (sine1 + sine2 + sine3 + noise) * envelope * amp;

          if (x === 0) ctx.moveTo(x, y);
          else ctx.lineTo(x, y);
        }
        ctx.stroke();

        // Harmonic trace
        ctx.strokeStyle = isLiveMicActive ? "rgba(78, 222, 163, 0.35)" : "rgba(16, 185, 129, 0.25)";
        ctx.beginPath();
        for (let x = 0; x < canvas.width; x++) {
          const normX = x / canvas.width;
          const sine = Math.cos(normX * 24 + phase * 0.8) * 14 * Math.sin(normX * Math.PI);
          const y = midY + sine * amp;
          if (x === 0) ctx.moveTo(x, y);
          else ctx.lineTo(x, y);
        }
        ctx.stroke();
      } else if (visualizerMode === "spectrogram") {
        // FFT Frequency Spectrogram Bars
        const barCount = 48;
        const barWidth = canvas.width / barCount;
        for (let i = 0; i < barCount; i++) {
          const normI = i / barCount;
          const barHeight = Math.abs(
            Math.sin(normI * 12 + phase * 1.2) * 45 +
            Math.cos(normI * 20 - phase * 0.8) * 30 +
            (isLiveMicActive ? Math.random() * 25 : 0)
          ) * Math.sin(normI * Math.PI) * (amp * 1.2);

          const grad = ctx.createLinearGradient(0, canvas.height, 0, canvas.height - barHeight);
          grad.addColorStop(0, "#059669");
          grad.addColorStop(0.6, "#10b981");
          grad.addColorStop(1, "#34d399");

          ctx.fillStyle = grad;
          ctx.fillRect(i * barWidth + 1, canvas.height - barHeight - 4, barWidth - 2, barHeight + 4);
        }
      }

      // Interruption Shockwave Pulse Animation
      if (shockwaveRadiusRef.current > 0) {
        ctx.strokeStyle = "rgba(217, 67, 38, 0.85)";
        ctx.lineWidth = 3;
        ctx.beginPath();
        ctx.arc(canvas.width / 2, canvas.height / 2, shockwaveRadiusRef.current, 0, Math.PI * 2);
        ctx.stroke();

        // Secondary harmonic echo ring
        ctx.strokeStyle = "rgba(217, 67, 38, 0.4)";
        ctx.lineWidth = 1.5;
        ctx.beginPath();
        ctx.arc(canvas.width / 2, canvas.height / 2, Math.max(0, shockwaveRadiusRef.current - 24), 0, Math.PI * 2);
        ctx.stroke();

        shockwaveRadiusRef.current += 16;
        if (shockwaveRadiusRef.current > canvas.width) {
          shockwaveRadiusRef.current = 0;
        }
      }

      // Cursor hairline probe
      if (cursorCoords) {
        ctx.strokeStyle = "rgba(217, 67, 38, 0.7)";
        ctx.lineWidth = 1;
        ctx.setLineDash([2, 2]);
        ctx.beginPath();
        ctx.moveTo(cursorCoords.x, 0);
        ctx.lineTo(cursorCoords.x, canvas.height);
        ctx.stroke();
        ctx.setLineDash([]);
      }

      phase += isLiveMicActive ? 0.065 : 0.035;
      animationFrameId = requestAnimationFrame(render);
    };

    render();

    return () => {
      cancelAnimationFrame(animationFrameId);
    };
  }, [visualizerMode, isLiveMicActive, cursorCoords]);

  // Handle canvas mouse move for frequency probe and Chladni modal shifting
  const handleCanvasMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    const normX = Math.max(0, Math.min(1, x / rect.width));
    const normY = Math.max(0, Math.min(1, y / rect.height));

    // Shift modal numbers based on mouse cadence
    const n = Math.max(2, Math.min(7, Math.round(2 + normX * 5)));
    const m = Math.max(2, Math.min(6, Math.round(2 + normY * 4)));
    modalParamsRef.current = { n, m };

    const freq = Math.round(80 + normX * 18000);
    const db = Number((-6 - (1 - Math.sin(normX * Math.PI)) * 36).toFixed(1));
    setCursorCoords({ x: (x / rect.width) * canvas.width, freq, db });
    setModalFrequencies({ n, m, hz: Math.round(n * m * 140 + 200) });
  };

  const handleCanvasMouseLeave = () => {
    setCursorCoords(null);
    modalParamsRef.current = { n: 4, m: 3 };
    setModalFrequencies({ n: 4, m: 3, hz: 1840 });
  };

  // Simulate Interruption with Shockwave Pulse
  const handleInterrupt = () => {
    amplitudeMultiplierRef.current = 0.02;
    shockwaveRadiusRef.current = 8;
    setDbfs("-48.0 dBFS");
    setBufferWidth("0%");
    setIsInterrupted(true);

    if (isPlayingAudio) {
      window.speechSynthesis?.cancel();
      setIsPlayingAudio(false);
    }

    setTimeout(() => {
      amplitudeMultiplierRef.current = 1.0;
      setDbfs("-14.2 dBFS");
      setBufferWidth("80%");
      setTimeout(() => {
        setIsInterrupted(false);
      }, 3500);
    }, 600);
  };

  // Toggle Live Microphone Simulation
  const toggleLiveMic = () => {
    if (isLiveMicActive) {
      setIsLiveMicActive(false);
      if (liveMicIntervalRef.current) clearInterval(liveMicIntervalRef.current);
      setDbfs("-14.2 dBFS");
      setIsSimulatingSpeech(false);
    } else {
      setIsLiveMicActive(true);
      setIsSimulatingSpeech(true);
      setDbfs("-9.4 dBFS");

      // Stream text simulation
      const textToStream = sampleTurns[turnIndex].text;
      const words = textToStream.split(" ");
      let currentIdx = 0;
      setLiveStreamingText("");

      if (liveMicIntervalRef.current) clearInterval(liveMicIntervalRef.current);
      liveMicIntervalRef.current = setInterval(() => {
        currentIdx += 2;
        if (currentIdx <= words.length) {
          setLiveStreamingText(words.slice(0, currentIdx).join(" "));
        } else {
          setLiveStreamingText(textToStream);
          if (liveMicIntervalRef.current) clearInterval(liveMicIntervalRef.current);
          setIsSimulatingSpeech(false);
        }
      }, 120);
    }
  };

  // Cycle Dialogue
  const handleCycleDialogue = () => {
    const nextIdx = (turnIndex + 1) % sampleTurns.length;
    setTurnIndex(nextIdx);
    setLiveStreamingText(sampleTurns[nextIdx].text);
  };

  // Audio Playback using SpeechSynthesis
  const handlePlayAudio = () => {
    if (isPlayingAudio) {
      window.speechSynthesis?.cancel();
      setIsPlayingAudio(false);
      return;
    }

    if (typeof window !== "undefined" && "speechSynthesis" in window) {
      window.speechSynthesis.cancel();
      const utterance = new SpeechSynthesisUtterance(sampleTurns[turnIndex].text);
      utterance.rate = 1.05;
      utterance.pitch = 1.0;

      utterance.onstart = () => {
        setIsPlayingAudio(true);
        amplitudeMultiplierRef.current = 1.6;
        setDbfs("-8.5 dBFS");
      };

      utterance.onend = () => {
        setIsPlayingAudio(false);
        amplitudeMultiplierRef.current = 1.0;
        setDbfs("-14.2 dBFS");
      };

      utterance.onerror = () => {
        setIsPlayingAudio(false);
        amplitudeMultiplierRef.current = 1.0;
      };

      window.speechSynthesis.speak(utterance);
    } else {
      // Fallback pulse simulation
      setIsPlayingAudio(true);
      amplitudeMultiplierRef.current = 1.5;
      setTimeout(() => {
        setIsPlayingAudio(false);
        amplitudeMultiplierRef.current = 1.0;
      }, 3000);
    }
  };

  return (
    <div
      className="bg-white border border-hairline rounded-sm shadow-monograph overflow-hidden crosshair-corner interactive-plate"
      data-purpose="hero-interactive-telemetry"
      id="pipeline-demo"
    >
      {/* Studio Header Bar */}
      <div className="bg-paper-card px-8 py-5 hairline-b flex flex-wrap items-center justify-between gap-4 text-xs font-mono">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2.5">
            <Logo className="w-4 h-4 text-sumi" size={16} />
            <span className={`h-2.5 w-2.5 rounded-full ${isLiveMicActive ? "bg-vermilion animate-ping" : "bg-emerald-500 live-pulse"}`}></span>
            <span className="font-bold text-sumi uppercase tracking-wider">
              {isLiveMicActive ? "LIVE MIC INGEST // TRANSCRIBING" : "STREAM 01: ACTIVE CONVERSATIONAL INGEST"}
            </span>
          </div>
          <span className="text-hairline">|</span>
          <span className="text-sumi-muted text-[11px]">
            OPUS-48k-STEREO • DIRECT WEBSOCKET LINK
          </span>
        </div>
        <div className="flex items-center gap-5 text-[11px]">
          <span className="text-sumi-muted uppercase tracking-wider">INTERVIEWEE:</span>
          <span className="font-semibold text-sumi bg-white px-2.5 py-1 rounded border border-hairline shadow-xs">
            Elena Rostova [UID: #ARC-88194]
          </span>
          <span className="text-sumi-muted uppercase tracking-wider">ROLE:</span>
          <span className="text-sumi font-medium">Staff Distributed Systems Architect</span>
        </div>
      </div>

      {/* Studio Main Deck */}
      <div className="grid grid-cols-1 lg:grid-cols-12 hairline-b">
        {/* Left Deck: Visualizer & Audio Dynamics */}
        <div className="lg:col-span-5 p-8 sm:p-10 hairline-r flex flex-col justify-between bg-paper-card/40">
          <div>
            <div className="flex items-center justify-between text-xs font-mono text-sumi-muted mb-4">
              <div className="flex items-center gap-2">
                <span className="tracking-widest uppercase">FIG 01.1 // ACOUSTIC SPECTRUM DYNAMICS</span>
                {cursorCoords && (
                  <span className="text-[10px] text-vermilion font-bold bg-white px-1.5 py-0.5 rounded border border-hairline">
                    {cursorCoords.freq}Hz • {cursorCoords.db}dB
                  </span>
                )}
              </div>
              <span className={`font-mono font-semibold transition-colors ${isLiveMicActive ? "text-vermilion" : "text-emerald-600"}`}>
                {dbfs}
              </span>
            </div>

            {/* Dynamic Waveform Canvas with Mode Controls */}
            <div
              className="h-40 w-full bg-[#141413] rounded p-3 flex items-center justify-center relative overflow-hidden group cursor-crosshair"
              data-purpose="canvas-container"
              onMouseMove={handleCanvasMouseMove}
              onMouseLeave={handleCanvasMouseLeave}
            >
              <canvas
                ref={canvasRef}
                className="w-full h-full block"
                height={140}
                width={520}
              />
              <div className="absolute top-2 left-3 text-[9px] font-mono text-neutral-400 flex items-center gap-2">
                <span>CH-A: {visualizerMode.toUpperCase()}</span>
                {isLiveMicActive && (
                  <span className="text-vermilion font-bold animate-pulse">● MIC INGEST ACTIVE</span>
                )}
              </div>
              <div className="absolute bottom-2 right-3 text-[10px] font-mono text-emerald-400/90 tracking-wider">
                48000Hz • 24-bit PCM • LATENCY 12ms
              </div>
            </div>

            {/* Mode Switcher Buttons */}
            <div className="mt-3 flex items-center justify-between font-mono text-[10px]">
              <div className="flex items-center gap-1.5 text-sumi-muted">
                <span className="uppercase tracking-wider">Mode:</span>
                {visualizerMode === "cymatics" && (
                  <span className="text-emerald-700 font-semibold bg-emerald-50 px-1.5 py-0.5 rounded border border-emerald-200">
                    Chladni (n={modalFrequencies.n}, m={modalFrequencies.m}) • {modalFrequencies.hz}Hz
                  </span>
                )}
              </div>
              <div className="flex items-center gap-1 bg-paper-card p-1 rounded border border-hairline">
                <button
                  type="button"
                  onClick={() => setVisualizerMode("cymatics")}
                  className={`px-2 py-0.5 rounded cursor-pointer transition-colors ${visualizerMode === "cymatics" ? "bg-sumi text-white font-bold" : "text-sumi-muted hover:text-sumi"}`}
                >
                  Cymatics
                </button>
                <button
                  type="button"
                  onClick={() => setVisualizerMode("oscilloscope")}
                  className={`px-2 py-0.5 rounded cursor-pointer transition-colors ${visualizerMode === "oscilloscope" ? "bg-sumi text-white font-bold" : "text-sumi-muted hover:text-sumi"}`}
                >
                  Oscilloscope
                </button>
                <button
                  type="button"
                  onClick={() => setVisualizerMode("spectrogram")}
                  className={`px-2 py-0.5 rounded cursor-pointer transition-colors ${visualizerMode === "spectrogram" ? "bg-sumi text-white font-bold" : "text-sumi-muted hover:text-sumi"}`}
                >
                  FFT
                </button>
              </div>
            </div>
          </div>

          {/* Live VU and Pipelining States */}
          <div className="mt-6 space-y-5 font-mono text-xs">
            <div className="space-y-2">
              <div className="flex justify-between text-[11px] text-sumi-muted">
                <span className="tracking-wider">SENTENCE BOUNDARY TTS BUFFER</span>
                <span className="text-sumi font-semibold">
                  {isPlayingAudio ? "STREAMING PLAYBACK CHUNKS" : "BUFFER HYDRATED (1st delimiter hit)"}
                </span>
              </div>
              <div className="w-full bg-paper-subtle h-2 rounded-full overflow-hidden border border-hairline relative">
                <div
                  className={`h-full transition-all duration-300 ${isPlayingAudio ? "bg-ultramarine shimmer-active" : "bg-emerald-600"}`}
                  style={{ width: bufferWidth }}
                />
              </div>
            </div>

            {/* Interactive Simulation Action Controls */}
            <div className="pt-2 grid grid-cols-2 gap-2.5">
              <button
                className={`py-2.5 px-3 rounded text-xs font-mono font-medium transition-all flex items-center justify-center gap-2 cursor-pointer border ${
                  isLiveMicActive
                    ? "bg-vermilion text-white border-vermilion shadow-sm"
                    : "bg-white border-hairline hover:border-sumi text-sumi hover:bg-paper-card"
                }`}
                onClick={toggleLiveMic}
                type="button"
              >
                <span className={`w-2 h-2 rounded-full ${isLiveMicActive ? "bg-white animate-ping" : "bg-vermilion"}`}></span>
                <span>{isLiveMicActive ? "Stop Mic" : "🎙️ Test Mic Ingest"}</span>
              </button>

              <button
                className={`py-2.5 px-3 rounded text-xs font-mono font-medium transition-all flex items-center justify-center gap-2 cursor-pointer border ${
                  isPlayingAudio
                    ? "bg-ultramarine text-white border-ultramarine shadow-sm"
                    : "bg-white border-hairline hover:border-sumi text-sumi hover:bg-paper-card"
                }`}
                onClick={handlePlayAudio}
                type="button"
              >
                <span>{isPlayingAudio ? "⏹ Stop Playback" : "▶ Audio Playback"}</span>
              </button>

              <button
                className="col-span-2 py-2.5 px-3 bg-white border border-vermilion/50 text-vermilion hover:bg-vermilion hover:text-white rounded text-xs font-mono font-medium transition-colors flex items-center justify-center gap-2 cursor-pointer"
                onClick={handleInterrupt}
                type="button"
              >
                <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                  />
                </svg>
                <span>Simulate Interruption (context_cancel)</span>
              </button>
            </div>
          </div>
        </div>

        {/* Right Deck: Real-time Transcript & Streaming Events */}
        <div className="lg:col-span-7 p-8 sm:p-10 flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between text-xs font-mono text-sumi-muted mb-4">
              <span className="uppercase tracking-widest">
                FIG 01.2 // STREAMING WEBSOCKET TRANSCRIPT (DEEPGRAM NOVA-2)
              </span>
              <div className="flex items-center gap-3">
                <span className="inline-flex items-center gap-1.5 text-[11px] text-emerald-600 font-medium">
                  <span className="w-1.5 h-1.5 bg-emerald-500 rounded-full"></span>
                  Confidence: 99.4%
                </span>
                <button
                  type="button"
                  onClick={handleCycleDialogue}
                  className="px-2.5 py-1 bg-white hover:bg-paper-subtle border border-hairline hover:border-sumi rounded text-[11px] font-mono text-sumi transition-colors cursor-pointer"
                >
                  Next Scenario →
                </button>
              </div>
            </div>

            {/* Live Transcript Box */}
            <div className="p-6 bg-paper-card border border-hairline rounded font-mono text-xs space-y-4 min-h-[190px] relative">
              {/* Scenario Topic Badge */}
              <div className="flex items-center justify-between text-[10px] text-sumi-muted pb-2 hairline-b">
                <span>SCENARIO TOPIC: <strong className="text-sumi">{sampleTurns[turnIndex].topic}</strong></span>
                <span className="text-emerald-700 font-semibold">TURN #{turnIndex + 14}</span>
              </div>

              {/* AI Evaluator Prompt */}
              <div className="text-sumi-muted text-[11px] leading-relaxed">
                <span className="font-bold text-sumi tracking-wider">[{sampleTurns[turnIndex].time}] EVALUATOR:</span>{" "}
                &quot;{sampleTurns[turnIndex].prompt}&quot;
              </div>

              {/* Candidate Response */}
              <div className="text-sumi text-xs leading-relaxed border-l-2 border-sumi pl-4 py-1.5 bg-white/70 rounded-r">
                <span className="font-bold text-emerald-600 tracking-wider">[{sampleTurns[turnIndex].time}] {sampleTurns[turnIndex].role}:</span>{" "}
                <span className="text-sumi font-medium">{liveStreamingText}</span>
                <span className="inline-block w-1.5 h-3.5 bg-sumi ml-1 animate-pulse align-middle"></span>
              </div>
            </div>

            {/* Event Log Pill Feed with dynamic statuses */}
            <div className="mt-4 flex flex-wrap gap-2 text-[10px] font-mono">
              <span className={`px-2.5 py-1 rounded border transition-colors ${isLiveMicActive ? "bg-emerald-50 border-emerald-300 text-emerald-800 font-semibold" : "bg-paper-subtle border-hairline text-sumi-secondary"}`}>
                EVENT: {isLiveMicActive ? "streaming_audio_frame_250ms" : "sentence_boundary_detected"}
              </span>
              <span className="px-2.5 py-1 rounded bg-emerald-50 border border-emerald-200 text-emerald-700">
                ElevenLabs Turbo v2 TTFB: 198ms
              </span>
              <span className="px-2.5 py-1 rounded bg-paper-subtle border border-hairline text-sumi-secondary">
                HNSW_COSINE_SIM: 0.941
              </span>
              <span className="px-2.5 py-1 rounded bg-paper-subtle border border-hairline text-sumi-muted">
                SKIP LOCKED: 0 wait
              </span>
            </div>
          </div>

          {/* Immediate Cancel Interrupt Feedback Banner */}
          {isInterrupted && (
            <div className="mt-6 p-4 rounded bg-amber-50 border border-amber-300 text-amber-900 text-[11px] font-mono flex items-center justify-between transition-all shadow-sm">
              <div className="flex items-center gap-2.5">
                <span className="h-2 w-2 rounded-full bg-amber-600 animate-ping"></span>
                <span>
                  <strong>INTERRUPTION DETECTED:</strong> Dispatched <code>context_cancel</code> signal in 12ms. Audio playback truncated. WebSocket turn re-synchronized.
                </span>
              </div>
              <span className="text-[10px] text-amber-700 font-bold uppercase tracking-wider bg-amber-100 px-2 py-0.5 rounded border border-amber-300">
                ABORT_ACK
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Bottom Telemetry Benchmark Bar */}
      <div className="px-8 py-4 bg-[#141413] text-paper text-xs font-mono flex flex-wrap items-center justify-between gap-5">
        <div className="flex items-center gap-3">
          <span className="text-emerald-400 font-bold tracking-wider">TOTAL RTT:</span>
          <span className="text-white font-bold text-base">742ms</span>
          <span className="text-neutral-400 text-[11px]">(Budget: &lt;800ms)</span>
        </div>
        <div className="flex flex-wrap items-center gap-6 sm:gap-8 text-[11px] text-neutral-300">
          <div>
            <span className="text-neutral-400">Deepgram Nova-2 STT:</span>{" "}
            <span className="text-white font-medium ml-1">184ms</span>
          </div>
          <div>
            <span className="text-neutral-400">Groq Llama 3.3 70B TTFB:</span>{" "}
            <span className="text-white font-medium ml-1">210ms</span>
          </div>
          <div>
            <span className="text-neutral-400">ElevenLabs Turbo v2:</span>{" "}
            <span className="text-white font-medium ml-1">198ms</span>
          </div>
          <div>
            <span className="text-neutral-400">Network Transit / Jitter:</span>{" "}
            <span className="text-emerald-400 font-medium ml-1">150ms ±1.2ms</span>
          </div>
        </div>
      </div>
    </div>
  );
}
