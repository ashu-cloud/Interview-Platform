# Interview Orchestrator (`session-runtime`)

The `interview-orchestrator` is the core real-time voice microservice of VoxAssess. It terminates full-duplex WebSocket connections from candidate browsers, coordinates streaming STT, LLM, and TTS pipelines, manages distributed session state in Redis, speculatively prefetches rubric context, dynamically guides interview questions via adaptive competency steering, and dispatches conversational events to Kafka or PostgreSQL Outbox.

---

## Architectural Role & Concurrency Model

```
Browser (Web Audio PCM)
        │
        ▼ (Full-Duplex WebSocket: /ws/interview/{id})
+───────────────────────────────────────────────────────────────────────+
| interview-orchestrator (session-runtime)                              |
|                                                                       |
|  Goroutine A: Main WebSocket Read Loop                                |
|  ├── Validates token against Redis (fast-path) or Postgres            |
|  ├── Acquires Redis distributed lock: session:lock:{id} (30s TTL)    |
|  └── Forwards binary PCM audio frames to Goroutine B                  |
|                                                                       |
|  Goroutine B: STT Streaming & Barge-In Detector                       |
|  ├── Streams audio to Deepgram Nova-2 WebSocket                       |
|  ├── Emits interim transcripts -> triggers speculative prefetcher     |
|  ├── Detects barge-in speech -> cancels in-flight turn context        |
|  └── On speech end -> forwards final transcript to Goroutine C        |
|                                                                       |
|  Goroutine C: Assessor Response Generator (Critical Path)            |
|  ├── Awaits initialGreetingDone barrier                               |
|  ├── Reads prefetched context from Redis (p50: 1.15ms)                |
|  ├── Updates 3-factor adaptive competency matrix                      |
|  ├── Synthesizes system prompt with dynamic CURRENT FOCUS directive   |
|  └── Initiates streaming LLM generation (Groq Llama 3.3 70B)          |
|                                                                       |
|  Goroutine D: Sentence Boundary Segmenter & Pipelined TTS             |
|  ├── Buffers LLM tokens until sentence terminal punctuation ('.','?','!')|
|  ├── Dispatches Sentence 1 to ElevenLabs Turbo streaming TTS          |
|  └── Streams binary MP3 chunks to browser (p50: 273.7ms dead air)    |
+───────────────────────────────────────────────────────────────────────+
        │
        ├──► Turn Events ───────────► Kafka [session.turns] / Postgres [outbox_turns]
        └──► Completed Trigger ─────► Kafka [session.completed] / Postgres [outbox_completed]
```

---

## Key Subsystems & Packages

### 1. Speculative RAG Prefetcher (`pkg/rag`)
- **Trigger**: Non-blocking async pgvector search launched on early interim STT events.
- **Deduplication**: `sync.Map` locks turn keys to prevent redundant execution across multiple interim frames.
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
- Truncates `ConversationHistory` strictly to sentences fully delivered to the client, preventing prompt hallucination.

---

## Environment Variables

| Variable | Default | Purpose |
| :--- | :--- | :--- |
| `PORT` | `8002` | HTTP / WebSocket server listen port |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection for state, locking, and prefetch cache |
| `KAFKA_BROKERS` | `""` | Comma-separated broker addresses (if empty, uses Postgres Outbox) |
| `DEEPGRAM_API_KEY` | — | Deepgram Nova-2 streaming STT key |
| `GROQ_API_KEY` | — | Groq Llama 3.3 70B LLM key |
| `ELEVENLABS_API_KEY` | — | ElevenLabs Turbo streaming TTS key |
| `ELEVENLABS_VOICE_ID` | `21m00Tcm4TlvDq8ikWAM` | Voice ID for synthetic voice persona |
| `INTERVIEW_DEMO_MODE` | `false` | When true, runs deterministic offline mock providers without API keys |

---

## Running & Testing

```bash
# Run unit tests
go test -v ./...

# Run voice experience latency tests
go test -v -run=TestOrchestratorPipeliningAndInterruption ./internal/handler/...

# Run sentence segmentation benchmark
go test -bench=BenchmarkSentenceSegmentation ./internal/handler/...

# Run adaptive engine ablation simulation
go test -v -run=TestAdaptiveEngineAblation ./internal/adaptive/...
```
