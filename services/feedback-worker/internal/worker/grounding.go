package worker

import (
	"strings"
	"unicode"

	"github.com/yourusername/ai-interview/services/feedback-worker/internal/repository"
)

type SessionTurn = repository.SessionTurn

const (
	// MinSubstantiveWords is the minimum number of substantive words required for an evidence quote.
	MinSubstantiveWords = 3
	// MinSubstantiveChars is the minimum character length for an evidence quote.
	MinSubstantiveChars = 15
)

// NormalizeQuote strips punctuation, smart quotes, and collapsed whitespace for contract verification.
func NormalizeQuote(s string) string {
	// 1. Replace Unicode smart quotes and apostrophes with ASCII
	replacer := strings.NewReplacer(
		"“", "\"",
		"”", "\"",
		"‘", "'",
		"’", "'",
		"—", " ",
		"–", " ",
	)
	s = replacer.Replace(s)

	// 2. Trim whitespace and boundary punctuation
	s = strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune("\"'`.,!?;:-()[]{}", r)
	})

	// 3. Normalize internal whitespace
	fields := strings.Fields(s)
	return strings.ToLower(strings.Join(fields, " "))
}

// IsSubstantiveEvidence validates that an evidence quote is substantive rather than a generic stop-token.
// Contract Policy:
// The assessment contract requires candidate claims to be backed by verifiable verbatim quotes.
// Single stop-words or generic phrases (e.g. "we", "to", "in a") match indiscriminately across any transcript
// and do not constitute legitimate evidence of competence. A quote must contain either >= 3 substantive words
// (words with >= 3 characters) or >= 15 normalized characters.
func IsSubstantiveEvidence(normalizedQuote string) bool {
	if len(normalizedQuote) < MinSubstantiveChars {
		words := strings.Fields(normalizedQuote)
		substantiveWordCount := 0
		for _, w := range words {
			cleanWord := strings.TrimFunc(w, unicode.IsPunct)
			if len(cleanWord) >= 3 {
				substantiveWordCount++
			}
		}
		if substantiveWordCount < MinSubstantiveWords {
			return false
		}
	}
	return true
}

// ValidateGrounding verifies whether each result's evidence_quote exists verbatim in the cited turn.
//
// CONTRACT SPECIFICATION:
// 1. Verbatim Requirement: The evaluation prompt instructs the LLM: "evidence_quote MUST be copied verbatim
//    from the content of that turn. Do not invent or paraphrase." Therefore, paraphrased responses are rejected
//    as failing contract compliance; this rejection is expected behavior, not a false rejection.
// 2. Attribution Integrity: The verbatim quote must exist within the cited EvidenceTurnID. Quotes appearing in
//    other turns are rejected (prevents wrong-turn attribution).
// 3. Substantive Floor: Quotes failing IsSubstantiveEvidence (e.g. "we", "in a") are rejected to eliminate
//    false-positive matches on trivial tokens.
// 4. Normalization: Whitespace is collapsed, case is folded, and surrounding punctuation/quotes (including unicode
//    curly quotes) are normalized before containment check.
func ValidateGrounding(results []CriterionResult, turns []SessionTurn) []CriterionResult {
	turnContent := make(map[string]string, len(turns))
	for _, t := range turns {
		// Store normalized turn content
		turnContent[t.ID] = NormalizeQuote(t.Content)
	}

	for i, r := range results {
		if r.Error != nil {
			results[i].Grounded = false
			continue
		}
		normalizedTurn, ok := turnContent[r.EvidenceTurnID]
		if !ok {
			results[i].Grounded = false
			continue
		}

		quoteNorm := NormalizeQuote(r.EvidenceQuote)
		if quoteNorm == "" {
			results[i].Grounded = false
			continue
		}

		if !IsSubstantiveEvidence(quoteNorm) {
			results[i].Grounded = false
			continue
		}

		results[i].Grounded = strings.Contains(normalizedTurn, quoteNorm)
	}
	return results
}
