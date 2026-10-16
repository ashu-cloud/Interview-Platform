# Speculative RAG Retrieval & Prefetch Architecture

This document provides a comprehensive technical audit and architectural trace of the speculative retrieval system implemented in `services/interview-orchestrator/pkg/rag`. It details the end-to-end execution path from interim speech-to-text (STT) hypotheses to cached context injection, reports on corpus reality and time budgets, documents the semantic drift defect under extended speech, and presents measured benchmarks from the replay harness.

---

## 1. End-to-End System Architecture

```mermaid
flowchart TD
    subgraph STTTier["Speech Recognition Tier"]
        Mic["Candidate Audio"] --> STT["Deepgram Nova-2 STT"]
        STT -->|Interim Transcripts<br/>(~150ms intervals)| InterimMsg["websocket.go:494<br/>TriggerPrefetch()"]
        STT -->|Final Transcript<br/>(Speech End / VAD)| FinalMsg["websocket.go:576<br/>GetContext()"]
    end

    subgraph PrefetchTier["Speculative Prefetcher (Non-Blocking Background)"]
        InterimMsg --> Dedup{"inflight.LoadOrStore(key)<br/>Already running?"}
        Dedup -- Yes --> Drop["Drop duplicate interim"]
        Dedup -- No --> CacheCheck{"cacheStore.Exists(key)?"}
        CacheCheck -- Yes --> ReturnEarly["Skip (Cached)"]
        CacheCheck -- No --> SpawnBG["Spawn Goroutine<br/>(Timeout: prefetchBudget = 2.0s)"]

        SpawnBG --> EmbedCall["Embedding API Call<br/>(768-dim unit vector)"]
        EmbedCall --> PgVector["PostgreSQL pgvector Query<br/>ORDER BY (embedding <=> vec) ASC<br/>LIMIT 3"]
        PgVector --> FormatChunks["Format Chunks"]
        FormatChunks --> WriteRedis["Write Redis Cache<br/>rag:prefetch:{session}:{turn}<br/>TTL = 30s"]
    end

    subgraph LiveTurnTier["Live Turn Critical Path (Blocking <150ms)"]
        FinalMsg --> ReadCache["Read Redis Cache<br/>rag:prefetch:{session}:{turn}"]
        ReadCache --> HitCheck{"Cache Hit?"}
        HitCheck -- Yes (<2ms) --> ContextReady["ragContext populated<br/>Inject into prompt"]
        HitCheck -- No / Timeout --> ColdQuery["Cold Query Fallback<br/>(Hard budget: queryBudget = 150ms)"]

        ColdQuery --> ColdSuccess{"Cold Query <= 150ms?"}
        ColdSuccess -- Yes --> ContextReady
        ColdSuccess -- No (Timeout) --> StaticFallback["Fallback to Static Rubric<br/>(ragContext = '')"]

        ContextReady --> PromptBuilder["buildSystemPrompt()<br/>(docket + rubric + ragContext)"]
        StaticFallback --> PromptBuilder
        PromptBuilder --> LLM["Groq Llama 3.3 70B"]
    end
```

---

## 2. In-Depth System Audit Findings

### A. Corpus Contents Reality
- **Implementation Reality**: The `document_chunks` table does **not** store external system design papers, engineering manuals, architecture wikis, or candidate CVs.
- **Data Source**: As defined in `services/job-candidate-service/internal/service/docket.go`, chunks are derived strictly from:
  1. `assessment_prompt`: The blueprint's own interview questions (`p.QuestionText`).
  2. `evaluation_criterion`: The blueprint's rubric criteria names and descriptions (`c.Name + ": " + c.Description`).
- **Scale**: A typical blueprint contains 4–8 questions and 3–6 criteria (total **7 to 14 rows** in PostgreSQL).
- **Index Type**: [`migrations/000003_foundation.up.sql`](file:///c:/Users/Ashu/Desktop/work/web%20dev/web%20d/projects/Ai%20Interview%20Platform/migrations/000003_foundation.up.sql) contains an index on `blueprint_id` only. No HNSW or IVFFlat vector index exists on `embedding`. Postgres performs an exact sequential scan over the ~10 rows belonging to the given `blueprint_id`.

### B. Vector-Only vs. Hybrid Mechanics
- Although the Go method is named `hybridSearch`, the SQL implementation in `prefetcher.go` is:
  ```sql
  SELECT chunk_text
  FROM document_chunks
  WHERE blueprint_id = $1
  ORDER BY (embedding <=> $2::vector) ASC
  LIMIT 3;
  ```
- **Finding**: Retrieval is **100% vector-only cosine distance** (`<=>`).
- There is **no BM25**, no PostgreSQL full-text search (`tsvector` / `tsquery`), no lexical inverted index, and no Reciprocal Rank Fusion (RRF).

### C. Time Budgets vs. Embedding Latency
- **Background Prefetch Budget**: `2 * time.Second` (2,000ms).
  - Background embedding (~45–65ms) + pgvector query (~10ms) easily completes within 2,000ms.
- **Cold Query Live Budget**: `150 * time.Millisecond` (150ms).
  - If a turn experiences a cache miss, it must run the embedding API call and database query sequentially on the live turn path.
  - While p50 total retrieval latency is ~64.8ms, real-world public cloud network jitter (OpenAI embedding API p95 > 120ms) can easily cause the 150ms context deadline to expire, aborting retrieval and triggering empty context fallback.

### D. Semantic Drift & Stale Context Defect
A major design defect in eager prefetching is **context staleness under extended speech**:
1. `TriggerPrefetch` deduplicates using `inflight.LoadOrStore(key, struct{}{})`.
2. The prefetch fires **only once** on the candidate's very first interim transcript event (typically at 120ms–200ms into speech, containing 2–4 words like *"Sure, let's talk..."* or *"Um, well..."*).
3. All subsequent interim transcripts emitted over the next 3 to 10 seconds are dropped by the `inflight` check.
4. If the candidate pivots topics mid-turn or elaborates on a nuanced architecture (*"...actually, we deprecated REST and switched to gRPC with Protobuf because JSON parsing overhead throttled our gateway"*), the Redis cache holds chunks retrieved from the opening words.
5. At the final transcript, `GetContext` technically records a cache hit, but injects **semantically stale context**.

---

## 3. Replay Harness Benchmark Measurements (`bench/prefetch_replay_test.go`)

We executed an end-to-end replay harness using 30 recorded speech turns (simulated 1:1 timeline playback) with calibrated embedding (p50 ~54ms) and pgvector query latencies (p50 ~10ms):

| Metric | Measured Value | Analysis |
| :--- | :--- | :--- |
| **Embedding Latency (isolated)** | **p50 = 54.3ms** / **p95 = 61.5ms** | Accounts for 83.8% of total cold retrieval time |
| **pgvector Query Latency (isolated)** | **p50 = 10.5ms** / **p95 = 10.7ms** | Accounts for 16.2% of total cold retrieval time |
| **Total Cold Retrieval Latency** | **p50 = 64.8ms** / **p95 = 72.1ms** | Required when prefetch misses or fails |
| **Raw Prefetch Cache Hit Rate** | **90.0%** (27 / 30 turns) | High nominal hit rate across speaking turns |
| **Fresh Valid Hits** | **70.0%** (21 / 30 turns) | Hits where prefetched context matched final intent |
| **Stale / Drifted Context Hits** | **22.2% of hits** (6 / 27 hits) | **Defect**: Hits where interim prefix diverged from final speech |
| **Cold Query Fallback Rate** | **10.0%** (3 / 30 turns) | Occurred on short turns (<200ms) that beat background prefetch |
| **Static Rubric Fallback Rate** | **0.0%** (0 / 30 turns) | Cold query fit inside 150ms budget under calibrated conditions |
| **Critical Path Retrieval Latency** | **p50 = 1.15ms** / **p99 = 65.22ms** | **56.3x speedup** on cache hit vs cold query |

---

## 4. Architectural Mitigation & Improvements

To resolve the 22.2% semantic staleness rate and eliminate cold query race conditions on short turns, the following improvements are specified:

```go
// 1. Debounce & Word-Count Stability Gate
// Wait until candidate has spoken >= 5 words or >= 400ms before firing prefetch
const minPrefetchWords = 5
const minSpeechDuration = 400 * time.Millisecond

func shouldTriggerPrefetch(text string, duration time.Duration) bool {
    return len(strings.Fields(text)) >= minPrefetchWords && duration >= minSpeechDuration
}

// 2. Semantic Drift Re-Prefetching
// If subsequent interim embedding drifts > 0.35 cosine distance, invalidate cache & re-fetch
func (p *Prefetcher) InvalidateIfDrifted(ctx context.Context, sessionID string, turnNum int, newText string, initialVec []float32) {
    newVec, _ := p.embClient.Embed(ctx, newText)
    if cosineDistance(initialVec, newVec) > 0.35 {
        p.cacheStore.Delete(ctx, prefetchKey(sessionID, turnNum))
        go p.executePrefetch(ctx, sessionID, turnNum, newText)
    }
}

// 3. Singleflight In-Flight Await
// If speech ends while prefetch is in-flight (t < 65ms), await the running channel
// rather than issuing an immediate duplicate cold query.
```
