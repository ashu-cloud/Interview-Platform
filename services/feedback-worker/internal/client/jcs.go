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
	BaseURL    string
	InternalKey string
	HTTPClient *http.Client
}

func NewJCSClient(baseURL, internalKey string) *JCSClient {
	return &JCSClient{
		BaseURL:    baseURL,
		InternalKey: internalKey,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *JCSClient) GetDocket(ctx context.Context, docketID string) (*model.Docket, error) {
	url := fmt.Sprintf("%s/api/v1/dockets/%s", c.BaseURL, docketID)
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

	var d model.Docket
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &d, nil
}

func (c *JCSClient) GetCandidate(ctx context.Context, candidateID string) (*model.Candidate, error) {
	url := fmt.Sprintf("%s/api/v1/candidates/%s", c.BaseURL, candidateID)
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

	var cand model.Candidate
	if err := json.NewDecoder(resp.Body).Decode(&cand); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &cand, nil
}
