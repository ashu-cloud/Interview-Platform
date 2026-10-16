# Turn Projector Service (`turn-projector`)

The `turn-projector` is a headless projection worker that decouples real-time conversational voice interaction from relational database writes. It consumes conversation turn events from Kafka or the PostgreSQL Outbox and writes them idempotently to the `session_turns` table.

---

## Architectural Purpose

During an active voice assessment turn, the `interview-orchestrator` critical path must remain sub-second and in-memory. Awaiting PostgreSQL transactions or disk writes directly within the WebSocket loop introduces latency spikes and dead-air audio jitter.

The `turn-projector` solves this by handling database projection asynchronously:

```
interview-orchestrator (session-runtime)
        │
        ├── (Non-blocking async publish)
        │
        ▼
Event Stream:
  ├── Kafka Mode: topic session.turns
  └── Outbox Mode: table outbox_turns (FOR UPDATE SKIP LOCKED)
        │
        ▼
+───────────────────────────────────────────────────────────────────────+
| turn-projector                                                        |
|                                                                       |
| 1. Dual-Broker Ingestion: Kafka Consumer Group / Outbox Poller        |
| 2. Payload Validation & Deserialization                               |
| 3. Idempotent Ingestion:                                              |
|    INSERT INTO session_turns (session_id, org_id, turn_number,        |
|                               role, content)                          |
|    VALUES ($1, $2, $3, $4, $5)                                        |
|    ON CONFLICT (session_id, turn_number, role) DO NOTHING;            |
| 4. Outbox State Acknowledgment: status = 'completed'                  |
+───────────────────────────────────────────────────────────────────────+
        │
        ▼
PostgreSQL (session_turns)
```

---

## Key Guarantees

1. **Zero Contention (`FOR UPDATE SKIP LOCKED`)**:
   In Outbox mode, multiple projector replicas poll concurrently without database serialization deadlocks or lock contention.
2. **Idempotence**:
   The unique constraint on `(session_id, turn_number, role)` ensures that duplicate network deliveries or consumer restarts never produce duplicate conversational turns.
3. **Poison Pill Defense**:
   Malicious or malformed messages that fail unmarshaling are retried up to `max_retries = 5` before transitioning to `status = 'failed'`, preventing head-of-line blocking.

---

## Environment Variables

| Variable | Default | Purpose |
| :--- | :--- | :--- |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `KAFKA_BROKERS` | `""` | Comma-separated Kafka broker addresses (uses Outbox mode if empty) |

---

## Running & Testing

```bash
go test -v ./...
```
