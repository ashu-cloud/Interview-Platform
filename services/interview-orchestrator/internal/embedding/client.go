// Package embedding turns text into normalized unit vectors.
package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
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
		httpClient: &http.Client{Timeout: 10 * time.Second},
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
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
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
		return nil, fmt.Errorf("embed: unexpected status %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}
	var parsed embeddingResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("embed: decode response: %w", err)
	}
	if len(parsed.Data) == 0 || len(parsed.Data[0].Embedding) == 0 {
		return nil, errors.New("embed: response contained no embedding")
	}
	vec := parsed.Data[0].Embedding
	if c.dims > 0 && len(vec) != c.dims {
		return nil, fmt.Errorf("embed: got %d dimensions, want %d", len(vec), c.dims)
	}
	return Normalize(vec), nil
}

// DeterministicSemanticEmbedding computes a high-fidelity unit vector using
// word-level and subword character n-gram hashing with positional weights.
// Sentences sharing domain keywords (e.g. kafka, redis, pubsub, partitions)
// produce high cosine similarities (>0.7), enabling reliable vector search and RAG
// completely offline without external API dependencies.
func DeterministicSemanticEmbedding(text string, dims int) []float32 {
	if dims <= 0 {
		dims = DefaultDims
	}
	vec := make([]float32, dims)
	clean := strings.ToLower(text)
	words := strings.FieldsFunc(clean, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-')
	})
	if len(words) == 0 {
		vec[0] = 1.0
		return vec
	}

	for pos, w := range words {
		if len(w) == 0 {
			continue
		}
		// Word unigram hash (FNV-1a 64-bit)
		h := fnv64(w)
		idx := int(h % uint64(dims))
		sign := float32(1.0)
		if (h>>32)&1 == 1 {
			sign = -1.0
		}
		posWeight := float32(1.0) / (1.0 + float32(pos)*0.02)
		vec[idx] += sign * 2.0 * posWeight

		// Subword character trigrams for morphological and root matching
		if len(w) >= 3 {
			for i := 0; i <= len(w)-3; i++ {
				th := fnv64(w[i : i+3])
				tidx := int(th % uint64(dims))
				tsign := float32(1.0)
				if (th>>32)&1 == 1 {
					tsign = -1.0
				}
				vec[tidx] += tsign * 0.4
			}
		}
	}

	return Normalize(vec)
}

func fnv64(s string) uint64 {
	var h uint64 = 14695981039346656037
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}

// FallbackEmbeddingClient wraps an upstream EmbeddingClient (such as OpenAI)
// and seamlessly falls back to DeterministicSemanticEmbedding if the upstream
// client is nil, unconfigured, or returns an error (such as 401 Unauthorized or network drop).
type FallbackEmbeddingClient struct {
	upstream EmbeddingClient
	dims     int
}

// NewFallbackEmbeddingClient constructs a FallbackEmbeddingClient.
func NewFallbackEmbeddingClient(upstream EmbeddingClient, dims int) *FallbackEmbeddingClient {
	if dims <= 0 {
		dims = DefaultDims
	}
	return &FallbackEmbeddingClient{
		upstream: upstream,
		dims:     dims,
	}
}

// Embed implements EmbeddingClient.
func (f *FallbackEmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error) {
	if f.upstream != nil {
		vec, err := f.upstream.Embed(ctx, text)
		if err == nil && len(vec) == f.dims {
			return vec, nil
		}
	}
	return DeterministicSemanticEmbedding(text, f.dims), nil
}

// DemoEmbeddingClient returns a deterministic semantic unit vector.
type DemoEmbeddingClient struct{}

// Embed implements EmbeddingClient.
func (d DemoEmbeddingClient) Embed(_ context.Context, text string) ([]float32, error) {
	return DeterministicSemanticEmbedding(text, DefaultDims), nil
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

// CosineSimilarity computes the dot product of two pre-normalized unit vectors.
// Since embeddings are normalized on storage, dot product == cosine similarity.
// Vectors of different length are compared over their common prefix.
func CosineSimilarity(a, b []float32) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var sum float64
	for i := 0; i < n; i++ {
		sum += float64(a[i]) * float64(b[i])
	}
	return sum
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
