-- Reverts the table renames (old names: job_roles, dockets, docket_questions,
-- rubric_criteria, candidates, interview_invites, interviews, feedback_reports,
-- feedback_scores).
ALTER TABLE evaluation_scores    RENAME TO feedback_scores;
ALTER TABLE evaluation_reports   RENAME TO feedback_reports;
ALTER TABLE assessment_sessions  RENAME TO interviews;
ALTER TABLE assessment_invites   RENAME TO interview_invites;
ALTER TABLE participants         RENAME TO candidates;
ALTER TABLE evaluation_criteria  RENAME TO rubric_criteria;
ALTER TABLE blueprint_prompts    RENAME TO docket_questions;
ALTER TABLE assessment_blueprints RENAME TO dockets;
ALTER TABLE assessment_tracks    RENAME TO job_roles;
