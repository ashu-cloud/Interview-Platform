# VoxAssess — Low-Latency Adaptive Voice Assessment Platform

VoxAssess is an enterprise-grade AI voice platform designed for automated, structured technical and behavioral candidate assessments. It features real-time bidirectional audio streaming, speculative RAG prefetching, adaptive competency steering, and asynchronous rubric evaluations with verbatim evidence attribution.

---

## Architecture Highlights

```
+--------------------------------------------------------------------------------------------------+
| TRACK 1: REAL-TIME VOICE PIPELINE (Sub-second critical path, in-memory, non-blocking)           |
|                                                                                                  |
| [Candidate Browser] <=== Full-Duplex PCM WebSocket ===> [interview-orchestrator (:8004)]        |
|                                                                 |                                |
|        +-------------------------+------------------------------+--------------------+           |
|        |                         |                              |                    |           |
|  [Streaming STT]       [Speculative Prefetch]            [Streaming LLM]       [Streaming TTS]   |
| (Deepgram Nova-2)      (pgvector + Redis Cache)        (Groq Llama 3.3 70B)  (ElevenLabs Turbo)  |
|                                                                 |                                |
|        +--------------------------------------------------------+                                |
|        v (Pipelined Sentence Synthesis: p50 = 273.7ms dead-air delay)                            |
|        v (Barge-in Interruption: Cancels in-flight turns & truncates history to delivered audio) |
|        v (Audio Pre-Roll & Anti-Aliasing: ~250ms ring buffer prevents syllable clipping)         |
+--------------------------------------------------------------------------------------------------+
                                  |
               (Non-blocking transactional PostgreSQL outbox events)
                                  |
                                  v
+--------------------------------------------------------------------------------------------------+
| TRACK 2: ASYNC PROJECTION & GROUNDED EVALUATION (ACID, idempotent, concurrency-bounded)          |
|                                                                                                  |
| Event Bus:                                                                                       |
|   - PostgreSQL Outbox (outbox_turns, outbox_completed polled via FOR UPDATE SKIP LOCKED)         |
|   - Zero broker overhead: No Apache Kafka required for local development or cloud deployment     |
|                                                                                                  |
| Background Processors:                                                                           |
|   - Embedded Turn Projector (in job-candidate-service): Ingests turns into session_turns         |
|   - feedback-worker: Bounded-concurrent (N=5) scatter-gather rubric evaluation with evidence     |
|   - notification-service: Event-driven alerts and email dispatch via Resend                      |
+--------------------------------------------------------------------------------------------------+
```

---

## Core Engineering Innovations

### 1. Pipelined Sentence Synthesis (2.8x Speedup)
Conventional voice bots await full LLM token completion (45+ tokens) before synthesizing speech, causing 750ms–1,200ms of dead air. VoxAssess segments incoming LLM streams at terminal punctuation boundaries (`.`, `!`, `?`), streaming Sentence 1 to ElevenLabs Turbo while Sentence 2 continues generating.
- **Sequential Baseline**: $\text{p50} = 753.9\text{ms}$.
- **Pipelined Treatment**: $\text{p50} = 273.7\text{ms}$ (**480.2ms dead-air reduction**).

### 2. Speculative RAG Retrieval & Prefetching
While the candidate speaks, interim STT hypotheses trigger background embedding generation (Google Gemini `gemini-embedding-001` or OpenAI `text-embedding-3-small`, 768 dimensions with L2 normalization) and nearest-neighbor vector search over blueprint rubric criteria. Results are cached in Redis (`rag:prefetch:{session}:{turn}`).
- **Prefetched Cache Hit Retrieval**: **1.15ms** (p50) on critical path.
- **Cold Query Fallback**: 64.8ms (p50) bounded by a strict 150ms timeout that falls back to static rubric criteria on timeout or network jitter.

### 3. Audio Ingestion Quality & Syllable Preservation
To eliminate speech recognition failures caused by microphone initialization delay or audio aliasing:
- **Anti-Aliasing Box Downsampler**: Downsamples browser input to 16kHz PCM using linear box filtering, preventing Nyquist frequency foldover.
- **250ms Circular Pre-Roll Buffer**: Flushes the preceding 250ms of audio upon speech detection so initial phonemes and syllables are never clipped.
- **Atomic Barge-In Synchronization**: Interruption signals synchronize atomic turn counters between client and server, preventing post-interruption audio chunk drops.

### 4. Adaptive 3-Factor Competency Steering
Rather than executing static question lists, the orchestrator computes live candidate competency deficits across three mathematical dimensions:
- **Topic Coverage**: Monotonic historical max similarity.
- **Demonstrated Competence**: Capped at $0.50$ on single mentions; requires $\ge 2$ substantive turns to reach full score.
- **Recency Decay**: Exponential decay ($0.85^{\Delta \text{turns}}$) to prevent unaddressed criteria from being starved of follow-up questions.
- Injects dynamic `CURRENT FOCUS:` directives into the interviewer system prompt.

### 5. Evidence Grounding & Verbatim Quote Attribution
Automated evaluation reports reject hallucinated evidence quotes. The `IsSubstantiveEvidence` contract requires $\ge 3$ substantive words or $\ge 15$ characters, filters conversational stop-tokens (*"yes"*, *"sure"*, *"okay"*), and verifies exact substring matching against the candidate transcript, dropping affirmative false accepts from **80.0% to 0.0%**.

---

## Services & Port Topology

| Service | Directory | Host Port | Internal Port | Protocol | Description |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Frontend Web App** | `frontend/` | `:3000` | `:3000` | HTTP / WS | Next.js 16 App Router UI (Manager Hub, Blueprint Studio, Candidate Stage) |
| **Platform Service** | `services/job-candidate-service` | `:8003` | `:8001` | HTTP REST | Auth, tracks, blueprints, public apply, candidate invites, and embedded turn projector |
| **Session Runtime** | `services/interview-orchestrator` | `:8004` | `:8002` | WebSocket / HTTP | Real-time voice orchestrator, STT, LLM, TTS, speculative RAG prefetch, `/debug/voice` probe |
| **Evaluation Worker** | `services/feedback-worker` | — | — | Outbox Worker | Grounded transcript assessment, rubric grading, JSON repair |
| **Notification Service** | `services/notification-service` | — | — | Outbox Worker | Candidate invite and completion notifications via Resend |
| **PostgreSQL 16** | `migrations/` | `:5432` | `:5432` | SQL / pgvector | Relational data store, outbox tables (`outbox_turns`, `outbox_completed`), 768-dim vector embeddings |
| **Redis 7.2** | — | `:6379` | `:6379` | RESP | Distributed mutex locks, session state, and RAG prefetch cache |
| **MinIO** | — | `:9000` / `:9001` | `:9000` / `:9001` | S3 API / Web | Object storage for archived assessment session audio |

---

## Canonical Application Routes

### Manager / Recruiter Hub
- `/console` (Primary Dashboard): Overview of active tracks, candidate dossiers, and live assessment streams.
- `/candidates`: Candidate pipeline management and dossier review.
- `/tracks`: Evaluation tracks and required skill specifications.
- `/blueprints/new`: Visual blueprint studio for question prompts and rubric criteria.
- `/report/:id`: Detailed candidate evaluation dossier with competency breakdown, composite scores, and verbatim evidence quotes.

### Candidate Portal
- `/apply/:trackId`: Public self-service application portal. Candidates apply with full name and email to receive an instant interview invite token.
- `/interview/:token`: Real-time voice interview stage with microphone diagnostics, live speech waveform visualizer, and optional telemetry overlay (`?debug=1`).

---

## Quick Start (Docker Compose)

### 1. Prerequisites
- Docker Engine 24+ & Docker Compose v2
- Go 1.22+ (for running local microbenchmarks)

### 2. Configure Environment
```bash
cp .env.example .env
```

Configure your API keys in `.env`:
```env
# Speech-to-Text & Text-to-Speech
DEEPGRAM_API_KEY=your_deepgram_api_key
ELEVENLABS_API_KEY=your_elevenlabs_api_key
ELEVENLABS_VOICE_ID=21m00Tcm4TlvDq8ikWAM

# Fast LLM Inference (Groq Cloud)
GROQ_API_KEY=your_groq_api_key
GROQ_MODEL=openai/gpt-oss-120b

# Embeddings (Google Gemini API - Free Tier)
EMBEDDING_API_KEY=your_gemini_api_key
EMBEDDING_BASE_URL=https://generativelanguage.googleapis.com/v1beta/openai/
EMBEDDING_MODEL=gemini-embedding-001
EMBEDDING_DIMS=768
```

> **Offline Demo Mode**: Set `INTERVIEW_DEMO_MODE=true` in `.env` to execute the full interactive assessment pipeline locally without requiring paid API keys for Deepgram, Groq, or ElevenLabs.

### 3. Launch Services
```bash
docker compose up -d
```

### 4. Access the Web Application
- **Hiring Manager Hub**: [http://localhost:3000/console](http://localhost:3000/console) (Log in or register to create assessment tracks and view dossiers).
- **Public Candidate Portal**: `http://localhost:3000/apply/<track_id>`.
- **MinIO Storage Console**: [http://localhost:9001](http://localhost:9001) (`minioadmin` / `minioadmin`).
- **Voice Diagnostics Endpoint**: `http://localhost:8004/debug/voice` (Probes Redis, Platform API, Deepgram, Groq, and ElevenLabs).

---

## Specialized Architecture Documentation

Detailed sequence diagrams, failure modes, code citations, and empirical measurements:

- **[Real-Time Voice Pipeline Flow](docs/voice_pipeline_flow.md)**: PCM audio framing, Deepgram streaming, pipelined TTS, barge-in interruption, sequence barriers.
- **[Speculative RAG Retrieval & Prefetch Flow](docs/speculative_retrieval_flow.md)**: End-to-end prefetch trace, pgvector queries, Redis caching, 150ms budget, and semantic drift mitigation.
- **[Adaptive Competency Steering](docs/adaptive_competency_steering.md)**: 3-factor formulation, deficit calculation, and dynamic prompt injection.
- **[Asynchronous Evaluation & Grounding Pipeline](docs/evaluation_and_grounding_flow.md)**: Completeness barriers, $N=5$ scatter-gather grading, JSON repair, and quote verification.
- **[System Architecture Guide](docs/architecture.md)**: Outbox event architecture, database schema, Redis keys, and deployment topologies.
- **[REST & WebSocket API Reference](docs/api.md)**: Endpoints, authentication, public candidate apply, `/debug/voice`, and payload schemas.

---

## Verification & Empirical Benchmarks

The platform includes an automated benchmark and test suite:

```bash
# Run unit and integration tests across all microservices
make test-platform      # Tests job-candidate-service
make test-runtime       # Tests interview-orchestrator
make test-evaluation    # Tests feedback-worker

# Run the fair comparison microbenchmark suite (Pipelining, Prefetch, Scatter-Gather)
cd bench
go test -v -run=TestFairComparison

# Run the 30-turn speculative prefetch replay harness
go test -v -run=TestSpeculativeRetrievalReplayHarness
```
