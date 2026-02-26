package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"app/internal/config"
	"app/internal/deck"
)

// setupAppWithDeck creates a temp collection with a deck and returns the App.
func setupAppWithDeck(t *testing.T) (*App, string, func()) {
	t.Helper()

	// Create temp collection dir with decks/
	tmpDir, err := os.MkdirTemp("", "app-cardops-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	decksDir := filepath.Join(tmpDir, "decks", "test-deck")
	if err := os.MkdirAll(decksDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("create deck dir: %v", err)
	}

	// Write deck.txt
	deckContent := `1x Lightning Bolt
1x Counterspell
2x Sol Ring #commander
`
	if err := os.WriteFile(filepath.Join(decksDir, "deck.txt"), []byte(deckContent), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("write deck.txt: %v", err)
	}

	// Write info.md
	infoContent := `# Test Deck

- **Status:** ✅ Owned
- **Colors:** UR
- **Commander:** Sol Ring
`
	if err := os.WriteFile(filepath.Join(decksDir, "info.md"), []byte(infoContent), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("write info.md: %v", err)
	}

	// Create app config
	cfg := &config.Config{
		AppDataDir:     tmpDir,
		ConfigFilePath: filepath.Join(tmpDir, "config.json"),
		File: config.ConfigFile{
			ActiveCollection: tmpDir,
			Collections: []config.Collection{
				{Path: tmpDir, Label: "Test", LastOpened: time.Now()},
			},
			Preferences: config.DefaultPreferences(),
		},
	}
	if err := cfg.Save(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("save config: %v", err)
	}

	a := &App{config: cfg}
	a.loadDecks()

	cleanup := func() { os.RemoveAll(tmpDir) }
	return a, tmpDir, cleanup
}

// =====================
// CRITICAL: ToggleCardTag Tests
// =====================

func TestToggleCardTag_AddCommander(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	// Add commander tag to Lightning Bolt
	d := a.ToggleCardTag("test-deck", "Lightning Bolt", "commander")
	if d == nil {
		t.Fatal("ToggleCardTag returned nil")
	}

	// Find the card
	var bolt *deck.Card
	for i := range d.Cards {
		if d.Cards[i].Name == "Lightning Bolt" {
			bolt = &d.Cards[i]
			break
		}
	}
	if bolt == nil {
		t.Fatal("Lightning Bolt not found in deck")
	}
	if !bolt.HasTag("commander") {
		t.Error("Lightning Bolt should have commander tag")
	}

	// Verify on disk
	content, _ := os.ReadFile(filepath.Join(a.config.DecksDir(), "test-deck", "deck.txt"))
	if !contains(string(content), "#commander") {
		t.Error("deck.txt should contain #commander")
	}
}

func TestToggleCardTag_RemoveCommander(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	// Sol Ring already has #commander in our test data
	// Remove it
	d := a.ToggleCardTag("test-deck", "Sol Ring", "commander")
	if d == nil {
		t.Fatal("ToggleCardTag returned nil")
	}

	var solRing *deck.Card
	for i := range d.Cards {
		if d.Cards[i].Name == "Sol Ring" {
			solRing = &d.Cards[i]
			break
		}
	}
	if solRing == nil {
		t.Fatal("Sol Ring not found")
	}
	if solRing.HasTag("commander") {
		t.Error("Sol Ring should NOT have commander tag after removal")
	}
}

func TestToggleCardTag_ProxyTag(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	d := a.ToggleCardTag("test-deck", "Lightning Bolt", "proxy")
	if d == nil {
		t.Fatal("ToggleCardTag returned nil")
	}

	// Find and check
	var bolt *deck.Card
	for i := range d.Cards {
		if d.Cards[i].Name == "Lightning Bolt" {
			bolt = &d.Cards[i]
			break
		}
	}
	if bolt == nil {
		t.Fatal("Lightning Bolt not found")
	}
	if !bolt.HasTag("proxy") {
		t.Error("Lightning Bolt should have proxy tag")
	}
}

func TestToggleCardTag_WishlistTag(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	d := a.ToggleCardTag("test-deck", "Lightning Bolt", "wishlist")
	if d == nil {
		t.Fatal("ToggleCardTag returned nil")
	}

	var bolt *deck.Card
	for i := range d.Cards {
		if d.Cards[i].Name == "Lightning Bolt" {
			bolt = &d.Cards[i]
			break
		}
	}
	if !bolt.HasTag("wishlist") {
		t.Error("Lightning Bolt should have wishlist tag")
	}
}

func TestToggleCardTag_NoDeck(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	d := a.ToggleCardTag("nonexistent-deck", "Card", "commander")
	if d != nil {
		t.Error("should return nil for nonexistent deck")
	}
}

func TestToggleCardTag_NoCard(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	d := a.ToggleCardTag("test-deck", "Nonexistent Card", "commander")
	if d != nil {
		t.Error("should return nil for nonexistent card")
	}
}

// =====================
// CRITICAL: AddCards Tests
// =====================

func TestAddCards_NewCard(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	errMsg := a.AddCards("test-deck", "1x Demonic Tutor\n")
	if errMsg != "" {
		t.Fatalf("AddCards returned error: %s", errMsg)
	}

	// Verify card was added
	d := a.ToggleCardTag("test-deck", "Demonic Tutor", "") // just to reload
	_ = d

	// Read deck file
	content, _ := os.ReadFile(filepath.Join(a.config.DecksDir(), "test-deck", "deck.txt"))
	if !contains(string(content), "Demonic Tutor") {
		t.Error("deck.txt should contain Demonic Tutor")
	}
}

func TestAddCards_ExistingCard(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	// Add Lightning Bolt again — should increment quantity
	errMsg := a.AddCards("test-deck", "1x Lightning Bolt\n")
	if errMsg != "" {
		t.Fatalf("AddCards error: %s", errMsg)
	}

	// Verify quantity is now 2
	d := a.ToggleCardTag("test-deck", "Lightning Bolt", "")
	if d != nil {
		for _, c := range d.Cards {
			if c.Name == "Lightning Bolt" && c.Quantity != 2 {
				t.Errorf("Lightning Bolt quantity = %d, want 2", c.Quantity)
			}
		}
	}
}

func TestAddCards_SkipsComments(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	errMsg := a.AddCards("test-deck", "# This is a comment\n1x New Card\n")
	if errMsg != "" {
		t.Fatalf("AddCards error: %s", errMsg)
	}

	// Verify only New Card was added
	content, _ := os.ReadFile(filepath.Join(a.config.DecksDir(), "test-deck", "deck.txt"))
	if contains(string(content), "This is a comment") {
		t.Error("comments should be skipped")
	}
}

// =====================
// CRITICAL: UpdateCardQuantity Tests
// =====================

func TestUpdateCardQuantity_Increase(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	errMsg := a.UpdateCardQuantity("test-deck", "Lightning Bolt", 5)
	if errMsg != "" {
		t.Fatalf("UpdateCardQuantity error: %s", errMsg)
	}

	// Verify
	d := a.ToggleCardTag("test-deck", "Lightning Bolt", "")
	if d != nil {
		for _, c := range d.Cards {
			if c.Name == "Lightning Bolt" && c.Quantity != 5 {
				t.Errorf("Lightning Bolt quantity = %d, want 5", c.Quantity)
			}
		}
	}
}

func TestUpdateCardQuantity_Remove(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	errMsg := a.UpdateCardQuantity("test-deck", "Lightning Bolt", 0)
	if errMsg != "" {
		t.Fatalf("UpdateCardQuantity error: %s", errMsg)
	}

	// Verify removed from file
	content, _ := os.ReadFile(filepath.Join(a.config.DecksDir(), "test-deck", "deck.txt"))
	if contains(string(content), "Lightning Bolt") {
		t.Error("Lightning Bolt should be removed from deck.txt")
	}
}

// =====================
// CRITICAL: RemoveCard Tests
// =====================

func TestRemoveCard(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	errMsg := a.RemoveCard("test-deck", "Lightning Bolt")
	if errMsg != "" {
		t.Fatalf("RemoveCard error: %s", errMsg)
	}

	// Verify
	content, _ := os.ReadFile(filepath.Join(a.config.DecksDir(), "test-deck", "deck.txt"))
	if contains(string(content), "Lightning Bolt") {
		t.Error("Lightning Bolt should be removed")
	}
}

// =====================
// CRITICAL: UpdateCardText Tests
// =====================

func TestUpdateCardText_ChangeName(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	// Rename Lightning Bolt to "Fire Bolt"
	errMsg := a.UpdateCardText("test-deck", "Lightning Bolt", "1x Fire Bolt (M19) 126")
	if errMsg != "" {
		t.Fatalf("UpdateCardText error: %s", errMsg)
	}

	// Verify
	content, _ := os.ReadFile(filepath.Join(a.config.DecksDir(), "test-deck", "deck.txt"))
	if !contains(string(content), "Fire Bolt") {
		t.Error("deck.txt should contain Fire Bolt")
	}
	if contains(string(content), "Lightning Bolt") {
		t.Error("deck.txt should NOT contain old name")
	}
}

func TestUpdateCardText_InvalidLine(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	errMsg := a.UpdateCardText("test-deck", "Lightning Bolt", "invalid")
	if errMsg == "" {
		t.Error("should return error for invalid card line")
	}
}

// =====================
// MEDIUM: ReloadDecks Tests
// =====================

func TestReloadDecks(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	// Modify deck file externally
	deckPath := filepath.Join(a.config.DecksDir(), "test-deck", "deck.txt")
	os.WriteFile(deckPath, []byte("1x New Card\n"), 0644)

	// Reload
	summaries := a.ReloadDecks()
	if len(summaries) != 1 {
		t.Fatalf("expected 1 deck, got %d", len(summaries))
	}

	// The reloaded deck should have the new card
	found := false
	for _, d := range a.decks {
		if d.Slug == "test-deck" {
			for _, c := range d.Cards {
				if c.Name == "New Card" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("Reloaded deck should contain New Card")
	}
}

// =====================
// MEDIUM: Card Count Updates After Operations
// =====================

func TestCardCount_AfterAddCards(t *testing.T) {
	a, _, cleanup := setupAppWithDeck(t)
	defer cleanup()

	initialCount := 0
	for _, d := range a.decks {
		if d.Slug == "test-deck" {
			initialCount = d.CardCount
		}
	}

	a.AddCards("test-deck", "1x New Card\n")

	// Reload to get updated count
	a.loadDecks()
	newCount := 0
	for _, d := range a.decks {
		if d.Slug == "test-deck" {
			newCount = d.CardCount
		}
	}

	if newCount != initialCount+1 {
		t.Errorf("CardCount = %d, want %d", newCount, initialCount+1)
	}
}

// Helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
