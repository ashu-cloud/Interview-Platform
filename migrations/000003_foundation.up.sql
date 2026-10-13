-- pgvector
CREATE EXTENSION IF NOT EXISTS vector;

-- Embedding columns (populated in Phase 3)
ALTER TABLE evaluation_criteria ADD COLUMN IF NOT EXISTS embedding vector(768);
ALTER TABLE blueprint_prompts   ADD COLUMN IF NOT EXISTS embedding vector(768);

-- Materialized session turns
CREATE TABLE IF NOT EXISTS session_turns (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  UUID NOT NULL REFERENCES assessment_sessions(id) ON DELETE CASCADE,
    org_id      UUID NOT NULL REFERENCES organizations(id),
    turn_number INT NOT NULL,
    role        VARCHAR(20) NOT NULL CHECK (role IN ('participant', 'assessor', 'system')),
    content     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (session_id, turn_number, role)
);
CREATE INDEX IF NOT EXISTS idx_session_turns_session_id ON session_turns(session_id);

-- Document chunks for RAG. Vector indexes (ivfflat/HNSW) are deliberately NOT
-- created here; they are added in Phase 3 after benchmarking.
CREATE TABLE IF NOT EXISTS document_chunks (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blueprint_id UUID NOT NULL REFERENCES assessment_blueprints(id) ON DELETE CASCADE,
    org_id       UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    chunk_text   TEXT NOT NULL,
    embedding    vector(768),
    chunk_index  INT NOT NULL,
    source       VARCHAR(255),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_document_chunks_blueprint_id ON document_chunks(blueprint_id);

-- Audio recording references
CREATE TABLE IF NOT EXISTS audio_recordings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  UUID NOT NULL REFERENCES assessment_sessions(id) ON DELETE CASCADE,
    turn_number INT,
    storage_uri TEXT NOT NULL,
    size_bytes  BIGINT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audio_recordings_session_id ON audio_recordings(session_id);

-- Outbox tables: Postgres-native event queue used when KAFKA_BROKERS is unset.
CREATE TABLE IF NOT EXISTS outbox_turns (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id   UUID NOT NULL,
    org_id       UUID NOT NULL,
    turn_number  INT NOT NULL,
    role         VARCHAR(20) NOT NULL CHECK (role IN ('participant', 'assessor', 'system')),
    content      TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status       TEXT NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    retry_count  INT NOT NULL DEFAULT 0,
    processed_at TIMESTAMPTZ,
    failed_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_outbox_turns_pending
    ON outbox_turns(created_at) WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS outbox_completed (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id     UUID NOT NULL,
    org_id         UUID NOT NULL,
    participant_id UUID NOT NULL,
    blueprint_id   UUID NOT NULL,
    turn_count     INT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status         TEXT NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    retry_count    INT NOT NULL DEFAULT 0,
    processed_at   TIMESTAMPTZ,
    failed_at      TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_outbox_completed_pending
    ON outbox_completed(created_at) WHERE status = 'pending';
