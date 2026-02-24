package app

import (
	"app/internal/config"
	"app/internal/db"
	"app/internal/deck"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// App struct — bound to the Svelte frontend via Wails.
type App struct {
	ctx    context.Context
	config *config.Config
	decks  []deck.Deck
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
