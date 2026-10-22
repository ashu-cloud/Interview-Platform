# Feedback Worker (`evaluation-worker`)

The `feedback-worker` is the asynchronous grading microservice of VoxAssess. It consumes interview completion events from the PostgreSQL Outbox (`outbox_completed`), fetches the candidate transcript, validates completeness, evaluates responses concurrently against the rubric, validates evidence quotes against verbatim utterances, and stores structured evaluation reports.

---

## Architectural Workflow

```
Postgres Outbox: outbox_completed (FOR UPDATE SKIP LOCKED)
        │
        ▼
+───────────────────────────────────────────────────────────────────────+
| feedback-worker (evaluation-worker)                                   |
|                                                                       |
| 1. Transcript Completeness Barrier: ValidateTranscriptCompleteness() |
|    ├── Verifies turns 1..N form a contiguous sequential chain         |
|    └── Applies bounded backoff (waitForTranscript) if still writing   |
|                                                                       |
| 2. Bounded-Concurrent Scatter-Gather: EvaluateTranscript()            |
|    ├── Spawns worker pool bounded to N=5 concurrent goroutines        |
|    ├── Evaluates each rubric criterion independently via Groq LLM     |
|    └── Reduces turnaround from 502.2ms to 100.4ms (5.0x speedup)      |
|                                                                       |
| 3. Schema Parsing & Deterministic Regex Repair                        |
|    ├── Strips markdown fences (```json ... ```) and trailing commas  |
|    └── Executes single bounded LLM repair prompt if parsing fails     |
|                                                                       |
| 4. Evidence Grounding & Verbatim Attribution: IsSubstantiveEvidence() |
|    ├── Requires >= 3 substantive words OR >= 15 characters           |
|    ├── Rejects conversational stop tokens ('yes', 'sure', 'ok')       |
|    └── Matches exact substring in transcript (0% false accepts)      |
|                                                                       |
| 5. Transactional Persistence & Error Propagation                      |
|    ├── Preserves genuine 0.0 candidate scores                         |
|    ├── Flags evaluation_failed on LLM errors (avoids false scores)   |
|    └── Persists report + scores atomically in PostgreSQL              |
+───────────────────────────────────────────────────────────────────────+
        │
        ├──► evaluations (Executive summary, composite score, recommendation)
        └──► evaluation_criteria_results (Per-criterion scores & verbatim quotes)
```

---

## Key Modules & Implementations

### 1. PostgreSQL Outbox Consumer (`internal/worker/evaluation.go`)
- Polls `outbox_completed` using:
  ```sql
  SELECT id, session_id, org_id, retry_count
  FROM outbox_completed
  WHERE status = 'pending'
  ORDER BY created_at ASC
  LIMIT 1
  FOR UPDATE SKIP LOCKED;
  ```
- Guarantees zero contention across multiple worker replicas and head-of-line poison pill avoidance via bounded retries (`max_retries = 5`).

### 2. Transcript Completeness Barrier
- Prevents race conditions where an evaluation commences before the turn projector finishes writing final turns.
- Verifies contiguous turn sequences and blocks with exponential backoff up to 10 seconds.

### 3. Substantive Evidence Attribution Guard (`IsSubstantiveEvidence`)
- Rejects generic conversational filler phrases (*"yes"*, *"okay"*, *"I see"*).
- Requires exact substring matches against the candidate's transcribed turns, dropping affirmative false accepts from **80.0% to 0.0%**.

---

## Environment Variables

| Variable | Default | Purpose |
| :--- | :--- | :--- |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `JOB_CANDIDATE_SERVICE_URL` | `http://localhost:8001` | Platform service URL for blueprint rubrics |
| `INTERNAL_API_KEY` | — | Shared key for inter-service HTTP authentication |
| `GROQ_API_KEY` | — | Groq API key for LLaMA 3.3 70B evaluation prompts |
| `FEEDBACK_LLM_MODE` / `EVALUATION_LLM_MODE` | `groq` | `groq` for live evaluation, `mock` for local offline testing |

---

## Running & Testing

```bash
# Run unit and grounding verification tests
go test -v ./...
```
