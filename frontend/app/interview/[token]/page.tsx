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
  const [debugMode, setDebugMode] = useState<boolean>(false);
  const [consentChecked, setConsentChecked] = useState<boolean>(false);
  const [chimePlaying, setChimePlaying] = useState<boolean>(false);

  useEffect(() => {
    if (typeof window !== "undefined") {
      const sp = new URLSearchParams(window.location.search);
      setDebugMode(sp.get("debug") === "1");
    }
  }, []);

  // References
  const wsRef = useRef<WebSocket | null>(null);
  const audioContextRef = useRef<AudioContext | null>(null);
  const vadAudioCtxRef = useRef<AudioContext | null>(null);
  const activeSourcesRef = useRef<AudioBufferSourceNode[]>([]);
  const nextAudioStartTimeRef = useRef<number>(0);
  const interruptedTurnRef = useRef<number>(0);
  const micStreamRef = useRef<MediaStream | null>(null);
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const preflightVideoRef = useRef<HTMLVideoElement | null>(null);
  const cameraStreamRef = useRef<MediaStream | null>(null);
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

    async function startCamera() {
      try {
        cameraStream = await navigator.mediaDevices.getUserMedia({
          video: true,
          audio: false,
        });
        cameraStreamRef.current = cameraStream;
        if (videoRef.current) {
          videoRef.current.srcObject = cameraStream;
        }
        if (preflightVideoRef.current) {
          preflightVideoRef.current.srcObject = cameraStream;
        }
        setIsCamActive(true);
        cameraStream.getVideoTracks().forEach((track) => {
          track.onended = () => setIsCamActive(false);
          track.onmute = () => setIsCamActive(false);
          track.onunmute = () => setIsCamActive(true);
        });
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
      cameraStreamRef.current = null;
    };
  }, [validation]);

  // Synchronize camera video feeds across preflight & live stage transitions
  useEffect(() => {
    if (cameraStreamRef.current && isCamActive) {
      if (videoRef.current && videoRef.current.srcObject !== cameraStreamRef.current) {
        videoRef.current.srcObject = cameraStreamRef.current;
      }
      if (preflightVideoRef.current && preflightVideoRef.current.srcObject !== cameraStreamRef.current) {
        preflightVideoRef.current.srcObject = cameraStreamRef.current;
      }
    }
  }, [hasEnteredStage, isCamActive]);

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
    if (!validation?.valid || !token || !hasEnteredStage) return;

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
    };

    ws.onmessage = async (event) => {
      try {
        const msg = JSON.parse(event.data);

        switch (msg.type) {
          case "connected":
            setIsConnected(true);
            if (typeof msg.turn === "number" && msg.turn > 0) {
              setTurnNumber(msg.turn);
              turnNumberRef.current = msg.turn;
            }
            if (msg.has_history) {
              setTurnStatusText("Reconnected to ongoing assessment.");
            } else {
              setTurnStatusText("Stage online. Ready to begin...");
            }
            break;

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

          case "assessment_complete":
          case "interview_complete":
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
  }, [validation, token, hasEnteredStage, stopAllAudioPlayback]);

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
        setIsMicActive(true);
        stream.getAudioTracks().forEach((track) => {
          track.onended = () => setIsMicActive(false);
          track.onmute = () => setIsMicActive(false);
          track.onunmute = () => setIsMicActive(true);
        });

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

          if (!isAudioActivelyPlaying) {
            noiseFloorRef.current = Math.min(
              Math.max(noiseFloorRef.current * 0.98 + voiceBandAvg * 0.02, 12),
              55
            );
          }

          // Barge-in requires deliberate energy over speakers (threshold 40+, 12 frames)
          const speechThreshold = isAudioActivelyPlaying
            ? Math.max(40, noiseFloorRef.current * 2.2)
            : Math.max(20, noiseFloorRef.current * 1.3);

          const requiredFrames = isAudioActivelyPlaying ? 12 : 4;

          if (voiceBandAvg > speechThreshold && isMicActiveRef.current) {
            consecutiveSpeechFrames++;
            if (consecutiveSpeechFrames >= requiredFrames) {
              setIsUserSpeaking(true);
            }

            // Candidate interruption detected ONLY while audio is actively playing out of speakers
            if (isAudioActivelyPlaying && consecutiveSpeechFrames >= 12) {
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
            consecutiveSpeechFrames = Math.max(0, consecutiveSpeechFrames - 1);
            if (consecutiveSpeechFrames === 0) {
              setIsUserSpeaking(false);
            }
          }

          animFrameRef.current = requestAnimationFrame(updateEnergyLoop);
        };

        animFrameRef.current = requestAnimationFrame(updateEnergyLoop);

        // High-Performance 16kHz Linear PCM Streamer
        // Downsamples Float32 PCM to 16,000Hz mono Int16
        const bufferSize = 2048;
        processorNode = vadAudioCtx.createScriptProcessor(bufferSize, 1, 1);
        const targetSampleRate = 16000;
        const sampleRatio = vadAudioCtx.sampleRate / targetSampleRate;

        processorNode.onaudioprocess = (e) => {
          if (!isSubscribed || !wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return;
          if (!isMicActiveRef.current) return;

          const inputData = e.inputBuffer.getChannelData(0);
          const outputLength = Math.floor(inputData.length / sampleRatio);
          const pcmData = new Int16Array(outputLength);

          // Anti-aliasing box filter + Float32 to Int16 conversion
          for (let i = 0; i < outputLength; i++) {
            const startIdx = Math.floor(i * sampleRatio);
            const endIdx = Math.min(inputData.length, Math.floor((i + 1) * sampleRatio));
            let sum = 0;
            const count = Math.max(1, endIdx - startIdx);
            for (let j = startIdx; j < endIdx; j++) {
              sum += inputData[j];
            }
            const s = Math.max(-1, Math.min(1, sum / count));
            pcmData[i] = s < 0 ? s * 0x8000 : s * 0x7FFF;
          }

          // Convert Int16Array to base64
          const bytes = new Uint8Array(pcmData.buffer);
          let binary = "";
          for (let i = 0; i < bytes.byteLength; i++) {
            binary += String.fromCharCode(bytes[i]);
          }
          const base64 = btoa(binary);

          const isAudioActivelyPlaying =
            activeSourcesRef.current.length > 0 ||
            (audioContextRef.current !== null &&
              nextAudioStartTimeRef.current > audioContextRef.current.currentTime);

          // Continuous audio stream to Deepgram STT when Voxi is not speaking.
          // Deepgram's server-side neural VAD accurately segments speech and handles natural pauses.
          // When Voxi IS speaking, audio streams if candidate interrupts.
          const shouldStreamAudio = !isAudioActivelyPlaying || isUserSpeakingRef.current;

          if (!shouldStreamAudio) {
            return;
          }

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

  // Re-enable / Toggle Microphone
  const toggleMic = async () => {
    if (micStreamRef.current && micStreamRef.current.getAudioTracks().length > 0) {
      const track = micStreamRef.current.getAudioTracks()[0];
      if (track.readyState !== "ended") {
        const next = !isMicActive;
        track.enabled = next;
        setIsMicActive(next);
        return;
      }
    }
    try {
      const fresh = await navigator.mediaDevices.getUserMedia({
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true,
        },
      });
      micStreamRef.current = fresh;
      setIsMicActive(true);
      fresh.getAudioTracks().forEach((track) => {
        track.onended = () => setIsMicActive(false);
        track.onmute = () => setIsMicActive(false);
        track.onunmute = () => setIsMicActive(true);
      });
    } catch {
      setIsMicActive(false);
    }
  };

  // Re-enable / Toggle Camera
  const toggleCam = async () => {
    if (cameraStreamRef.current && cameraStreamRef.current.getVideoTracks().length > 0) {
      const track = cameraStreamRef.current.getVideoTracks()[0];
      if (track.readyState !== "ended") {
        const next = !isCamActive;
        track.enabled = next;
        setIsCamActive(next);
        return;
      }
    }
    try {
      const cameraStream = await navigator.mediaDevices.getUserMedia({ video: true });
      cameraStreamRef.current = cameraStream;
      if (videoRef.current) {
        videoRef.current.srcObject = cameraStream;
      }
      if (preflightVideoRef.current) {
        preflightVideoRef.current.srcObject = cameraStream;
      }
      setIsCamActive(true);
      cameraStream.getVideoTracks().forEach((track) => {
        track.onended = () => setIsCamActive(false);
        track.onmute = () => setIsCamActive(false);
        track.onunmute = () => setIsCamActive(true);
      });
    } catch {
      setIsCamActive(false);
    }
  };

  const playTestChime = () => {
    try {
      setChimePlaying(true);
      const AudioCtxClass =
        window.AudioContext ||
        (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
      const ctx = new AudioCtxClass();
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.type = "sine";
      osc.frequency.setValueAtTime(440, ctx.currentTime);
      gain.gain.setValueAtTime(0.12, ctx.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.4);
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.start();
      osc.stop(ctx.currentTime + 0.4);
      setTimeout(() => setChimePlaying(false), 500);
    } catch (e) {
      console.warn("Test chime failed:", e);
      setChimePlaying(false);
    }
  };

  const handleFinishSpeaking = useCallback(() => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN && !isAgentTurn) {
      wsRef.current.send(JSON.stringify({ type: "finish_turn" }));
      setTurnStatusText("Submitted answer. Voxi is responding...");
    }
  }, [isAgentTurn]);

  // Keyboard shortcut: Press Enter to Finish Answer during candidate turn
  useEffect(() => {
    if (!hasEnteredStage) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement
      ) {
        return;
      }
      if (e.key === "Enter" && !isAgentTurn) {
        e.preventDefault();
        handleFinishSpeaking();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [hasEnteredStage, isAgentTurn, handleFinishSpeaking]);

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
  };

  const handleEndSession = () => {
    stopAllAudioPlayback();
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: "end_assessment" }));
    }
    if (micStreamRef.current) {
      micStreamRef.current.getTracks().forEach((t) => t.stop());
    }
    // Graceful fallback to close if server acknowledgment is delayed
    setTimeout(() => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.close();
      }
      setIsSessionCompleted(true);
      setShowEndModal(false);
    }, 1500);
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

  if (!hasEnteredStage) {
    return (
      <div className="min-h-screen bg-[#fbf9f5] text-[#141413] flex flex-col font-sans selection:bg-[#141413] selection:text-[#fbf9f5]">
        {/* Monograph Top Navigation */}
        <header className="w-full border-b border-[#e5e2da] bg-white px-6 lg:px-12 py-4 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link href="/" className="flex items-center gap-2 group">
              <Logo size={20} />
              <span className="font-mono text-xs font-bold tracking-tight text-[#141413]">
                VOXI ASSESSMENT ENGINE
              </span>
            </Link>
            <span className="text-[#c7c7c0] font-mono text-xs">/</span>
            <span className="font-mono text-[11px] tracking-wider uppercase text-[#797670]">
              PLATE NO. 02 • PRE-FLIGHT VERIFICATION
            </span>
          </div>

          <div className="flex items-center gap-3 font-mono text-[11px] text-[#57534e]">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
            <span className="tracking-wider uppercase font-semibold text-[#141413]">
              GATEWAY READY
            </span>
            <span className="text-[#c7c7c0]">•</span>
            <span className="text-[#797670] font-mono">
              tok_{token ? token.slice(0, 8) : "valid"}
            </span>
          </div>
        </header>

        {/* Main Monograph Setup Stage Body */}
        <main className="flex-1 w-full max-w-5xl mx-auto px-6 py-10 lg:py-14 flex flex-col justify-center space-y-10">
          {/* Folio Headstone */}
          <div className="space-y-2 pb-6 border-b border-[#e5e2da]">
            <div className="flex items-center gap-2 font-mono text-[10px] tracking-[0.2em] text-[#797670] uppercase">
              <span>FOLIO REF. STAGE-01</span>
              <span>•</span>
              <span>AUTONOMOUS ACOUSTIC &amp; VIDEO CALIBRATION</span>
            </div>
            <h1 className="text-3xl md:text-4xl font-light tracking-tight text-[#141413]">
              Participant Entry Stage
            </h1>
            <p className="font-serif italic text-lg md:text-xl text-[#57534e] leading-relaxed">
              Verify your camera framing, acoustic microphone sensitivity, and audio output before entering the dialogue sanctuary.
            </p>
          </div>

          {/* Candidate & Session Dossier */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 p-5 bg-white border border-[#e5e2da] rounded-xs font-mono text-xs">
            <div className="space-y-1 md:col-span-2">
              <span className="text-[10px] uppercase tracking-wider text-[#797670] block">Candidate Dossier</span>
              <div className="font-sans text-base font-semibold text-[#141413]">
                {validation.candidate?.full_name || validation.participant?.full_name || "Technical Candidate"}
              </div>
              <div className="text-[11px] text-[#797670]">
                {validation.job_role?.title || validation.track?.title || "Technical System Assessment"}
              </div>
            </div>
            <div className="space-y-1">
              <span className="text-[10px] uppercase tracking-wider text-[#797670] block">Interviewer</span>
              <div className="font-sans text-sm font-medium text-[#141413]">Voxi (Autonomous Assessor)</div>
              <div className="text-[11px] text-[#797670]">Duplex Acoustic Engine</div>
            </div>
            <div className="space-y-1">
              <span className="text-[10px] uppercase tracking-wider text-[#797670] block">Session Protocol</span>
              <div className="font-sans text-sm font-medium text-[#141413]">Turn-Based Dialectic</div>
              <div className="text-[11px] text-[#797670]">2.2s Pause Wait + Fast Submit</div>
            </div>
          </div>

          {/* 3 Hardware & Acoustic Verification Stations */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {/* Station 01: Live Camera Feed */}
            <div className="p-5 bg-white border border-[#e5e2da] rounded-xs space-y-4 flex flex-col justify-between">
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="font-mono text-[10px] tracking-wider uppercase text-[#797670]">01 / Optical Feed</span>
                  <span className={`w-2 h-2 rounded-full ${isCamActive ? "bg-emerald-500 animate-pulse" : "bg-red-500"}`} />
                </div>
                <div className="font-sans text-sm font-semibold text-[#141413]">HD Video Camera</div>
                <div className="aspect-video w-full bg-[#141413] rounded-xs overflow-hidden relative flex items-center justify-center border border-[#e5e2da]">
                  <video
                    ref={preflightVideoRef}
                    autoPlay
                    playsInline
                    muted
                    className={`w-full h-full object-cover ${!isCamActive ? "hidden" : ""}`}
                  />
                  {!isCamActive && (
                    <div className="text-center p-3 space-y-1">
                      <span className="text-xs text-red-400 block font-mono">Camera Disabled</span>
                      <span className="text-[10px] text-neutral-400">Click below to grant access</span>
                    </div>
                  )}
                  <div className="absolute bottom-1.5 left-1.5 bg-black/70 px-2 py-0.5 rounded-xs font-mono text-[9px] text-white uppercase">
                    1080p Live Feed
                  </div>
                </div>
              </div>

              <div className="pt-2">
                <button
                  type="button"
                  onClick={toggleCam}
                  className={`w-full py-2 font-mono text-xs uppercase tracking-wider font-semibold rounded-xs border transition-colors cursor-pointer flex items-center justify-center gap-2 ${
                    isCamActive
                      ? "border-emerald-300 bg-emerald-50 text-emerald-800"
                      : "border-red-300 bg-red-600 text-white hover:bg-red-700"
                  }`}
                >
                  <span>{isCamActive ? "✓ Camera Verified & Active" : "Grant Camera Access"}</span>
                </button>
              </div>
            </div>

            {/* Station 02: Studio Microphone & VU Meter */}
            <div className="p-5 bg-white border border-[#e5e2da] rounded-xs space-y-4 flex flex-col justify-between">
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="font-mono text-[10px] tracking-wider uppercase text-[#797670]">02 / Acoustic Input</span>
                  <span className={`w-2 h-2 rounded-full ${isMicActive ? "bg-emerald-500 animate-pulse" : "bg-red-500"}`} />
                </div>
                <div className="font-sans text-sm font-semibold text-[#141413]">Studio Microphone</div>

                {/* Real-time VU Meter Bar Graph */}
                <div className="aspect-video w-full bg-[#fbf9f5] border border-[#e5e2da] rounded-xs p-4 flex flex-col items-center justify-center space-y-3">
                  <div className="flex items-center gap-1.5 h-12">
                    {vuLevels.map((lvl, idx) => (
                      <div
                        key={idx}
                        style={{ height: `${Math.max(lvl, 4)}px` }}
                        className={`w-2 rounded-xs transition-all duration-75 ${
                          !isMicActive
                            ? "bg-red-300"
                            : isUserSpeaking
                            ? "bg-emerald-600"
                            : "bg-[#141413]"
                        }`}
                      />
                    ))}
                  </div>
                  <div className="font-mono text-[10px] text-[#797670] flex items-center gap-2">
                    <span>Signal:</span>
                    <span className={isUserSpeaking ? "text-emerald-700 font-bold" : "text-[#141413]"}>
                      {isMicActive ? (isUserSpeaking ? "Voice Speaking Active" : "Detecting Ambient Level") : "Mic Inactive"}
                    </span>
                  </div>
                </div>
              </div>

              <div className="space-y-2 pt-2">
                <button
                  type="button"
                  onClick={toggleMic}
                  className={`w-full py-2 font-mono text-xs uppercase tracking-wider font-semibold rounded-xs border transition-colors cursor-pointer flex items-center justify-center gap-2 ${
                    isMicActive
                      ? "border-emerald-300 bg-emerald-50 text-emerald-800"
                      : "border-red-300 bg-red-600 text-white hover:bg-red-700"
                  }`}
                >
                  <span>{isMicActive ? "✓ Microphone Verified" : "Grant Microphone Access"}</span>
                </button>
              </div>
            </div>

            {/* Station 03: Audio Monitoring & Cadence Guidelines */}
            <div className="p-5 bg-white border border-[#e5e2da] rounded-xs space-y-4 flex flex-col justify-between">
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="font-mono text-[10px] tracking-wider uppercase text-[#797670]">03 / Audio &amp; Cadence</span>
                  <span className="w-2 h-2 rounded-full bg-emerald-500" />
                </div>
                <div className="font-sans text-sm font-semibold text-[#141413]">Audio Output &amp; Pacing</div>

                <div className="p-3.5 bg-[#fbf9f5] border border-[#e5e2da] rounded-xs space-y-2 font-sans text-xs text-[#57534e]">
                  <div className="font-semibold text-[#141413] flex items-center gap-1.5 font-mono text-[11px] uppercase">
                    <span>Conversational Cadence</span>
                  </div>
                  <ul className="space-y-1.5 text-[11px] leading-relaxed list-disc list-inside">
                    <li><strong className="text-[#141413]">2.2s Pause Wait:</strong> Take your time to think; natural pauses will never cut you off.</li>
                    <li><strong className="text-[#141413]">Fast Submit:</strong> Hit <span className="font-mono font-semibold">Enter</span> or click &ldquo;Finish Answer&rdquo; whenever done.</li>
                    <li><strong className="text-[#141413]">Interruptible:</strong> Speak at any second to interrupt Voxi naturally.</li>
                  </ul>
                </div>
              </div>

              <div className="pt-2">
                <button
                  type="button"
                  onClick={playTestChime}
                  disabled={chimePlaying}
                  className="w-full py-2 font-mono text-xs uppercase tracking-wider font-semibold rounded-xs border border-[#e5e2da] bg-[#fbf9f5] hover:bg-neutral-800 hover:text-white text-[#141413] transition-colors cursor-pointer flex items-center justify-center gap-2"
                >
                  <span>{chimePlaying ? "♪ Playing 440Hz Chime..." : "Test 440Hz Audio Output"}</span>
                </button>
              </div>
            </div>
          </div>

          {/* Consent Checkbox & Primary Enter Action */}
          <div className="pt-6 border-t border-[#e5e2da] space-y-6">
            <label className="flex items-start gap-3.5 cursor-pointer group">
              <input
                type="checkbox"
                checked={consentChecked}
                onChange={(e) => setConsentChecked(e.target.checked)}
                className="mt-0.5 rounded-none border-[#141413]/40 text-[#141413] focus:ring-0 w-4 h-4 accent-[#141413] cursor-pointer"
              />
              <span className="text-xs text-[#57534e] leading-relaxed group-hover:text-[#141413] transition-colors select-none">
                I consent to continuous camera capture, microphone audio recording, and automated dialectic assessment evaluation for this interview.
              </span>
            </label>

            <div className="flex flex-col sm:flex-row items-center gap-4">
              <button
                type="button"
                disabled={!isCamActive || !isMicActive || !consentChecked}
                onClick={handleEnterStage}
                className={`w-full sm:w-auto px-8 py-4 font-mono text-xs uppercase tracking-[0.14em] font-semibold rounded-xs transition-all flex items-center justify-center gap-3 shadow-sm ${
                  !isCamActive || !isMicActive || !consentChecked
                    ? "bg-neutral-300 text-neutral-500 cursor-not-allowed opacity-60"
                    : "bg-[#141413] hover:bg-neutral-800 text-white cursor-pointer"
                }`}
              >
                <span>
                  {!isCamActive
                    ? "Camera Required"
                    : !isMicActive
                    ? "Microphone Required"
                    : !consentChecked
                    ? "Check Consent to Enter"
                    : "Enter Interview Stage"}
                </span>
                <span>→</span>
              </button>

              <div className="font-mono text-[11px] text-[#797670]">
                {!isCamActive || !isMicActive
                  ? "Both camera and microphone must be active to proceed."
                  : !consentChecked
                  ? "Please accept the consent declaration above."
                  : "All devices verified. Voxi will introduce itself when you enter."}
              </div>
            </div>
          </div>
        </main>
      </div>
    );
  }

  return (
    <div className="h-screen w-screen overflow-hidden flex flex-col bg-[#fbf9f5] text-sumi relative">

      {/* Mandatory Camera & Microphone Active Enforcement Overlay during Stage */}
      {hasEnteredStage && (!isCamActive || !isMicActive) && (
        <div className="fixed inset-0 z-50 bg-[#141413]/85 backdrop-blur-md flex items-center justify-center p-4">
          <div className="bg-white border-2 border-red-500 rounded-xs shadow-2xl max-w-md w-full p-6 text-center space-y-4">
            <div className="w-12 h-12 rounded-full bg-red-100 border border-red-300 flex items-center justify-center mx-auto text-red-600 font-bold text-lg">
              !
            </div>
            <div>
              <span className="font-mono text-[10px] uppercase tracking-wider text-red-600 font-bold block">
                Proctoring Requirement
              </span>
              <h3 className="font-sans text-xl font-semibold text-sumi mt-1">
                Camera &amp; Microphone Required
              </h3>
              <p className="font-sans text-xs text-sumi-secondary mt-2 leading-relaxed">
                Voxi requires both your camera and microphone to remain active continuously during this technical interview. The assessment dialogue is paused until both feeds are restored.
              </p>
            </div>

            <div className="p-3 bg-[#fbf9f5] rounded-xs border border-[#e5e2da] text-left font-mono text-xs space-y-2">
              <div className="flex items-center justify-between">
                <span>Camera Feed:</span>
                <span className={isCamActive ? "text-emerald-700 font-bold" : "text-red-600 font-bold"}>
                  {isCamActive ? "Active ✓" : "Offline / Muted ✕"}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span>Microphone:</span>
                <span className={isMicActive ? "text-emerald-700 font-bold" : "text-red-600 font-bold"}>
                  {isMicActive ? "Active ✓" : "Offline / Muted ✕"}
                </span>
              </div>
            </div>

            <div className="pt-2 flex flex-col gap-2">
              {!isCamActive && (
                <button
                  type="button"
                  onClick={toggleCam}
                  className="w-full py-2.5 bg-sumi text-white hover:bg-neutral-800 font-mono text-xs uppercase tracking-wider font-semibold rounded-xs transition-colors cursor-pointer"
                >
                  Turn On Camera
                </button>
              )}
              {!isMicActive && (
                <button
                  type="button"
                  onClick={toggleMic}
                  className="w-full py-2.5 bg-sumi text-white hover:bg-neutral-800 font-mono text-xs uppercase tracking-wider font-semibold rounded-xs transition-colors cursor-pointer"
                >
                  Unmute Microphone
                </button>
              )}
            </div>
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

              {/* Floating Bottom Subtitle / Live STT Feed & Manual Finish Action */}
              <div className="w-full max-w-2xl flex flex-col sm:flex-row items-center gap-3 z-10">
                <div className="flex-1 bg-[#fbf9f5] border border-[#e5e2da] rounded-xs px-4 py-3 min-h-[52px] flex items-center justify-between w-full shadow-2xs">
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
                          : "Listening to your answer naturally (2.2s pause grace period)...")}
                    </span>
                  </div>
                </div>

                {!isAgentTurn && (
                  <button
                    type="button"
                    onClick={handleFinishSpeaking}
                    className="px-5 py-3.5 bg-[#141413] hover:bg-neutral-800 text-white font-mono text-xs uppercase tracking-wider font-semibold rounded-xs transition-all flex items-center justify-center gap-2 shrink-0 shadow-sm cursor-pointer"
                    title="Submit your answer immediately without waiting for the 2.2s silence timer (Shortcut: Enter key)"
                  >
                    <span>Finish Answer</span>
                    <span className="text-[10px] opacity-70 font-mono">(Enter ↵)</span>
                  </button>
                )}
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
                  className={`px-3 py-1.5 rounded-xs font-mono text-xs border transition-colors cursor-pointer flex items-center gap-1.5 ${
                    isMicActive
                      ? "border-emerald-300 bg-emerald-50 text-emerald-800 font-semibold"
                      : "border-red-300 bg-red-50 text-red-700 font-semibold animate-pulse"
                  }`}
                  title={isMicActive ? "Microphone active & required" : "Microphone offline! Click to enable"}
                >
                  <span className={`w-1.5 h-1.5 rounded-full ${isMicActive ? "bg-emerald-600" : "bg-red-600"}`} />
                  <span>{isMicActive ? "Mic Live (Required)" : "Enable Mic"}</span>
                </button>
                <button
                  type="button"
                  onClick={toggleCam}
                  className={`px-3 py-1.5 rounded-xs font-mono text-xs border transition-colors cursor-pointer flex items-center gap-1.5 ${
                    isCamActive
                      ? "border-emerald-300 bg-emerald-50 text-emerald-800 font-semibold"
                      : "border-red-300 bg-red-50 text-red-700 font-semibold animate-pulse"
                  }`}
                  title={isCamActive ? "Camera active & required" : "Camera offline! Click to enable"}
                >
                  <span className={`w-1.5 h-1.5 rounded-full ${isCamActive ? "bg-emerald-600" : "bg-red-600"}`} />
                  <span>{isCamActive ? "Cam Live (Required)" : "Enable Cam"}</span>
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

      {/* Voice Diagnostics Telemetry Overlay */}
      {debugMode && (
        <aside aria-label="Voice Debug Telemetry" className="fixed bottom-4 right-4 z-50 bg-[#1b1c1a]/95 text-white border border-[#44403c] rounded p-3 font-mono text-[10px] space-y-1 shadow-lg max-w-xs">
          <div className="font-bold text-emerald-400 border-b border-[#44403c] pb-1 uppercase tracking-wider">
            Voice Diagnostics
          </div>
          <div className="flex justify-between">
            <span className="text-[#a8a29e]">WS:</span>
            <span>{isConnected ? "Connected" : "Disconnected"}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-[#a8a29e]">AudioContext:</span>
            <span>{audioContextRef.current?.state || "none"}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-[#a8a29e]">Turn:</span>
            <span>#{turnNumber}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-[#a8a29e]">Voxi Speaking:</span>
            <span className={isAgentTurn ? "text-amber-400" : "text-[#78716c]"}>{isAgentTurn ? "YES" : "NO"}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-[#a8a29e]">Candidate Speaking:</span>
            <span className={isUserSpeaking ? "text-emerald-400" : "text-[#78716c]"}>{isUserSpeaking ? "YES" : "NO"}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-[#a8a29e]">Noise Floor:</span>
            <span>{Math.round(noiseFloorRef.current)}</span>
          </div>
          {liveSttText && (
            <div className="border-t border-[#44403c] pt-1 text-[9px] text-[#d6d3d1] truncate">
              STT: {liveSttText}
            </div>
          )}
        </aside>
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
