-- Rebrand: rename tables to assessment vocabulary.
-- Foreign keys, indexes and sequences follow the renamed tables automatically.
-- Column names are intentionally unchanged in this migration.
ALTER TABLE job_roles          RENAME TO assessment_tracks;
ALTER TABLE dockets            RENAME TO assessment_blueprints;
ALTER TABLE docket_questions   RENAME TO blueprint_prompts;
ALTER TABLE rubric_criteria    RENAME TO evaluation_criteria;
ALTER TABLE candidates         RENAME TO participants;
ALTER TABLE interview_invites  RENAME TO assessment_invites;
ALTER TABLE interviews         RENAME TO assessment_sessions;
ALTER TABLE feedback_reports   RENAME TO evaluation_reports;
ALTER TABLE feedback_scores    RENAME TO evaluation_scores;
