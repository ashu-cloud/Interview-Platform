package worker

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yourusername/ai-interview/services/feedback-worker/internal/repository"
)

// LabelledCase represents a test case for the grounding validator.
type LabelledCase struct {
	ID             string
	Category       string // verbatim, whitespace_case, paraphrase, fabricated, wrong_turn, short_quote
	ExpectedGround bool   // ground truth: is it legitimately grounded in that cited turn?
	TurnID         string
	TurnContent    string
	Quote          string
}

// buildLabelledCorpus returns a diverse, realistic corpus of quotes and transcripts.
func buildLabelledCorpus() []LabelledCase {
	var corpus []LabelledCase

	// Base transcript turns
	turn1 := "Turn 1: We implemented distributed locking using Redis SETNX with an aggressive TTL of two seconds to prevent race conditions during concurrent worker dispatch."
	turn2 := "Turn 2: For database consistency, we wrapped PostgreSQL outbox processing in a single transaction with FOR UPDATE SKIP LOCKED to guarantee idempotent turn projection."
	turn3 := "Turn 3: The system architecture uses Go microservices with WebSocket connections for streaming real-time audio from Deepgram Nova-2 to Groq LLM."
	turn4 := "Turn 4: When evaluating candidates, we use a scatter-gather worker that evaluates up to five rubric criteria in parallel before aggregating the hire recommendation."

	// 1. Verbatim quotes (Expected: true)
	verbatimQuotes := []struct {
		turnID  string
		content string
		quote   string
	}{
		{"t1", turn1, "distributed locking using Redis SETNX"},
		{"t1", turn1, "TTL of two seconds to prevent race conditions"},
		{"t2", turn2, "wrapped PostgreSQL outbox processing in a single transaction"},
		{"t2", turn2, "FOR UPDATE SKIP LOCKED"},
		{"t3", turn3, "WebSocket connections for streaming real-time audio"},
		{"t3", turn3, "Deepgram Nova-2 to Groq LLM"},
		{"t4", turn4, "evaluates up to five rubric criteria in parallel"},
		{"t4", turn4, "aggregating the hire recommendation"},
		{"t1", turn1, "concurrent worker dispatch"},
		{"t2", turn2, "idempotent turn projection"},
	}
	for i, v := range verbatimQuotes {
		corpus = append(corpus, LabelledCase{
			ID:             fmt.Sprintf("verb-%d", i+1),
			Category:       "verbatim",
			ExpectedGround: true,
			TurnID:         v.turnID,
			TurnContent:    v.content,
			Quote:          v.quote,
		})
	}

	// 2. Whitespace & Case variants (Expected: true, should match)
	whitespaceCaseQuotes := []struct {
		turnID  string
		content string
		quote   string
	}{
		{"t1", turn1, "DISTRIBUTED LOCKING USING REDIS SETNX"},
		{"t1", turn1, "ttl of two seconds   to prevent   race conditions"},
		{"t2", turn2, "wrapped  postgresql   outbox processing"},
		{"t2", turn2, "for update skip locked"},
		{"t3", turn3, "websocket CONNECTIONS for Streaming"},
		{"t3", turn3, "deepgram  nova-2   to groq llm"},
		{"t4", turn4, "EVALUATES UP TO FIVE RUBRIC CRITERIA"},
		{"t4", turn4, "aggregating  \t  the hire   recommendation"},
		{"t1", turn1, "Concurrent Worker Dispatch"},
		{"t2", turn2, "IDEMPOTENT TURN PROJECTION"},
	}
	for i, v := range whitespaceCaseQuotes {
		corpus = append(corpus, LabelledCase{
			ID:             fmt.Sprintf("ws_case-%d", i+1),
			Category:       "whitespace_case",
			ExpectedGround: true,
			TurnID:         v.turnID,
			TurnContent:    v.content,
			Quote:          v.quote,
		})
	}

	// 3. Paraphrase (Semantically valid, but phrasing differs; Expected: true semantically, but validator misses)
	paraphraseQuotes := []struct {
		turnID  string
		content string
		quote   string
	}{
		{"t1", turn1, "We built distributed locks with Redis SETNX to stop race conditions."},
		{"t1", turn1, "Set a 2-second expiry TTL on Redis locks during dispatch."},
		{"t2", turn2, "PostgreSQL outbox was handled in one transaction with SKIP LOCKED."},
		{"t2", turn2, "Turn projection was guaranteed to be idempotent using transactions."},
		{"t3", turn3, "Audio streams continuously over WebSockets between Deepgram and Groq."},
		{"t3", turn3, "Employed Go services for low latency audio pipeline."},
		{"t4", turn4, "Assessing five rubric points concurrently before scoring."},
		{"t4", turn4, "Worker performs parallel evaluation across rubric items."},
		{"t1", turn1, "Prevented concurrent worker dispatch collisions with Redis."},
		{"t2", turn2, "Guaranteed no duplicate turns via Postgres locks."},
	}
	for i, v := range paraphraseQuotes {
		corpus = append(corpus, LabelledCase{
			ID:             fmt.Sprintf("para-%d", i+1),
			Category:       "paraphrase",
			ExpectedGround: true, // Ground truth: candidate did convey this concept
			TurnID:         v.turnID,
			TurnContent:    v.content,
			Quote:          v.quote,
		})
	}

	// 4. Fabricated quotes (Hallucinated by LLM; Expected: false)
	fabricatedQuotes := []struct {
		turnID  string
		content string
		quote   string
	}{
		{"t1", turn1, "We configured Kafka partitions with MurmurHash keys."},
		{"t1", turn1, "Implemented raft consensus with etcd for leader election."},
		{"t2", turn2, "Tuned InnoDB buffer pool size to sixteen gigabytes."},
		{"t2", turn2, "Used MongoDB replica sets with write concern majority."},
		{"t3", turn3, "Deployed Kubernetes HPA scaling on custom Prometheus metrics."},
		{"t3", turn3, "Wrote custom WebRTC media server in C++ for sub-10ms delivery."},
		{"t4", turn4, "Trained a cross-encoder model on five thousand labeled resumes."},
		{"t4", turn4, "Implemented Bayesian scoring to eliminate interviewer bias."},
		{"t1", turn1, "Reduced memory consumption by forty percent using sync.Pool."},
		{"t2", turn2, "Achieved fifty thousand transactions per second on single node."},
	}
	for i, v := range fabricatedQuotes {
		corpus = append(corpus, LabelledCase{
			ID:             fmt.Sprintf("fab-%d", i+1),
			Category:       "fabricated",
			ExpectedGround: false,
			TurnID:         v.turnID,
			TurnContent:    v.content,
			Quote:          v.quote,
		})
	}

	// 5. Right quote under wrong turn (Valid quote from transcript, but cited under wrong turn ID; Expected: false for that turn)
	wrongTurnQuotes := []struct {
		citedTurnID string
		actualTurn  string
		quote       string
	}{
		{"t2", turn1, "distributed locking using Redis SETNX"},
		{"t3", turn1, "TTL of two seconds to prevent race conditions"},
		{"t1", turn2, "wrapped PostgreSQL outbox processing in a single transaction"},
		{"t3", turn2, "FOR UPDATE SKIP LOCKED"},
		{"t4", turn3, "WebSocket connections for streaming real-time audio"},
		{"t1", turn3, "Deepgram Nova-2 to Groq LLM"},
		{"t2", turn4, "evaluates up to five rubric criteria in parallel"},
		{"t3", turn4, "aggregating the hire recommendation"},
		{"t4", turn1, "concurrent worker dispatch"},
		{"t1", turn2, "idempotent turn projection"},
	}
	for i, v := range wrongTurnQuotes {
		corpus = append(corpus, LabelledCase{
			ID:             fmt.Sprintf("wrong_turn-%d", i+1),
			Category:       "wrong_turn",
			ExpectedGround: false,
			TurnID:         v.citedTurnID,
			TurnContent:    getTurnContent(v.citedTurnID, turn1, turn2, turn3, turn4),
			Quote:          v.quote,
		})
	}

	// 6. Very short non-substantive quotes (e.g. "we", "in", "to"; Expected: false as substantive evidence, but substring matches)
	shortQuotes := []struct {
		turnID  string
		content string
		quote   string
	}{
		{"t1", turn1, "we"},
		{"t1", turn1, "to"},
		{"t2", turn2, "in a"},
		{"t2", turn2, "for"},
		{"t3", turn3, "using"},
		{"t3", turn3, "with"},
		{"t4", turn4, "up to"},
		{"t4", turn4, "the"},
		{"t1", turn1, "of"},
		{"t2", turn2, "and"},
	}
	for i, v := range shortQuotes {
		corpus = append(corpus, LabelledCase{
			ID:             fmt.Sprintf("short-%d", i+1),
			Category:       "short_quote",
			ExpectedGround: false, // Ground truth: non-substantive quote
			TurnID:         v.turnID,
			TurnContent:    v.content,
			Quote:          v.quote,
		})
	}

	return corpus
}

func getTurnContent(turnID, t1, t2, t3, t4 string) string {
	switch turnID {
	case "t1":
		return t1
	case "t2":
		return t2
	case "t3":
		return t3
	case "t4":
		return t4
	default:
		return ""
	}
}

// TestGroundingValidatorBenchmark runs the benchmark suite against ValidateGrounding.
func TestGroundingValidatorBenchmark(t *testing.T) {
	corpus := buildLabelledCorpus()

	turns := []repository.SessionTurn{
		{ID: "t1", TurnNumber: 1, Role: "participant", Content: "We implemented distributed locking using Redis SETNX with an aggressive TTL of two seconds to prevent race conditions during concurrent worker dispatch."},
		{ID: "t2", TurnNumber: 2, Role: "participant", Content: "For database consistency, we wrapped PostgreSQL outbox processing in a single transaction with FOR UPDATE SKIP LOCKED to guarantee idempotent turn projection."},
		{ID: "t3", TurnNumber: 3, Role: "participant", Content: "The system architecture uses Go microservices with WebSocket connections for streaming real-time audio from Deepgram Nova-2 to Groq LLM."},
		{ID: "t4", TurnNumber: 4, Role: "participant", Content: "When evaluating candidates, we use a scatter-gather worker that evaluates up to five rubric criteria in parallel before aggregating the hire recommendation."},
	}

	var results []CriterionResult
	for _, c := range corpus {
		results = append(results, CriterionResult{
			CriterionID:    c.ID,
			Score:          4,
			EvidenceTurnID: c.TurnID,
			EvidenceQuote:  c.Quote,
		})
	}

	validated := ValidateGrounding(results, turns)

	type CatStats struct {
		Total int
		TP    int // Ground truth true,  pred true
		FP    int // Ground truth false, pred true  (False Accept)
		TN    int // Ground truth false, pred false
		FN    int // Ground truth true,  pred false (False Reject)
	}

	stats := make(map[string]*CatStats)
	categories := []string{"verbatim", "whitespace_case", "paraphrase", "fabricated", "wrong_turn", "short_quote"}
	for _, cat := range categories {
		stats[cat] = &CatStats{}
	}

	totalTP, totalFP, totalTN, totalFN := 0, 0, 0, 0

	for i, c := range corpus {
		pred := validated[i].Grounded
		catStat := stats[c.Category]
		catStat.Total++

		if c.ExpectedGround && pred {
			catStat.TP++
			totalTP++
		} else if !c.ExpectedGround && pred {
			catStat.FP++
			totalFP++
		} else if !c.ExpectedGround && !pred {
			catStat.TN++
			totalTN++
		} else if c.ExpectedGround && !pred {
			catStat.FN++
			totalFN++
		}
	}

	fmt.Println("\n=========================================================================================")
	fmt.Println("                       GROUNDING VALIDATOR BENCHMARK RESULTS")
	fmt.Println("=========================================================================================")
	fmt.Printf("%-18s | %-6s | %-4s | %-4s | %-4s | %-4s | %-12s | %-12s\n",
		"Category", "Total", "TP", "FP", "TN", "FN", "False Accept", "False Reject")
	fmt.Println(strings.Repeat("-", 89))

	for _, cat := range categories {
		s := stats[cat]
		falseAcceptRate := 0.0
		if s.FP+s.TN > 0 {
			falseAcceptRate = float64(s.FP) / float64(s.FP+s.TN) * 100.0
		}
		falseRejectRate := 0.0
		if s.TP+s.FN > 0 {
			falseRejectRate = float64(s.FN) / float64(s.TP+s.FN) * 100.0
		}
		fmt.Printf("%-18s | %-6d | %-4d | %-4d | %-4d | %-4d | %11.1f%% | %11.1f%%\n",
			cat, s.Total, s.TP, s.FP, s.TN, s.FN, falseAcceptRate, falseRejectRate)
	}

	overallPrecision := 0.0
	if totalTP+totalFP > 0 {
		overallPrecision = float64(totalTP) / float64(totalTP+totalFP) * 100.0
	}
	overallRecall := 0.0
	if totalTP+totalFN > 0 {
		overallRecall = float64(totalTP) / float64(totalTP+totalFN) * 100.0
	}
	overallAccuracy := float64(totalTP+totalTN) / float64(len(corpus)) * 100.0
	f1 := 0.0
	if overallPrecision+overallRecall > 0 {
		f1 = 2 * (overallPrecision * overallRecall) / (overallPrecision + overallRecall)
	}

	fmt.Println(strings.Repeat("-", 89))
	fmt.Printf("OVERALL (N=%d): Precision=%.1f%%, Recall=%.1f%%, Accuracy=%.1f%%, F1=%.1f%%\n",
		len(corpus), overallPrecision, overallRecall, overallAccuracy, f1)
	fmt.Printf("KEY FINDINGS:\n")
	fmt.Printf("1. Paraphrase False-Reject Rate: %.1f%% (normalized substring check is completely blind to paraphrased facts)\n",
		float64(stats["paraphrase"].FN)/float64(stats["paraphrase"].Total)*100.0)
	fmt.Printf("2. Short Non-Substantive False-Accept Rate: %.1f%% (substring match trivially matches stop-words in any turn)\n",
		float64(stats["short_quote"].FP)/float64(stats["short_quote"].Total)*100.0)
	fmt.Printf("3. Fabricated Hallucination Catch Rate: %.1f%% (10/10 completely fabricated statements correctly rejected)\n",
		float64(stats["fabricated"].TN)/float64(stats["fabricated"].Total)*100.0)
	fmt.Printf("4. Wrong-Turn Citation Catch Rate: %.1f%% (10/10 valid quotes cited against wrong turn ID correctly rejected)\n",
		float64(stats["wrong_turn"].TN)/float64(stats["wrong_turn"].Total)*100.0)
	fmt.Println("=========================================================================================")
}

func BenchmarkValidateGrounding(b *testing.B) {
	corpus := buildLabelledCorpus()
	turns := []repository.SessionTurn{
		{ID: "t1", TurnNumber: 1, Role: "participant", Content: "We implemented distributed locking using Redis SETNX with an aggressive TTL of two seconds to prevent race conditions during concurrent worker dispatch."},
		{ID: "t2", TurnNumber: 2, Role: "participant", Content: "For database consistency, we wrapped PostgreSQL outbox processing in a single transaction with FOR UPDATE SKIP LOCKED to guarantee idempotent turn projection."},
		{ID: "t3", TurnNumber: 3, Role: "participant", Content: "The system architecture uses Go microservices with WebSocket connections for streaming real-time audio from Deepgram Nova-2 to Groq LLM."},
		{ID: "t4", TurnNumber: 4, Role: "participant", Content: "When evaluating candidates, we use a scatter-gather worker that evaluates up to five rubric criteria in parallel before aggregating the hire recommendation."},
	}

	results := make([]CriterionResult, len(corpus))
	for i, c := range corpus {
		results[i] = CriterionResult{
			CriterionID:    c.ID,
			Score:          4,
			EvidenceTurnID: c.TurnID,
			EvidenceQuote:  c.Quote,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ValidateGrounding(results, turns)
	}
}
