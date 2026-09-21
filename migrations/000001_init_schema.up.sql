-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Organizations (tenants)
CREATE TABLE organizations (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Hiring Managers (users)
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    full_name       VARCHAR(255) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_org_id ON users(org_id);

-- Job Roles
CREATE TABLE job_roles (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_by      UUID NOT NULL REFERENCES users(id),
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    required_skills TEXT[],
    status          VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_job_roles_org_id ON job_roles(org_id);

-- Dockets (interview question bank + rubric per job role)
CREATE TABLE dockets (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_role_id     UUID NOT NULL REFERENCES job_roles(id) ON DELETE CASCADE,
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    title           VARCHAR(255) NOT NULL,
    instructions    TEXT,           -- system prompt context for LLM
    duration_mins   INT NOT NULL DEFAULT 30,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dockets_job_role_id ON dockets(job_role_id);

-- Docket Questions
CREATE TABLE docket_questions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    docket_id       UUID NOT NULL REFERENCES dockets(id) ON DELETE CASCADE,
    question_text   TEXT NOT NULL,
    question_type   VARCHAR(50) NOT NULL CHECK (question_type IN ('technical', 'behavioral', 'situational')),
    order_index     INT NOT NULL,
    follow_up_hint  TEXT,           -- hint for LLM on what to probe deeper
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_docket_questions_docket_id ON docket_questions(docket_id);

-- Rubric Criteria (scoring dimensions)
CREATE TABLE rubric_criteria (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    docket_id       UUID NOT NULL REFERENCES dockets(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,   -- e.g. "Problem Solving", "Communication"
    description     TEXT NOT NULL,
    max_score       INT NOT NULL DEFAULT 10,
    order_index     INT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rubric_criteria_docket_id ON rubric_criteria(docket_id);

-- Candidates
CREATE TABLE candidates (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    job_role_id     UUID NOT NULL REFERENCES job_roles(id),
    email           VARCHAR(255) NOT NULL,
    full_name       VARCHAR(255) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, email, job_role_id)
);

CREATE INDEX idx_candidates_org_id ON candidates(org_id);
CREATE INDEX idx_candidates_job_role_id ON candidates(job_role_id);

-- Interview Invites (unique time-bound links)
CREATE TABLE interview_invites (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidate_id    UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    docket_id       UUID NOT NULL REFERENCES dockets(id),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    token           VARCHAR(255) NOT NULL UNIQUE,   -- JWT or random UUID used in link
    expires_at      TIMESTAMPTZ NOT NULL,
    used_at         TIMESTAMPTZ,                    -- null = not yet used
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_interview_invites_token ON interview_invites(token);
CREATE INDEX idx_interview_invites_candidate_id ON interview_invites(candidate_id);

-- Interviews
CREATE TABLE interviews (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invite_id       UUID NOT NULL REFERENCES interview_invites(id),
    candidate_id    UUID NOT NULL REFERENCES candidates(id),
    docket_id       UUID NOT NULL REFERENCES dockets(id),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    status          VARCHAR(50) NOT NULL DEFAULT 'invited'
                    CHECK (status IN ('invited', 'in_progress', 'completed', 'feedback_ready')),
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_interviews_org_id ON interviews(org_id);
CREATE INDEX idx_interviews_candidate_id ON interviews(candidate_id);
CREATE INDEX idx_interviews_status ON interviews(status);

-- Feedback Reports
CREATE TABLE feedback_reports (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    interview_id    UUID NOT NULL UNIQUE REFERENCES interviews(id),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    summary         TEXT NOT NULL,
    overall_score   NUMERIC(4,2) NOT NULL,      -- e.g. 7.50
    recommendation  VARCHAR(50) NOT NULL CHECK (recommendation IN ('strong_hire', 'hire', 'no_hire', 'strong_no_hire')),
    transcript_json JSONB NOT NULL,             -- full conversation [{role, content, timestamp}]
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Feedback Scores (per rubric criterion)
CREATE TABLE feedback_scores (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    feedback_id     UUID NOT NULL REFERENCES feedback_reports(id) ON DELETE CASCADE,
    criterion_id    UUID NOT NULL REFERENCES rubric_criteria(id),
    score           INT NOT NULL,
    evidence        TEXT NOT NULL,          -- quote from transcript supporting score
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_feedback_scores_feedback_id ON feedback_scores(feedback_id);
