package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/ai-interview/services/feedback-worker/internal/model"
)

type JCSClient struct {
	BaseURL     string
	InternalKey string
	HTTPClient  *http.Client
}

type PlatformClient = JCSClient

func NewJCSClient(baseURL, internalKey string) *JCSClient {
	return &JCSClient{
		BaseURL:     baseURL,
		InternalKey: internalKey,
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

func NewPlatformClient(baseURL, internalKey string) *PlatformClient {
	return NewJCSClient(baseURL, internalKey)
}

func (c *JCSClient) GetBlueprint(ctx context.Context, blueprintID string) (*model.AssessmentBlueprint, error) {
	// Try canonical /api/v1/blueprints/:id first
	blueprint, err := c.fetchBlueprint(ctx, fmt.Sprintf("%s/api/v1/blueprints/%s", c.BaseURL, blueprintID))
	if err == nil {
		return blueprint, nil
	}

	// Fallback to legacy /api/v1/dockets/:id
	legacyBlueprint, legacyErr := c.fetchBlueprint(ctx, fmt.Sprintf("%s/api/v1/dockets/%s", c.BaseURL, blueprintID))
	if legacyErr == nil {
		return legacyBlueprint, nil
	}

	return nil, fmt.Errorf("get blueprint %s: %w", blueprintID, err)
}

func (c *JCSClient) fetchBlueprint(ctx context.Context, url string) (*model.AssessmentBlueprint, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Internal-Key", c.InternalKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var d model.AssessmentBlueprint
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(d.Criteria) == 0 && len(d.Rubric) > 0 {
		d.Criteria = d.Rubric
	}
	return &d, nil
}

func (c *JCSClient) GetDocket(ctx context.Context, docketID string) (*model.Docket, error) {
	return c.GetBlueprint(ctx, docketID)
}

func (c *JCSClient) GetParticipant(ctx context.Context, participantID string) (*model.Participant, error) {
	// Try canonical /api/v1/participants/:id first
	part, err := c.fetchParticipant(ctx, fmt.Sprintf("%s/api/v1/participants/%s", c.BaseURL, participantID))
	if err == nil {
		return part, nil
	}

	// Fallback to legacy /api/v1/candidates/:id
	legacyPart, legacyErr := c.fetchParticipant(ctx, fmt.Sprintf("%s/api/v1/candidates/%s", c.BaseURL, participantID))
	if legacyErr == nil {
		return legacyPart, nil
	}

	return nil, fmt.Errorf("get participant %s: %w", participantID, err)
}

func (c *JCSClient) fetchParticipant(ctx context.Context, url string) (*model.Participant, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Internal-Key", c.InternalKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var cand model.Participant
	if err := json.NewDecoder(resp.Body).Decode(&cand); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &cand, nil
}

func (c *JCSClient) GetCandidate(ctx context.Context, candidateID string) (*model.Candidate, error) {
	return c.GetParticipant(ctx, candidateID)
}

