ALTER TABLE assessment_sessions DROP CONSTRAINT IF EXISTS assessment_sessions_status_check;
ALTER TABLE assessment_sessions ADD CONSTRAINT interviews_status_check
    CHECK (status IN ('invited', 'in_progress', 'completed', 'feedback_ready'));
ALTER TABLE evaluation_scores DROP COLUMN IF EXISTS evidence_turn_id;
ALTER TABLE evaluation_scores DROP COLUMN IF EXISTS is_grounded;

