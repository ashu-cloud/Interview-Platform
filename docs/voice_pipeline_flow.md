# Real-Time Voice Pipeline Architecture & Flow

This document details the end-to-end real-time conversational voice pipeline implemented in `services/interview-orchestrator` and `frontend/app/interview/[token]/page.tsx`. It traces audio capture, pre-roll buffering, streaming transcription, speculative context prefetching, streaming LLM dialogue generation, sentence-boundary pipelined text-to-speech (TTS), low-latency barge-in interruption, and diagnostic observability.

---

## 1. High-Level Pipeline Sequence

```mermaid
sequenceDiagram
    autonumber
    actor Candidate as Candidate (Microphone)
    participant UI as Browser Stage UI (page.tsx)
    participant WS as interview-orchestrator (WebSocket)
    participant STT as Deepgram Nova-2 (Streaming STT)
    participant SpecRAG as Speculative Prefetcher (Redis + pgvector)
    participant LLM as Groq / Llama 3.3 70B (Streaming LLM)
    participant TTS as ElevenLabs Turbo (Streaming TTS)
    actor Speaker as Candidate (Headphones/Speaker)

    Note over Candidate,UI: 1. Audio Conditioning & Capture
    Candidate->>UI: Candidate Speaks ("I optimized Kafka consumer lag...")
    UI->>UI: Box Anti-Aliasing Downsampler (44.1/48kHz -> 16kHz Linear PCM)
    UI->>UI: Flush ~250ms Pre-Roll Ring Buffer (prevents syllable clipping)
    UI->>WS: Binary Frame Tag 0x01 + Raw 16kHz PCM Audio Chunks
    WS->>STT: Stream Audio Buffer to Deepgram

    Note over STT,SpecRAG: 2. Speculative Context Prefetching
    STT-->>WS: Interim Transcript ("I optimized Kafka...")
    WS->>SpecRAG: TriggerPrefetch(sessionID, turnNum, interimText)
    Note over SpecRAG: Background async embedding (Gemini) + pgvector lookup (Non-blocking)

    Note over WS,LLM: 3. Turn Finalization & Prompt Assembly
    Candidate->>UI: Silence / Speech Ends (VAD Trigger)
    STT-->>WS: Final Transcript ("I optimized Kafka consumer lag using batch rebalancing.")
    WS->>UI: JSON {"type": "transcript", "role": "participant", "content": "..."}
    WS->>SpecRAG: GetContext(sessionID, turnNum, finalText)
    SpecRAG-->>WS: Return Prefetched Context (Cache Hit <1.5ms)
    WS->>LLM: Stream Prompt (System Directives + History + Prefetched Context)

    Note over LLM,Speaker: 4. Pipelined Sentence Synthesis (2.8x Speedup)
    LLM-->>WS: Stream Token Stream ("That is an interesting distributed systems challenge.")
    Note over WS: Sentence Segmenter detects terminal punct ('.', '!', '?')
    WS->>TTS: Stream Sentence 1 ("That is an interesting distributed systems challenge.")
    TTS-->>WS: MP3 Audio Chunk Stream (Sentence 1)
    WS-->>UI: Binary Frame Tag 0x02 + 24kHz MP3 Audio Chunk
    UI->>Speaker: Low-Latency Playback Begins (p50: 273.7ms dead air!)

    LLM-->>WS: Stream Sentence 2 ("How did you configure consumer heartbeat intervals?")
    WS->>TTS: Stream Sentence 2
    TTS-->>WS: MP3 Audio Chunk Stream (Sentence 2)
    WS-->>UI: Binary Frame Tag 0x02 + MP3 Audio Chunk 2
    UI->>Speaker: Continuous Seamless Audio Playback
```

---

## 2. Audio Ingestion Quality & Conditioning

To eliminate speech recognition failures and acoustic degradation:

### A. Anti-Aliasing Box Downsampler
Standard browser microphones operate at 44.1kHz or 48kHz, whereas speech-to-text models require 16,000Hz mono audio. Naive decimation (dropping samples) introduces Nyquist foldover distortion, degrading speech intelligibility. VoxAssess implements a **linear box filter downsampler** inside the Web Audio ScriptProcessor/AudioWorklet pipeline that computes weighted local sample averages before encoding into 16-bit linear PCM.

### B. ~250ms Circular Audio Pre-Roll Buffer
When a candidate starts speaking, Voice Activity Detection (VAD) algorithms require 150ms–250ms of audio energy above threshold to transition into speech mode. Without buffering, the candidate's opening syllable (e.g. *"Well..."*, *"In..."*) is permanently lost.
- The client maintains a rotating **circular ring buffer** holding the latest ~250ms of audio frames.
- Upon speech detection, the pre-roll buffer is immediately flushed ahead of the new audio stream, guaranteeing Deepgram receives complete phonetic utterances from the very first consonant.

---

## 3. Pipelined Sentence Synthesis (2.8x Dead-Air Reduction)

Conventional voice bots wait for the LLM to complete its entire response before initiating speech synthesis:

$$\text{Latency}_{\text{conventional}} = \text{TTFT} + (N_{\text{tokens}} \times \text{TPOT}) + \text{TTS}_{\text{synth}}$$

For a 50-token response, this introduces **750ms–1,200ms of dead air**.

### VoxAssess Pipelining Strategy
1. The orchestrator consumes streaming tokens from Groq (LLaMA 3.3 70B).
2. A deterministic tokenizer detects terminal punctuation boundaries (`.`, `!`, `?`, `\n`) while rejecting abbreviations (`e.g.`, `i.e.`, `vs.`, `Dr.`).
3. Sentence 1 (typically 12–18 tokens, ~250ms) is immediately dispatched to ElevenLabs Turbo streaming TTS.
4. While ElevenLabs streams back audio for Sentence 1, Groq continues generating Sentence 2 in the background.

#### Empirical Measurements:
- **Sequential Baseline**: $\text{p50} = 753.9\text{ms}$.
- **Pipelined Sentence Synthesis**: $\text{p50} = 273.7\text{ms}$.
- **Net Improvement**: **480.2ms reduction in dead air**.

---

## 4. Live Barge-In Interruption & Ledger Truncation

Natural conversational flow requires candidates to be able to interrupt the assessor without causing turn count desynchronization or prompt corruption.

```mermaid
flowchart TD
    CandidateSpeaks["Candidate Starts Speaking during Assessor Playback"] --> VADDetects["Browser Web Audio / VAD Detects Speech Energy"]
    VADDetects --> HaltLocal["Immediately Halt AudioContext Playback & Clear Audio Queue"]
    HaltLocal --> SendInterrupt["Send WS JSON: {'type': 'interrupt'}"]

    SendInterrupt --> BackendCancel["Backend Cancels Active turn Context (turnCancel())"]
    BackendCancel --> AbortStreams["Abort Active Groq Stream & ElevenLabs TTS HTTP Connection"]
    BackendCancel --> TruncateHistory["Truncate AI Message History to Delivered Sentences Only"]
    BackendCancel --> SyncTurn["Synchronize Atomic Turn Counter (curTurn.Load())"]

    SyncTurn --> SendInterrupted["Send WS JSON: {'type': 'ai_interrupted', 'turn': N}"]
    SendInterrupted --> ReArm["Re-arm STT Pipeline for Candidate's New Utterance"]
```

### Atomic Turn Synchronization
When an interruption occurs, `commitTurn` in `websocket.go` broadcasts `ai_interrupted` with `Turn: int(curTurn.Load())` rather than speculative targets. This ensures the client's local audio queue synchronization counter remains aligned with the backend, preventing the browser from mistakenly discarding subsequent assessor responses.

---

## 5. Diagnostic Observability & Telemetry

### A. Health Probe Endpoint (`GET /debug/voice`)
The orchestrator exposes an authenticated diagnostics endpoint (`http://localhost:8004/debug/voice`) gated by `VOICE_DEBUG=true` or `X-Internal-Key`. It tests each layer of the voice pipeline in isolation:
- **Redis Connectivity**: Measures round-trip ping latency.
- **Platform API**: Probes HTTP connectivity to `job-candidate-service`.
- **Deepgram STT**: Probes WebSocket handshake.
- **Groq LLM**: Probes streaming token generation latency.
- **ElevenLabs TTS**: Probes binary audio synthesis.

### B. In-Browser Telemetry HUD (`?debug=1`)
Opening any interview stage with `?debug=1` (e.g. `/interview/token_xyz?debug=1`) displays a real-time diagnostics HUD overlay showing:
- WebSocket connection state and latency.
- Audio buffer queue depth and playback status.
- Current server vs client turn indices.
- Deepgram STT frame rate and speech detection flags.
