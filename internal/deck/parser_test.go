package deck

import (
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
		})
	}
}
