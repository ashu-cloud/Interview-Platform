# Assessment Engine — API Specification

This document specifies the HTTP REST API and WebSocket protocols for the Assessment Engine.

---

## Authentication & Authorization

The system uses two authentication mechanisms:

1. **Bearer JWT** (Assessment Authors / Users):
   - Header: `Authorization: Bearer <jwt_token>`
   - Acquired via `/api/v1/auth/login` or `/api/v1/auth/register`.
   - Used for all authoring, participant management, and session querying endpoints.

2. **Internal API Key** (Service-to-Service):
   - Header: `X-Internal-Key: <internal_key>`
   - Used by `session-runtime` (`interview-orchestrator`) and background workers for internal synchronization.

---

## Canonical API Routes (v1)

### Assessment Tracks
Tracks represent job roles and competency definitions.

- `POST /api/v1/tracks`
  - Body: `{"title": string, "description": string, "required_skills": string[]}`
  - Response: `201 Created` — Track object
- `GET /api/v1/tracks`
  - Query: `?status=active|archived&page=1&limit=20`
  - Response: `200 OK` — `{"tracks": [...], "total": int, "page": int, "limit": int}`
- `GET /api/v1/tracks/:track_id`
  - Response: `200 OK` — Track object
- `PATCH /api/v1/tracks/:track_id`
  - Body: `{"title"?: string, "description"?: string, "required_skills"?: string[], "status"?: string}`
  - Response: `200 OK` — Updated Track object
- `DELETE /api/v1/tracks/:track_id`
  - Response: `204 No Content` (Returns `409 Conflict` if active blueprints or participants depend on it)

---

### Assessment Blueprints
Blueprints contain the question prompts and evaluation rubrics for an assessment.

- `POST /api/v1/tracks/:track_id/blueprints`
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
  - Response: `201 Created` — Blueprint object with nested prompts and rubric criteria.
- `GET /api/v1/tracks/:track_id/blueprints`
  - Response: `200 OK` — `{"blueprints": [...]}`
- `GET /api/v1/blueprints/:blueprint_id`
  - Response: `200 OK` — Blueprint object
- `PATCH /api/v1/blueprints/:blueprint_id`
  - Body: `{"title"?: string, "instructions"?: string, "duration_mins"?: int}`
  - Response: `200 OK` — Updated Blueprint object
- `DELETE /api/v1/blueprints/:blueprint_id`
  - Response: `204 No Content` (Returns `409 Conflict` if referenced by active invites)

---

### Participants & Assessment Links (Invites)

- `POST /api/v1/tracks/:track_id/participants`
  - Body: `{"email": string, "full_name": string}`
  - Response: `201 Created` — Participant object
- `GET /api/v1/tracks/:track_id/participants`
  - Response: `200 OK` — `{"participants": [{"id": ..., "full_name": ..., "email": ..., "interview_status": ...}]}`
- `GET /api/v1/participants/:participant_id`
  - Response: `200 OK` — Participant object
- `POST /api/v1/participants/:participant_id/invite`
  - Body: `{"docket_id": string, "expires_in_days": int}`
  - Response: `201 Created` — `{"invite_id": string, "invite_link": string, "token": string, "expires_at": string}`
- `GET /api/v1/invites/validate/:token` (Public / unauthenticated)
  - Validates token for participant entering assessment stage.
  - Response: `200 OK` — `{"valid": bool, "participant": {...}, "track": {...}, "blueprint_title": string, "expires_at": string}`

---

### Assessment Sessions

- `POST /api/v1/sessions`
  - Body: `{"invite_id": string}`
  - Response: `201 Created` (Idempotent: returns existing session if already initialized).
- `GET /api/v1/sessions`
  - Query: `?status=invited|in_progress|completed|feedback_ready&job_id=&page=1&limit=20`
  - Response: `200 OK` — `{"sessions": [...], "total": int}`
- `GET /api/v1/sessions/:session_id`
  - Response: `200 OK` — Detailed session object (status, participant info, blueprint title, timestamps, evaluation brief).
- `GET /api/v1/sessions/:session_id/turns`
  - Response: `200 OK` — `{"turns": [{"turn_number": int, "role": string, "content": string, "created_at": string}]}`
- `POST /api/v1/sessions/:session_id/evaluation`
  - Triggers asynchronous evaluation report generation (enqueues `outbox_completed` event).
  - Response: `202 Accepted` — `{"id": string, "session_id": string, "turn_count": int, "status": "pending"}`
- `GET /api/v1/sessions/:session_id/evaluation`
  - Response: `200 OK` — Scored report, overall recommendation, rubric breakdown, and evidence transcript.

---

## Backward-Compatibility Route Aliases

For seamless interoperability during migration, the following legacy paths remain supported:

| Legacy Route | Handled As | Status |
|---|---|---|
| `GET, POST /api/v1/jobs` | `/api/v1/tracks` | Supported (returns `jobs` envelope on GET) |
| `GET, PATCH /api/v1/jobs/:job_id` | `/api/v1/tracks/:track_id` | Supported |
| `GET, POST /api/v1/jobs/:job_id/dockets` | `/api/v1/tracks/:track_id/blueprints` | Supported (returns `dockets` envelope on GET) |
| `GET /api/v1/dockets/:docket_id` | `/api/v1/blueprints/:blueprint_id` | Supported |
| `GET, POST /api/v1/jobs/:job_id/candidates` | `/api/v1/tracks/:track_id/participants` | Supported (returns `candidates` envelope on GET) |
| `GET /api/v1/candidates/:candidate_id` | `/api/v1/participants/:participant_id` | Supported |
| `POST /api/v1/candidates/:candidate_id/invite` | `/api/v1/participants/:participant_id/invite` | Supported |
| `GET /api/v1/interviews` | `/api/v1/sessions` | Supported (returns `interviews` envelope on GET) |
| `GET /api/v1/interviews/:interview_id` | `/api/v1/sessions/:session_id` | Supported |
| `GET /api/v1/interviews/:interview_id/feedback` | `/api/v1/sessions/:session_id/evaluation` | Supported |
| `PATCH /api/v1/interviews/:interview_id/status` | `/internal/sessions/:session_id/status` | Supported (X-Internal-Key) |
| `PATCH /api/v1/invites/:invite_id/use` | Marks invite used | Supported (X-Internal-Key) |

---

## Real-Time WebSocket Protocol

- **Endpoint**: `/ws/interview/?token=<invite-token>` (handled by `session-runtime`)

### Client → Server Messages

- `speech_started`: Participant begins speaking.
- `speech_ended`: Participant stops speaking.
- `audio_frame`: Base64 binary chunk of real-time audio.
- `audio_end`: Finalized turn audio frame blob.
- `heartbeat`: Ping message sent every 10 seconds to maintain connection.
- `end_assessment`: Participant requests to finalize session. (Legacy `end_interview` is also supported).

### Server → Client Messages

- `connected`: Session connected and initialized.
- `ai_text_chunk`: Streamed token of the assessor's response.
- `ai_text_end`: Completed text turn of the assessor.
- `ai_audio_chunk`: Synthesized binary audio chunk (ElevenLabs streaming).
- `ai_audio_end`: Completed audio playback chunk for current turn.
- `ai_interrupted`: Sent if the participant interrupts the assessor during TTS playback.
- `interim_transcript`: Real-time streaming transcription from STT.
- `final_transcript`: Completed transcription turn.
- `assessment_complete`: Session concluded and marked completed. (Legacy `interview_complete` is also supported).
- `heartbeat_ack`: Acknowledgement of client heartbeat.
- `error`: Error payload `{"type": "error", "message": string}`.
