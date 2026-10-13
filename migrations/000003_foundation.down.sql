DROP TABLE IF EXISTS outbox_completed;
DROP TABLE IF EXISTS outbox_turns;
DROP TABLE IF EXISTS audio_recordings;
DROP TABLE IF EXISTS document_chunks;
DROP TABLE IF EXISTS session_turns;
ALTER TABLE blueprint_prompts   DROP COLUMN IF EXISTS embedding;
ALTER TABLE evaluation_criteria DROP COLUMN IF EXISTS embedding;
DROP EXTENSION IF EXISTS vector;
