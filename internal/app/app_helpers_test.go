package app

import (
	"testing"
)

// =====================
// MEDIUM: Slug Generation Tests
// =====================

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple title", "My Deck", "my-deck"},
		{"uppercase", "ATRAXA SUPERFRIENDS", "atraxa-superfriends"},
		{"special characters", "Deck #1 (Best!)", "deck-1-best"},
		{"underscores to hyphens", "my_cool_deck", "my-cool-deck"},
		{"multiple spaces", "my   deck   name", "my-deck-name"},
		{"leading trailing hyphens", "---deck---", "deck"},
		{"numbers", "deck 42", "deck-42"},
		{"empty string", "", ""},
		{"only special chars", "!@#$%", ""},
		{"very long title", "this is an extremely long deck title that should get truncated at sixty characters because we dont want slugs to be too long", "this-is-an-extremely-long-deck-title-that-should-get-truncat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSlug(tt.input)
			if got != tt.want {
				t.Errorf("generateSlug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// =====================
// MEDIUM: Status Normalization Tests
// =====================

func TestNormalizeStatus(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"owned with emoji", "✅ Owned", "Owned"},
		{"planned with emoji", "📋 Planned", "Planned"},
		{"disassembled with emoji", "🔧 Disassembled", "Disassembled"},
		{"plain owned", "Owned", "Owned"},
		{"plain planned", "Planned", "Planned"},
		{"plain disassembled", "Disassembled", "Disassembled"},
		{"case insensitive owned", "OWNED", "Owned"},
		{"case insensitive planned", "planned", "Planned"},
		{"unknown status", "WIP", "WIP"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeStatus(tt.input)
			if got != tt.want {
				t.Errorf("normalizeStatus(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
