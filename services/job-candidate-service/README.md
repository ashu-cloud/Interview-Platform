# Job Candidate Service (`platform-service`)

The `job-candidate-service` is the administrative and configuration REST microservice of VoxAssess. It handles multi-tenant organization management, assessment tracks, blueprints, vector embedding generation for prompts and rubric criteria, public candidate self-application, invites, and embedded outbox turn projection.

---

## Architecture & Responsibilities

```
Hiring Manager Browser              Public Candidate Browser
        │                                      │
        ▼ (Authenticated REST)                 ▼ (Public REST)
+───────────────────────────────────────────────────────────────────────+
| job-candidate-service (platform-service :8001 / Host: :8003)          |
|                                                                       |
| 1. Authentication & Multi-Tenancy                                     |
|    ├── JWT HS256 authentication for hiring managers                   |
|    └── Multi-tenant organization scoping across all models            |
|                                                                       |
| 2. Assessment Tracks & Blueprints                                     |
|    ├── CRUD for assessment tracks (job roles, required skills)        |
|    ├── CRUD for blueprints (question prompts, instructions, rubrics)   |
|    └── Vector embeddings via Google Gemini or OpenAI (768-dim)        |
|                                                                       |
| 3. Public Candidate Self-Application                                  |
|    ├── GET  /api/v1/public/tracks/:job_id (Overview & blueprint)       |
|    ├── POST /api/v1/public/tracks/:job_id/apply (Self-application)    |
|    └── Fast-path Redis caching: invite:token:{token} (TTL: 7d)        |
|                                                                       |
| 4. Embedded Turn Projector (Outbox Mode)                              |
|    ├── Polls outbox_turns using FOR UPDATE SKIP LOCKED                |
|    └── Idempotently projects turns into relational session_turns       |
+───────────────────────────────────────────────────────────────────────+
        │
        ├──► PostgreSQL 16 (Relational models, outbox_turns, pgvector)
        └──► Redis 7.2 (Invite token cache)
```

---

## API Endpoints (v1)

### Public Candidate Portal (Unauthenticated)
- `GET /api/v1/public/tracks/:job_id`: Fetch public track details and active blueprint summary.
- `POST /api/v1/public/tracks/:job_id/apply`: Candidate self-application (creates participant, generates invite, returns token).
- `GET /api/v1/invites/validate/:token`: Validate token before opening WebSocket stage.

### Authentication & Management (Authenticated)
- `POST /api/v1/auth/register`: Create organization and manager account.
- `POST /api/v1/auth/login`: Authenticate and obtain Bearer JWT.
- `POST /api/v1/tracks`: Create assessment track.
- `GET /api/v1/tracks`: List tracks for organization.
- `POST /api/v1/tracks/:id/blueprints`: Create blueprint with nested questions and rubric criteria.
- `GET /api/v1/blueprints/:id`: Fetch blueprint details.
- `GET /api/v1/sessions`: List assessment sessions and progress.
- `GET /api/v1/sessions/:id`: Get session details.
- `GET /api/v1/sessions/:id/turns`: Get audited transcript turns.
- `GET /api/v1/sessions/:id/evaluation`: Get scored evaluation dossier.

---

## Environment Variables

| Variable | Default | Purpose |
| :--- | :--- | :--- |
| `PORT` | `8001` | HTTP server listen port (mapped to `:8003` in Docker Compose) |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection for invite token caching |
| `JWT_SECRET` | — | HMAC secret for signing manager JWT tokens |
| `INTERNAL_API_KEY` | — | Shared key for inter-service HTTP authentication |
| `APP_DOMAIN` | `http://localhost:3000` | Frontend web domain for generating invite links |
| `EMBEDDING_API_KEY` | — | Gemini or OpenAI API key for 768-dim embeddings |
| `EMBEDDING_BASE_URL`| — | `https://generativelanguage.googleapis.com/v1beta/openai/` for Gemini |
| `EMBEDDING_MODEL`   | `gemini-embedding-001` | Embedding model name |
| `EMBEDDING_DIMS`    | `768` | Embedding dimensionality with L2 normalization |

---

## Running & Testing

```bash
# Run unit and integration tests
go test -v ./...
```
