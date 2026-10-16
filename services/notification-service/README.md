# Notification Service (`notification-service`)

The `notification-service` is an asynchronous event-driven worker responsible for dispatching transactional alerts, candidate interview invitation emails, and completed assessment notifications.

---

## Architectural Workflow

```
Platform Service (Candidate Invited) / Feedback Worker (Evaluation Finished)
        │
        ├──► Kafka: notification.events / PostgreSQL Outbox
        │
        ▼
+───────────────────────────────────────────────────────────────────────+
| notification-service                                                  |
|                                                                       |
| 1. Consumes notification payload                                      |
| 2. Resolves recipient candidate / hiring manager email address        |
| 3. Renders HTML email template (assessment link / score summary)      |
| 4. Dispatches email via Resend API (or logs payload in local dev)    |
+───────────────────────────────────────────────────────────────────────+
```

---

## Environment Variables

| Variable | Default | Purpose |
| :--- | :--- | :--- |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `KAFKA_BROKERS` | `""` | Comma-separated Kafka broker addresses |
| `RESEND_API_KEY` | — | Resend transactional email API key |
| `EMAIL_FROM` | `assessments@voxassess.io` | Sender email address |

---

## Running & Testing

```bash
go test -v ./...
```
