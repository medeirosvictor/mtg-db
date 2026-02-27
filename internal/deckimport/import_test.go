package deckimport

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// =====================
// CRITICAL: Import Detection Tests
// =====================

func TestDetectAndImport_MoxfieldURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"name":"Test Deck","mainboard":{"Lightning Bolt":{"quantity":4,"card":{"name":"Lightning Bolt"}}},"commanders":{},"sideboard":{}}`)
	}))
	defer srv.Close()

	origBase := moxfieldAPIBase
	moxfieldAPIBase = srv.URL
	defer func() { moxfieldAPIBase = origBase }()

	tests := []struct {
		name  string
		input string
	}{
		{"full url https", "https://moxfield.com/decks/abc123"},
		{"full url http", "http://moxfield.com/decks/abc123"},
		{"www url", "www.moxfield.com/decks/abc123"},
		{"no protocol", "moxfield.com/decks/abc123"},
		{"with underscores", "moxfield.com/decks/abc_123-def"},
		{"with numbers", "moxfield.com/decks/AbC123XyZ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectAndImport(tt.input)
			if result.Error != "" {
				t.Errorf("DetectAndImport(%q) returned error: %s", tt.input, result.Error)
			}
			if result.Source != "moxfield" {
				t.Errorf("DetectAndImport(%q) source = %q, want %q", tt.input, result.Source, "moxfield")
			}
		})
	}
}

func TestDetectAndImport_ArchidektURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"name":"Test Deck","cards":[{"quantity":4,"card":{"oracleCard":{"name":"Lightning Bolt"}}}]}`)
	}))
	defer srv.Close()

	origBase := archidektAPIBase
	archidektAPIBase = srv.URL
	defer func() { archidektAPIBase = origBase }()

	tests := []struct {
		name  string
		input string
	}{
		{"full url https", "https://archidekt.com/decks/12345"},
		{"full url http", "http://archidekt.com/decks/12345"},
		{"www url", "www.archidekt.com/decks/12345"},
		{"no protocol", "archidekt.com/decks/12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectAndImport(tt.input)
			if result.Error != "" {
				t.Errorf("DetectAndImport(%q) returned error: %s", tt.input, result.Error)
			}
			if result.Source != "archidekt" {
				t.Errorf("DetectAndImport(%q) source = %q, want %q", tt.input, result.Source, "archidekt")
			}
		})
	}
}

func TestDetectAndImport_Text(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{
			name:    "simple list",
			input:   "1x Lightning Bolt\n2x Counterspell",
			wantLen: 2,
		},
		{
			name:    "list without x",
			input:   "1 Lightning Bolt\n3 Counterspell",
			wantLen: 2,
		},
		{
			name:    "with set code",
			input:   "1x Lightning Bolt (M19)",
			wantLen: 1,
		},
		{
			name:    "with foil",
			input:   "1x Lightning Bolt *F*",
			wantLen: 1,
		},
		{
			name:    "with tags",
			input:   "1x Lightning Bolt #commander",
			wantLen: 1,
		},
		{
			name:    "comments ignored",
			input:   "# This is a comment\n1x Lightning Bolt\n// Another comment",
			wantLen: 1,
		},
		{
			name:    "plain card name only",
			input:   "Lightning Bolt",
			wantLen: 1,
		},
		{
			name:    "multiple cards",
			input:   "1x Sol Ring\n1x Dark Ritual\n1x Demonic Tutor\n4x Lightning Bolt",
			wantLen: 4,
		},
		{
			name:    "empty lines ignored",
			input:   "1x Card A\n\n\n2x Card B\n",
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectAndImport(tt.input)
			if result.Source != "text" {
				t.Errorf("expected source text, got %s", result.Source)
			}
			if len(result.Cards) != tt.wantLen {
				t.Errorf("got %d cards, want %d", len(result.Cards), tt.wantLen)
			}
		})
	}
}

func TestImportFromText_Complex(t *testing.T) {
	input := `# Commander Deck

1x Sol Ring
1x Mana Crypt
1x Demonic Tutor #wishlist
2x Lightning Bolt (M19) 126 *F*
1x Counterspell (ME4)
`
	result := ImportFromText(input)

	if result.Source != "text" {
		t.Fatalf("expected source text, got %s", result.Source)
	}

	if len(result.Cards) != 5 {
		t.Fatalf("expected 5 cards, got %d", len(result.Cards))
	}

	// Check specific cards
	expected := []struct {
		name string
		qty  int
	}{
		{"Sol Ring", 1},
		{"Mana Crypt", 1},
		{"Demonic Tutor", 1},
		{"Lightning Bolt", 2},
		{"Counterspell", 1},
	}

	for i, exp := range expected {
		if i >= len(result.Cards) {
			break
		}
		if result.Cards[i].Name != exp.name {
			t.Errorf("card[%d] = %q, want %q", i, result.Cards[i].Name, exp.name)
		}
		if result.Cards[i].Quantity != exp.qty {
			t.Errorf("card[%d] qty = %d, want %d", i, result.Cards[i].Quantity, exp.qty)
		}
	}
}

// =====================
// HIGH: URL Regex Tests
// =====================

func TestMoxfieldURLRegex(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"moxfield.com/decks/abc", true},
		{"https://moxfield.com/decks/abc", true},
		{"https://www.moxfield.com/decks/abc", true},
		{"moxfield.com/decks/abc-123_def", true},
		{"archidekt.com/decks/123", false},
		{"random.com/decks/abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := moxfieldURLRegex.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestArchidektURLRegex(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"archidekt.com/decks/123", true},
		{"https://archidekt.com/decks/123", true},
		{"https://www.archidekt.com/decks/123", true},
		{"archidekt.com/decks/123456789", true},
		{"moxfield.com/decks/abc", false},
		{"random.com/decks/123", false},
		{"archidekt.com/decks/abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := archidektURLRegex.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
