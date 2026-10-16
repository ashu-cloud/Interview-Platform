"use client";

import React, { useState, useEffect, useRef, Suspense, useCallback } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import Logo from "@/components/ui/Logo";
import { api, WS_BASE_URL, InviteValidation } from "@/lib/api";

function LiveCandidateStageContent() {
  const params = useParams();
  const token = typeof params?.token === "string" ? params.token : "";

  // Validation State - properly initialized to avoid cascading render warnings
  const [validation, setValidation] = useState<InviteValidation | null>(null);
  const [validating, setValidating] = useState<boolean>(Boolean(token));
  const [validationError, setValidationError] = useState<string | null>(
    token ? null : "No assessment invitation token was provided."
  );

  // Stage Entry Gate (Ensures browser audio autoplay policy is unlocked within a user gesture)
  const [hasEnteredStage, setHasEnteredStage] = useState(false);

  // Connection & Session State
  const [isConnected, setIsConnected] = useState(false);
  const [isAgentTurn, setIsAgentTurn] = useState(false);
  const [turnStatusText, setTurnStatusText] = useState("Stage ready. Click Enter Stage to begin...");
  const [turnNumber, setTurnNumber] = useState(1);
  const [liveSttText, setLiveSttText] = useState("");
  const [transcriptHistory, setTranscriptHistory] = useState<
    { speaker: "Voxi" | "You"; text: string; time: string }[]
  >([]);

  // Hardware & Mic State
  const [isMicActive, setIsMicActive] = useState(true);
  const [isUserSpeaking, setIsUserSpeaking] = useState(false);
  const [isCamActive, setIsCamActive] = useState(true);
  const [vuLevels, setVuLevels] = useState([4, 4, 4, 4, 4, 4, 4, 4]);

  // Drawer & Modals
  const [isDrawerExpanded, setIsDrawerExpanded] = useState(false);
  const [showEndModal, setShowEndModal] = useState(false);
  const [isSessionCompleted, setIsSessionCompleted] = useState(false);
  const [isRestarting, setIsRestarting] = useState<boolean>(false);

  // References
  const wsRef = useRef<WebSocket | null>(null);
  const audioContextRef = useRef<AudioContext | null>(null);
  const vadAudioCtxRef = useRef<AudioContext | null>(null);
  const activeSourcesRef = useRef<AudioBufferSourceNode[]>([]);
  const nextAudioStartTimeRef = useRef<number>(0);
  const interruptedTurnRef = useRef<number>(0);
  const micStreamRef = useRef<MediaStream | null>(null);
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const analyserRef = useRef<AnalyserNode | null>(null);
  const animFrameRef = useRef<number | null>(null);
  const isAgentTurnRef = useRef<boolean>(false);
  const isMicActiveRef = useRef<boolean>(true);
  const isUserSpeakingRef = useRef<boolean>(false);
  const turnNumberRef = useRef<number>(1);
  const hasEnteredStageRef = useRef<boolean>(false);
  const aiLiveTextRef = useRef<string>("");
  const noiseFloorRef = useRef<number>(16);

  // Keep refs synchronized with React state for Web Audio worklet/event callbacks
  useEffect(() => {
    hasEnteredStageRef.current = hasEnteredStage;
  }, [hasEnteredStage]);

  useEffect(() => {
    isAgentTurnRef.current = isAgentTurn;
  }, [isAgentTurn]);

  useEffect(() => {
    isMicActiveRef.current = isMicActive;
  }, [isMicActive]);

  useEffect(() => {
    isUserSpeakingRef.current = isUserSpeaking;
  }, [isUserSpeaking]);

  useEffect(() => {
    turnNumberRef.current = turnNumber;
  }, [turnNumber]);

  // 1. Validate Invite Token on Mount
  useEffect(() => {
    if (!token) return;
    let isCancelled = false;

    async function checkToken() {
      try {
        const res = await api.validateInvite(token);
        if (isCancelled) return;
        if (!res.valid) {
          setValidationError("This assessment token is either invalid or has expired.");
        } else {
          setValidation(res);
        }
      } catch (err: unknown) {
        if (isCancelled) return;
        const msg = err instanceof Error ? err.message : "Token validation failed";
        setValidationError(msg);
      } finally {
        if (!isCancelled) {
          setValidating(false);
        }
      }
    }

    checkToken();
    return () => {
      isCancelled = true;
    };
  }, [token]);

  // 2. Camera Video Stream
  useEffect(() => {
    if (!validation?.valid) return;
    let cameraStream: MediaStream | null = null;
    const currentVideo = videoRef.current;

    async function startCamera() {
      try {
        cameraStream = await navigator.mediaDevices.getUserMedia({
          video: true,
          audio: false,
        });
        if (videoRef.current) {
          videoRef.current.srcObject = cameraStream;
        }
      } catch (err) {
        console.warn("Camera access declined or unavailable:", err);
        setIsCamActive(false);
      }
    }
    startCamera();

    return () => {
      if (cameraStream) {
        cameraStream.getTracks().forEach((t) => t.stop());
      }
      if (currentVideo && currentVideo.srcObject) {
        (currentVideo.srcObject as MediaStream).getTracks().forEach((t) => t.stop());
      }
    };
  }, [validation]);

  // Helper: Stop all currently scheduled and playing audio buffers (Barge-In)
  const stopAllAudioPlayback = useCallback(() => {
    if (activeSourcesRef.current.length > 0) {
      activeSourcesRef.current.forEach((src) => {
        try {
          src.stop();
        } catch {
          // Ignore if already stopped
        }
      });
      activeSourcesRef.current = [];
    }
    if (audioContextRef.current) {
      nextAudioStartTimeRef.current = audioContextRef.current.currentTime;
    }
  }, []);

  // 3. Setup WebSocket & Conversational Audio Playback
  useEffect(() => {
    if (!validation?.valid || !token) return;

    // Web Audio Context for gapless chunk playback
    const AudioContextClass =
      window.AudioContext ||
      (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
    const audioCtx = new AudioContextClass();
    audioContextRef.current = audioCtx;
    nextAudioStartTimeRef.current = audioCtx.currentTime;

    const wsUrl = `${WS_BASE_URL}/ws/interview/?token=${encodeURIComponent(token)}`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setIsConnected(true);
      setTurnStatusText("Stage online. Ready to begin...");
      if (hasEnteredStageRef.current) {
        ws.send(JSON.stringify({ type: "reset_session" }));
      }
    };

    ws.onmessage = async (event) => {
      try {
        const msg = JSON.parse(event.data);

        switch (msg.type) {
          case "ai_speaking":
            setIsAgentTurn(msg.speaking);
            if (msg.speaking) {
              setLiveSttText("");
              setTurnStatusText("Voxi speaking (Speak to interrupt anytime)...");
            } else {
              setTurnStatusText("Listening to you naturally...");
            }
            break;

          case "ai_audio_chunk":
          case "audio_chunk":
            // Drop stale chunks that arrived after an interruption
            if (msg.turn && msg.turn <= interruptedTurnRef.current) {
              return;
            }

            if (msg.data && audioContextRef.current) {
              if (audioContextRef.current.state === "suspended") {
                await audioContextRef.current.resume().catch(() => {});
              }

              const binary = atob(msg.data);
              const bytes = new Uint8Array(binary.length);
              for (let i = 0; i < binary.length; i++) {
                bytes[i] = binary.charCodeAt(i);
              }
              try {
                const buffer = await audioContextRef.current.decodeAudioData(bytes.buffer.slice(0));
                const source = audioContextRef.current.createBufferSource();
                source.buffer = buffer;
                source.connect(audioContextRef.current.destination);

                activeSourcesRef.current.push(source);
                setIsAgentTurn(true);
                setLiveSttText("");
                setTurnStatusText("Voxi speaking (Speak to interrupt anytime)...");

                source.onended = () => {
                  activeSourcesRef.current = activeSourcesRef.current.filter((s) => s !== source);
                  if (activeSourcesRef.current.length === 0) {
                    setIsAgentTurn(false);
                    setTurnStatusText("Listening to you naturally...");
                  }
                };

                const startTime = Math.max(
                  audioContextRef.current.currentTime,
                  nextAudioStartTimeRef.current
                );
                source.start(startTime);
                nextAudioStartTimeRef.current = startTime + buffer.duration;
              } catch (decodeErr) {
                console.warn("Audio buffer decode error:", decodeErr);
                if (audioContextRef.current) {
                  nextAudioStartTimeRef.current = Math.max(
                    audioContextRef.current.currentTime,
                    nextAudioStartTimeRef.current
                  );
                }
              }
            }
            break;

          case "session_reset":
            stopAllAudioPlayback();
            setTranscriptHistory([]);
            setLiveSttText("");
            setTurnNumber(1);
            turnNumberRef.current = 1;
            interruptedTurnRef.current = 0;
            setIsAgentTurn(true);
            setIsRestarting(false);
            setTurnStatusText("Stage initialized. Voxi is speaking...");
            break;

          case "ai_interrupted":
            stopAllAudioPlayback();
            interruptedTurnRef.current = msg.turn || turnNumberRef.current;
            if (aiLiveTextRef.current.trim()) {
              const now = new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
              const partial = aiLiveTextRef.current.trim() + " [interrupted]";
              aiLiveTextRef.current = "";
              setTranscriptHistory((prev) => [
                ...prev,
                { speaker: "Voxi", text: partial, time: now },
              ]);
            }
            setIsAgentTurn(false);
            setTurnStatusText("Listening to your voice...");
            break;

          case "interim_transcript":
            // Suppress ghost interim updates while Voxi is speaking
            if (!isAgentTurnRef.current) {
              setLiveSttText(msg.text || "");
            }
            break;

          case "final_transcript":
            if (msg.text) {
              const now = new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
              setTranscriptHistory((prev) => [
                ...prev,
                { speaker: "You", text: msg.text, time: now },
              ]);
              setLiveSttText("");
            }
            break;

          case "ai_text_chunk":
            if (msg.text) {
              aiLiveTextRef.current += msg.text;
            }
            break;

          case "ai_text_end":
            if (aiLiveTextRef.current.trim()) {
              const now = new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
              const fullText = aiLiveTextRef.current.trim();
              aiLiveTextRef.current = "";
              setTranscriptHistory((prev) => [
                ...prev,
                { speaker: "Voxi", text: fullText, time: now },
              ]);
            }
            break;

          case "turn_completed":
            setLiveSttText("");
            setTurnNumber((prev) => {
              const next = prev + 1;
              turnNumberRef.current = next;
              return next;
            });
            break;

          case "session_completed":
            setIsSessionCompleted(true);
            setTurnStatusText("Assessment session completed.");
            break;

          case "error":
            setTurnStatusText(`Notice: ${msg.message || msg.code}`);
            break;
        }
      } catch (err) {
        console.error("Error handling incoming WebSocket message:", err);
      }
    };

    ws.onerror = () => {
      setTurnStatusText("Connection error. Ensure orchestrator is running.");
      setIsConnected(false);
    };

    ws.onclose = () => {
      setIsConnected(false);
    };

    return () => {
      ws.close();
      stopAllAudioPlayback();
      if (audioContextRef.current) {
        audioContextRef.current.close().catch(() => {});
      }
    };
  }, [validation, token, stopAllAudioPlayback]);

  // 4. Continuous Microphone Ingest with WebRTC AEC, Voice-Bandpass VAD & Free-Tier 16kHz Linear PCM Streamer
  useEffect(() => {
    if (!validation?.valid) return;

    let isSubscribed = true;
    let vadAudioCtx: AudioContext | null = null;
    let processorNode: ScriptProcessorNode | null = null;
    let micSource: MediaStreamAudioSourceNode | null = null;
    let gainNode: GainNode | null = null;

    async function initContinuousMicrophone() {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          audio: {
            echoCancellation: true,
            noiseSuppression: true,
            autoGainControl: true,
            channelCount: 1,
            sampleRate: 16000,
          },
        });

        if (!isSubscribed) {
          stream.getTracks().forEach((t) => t.stop());
          return;
        }

        micStreamRef.current = stream;

        // Web Audio Analyser for VAD & dynamic VU meters
        const AudioContextClass =
          window.AudioContext ||
          (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
        vadAudioCtx = new AudioContextClass();
        vadAudioCtxRef.current = vadAudioCtx;
        micSource = vadAudioCtx.createMediaStreamSource(stream);
        const analyser = vadAudioCtx.createAnalyser();
        analyser.fftSize = 256;
        analyser.smoothingTimeConstant = 0.3;
        micSource.connect(analyser);
        analyserRef.current = analyser;

        const dataArray = new Uint8Array(analyser.frequencyBinCount);
        let consecutiveSpeechFrames = 0;

        const updateEnergyLoop = () => {
          if (!isSubscribed) return;

          analyser.getByteFrequencyData(dataArray);

          // Human Vocal Bandpass Filter (ITU-T G.711 voice band: bins 4 to 55, ~250 Hz to 3,437 Hz)
          // Excludes <200 Hz (fan hum, desk bumps, AC rumble) and >3,500 Hz (ambient room hiss)
          let voiceSum = 0;
          const startBin = 4;
          const endBin = Math.min(55, dataArray.length);
          for (let i = startBin; i < endBin; i++) {
            voiceSum += dataArray[i];
          }
          const voiceBandAvg = voiceSum / (endBin - startBin);
          const normalized = Math.min(Math.max((voiceBandAvg / 128) * 24, 4), 28);

          // Update VU bar array with smooth jitter
          setVuLevels([
            Math.round(normalized * 0.6),
            Math.round(normalized * 0.8),
            Math.round(normalized * 1.0),
            Math.round(normalized * 1.2),
            Math.round(normalized * 1.1),
            Math.round(normalized * 0.9),
            Math.round(normalized * 0.7),
            Math.round(normalized * 0.5),
          ]);

          // Check if Voxi is actively producing audible sound
          const isAudioActivelyPlaying =
            activeSourcesRef.current.length > 0 ||
            (audioContextRef.current !== null &&
              nextAudioStartTimeRef.current > audioContextRef.current.currentTime);

          // VAD Adaptation Freeze: ONLY adapt noise floor when candidate is NOT speaking and Voxi is NOT emitting sound
          // Prevents the candidate's own voice from poisoning the threshold during long answers
          if (!isUserSpeakingRef.current && !isAudioActivelyPlaying) {
            noiseFloorRef.current = Math.min(
              Math.max(noiseFloorRef.current * 0.96 + voiceBandAvg * 0.04, 16),
              65
            );
          }

          // Elevated Adaptive Thresholds
          // Deliberate barge-in requires strong energy over speakers (threshold 45+, 18 frames)
          // Normal speaking requires voice-band energy above 30, sustained for 10 frames (~160ms) to ignore breaths/clicks
          const speechThreshold = isAudioActivelyPlaying
            ? Math.max(45, noiseFloorRef.current * 2.4)
            : Math.max(30, noiseFloorRef.current * 1.85);

          const requiredFrames = isAudioActivelyPlaying ? 18 : 10;

          if (voiceBandAvg > speechThreshold && isMicActiveRef.current) {
            consecutiveSpeechFrames++;
            if (consecutiveSpeechFrames >= requiredFrames) {
              setIsUserSpeaking(true);
            }

            // Candidate interruption detected ONLY while audio is actively playing out of speakers
            if (isAudioActivelyPlaying && consecutiveSpeechFrames >= 18) {
              stopAllAudioPlayback();
              interruptedTurnRef.current = turnNumberRef.current;
              setIsAgentTurn(false);
              setTurnStatusText("Listening to you (Interrupted)...");

              if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
                wsRef.current.send(
                  JSON.stringify({ type: "user_interrupted", turn: turnNumberRef.current })
                );
              }
            }
          } else {
            consecutiveSpeechFrames = 0;
            setIsUserSpeaking(false);
          }

          animFrameRef.current = requestAnimationFrame(updateEnergyLoop);
        };

        animFrameRef.current = requestAnimationFrame(updateEnergyLoop);

        // High-Performance 16kHz Linear PCM Streamer (Zero container headers, Free-Tier Resource Efficient)
        // Downsamples Float32 PCM to 16,000Hz mono Int16
        const bufferSize = 2048;
        processorNode = vadAudioCtx.createScriptProcessor(bufferSize, 1, 1);
        const targetSampleRate = 16000;
        const sampleRatio = vadAudioCtx.sampleRate / targetSampleRate;

        let speechHangover = 0;

        processorNode.onaudioprocess = (e) => {
          if (!isSubscribed || !wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return;
          if (!isMicActiveRef.current) return;

          const isAudioActivelyPlaying =
            activeSourcesRef.current.length > 0 ||
            (audioContextRef.current !== null &&
              nextAudioStartTimeRef.current > audioContextRef.current.currentTime);

          // Natural Ambient Post-Speech Hangover Window:
          // Keep sending ambient audio for 1.2s (~25 cycles) so Deepgram reliably detects endpointing silence
          if (isUserSpeakingRef.current) {
            speechHangover = 25;
          } else if (speechHangover > 0) {
            speechHangover--;
          }

          // FREE TIER CONSERVATION GUARD:
          // 1. If Voxi is speaking, DO NOT send mic audio to Deepgram unless candidate is deliberately interrupting
          // 2. If Voxi is quiet, only stream when user is actively speaking or in ambient hangover window
          const shouldStreamAudio = speechHangover > 0 || (!isAudioActivelyPlaying && consecutiveSpeechFrames > 0);
          if (!shouldStreamAudio) return;

          const inputData = e.inputBuffer.getChannelData(0);
          const outputLength = Math.floor(inputData.length / sampleRatio);
          const pcmData = new Int16Array(outputLength);

          // Fast linear resampler to 16kHz + Float32 to Int16 conversion
          for (let i = 0; i < outputLength; i++) {
            const index = Math.floor(i * sampleRatio);
            const s = Math.max(-1, Math.min(1, inputData[index]));
            pcmData[i] = s < 0 ? s * 0x8000 : s * 0x7FFF;
          }

          // Convert Int16Array to base64
          const bytes = new Uint8Array(pcmData.buffer);
          let binary = "";
          for (let i = 0; i < bytes.byteLength; i++) {
            binary += String.fromCharCode(bytes[i]);
          }
          const base64 = btoa(binary);

          wsRef.current.send(
            JSON.stringify({
              type: "audio_frame",
              data: base64,
              mime_type: "audio/l16;rate=16000",
            })
          );
        };

        micSource.connect(processorNode);
        gainNode = vadAudioCtx.createGain();
        gainNode.gain.value = 0;
        processorNode.connect(gainNode);
        gainNode.connect(vadAudioCtx.destination);
      } catch (err) {
        console.error("Microphone hardware initialization failed:", err);
      }
    }

    initContinuousMicrophone();

    return () => {
      isSubscribed = false;
      if (animFrameRef.current) {
        cancelAnimationFrame(animFrameRef.current);
      }
      if (processorNode) {
        processorNode.disconnect();
      }
      if (micSource) {
        micSource.disconnect();
      }
      if (gainNode) {
        gainNode.disconnect();
      }
      if (vadAudioCtx) {
        vadAudioCtx.close().catch(() => {});
      }
      if (micStreamRef.current) {
        micStreamRef.current.getTracks().forEach((t) => t.stop());
        micStreamRef.current = null;
      }
    };
  }, [validation, stopAllAudioPlayback]);

  // Toggle Microphone Mute
  const toggleMic = () => {
    setIsMicActive((prev) => {
      const next = !prev;
      if (micStreamRef.current) {
        micStreamRef.current.getAudioTracks().forEach((t) => {
          t.enabled = next;
        });
      }
      return next;
    });
  };

  // Toggle Camera Mute
  const toggleCam = () => {
    setIsCamActive((prev) => {
      const next = !prev;
      if (videoRef.current && videoRef.current.srcObject) {
        (videoRef.current.srcObject as MediaStream).getVideoTracks().forEach((t) => {
          t.enabled = next;
        });
      }
      return next;
    });
  };

  const handleEnterStage = async () => {
    // 1. Resume both AudioContext instances within this explicit user gesture
    if (audioContextRef.current && audioContextRef.current.state === "suspended") {
      try {
        await audioContextRef.current.resume();
      } catch (e) {
        console.warn("AudioContext resume failed:", e);
      }
    }
    if (vadAudioCtxRef.current && vadAudioCtxRef.current.state === "suspended") {
      try {
        await vadAudioCtxRef.current.resume();
      } catch (e) {
        console.warn("vadAudioCtx resume failed:", e);
      }
    }
    setHasEnteredStage(true);
    hasEnteredStageRef.current = true;
    // 2. Trigger fresh session reset and greeting turn
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: "reset_session" }));
    }
  };

  const handleEndSession = () => {
    stopAllAudioPlayback();
    if (micStreamRef.current) {
      micStreamRef.current.getTracks().forEach((t) => t.stop());
    }
    if (wsRef.current) {
      wsRef.current.close();
    }
    setIsSessionCompleted(true);
    setShowEndModal(false);
  };

  const handleResetSession = useCallback(() => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return;
    setIsRestarting(true);
    stopAllAudioPlayback();
    setTranscriptHistory([]);
    setLiveSttText("");
    setTurnStatusText("Resetting assessment stage...");
    wsRef.current.send(JSON.stringify({ type: "reset_session" }));
  }, [stopAllAudioPlayback]);

  if (validating) {
    return (
      <div className="min-h-screen bg-[#fbf9f5] flex items-center justify-center font-mono text-xs text-sumi">
        <div className="flex items-center gap-3">
          <span className="w-2 h-2 rounded-full bg-emerald-500 animate-ping"></span>
          <span>Verifying accredited assessment token...</span>
        </div>
      </div>
    );
  }

  if (validationError || !validation?.valid) {
    return (
      <div className="min-h-screen bg-[#fbf9f5] flex items-center justify-center p-6 text-sumi font-mono text-xs">
        <div className="max-w-md w-full bg-white border border-red-300 rounded p-8 shadow-sm space-y-4 text-center">
          <div className="w-10 h-10 rounded-full bg-red-100 border border-red-300 flex items-center justify-center mx-auto text-red-600 font-bold">
            ✕
          </div>
          <h2 className="font-sans text-lg font-medium text-sumi">
            Invalid Assessment Token
          </h2>
          <p className="font-sans text-xs text-sumi-secondary leading-relaxed">
            {validationError ||
              "This link is invalid or has already been used. Please request a new invitation from your hiring manager."}
          </p>
          <div className="pt-2">
            <Link
              href="/"
              className="inline-block px-4 py-2 bg-sumi text-white rounded-xs uppercase tracking-wider font-semibold"
            >
              Return to Home
            </Link>
          </div>
        </div>
      </div>
    );
  }

  if (isSessionCompleted) {
    return (
      <div className="min-h-screen bg-[#fbf9f5] flex items-center justify-center p-6 text-sumi font-mono text-xs">
        <div className="max-w-md w-full bg-white border border-[#e5e2da] rounded-xs p-8 shadow-sm space-y-5 text-center">
          <div className="w-12 h-12 rounded-full bg-emerald-50 border border-emerald-300 flex items-center justify-center mx-auto text-emerald-700 font-bold text-sm">
            ✓
          </div>
          <div className="space-y-1">
            <h2 className="font-sans text-xl font-medium text-sumi">
              Assessment Concluded
            </h2>
            <p className="text-sumi-muted text-[11px] uppercase tracking-wider">
              Verification Docket #{validation.interview_id?.slice(0, 8)}
            </p>
          </div>
          <p className="font-sans text-xs text-sumi-secondary leading-relaxed">
            Thank you, {validation.participant?.full_name}. Your interview audio, dialogue turns, and technical rubric metrics have been submitted to the evaluation engine.
          </p>
          <div className="p-3 bg-[#fbf9f5] rounded border border-[#e5e2da] text-left space-y-1 font-mono text-[11px] text-sumi-secondary">
            <div>Candidate: {validation.participant?.full_name}</div>
            <div>Role: {validation.job_role?.title || validation.track?.title}</div>
            <div>Status: Pending Dialectic Assessment Evaluation</div>
          </div>
          <div className="pt-2">
            <Link
              href="/"
              className="inline-block px-4 py-2 bg-sumi text-white rounded-xs uppercase tracking-wider font-semibold"
            >
              Exit Stage
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen w-screen overflow-hidden flex flex-col bg-[#fbf9f5] text-sumi relative">
      {/* Entry Stage Gate Modal (Guarantees user gesture for Web Audio autoplay & mic setup) */}
      {!hasEnteredStage && (
        <div className="fixed inset-0 z-50 bg-[#1b1c1a]/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-[#fbf9f5] border border-[#e5e2da] rounded-xs shadow-2xl max-w-lg w-full p-6 space-y-6 text-left">
            <div className="border-b border-[#e5e2da] pb-4 flex items-center justify-between">
              <div>
                <span className="font-mono text-[10px] uppercase tracking-wider text-[#797670] block">
                  Candidate Assessment Checkpoint
                </span>
                <h2 className="font-sans text-lg font-semibold text-[#1b1c1a] mt-0.5">
                  {validation?.candidate?.full_name || validation?.participant?.full_name || "Candidate Assessment"}
                </h2>
              </div>
              <div className="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-pulse" />
            </div>

            <div className="space-y-3 font-mono text-xs text-[#57534e]">
              <div className="bg-white border border-[#e5e2da] p-3.5 rounded-xs space-y-2">
                <div className="flex justify-between text-[11px]">
                  <span className="text-[#797670]">Assessment Track:</span>
                  <span className="font-semibold text-[#1b1c1a]">
                    {validation?.job_role?.title || validation?.track?.title || "Technical System Evaluation"}
                  </span>
                </div>
                <div className="flex justify-between text-[11px]">
                  <span className="text-[#797670]">Interviewer:</span>
                  <span className="font-semibold text-[#1b1c1a]">Voxi (AI Voice Assessor)</span>
                </div>
                <div className="flex justify-between text-[11px]">
                  <span className="text-[#797670]">Mode:</span>
                  <span className="text-emerald-700 font-medium">Hands-Free Conversational Voice</span>
                </div>
              </div>

              <div className="p-3 bg-[#f2efe9] rounded-xs border border-[#e5e2da] space-y-1">
                <p className="text-[11px] text-[#57534e] leading-relaxed">
                  Click the button below to initialize your speakers and microphone. Once connected, Voxi will welcome you and ask the first question. Speak naturally at any time.
                </p>
              </div>
            </div>

            <button
              type="button"
              onClick={handleEnterStage}
              className="w-full py-3.5 bg-[#1b1c1a] hover:bg-neutral-800 text-white font-mono text-xs font-semibold tracking-wider uppercase rounded-xs transition-colors flex items-center justify-center gap-2 cursor-pointer shadow-sm"
            >
              <span>Enter Interview Stage</span>
              <span className="text-xs">→</span>
            </button>
          </div>
        </div>
      )}

      {/* Top Header Bar */}
      <header className="h-12 border-b border-[#e5e2da] bg-white flex items-center justify-between px-6 z-20 shrink-0">
        <div className="flex items-center gap-4">
          <Link href="/" className="flex items-center gap-2 group">
            <Logo size={18} />
            <span className="font-mono text-xs font-bold tracking-tight text-[#1b1c1a] group-hover:text-sumi transition-colors">
              VOXI
            </span>
          </Link>
          <span className="text-[#c7c7c0]">/</span>
          <div className="flex items-center gap-2 font-mono text-xs text-[#57534e]">
            <span className="uppercase text-[#1b1c1a] font-semibold">
              {validation.candidate?.full_name || validation.participant?.full_name || "Candidate"}
            </span>
            <span className="text-[#c7c7c0]">•</span>
            <span>
              {validation.job_role?.title ||
                validation.track?.title ||
                validation.docket_title ||
                validation.blueprint_title ||
                "Technical Assessment"}
            </span>
          </div>
        </div>

        <div className="flex items-center gap-4">
          <div className="flex items-center gap-1.5 font-mono text-xs text-[#1b1c1a] bg-[#efeeea] px-3 py-1 rounded-xs border border-[#e5e2da]">
            <span className="font-mono text-[11px] font-bold tracking-wider">
              TURN #{turnNumber}
            </span>
          </div>

          <button
            type="button"
            onClick={handleResetSession}
            disabled={isRestarting}
            className="px-3 py-1 bg-[#efeeea] hover:bg-neutral-800 hover:text-white text-[#57534e] rounded-xs font-mono text-xs transition-colors flex items-center gap-1.5 border border-[#e5e2da] cursor-pointer disabled:opacity-50"
            title="Reset interview session and restart from Turn #1"
          >
            <span>{isRestarting ? "Restarting..." : "Restart Stage"}</span>
          </button>

          <button
            type="button"
            onClick={() => setShowEndModal(true)}
            className="px-3 py-1 bg-[#efeeea] hover:bg-red-600 hover:text-white text-red-600 rounded-xs font-mono text-xs transition-colors flex items-center gap-1.5 border border-red-200 cursor-pointer"
          >
            <span>Conclude</span>
          </button>
        </div>
      </header>

      {/* Main Workspace Layout */}
      <main className="w-full flex-1 min-h-0 overflow-hidden px-4 sm:px-6 py-2.5 max-w-[1780px] mx-auto flex flex-col">
        <div className="grid grid-cols-1 xl:grid-cols-12 gap-3.5 h-full min-h-0 relative">
          {/* Left/Center Sanctuary: Voxi Stage Viewport */}
          <div
            className={`${
              isDrawerExpanded ? "xl:col-span-8" : "xl:col-span-9"
            } flex flex-col gap-3 h-full min-h-0 transition-all duration-300`}
          >
            {/* Visualizer & Stage Viewport Card */}
            <div className="flex-1 bg-white border border-[#e5e2da] rounded-xs shadow-xs p-6 flex flex-col items-center justify-between relative overflow-hidden">
              {/* Subtle Swiss Grid Ambient Lines */}
              <div className="absolute inset-0 bg-[linear-gradient(to_right,#f5f4ef_1px,transparent_1px),linear-gradient(to_bottom,#f5f4ef_1px,transparent_1px)] bg-[size:40px_40px] pointer-events-none opacity-40"></div>

              {/* Status Header inside Stage */}
              <div className="w-full flex items-center justify-between z-10 font-mono text-xs text-sumi-secondary">
                <div className="flex items-center gap-2">
                  <span
                    className={`w-2 h-2 rounded-full ${
                      isConnected ? "bg-emerald-500 animate-pulse" : "bg-red-400"
                    }`}
                  />
                  <span className="uppercase text-[11px] tracking-wider text-sumi font-semibold">
                    {isConnected ? "Stage Online" : "Connecting..."}
                  </span>
                </div>
                <div className="text-[11px] text-sumi-muted tracking-tight">
                  AEC Echo-Canceled • 16kHz PCM Lossless
                </div>
              </div>

              {/* Central Dynamic Waveform Orb & Avatar */}
              <div className="my-auto flex flex-col items-center justify-center gap-6 z-10 py-6">
                <div className="relative flex items-center justify-center">
                  {/* Waveform Halo Pulse */}
                  <div
                    className={`absolute rounded-full transition-all duration-500 pointer-events-none ${
                      isAgentTurn
                        ? "w-48 h-48 bg-emerald-500/10 animate-ping opacity-30"
                        : isUserSpeaking
                        ? "w-48 h-48 bg-vermilion/10 animate-ping opacity-30"
                        : "w-36 h-36 bg-sumi/5 opacity-10"
                    }`}
                  />
                  <div
                    className={`w-28 h-28 rounded-full border flex items-center justify-center transition-all duration-300 shadow-sm ${
                      isAgentTurn
                        ? "bg-emerald-50 border-emerald-400 scale-105"
                        : isUserSpeaking
                        ? "bg-vermilion/5 border-vermilion scale-105"
                        : "bg-[#fbf9f5] border-[#e5e2da]"
                    }`}
                  >
                    <div className="flex items-center gap-1.5 h-10">
                      {vuLevels.map((lvl, idx) => (
                        <div
                          key={idx}
                          style={{ height: `${Math.max(lvl, 4)}px` }}
                          className={`w-1 rounded-full transition-all duration-75 ${
                            isAgentTurn
                              ? "bg-emerald-600"
                              : isUserSpeaking
                              ? "bg-vermilion"
                              : "bg-[#c7c7c0]"
                          }`}
                        />
                      ))}
                    </div>
                  </div>
                </div>

                {/* Real-time Status Callout */}
                <div className="text-center space-y-1">
                  <h3 className="font-mono text-xs uppercase tracking-widest text-[#57534e]">
                    {turnStatusText}
                  </h3>
                  <p className="font-sans text-xs text-sumi-muted max-w-md mx-auto">
                    {isAgentTurn
                      ? "Voxi is currently speaking. You can speak naturally at any time to interrupt."
                      : "Microphone is live and AEC echo-canceled. Answer at your normal speaking pace."}
                  </p>
                </div>
              </div>

              {/* Floating Bottom Subtitle / Live STT Feed */}
              <div className="w-full max-w-2xl bg-[#fbf9f5] border border-[#e5e2da] rounded-xs px-4 py-3 z-10 min-h-[52px] flex items-center justify-between">
                <div className="flex items-center gap-2 overflow-hidden w-full font-mono text-xs">
                  <span className="text-sumi-muted uppercase text-[10px] shrink-0 font-semibold">
                    {liveSttText ? "Hearing You:" : "Voxi Status:"}
                  </span>
                  <span
                    className={`truncate ${
                      liveSttText ? "text-vermilion font-medium" : "text-sumi-secondary"
                    }`}
                  >
                    {liveSttText ||
                      (isAgentTurn
                        ? "Voxi speaking..."
                        : "Ready for your answer...")}
                  </span>
                </div>
              </div>
            </div>

            {/* Candidate PiP & Diagnostics Bar */}
            <div className="h-28 bg-white border border-[#e5e2da] rounded-xs p-3 flex items-center justify-between shrink-0 shadow-xs">
              {/* Left: Video Preview */}
              <div className="flex items-center gap-4 h-full">
                <div className="h-full aspect-video bg-[#1b1c1a] rounded-xs border border-[#e5e2da] overflow-hidden relative flex items-center justify-center">
                  <video
                    ref={videoRef}
                    autoPlay
                    playsInline
                    muted
                    className={`w-full h-full object-cover ${!isCamActive ? "hidden" : ""}`}
                  />
                  {!isCamActive && (
                    <span className="text-[10px] font-mono text-sumi-muted uppercase">
                      Cam Muted
                    </span>
                  )}
                  <div className="absolute bottom-1 left-1 bg-black/60 px-1.5 py-0.5 rounded-xs font-mono text-[9px] text-white uppercase">
                    You (PiP)
                  </div>
                </div>

                <div className="space-y-1 font-mono text-xs">
                  <div className="font-semibold text-sumi">
                    {validation.candidate?.full_name || validation.participant?.full_name}
                  </div>
                  <div className="text-[11px] text-sumi-muted flex items-center gap-2">
                    <span>
                      Mic:{" "}
                      <span className={isMicActive ? "text-emerald-700" : "text-red-600 font-bold"}>
                        {isMicActive ? "Active" : "Muted"}
                      </span>
                    </span>
                    <span>•</span>
                    <span>
                      Cam:{" "}
                      <span className={isCamActive ? "text-emerald-700" : "text-red-600 font-bold"}>
                        {isCamActive ? "Active" : "Muted"}
                      </span>
                    </span>
                  </div>
                </div>
              </div>

              {/* Right: Quick Controls */}
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={toggleMic}
                  className={`px-3 py-1.5 rounded-xs font-mono text-xs border transition-colors cursor-pointer ${
                    isMicActive
                      ? "border-[#e5e2da] hover:border-sumi bg-white text-sumi"
                      : "border-red-300 bg-red-50 text-red-700 font-semibold"
                  }`}
                >
                  {isMicActive ? "Mute Mic" : "Unmute Mic"}
                </button>
                <button
                  type="button"
                  onClick={toggleCam}
                  className={`px-3 py-1.5 rounded-xs font-mono text-xs border transition-colors cursor-pointer ${
                    isCamActive
                      ? "border-[#e5e2da] hover:border-sumi bg-white text-sumi"
                      : "border-red-300 bg-red-50 text-red-700 font-semibold"
                  }`}
                >
                  {isCamActive ? "Stop Cam" : "Start Cam"}
                </button>
              </div>
            </div>
          </div>

          {/* Right Sanctuary: Live Transcript Docket & Rubric Criteria */}
          <div
            className={`${
              isDrawerExpanded ? "xl:col-span-4" : "xl:col-span-3"
            } h-full min-h-0 bg-white border border-[#e5e2da] rounded-xs shadow-xs flex flex-col overflow-hidden transition-all duration-300`}
          >
            {/* Transcript Drawer Header */}
            <div className="p-3.5 border-b border-[#e5e2da] flex items-center justify-between shrink-0 bg-[#fbf9f5]">
              <div className="flex items-center gap-2">
                <span className="font-mono text-xs font-bold uppercase tracking-wider text-sumi">
                  Turn Transcript
                </span>
                <span className="font-mono text-[10px] bg-[#efeeea] px-1.5 py-0.5 rounded-xs border border-[#e5e2da] text-sumi-secondary">
                  {transcriptHistory.length} Turns
                </span>
              </div>
              <button
                type="button"
                onClick={() => setIsDrawerExpanded((prev) => !prev)}
                className="font-mono text-[11px] text-sumi-muted hover:text-sumi transition-colors cursor-pointer"
                title="Expand or collapse transcript drawer"
              >
                {isDrawerExpanded ? "Collapse ⇥" : "Expand ⇤"}
              </button>
            </div>

            {/* Transcript Feed Body */}
            <div className="flex-1 min-h-0 overflow-y-auto p-4 space-y-4 font-mono text-xs">
              {transcriptHistory.length === 0 ? (
                <div className="h-full flex flex-col items-center justify-center text-center p-6 text-sumi-muted space-y-2">
                  <div className="w-8 h-8 rounded-full border border-dashed border-[#c7c7c0] flex items-center justify-center">
                    💬
                  </div>
                  <p className="text-[11px] leading-relaxed">
                    Conversation history will appear here in real-time as you and Voxi speak.
                  </p>
                </div>
              ) : (
                transcriptHistory.map((item, idx) => (
                  <div
                    key={idx}
                    className={`space-y-1 p-3 rounded-xs border text-left ${
                      item.speaker === "Voxi"
                        ? "bg-[#fbf9f5] border-[#e5e2da]"
                        : "bg-white border-[#d8d5cd]"
                    }`}
                  >
                    <div className="flex items-center justify-between text-[10px] text-sumi-muted">
                      <span
                        className={`font-bold uppercase tracking-wider ${
                          item.speaker === "Voxi" ? "text-emerald-800" : "text-sumi"
                        }`}
                      >
                        {item.speaker === "Voxi" ? "AI Interviewer (Voxi)" : "You (Candidate)"}
                      </span>
                      <span>{item.time}</span>
                    </div>
                    <p className="font-sans text-xs text-sumi leading-relaxed whitespace-pre-wrap">
                      {item.text}
                    </p>
                  </div>
                ))
              )}
            </div>

            {/* Competency & Docket Criteria Summary */}
            <div className="p-3 border-t border-[#e5e2da] bg-[#fbf9f5] shrink-0 font-mono text-[11px] space-y-1">
              <div className="text-sumi-muted uppercase text-[10px] font-semibold">
                Evaluation Blueprint
              </div>
              <div className="text-sumi font-medium truncate">
                {validation.docket_title || validation.blueprint_title || "Technical Dialogue Track"}
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Persistent Swiss Minimalist Footer */}
      <footer className="h-10 border-t border-[#e5e2da] bg-white flex items-center justify-between px-6 z-20 shrink-0">
        <div className="text-xs font-mono text-sumi-secondary flex items-center gap-2">
          <span className="font-medium text-sumi">Two-Way Full-Duplex Audio Active</span>
          <span className="text-[#c7c7c0]">•</span>
          <span className="text-sumi-muted">Hands-Free Conversational Stage</span>
        </div>

        {/* Center Conversational State Pill */}
        <div className="flex items-center gap-3">
          <div
            className={`px-4 py-1.5 rounded-xs border font-mono text-xs flex items-center gap-2 transition-all ${
              isAgentTurn
                ? "bg-emerald-50 border-emerald-300 text-emerald-900"
                : isUserSpeaking
                ? "bg-vermilion/10 border-vermilion/40 text-vermilion font-semibold"
                : "bg-[#f5f3ef] border-[#e5e2da] text-sumi"
            }`}
          >
            <span>
              {isAgentTurn
                ? "🔊 Voxi Speaking (Speak to interrupt)"
                : isUserSpeaking
                ? "🎙️ You Are Speaking..."
                : "👂 Voxi Listening..."}
            </span>
          </div>

          <button
            type="button"
            onClick={toggleMic}
            className={`px-3 py-1 rounded-xs font-mono text-xs border transition-colors cursor-pointer ${
              isMicActive
                ? "border-[#e5e2da] hover:border-sumi bg-white text-sumi"
                : "border-red-300 bg-red-50 text-red-700 font-semibold"
            }`}
          >
            {isMicActive ? "Mute Mic" : "Unmute Mic"}
          </button>
        </div>

        <div className="text-xs font-mono text-sumi-muted hidden sm:flex items-center gap-2">
          <span>AEC &amp; VAD Enabled</span>
        </div>
      </footer>

      {/* Modal: End Session Confirmation */}
      {showEndModal && (
        <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white border border-[#e5e2da] rounded-sm max-w-sm w-full p-6 shadow-2xl text-center space-y-4 font-mono text-xs">
            <h3 className="font-bold text-sumi uppercase tracking-wider text-sm">
              End Assessment Session?
            </h3>
            <p className="font-sans text-xs text-sumi-secondary leading-relaxed">
              Are you sure you want to conclude the assessment? Your recorded turns will be submitted for scoring.
            </p>
            <div className="flex items-center justify-center gap-3 pt-2">
              <button
                type="button"
                onClick={() => setShowEndModal(false)}
                className="px-4 py-2 border border-[#e5e2da] hover:border-sumi bg-white rounded uppercase cursor-pointer"
              >
                Continue Interview
              </button>
              <button
                type="button"
                onClick={handleEndSession}
                className="px-4 py-2 bg-red-600 text-white hover:bg-red-700 rounded uppercase font-semibold cursor-pointer"
              >
                Conclude &amp; Submit
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default function InterviewPage() {
  return (
    <Suspense
      fallback={
        <div className="p-12 text-center font-mono text-xs text-sumi">
          Initializing Voice Stage...
        </div>
      }
    >
      <LiveCandidateStageContent />
    </Suspense>
  );
}
