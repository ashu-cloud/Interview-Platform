# VoxAssess — Low-Latency Adaptive Voice Assessment Platform

VoxAssess is an enterprise-grade AI voice platform designed for automated, structured technical and behavioral candidate assessments. It features real-time bidirectional audio streaming, speculative RAG prefetching, adaptive competency steering, and asynchronous rubric evaluations with verbatim evidence attribution.

---

## Architecture Highlights

```
+--------------------------------------------------------------------------------------------------+
| TRACK 1: REAL-TIME VOICE PIPELINE (Sub-second critical path, in-memory, non-blocking)           |
|                                                                                                  |
| [Candidate Browser] <=== Full-Duplex PCM WebSocket ===> [interview-orchestrator]                |
|                                                                 |                                |
|        +-------------------------+------------------------------+--------------------+           |
|        |                         |                              |                    |           |
|  [Streaming STT]       [Speculative Prefetch]            [Streaming LLM]       [Streaming TTS]   |
| (Deepgram Nova-2)      (pgvector + Redis Cache)        (Groq Llama 3.3 70B)  (ElevenLabs Turbo)  |
|                                                                 |                                |
|        +--------------------------------------------------------+                                |
|        v (Pipelined Sentence Synthesis: p50 = 273.7ms dead-air delay)                            |
|        v (Barge-in Interruption: Cancels in-flight turns & truncates history to delivered audio) |
+--------------------------------------------------------------------------------------------------+
                                  |
               (Non-blocking async outbox / Kafka events)
                                  |
                                  v
+--------------------------------------------------------------------------------------------------+
| TRACK 2: ASYNC PROJECTION & GROUNDED EVALUATION (ACID, idempotent, concurrency-bounded)          |
|                                                                                                  |
| Event Bus (Dual-Mode):                                                                           |
|   - Local Dev: Apache Kafka 3.7 (session.turns, session.completed)                               |
|   - Cloud/Low-RAM: PostgreSQL Outbox (FOR UPDATE SKIP LOCKED, zero broker overhead)             |
|                                                                                                  |
| Workers:                                                                                         |
|   - turn-projector: Ingests turns idempotently into session_turns                                |
|   - feedback-worker: Bounded-concurrent (N=5) evaluation with verbatim quote attribution        |
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
While the candidate speaks, interim STT hypotheses trigger background embedding generation (`text-embedding-3-small`) and nearest-neighbor vector search over blueprint rubric criteria. Results are cached in Redis (`rag:prefetch:{session}:{turn}`).
- **Prefetched Cache Hit Retrieval**: **1.15ms** (p50) on critical path.
- **Cold Query Fallback**: 64.8ms (p50) bounded by a strict 150ms timeout that falls back to static rubric criteria on timeout or network jitter.

### 3. Adaptive 3-Factor Competency Steering
Rather than executing static question lists, the orchestrator computes live candidate competency deficits across three mathematical dimensions:
- **Topic Coverage**: Monotonic historical max similarity.
- **Demonstrated Competence**: Capped at $0.50$ on single mentions; requires $\ge 2$ substantive turns to reach full score.
- **Recency Decay**: Exponential decay ($0.85^{\Delta \text{turns}}$) to prevent unaddressed criteria from being starved of follow-up questions.
- Injects dynamic `CURRENT FOCUS:` directives into the interviewer system prompt.

### 4. Evidence Grounding & Verbatim Quote Attribution
Automated evaluation reports reject hallucinated evidence quotes. The `IsSubstantiveEvidence` contract requires $\ge 3$ substantive words or $\ge 15$ characters, filters conversational stop-tokens (*"yes"*, *"sure"*, *"okay"*), and verifies exact substring matching against the candidate transcript, dropping affirmative false accepts from **80.0% to 0.0%**.

---

## Specialized Architecture Documentation

Detailed sequence diagrams, failure modes, code citations, and empirical measurements:

- **[Real-Time Voice Pipeline Flow](docs/voice_pipeline_flow.md)**: PCM audio framing, Deepgram streaming, pipelined TTS, barge-in interruption, sequence barriers.
- **[Speculative RAG Retrieval & Prefetch Flow](docs/speculative_retrieval_flow.md)**: End-to-end prefetch trace, pgvector queries, Redis caching, 150ms budget, and semantic drift mitigation.
- **[Adaptive Competency Steering](docs/adaptive_competency_steering.md)**: 3-factor formulation, deficit calculation, and dynamic prompt injection.
- **[Asynchronous Evaluation & Grounding Pipeline](docs/evaluation_and_grounding_flow.md)**: Completeness barriers, $N=5$ scatter-gather grading, JSON repair, and quote verification.
- **[System Architecture Guide](docs/architecture.md)**: Dual-track architecture, database schema, Redis keys, and deployment topologies.
- **[REST & WebSocket API Reference](docs/api.md)**: Endpoints, authentication, and payload schemas.

---

## Services & Port Topology

| Service | Directory | Port | Protocol | Description |
| :--- | :--- | :--- | :--- | :--- |
| **Frontend Gateway** | `frontend/` | `:8080` | HTTP / WS | Next.js 14 App Router UI & Nginx Reverse Proxy |
| **Platform Service** | `services/job-candidate-service` | `:8001` | HTTP REST | Auth, tracks, blueprints, prompts, rubric vectors, candidate invites |
| **Session Runtime** | `services/interview-orchestrator` | `:8002` | WebSocket | Real-time voice orchestrator, STT, LLM, TTS, speculative RAG prefetch |
| **Evaluation Worker** | `services/feedback-worker` | — | Kafka / Outbox | Grounded transcript assessment, rubric grading, JSON repair |
| **Turn Projector** | `services/turn-projector` | — | Kafka / Outbox | Idempotent turn ingestion into relational `session_turns` |
| **Notification Service** | `services/notification-service` | — | Kafka / Outbox | Candidate invite and completion notifications via Resend |
| **PostgreSQL 16** | `migrations/` | `:5432` | SQL / pgvector | Relational data store with 768-dim vector embeddings |
| **Redis 7.2** | — | `:6379` | RESP | Distributed mutex locks, session state, and RAG prefetch cache |
| **Apache Kafka 3.7** | — | `:9092` | Kafka (KRaft) | Event broker for turns and completion events |

---

## Quick Start (Docker Compose)

### 1. Prerequisites
- Docker Engine 24+ & Docker Compose v2
- Go 1.22+ (for running local benchmarks)

### 2. Configure Environment
```bash
cp .env.example .env
```
> **Offline Demo Mode**: Set `INTERVIEW_DEMO_MODE=true` in `.env` to execute the full interactive voice assessment pipeline locally without requiring paid API keys for Deepgram, Groq, or ElevenLabs.

### 3. Launch Services
```bash
docker compose up -d
```

### 4. Access the Web Dashboard
Navigate to `http://localhost:8080` to log in, create an assessment track and blueprint, generate a candidate invite link, and conduct an automated voice interview.

---

## Verification & Empirical Benchmarks

The platform includes an automated benchmark suite under `bench/`:

```bash
# Run the fair comparison microbenchmark suite (Pipelining, Prefetch, Scatter-Gather)
cd bench
go test -v -run=TestFairComparison

# Run the 30-turn speculative prefetch replay harness
go test -v -run=TestSpeculativeRetrievalReplayHarness

# Run all project benchmark suites
go run run_benchmarks.go
```
