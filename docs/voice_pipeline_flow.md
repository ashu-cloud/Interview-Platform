# Real-Time Voice Pipeline Architecture & Flow

This document details the end-to-end real-time conversational voice pipeline implemented in `services/interview-orchestrator`. It traces audio capture, streaming transcription, speculative context prefetching, streaming LLM dialogue generation, sentence-boundary pipelined text-to-speech (TTS), and low-latency barge-in interruption handling.

---

## 1. High-Level Pipeline Diagram

```mermaid
sequenceDiagram
    autonumber
    actor Candidate as Candidate (Microphone)
    participant WS as session-runtime (WebSocket)
    participant STT as Deepgram Nova-2 (Streaming STT)
    participant SpecRAG as Speculative Prefetcher (Redis + pgvector)
    participant LLM as Groq / Llama 3.3 70B (Streaming LLM)
    participant TTS as ElevenLabs Turbo (Streaming TTS)
    actor Speaker as Candidate (Headphones/Speaker)

    Note over Candidate,Speaker: Candidate Speaking Phase
    Candidate->>WS: Raw PCM Audio Frames (16kHz, 16-bit, mono) via WS Binary
    WS->>STT: Stream Audio Buffer
    STT-->>WS: Interim Transcript Event ("I worked on...")
    WS->>SpecRAG: TriggerPrefetch(sessionID, turnNum, interimText)
    Note over SpecRAG: Background async embedding & pgvector lookup (Non-blocking)

    Candidate->>WS: Silence / Speech Ends (VAD Trigger)
    WS->>STT: Flush Audio Buffer
    STT-->>WS: Final Transcript Event ("I worked on Kafka consumer lag optimization.")

    Note over WS,LLM: Turn Finalization & Critical Path (<300ms target)
    WS->>SpecRAG: GetContext(sessionID, turnNum, finalText)
    SpecRAG-->>WS: Return Prefetched Context (Cache Hit <1.5ms)
    WS->>LLM: Stream Prompt (System Directives + History + Prefetched Context)

    Note over LLM,Speaker: Pipelined Sentence Synthesis
    LLM-->>WS: Stream Token Stream ("That is an interesting challenge.")
    Note over WS: Sentence Boundary Segmenter detects punct (".", "!", "?")
    WS->>TTS: Stream Sentence 1 ("That is an interesting challenge.")
    TTS-->>WS: MP3 Audio Chunk Stream (Sentence 1)
    WS-->>Speaker: Binary Audio Message (Chunk 1) - Playback Begins!

    LLM-->>WS: Stream Sentence 2 ("How did you configure consumer group rebalance timeouts?")
    WS->>TTS: Stream Sentence 2
    TTS-->>WS: MP3 Audio Chunk Stream (Sentence 2)
    WS-->>Speaker: Binary Audio Message (Chunk 2)
```

---

## 2. Component Breakdown & Protocols

### A. Audio Ingestion & Frame Streaming
- **Transport**: Persistent full-duplex WebSocket connection (`/ws/interview/{id}?token={token}`).
- **Audio Specification**: Linear PCM, 16,000 Hz sample rate, 16-bit depth, single-channel (mono).
- **Framing**: The browser client captures audio chunks via Web Audio API (`AudioWorkletNode`) every 100ms–250ms and streams them as raw binary WebSocket messages.
- **WebSocket Protocol**:
  - `BinaryMessage`: Raw PCM audio bytes forwarded immediately to the active STT session.
  - `TextMessage`: JSON control payloads (`session_start`, `ping`, `client_ready`).

### B. Streaming Speech-to-Text (STT)
- **Engine**: Deepgram Nova-2 Streaming API via WebSocket connection.
- **Interim vs. Final Events**:
  - `is_final == false`: Interim hypothesis emitted every ~150ms–300ms as speech progresses. Triggers the speculative retrieval prefetcher.
  - `is_final == true`: Speech utterance terminated by voice activity detection (VAD). Sent to `finalTranscriptChan` to initiate assessor dialogue generation.
- **Resilience & Auto-Reconnect**: If the Deepgram WebSocket drops mid-turn due to network timeout or provider reset, `websocket.go` detects the closed socket, instantiates a fresh connection with exponential backoff, and resumes streaming without aborting the candidate's interview session.

### C. Sentence Boundary Segmentation & Pipelined TTS
Conventional voice architectures wait for the LLM to generate an entire paragraph before dispatching text to a TTS engine, introducing 700ms–1,200ms of dead-air delay.

VoxAssess implements **Pipelined Sentence Synthesis**:
1. An asynchronous stream reader processes LLM tokens as they arrive over HTTP streaming.
2. A deterministic tokenizer detects terminal punctuation boundaries (`.`, `!`, `?`, `\n`) while filtering common abbreviations (`e.g.`, `i.e.`, `vs.`, `Dr.`, `Mr.`).
3. As soon as the first sentence boundary is resolved (typically 12–18 tokens, ~250ms), it is immediately dispatched to ElevenLabs Turbo streaming TTS.
4. Subsequent sentences are dispatched concurrently or enqueued sequentially, overlapping LLM generation of sentence $N+1$ with TTS synthesis of sentence $N$.

#### Measured Impact:
- **Baseline (Sequential Completion-Then-TTS)**: $\text{p50} = 753.9\text{ms}$.
- **Pipelined Sentence Synthesis**: $\text{p50} = 273.7\text{ms}$.
- **Speedup**: **2.8x reduction in candidate dead-air delay** (saving $480.2\text{ms}$).

---

## 3. Barge-In Interruption & History Truncation

A critical requirement of natural human dialogue is the ability to interrupt the speaker without desynchronizing state.

```mermaid
flowchart TD
    CandidateSpeaks["Candidate Starts Speaking during AI Audio Playback"] --> STTDetects["Deepgram Emits Non-Empty Interim Transcript"]
    STTDetects --> FilterFillers{"Is Filler Only?<br/>('um', 'uh', 'ah')"}
    FilterFillers -- Yes --> Ignore["Ignore Interim<br/>Continue Playing AI Speech"]
    FilterFillers -- No --> TriggerInterrupt["Trigger Barge-In Interruption"]

    TriggerInterrupt --> CancelTurnCtx["Cancel active turn Context (turnCancel())"]
    TriggerInterrupt --> StopTTS["Close TTS Audio Stream & Signal Browser UI"]
    TriggerInterrupt --> SendWS["Send WS JSON: {'type': 'ai_interrupted', 'turn': N}"]
    TriggerInterrupt --> TruncateHistory["Truncate AI Message History to Delivered Sentences Only"]

    TruncateHistory --> SaveRedis["Persist Clean State to Redis"]
```

### Detailed Mechanics:
1. **Filler Filtering (`isFillerOnly`)**:
   Candidate hesitations such as *"um"*, *"uh"*, *"ah"*, or *"hmm"* do not trigger interruptions. The assessor continues speaking uninterrupted.
2. **Turn Context Cancellation**:
   When substantive speech is detected while `tc.isSpeaking() == true`, the active turn context is cancelled via `turnCancel()`. In-flight LLM HTTP requests and TTS streaming goroutines abort immediately.
3. **Audio Playback Cutoff**:
   The orchestrator sends an `ai_interrupted` WebSocket control frame to the browser. The frontend immediately halts its Web Audio buffer playback and clears the speaker queue.
4. **Delivered Text Truncation**:
   Only sentences that were completely synthesized and transmitted to the client prior to interruption are appended to `sess.ConversationHistory`. This prevents prompt hallucination in subsequent turns where the LLM assumes the candidate heard points that were cut off mid-sentence.

---

## 4. Sequence Barriers & Concurrency Controls

### Initial Greeting Completion Barrier
- **The Problem**: If a candidate speaks immediately upon connecting (or emits ambient noise), their audio could trigger Turn 2 while Turn 1 (the initial interviewer welcome greeting) is still generating or persisting, causing out-of-order turns and corrupted database history.
- **The Solution**: An explicit synchronization barrier (`initialGreetingDone chan struct{}`) blocks Goroutine C (the AI response generator) from processing participant input until the initial greeting turn has finalized its Kafka publication and database state save.

### Distributed Session Locking
- To prevent split-brain dual-tab sessions, `session-runtime` acquires a Redis distributed lock (`session:lock:{id}`) with a 30-second TTL upon WebSocket upgrade.
- An internal heartbeat goroutine refreshes the lock every 15 seconds. If a secondary connection attempts to attach to the same session, it is rejected with `409 Conflict`.
