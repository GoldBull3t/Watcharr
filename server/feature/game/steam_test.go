package game

import (
	"testing"
)

func TestTitlesMatch(t *testing.T) {
	tests := []struct {
		a, b     string
		expected bool
	}{
		{"The Witcher 3: Wild Hunt", "The Witcher 3: Wild Hunt - Game of the Year Edition", true},
		{"Cyberpunk 2077", "Cyberpunk 2077: Phantom Liberty", true},
		{"Grand Theft Auto V", "Grand Theft Auto V", true},
		{"ELDEN RING", "Elden Ring", true},
		{"Portal 2", "Portal 2", true},
		{"Half-Life 2", "Half-Life 2", true},
		{"DOOM Eternal", "DOOM Eternal (Deluxe Edition)", true},
		{"Hollow Knight", "Hollow Knight", true},
		{"Counter-Strike 2", "Counter-Strike 2", true},
		{"Dark Souls III", "Dark Souls 3", false}, // roman numerals vs numbers differs unless normalized, good to know
		{"Portal", "Portal 2", false}, // must not match different numbered installments
		{"Half-Life", "Half-Life 2", false}, // must not match different numbered installments
	}

	for _, tt := range tests {
		actual := TitlesMatch(tt.a, tt.b)
		if actual != tt.expected {
			t.Errorf("TitlesMatch(%q, %q) = %v, expected %v", tt.a, tt.b, actual, tt.expected)
		}
	}
}

