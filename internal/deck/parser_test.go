package deck

import (
	"os"
	"path/filepath"
	"testing"
)

// Test helper: create temp directory with deck files
func createTestDeckDir(t *testing.T, files map[string]string) string {
	tmpDir, err := os.MkdirTemp("", "deck-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	return tmpDir
}

func cleanupTestDir(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Logf("failed to cleanup temp dir: %v", err)
	}
}

// =====================
// CRITICAL: Card Parsing Tests
// =====================

func TestParseCardLine_QuantityFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantQty  int
		wantName string
		wantSet  string
		wantFoil bool
	}{
		{"1x format", "1x Lightning Bolt", 1, "Lightning Bolt", "", false},
		{"1x with set", "1x Lightning Bolt (M19)", 1, "Lightning Bolt", "M19", false},
		{"1x with set and collector", "1x Lightning Bolt (M19) 126", 1, "Lightning Bolt", "M19", false},
		{"1x foil", "1x Lightning Bolt *F*", 1, "Lightning Bolt", "", true},
		{"1x set collector foil", "1x Lightning Bolt (M19) 126 *F*", 1, "Lightning Bolt", "M19", true},
		{"plain quantity", "3 Counterspell", 3, "Counterspell", "", false},
		{"quantity without x", "4 Sol Ring", 4, "Sol Ring", "", false},
		{"double-faced card", "1x Card A / Card B (SET) 123", 1, "Card A / Card B", "SET", false},
		{"non-standard collector", "1x Card (PLST) CMA-30", 1, "Card", "PLST", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card, err := ParseCardLine(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if card.Quantity != tt.wantQty {
				t.Errorf("quantity = %d, want %d", card.Quantity, tt.wantQty)
			}
			if card.Name != tt.wantName {
				t.Errorf("name = %q, want %q", card.Name, tt.wantName)
			}
			if card.SetCode != tt.wantSet {
				t.Errorf("setCode = %q, want %q", card.SetCode, tt.wantSet)
			}
			if card.Foil != tt.wantFoil {
				t.Errorf("foil = %v, want %v", card.Foil, tt.wantFoil)
			}
		})
	}
}

func TestParseCardLine_Tags(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTags  []string
		wantName  string
	}{
		{"commander tag", "1x Sol Ring #commander", []string{"commander"}, "Sol Ring"},
		{"proxy tag", "1x Sol Ring #proxy", []string{"proxy"}, "Sol Ring"},
		{"multiple tags", "1x Sol Ring #commander #proxy", []string{"commander", "proxy"}, "Sol Ring"},
		{"mixed case tag", "1x Sol Ring #Commander #Proxy", []string{"commander", "proxy"}, "Sol Ring"},
		{"wishlist tag", "1x Sol Ring #wishlist", []string{"wishlist"}, "Sol Ring"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card, err := ParseCardLine(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(card.Tags) != len(tt.wantTags) {
				t.Errorf("tags = %v, want %v", card.Tags, tt.wantTags)
			}
			for i, tag := range tt.wantTags {
				if i >= len(card.Tags) || card.Tags[i] != tag {
					t.Errorf("tags[%d] = %q, want %q", i, card.Tags[i], tag)
				}
			}
		})
	}
}

func TestParseCardLine_SkipInvalid(t *testing.T) {
	invalid := []string{
		"",           // empty
		"# comment",  // comment
		"// comment", // comment
		"   ",        // whitespace
	}

	for _, line := range invalid {
		_, err := ParseCardLine(line)
		if err == nil {
			t.Errorf("expected error for %q, got nil", line)
		}
	}
}

func TestParseDeckFile(t *testing.T) {
	content := `1x Lightning Bolt
1x Counterspell #commander
2x Forest #wishlist
# This is a comment
1x Sol Ring (M19) 123 *F*
`
	dir := createTestDeckDir(t, map[string]string{
		"deck.txt": content,
	})
	defer cleanupTestDir(t, dir)

	cards, err := ParseDeckFile(filepath.Join(dir, "deck.txt"))
	if err != nil {
		t.Fatalf("ParseDeckFile error: %v", err)
	}

	if len(cards) != 4 {
		t.Errorf("got %d cards, want 4", len(cards))
	}

	// Check first card
	if cards[0].Name != "Lightning Bolt" {
		t.Errorf("first card name = %q, want %q", cards[0].Name, "Lightning Bolt")
	}
	if cards[0].Quantity != 1 {
		t.Errorf("first card quantity = %d, want 1", cards[0].Quantity)
	}

	// Check card with commander tag
	commanderCard := cards[1]
	if !commanderCard.HasTag("commander") {
		t.Errorf("expected commander tag on %s", commanderCard.Name)
	}

	// Check wishlist card
	wishlistCard := cards[2]
	if !wishlistCard.HasTag("wishlist") {
		t.Errorf("expected wishlist tag on %s", wishlistCard.Name)
	}

	// Check foil card
	foilCard := cards[3]
	if !foilCard.Foil {
		t.Errorf("expected foil on %s", foilCard.Name)
	}
}

// =====================
// CRITICAL: Deck Info Parsing Tests
// =====================

func TestParseInfoFile(t *testing.T) {
	content := `# My Awesome Deck

- **Status:** ✅ Owned
- **Colors:** WUG
- **Commander:** Teferi, Hero of Dominaria
- **Strategy:** Control the board and win with card advantage
- **Universe:** Custom
`
	dir := createTestDeckDir(t, map[string]string{
		"info.md": content,
	})
	defer cleanupTestDir(t, dir)

	info, err := ParseInfoFile(filepath.Join(dir, "info.md"))
	if err != nil {
		t.Fatalf("ParseInfoFile error: %v", err)
	}

	if info.Title != "My Awesome Deck" {
		t.Errorf("title = %q, want %q", info.Title, "My Awesome Deck")
	}
	if info.Status != "✅ Owned" {
		t.Errorf("status = %q, want %q", info.Status, "✅ Owned")
	}
	if info.Colors != "WUG" {
		t.Errorf("colors = %q, want %q", info.Colors, "WUG")
	}
	if info.Commander != "Teferi, Hero of Dominaria" {
		t.Errorf("commander = %q, want %q", info.Commander, "Teferi, Hero of Dominaria")
	}
	if info.Strategy != "Control the board and win with card advantage" {
		t.Errorf("strategy = %q, want %q", info.Strategy, "Control the board and win with card advantage")
	}
	if info.Universe != "Custom" {
		t.Errorf("universe = %q, want %q", info.Universe, "Custom")
	}
}

func TestWriteInfoFile(t *testing.T) {
	info := DeckInfo{
		Title:     "Test Deck",
		Status:    "📋 Planned",
		Colors:    "BR",
		Commander: "Edgar Markov",
		Strategy:  "Go wide with vampires",
		Universe:  "MTG",
	}

	tmpDir := createTestDeckDir(t, map[string]string{})
	defer cleanupTestDir(t, tmpDir)

	path := filepath.Join(tmpDir, "info.md")
	if err := WriteInfoFile(path, info); err != nil {
		t.Fatalf("WriteInfoFile error: %v", err)
	}

	// Read back and verify
	loaded, err := ParseInfoFile(path)
	if err != nil {
		t.Fatalf("ParseInfoFile error: %v", err)
	}

	if loaded.Title != info.Title {
		t.Errorf("title = %q, want %q", loaded.Title, info.Title)
	}
	if loaded.Status != info.Status {
		t.Errorf("status = %q, want %q", loaded.Status, info.Status)
	}
	if loaded.Colors != info.Colors {
		t.Errorf("colors = %q, want %q", loaded.Colors, info.Colors)
	}
}

// =====================
// HIGH: Card Operations Tests
// =====================

func TestAddCard(t *testing.T) {
	cards := []Card{}

	// Add new card
	cards = AddCard(cards, "Lightning Bolt", 1)
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Name != "Lightning Bolt" || cards[0].Quantity != 1 {
		t.Errorf("card = %+v", cards[0])
	}

	// Add to existing card
	cards = AddCard(cards, "Lightning Bolt", 2)
	if len(cards) != 1 {
		t.Fatalf("expected 1 card (incremented), got %d", len(cards))
	}
	if cards[0].Quantity != 3 {
		t.Errorf("quantity = %d, want 3", cards[0].Quantity)
	}

	// Add different card
	cards = AddCard(cards, "Counterspell", 2)
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
}

func TestRemoveCard(t *testing.T) {
	cards := []Card{
		{Name: "Lightning Bolt", Quantity: 1},
		{Name: "Counterspell", Quantity: 2},
		{Name: "Sol Ring", Quantity: 1},
	}

	cards = RemoveCard(cards, "Lightning Bolt")
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	if cards[0].Name != "Counterspell" {
		t.Errorf("remaining cards should start with Counterspell, got %s", cards[0].Name)
	}

	// Test case-insensitive
	cards = RemoveCard(cards, "COUNTERSPELL")
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
}

func TestUpdateCardQty(t *testing.T) {
	cards := []Card{
		{Name: "Lightning Bolt", Quantity: 1},
		{Name: "Counterspell", Quantity: 2},
	}

	// Update existing
	cards = UpdateCardQty(cards, "Lightning Bolt", 4)
	if cards[0].Quantity != 4 {
		t.Errorf("quantity = %d, want 4", cards[0].Quantity)
	}

	// Add new
	cards = UpdateCardQty(cards, "New Card", 3)
	if len(cards) != 3 {
		t.Fatalf("expected 3 cards, got %d", len(cards))
	}
	if cards[2].Name != "New Card" || cards[2].Quantity != 3 {
		t.Errorf("new card = %+v", cards[2])
	}

	// Remove (quantity <= 0)
	cards = UpdateCardQty(cards, "Counterspell", 0)
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards after remove, got %d", len(cards))
	}
}

func TestSetCardTag(t *testing.T) {
	cards := []Card{
		{Name: "Lightning Bolt", Tags: []string{"commander"}},
		{Name: "Counterspell", Tags: []string{}},
	}

	// Add tag
	cards = SetCardTag(cards, "Lightning Bolt", "proxy", true)
	if !cards[0].HasTag("proxy") {
		t.Errorf("expected proxy tag on Lightning Bolt")
	}

	// Add tag to card without tags
	cards = SetCardTag(cards, "Counterspell", "proxy", true)
	if !cards[1].HasTag("proxy") {
		t.Errorf("expected proxy tag on Counterspell")
	}

	// Remove tag
	cards = SetCardTag(cards, "Lightning Bolt", "commander", false)
	if cards[0].HasTag("commander") {
		t.Errorf("commander tag should be removed")
	}
}

// =====================
// MEDIUM: Format and Load Tests
// =====================

func TestFormatCardLine(t *testing.T) {
	card := Card{
		Quantity:        2,
		Name:            "Lightning Bolt",
		SetCode:         "M19",
		CollectorNumber: "126",
		Foil:            true,
		Tags:            []string{"commander", "proxy"},
	}

	result := FormatCardLine(card)
	expected := "2x Lightning Bolt (M19) 126 *F* #commander #proxy"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestLoadDeck(t *testing.T) {
	dir := createTestDeckDir(t, map[string]string{
		"deck.txt": `1x Lightning Bolt
1x Counterspell #commander
`,
		"info.md": `# Test Deck

- **Status:** ✅ Owned
- **Colors:** UR
`,
	})
	defer cleanupTestDir(t, dir)

	deck, err := LoadDeck(dir)
	if err != nil {
		t.Fatalf("LoadDeck error: %v", err)
	}

	if deck.Info.Title != "Test Deck" {
		t.Errorf("title = %q, want %q", deck.Info.Title, "Test Deck")
	}
	if deck.Info.Status != "✅ Owned" {
		t.Errorf("status = %q, want %q", deck.Info.Status, "✅ Owned")
	}
	if deck.Info.Colors != "UR" {
		t.Errorf("colors = %q, want %q", deck.Info.Colors, "UR")
	}
	if len(deck.Cards) != 2 {
		t.Errorf("got %d cards, want 2", len(deck.Cards))
	}
	if deck.CardCount != 2 {
		t.Errorf("cardCount = %d, want 2", deck.CardCount)
	}

	// Check commander derived from tags
	if deck.Info.Commander != "Counterspell" {
		t.Errorf("commander = %q, want %q", deck.Info.Commander, "Counterspell")
	}
}

func TestGetCommanders(t *testing.T) {
	cards := []Card{
		{Name: "Card1", Tags: []string{"commander"}},
		{Name: "Card2", Tags: []string{}},
		{Name: "Card3", Tags: []string{"commander", "proxy"}},
	}

	commanders := GetCommanders(cards)
	if len(commanders) != 2 {
		t.Errorf("got %d commanders, want 2", len(commanders))
	}
	// Should be sorted
	if commanders[0] != "Card1" || commanders[1] != "Card3" {
		t.Errorf("commanders = %v, want [Card1 Card3]", commanders)
	}
}
