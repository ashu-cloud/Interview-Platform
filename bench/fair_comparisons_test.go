package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// RunStats holds sample size, success/failure counts, and latency percentiles.
type RunStats struct {
	Name         string
	SampleCount  int
	SuccessCount int
	FailureCount int
	TimeoutCount int
	P50          time.Duration
	P90          time.Duration
	P99          time.Duration
	Min          time.Duration
	Max          time.Duration
	Durations    []time.Duration
}

func computeStats(name string, durations []time.Duration, failures, timeouts int) RunStats {
	n := len(durations)
	if n == 0 {
		return RunStats{Name: name, FailureCount: failures, TimeoutCount: timeouts}
	}
	sorted := make([]time.Duration, n)
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	p50Idx := int(math.Round(float64(n-1) * 0.50))
	p90Idx := int(math.Round(float64(n-1) * 0.90))
	p99Idx := int(math.Round(float64(n-1) * 0.99))

	return RunStats{
		Name:         name,
		SampleCount:  n + failures + timeouts,
		SuccessCount: n,
		FailureCount: failures,
		TimeoutCount: timeouts,
		P50:          sorted[p50Idx],
		P90:          sorted[p90Idx],
		P99:          sorted[p99Idx],
		Min:          sorted[0],
		Max:          sorted[n-1],
		Durations:    sorted,
	}
}

// ============================================================================
// COMPARISON A: Sequential Completion-then-TTS vs Pipelined Synthesis
// Constant Workload: 3 sentences, 45 tokens. TTFT=25ms, TPOT=15ms. TTS=80ms/sentence (or 200ms whole text).
// ============================================================================
func TestFairComparison_SequentialVsPipelinedTTS(t *testing.T) {
	const sampleCount = 20
	const (
		numSentences   = 3
		tokensPerSent  = 15
		ttft           = 25 * time.Millisecond
		tpot           = 12 * time.Millisecond
		ttsPerSentence = 80 * time.Millisecond
		ttsWholeText   = 200 * time.Millisecond
	)

	// Baseline: Sequential Completion-then-TTS (waits for full LLM generation before invoking TTS)
	var baselineDurations []time.Duration
	for i := 0; i < sampleCount; i++ {
		start := time.Now()

		// 1. Full LLM completion
		time.Sleep(ttft + time.Duration(numSentences*tokensPerSent-1)*tpot)

		// 2. TTS on complete block of text
		time.Sleep(ttsWholeText)

		baselineDurations = append(baselineDurations, time.Since(start))
	}
	baselineStats := computeStats("Baseline: Sequential Completion-Then-TTS", baselineDurations, 0, 0)

	// Treatment: Pipelined Sentence Synthesis (dispatches first sentence to TTS as soon as sentence 1 punctuates)
	var treatmentDurations []time.Duration
	for i := 0; i < sampleCount; i++ {
		start := time.Now()

		// 1. LLM streams up to first sentence boundary (tokensPerSent)
		time.Sleep(ttft + time.Duration(tokensPerSent-1)*tpot)

		// 2. Concurrent TTS on sentence 1 begins immediately while subsequent tokens continue streaming
		time.Sleep(ttsPerSentence)

		treatmentDurations = append(treatmentDurations, time.Since(start))
	}
	treatmentStats := computeStats("Treatment: Pipelined Sentence Synthesis", treatmentDurations, 0, 0)

	printComparisonReport(
		"COMPARISON A: SEQUENTIAL COMPLETION-THEN-TTS vs PIPELINED SYNTHESIS",
		"Measures candidate dead-air delay from speech end to first audible playback.",
		baselineStats,
		treatmentStats,
		"Mock-provider deterministic microbenchmark (constant tokens=45, ttft=25ms, tpot=12ms, tts=80ms).",
		"Live provider network jitter excluded; measures architectural pipelining delta.",
	)
}

// ============================================================================
// COMPARISON B: Retrieval After Final Transcription vs Interim Prefetch
// Constant Workload: Embedding=30ms, pgvector lookup=15ms. STT stabilization window=250ms.
// ============================================================================
func TestFairComparison_PostTranscriptRetrievalVsInterimPrefetch(t *testing.T) {
	const sampleCount = 20
	const (
		embeddingLatency = 30 * time.Millisecond
		pgvectorLatency  = 15 * time.Millisecond
		cacheHitLookup   = 100 * time.Microsecond
	)

	// Baseline: Retrieval After Final Transcription
	// Final STT transcript completes at t=0; candidate waits for embedding + vector search before prompt assembly
	var baselineDurations []time.Duration
	for i := 0; i < sampleCount; i++ {
		start := time.Now()
		// Compute embedding + vector query after final transcript
		time.Sleep(embeddingLatency + pgvectorLatency)
		baselineDurations = append(baselineDurations, time.Since(start))
	}
	baselineStats := computeStats("Baseline: Post-Transcript Retrieval", baselineDurations, 0, 0)

	// Treatment: Interim Prefetch
	// Triggered on interim STT transcript 250ms prior to finalization; cache is warm when final STT arrives
	var treatmentDurations []time.Duration
	for i := 0; i < sampleCount; i++ {
		start := time.Now()
		// By the time final STT arrives, cache lookup takes <1ms
		time.Sleep(cacheHitLookup)
		treatmentDurations = append(treatmentDurations, time.Since(start))
	}
	treatmentStats := computeStats("Treatment: Interim Transcript Prefetch", treatmentDurations, 0, 0)

	printComparisonReport(
		"COMPARISON B: RETRIEVAL AFTER FINAL TRANSCRIPTION vs INTERIM PREFETCH",
		"Measures retrieval latency overhead blocking LLM prompt assembly after speech finalization.",
		baselineStats,
		treatmentStats,
		"Mock vector store & embedding client (constant embedding=30ms, pgvector=15ms).",
		"Assumes interim transcript semantic stability; false-trigger cache invalidation excluded.",
	)
}

// ============================================================================
// COMPARISON C: Serial vs Bounded-Concurrent Assessment
// Constant Workload: 5 evaluation criteria, each taking 100ms LLM evaluation time. Concurrency=5.
// ============================================================================
func TestFairComparison_SerialVsConcurrentEvaluation(t *testing.T) {
	const sampleCount = 15
	const (
		criteriaCount   = 5
		evalLatencyCrit = 100 * time.Millisecond
		maxConcurrent   = 5
	)

	// Baseline: Serial Assessment (evaluates 1 criterion after another)
	var baselineDurations []time.Duration
	for i := 0; i < sampleCount; i++ {
		start := time.Now()
		for c := 0; c < criteriaCount; c++ {
			time.Sleep(evalLatencyCrit)
		}
		baselineDurations = append(baselineDurations, time.Since(start))
	}
	baselineStats := computeStats("Baseline: Serial Evaluation (1 at a time)", baselineDurations, 0, 0)

	// Treatment: Bounded-Concurrent Assessment (scatter-gather worker with sem=5)
	var treatmentDurations []time.Duration
	for i := 0; i < sampleCount; i++ {
		start := time.Now()
		sem := make(chan struct{}, maxConcurrent)
		var wg sync.WaitGroup
		for c := 0; c < criteriaCount; c++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				time.Sleep(evalLatencyCrit)
			}()
		}
		wg.Wait()
		treatmentDurations = append(treatmentDurations, time.Since(start))
	}
	treatmentStats := computeStats("Treatment: Bounded-Concurrent Scatter-Gather (N=5)", treatmentDurations, 0, 0)

	printComparisonReport(
		"COMPARISON C: SERIAL vs BOUNDED-CONCURRENT ASSESSMENT",
		"Measures total worker execution turnaround to grade a 5-criterion blueprint.",
		baselineStats,
		treatmentStats,
		"Deterministic mock evaluator (5 criteria, 100ms evaluation latency per criterion).",
		"Provider rate limit backoff (TPM/RPM throttling) not triggered under mock provider.",
	)
}

func printComparisonReport(title, description string, base, treat RunStats, env, limitations string) {
	speedup := float64(base.P50) / float64(treat.P50)
	latencySaved := base.P50 - treat.P50

	fmt.Println("\n=========================================================================================")
	fmt.Printf("   %s\n", title)
	fmt.Println("=========================================================================================")
	fmt.Printf("Description: %s\n", description)
	fmt.Printf("Environment: %s\n", env)
	fmt.Printf("Limitations: %s\n", limitations)
	fmt.Println(strings.Repeat("-", 89))
	fmt.Printf("%-44s | %-7s | %-9s | %-9s | %-9s | %-9s\n",
		"Configuration", "Samples", "p50 (Med)", "p90", "p99", "Status")
	fmt.Println(strings.Repeat("-", 89))

	fmt.Printf("%-44s | %7d | %7.1fms | %7.1fms | %7.1fms | %d ok / %d err\n",
		base.Name, base.SampleCount,
		float64(base.P50.Microseconds())/1000.0,
		float64(base.P90.Microseconds())/1000.0,
		float64(base.P99.Microseconds())/1000.0,
		base.SuccessCount, base.FailureCount)

	fmt.Printf("%-44s | %7d | %7.1fms | %7.1fms | %7.1fms | %d ok / %d err\n",
		treat.Name, treat.SampleCount,
		float64(treat.P50.Microseconds())/1000.0,
		float64(treat.P90.Microseconds())/1000.0,
		float64(treat.P99.Microseconds())/1000.0,
		treat.SuccessCount, treat.FailureCount)

	fmt.Println(strings.Repeat("-", 89))
	fmt.Printf("COMPARATIVE FINDING: P50 latency reduced from %.1fms to %.1fms (saved %.1fms, %.1fx speedup)\n",
		float64(base.P50.Microseconds())/1000.0,
		float64(treat.P50.Microseconds())/1000.0,
		float64(latencySaved.Microseconds())/1000.0,
		speedup)
	fmt.Println("=========================================================================================")
}
