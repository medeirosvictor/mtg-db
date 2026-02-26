package app

import (
	"app/internal/config"
	"app/internal/db"
	"app/internal/deck"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// App struct — bound to the Svelte frontend via Wails.
type App struct {
	ctx    context.Context
	config *config.Config
	decks  []deck.Deck
	mu     sync.RWMutex // Protects decks slice
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: config load failed: %v\n", err)
		return
	}
	a.config = cfg

	// Initialize SQLite database for card cache
	if a.config.HasActiveCollection() {
		dataDir := filepath.Join(a.config.AppDataDir, "images", "cache")
		if err := db.Init(dataDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: database init failed: %v\n", err)
		}
	}

	// If we have an active collection, validate it and load decks
	if a.config.HasActiveCollection() {
		if err := config.ValidateCollectionDir(a.config.ActiveCollectionPath()); err != nil {
			fmt.Fprintf(os.Stderr, "warning: saved collection path invalid: %v\n", err)
			return
		}
		a.loadDecks()
	}
}

func (a *App) loadDecks() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.config == nil || !a.config.HasActiveCollection() {
		a.decks = nil
		return
	}
	decks, err := deck.LoadAllDecks(a.config.DecksDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load decks: %v\n", err)
		a.decks = nil
		return
	}
	a.decks = decks
}

// getDeck returns a pointer to a deck by slug (thread-safe)
func (a *App) getDeck(slug string) *deck.Deck {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for i := range a.decks {
		if a.decks[i].Slug == slug {
			return &a.decks[i]
		}
	}
	return nil
}

// getDeckIndex returns the index of a deck by slug (caller must hold lock)
func (a *App) getDeckIndex(slug string) int {
	for i := range a.decks {
		if a.decks[i].Slug == slug {
			return i
		}
	}
	return -1
}

// setDeck updates a deck at the given index (caller must hold write lock)
func (a *App) setDeck(index int, d *deck.Deck) {
	if index >= 0 && index < len(a.decks) {
		a.decks[index] = *d
	}
}
