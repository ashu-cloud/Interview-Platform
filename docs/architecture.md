# Assessment Engine — System Architecture

## 1. System Overview

The **Assessment Engine** is a low-latency, real-time adaptive voice assessment platform. It conducts structured technical and behavioral evaluations over streaming audio, continuously tracks competency coverage across live conversational turns, dynamically guides LLM prompting to explore uncovered criteria, and asynchronously produces evidence-grounded evaluation reports citing exact conversation turns.

The platform is designed to operate seamlessly across two distinct runtime environments:
1. **Local Development**: Standard distributed microservices orchestrated via Docker Compose using Kafka, PostgreSQL with `pgvector`, Redis, and MinIO object storage.
2. **Resource-Constrained Cloud Deployment (Railway Free Tier)**: Consolidated 3-service deployment utilizing a lightweight PostgreSQL Outbox pattern (`FOR UPDATE SKIP LOCKED`) in place of Kafka, Neon Postgres, Upstash Redis, and Cloudflare R2 object storage.

---

## 2. Dual-Track Architecture

The system decouples real-time audio interaction from persistence and evaluation through two architectural tracks:

```
+--------------------------------------------------------------------------------------------------+
| TRACK 1: LIVE CRITICAL PATH (Sub-second, in-memory, non-blocking)                               |
|                                                                                                  |
| [Participant] <--- WebSocket ---> [session-runtime]                                              |
|                                         |                                                        |
|   +-------------------------------------+---------------------------------------+                |
|   |                  |                              |                           |                |
| [Streaming STT]  [Redis State]             [Streaming LLM]               [Streaming TTS]         |
| (Deepgram Nova-2) (Context / Locks)   (Groq / Llama / Demo Provider)   (ElevenLabs Turbo)        |
+--------------------------------------------------------------------------------------------------+
                                          |
                      (Non-blocking async fire-and-forget)
                                          |
                                          v
+--------------------------------------------------------------------------------------------------+
| TRACK 2: ASYNC PROJECTION & EVALUATION PATH (ACID, idempotent, latency-tolerant)                |
|                                                                                                  |
| Local Mode (KAFKA_BROKERS set):                                                                  |
|   session-runtime ---> Kafka [session.turns]     ---> [turn-projector]   ---> [session_turns]    |
|   session-runtime ---> Kafka [session.completed] ---> [evaluation-worker] ---> [evaluation_reports]|
|                                                                                                  |
| Deployed Outbox Mode (KAFKA_BROKERS unset):                                                      |
|   session-runtime ---> Postgres [outbox_turns]     ---> [turn-projector]   ---> [session_turns]  |
|   session-runtime ---> Postgres [outbox_completed] ---> [evaluation-worker] ---> [evaluation_reports]|
+--------------------------------------------------------------------------------------------------+
```

### Track 1: Live Critical Path
- **Sub-second latency constraint**: Time from participant speech end to assessor audio playback target is sub-second.
- **In-memory state**: During an active turn, state resides strictly in Redis (`session:state:{id}` and distributed locks `session:lock:{id}`).
- **Zero blocking database writes**: The critical WebSocket loop never awaits disk I/O or database transactions. Turn events and audio recordings are dispatched via background goroutines and buffered channels.

### Track 2: Async Projection & Evaluation Path
- **Turn Projection**: Consumes individual conversation turns, validates payload models, and performs idempotent writes into the `session_turns` table.
- **Evaluation Generation**: Upon session completion, the evaluation worker processes the transcript, generates scoring rubrics via parallel LLM evaluation, and records reports into `evaluation_reports` and `evaluation_scores`.

---

## 3. Event Bus & The PostgreSQL Outbox Pattern

To eliminate the heavy RAM requirements of running managed or self-hosted Kafka brokers in lightweight production deployments (e.g. Railway's 512MB RAM free tier limit), the engine implements a zero-broker outbox pattern that provides identical logical delivery and ordering guarantees.

### Automatic Broker Selection
The runtime and worker components detect broker configuration automatically at startup:
- `KAFKA_BROKERS` is non-empty $\rightarrow$ **Kafka Mode** (publishes to `session.turns` and `session.completed`).
- `KAFKA_BROKERS` is empty/unset $\rightarrow$ **PostgreSQL Outbox Mode** (publishes to `outbox_turns` and `outbox_completed`).

### Outbox Dequeue Mechanics
Turn projectors and evaluation workers running in Outbox mode poll the database on a configurable interval (default: 500ms):

```sql
SELECT id, session_id, org_id, turn_number, role, content, retry_count
FROM outbox_turns
WHERE status = 'pending'
ORDER BY created_at ASC, id ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;
```

#### Key Guarantees:
1. **Concurrency Safety (`FOR UPDATE SKIP LOCKED`)**: When multiple worker instances or goroutines poll concurrently, locked rows are skipped without blocking or serialization deadlocks.
2. **Idempotent Ingestion**: Projected rows write to `session_turns` with unique constraints:
   ```sql
   INSERT INTO session_turns (session_id, org_id, turn_number, role, content)
   VALUES ($1, $2, $3, $4, $5)
   ON CONFLICT (session_id, turn_number, role) DO NOTHING;
   ```
3. **Atomic State Transition**: Rows are updated to `status = 'completed'` or given an incremented `retry_count` in the same transaction or immediately upon completion.
4. **Bounded Retries & Poison Pill Defense**: Messages exceeding the max retry limit (default: 5) transition to `status = 'failed'` to avoid head-of-line blocking.

---

## 4. Data Layer Architecture

### Database Schema (PostgreSQL 16 + pgvector)

```
                    +-------------------+
                    |   organizations   |
                    +---------+---------+
                              |
            +-----------------+-----------------+
            |                                   |
    +-------v-------+                   +-------v-------+
    |     users     |                   | assessment_   |
    +---------------+                   |    tracks     |
                                        +-------+-------+
                                                |
                                        +-------v-------+
                                        |  assessment_  |
                                        |  blueprints   |
                                        +---+-------+---+
                                            |       |
            +-------------------------------+       +-------------------------------+
            |                                                                       |
    +-------v---------------+                                               +-------v---------------+
    |   blueprint_prompts   |                                               |  evaluation_criteria  |
    | (embedding vector768) |                                               | (embedding vector768) |
    +-----------------------+                                               +-----------------------+
                                                |
                                        +-------v-------+
                                        |  participants |
                                        +-------+-------+
                                                |
                                        +-------v-------+
                                        |  assessment_  |
                                        |    invites    |
                                        +-------+-------+
                                                |
                                        +-------v-------+
                                        |  assessment_  |
                                        |   sessions    |
                                        +---+-------+---+
                                            |       |
            +-------------------------------+       +-------------------------------+
            |                                                                       |
    +-------v-------+                                                       +-------v-------+
    | session_turns |                                                       |  evaluation_  |
    | (projected)   |                                                       |    reports    |
    +---------------+                                                       +-------+-------+
            |                                                                       |
    +-------v-------+                                                       +-------v-------+
    |     audio_    |                                                       |  evaluation_  |
    |  recordings   |                                                       |    scores     |
    +---------------+                                                       +---------------+
```

#### New Phase 1 Foundation Tables:
- `session_turns`: Authoritative sequential log of participant and assessor utterances per assessment session.
- `outbox_turns`: Transactional outbox buffer for conversational turn events.
- `outbox_completed`: Transactional outbox buffer for session completion triggers.
- `audio_recordings`: Metadata and object storage URIs for recorded session audio chunks.
- `document_chunks`: Chunks and embeddings for RAG-assisted contextual assessment prompts.

### Redis Caching & Synchronization Schema

| Key Format | TTL | Purpose |
|---|---|---|
| `session:state:{session_id}` | 2 hours | Active session context, conversational turn history, and competency matrix |
| `session:lock:{session_id}` | 30 seconds (15s heartbeat) | Distributed mutex preventing duplicate active WebSocket sessions |
| `invite:token:{token}` | Invite expiry | Fast-path invite validation avoiding PostgreSQL lookups during handshake |
| `rag:prefetch:{session_id}:{turn}` | 30 seconds | Pre-computed relevant knowledge chunks for subsequent LLM turns |
| `org:connections:{org_id}` | 2 hours | Soft concurrency quota counter per organization |

---

## 5. Service Topology & Deployment Mapping

### Local Development Environment (5 Go Services + Supporting Infrastructure)
- **`platform-service`** (`services/job-candidate-service`, port `8001`): Admin REST API for tracks, blueprints, participants, and invites.
- **`session-runtime`** (`services/interview-orchestrator`, port `8002`): Live WebSocket assessment engine; STT, LLM, and TTS orchestration.
- **`turn-projector`** (`services/turn-projector`): Headless worker consuming `session.turns` / `outbox_turns` into `session_turns`.
- **`evaluation-worker`** (`services/feedback-worker`): Headless worker generating async evaluations.
- **`notification-service`**: Webhook and email dispatch stub.
- **Infrastructure Containers**: PostgreSQL 16 (`pgvector`), Redis 7.2, Apache Kafka 3.7, MinIO (`assessment-audio` bucket), Nginx frontend (:8080).

### Railway Free-Tier Deployment (3-Service Consolidation)
To operate within Railway's 3-service deployment quota without sacrificing architectural boundaries:

1. **`platform` Service**:
   - Runs `platform-service` HTTP server.
   - Spawns `turn-projector` worker as a managed background goroutine.
2. **`runtime` Service**:
   - Runs `session-runtime` WebSocket server on dedicated persistent connections.
3. **`evaluation` Service**:
   - Runs `evaluation-worker` and `notification-service` background processors.
4. **Cloud Managed Backends**:
   - Database: Neon Serverless PostgreSQL with `pgvector`.
   - Redis: Upstash Redis.
   - Storage: Cloudflare R2 (S3 API compatible).
   - Frontend: Cloudflare Pages static web application.

---

## 6. Security & Inter-Service Communication

1. **User / Assessment Author Authentication**:
   - Bearer JWT tokens signed with HS256 using `JWT_SECRET`.
   - Injected in HTTP requests via `Authorization: Bearer <token>`.
2. **Participant Session Handshake**:
   - Single-use invite tokens passed in query string `?token=<invite_token>`.
   - Validated against Redis cache or PostgreSQL `assessment_invites` table.
3. **Internal Microservice Authentication**:
   - HTTP requests between services require `X-Internal-Key` header matching `INTERNAL_API_KEY`.

---

## 7. Specialized Subsystem Architecture Flows

For in-depth mechanical breakdowns, sequence diagrams, failure modes, and code citations of critical subsystems, consult the following dedicated engineering guides:

1. **[Real-Time Voice Pipeline Flow](voice_pipeline_flow.md)**:
   - Web Audio PCM binary frame streaming and Deepgram Nova-2 streaming STT.
   - Pipelined sentence-boundary synthesis (2.8x dead-air reduction: 753.9ms $\to$ 273.7ms).
   - Barge-in interruption mechanics, context cancellation, and delivered-sentence history truncation.
   - Turn sequence barriers and Redis distributed locking (`session:lock:{id}`).

2. **[Speculative RAG Retrieval & Prefetch Flow](speculative_retrieval_flow.md)**:
   - Interim STT triggered prefetch, 768-dim embeddings, pgvector cosine search, and Redis caching.
   - Latency breakdown: 54.3ms embedding vs 10.5ms pgvector query (p50).
   - Cache hit vs fallback dynamics (1.15ms cache hit vs 64.8ms cold query fallback).
   - Architectural defect analysis: 22.2% semantic staleness under topic drift and proposed debounce/invalidation fixes.

3. **[Adaptive Competency Steering](adaptive_competency_steering.md)**:
   - Mathematical model of the 3-factor evaluation engine (Topic Coverage, Demonstrated Competence, Recency Decay).
   - Single-mention capping at 0.50 competence and $\ge 2$ substantive turns requirement.
   - Dynamic `CURRENT FOCUS:` system prompt injection based on maximum coverage deficit.

4. **[Asynchronous Evaluation & Grounding Pipeline](evaluation_and_grounding_flow.md)**:
   - Kafka and Outbox consumer mechanics with transcript completeness verification barriers.
   - Bounded-concurrent scatter-gather grading ($N=5$, 5.0x speedup: 502.2ms $\to$ 100.4ms).
   - Verbatim quote attribution (`IsSubstantiveEvidence`) eliminating stop-token false accepts (80.0% $\to$ 0.0%).
   - Bounded regex JSON schema repair and failure-preserving score aggregation.

---

## 8. Empirical Verification & Benchmarks

The repository includes an automated test and replay suite under `bench/` that empirically verifies latency and reliability invariants:
- `bench/fair_comparisons_test.go`: Microbenchmarks isolating pipelined synthesis (A), prefetch retrieval (B), and scatter-gather evaluation (C).
- `bench/prefetch_replay_test.go`: 30-turn 1:1 timeline replay measuring speculative retrieval hit rate, cold query fallback rate, and semantic drift.
- `services/feedback-worker/internal/worker/grounding_bench_test.go`: Precision/recall verification of evidence quote attribution against adversarial false accepts.
- `services/interview-orchestrator/internal/adaptive/adaptive_sim_test.go`: Convergence and topic drift ablation across multi-turn conversational transcripts.

