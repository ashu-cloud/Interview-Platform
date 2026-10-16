package handler

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkSentenceSegmentation measures findSentenceEnd across realistic AI token streams.
func BenchmarkSentenceSegmentation(b *testing.B) {
	cases := []struct {
		name string
		text string
	}{
		{"ShortSingleClause", "We implemented distributed locking using Redis."},
		{"DecimalVersionNumbers", "In version 2.1 of our microservice, we achieved a 3.14x speedup."},
		{"UnpunctuatedLongBuffer", "This is a very long stream of tokens coming from the LLM that does not contain any punctuation mark until more than one hundred characters later where it finally breaks"},
		{"MultipleSentencesWithNewlines", "First sentence ends here.\nSecond sentence begins now and continues further."},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = findSentenceEnd(tc.text)
			}
		})
	}
}

// TestOrchestratorPipeliningAndInterruption measures orchestrator overhead and cancellation speed.
func TestOrchestratorPipeliningAndInterruption(t *testing.T) {
	// 1. Measure Sentence Chunking Dispatch Overhead (excluding provider latency)
	t.Run("SentenceChunkingOverhead", func(t *testing.T) {
		tokens := []string{
			"We ", "designed ", "a ", "distributed ", "locking ", "mechanism. ",
			"It ", "uses ", "Redis ", "SETNX ", "with ", "a ", "two ", "second ", "TTL. ",
			"This ", "prevents ", "any ", "concurrent ", "worker ", "double-booking. ",
		}

		start := time.Now()
		var fullText strings.Builder
		var sentenceBuf strings.Builder
		var extractedSentences []string

		for _, chunk := range tokens {
			fullText.WriteString(chunk)
			sentenceBuf.WriteString(chunk)

			for {
				bufStr := sentenceBuf.String()
				endIdx := findSentenceEnd(bufStr)
				if endIdx < 0 {
					break
				}
				sentence := strings.TrimSpace(bufStr[:endIdx+1])
				remaining := strings.TrimLeft(bufStr[endIdx+1:], " \t\r\n")
				sentenceBuf.Reset()
				sentenceBuf.WriteString(remaining)

				if sentence != "" {
					extractedSentences = append(extractedSentences, sentence)
				}
			}
		}
		duration := time.Since(start)

		if len(extractedSentences) != 3 {
			t.Fatalf("expected 3 sentences, got %d: %v", len(extractedSentences), extractedSentences)
		}
		t.Logf("Sentence chunking overhead for 3 sentences (%d tokens): %v (%.2f µs per sentence)",
			len(tokens), duration, float64(duration.Microseconds())/3.0)
	})

	// 2. Measure Barge-In Cancellation Propagation Time
	t.Run("BargeInCancellationLatency", func(t *testing.T) {
		tc := &turnController{}
		ctx, cancel := tc.startTurn()

		var chunksEmitted atomic.Int64
		var postCancelEmitted atomic.Int64
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(2 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					chunksEmitted.Add(1)
					if !tc.isSpeaking() {
						postCancelEmitted.Add(1)
					}
				}
			}
		}()

		// Let it stream for 20ms
		time.Sleep(20 * time.Millisecond)

		// Server cancellation request -> task exit
		cancelStart := time.Now()
		cancel()
		tc.interrupt()
		wg.Wait()
		taskExitDuration := time.Since(cancelStart)

		t.Logf("Server cancellation request -> task exit duration: %v (total emitted: %d, post-cancel: %d)",
			taskExitDuration, chunksEmitted.Load(), postCancelEmitted.Load())

		if taskExitDuration > 10*time.Millisecond {
			t.Errorf("task exit took too long: %v", taskExitDuration)
		}
	})
}
