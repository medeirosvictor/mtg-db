package deck

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAllDecks(t *testing.T) {
	// Find the decks directory relative to the repo root
	decksDir := filepath.Join("..", "..", "decks")
	if _, err := os.Stat(decksDir); os.IsNotExist(err) {
		t.Skip("decks directory not found — skipping integration test")
	}

	decks, err := LoadAllDecks(decksDir)
	if err != nil {
		t.Fatalf("LoadAllDecks: %v", err)
	}

	if len(decks) == 0 {
		t.Fatal("expected at least one deck")
	}

	expectedDecks := map[string]bool{
		"abzan-desert":           true,
		"avatar-ally":            true,
		"desert-dune":            true,
		"finalfantasy-voltron":   true,
		"jumpscare":              true,
		"lotr-aragorn":           true,
		"sultai-rogues":          true,
		"warhammer-spellslinger": true,
	}

	for _, d := range decks {
		t.Logf("Loaded: %-25s | %3d cards | commander: %s | status: %s",
			d.Slug, d.CardCount, d.Info.Commander, d.Info.Status)

		if !expectedDecks[d.Slug] {
			t.Errorf("unexpected deck: %s", d.Slug)
		}
		delete(expectedDecks, d.Slug)

		if d.CardCount == 0 {
			t.Errorf("deck %s has 0 cards", d.Slug)
		}
		if d.Info.Title == "" {
			t.Errorf("deck %s has no title", d.Slug)
		}
		if d.Info.Commander == "" {
			t.Errorf("deck %s has no commander", d.Slug)
		}
	}

	for missing := range expectedDecks {
		t.Errorf("expected deck not loaded: %s", missing)
	}
}

func TestParseAllCardLines(t *testing.T) {
	// Parse every single line across all deck files to find any parsing gaps
	decksDir := filepath.Join("..", "..", "decks")
	if _, err := os.Stat(decksDir); os.IsNotExist(err) {
		t.Skip("decks directory not found — skipping integration test")
	}

	entries, err := os.ReadDir(decksDir)
	if err != nil {
		t.Fatal(err)
	}

	totalCards := 0
	parseErrors := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		deckFile := filepath.Join(decksDir, entry.Name(), "deck.txt")
		cards, err := ParseDeckFile(deckFile)
		if err != nil {
			t.Errorf("failed to parse %s: %v", deckFile, err)
			continue
		}
		for _, c := range cards {
			totalCards += c.Quantity
			if c.Name == "" {
				t.Errorf("empty card name in %s", entry.Name())
				parseErrors++
			}
		}
	}

	t.Logf("Parsed %d total cards across all decks (%d parse errors)", totalCards, parseErrors)

	if parseErrors > 0 {
		t.Errorf("%d cards had parse errors", parseErrors)
	}
}
