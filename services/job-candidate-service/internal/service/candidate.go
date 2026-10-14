package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/repository"
)

// CandidateService implements business logic for candidates.
type CandidateService struct {
	repo    *repository.CandidateRepository
	jobRepo *repository.JobRepository
}

// NewCandidateService constructs a CandidateService.
func NewCandidateService(repo *repository.CandidateRepository, jobRepo *repository.JobRepository) *CandidateService {
	return &CandidateService{repo: repo, jobRepo: jobRepo}
}

// CreateCandidate adds a candidate to a job role.
// Verifies the job belongs to the caller's org before inserting.
// Returns ErrConflict if candidate email already exists for this job.
func (s *CandidateService) CreateCandidate(
	ctx context.Context,
	orgID, jobID uuid.UUID,
	req *model.CreateCandidateRequest,
) (*model.Candidate, error) {
	// Ownership check: job must belong to this org
	if _, err := s.jobRepo.GetByID(ctx, orgID, jobID); err != nil {
		return nil, fmt.Errorf("CandidateService.CreateCandidate: verify job: %w", err)
	}

	c, err := s.repo.Create(ctx, orgID, jobID, req)
	if err != nil {
		return nil, fmt.Errorf("CandidateService.CreateCandidate: %w", err)
	}
	return c, nil
}

// ListCandidates returns all candidates for a job role in this org.
func (s *CandidateService) ListCandidates(
	ctx context.Context,
	orgID, jobID uuid.UUID,
) ([]model.CandidateWithStatus, error) {
	// Ownership check: job must belong to this org
	if _, err := s.jobRepo.GetByID(ctx, orgID, jobID); err != nil {
		return nil, fmt.Errorf("CandidateService.ListCandidates: verify job: %w", err)
	}

	candidates, err := s.repo.ListByJob(ctx, orgID, jobID)
	if err != nil {
		return nil, fmt.Errorf("CandidateService.ListCandidates: %w", err)
	}
	return candidates, nil
}

// GetCandidate returns a candidate scoped to the caller's org.
func (s *CandidateService) GetCandidate(ctx context.Context, orgID, candidateID uuid.UUID) (*model.Candidate, error) {
	candidate, err := s.repo.GetByID(ctx, orgID, candidateID)
	if err != nil {
		return nil, fmt.Errorf("CandidateService.GetCandidate: %w", err)
	}
	return candidate, nil
}

// GetByIDInternal fetches a candidate internally (without org check).
func (s *CandidateService) GetByIDInternal(ctx context.Context, candidateID uuid.UUID) (*model.Candidate, error) {
	c, err := s.repo.GetByIDInternal(ctx, candidateID)
	if err != nil {
		return nil, fmt.Errorf("CandidateService.GetByIDInternal: %w", err)
	}
	return c, nil
}
