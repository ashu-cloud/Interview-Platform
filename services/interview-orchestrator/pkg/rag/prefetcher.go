// Package rag prefetches knowledge-base context for the assessor while the
// participant is still speaking.
package rag

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/redis/go-redis/v9"
)

const (
	prefetchTTL = 30 * time.Second
	// prefetchBudget bounds the background (non-blocking) prefetch, which is
	// allowed to be slower than the cold path because no turn waits on it.
	prefetchBudget = 2 * time.Second
	chunkLimit     = 3
)

// EmbeddingClient defines the interface for generating vector embeddings.
type EmbeddingClient interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorStore defines the interface for querying nearest-neighbor chunks.
type VectorStore interface {
	QueryChunks(ctx context.Context, blueprintID string, vec []float32, limit int) ([]string, error)
}

// CacheStore defines the interface for storing and retrieving prefetched RAG context.
type CacheStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, val string, ttl time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
}

// QueryTiming captures the breakdown between embedding and vector query latencies.
type QueryTiming struct {
	EmbeddingDuration   time.Duration
	VectorQueryDuration time.Duration
	TotalDuration       time.Duration
}

type pgxVectorStore struct {
	pool *pgxpool.Pool
}

func (s *pgxVectorStore) QueryChunks(ctx context.Context, blueprintID string, vec []float32, limit int) ([]string, error) {
	if s.pool == nil {
		return nil, errors.New("pgxpool is nil")
	}
	rows, err := s.pool.Query(ctx, `
		SELECT chunk_text
		FROM document_chunks
		WHERE blueprint_id = $1
		ORDER BY (embedding <=> $2::vector) ASC
		LIMIT $3
	`, blueprintID, pgvector.NewVector(vec), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chunks []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			continue
		}
		chunks = append(chunks, t)
	}
	return chunks, rows.Err()
}

type redisCacheStore struct {
	client *redis.Client
}

func (s *redisCacheStore) Get(ctx context.Context, key string) (string, error) {
	if s.client == nil {
		return "", errors.New("redis client is nil")
	}
	return s.client.Get(ctx, key).Result()
}

func (s *redisCacheStore) Set(ctx context.Context, key string, val string, ttl time.Duration) error {
	if s.client == nil {
		return errors.New("redis client is nil")
	}
	return s.client.Set(ctx, key, val, ttl).Err()
}

func (s *redisCacheStore) Exists(ctx context.Context, key string) (bool, error) {
	if s.client == nil {
		return false, errors.New("redis client is nil")
	}
	n, err := s.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Prefetcher runs pgvector searches ahead of the final transcript and caches
// the formatted result in Redis.
type Prefetcher struct {
	pool        *pgxpool.Pool
	redisClient *redis.Client
	vecStore    VectorStore
	cacheStore  CacheStore
	embClient   EmbeddingClient
	queryBudget time.Duration // hard timeout for cold queries on the live turn
	inflight    sync.Map      // prefetch keys that have been started locally
}

// NewPrefetcher builds a Prefetcher with concrete pgx and redis connections.
func NewPrefetcher(pool *pgxpool.Pool, redisClient *redis.Client,
	embClient EmbeddingClient, queryBudget time.Duration) *Prefetcher {
	if queryBudget <= 0 {
		queryBudget = 150 * time.Millisecond
	}
	var vs VectorStore
	if pool != nil {
		vs = &pgxVectorStore{pool: pool}
	}
	var cs CacheStore
	if redisClient != nil {
		cs = &redisCacheStore{client: redisClient}
	}
	return &Prefetcher{
		pool:        pool,
		redisClient: redisClient,
		vecStore:    vs,
		cacheStore:  cs,
		embClient:   embClient,
		queryBudget: queryBudget,
	}
}

// NewCustomPrefetcher allows injecting custom or in-memory vector/cache stores for replay benchmarking.
func NewCustomPrefetcher(vecStore VectorStore, cacheStore CacheStore,
	embClient EmbeddingClient, queryBudget time.Duration) *Prefetcher {
	if queryBudget <= 0 {
		queryBudget = 150 * time.Millisecond
	}
	return &Prefetcher{
		vecStore:    vecStore,
		cacheStore:  cacheStore,
		embClient:   embClient,
		queryBudget: queryBudget,
	}
}

func PrefetchKey(sessionID string, turnNum int) string {
	return fmt.Sprintf("rag:prefetch:%s:%d", sessionID, turnNum)
}

func prefetchKey(sessionID string, turnNum int) string {
	return PrefetchKey(sessionID, turnNum)
}

func (p *Prefetcher) ready() bool {
	return p != nil && p.vecStore != nil && p.cacheStore != nil && p.embClient != nil
}

// TriggerPrefetch fires a non-blocking async pgvector search on interim transcript text.
// Deduplicated per (sessionID, turnNum) — only the first call fires the query.
func (p *Prefetcher) TriggerPrefetch(ctx context.Context,
	sessionID, blueprintID string, turnNum int, interimText string) {
	if !p.ready() || strings.TrimSpace(interimText) == "" {
		return
	}
	key := prefetchKey(sessionID, turnNum)
	// Interim results arrive many times per second; the local set closes the
	// window before the first query has written its Redis key.
	if _, loaded := p.inflight.LoadOrStore(key, struct{}{}); loaded {
		return
	}
	if exists, _ := p.cacheStore.Exists(ctx, key); exists {
		return
	}
	go func() {
		defer time.AfterFunc(prefetchTTL, func() { p.inflight.Delete(key) })
		pCtx, cancel := context.WithTimeout(context.Background(), prefetchBudget)
		defer cancel()
		result, err := p.hybridSearch(pCtx, blueprintID, interimText)
		if err != nil {
			slog.Debug("rag prefetch failed", "session_id", sessionID, "turn", turnNum, "error", err)
			p.inflight.Delete(key) // allow a later interim result to retry
			return
		}
		if err := p.cacheStore.Set(pCtx, key, result, prefetchTTL); err != nil {
			slog.Debug("rag prefetch cache write failed", "key", key, "error", err)
		}
	}()
}

// GetContext returns the prefetched RAG context.
// If not ready, runs a cold query with p.queryBudget timeout.
// On failure returns ("", nil) — never blocks the turn.
func (p *Prefetcher) GetContext(ctx context.Context,
	sessionID, blueprintID string, turnNum int, finalText string) (string, error) {
	if !p.ready() || strings.TrimSpace(finalText) == "" {
		return "", nil
	}
	key := prefetchKey(sessionID, turnNum)
	if cached, err := p.cacheStore.Get(ctx, key); err == nil && cached != "" {
		return cached, nil
	}
	qCtx, cancel := context.WithTimeout(ctx, p.queryBudget)
	defer cancel()
	result, err := p.hybridSearch(qCtx, blueprintID, finalText)
	if err != nil {
		slog.Debug("rag cold query failed", "session_id", sessionID, "turn", turnNum, "error", err)
		return "", nil
	}
	return result, nil
}

// SearchWithTiming executes embedding and vector query while measuring each stage's duration.
func (p *Prefetcher) SearchWithTiming(ctx context.Context,
	blueprintID, text string) (string, QueryTiming, error) {
	t0 := time.Now()
	vec, err := p.embClient.Embed(ctx, text)
	t1 := time.Now()
	if err != nil {
		return "", QueryTiming{EmbeddingDuration: t1.Sub(t0), TotalDuration: time.Since(t0)}, err
	}
	chunks, err := p.vecStore.QueryChunks(ctx, blueprintID, vec, chunkLimit)
	t2 := time.Now()
	timing := QueryTiming{
		EmbeddingDuration:   t1.Sub(t0),
		VectorQueryDuration: t2.Sub(t1),
		TotalDuration:       t2.Sub(t0),
	}
	if err != nil {
		return "", timing, err
	}
	return formatChunks(chunks), timing, nil
}

func (p *Prefetcher) hybridSearch(ctx context.Context,
	blueprintID, text string) (string, error) {
	result, _, err := p.SearchWithTiming(ctx, blueprintID, text)
	return result, err
}

func FormatChunks(chunks []string) string {
	if len(chunks) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("[Context from assessment knowledge base]\n")
	for i, c := range chunks {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, c))
	}
	return sb.String()
}

func formatChunks(chunks []string) string {
	return FormatChunks(chunks)
}
