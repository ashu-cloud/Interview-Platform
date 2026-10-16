package rag

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/embedding"
	pkgrag "github.com/yourusername/ai-interview/services/interview-orchestrator/pkg/rag"
)

type Prefetcher = pkgrag.Prefetcher
type VectorStore = pkgrag.VectorStore
type CacheStore = pkgrag.CacheStore
type QueryTiming = pkgrag.QueryTiming

func NewPrefetcher(pool *pgxpool.Pool, redisClient *redis.Client,
	embClient embedding.EmbeddingClient, queryBudget time.Duration) *Prefetcher {
	return pkgrag.NewPrefetcher(pool, redisClient, embClient, queryBudget)
}

func NewCustomPrefetcher(vecStore VectorStore, cacheStore CacheStore,
	embClient embedding.EmbeddingClient, queryBudget time.Duration) *Prefetcher {
	return pkgrag.NewCustomPrefetcher(vecStore, cacheStore, embClient, queryBudget)
}

var prefetchKey = pkgrag.PrefetchKey
var formatChunks = pkgrag.FormatChunks
