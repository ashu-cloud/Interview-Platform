# VoxAssess UI/UX Redesign Blueprint

Status: Design proposal based on the current repository, README, frontend routes, and API documentation. This document is a design spec, not a claim that the proposed screens have been implemented.

## Product design principle

VoxAssess is a low-latency adaptive voice-assessment engine that conducts structured technical and behavioral conversations and produces evaluation reports grounded in verbatim transcript evidence. The UI should communicate confidence, clarity, candidate respect, and technical rigor. It must not feel like a generic HR dashboard or a sci-fi AI toy.

Two different user contexts require two distinct visual modes:
1. Hiring team workspace: information-dense but calm, readable, editorial, decisive.
2. Candidate stage: quiet, immersive, accessible, low-distraction and focused on speaking/listening.

A public marketing site can demonstrate the product, but it must not overshadow the actual workflows.

## Current repository observations

- Front page is a long technical showcase built from HeroMonograph, MetricsPlates, CoverageDeepDive, VoicePipelineFlow, ResilienceGrid, InterfaceShowcaseTabs, TechStackGrid and FinalCTA.
- Current visual tokens include paper/ivory (#fbf9f5), dark ink (#141413), warm gray/hairlines, vermilion, ultramarine and emerald; fonts Space Grotesk, Cormorant Garamond and JetBrains Mono.
- Hiring workspace navigation currently exposes a console/dashboard, candidate dossiers, evaluation tracks, and blueprint studio.
- Existing relevant routes include /auth, /console (also /dashboard and /orchestration aliases), /tracks, /blueprints/new, /candidates, /interview, /interview/[token], /report/[id], /reports/[id].
- Candidate stage already includes invite-token validation, microphone/audio streaming, live text/transcript state, speaking/listening state, mic/camera indicators, transcript drawer, session completion and reconnect/restart-related state.
- API docs define tracks, blueprints, participants/invites, sessions, session turns, evaluation creation/retrieval and invite validation.
- Documented session states: invited, in_progress, completed, feedback_ready. Evaluation states include pending, evaluating, completed, evaluation_failed.
- Do not fabricate backend capabilities. Label new concepts as proposals and only expose controls when supported.

## UX goals

1. Make the next best action obvious from each page.
2. Turn role → blueprint → invite → interview → evaluation into one discoverable lifecycle.
3. Make evidence grounding visible in reports: scores must be inspectable alongside exact candidate quotes and transcript context.
4. Clearly distinguish status from recommendation; never show a score while evaluation is pending.
5. Reduce jargon on the hiring-facing surface. Keep technical diagnostics in optional detail panes.
6. Make candidate readiness, connection, microphone, listening/speaking, interruption, reconnection and completion states unambiguous.
7. Support empty/loading/error/permission-denied/expired-token/evaluation-failed states with recovery paths.
8. Restore standard pointer and keyboard behavior; the current global cursor suppression should not carry over.

## Art direction: “Signal Room”

The redesigned product should look materially different from the current warm-paper monograph UI. Do not preserve the ivory canvas, museum/archival metaphor, pale card grid, or current page composition. Keep the product's technical precision, but express it as a tactile, dark assessment instrument—not as a sci-fi control-room cliché.

### Palette
- App background: deep graphite-green #0D1210
- Sidebar: near-black #090D0B
- Primary surface: #141B17
- Elevated surface: #19221C
- Secondary surface: #202B23
- Main text: soft white-green #EEF4EC
- Secondary text: #9AA99B
- Quiet metadata: #667568
- Brand accent: signal lime #D5FA67
- Accent background: #293719
- Secondary information accent: cool blue #8AC7FF
- Attention: amber #F1C06B
- Error: soft coral #FF8F86
- Candidate stage: deep forest-black with restrained green ambience
Use lime sparingly for active navigation, important actions, live audio states, focused chart traces and evidence selection. It should not fill every icon or become a neon glow effect. Use blue for neutral technical telemetry, amber for attention, coral for errors.

### Typography
- Main UI: Space Grotesk or a similarly confident, highly legible grotesk.
- Long-form report text: highly readable sans-serif at comfortable line height.
- IDs, timing, transcript turn numbers, technical measurements and small status metadata: JetBrains Mono, sparingly.
- Strong scale contrast: confident page titles, compact labels, readable body copy.
- Avoid all-caps for every heading; use it only for small section markers and telemetry metadata.

### Shape, material, composition
- Deep layered surfaces separated by precise 1px borders.
- Medium-small radii (8–12px), not pill-shaped everything.
- Strong typographic hierarchy and asymmetric compositions.
- Data tables should feel like operational ledgers, not generic white spreadsheet tables.
- Use a restrained audio waveform language and subtle animated signal traces where they explain state.
- No blurred glass cards, giant gradient blobs, generic AI sparkle, excessive glow, floating decorative orbs, fake telemetry, permanent custom cursor, or grid wallpaper behind every screen.
- The interface should feel engineered, tactile, precise and considered—with fewer, stronger elements instead of a screen filled with decoration.
- Candidate experience should be immersive and darker than the manager workspace, but share the same signal-lime accent and typography.

### Motion
- UI transitions around 140–220ms.
- Waveform animation must represent a demo/real audio state, not imply live audio where none exists.
- Use clear state transitions for listening, speaking, interruption, reconnecting and completion.
- Respect reduced motion; do not animate decorative elements without purpose.

## Shared application shell

### Hiring workspace shell
- Left sidebar, approximately 248px on desktop, collapsible.
- Top bar: breadcrumbs/page title, global candidate/track search if implemented, notification/refresh only if functional, profile menu.
- Main canvas: page title + one-sentence purpose + primary action at top.
- No duplicated navigation terms or internal engineering implementation details in page headers.
- Footer only where useful. Avoid sticky headers that consume too much vertical space.

Navigation proposal:
- Overview
- Candidates
- Role tracks
- Blueprint library
- Reports
- Settings (only supported settings)

“Create assessment” should be the primary action and guide role/blueprint selection before invite creation. Preserve current route compatibility if the app uses existing route names.

### Candidate shell
- No recruiter sidebar.
- Minimal VoxAssess wordmark, assessment name/role, progress/status and support/help access if available.
- Dark immersive stage during live conversation; light/neutral preflight and completion screens.
- Large, obvious status copy and central waveform/voice visualization.
- No unnecessary candidate-facing scores/recommendation disclosure.

## Complete screen inventory and specification

### A. Public entry

#### A01 — Product landing
Purpose: explain the product and get hiring teams into an assessment workflow.
- Hero statement should communicate adaptive structured voice assessments and evidence-linked reports.
- Primary CTA: “Open workspace” or “Start an assessment” depending on auth model.
- Secondary CTA: “Explore the assessment flow”.
- Product preview should show three steps: configure, converse, review evidence.
- Supporting modules: Adaptive conversation, low-latency voice, structured rubric evaluation, evidence-linked reports.
- Avoid opening with dense architecture specs and unsupported product claims.
- Retain detailed engineering/architecture material as an expandable technical section or separate docs page.

#### A02 — Login/register
- Clear title and form labels.
- Password visibility; inline validation; loading, invalid credentials, network error and success behavior.
- Show whether the action is login or registration.
- Candidate entry must be distinct from hiring-team authentication.
- Existing auth behavior is the source of truth; do not invent SSO/OAuth unless implemented.

#### A03 — Candidate invite entry (/interview)
- Manual token/link entry remains fallback.
- Explain that the personal invite link is normally sufficient.
- Loading/validating state.
- Invalid, expired, already-completed or network-failed states need distinct explanation when the backend exposes the relevant reason.
- Never expose the raw token in unnecessary UI.

### B. Hiring workspace

#### B01 — Overview (/console, /dashboard, /orchestration aliases)
Primary question: “What needs attention now?”
Top section:
- Greeting/page heading “Assessment overview”.
- Primary action “Create assessment”.
- Small context line explaining the operational scope.

Metrics:
- Active role tracks
- Sessions in progress
- Evaluations ready
- Evaluations needing attention

Do not average scores across irrelevant statuses. Avoid a KPI purely to fill a card.

Main panels:
- “Needs your attention”: report ready, evaluation failed, invite expiring/expired if known, session needing recovery if available.
- “Recent candidates”: candidate, role, status, last activity, score only if report is ready, action.
- “Your role tracks”: active track names and counts.
- Empty-state action when no tracks/candidates exist.

Interactions:
- Clicking a metric filters the relevant list.
- All dashboard links go to real screens.
- Refresh/loading/error state remains in place without replacing the page with a blank shell.

#### B02 — Candidate pipeline (/candidates)
Top:
- “Candidates” + purpose line + “Invite candidate”.
- Search by candidate name/email/role.
- Filters: status, role track, date/time if data supports it.
- Clear active-filter chips and a one-click reset.

Table fields:
- Candidate identity and email
- Role track / assessment blueprint
- Invitation/session status
- Invited/last activity date when available
- Score/recommendation only when evaluation is available
- Action: View details / Review report

Status presentation:
- Invited
- In progress
- Completed
- Evaluation pending/evaluating
- Report ready
- Evaluation failed
Do not conflate session completion with report readiness.

Candidate detail/dossier:
- Summary of candidate, role, invite state, session timestamps.
- Session timeline and transcript access.
- Report status and action.
- Re-invite/retry only if supported and authorized.

#### B03 — Role tracks (/tracks)
- List/table first, not just a wall of cards.
- Role title, brief, required skills, associated blueprints/candidate count when available, state.
- Search/sort/filter if list grows.
- New role track form: title, description, required skills.
- Inline validation; preserve entered values after API error.
- Track details include associated blueprints and candidates if available.
- Deletion requires confirmation and explanation when dependent resources block deletion.
- Empty state explains role tracks as assessment contexts.

#### B04 — Blueprint library (/blueprints)
- Separate browse/list screen is a proposed improvement; preserve existing new-builder route.
- Blueprint row: name, role, duration, question count, rubric criteria count, status if available, updated date if available.
- Actions: open/edit/duplicate only if implemented; preview; create new.
- Distinguish saved/published states only if the backend models them.
- No fake “AI-generated” badge.

#### B05 — Blueprint builder (/blueprints/new)
A guided editor with a visible stepper:
1. Role and basics
2. Interview guidance
3. Questions
4. Evaluation rubric
5. Review

Basics:
- Role track selector
- Blueprint title
- Duration
- Instructions/assessment focus

Questions:
- Ordered list of question prompts.
- Question text, type if supported, optional follow-up hints.
- Add, edit, remove, reorder actions (reordering is proposed if not implemented).
- Inline validation for empty questions.
- Allow technical and behavioral question types only if supported by current data model.

Rubric:
- Criteria name
- Description
- Max score
- Order
- Total possible score preview.
- Explain that criteria anchor post-interview evaluation.

Review:
- Candidate-facing summary preview.
- Duration, prompts and criteria overview.
- Clear save/publish CTA using the real backend semantics.
- Warn about missing prompts or rubric before submission.
- Preserve unsaved work on validation errors.
- Do not claim a draft auto-save exists unless implemented.

#### B06 — Invite candidate flow (within /candidates or a dedicated step)
Prefer a focused multi-step modal/page:
1. Candidate details: name, email.
2. Select role track.
3. Select blueprint.
4. Set expiration if supported.
5. Review and create invite.

Success state:
- Candidate identity and assessment name
- Expiry time if provided by API
- Copy invite link
- Copy confirmation
- Send notification status only if notification result is available.
- “Invite another” and “Return to candidate pipeline”.

Error handling:
- Track has no blueprint → link to configure one.
- Invalid email → inline message.
- Expiration invalid → inline message.
- Network/API failure → preserve form and allow retry.
Do not claim an email was sent merely because an invite link was created.

#### B07 — Live session monitor
Only show available telemetry. The platform can expose active session state and turn progression; detailed transport/debug data belongs in optional diagnostics.
- Header: candidate, role/blueprint, status, elapsed duration if reliable.
- Session status: connecting / live / reconnecting / completed / interrupted.
- Turn timeline or turn counter.
- Candidate/assessor speaking event indicator.
- Transcript preview if authorized and data is available.
- Do not enable recruiter controls that interfere with the live session unless backend behavior exists.
- On disconnection, separate a transient connection state from a failed session.

#### B08 — Evaluation report (/report/[id], /reports/[id])
This is the signature product screen.
Header:
- Candidate name and role
- Session date/duration if known
- Report status
- Export/print only if implemented; otherwise mark as later proposal.

Summary:
- Recommendation only when the evaluation has a recommendation.
- Overall score only when report provides a score.
- Criteria coverage and evidence availability.
- Clear explanation that automated findings should be reviewed with supporting evidence.

Main report layout:
- Left/main column: rubric criteria and scores, each with rationale and exact quoted evidence.
- Right/side panel: transcript navigation, turn index and selected excerpt.
- Selecting an evidence quote highlights/jumps to the exact transcript turn.
- Identify speaker and turn number.
- Distinguish the candidate's verbatim words, model-generated rationale and numeric score.
- If a quote fails evidence validation or there is no evidence, show an explicit “No supporting quote available”; never invent a quote.
- Show uncertainty/incomplete criteria if report schema exposes them; otherwise do not add unsupported fields.

Pending:
- “Evaluation in progress”.
- Explain that the transcript is available only if it is already present.
- Do not show an interim score as final.

Failed:
- “We couldn’t generate this report”.
- Display safe error summary if available.
- Retry action only for authorized users and only if existing evaluation API permits it.
- Keep transcript/session record accessible.

#### B09 — Settings/account
- User profile and sign-out.
- Audio/platform preferences only if current settings model supports them.
- Organization/team permissions are proposals, not assumptions.

### C. Candidate assessment

#### C01 — Invite/welcome (/interview/[token] preflight)
- Validate invite before the candidate is asked to prepare.
- Display role/assessment name if provided.
- Explain duration, microphone needs, expected format and whether transcript is displayed.
- Show simple checklist: quiet environment, microphone permission, stable connection.
- Explicit “Begin setup” action.
- Don't request camera permission unless camera is a genuine requirement. The current client code attempts camera capture after invite validation; UX should make the purpose and optional/required nature clear before permission is requested.

#### C02 — Audio and device check
Proposed improvement to the existing stage:
- Microphone permission state
- Live input level meter
- Selected input/device label where supported
- Test voice instruction
- Network/readiness state only if accurately measurable
- Clear retry and browser-permission guidance
- Continue only when mandatory checks pass; provide an accessible fallback when a nonmandatory check fails.

#### C03 — Live assessment (/interview/[token])
The candidate interface should be calm and immersive rather than look like a telemetry dashboard.
Layout:
- Small header with VoxAssess, assessment title and session state.
- Center: animated waveform/listening orb with distinct states.
- Status text: “Listening”, “Thinking”, “Speaking”, “Reconnecting”, “Complete”.
- Live captions/transcript should be readable and optionally collapsible if the product allows this.
- Controls: mic state, help/technical issue, end assessment with confirmation.
- If camera is active, explain why and offer a clear indicator/control consistent with actual requirements.
- Barge-in is a core behavior: interruption must feel immediate, and the UI should switch to listening without confusing overlapping indicators.
- Do not show fake response latency data to the candidate.

States:
- Validating invite
- Ready to begin
- Requesting microphone
- Connecting
- Listening to candidate
- Assessor speaking
- Candidate interrupts assessor
- Recovering/reconnecting
- Assessment complete
- Fatal/nonrecoverable error

#### C04 — Completion
- Confirm completion.
- Thank the candidate without disclosing a recruiter-only recommendation.
- Explain that the hiring team will review the assessment.
- Avoid promises about response timelines unless configured by the hiring team.
- Offer safe close-page instructions.

## End-to-end flow

Hiring team:
Login → Overview → Create/select role track → Create/select blueprint → Review prompts/rubric → Invite candidate → Candidate pipeline → Live status if available → Session completed → Evaluation pending → Report ready → Review score and verbatim evidence.

Candidate:
Open unique invite link → Validate invitation → Read instructions → Check microphone/audio → Begin assessment → Speak/listen/interrupt naturally → Handle reconnect if needed → Finish → Confirmation.

Report flow:
Session complete → asynchronous evaluation triggered/queued → evaluation status pending/evaluating → report completed or evaluation_failed → reviewer inspects criterion score → clicks evidence quote → transcript context highlighted.

## Interaction and component system

Shared components:
- AppShell
- SidebarNavigation
- PageHeader with breadcrumbs and primary action
- SearchAndFilters
- StatusBadge with semantic mapping
- DataTable with empty/loading/error states
- CandidateAvatar/Initials
- TrackRow
- BlueprintStepper
- QuestionEditor
- RubricCriterionEditor
- InviteWizard
- InviteLinkPanel
- SessionStateIndicator
- VoiceWaveform
- AudioLevelMeter
- TranscriptViewer
- EvidenceQuote
- CriterionScorePanel
- RecommendationSummary
- EmptyState
- ErrorState with retry
- ConfirmDialog
- Toast with success/error/neutral semantics
- Skeletons matching content layout

Interaction standards:
- Standard cursor remains enabled.
- Visible keyboard focus and semantic HTML.
- Escape closes dialogs where safe.
- Confirm before irreversible deletion or ending an assessment.
- Toasts must not be the only place an important error is communicated.
- Disabled controls explain why they are disabled.
- Inputs preserve values on errors.
- Avoid loading spinners for the whole screen when a local section is refreshing.
- Respect reduced motion.
- Touch targets should be comfortable on mobile.

## Status semantics

Use status from the real domain model; don't create visual states unsupported by the server:
- invited: neutral/blue
- in_progress: green/active only if connection is verified; otherwise use neutral in-progress
- completed: neutral completed
- feedback_ready/report ready: forest green
- pending/evaluating: amber/neutral progress
- evaluation_failed: muted brick error

The UI must not infer “online” from a stale session record or mark “report ready” solely because the candidate finished.

## Responsive design
- Desktop hiring workspace: persistent sidebar, content max-width around 1440px, wide tables with clear row actions.
- Tablet: sidebar collapses to icon rail/drawer, filters wrap without overlap.
- Mobile: single-column forms, horizontally scrollable data table or stacked rows, accessible drawers.
- Candidate stage: responsive, central content stays visually focused with no tiny controls.
- Report: stack summary then criteria then transcript on narrow screens.

## What not to do
- No generic SaaS gradient hero, floating glass cards or excessive rounded pills.
- No fake operational metrics.
- No “zero hallucinations” guarantee as a UI claim.
- No unreadable microtext just to look technical.
- No default hidden cursor.
- No recruiter controls on candidate stage.
- No candidate score disclosure by accident.
- No mocked backend activity presented as live data.
- No unsupported admin/settings concepts.
- No reimplementation of the backend as part of a UI redesign.

## Suggested implementation order
1. Establish tokens, app shell, cursor/accessibility fixes.
2. Redesign Overview and Candidate Pipeline.
3. Redesign Role Tracks and Blueprint Editor.
4. Redesign Candidate Invite flow.
5. Redesign Candidate preflight and live assessment.
6. Redesign Report viewer with evidence-to-transcript interaction.
7. Add all system states and responsive behavior.
8. Verify existing routes and API wiring are preserved.

## Prototype acceptance checklist
- Can navigate all primary routes.
- Candidate invite flow ends in a useful success panel.
- Role/blueprint editor demonstrates ordered prompts and rubric criteria.
- Live assessment mock clearly demonstrates listening, speaking, interruption and reconnecting states.
- Report includes criterion scores with exact quotes linked to transcript turns.
- Pending and failed evaluation states are visually distinct.
- All sample/demo data is labeled as prototype data.
- All controls are either functional or visibly marked as nonfunctional prototype elements.
- Keyboard navigation and normal mouse cursor work.
- Existing API routes and data contracts are not changed by a visual redesign.
