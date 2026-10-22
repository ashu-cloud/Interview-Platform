# Assessment Engine — API Specification

This document specifies the HTTP REST API and WebSocket protocols for the Assessment Engine.

---

## Authentication & Authorization

The system uses three access tiers:

1. **Bearer JWT** (Assessment Authors / Hiring Managers):
   - Header: `Authorization: Bearer <jwt_token>`
   - Acquired via `/api/v1/auth/login` or `/api/v1/auth/register`.
   - Used for all authoring, participant management, and session querying endpoints.

2. **Public / Unauthenticated Access** (Candidates):
   - Public endpoints allow candidates to self-apply, inspect track criteria, and validate invite tokens without a manager account.
   - Used at `/api/v1/public/*` and `/api/v1/invites/validate/:token`.

3. **Internal API Key** (Service-to-Service):
   - Header: `X-Internal-Key: <internal_key>`
   - Used by `interview-orchestrator` and background workers for internal synchronization and diagnostics.

---

## Canonical API Routes (v1)

### Public Candidate Portal (Unauthenticated)

- `GET /api/v1/public/tracks/:job_id`
  - Fetches public track overview and first active blueprint metadata for prospective applicants.
  - Response: `200 OK`
    ```json
    {
      "track": {
        "id": "uuid",
        "title": "Senior Distributed Systems Engineer",
        "description": "Design resilient distributed architectures...",
        "required_skills": ["Go", "Distributed Systems", "Raft"]
      },
      "blueprint_id": "uuid",
      "blueprint_title": "Core Systems & Consensus Blueprint",
      "duration_mins": 30
    }
    ```

- `POST /api/v1/public/tracks/:job_id/apply`
  - Allows candidates to self-apply. Automatically finds or creates a participant record, attaches the active blueprint, initializes an interview session, caches the invite token in Redis (TTL: 7 days), and returns an invite token.
  - Body:
    ```json
    {
      "full_name": "Jane Doe",
      "email": "jane.doe@example.com"
    }
    ```
  - Response: `201 Created`
    ```json
    {
      "invite_id": "uuid",
      "invite_link": "http://localhost:3000/interview/token_xyz",
      "token": "token_xyz",
      "expires_at": "2026-10-27T12:00:00Z"
    }
    ```

- `GET /api/v1/invites/validate/:token`
  - Validates candidate invite token prior to opening the real-time WebSocket stage.
  - Response: `200 OK`
    ```json
    {
      "valid": true,
      "candidate": { "id": "uuid", "full_name": "Jane Doe" },
      "track": { "id": "uuid", "title": "Senior Distributed Systems Engineer" },
      "blueprint_title": "Core Systems & Consensus Blueprint",
      "duration_mins": 30,
      "expires_at": "2026-10-27T12:00:00Z"
    }
    ```

---

### Assessment Tracks (Authenticated)

- `POST /api/v1/tracks`: Create assessment track.
- `GET /api/v1/tracks`: List tracks for organization (`?status=active|archived&page=1&limit=20`).
- `GET /api/v1/tracks/:track_id`: Get track details by ID.
- `PATCH /api/v1/tracks/:track_id`: Partial update of track.
- `DELETE /api/v1/tracks/:track_id`: Delete track (returns `409 Conflict` if active blueprints or candidates depend on it).

---

### Assessment Blueprints (Authenticated)

- `POST /api/v1/tracks/:track_id/blueprints`
  - Creates blueprint with nested questions and rubric criteria. Generates vector embeddings for criteria asynchronously.
  - Body:
    ```json
    {
      "title": "Senior Distributed Systems",
      "instructions": "Probe concurrency and reliability trade-offs",
      "duration_mins": 30,
      "questions": [
        {
          "question_text": "Explain how you implement distributed locking in Redis.",
          "question_type": "technical",
          "order_index": 1,
          "follow_up_hint": "Ask about TTL renewal and clock drift."
        }
      ],
      "rubric": [
        {
          "name": "Concurrency Control",
          "description": "Demonstrates mutex, channel, or distributed lock mechanics",
          "max_score": 10,
          "order_index": 1
        }
      ]
    }
    ```
- `GET /api/v1/tracks/:track_id/blueprints`: List blueprints for track.
- `GET /api/v1/blueprints/:blueprint_id`: Fetch blueprint details with prompts and rubric.
- `PATCH /api/v1/blueprints/:blueprint_id`: Update blueprint.
- `DELETE /api/v1/blueprints/:blueprint_id`: Delete blueprint.

---

### Participants & Invites (Authenticated)

- `POST /api/v1/tracks/:track_id/participants`: Add participant record manually.
- `GET /api/v1/tracks/:track_id/participants`: List participants and their interview statuses.
- `GET /api/v1/participants/:participant_id`: Fetch participant details.
- `POST /api/v1/participants/:participant_id/invite`: Generate single-use invite token.

---

### Assessment Sessions & Dossiers (Authenticated)

- `POST /api/v1/sessions`: Initialize or resume an assessment session (`{"invite_id": string}`).
- `GET /api/v1/sessions`: List sessions across organization (`?status=invited|in_progress|completed|feedback_ready&page=1&limit=20`).
- `GET /api/v1/sessions/:session_id`: Fetch session details, candidate information, and score summaries.
- `GET /api/v1/sessions/:session_id/turns`: List audited conversation turns from `session_turns`.
- `POST /api/v1/sessions/:session_id/evaluation`: Trigger or retry asynchronous evaluation generation (`202 Accepted`).
- `GET /api/v1/sessions/:session_id/evaluation`: Retrieve scored evaluation report, executive summary, and verbatim cited evidence quotes.

---

### Diagnostics & Observability

- `GET /debug/voice` (interview-orchestrator on `:8004` / `:8002`)
  - Gated by `VOICE_DEBUG=true` or `X-Internal-Key`.
  - Probes live upstream dependencies:
    - **Redis**: PING latency test.
    - **Platform API**: Checks connectivity to `job-candidate-service`.
    - **Deepgram STT**: Probes Deepgram WebSocket handshaking.
    - **Groq LLM**: Probes minimal streaming generation latency.
    - **ElevenLabs TTS**: Probes audio synthesis chunk delivery.
  - Response:
    ```json
    {
      "status": "healthy",
      "timestamp": "2026-10-20T16:00:00Z",
      "checks": {
        "redis": { "status": "ok", "latency_ms": 1.2 },
        "platform_api": { "status": "ok", "latency_ms": 4.5 },
        "deepgram_stt": { "status": "ok" },
        "groq_llm": { "status": "ok", "model": "openai/gpt-oss-120b" },
        "elevenlabs_tts": { "status": "ok", "voice_id": "21m00Tcm4TlvDq8ikWAM" }
      }
    }
    ```

---

## Real-Time WebSocket Protocol

- **Endpoint**: `/ws/interview/?token=<invite-token>` (handled by `interview-orchestrator` on `:8004` / `:8002`)

### Binary Protocol Tags
All binary WebSocket frames are prefixed with a 1-byte opcode tag:
- **`0x01` (Client -> Server)**: Raw Linear PCM audio chunk (16kHz, 16-bit mono Little Endian).
- **`0x02` (Server -> Client)**: Synthesized MP3 audio chunk (24kHz) for immediate client playback.

### Control Frames (JSON Text)

#### Client → Server Messages:
- `{"type": "client_ready"}`: Client signals audio context is active and ready for interview greeting.
- `{"type": "interrupt"}`: Client signals local speech detection, halting assessor playback.
- `{"type": "end_assessment"}`: Client requests graceful completion of assessment session.
- `{"type": "ping"}`: Heartbeat frame sent every 10 seconds.

#### Server → Client Messages:
- `{"type": "connected", "session_id": "...", "has_history": bool}`: Connection established. `has_history: true` signals a reconnected session with existing conversation history.
- `{"type": "transcript", "role": "participant"|"agent", "content": "..."}`: Verbatim transcribed utterance.
- `{"type": "interim_transcript", "content": "..."}`: Real-time interim speech hypothesis.
- `{"type": "ai_interrupted", "turn": int}`: Signals server has cancelled in-flight response and synchronized atomic turn counter.
- `{"type": "assessment_complete"}`: Assessment concluded and marked `completed` in outbox ledger.
- `{"type": "error", "message": string}`: Error payload.
