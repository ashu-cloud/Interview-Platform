package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"
)

type JCSClient interface {
	ValidateInvite(ctx context.Context, token string) (*InviteValidateResponse, error)
	GetDocket(ctx context.Context, docketID string) (*model.Docket, error)
	UpdateInterviewStatus(ctx context.Context, interviewID string, status string) (bool, error)
	MarkInviteUsed(ctx context.Context, inviteID string) error
}

type jcsClient struct {
	baseURL        string
	internalAPIKey string
	httpClient     *http.Client
}

func NewJCSClient(baseURL, internalAPIKey string) JCSClient {
	return &jcsClient{
		baseURL:        baseURL,
		internalAPIKey: internalAPIKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type InviteValidateCandidate struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
}

type InviteValidateJobRole struct {
	Title string `json:"title"`
}

type InviteValidateResponse struct {
	Valid           bool                    `json:"valid"`
	InterviewID     string                  `json:"interview_id"`
	InterviewStatus string                  `json:"interview_status"`
	DocketID        string                  `json:"docket_id"`
	OrgID           string                  `json:"org_id"`
	Candidate       InviteValidateCandidate `json:"candidate"`
	JobRole         InviteValidateJobRole   `json:"job_role"`
	DocketTitle     string                  `json:"docket_title"`
	ExpiresAt       time.Time               `json:"expires_at"`
}

func (c *jcsClient) ValidateInvite(ctx context.Context, token string) (*InviteValidateResponse, error) {
	url := fmt.Sprintf("%s/api/v1/invites/validate/%s", c.baseURL, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(b))
	}

	var res InviteValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return &res, nil
}

func (c *jcsClient) GetDocket(ctx context.Context, docketID string) (*model.Docket, error) {
	url := fmt.Sprintf("%s/api/v1/dockets/%s", c.baseURL, docketID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("X-Internal-Key", c.internalAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(b))
	}

	var res model.Docket
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return &res, nil
}

func (c *jcsClient) UpdateInterviewStatus(ctx context.Context, interviewID string, status string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/interviews/%s/status", c.baseURL, interviewID)
	
	body := map[string]string{"status": status}
	b, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(b))
	if err != nil {
		return false, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("X-Internal-Key", c.internalAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(b))
	}

	var res struct {
		Message string `json:"message"`
		Changed bool   `json:"changed"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, fmt.Errorf("decode response failed: %w", err)
	}

	return res.Changed, nil
}

func (c *jcsClient) MarkInviteUsed(ctx context.Context, inviteID string) error {
	url := fmt.Sprintf("%s/api/v1/invites/%s/use", c.baseURL, inviteID)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("X-Internal-Key", c.internalAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(b))
	}

	return nil
}
