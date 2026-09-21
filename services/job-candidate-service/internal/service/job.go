package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/repository"
)

// JobService implements business logic for job roles.
type JobService struct {
	repo *repository.JobRepository
}

// NewJobService constructs a JobService.
func NewJobService(repo *repository.JobRepository) *JobService {
	return &JobService{repo: repo}
}

// CreateJob creates a new job role owned by the authenticated user's org.
func (s *JobService) CreateJob(
	ctx context.Context,
	orgID, userID uuid.UUID,
	req *model.CreateJobRequest,
) (*model.JobRole, error) {
	job, err := s.repo.Create(ctx, orgID, userID, req)
	if err != nil {
		return nil, fmt.Errorf("JobService.CreateJob: %w", err)
	}
	return job, nil
}

// ListJobs returns paginated job roles for an org with optional status filter.
func (s *JobService) ListJobs(
	ctx context.Context,
	orgID uuid.UUID,
	status string,
	page, limit int,
) (*model.JobListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	jobs, total, err := s.repo.List(ctx, orgID, status, page, limit)
	if err != nil {
		return nil, fmt.Errorf("JobService.ListJobs: %w", err)
	}
	return &model.JobListResponse{
		Jobs:  jobs,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

// GetJob returns a single job role scoped to the org.
func (s *JobService) GetJob(ctx context.Context, orgID, jobID uuid.UUID) (*model.JobRole, error) {
	job, err := s.repo.GetByID(ctx, orgID, jobID)
	if err != nil {
		return nil, fmt.Errorf("JobService.GetJob: %w", err)
	}
	return job, nil
}

// PatchJob applies a partial update to a job role.
func (s *JobService) PatchJob(
	ctx context.Context,
	orgID, jobID uuid.UUID,
	req *model.PatchJobRequest,
) (*model.JobRole, error) {
	job, err := s.repo.Patch(ctx, orgID, jobID, req)
	if err != nil {
		return nil, fmt.Errorf("JobService.PatchJob: %w", err)
	}
	return job, nil
}
