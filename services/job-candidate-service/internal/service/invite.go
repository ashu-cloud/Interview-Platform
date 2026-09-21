package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/apperror"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/repository"
)

const inviteTokenCacheKeyPrefix = "invite:token:"

// InviteService implements business logic for invites, interviews, and feedback.
type InviteService struct {
	repo          *repository.InviteRepository
	candidateRepo *repository.CandidateRepository
	docketRepo    *repository.DocketRepository
	redis         *redis.Client
	appDomain     string
}

// NewInviteService constructs an InviteService.
func NewInviteService(
	repo *repository.InviteRepository,
	candidateRepo *repository.CandidateRepository,
	docketRepo *repository.DocketRepository,
	redisClient *redis.Client,
	appDomain string,
) *InviteService {
	return &InviteService{
		repo:          repo,
		candidateRepo: candidateRepo,
		docketRepo:    docketRepo,
		redis:         redisClient,
		appDomain:     appDomain,
	}
}

// CreateInvite creates an interview invite for a candidate.
//
// Ownership validation (correction #3):
//  1. candidate must belong to orgID
//  2. docket must belong to orgID
//  3. docket.job_role_id must match candidate.job_role_id
func (s *InviteService) CreateInvite(
	ctx context.Context,
	orgID, candidateID uuid.UUID,
	req *model.CreateInviteRequest,
) (*model.InviteResponse, error) {
	// 1. Verify candidate belongs to this org
	candidate, err := s.candidateRepo.GetByID(ctx, orgID, candidateID)
	if err != nil {
		return nil, fmt.Errorf("InviteService.CreateInvite: get candidate: %w", err)
	}

	// 2. Parse and verify docket belongs to this org
	docketID, err := uuid.Parse(req.DocketID)
	if err != nil {
		return nil, fmt.Errorf("InviteService.CreateInvite: parse docket_id: %w", apperror.ErrBadRequest)
	}
	docket, err := s.docketRepo.GetByID(ctx, orgID, docketID)
	if err != nil {
		return nil, fmt.Errorf("InviteService.CreateInvite: get docket: %w", err)
	}

	// 3. Verify docket and candidate are for the same job role
	if docket.JobRoleID != candidate.JobRoleID {
		return nil, fmt.Errorf("InviteService.CreateInvite: docket job_role_id does not match candidate: %w", apperror.ErrBadRequest)
	}

	// Generate token and compute expiry
	token := uuid.New().String()
	expiresAt := time.Now().UTC().Add(time.Duration(req.ExpiresInDays) * 24 * time.Hour)

	invite, interview, err := s.repo.CreateInviteAndInterview(ctx, candidateID, docketID, orgID, token, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("InviteService.CreateInvite: create: %w", err)
	}

	// Cache in Redis — TTL = time until expiry
	ttl := time.Until(expiresAt)
	cacheVal := model.InviteCacheValue{
		InviteID:    invite.ID.String(),
		InterviewID: interview.ID.String(),
		CandidateID: candidateID.String(),
		DocketID:    docketID.String(),
		OrgID:       orgID.String(),
		ExpiresAt:   expiresAt.Format(time.RFC3339),
	}
	if data, err := json.Marshal(cacheVal); err == nil {
		if redisErr := s.redis.Set(ctx, inviteTokenCacheKeyPrefix+token, data, ttl).Err(); redisErr != nil {
			slog.Warn("InviteService.CreateInvite: failed to cache invite", "token", token, "error", redisErr)
		}
	}

	return &model.InviteResponse{
		InviteID:   invite.ID.String(),
		InviteLink: fmt.Sprintf("%s/interview/%s", s.appDomain, token),
		Token:      token,
		ExpiresAt:  expiresAt,
	}, nil
}

// ValidateToken validates an invite token.
// Tries Redis first; falls back to DB on cache miss.
// Returns ErrNotFound for unknown tokens, ErrConflict for used/completed tokens.
// In-progress interviews remain valid to permit candidates to reconnect.
func (s *InviteService) ValidateToken(ctx context.Context, token string) (*model.InviteValidateResponse, error) {
	// Try Redis cache first
	raw, err := s.redis.Get(ctx, inviteTokenCacheKeyPrefix+token).Bytes()
	if err == nil {
		// Cache hit — token exists in Redis (not yet completed, not yet expired by TTL)
		var cached model.InviteCacheValue
		if jsonErr := json.Unmarshal(raw, &cached); jsonErr == nil {
			expiresAt, _ := time.Parse(time.RFC3339, cached.ExpiresAt)
			// Resolve candidate/job details from DB for full response
			return s.buildValidateResponseFromDB(ctx, token, expiresAt)
		}
	}

	// Cache miss — fall back to DB
	return s.buildValidateResponseFromDB(ctx, token, time.Time{})
}

// buildValidateResponseFromDB queries the DB to build the validate response.
// Handles used/expired/completed checks.
func (s *InviteService) buildValidateResponseFromDB(
	ctx context.Context,
	token string,
	cachedExpiresAt time.Time,
) (*model.InviteValidateResponse, error) {
	details, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("InviteService.ValidateToken: %w", err)
	}

	// Make sure a completed interview cannot be rejoined
	if details.UsedAt != nil || details.InterviewStatus == "completed" || details.InterviewStatus == "feedback_ready" {
		return nil, fmt.Errorf("InviteService.ValidateToken: interview already completed: %w", apperror.ErrConflict)
	}

	expiresAt := details.ExpiresAt
	if !cachedExpiresAt.IsZero() {
		expiresAt = cachedExpiresAt
	}
	if time.Now().After(expiresAt) {
		return nil, fmt.Errorf("InviteService.ValidateToken: invite expired: %w", apperror.ErrNotFound)
	}

	return &model.InviteValidateResponse{
		Valid:           true,
		InterviewID:     details.InterviewID.String(),
		InterviewStatus: details.InterviewStatus,
		DocketID:        details.DocketID.String(),
		OrgID:           details.OrgID.String(),
		Candidate: model.InviteValidateCandidate{
			ID:       details.CandidateID.String(),
			FullName: details.CandidateFullName,
		},
		JobRole:     model.InviteValidateJobRole{Title: details.JobTitle},
		DocketTitle: details.DocketTitle,
		ExpiresAt:   expiresAt,
	}, nil
}

// MarkInviteUsed marks an invite as used and invalidates the Redis cache.
//
// Cache invalidation (correction #2):
//  1. Fetch invite from DB to retrieve the token string
//  2. Mark used_at in DB
//  3. DEL invite:token:{token} from Redis
//     Redis Del failure is logged but does not fail the operation.
func (s *InviteService) MarkInviteUsed(ctx context.Context, inviteID uuid.UUID) error {
	// Fetch to get token for cache key
	invite, err := s.repo.GetByInviteID(ctx, inviteID)
	if err != nil {
		return fmt.Errorf("InviteService.MarkInviteUsed: get invite: %w", err)
	}

	// Mark used in DB
	if err := s.repo.MarkUsed(ctx, inviteID, time.Now().UTC()); err != nil {
		return fmt.Errorf("InviteService.MarkInviteUsed: mark used: %w", err)
	}

	// Invalidate Redis cache — log on failure, do not propagate
	if redisErr := s.redis.Del(ctx, inviteTokenCacheKeyPrefix+invite.Token).Err(); redisErr != nil {
		slog.Warn("InviteService.MarkInviteUsed: failed to delete cache key",
			"invite_id", inviteID, "error", redisErr)
	}

	return nil
}

// UpdateInterviewStatus updates an interview's status.
// NOTE: Spec gap resolution — the spec requires this operation but defines no HTTP endpoint.
// Semantics:
// - in_progress: Started timestamp set, invite remains active and valid for reconnect.
// - completed / feedback_ready: Completed timestamp set, invite is marked used, and Redis cache is invalidated.
func (s *InviteService) UpdateInterviewStatus(
	ctx context.Context,
	interviewID uuid.UUID,
	req *model.UpdateInterviewStatusRequest,
) (bool, error) {
	changed, err := s.repo.UpdateInterviewStatus(ctx, interviewID, req.Status)
	if err != nil {
		return false, fmt.Errorf("InviteService.UpdateInterviewStatus: %w", err)
	}

	// Only invalidate the invite / mark it consumed once the interview is actually completed
	// and only if this instance actually transitioned the status
	if changed && (req.Status == "completed" || req.Status == "feedback_ready") {
		invite, err := s.repo.GetInviteByInterviewID(ctx, interviewID)
		if err == nil {
			_ = s.repo.MarkUsed(ctx, invite.ID, time.Now().UTC())
			if redisErr := s.redis.Del(ctx, inviteTokenCacheKeyPrefix+invite.Token).Err(); redisErr != nil {
				slog.Warn("InviteService.UpdateInterviewStatus: failed to delete cache key",
					"interview_id", interviewID, "error", redisErr)
			}
		}
	}

	return changed, nil
}

// ListInterviews returns paginated interviews for an org.
func (s *InviteService) ListInterviews(
	ctx context.Context,
	orgID uuid.UUID,
	statusFilter, jobIDFilter string,
	page, limit int,
) (*model.InterviewListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	interviews, total, err := s.repo.GetInterviews(ctx, orgID, statusFilter, jobIDFilter, page, limit)
	if err != nil {
		return nil, fmt.Errorf("InviteService.ListInterviews: %w", err)
	}
	return &model.InterviewListResponse{
		Interviews: interviews,
		Total:      total,
		Page:       page,
		Limit:      limit,
	}, nil
}

// GetInterviewByID returns a single interview scoped to the org.
func (s *InviteService) GetInterviewByID(
	ctx context.Context,
	orgID, interviewID uuid.UUID,
) (*model.InterviewDetail, error) {
	d, err := s.repo.GetInterviewByID(ctx, orgID, interviewID)
	if err != nil {
		return nil, fmt.Errorf("InviteService.GetInterviewByID: %w", err)
	}
	return d, nil
}

// GetFeedback returns the structured feedback for an interview.
func (s *InviteService) GetFeedback(
	ctx context.Context,
	orgID, interviewID uuid.UUID,
) (*model.FeedbackDetailResponse, error) {
	resp, err := s.repo.GetFeedback(ctx, orgID, interviewID)
	if err != nil {
		return nil, fmt.Errorf("InviteService.GetFeedback: %w", err)
	}
	return resp, nil
}
