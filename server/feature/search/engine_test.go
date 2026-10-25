package search_test

import (
	"testing"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/search"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World!", "hello world"},
		{"Amélie Poulain", "amelie poulain"},
		{"Spider-Man: No Way Home", "spider man no way home"},
		{"  Multiple   Spaces  --  ", "multiple spaces"},
		{"Coração D'Aço", "coracao d aco"},
	}

	for _, tt := range tests {
		got := search.NormalizeText(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeText(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestDamerauLevenshtein(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"kitten", "sitting", 3},
		{"rings", "rigns", 1}, // transposition
		{"oppenheimer", "oppenhaimer", 1}, // substitution
		{"batman", "batman", 0},
		{"matrix", "mtarix", 1}, // transposition
	}

	for _, tt := range tests {
		got := search.DamerauLevenshtein(tt.s1, tt.s2)
		if got != tt.expected {
			t.Errorf("DamerauLevenshtein(%q, %q) = %d; want %d", tt.s1, tt.s2, got, tt.expected)
		}
	}
}

func TestScoreText(t *testing.T) {
	// Exact Match
	scoreExact := search.ScoreText("Inception", "Inception")
	if scoreExact != 1.0 {
		t.Fatalf("expected exact score 1.0, got %f", scoreExact)
	}

	// Accent Invariance
	scoreAccent := search.ScoreText("amelie", "Amélie")
	if scoreAccent != 1.0 {
		t.Fatalf("expected accent invariant score 1.0, got %f", scoreAccent)
	}

	// Prefix Match
	scorePrefix := search.ScoreText("Incep", "Inception")
	if scorePrefix < 0.85 {
		t.Fatalf("expected prefix score >= 0.85, got %f", scorePrefix)
	}

	// Token Set Match
	scoreTokens := search.ScoreText("lord rings", "The Lord of the Rings")
	if scoreTokens < 0.75 {
		t.Fatalf("expected token set score >= 0.75, got %f", scoreTokens)
	}

	// Typo Tolerance - Transposition
	scoreTransposition := search.ScoreText("rigns", "The Lord of the Rings")
	if scoreTransposition < 0.65 {
		t.Fatalf("expected transposition typo score >= 0.65, got %f", scoreTransposition)
	}

	// Typo Tolerance - Missing Letter / Single Edit
	scoreTypo := search.ScoreText("openheimer", "Oppenheimer")
	if scoreTypo < 0.65 {
		t.Fatalf("expected typo score >= 0.65, got %f", scoreTypo)
	}

	// Irrelevant content should score near 0
	scoreIrrelevant := search.ScoreText("Harry Potter", "The Matrix")
	if scoreIrrelevant > 0.20 {
		t.Fatalf("expected irrelevant score <= 0.20, got %f", scoreIrrelevant)
	}
}

func TestRankAndFilterLibrary(t *testing.T) {
	items := []entity.Watched{
		{
			Content: &entity.Content{
				Title: "The Matrix Reloaded",
			},
		},
		{
			Content: &entity.Content{
				Title: "The Matrix",
			},
		},
		{
			Content: &entity.Content{
				Title: "Matrx Revolutions", // typo in library or test
			},
		},
		{
			Content: &entity.Content{
				Title: "Interstellar",
			},
		},
	}

	results := search.RankAndFilterLibrary(items, "The Matrix", 0.50)
	if len(results) < 2 {
		t.Fatalf("expected at least 2 matrix results, got %d", len(results))
	}

	// Exact match "The Matrix" must be ranked 1st!
	if results[0].Title != "The Matrix" {
		t.Errorf("expected 1st ranked to be 'The Matrix', got %q", results[0].Title)
	}
	if results[0].Score != 1.0 {
		t.Errorf("expected score 1.0 for exact match, got %f", results[0].Score)
	}

	// Second ranked should be "The Matrix Reloaded" (prefix/substring)
	if results[1].Title != "The Matrix Reloaded" {
		t.Errorf("expected 2nd ranked to be 'The Matrix Reloaded', got %q", results[1].Title)
	}

	// "Interstellar" should not appear in results for "The Matrix"
	for _, r := range results {
		if r.Title == "Interstellar" {
			t.Errorf("did not expect 'Interstellar' in matrix search results")
		}
	}
}
