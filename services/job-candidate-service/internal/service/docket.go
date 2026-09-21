package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/repository"
)

// DocketService implements business logic for dockets.
type DocketService struct {
	repo    *repository.DocketRepository
	jobRepo *repository.JobRepository
}

// NewDocketService constructs a DocketService.
func NewDocketService(repo *repository.DocketRepository, jobRepo *repository.JobRepository) *DocketService {
	return &DocketService{repo: repo, jobRepo: jobRepo}
}

// CreateDocket creates a docket for a job role.
// Verifies the job role belongs to the caller's org before creating.
func (s *DocketService) CreateDocket(
	ctx context.Context,
	orgID, jobID uuid.UUID,
	req *model.CreateDocketRequest,
) (*model.Docket, error) {
	// Ownership check: job must belong to this org
	if _, err := s.jobRepo.GetByID(ctx, orgID, jobID); err != nil {
		return nil, fmt.Errorf("DocketService.CreateDocket: verify job: %w", err)
	}

	docket, err := s.repo.Create(ctx, orgID, jobID, req)
	if err != nil {
		return nil, fmt.Errorf("DocketService.CreateDocket: %w", err)
	}
	return docket, nil
}

// GetDocket fetches a docket scoped to the caller's org.
func (s *DocketService) GetDocket(ctx context.Context, orgID, docketID uuid.UUID) (*model.Docket, error) {
	docket, err := s.repo.GetByID(ctx, orgID, docketID)
	if err != nil {
		return nil, fmt.Errorf("DocketService.GetDocket: %w", err)
	}
	return docket, nil
}

// GetDocketInternal fetches a docket without org scoping (called by the orchestrator).
// This endpoint is protected by the X-Internal-Key header middleware.
func (s *DocketService) GetDocketInternal(ctx context.Context, docketID uuid.UUID) (*model.Docket, error) {
	docket, err := s.repo.GetByIDInternal(ctx, docketID)
	if err != nil {
		return nil, fmt.Errorf("DocketService.GetDocketInternal: %w", err)
	}
	return docket, nil
}
