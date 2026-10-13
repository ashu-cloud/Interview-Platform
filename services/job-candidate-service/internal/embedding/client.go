// Package embedding turns text into normalized unit vectors for pgvector storage.
package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://api.openai.com/v1"
	DefaultModel   = "text-embedding-3-small"
	DefaultDims    = 768
)

// EmbeddingClient embeds text.
type EmbeddingClient interface {
	// Embed returns a normalized unit vector for the given text.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// OpenAIEmbeddingClient calls an OpenAI-compatible /embeddings endpoint.
type OpenAIEmbeddingClient struct {
	apiKey     string
	baseURL    string
	model      string
	dims       int
	httpClient *http.Client
}

// NewOpenAIEmbeddingClient builds a client with default model and base URL.
func NewOpenAIEmbeddingClient(apiKey string, dims int) *OpenAIEmbeddingClient {
	if dims <= 0 {
		dims = DefaultDims
	}
	return &OpenAIEmbeddingClient{
		apiKey:     apiKey,
		baseURL:    DefaultBaseURL,
		model:      DefaultModel,
		dims:       dims,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// WithModel overrides the embedding model (ignored when empty).
func (c *OpenAIEmbeddingClient) WithModel(model string) *OpenAIEmbeddingClient {
	if strings.TrimSpace(model) != "" {
		c.model = strings.TrimSpace(model)
	}
	return c
}

// WithBaseURL overrides the API base URL (ignored when empty).
func (c *OpenAIEmbeddingClient) WithBaseURL(baseURL string) *OpenAIEmbeddingClient {
	if b := strings.TrimRight(strings.TrimSpace(baseURL), "/"); b != "" {
		c.baseURL = b
	}
	return c
}

type embeddingRequest struct {
	Model      string `json:"model"`
	Input      string `json:"input"`
	Dimensions int    `json:"dimensions,omitempty"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// Embed implements EmbeddingClient.
func (c *OpenAIEmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("embed: empty text")
	}
	body, err := json.Marshal(embeddingRequest{Model: c.model, Input: text, Dimensions: c.dims})
	if err != nil {
		return nil, fmt.Errorf("embed: marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embed: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embed: request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("embed: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		msg := string(raw)
		if len(msg) > 300 {
			msg = msg[:300] + "..."
		}
		return nil, fmt.Errorf("embed: unexpected status %d: %s", resp.StatusCode, msg)
	}
	var parsed embeddingResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("embed: decode response: %w", err)
	}
	if len(parsed.Data) == 0 || len(parsed.Data[0].Embedding) == 0 {
		return nil, errors.New("embed: response contained no embedding")
	}
	vec := parsed.Data[0].Embedding
	if len(vec) != c.dims {
		return nil, fmt.Errorf("embed: got %d dimensions, want %d", len(vec), c.dims)
	}
	return Normalize(vec), nil
}

// DemoEmbeddingClient returns a random unit vector so the pipeline runs
// end to end without an API key. Coverage scores derived from it are noise.
type DemoEmbeddingClient struct{}

// Embed implements EmbeddingClient.
func (DemoEmbeddingClient) Embed(_ context.Context, _ string) ([]float32, error) {
	vec := make([]float32, DefaultDims)
	for i := range vec {
		vec[i] = rand.Float32()*2 - 1
	}
	return Normalize(vec), nil
}

// Normalize scales vec to unit length in place and returns it.
func Normalize(vec []float32) []float32 {
	var sum float64
	for _, v := range vec {
		sum += float64(v) * float64(v)
	}
	if sum == 0 {
		return vec
	}
	norm := float32(math.Sqrt(sum))
	for i := range vec {
		vec[i] /= norm
	}
	return vec
}
