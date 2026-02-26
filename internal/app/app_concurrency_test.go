package app

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"app/internal/config"
)

// =====================
// CONCURRENCY: App Deck Operations Tests
// These tests verify behavior under concurrent deck operations.
// The App.decks slice is modified by multiple methods without synchronization.
// =====================

// setupConcurrentApp creates an App with a test deck for concurrency testing.
func setupConcurrentApp(t *testing.T) (*App, string, func()) {
	t.Helper()

	// Create temp collection dir
	tmpDir, err := os.MkdirTemp("", "app-concurrent-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	// Create decks directory
	decksDir := filepath.Join(tmpDir, "decks")
	if err := os.MkdirAll(decksDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("create decks dir: %v", err)
	}

	// Create test deck directory
	deckDir := filepath.Join(decksDir, "concurrent-deck")
	if err := os.MkdirAll(deckDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("create deck dir: %v", err)
	}

	// Write deck.txt
	deckContent := `1x Lightning Bolt
1x Counterspell
1x Sol Ring
1x Dark Ritual
1x Demonic Tutor
`
	if err := os.WriteFile(filepath.Join(deckDir, "deck.txt"), []byte(deckContent), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("write deck.txt: %v", err)
	}

	// Write info.md
	infoContent := `# Concurrent Deck

- **Status:** Active
- **Colors:** UB
`
	if err := os.WriteFile(filepath.Join(deckDir, "info.md"), []byte(infoContent), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("write info.md: %v", err)
	}

	// Create config and app
	cfg := &config.Config{
		AppDataDir:     tmpDir,
		ConfigFilePath: filepath.Join(tmpDir, "config.json"),
		File: config.ConfigFile{
			ActiveCollection: tmpDir,
		},
	}

	app := &App{
		config: cfg,
	}

	// Load decks
	app.loadDecks()

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return app, tmpDir, cleanup
}

// TestConcurrentDeckReads tests concurrent reads of deck data.
func TestConcurrentDeckReads(t *testing.T) {
	app, _, cleanup := setupConcurrentApp(t)
	defer cleanup()

	var wg sync.WaitGroup
	numGoroutines := 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Concurrent reads of deck data
			_ = app.GetAllDecks()
			_ = app.GetDeckBasic("concurrent-deck")
		}()
	}

	wg.Wait()
}

// TestConcurrentCardOperations tests concurrent card modifications.
// This test can expose race conditions in the App.decks slice.
func TestConcurrentCardOperations(t *testing.T) {
	app, _, cleanup := setupConcurrentApp(t)
	defer cleanup()

	var wg sync.WaitGroup
	numWorkers := 10

	// Concurrent card additions
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			cardLine := "1x Card-" + string(rune('A'+workerID))
			_ = app.AddCards("concurrent-deck", cardLine)
		}(w)
	}

	// Concurrent quantity updates
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			_ = app.UpdateCardQuantity("concurrent-deck", "Lightning Bolt", 2)
		}(w)
	}

	// Concurrent reads
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = app.GetDeckBasic("concurrent-deck")
			_ = app.GetAllDecks()
		}()
	}

	wg.Wait()

	// Verify deck is still accessible
	result := app.GetDeckBasic("concurrent-deck")
	if result == nil {
		t.Error("Deck became nil after concurrent operations")
	}
}

// TestConcurrentTagToggles tests concurrent tag toggling on cards.
func TestConcurrentTagToggles(t *testing.T) {
	app, _, cleanup := setupConcurrentApp(t)
	defer cleanup()

	// Get initial deck state
	deck := app.GetDeckBasic("concurrent-deck")
	if deck == nil {
		t.Fatal("Deck not found")
	}

	// Get a card name from the deck
	cardName := deck.Cards[0].Name

	var wg sync.WaitGroup

	// Concurrent tag toggles on the same card
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = app.ToggleCardTag("concurrent-deck", cardName, "test-tag")
		}()
	}

	wg.Wait()

	// Verify deck is still valid
	result := app.GetDeckBasic("concurrent-deck")
	if result == nil {
		t.Error("Deck is nil after concurrent tag toggles")
	}
}

// TestConcurrentMixedOperations tests a mix of deck operations concurrently.
func TestConcurrentMixedOperations(t *testing.T) {
	app, _, cleanup := setupConcurrentApp(t)
	defer cleanup()

	var wg sync.WaitGroup

	// Mix of different operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = app.GetAllDecks()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = app.GetDeckBasic("concurrent-deck")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = app.AddCards("concurrent-deck", "1x New Card\n")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = app.ReloadDecks()
		}()
	}

	wg.Wait()

	// Verify final state
	decks := app.GetAllDecks()
	if len(decks) != 1 {
		t.Errorf("Expected 1 deck, got %d", len(decks))
	}
}

// TestConcurrentDeckReload tests concurrent reload operations.
func TestConcurrentDeckReload(t *testing.T) {
	app, _, cleanup := setupConcurrentApp(t)
	defer cleanup()

	var wg sync.WaitGroup

	// Multiple concurrent reloads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = app.ReloadDecks()
		}()
	}

	wg.Wait()

	// Should still have decks loaded
	decks := app.GetAllDecks()
	if len(decks) != 1 {
		t.Errorf("Expected 1 deck after reloads, got %d", len(decks))
	}
}

// TestConcurrentReadAndReload tests concurrent reads while reloading.
func TestConcurrentReadAndReload(t *testing.T) {
	app, _, cleanup := setupConcurrentApp(t)
	defer cleanup()

	var wg sync.WaitGroup

	// Continuous reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = app.GetDeckBasic("concurrent-deck")
				_ = app.GetAllDecks()
			}
		}()
	}

	// Concurrent reloads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				_ = app.ReloadDecks()
			}
		}()
	}

	wg.Wait()

	// Final state should be valid
	decks := app.GetAllDecks()
	if len(decks) != 1 {
		t.Errorf("Expected 1 deck, got %d", len(decks))
	}
}

// TestConcurrentCardTextUpdates tests concurrent card text updates.
func TestConcurrentCardTextUpdates(t *testing.T) {
	app, _, cleanup := setupConcurrentApp(t)
	defer cleanup()

	var wg sync.WaitGroup
	cardName := "Lightning Bolt"

	// Update the same card concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			newLine := "2x Lightning Bolt"
			_ = app.UpdateCardText("concurrent-deck", cardName, newLine)
		}()
	}

	wg.Wait()

	// Deck should still be valid
	result := app.GetDeckBasic("concurrent-deck")
	if result == nil {
		t.Error("Deck is nil after concurrent updates")
	}
}
