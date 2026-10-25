package search

import (
	"math"
	"strings"
	"unicode"

	"github.com/sbondCo/Watcharr/database/entity"
)

// NormalizeText normalizes a string for search: lowercasing, folding unicode accents/diacritics,
// and simplifying punctuation to whitespace.
func NormalizeText(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == 'á' || r == 'à' || r == 'ã' || r == 'â' || r == 'ä' || r == 'å':
			b.WriteByte('a')
		case r == 'é' || r == 'è' || r == 'ê' || r == 'ë':
			b.WriteByte('e')
		case r == 'í' || r == 'ì' || r == 'î' || r == 'ï':
			b.WriteByte('i')
		case r == 'ó' || r == 'ò' || r == 'õ' || r == 'ô' || r == 'ö':
			b.WriteByte('o')
		case r == 'ú' || r == 'ù' || r == 'û' || r == 'ü':
			b.WriteByte('u')
		case r == 'ç':
			b.WriteByte('c')
		case r == 'ñ':
			b.WriteByte('n')
		case unicode.IsSpace(r) || unicode.IsPunct(r):
			b.WriteByte(' ')
		default:
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				b.WriteRune(r)
			} else {
				b.WriteByte(' ')
			}
		}
	}

	// Collapse multiple spaces
	return strings.Join(strings.Fields(b.String()), " ")
}

// GenerateTrigrams returns a set of 3-character n-grams from the string.
// The string is prefixed with "^ " and suffixed with " $" to capture boundary transitions.
func GenerateTrigrams(s string) map[string]struct{} {
	norm := NormalizeText(s)
	if len(norm) == 0 {
		return nil
	}

	padded := "^ " + norm + " $"
	runes := []rune(padded)
	trigrams := make(map[string]struct{})

	if len(runes) < 3 {
		trigrams[padded] = struct{}{}
		return trigrams
	}

	for i := 0; i <= len(runes)-3; i++ {
		tri := string(runes[i : i+3])
		trigrams[tri] = struct{}{}
	}

	return trigrams
}

// TrigramSimilarity calculates the Dice coefficient between the trigram sets of s1 and s2.
// Returns a value between 0.0 (no overlap) and 1.0 (identical trigrams).
func TrigramSimilarity(s1, s2 string) float64 {
	tri1 := GenerateTrigrams(s1)
	tri2 := GenerateTrigrams(s2)

	if len(tri1) == 0 && len(tri2) == 0 {
		return 1.0
	}
	if len(tri1) == 0 || len(tri2) == 0 {
		return 0.0
	}

	intersection := 0
	for t := range tri1 {
		if _, exists := tri2[t]; exists {
			intersection++
		}
	}

	total := len(tri1) + len(tri2)
	return (2.0 * float64(intersection)) / float64(total)
}

// DamerauLevenshtein computes the edit distance between two strings,
// counting insertions, deletions, substitutions, and adjacent transpositions.
func DamerauLevenshtein(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	l1, l2 := len(r1), len(r2)

	if l1 == 0 {
		return l2
	}
	if l2 == 0 {
		return l1
	}

	// 2-row rolling matrix optimization for memory efficiency
	h := make([][]int, l1+2)
	for i := range h {
		h[i] = make([]int, l2+2)
	}

	maxDist := l1 + l2
	h[0][0] = maxDist
	for i := 0; i <= l1; i++ {
		h[i+1][0] = maxDist
		h[i+1][1] = i
	}
	for j := 0; j <= l2; j++ {
		h[0][j+1] = maxDist
		h[1][j+1] = j
	}

	lastRow := make(map[rune]int)

	for i := 1; i <= l1; i++ {
		db := 0
		for j := 1; j <= l2; j++ {
			i1 := lastRow[r2[j-1]]
			j1 := db
			cost := 1
			if r1[i-1] == r2[j-1] {
				cost = 0
				db = j
			}

			h[i+1][j+1] = minInt(
				h[i][j+1]+1,                     // deletion
				h[i+1][j]+1,                     // insertion
				h[i][j]+cost,                    // substitution
				h[i1][j1]+(i-i1-1)+1+(j-j1-1),   // transposition
			)
		}
		lastRow[r1[i-1]] = i
	}

	return h[l1+1][l2+1]
}

func minInt(vars ...int) int {
	min := vars[0]
	for _, v := range vars[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// ScoredCandidate represents an item with its calculated match score.
type ScoredCandidate struct {
	Watched entity.Watched
	Title   string
	Score   float64
}

// ScoreText evaluates how well target matches query.
// Returns a score between 0.0 and 1.0, where:
// 1.0   = Exact match
// 0.90+ = Exact prefix match
// 0.80+ = Substring or full token-set match
// 0.50+ = High fuzzy/trigram match with typo tolerance
// < 0.50 = Weak or non-match
func ScoreText(query, target string) float64 {
	qNorm := NormalizeText(query)
	tNorm := NormalizeText(target)

	if qNorm == "" || tNorm == "" {
		return 0.0
	}

	// 1. Exact match
	if qNorm == tNorm {
		return 1.0
	}

	// 2. Exact prefix match (e.g. "bat" -> "batman")
	if strings.HasPrefix(tNorm, qNorm) {
		ratio := float64(len(qNorm)) / float64(len(tNorm))
		return 0.85 + (0.10 * ratio)
	}

	// 3. Word-boundary prefix match (e.g. "rings" -> "the lord of the rings")
	words := strings.Fields(tNorm)
	for _, w := range words {
		if w == qNorm {
			return 0.85
		}
		if strings.HasPrefix(w, qNorm) && len(qNorm) >= 3 {
			return 0.80
		}
	}

	// 4. Token set inclusion (e.g. "lord rings" -> "the lord of the rings")
	qTokens := strings.Fields(qNorm)
	if len(qTokens) > 1 {
		matchedTokens := 0
		for _, qt := range qTokens {
			for _, wt := range words {
				if wt == qt || strings.HasPrefix(wt, qt) {
					matchedTokens++
					break
				}
			}
		}
		if matchedTokens == len(qTokens) {
			tokenRatio := float64(len(qNorm)) / float64(len(tNorm))
			return 0.75 + (0.10 * tokenRatio)
		}
	}

	// 5. General substring match
	if strings.Contains(tNorm, qNorm) {
		ratio := float64(len(qNorm)) / float64(len(tNorm))
		return 0.70 + (0.10 * ratio)
	}

	// 6. Typo-tolerant fuzzy matching (Trigram & Damerau-Levenshtein)
	triSim := TrigramSimilarity(qNorm, tNorm)

	// Check edit distance for words of reasonable length
	minWordDist := math.MaxInt32
	qLen := len(qNorm)

	for _, w := range words {
		wLen := len(w)
		if math.Abs(float64(wLen-qLen)) <= 2 {
			d := DamerauLevenshtein(qNorm, w)
			if d < minWordDist {
				minWordDist = d
			}
		}
	}

	// Single edit typo on a word (e.g. "oppenhymer" -> "oppenheimer", "rigns" -> "rings")
	if minWordDist == 1 && qLen >= 4 {
		return math.Max(0.72, triSim)
	}
	// Double edit typo on long words (>= 7 characters)
	if minWordDist == 2 && qLen >= 7 {
		return math.Max(0.60, triSim)
	}

	// If trigram similarity is strong enough
	if triSim >= 0.50 {
		return triSim * 0.80
	}

	return 0.0
}

// RankAndFilterLibrary applies the fuzzy and trigram search algorithm across library items.
// Items meeting minScore are returned ordered from highest score to lowest.
func RankAndFilterLibrary(items []entity.Watched, query string, minScore float64) []ScoredCandidate {
	if minScore <= 0 {
		minScore = 0.50
	}

	var results []ScoredCandidate
	for _, item := range items {
		title := ""
		if item.Content != nil && item.Content.Title != "" {
			title = item.Content.Title
		} else if item.Game != nil && item.Game.Name != "" {
			title = item.Game.Name
		}

		if title == "" {
			continue
		}

		score := ScoreText(query, title)
		if score >= minScore {
			results = append(results, ScoredCandidate{
				Watched: item,
				Title:   title,
				Score:   score,
			})
		}
	}

	// Sort descending by score
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}
