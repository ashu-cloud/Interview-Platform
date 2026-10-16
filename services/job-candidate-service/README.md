# Job Candidate Service (`platform-service`)

The `job-candidate-service` is the administrative and configuration REST microservice of VoxAssess. It handles multi-tenant organization management, assessment tracks, blueprints, vector embedding generation for prompts and rubric criteria, candidate invitations, and session metadata.

---

## Architecture & Responsibilities

```
Hiring Manager Browser
        │
        ▼ (HTTP REST: /api/v1/...)
+───────────────────────────────────────────────────────────────────────+
| job-candidate-service (platform-service :8001)                        |
|                                                                       |
| 1. Authentication & Multi-Tenancy                                     |
|    ├── JWT HS256 authentication for hiring managers                   |
|    └── Multi-tenant organization scoping across all models            |
|                                                                       |
| 2. Assessment Tracks & Blueprints                                     |
|    ├── CRUD for assessment tracks (job roles, required skills)        |
|    └── CRUD for blueprints (question prompts, instructions, rubrics)   |
|                                                                       |
| 3. Asynchronous Vector Generation (docket.go)                         |
|    ├── Background embedding generation (768-dim text-embedding-3-small)|
|    └── Populates document_chunks for RAG prefetcher                   |
|                                                                       |
| 4. Candidate & Invite Management                                      |
|    ├── Single-use invite token generation                             |
|    ├── Fast-path Redis caching: invite:token:{token}                  |
|    └── Public token validation endpoint for entering candidates       |
+───────────────────────────────────────────────────────────────────────+
        │
        ├──► PostgreSQL 16 (Relational models & pgvector)
        └──► Redis 7.2 (Invite token cache)
```

---

## API Endpoints (v1)

### Authentication
- `POST /api/v1/auth/register`: Create organization and manager account.
- `POST /api/v1/auth/login`: Authenticate and obtain Bearer JWT.

### Tracks & Blueprints
- `POST /api/v1/tracks`: Create assessment track.
- `GET /api/v1/tracks`: List tracks for organization.
- `POST /api/v1/tracks/:id/blueprints`: Create blueprint with nested questions and rubric criteria.
- `GET /api/v1/blueprints/:id`: Fetch blueprint details.

### Candidates & Invites
- `POST /api/v1/tracks/:id/participants`: Add candidate to track.
- `POST /api/v1/participants/:id/invite`: Generate single-use assessment invite link.
- `GET /api/v1/invites/validate/:token`: Public endpoint validating candidate token before WebSocket connection.

---

## Environment Variables

| Variable | Default | Purpose |
| :--- | :--- | :--- |
| `PORT` | `8001` | HTTP server listen port |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection for invite token caching |
| `JWT_SECRET` | — | HMAC secret for signing manager JWT tokens |
| `INTERNAL_API_KEY` | — | Shared key for inter-service HTTP authentication |
| `OPENAI_API_KEY` | — | Used to generate 768-dim embeddings for blueprint chunks |

---

## Running & Testing

```bash
# Run unit and repository tests
go test -v ./...
```
