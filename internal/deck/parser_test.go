package deck

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCardLine(t *testing.T) {
	tests := []struct {
		input string
		want  Card
		err   bool
	}{
		// Basic formats
		{
			input: "1 Sol Ring",
			want:  Card{Quantity: 1, Name: "Sol Ring"},
		},
		{
			input: "1x Sol Ring",
			want:  Card{Quantity: 1, Name: "Sol Ring"},
		},
		{
			input: "7 Forest",
			want:  Card{Quantity: 7, Name: "Forest"},
		},
		// With set code and collector number
		{
			input: "1 Aang and Katara (TLE) 69",
			want:  Card{Quantity: 1, Name: "Aang and Katara", SetCode: "TLE", CollectorNumber: "69"},
		},
		{
			input: "1x Éowyn, Shieldmaiden",
			want:  Card{Quantity: 1, Name: "Éowyn, Shieldmaiden"},
		},
		// Double-faced cards
		{
			input: "1 Aang, at the Crossroads / Aang, Destined Savior (TLA) 203",
			want:  Card{Quantity: 1, Name: "Aang, at the Crossroads / Aang, Destined Savior", SetCode: "TLA", CollectorNumber: "203"},
		},
		// With foil marker — collector number "23s" is a promo variant
		{
			input: "1 Hakoda, Selfless Commander (PTLA) 23s *F*",
			want:  Card{Quantity: 1, Name: "Hakoda, Selfless Commander", SetCode: "PTLA", CollectorNumber: "23s", Foil: true},
		},
		// With {num} variant (from master purchase list)
		{
			input: "1x Avatar Aang {207} // Aang, Master of Elements {207}",
			want:  Card{Quantity: 1, Name: "Avatar Aang {207} // Aang, Master of Elements", CollectorNumber: "207"},
		},
		// Set code with hyphenated collector number (PLST cross-references)
		{
			input: "1 Arcane Denial (PLST) CMA-30",
			want:  Card{Quantity: 1, Name: "Arcane Denial", SetCode: "PLST", CollectorNumber: "CMA-30"},
		},
		// Multiple quantity
		{
			input: "3x Anger",
			want:  Card{Quantity: 3, Name: "Anger"},
		},
		// --- Tag tests ---
		{
			input: "1x Sol Ring #commander",
			want:  Card{Quantity: 1, Name: "Sol Ring", Tags: []string{"commander"}},
		},
		{
			input: "1x Mana Crypt #proxy #wishlist",
			want:  Card{Quantity: 1, Name: "Mana Crypt", Tags: []string{"proxy", "wishlist"}},
		},
		{
			input: "1 Avenger of Zendikar (ZNR) 178 #proxy",
			want:  Card{Quantity: 1, Name: "Avenger of Zendikar", SetCode: "ZNR", CollectorNumber: "178", Tags: []string{"proxy"}},
		},
		{
			input: "1x Hazezon, Shaper of Sand #commander",
			want:  Card{Quantity: 1, Name: "Hazezon, Shaper of Sand", Tags: []string{"commander"}},
		},
		// Comments and blanks
		{
			input: "# Cards to buy/print for this deck",
			err:   true,
		},
		{
			input: "",
			err:   true,
		},
		{
			input: "// comment",
			err:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseCardLine(tt.input)
			if tt.err {
				if err == nil {
					t.Errorf("expected error for %q, got %+v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error for %q: %v", tt.input, err)
				return
			}
			if got.Quantity != tt.want.Quantity {
				t.Errorf("quantity: got %d, want %d", got.Quantity, tt.want.Quantity)
			}
			if got.Name != tt.want.Name {
				t.Errorf("name: got %q, want %q", got.Name, tt.want.Name)
			}
			if got.SetCode != tt.want.SetCode {
				t.Errorf("setCode: got %q, want %q", got.SetCode, tt.want.SetCode)
			}
			if got.CollectorNumber != tt.want.CollectorNumber {
				t.Errorf("collectorNumber: got %q, want %q", got.CollectorNumber, tt.want.CollectorNumber)
			}
			if got.Foil != tt.want.Foil {
				t.Errorf("foil: got %v, want %v", got.Foil, tt.want.Foil)
			}
			// Compare tags
			if len(got.Tags) != len(tt.want.Tags) {
				t.Errorf("tags: got %v, want %v", got.Tags, tt.want.Tags)
			} else {
				for i := range got.Tags {
					if got.Tags[i] != tt.want.Tags[i] {
						t.Errorf("tags[%d]: got %q, want %q", i, got.Tags[i], tt.want.Tags[i])
					}
				}
			}
		})
	}
}

func TestFormatCardLine(t *testing.T) {
	tests := []struct {
		card Card
		want string
	}{
		{
			card: Card{Quantity: 1, Name: "Sol Ring"},
			want: "1x Sol Ring",
		},
		{
			card: Card{Quantity: 1, Name: "Sol Ring", Tags: []string{"commander"}},
			want: "1x Sol Ring #commander",
		},
		{
			card: Card{Quantity: 1, Name: "Mana Crypt", Tags: []string{"proxy", "wishlist"}},
			want: "1x Mana Crypt #proxy #wishlist",
		},
		{
			card: Card{Quantity: 1, Name: "Avenger of Zendikar", SetCode: "ZNR", CollectorNumber: "178"},
			want: "1x Avenger of Zendikar (ZNR) 178",
		},
		{
			card: Card{Quantity: 7, Name: "Forest"},
			want: "7x Forest",
		},
		{
			card: Card{Quantity: 1, Name: "Hakoda", SetCode: "PTLA", CollectorNumber: "23s", Foil: true},
			want: "1x Hakoda (PTLA) 23s *F*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatCardLine(tt.card)
			if got != tt.want {
				t.Errorf("FormatCardLine: got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetCardTag(t *testing.T) {
	cards := []Card{
		{Quantity: 1, Name: "Sol Ring"},
		{Quantity: 1, Name: "Hazezon, Shaper of Sand"},
		{Quantity: 1, Name: "Avenger of Zendikar", Tags: []string{"proxy"}},
	}

	// Add commander tag
	cards = SetCardTag(cards, "Hazezon, Shaper of Sand", TagCommander, true)
	if !cards[1].HasTag(TagCommander) {
		t.Error("expected Hazezon to have commander tag")
	}

	// Add proxy tag to Sol Ring
	cards = SetCardTag(cards, "Sol Ring", TagProxy, true)
	if !cards[0].HasTag(TagProxy) {
		t.Error("expected Sol Ring to have proxy tag")
	}

	// Remove proxy tag from Avenger
	cards = SetCardTag(cards, "Avenger of Zendikar", TagProxy, false)
	if cards[2].HasTag(TagProxy) {
		t.Error("expected Avenger to not have proxy tag")
	}

	// Adding same tag twice should not duplicate
	cards = SetCardTag(cards, "Hazezon, Shaper of Sand", TagCommander, true)
	count := 0
	for _, tag := range cards[1].Tags {
		if tag == TagCommander {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 commander tag, got %d", count)
	}
}

func TestRoundTrip(t *testing.T) {
	// Write cards → file → read back, ensure they match
	dir := t.TempDir()
	path := filepath.Join(dir, "deck.txt")

	original := []Card{
		{Quantity: 1, Name: "Sol Ring", Tags: []string{"commander"}},
		{Quantity: 1, Name: "Mana Crypt", Tags: []string{"proxy", "wishlist"}},
		{Quantity: 7, Name: "Forest"},
		{Quantity: 1, Name: "Avenger of Zendikar", SetCode: "ZNR", CollectorNumber: "178"},
		{Quantity: 1, Name: "Hakoda", SetCode: "PTLA", CollectorNumber: "23s", Foil: true},
	}

	err := WriteDeckFile(path, original)
	if err != nil {
		t.Fatalf("WriteDeckFile: %v", err)
	}

	// Verify the file content
	content, _ := os.ReadFile(path)
	t.Logf("Written file:\n%s", string(content))

	parsed, err := ParseDeckFile(path)
	if err != nil {
		t.Fatalf("ParseDeckFile: %v", err)
	}

	if len(parsed) != len(original) {
		t.Fatalf("expected %d cards, got %d", len(original), len(parsed))
	}

	for i, orig := range original {
		got := parsed[i]
		if got.Name != orig.Name {
			t.Errorf("[%d] name: got %q, want %q", i, got.Name, orig.Name)
		}
		if got.Quantity != orig.Quantity {
			t.Errorf("[%d] quantity: got %d, want %d", i, got.Quantity, orig.Quantity)
		}
		if len(got.Tags) != len(orig.Tags) {
			t.Errorf("[%d] tags: got %v, want %v", i, got.Tags, orig.Tags)
		}
	}
}
