// Package rag prefetches knowledge-base context for the assessor while the
// participant is still speaking.
package rag

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/redis/go-redis/v9"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/embedding"
)

const (
	prefetchTTL = 30 * time.Second
	// prefetchBudget bounds the background (non-blocking) prefetch, which is
	// allowed to be slower than the cold path because no turn waits on it.
	prefetchBudget = 2 * time.Second
	chunkLimit     = 3
)

// Prefetcher runs pgvector searches ahead of the final transcript and caches
// the formatted result in Redis.
type Prefetcher struct {
	pool        *pgxpool.Pool
	redisClient *redis.Client
	embClient   embedding.EmbeddingClient
	queryBudget time.Duration // hard timeout for cold queries on the live turn
	inflight    sync.Map      // prefetch keys that have been started locally
}

// NewPrefetcher builds a Prefetcher. queryBudget bounds cold queries.
func NewPrefetcher(pool *pgxpool.Pool, redisClient *redis.Client,
	embClient embedding.EmbeddingClient, queryBudget time.Duration) *Prefetcher {
	if queryBudget <= 0 {
		queryBudget = 150 * time.Millisecond
	}
	return &Prefetcher{pool: pool, redisClient: redisClient, embClient: embClient, queryBudget: queryBudget}
}

func prefetchKey(sessionID string, turnNum int) string {
	return fmt.Sprintf("rag:prefetch:%s:%d", sessionID, turnNum)
}

func (p *Prefetcher) ready() bool {
	return p != nil && p.pool != nil && p.redisClient != nil && p.embClient != nil
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
	if exists, _ := p.redisClient.Exists(ctx, key).Result(); exists > 0 {
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
		if err := p.redisClient.Set(pCtx, key, result, prefetchTTL).Err(); err != nil {
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
	if cached, err := p.redisClient.Get(ctx, key).Result(); err == nil {
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

func (p *Prefetcher) hybridSearch(ctx context.Context,
	blueprintID, text string) (string, error) {
	vec, err := p.embClient.Embed(ctx, text)
	if err != nil {
		return "", err
	}
	rows, err := p.pool.Query(ctx, `
		SELECT chunk_text
		FROM document_chunks
		WHERE blueprint_id = $1
		ORDER BY (embedding <=> $2::vector) ASC
		LIMIT $3
	`, blueprintID, pgvector.NewVector(vec), chunkLimit)
	if err != nil {
		return "", err
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
	if err := rows.Err(); err != nil {
		return "", err
	}
	return formatChunks(chunks), nil
}

func formatChunks(chunks []string) string {
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
