package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/embedding"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/repository"
)

// embeddingJobTimeout bounds one blueprint's background embedding run.
const embeddingJobTimeout = 90 * time.Second

// DocketService implements business logic for dockets.
type DocketService struct {
	repo            *repository.DocketRepository
	jobRepo         *repository.JobRepository
	embeddingClient embedding.EmbeddingClient // nil disables embedding generation
	logger          *slog.Logger
}

// NewDocketService constructs a DocketService.
func NewDocketService(repo *repository.DocketRepository, jobRepo *repository.JobRepository) *DocketService {
	return &DocketService{repo: repo, jobRepo: jobRepo, logger: slog.Default()}
}

// SetEmbedder enables background embedding generation for blueprints.
func (s *DocketService) SetEmbedder(client embedding.EmbeddingClient, logger *slog.Logger) {
	s.embeddingClient = client
	if logger != nil {
		s.logger = logger
	}
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
	s.generateEmbeddingsAsync(docket)
	return docket, nil
}

// ListDockets returns blueprints for a track owned by the caller's org.
func (s *DocketService) ListDockets(ctx context.Context, orgID, jobID uuid.UUID) ([]model.Docket, error) {
	if _, err := s.jobRepo.GetByID(ctx, orgID, jobID); err != nil {
		return nil, fmt.Errorf("DocketService.ListDockets: verify job: %w", err)
	}
	dockets, err := s.repo.ListByJob(ctx, orgID, jobID)
	if err != nil {
		return nil, fmt.Errorf("DocketService.ListDockets: %w", err)
	}
	return dockets, nil
}

// UpdateDocket updates a blueprint scoped to the caller's org.
func (s *DocketService) UpdateDocket(
	ctx context.Context,
	orgID, docketID uuid.UUID,
	req *model.UpdateDocketRequest,
) (*model.Docket, error) {
	docket, err := s.repo.Update(ctx, orgID, docketID, req)
	if err != nil {
		return nil, fmt.Errorf("DocketService.UpdateDocket: %w", err)
	}
	// Replacing prompts or criteria creates new rows without embeddings.
	if req.Questions != nil || req.Rubric != nil {
		s.generateEmbeddingsAsync(docket)
	}
	return docket, nil
}

// DeleteDocket removes a blueprint scoped to the caller's org.
func (s *DocketService) DeleteDocket(ctx context.Context, orgID, docketID uuid.UUID) error {
	if err := s.repo.Delete(ctx, orgID, docketID); err != nil {
		return fmt.Errorf("DocketService.DeleteDocket: %w", err)
	}
	return nil
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

// GetCriteriaEmbeddings returns criterion ID -> embedding for the runtime's
// adaptive engine. A blueprint with no embeddings yields an empty map.
func (s *DocketService) GetCriteriaEmbeddings(ctx context.Context, blueprintID uuid.UUID) (map[string][]float32, error) {
	embeddings, err := s.repo.GetCriteriaEmbeddings(ctx, blueprintID)
	if err != nil {
		return nil, fmt.Errorf("DocketService.GetCriteriaEmbeddings: %w", err)
	}
	return embeddings, nil
}

// generateEmbeddingsAsync embeds a blueprint's criteria and prompts and rebuilds
// its document_chunks in the background. It never affects the caller.
func (s *DocketService) generateEmbeddingsAsync(blueprint *model.Docket) {
	if s.embeddingClient == nil || blueprint == nil {
		return
	}
	// Copy so the goroutine does not share the response object with the handler.
	bp := *blueprint
	bp.Rubric = append([]model.RubricCriteria(nil), blueprint.Rubric...)
	bp.Questions = append([]model.DocketQuestion(nil), blueprint.Questions...)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), embeddingJobTimeout)
		defer cancel()

		criterionVecs := make(map[uuid.UUID][]float32)
		promptVecs := make(map[uuid.UUID][]float32)

		for _, c := range bp.Rubric {
			vec, err := s.embeddingClient.Embed(ctx, criterionText(c))
			if err != nil {
				s.logger.Error("embed criterion failed", "id", c.ID, "error", err)
				continue
			}
			criterionVecs[c.ID] = vec
			if err := s.repo.UpdateCriterionEmbedding(ctx, c.ID, vec); err != nil {
				s.logger.Error("save criterion embedding failed", "id", c.ID, "error", err)
			}
		}

		for _, p := range bp.Questions {
			vec, err := s.embeddingClient.Embed(ctx, p.QuestionText)
			if err != nil {
				s.logger.Error("embed prompt failed", "id", p.ID, "error", err)
				continue
			}
			promptVecs[p.ID] = vec
			if err := s.repo.UpdatePromptEmbedding(ctx, p.ID, vec); err != nil {
				s.logger.Error("save prompt embedding failed", "id", p.ID, "error", err)
			}
		}

		// Rebuild document_chunks (RAG knowledge base for this blueprint).
		var chunks []model.DocumentChunk
		for i, p := range bp.Questions {
			if vec, ok := promptVecs[p.ID]; ok {
				chunks = append(chunks, model.DocumentChunk{
					BlueprintID: bp.ID, OrgID: bp.OrgID, ChunkText: p.QuestionText,
					Embedding: vec, ChunkIndex: i, Source: "assessment_prompt",
				})
			}
		}
		for i, c := range bp.Rubric {
			if vec, ok := criterionVecs[c.ID]; ok {
				chunks = append(chunks, model.DocumentChunk{
					BlueprintID: bp.ID, OrgID: bp.OrgID, ChunkText: criterionText(c),
					Embedding: vec, ChunkIndex: i, Source: "evaluation_criterion",
				})
			}
		}
		if err := s.repo.DeleteDocumentChunks(ctx, bp.ID); err != nil {
			s.logger.Error("clear document chunks failed", "blueprint_id", bp.ID, "error", err)
			return
		}
		if err := s.repo.InsertDocumentChunks(ctx, chunks); err != nil {
			s.logger.Error("insert document chunks failed", "blueprint_id", bp.ID, "error", err)
			return
		}
		s.logger.Info("blueprint embeddings generated", "blueprint_id", bp.ID,
			"criteria", len(criterionVecs), "prompts", len(promptVecs), "chunks", len(chunks))
	}()
}

func criterionText(c model.RubricCriteria) string {
	return c.Name + ": " + c.Description
}
