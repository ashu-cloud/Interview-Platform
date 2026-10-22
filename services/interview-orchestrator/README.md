# Interview Orchestrator (`session-runtime`)

The `interview-orchestrator` is the core real-time voice microservice of VoxAssess. It terminates full-duplex WebSocket connections from candidate browsers, coordinates streaming STT, LLM, and TTS pipelines, manages distributed session state in Redis, speculatively prefetches rubric context, dynamically guides interview questions via adaptive competency steering, and dispatches conversational events to the PostgreSQL Outbox.

---

## Architectural Role & Concurrency Model

```
Browser (16kHz PCM Binary Frames)
        │
        ▼ (Full-Duplex WebSocket: /ws/interview/{id}?token={token})
+───────────────────────────────────────────────────────────────────────+
| interview-orchestrator (session-runtime :8002 / Host: :8004)          |
|                                                                       |
|  Goroutine A: Main WebSocket Read Loop                                |
|  ├── Validates token against Redis (fast-path) or Postgres            |
|  ├── Acquires Redis distributed lock: session:lock:{id} (30s TTL)    |
|  └── Forwards binary PCM audio frames (Tag 0x01) to Goroutine B       |
|                                                                       |
|  Goroutine B: STT Streaming & Barge-In Detector                       |
|  ├── Streams audio to Deepgram Nova-2 WebSocket                       |
|  ├── Emits interim transcripts -> triggers speculative prefetcher     |
|  ├── Detects barge-in speech -> cancels in-flight turn context        |
|  └── On speech end -> forwards final transcript to Goroutine C        |
|                                                                       |
|  Goroutine C: Assessor Response Generator (Critical Path)            |
|  ├── Awaits initialGreetingDone barrier (race-free greeting)          |
|  ├── Reads prefetched context from Redis (p50: 1.15ms)                |
|  ├── Updates 3-factor adaptive competency matrix                      |
|  ├── Synthesizes system prompt with dynamic CURRENT FOCUS directive   |
|  └── Initiates streaming LLM generation (Groq Llama 3.3 70B)          |
|                                                                       |
|  Goroutine D: Sentence Boundary Segmenter & Pipelined TTS             |
|  ├── Buffers LLM tokens until sentence terminal punctuation ('.','?','!')|
|  ├── Dispatches Sentence 1 to ElevenLabs Turbo streaming TTS          |
|  └── Streams binary MP3 chunks (Tag 0x02) to browser (p50: 273.7ms)   |
+───────────────────────────────────────────────────────────────────────+
        │
        ├──► Turn Events ───────────► Postgres [outbox_turns]
        └──► Completed Trigger ─────► Postgres [outbox_completed]
```

---

## Key Subsystems & Packages

### 1. Speculative RAG Prefetcher (`pkg/rag`)
- **Trigger**: Non-blocking async pgvector search launched on early interim STT events.
- **Embeddings**: Gemini `gemini-embedding-001` or OpenAI `text-embedding-3-small` (768-dim normalized).
- **Latency Split**: Embedding (p50: 54.3ms) + pgvector query (p50: 10.5ms). Total cold retrieval: 64.8ms.
- **Cache Hit**: Returns context from Redis in **1.15ms** (p50), bypassing embedding and database overhead on the prompt assembly critical path.
- **Fallback**: Strict 150ms timeout bounds cold queries; falls back cleanly to static blueprint questions and rubric directives on timeout.

### 2. Adaptive Competency Engine (`internal/adaptive`)
- Evaluates candidate utterances against rubric criteria using a 3-factor mathematical formulation:
  1. **Topic Coverage** (monotonic max similarity).
  2. **Demonstrated Competence** (capped at 0.50 on single mentions; requires $\ge 2$ substantive turns).
  3. **Recency Decay** (exponential decay: $0.85^{\Delta \text{turns}}$).
- Identifies the maximum deficit criterion and injects dynamic `CURRENT FOCUS:` directives into the prompt.

### 3. Pipelined Sentence Synthesis & Audio Streaming
- Employs streaming LLM token buffering with regex sentence segmentation.
- Drops candidate dead-air delay from 753.9ms to **273.7ms (p50)** by dispatching Sentence 1 to TTS while Sentence 2 continues streaming.

### 4. Barge-In Interruption & History Truncation
- Detects non-filler candidate speech while the assessor is speaking (`tc.isSpeaking()`).
- Cancels active turn context (`turnCancel()`), immediately terminating in-flight LLM HTTP connections and ElevenLabs TTS audio streams.
- Synchronizes the atomic turn counter and transmits `ai_interrupted` with `Turn: int(curTurn.Load())`.
- Truncates `ConversationHistory` strictly to sentences fully delivered to the client, preventing prompt hallucination.

### 5. Diagnostics & Observability (`/debug/voice`)
- Exposes `GET /debug/voice` health probe (gated by `VOICE_DEBUG=true` or `X-Internal-Key`).
- Sequentially validates Redis, Platform API, Deepgram STT, Groq LLM streaming, and ElevenLabs audio synthesis.

---

## Environment Variables

| Variable | Default | Purpose |
| :--- | :--- | :--- |
| `PORT` | `8002` | Listen port (mapped to `:8004` in Docker Compose) |
| `DATABASE_URL` | — | PostgreSQL connection string (for outbox persistence) |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection for state, locking, and prefetch cache |
| `JOB_CANDIDATE_SERVICE_URL` | `http://localhost:8001` | Platform service URL for token validation |
| `INTERNAL_API_KEY` | — | Shared key for inter-service authentication |
| `GROQ_API_KEY` | — | Groq API key for LLaMA 3.3 70B streaming inference |
| `DEEPGRAM_API_KEY` | — | Deepgram API key for streaming WebSocket STT |
| `ELEVENLABS_API_KEY` | — | ElevenLabs API key for streaming TTS |
| `ELEVENLABS_VOICE_ID` | `21m00Tcm4TlvDq8ikWAM` | Voice identifier (Rachel default) |
| `EMBEDDING_API_KEY` | — | Gemini or OpenAI key for 768-dim prefetch embeddings |
| `EMBEDDING_BASE_URL`| — | `https://generativelanguage.googleapis.com/v1beta/openai/` for Gemini |
| `EMBEDDING_MODEL`   | `gemini-embedding-001` | Embedding model name |
| `EMBEDDING_DIMS`    | `768` | Embedding dimensionality |
| `VOICE_DEBUG`       | `true` | Enables `/debug/voice` diagnostics endpoint |
| `INTERVIEW_DEMO_MODE` | `false` | Enables offline demo mode without paid API keys |

---

## Running & Testing

```bash
# Run unit and subsystem tests
go test -v ./...
```
