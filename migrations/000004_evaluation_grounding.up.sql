ALTER TABLE evaluation_scores ADD COLUMN IF NOT EXISTS
    is_grounded BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE evaluation_scores ADD COLUMN IF NOT EXISTS
    evidence_turn_id UUID REFERENCES session_turns(id);

ALTER TABLE assessment_sessions DROP CONSTRAINT IF EXISTS interviews_status_check;
ALTER TABLE assessment_sessions DROP CONSTRAINT IF EXISTS assessment_sessions_status_check;
ALTER TABLE assessment_sessions ADD CONSTRAINT assessment_sessions_status_check
    CHECK (status IN ('invited', 'in_progress', 'completed', 'feedback_ready', 'evaluation_ready'));

