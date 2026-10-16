# VoxAssess — Verification & Replay Benchmark Suite (`bench/`)

This directory contains the automated performance verification suites, fair comparison microbenchmarks, and speculative retrieval replay harnesses for VoxAssess.

---

## Benchmark Suites

### 1. Fair Comparison Microbenchmarks (`fair_comparisons_test.go`)
Empirically quantifies architectural treatment effects against conventional sequential voice bot designs under identical constant workloads:

- **Comparison A: Sequential Completion-Then-TTS vs. Pipelined Synthesis**
  - Measures candidate dead-air delay from speech termination to first audible audio chunk playback.
  - **Baseline**: Waits for full LLM generation (45 tokens) before calling TTS $\rightarrow$ $\text{p50} = 753.9\text{ms}$.
  - **Treatment**: Dispatches Sentence 1 to ElevenLabs as soon as punctuation resolves $\rightarrow$ $\text{p50} = 273.7\text{ms}$.
  - **Result**: **2.8x reduction in dead-air delay** (saving $480.2\text{ms}$).

- **Comparison B: Post-Transcript Retrieval vs. Interim Prefetch**
  - Measures retrieval delay blocking LLM prompt assembly after speech finalization.
  - **Baseline**: Synchronous embedding and pgvector query after speech end $\rightarrow$ $\text{p50} = 45.4\text{ms}$.
  - **Treatment**: Speculative prefetch on interim transcript cached in Redis $\rightarrow$ $\text{p50} = 0.8\text{ms}$.
  - **Result**: **59.2x speedup** on critical path.

- **Comparison C: Serial Evaluation vs. Bounded-Concurrent Scatter-Gather ($N=5$)**
  - Measures total execution turnaround to grade a 5-criterion rubric.
  - **Baseline**: Serial sequential evaluation loop $\rightarrow$ $\text{p50} = 502.2\text{ms}$.
  - **Treatment**: Scatter-gather worker pool bounded to $N=5$ $\rightarrow$ $\text{p50} = 100.4\text{ms}$.
  - **Result**: **5.0x throughput improvement**.

---

### 2. Speculative Retrieval Replay Harness (`prefetch_replay_test.go`)
Feeds 30 recorded speech turns (simulated 1:1 timeline playback) through the real `rag.Prefetcher`:
- **Component Latency Split**:
  - Embedding Latency: $\text{p50} = 54.3\text{ms}$, $\text{p95} = 61.5\text{ms}$.
  - pgvector Query Latency: $\text{p50} = 10.5\text{ms}$, $\text{p95} = 10.7\text{ms}$.
  - Total Cold Retrieval: $\text{p50} = 64.8\text{ms}$, $\text{p95} = 72.1\text{ms}$.
- **Hit Rate vs. Quality**:
  - Raw Cache Hit Rate: **90.0%** (27 of 30 turns).
  - Fresh Valid Hits: **70.0%** (21 of 30 turns).
  - Stale / Drifted Context Hits: **22.2% of hits** (6 of 27 hits). Highlights semantic drift under extended candidate speech.
- **Fallback Rate**:
  - Cold Query Fallback: **10.0%** (3 of 30 turns, on short responses $<200\text{ms}$ that beat the background prefetch).
  - Static Rubric Fallback: **0.0%** (cold queries completed within 150ms live budget).
- **Critical Path Latency**:
  - $\text{p50} = 1.15\text{ms}$ (instant Redis cache read).
  - $\text{p99} = 65.22\text{ms}$ (when falling back to cold retrieval).

---

## Running the Benchmarks

```bash
# Run fair comparison microbenchmarks
go test -v -run=TestFairComparison

# Run speculative retrieval replay harness
go test -v -run=TestSpeculativeRetrievalReplayHarness

# Run all project benchmark suites (Grounding, Turn Latency, Adaptive Ablation, Evaluation Reliability)
go run run_benchmarks.go
```
