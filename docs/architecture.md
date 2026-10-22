# Assessment Engine — System Architecture

## 1. System Overview

The **Assessment Engine** is a low-latency, real-time adaptive voice assessment platform. It conducts structured technical and behavioral evaluations over streaming audio, continuously tracks competency coverage across live conversational turns, dynamically guides LLM prompting to explore uncovered criteria, and asynchronously produces evidence-grounded evaluation reports citing exact conversation turns.

The platform employs a unified, zero-broker architecture:
- **Real-Time Voice Critical Path**: Sub-second full-duplex WebSocket connections terminated by `interview-orchestrator`, leveraging in-memory state in Redis, streaming STT via Deepgram, streaming LLM via Groq, pipelined sentence TTS via ElevenLabs, and pre-roll audio filtering.
- **Transactional Outbox Event Bus**: Microservices decouple state through PostgreSQL outbox tables (`outbox_turns`, `outbox_completed`) polled with `FOR UPDATE SKIP LOCKED`. This eliminates Kafka operational and RAM overhead, providing strict ordering and idempotency across both local development and cloud production deployments.

---

## 2. Dual-Track Architecture

The system decouples real-time audio interaction from persistence and evaluation through two architectural tracks:

```
+--------------------------------------------------------------------------------------------------+
| TRACK 1: LIVE CRITICAL PATH (Sub-second, in-memory, non-blocking)                               |
|                                                                                                  |
| [Candidate Browser] <--- Full-Duplex PCM WebSocket ---> [interview-orchestrator (:8004)]        |
|                                                                 |                                |
|        +-------------------------+------------------------------+--------------------+           |
|        |                         |                              |                    |           |
|  [Streaming STT]       [Speculative Prefetch]            [Streaming LLM]       [Streaming TTS]   |
| (Deepgram Nova-2)      (pgvector + Redis Cache)        (Groq Llama 3.3 70B)  (ElevenLabs Turbo)  |
+--------------------------------------------------------------------------------------------------+
                                           |
                       (Non-blocking transactional outbox persistence)
                                           |
                                           v
+--------------------------------------------------------------------------------------------------+
| TRACK 2: ASYNC PROJECTION & EVALUATION PATH (ACID, idempotent, latency-tolerant)                |
|                                                                                                  |
| PostgreSQL Outbox Pattern (FOR UPDATE SKIP LOCKED):                                             |
|                                                                                                  |
|   interview-orchestrator ---> Postgres [outbox_turns]     ---> [Embedded Turn Projector]        |
|                                                                 (in job-candidate-service)       |
|                                                                 └──► [session_turns]             |
|                                                                                                  |
|   interview-orchestrator ---> Postgres [outbox_completed] ---> [feedback-worker]                 |
|                                                                 └──► [evaluations]               |
+--------------------------------------------------------------------------------------------------+
```

### Track 1: Live Critical Path
- **Sub-second latency constraint**: Time from participant speech end to assessor audio playback target is sub-second (p50: 273.7ms).
- **In-memory state**: During an active turn, state resides strictly in Redis (`session:state:{id}` and distributed locks `session:lock:{id}`).
- **Zero blocking database writes**: The critical WebSocket loop never awaits disk I/O or heavy database transactions. Turn events are recorded via background goroutines and non-blocking outbox insertion.
- **Audio Conditioning**: Browser-side anti-aliased downsampling (16kHz PCM) and ~250ms circular audio pre-roll buffer prevent initial phoneme clipping.

### Track 2: Async Projection & Evaluation Path
- **Embedded Turn Projection**: The `job-candidate-service` embeds a concurrent turn projector polling `outbox_turns` every 500ms, validating payload models and performing idempotent writes into the `session_turns` table.
- **Evaluation Generation**: Upon session completion, `feedback-worker` polls `outbox_completed`, processes the transcript, generates scoring rubrics via parallel LLM evaluation, and records reports into `evaluations` and `evaluation_criteria_results`.

---

## 3. Event Bus & The PostgreSQL Outbox Pattern

To eliminate the operational complexity and heavy RAM requirements of Kafka brokers, the platform utilizes an ACID-compliant PostgreSQL outbox pattern that guarantees message delivery and strict ordering.

### Outbox Dequeue Mechanics
The embedded turn projector and evaluation workers poll the database on a configurable interval (default: 500ms):

```sql
SELECT id, session_id, org_id, turn_number, role, content, retry_count
FROM outbox_turns
WHERE status = 'pending'
ORDER BY created_at ASC, id ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;
```

#### Key Guarantees:
1. **Concurrency Safety (`FOR UPDATE SKIP LOCKED`)**: Multiple worker instances or goroutines poll concurrently without row contention, blocking, or serialization deadlocks.
2. **Idempotent Ingestion**: Projected rows write to `session_turns` with unique constraints:
   ```sql
   INSERT INTO session_turns (session_id, org_id, turn_number, role, content)
   VALUES ($1, $2, $3, $4, $5)
   ON CONFLICT (session_id, turn_number, role) DO NOTHING;
   ```
3. **Atomic State Transition**: Rows are updated to `status = 'completed'` in the same transaction or immediately upon completion.
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
            v                                   v
+-----------------------+           +-----------------------+
|         users         |           |   assessment_tracks   |
|   (Hiring Managers)   |           |      (Job Roles)      |
+-----------------------+           +-----------+-----------+
                                                |
                                    +-----------+-----------+
                                    |                       |
                                    v                       v
                        +-----------------------+   +---------------+
                        | assessment_blueprints |   |  participants |
                        +-----------+-----------+   +-------+-------+
                                    |                       |
            +-----------------------+                       |
            |                       |                       |
            v                       v                       v
+-----------------------+   +-------------------+   +--------------------+
|   blueprint_prompts   |   | blueprint_rubrics |   | participant_invites|
| (Questions / Vectors) |   | (Criteria/Vectors)|   +---------+----------+
+-----------------------+   +-------------------+             |
                                                              v
                                                    +--------------------+
                                                    | interview_sessions |
                                                    +---------+----------+
                                                              |
                                            +-----------------+-----------------+
                                            |                                   |
                                            v                                   v
                                +-----------------------+           +-----------------------+
                                |     session_turns     |           |      evaluations      |
                                |  (Audited Transcript) |           |  (Scored Dossiers)    |
                                +-----------------------+           +-----------+-----------+
                                                                                |
                                                                                v
                                                                    +-----------------------+
                                                                    |evaluation_criteria_   |
                                                                    |        results        |
                                                                    +-----------------------+
```

### Outbox Tables Schema
- **`outbox_turns`**: Stores serialized participant and assessor turns awaiting projection to `session_turns`.
- **`outbox_completed`**: Stores completed session notifications awaiting asynchronous evaluation processing.

### Redis Key Schema
- `invite:token:{token}`: JSON-cached invite metadata for instant validation (TTL: 7 days).
- `session:state:{session_id}`: Active session state, turn counters, and conversation history buffer.
- `session:lock:{session_id}`: Distributed mutex lock preventing duplicate connections (TTL: 30s with 15s heartbeats).
- `rag:prefetch:{session_id}:{turn}`: Speculatively fetched rubric criteria chunks (TTL: 60s).

---

## 5. Microservices & Deployment Topology

| Service | Directory | Host Port | Protocol | Description |
| :--- | :--- | :--- | :--- | :--- |
| **Frontend Web App** | `frontend/` | `:3000` | HTTP / WS | Next.js 16 App Router UI & Turbopack |
| **Platform Service** | `services/job-candidate-service` | `:8003` | HTTP REST | Auth, tracks, blueprints, public apply, invites, embedded turn projector |
| **Session Runtime** | `services/interview-orchestrator` | `:8004` | WebSocket | Real-time voice orchestrator, STT, LLM, TTS, speculative RAG prefetch, `/debug/voice` probe |
| **Evaluation Worker** | `services/feedback-worker` | — | Outbox Worker | Grounded transcript assessment, rubric grading, JSON repair |
| **Notification Service** | `services/notification-service` | — | Outbox Worker | Candidate invite and completion notifications via Resend |

---

## 6. Security & Candidate Access

1. **Manager Authentication**:
   - Bearer JWT tokens signed with HS256 using `JWT_SECRET`.
   - Required for all track, blueprint, and session management routes.
2. **Public Candidate Application Flow**:
   - Candidates self-apply via `POST /api/v1/public/tracks/:job_id/apply` without requiring administrative credentials.
   - Generates an interview token, initializes the session in `invited` state, and returns `{ token, invite_id, interview_id, invite_link }`.
3. **Participant Session Handshake**:
   - Single-use invite tokens passed in query string `?token=<invite_token>` during WebSocket connection.
   - Fast-path validation against Redis cache with fallback to PostgreSQL.
4. **Internal Microservice Authentication**:
   - Service-to-service communication requires the `X-Internal-Key` header matching `INTERNAL_API_KEY`.
