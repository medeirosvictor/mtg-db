package main

import (
	"app/internal/config"
	"app/internal/deck"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Determine root dir: prefer the directory where the executable lives,
	// but fall back to cwd. During development (wails dev), cwd is correct.
	rootDir, err := findRootDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not determine root dir: %v\n", err)
		rootDir = ""
	}

	cfg, err := config.New(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: config init failed: %v\n", err)
		return
	}
	a.config = cfg

	// Load all decks on startup
	a.loadDecks()
}

// findRootDir locates the project root (the directory containing the decks/ folder).
func findRootDir() (string, error) {
	// First try cwd
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if hasDecksDir(cwd) {
		return cwd, nil
	}

	// Try executable directory
	exe, err := os.Executable()
	if err != nil {
		return cwd, nil // fall back to cwd
	}
	exeDir := filepath.Dir(exe)
	if hasDecksDir(exeDir) {
		return exeDir, nil
	}

	// Default to cwd
	return cwd, nil
}

func hasDecksDir(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "decks"))
	return err == nil && info.IsDir()
}

func (a *App) loadDecks() {
	if a.config == nil {
		return
	}
	decks, err := deck.LoadAllDecks(a.config.DecksDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load decks: %v\n", err)
		return
	}
	a.decks = decks
}

// --- Methods bound to the frontend ---

// DeckSummary is a lightweight view of a deck for the dashboard.
type DeckSummary struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Commander string `json:"commander"`
	Colors    string `json:"colors"`
	Status    string `json:"status"`
	CardCount int    `json:"cardCount"`
	Universe  string `json:"universe,omitempty"`
}

// GetAllDecks returns summaries of all loaded decks.
func (a *App) GetAllDecks() []DeckSummary {
	var summaries []DeckSummary
	for _, d := range a.decks {
		status := normalizeStatus(d.Info.Status)
		summaries = append(summaries, DeckSummary{
			Slug:      d.Slug,
			Title:     d.Info.Title,
			Commander: d.Info.Commander,
			Colors:    d.Info.Colors,
			Status:    status,
			CardCount: d.CardCount,
			Universe:  d.Info.Universe,
		})
	}
	return summaries
}

// GetDeck returns the full deck data for a given slug.
func (a *App) GetDeck(slug string) *deck.Deck {
	for _, d := range a.decks {
		if d.Slug == slug {
			return &d
		}
	}
	return nil
}

// ReloadDecks re-reads all deck files from disk.
func (a *App) ReloadDecks() []DeckSummary {
	a.loadDecks()
	return a.GetAllDecks()
}

// normalizeStatus extracts a clean status label.
func normalizeStatus(raw string) string {
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "owned"):
		return "Owned"
	case strings.Contains(lower, "planned"):
		return "Planned"
	case strings.Contains(lower, "disassembled"):
		return "Disassembled"
	default:
		return raw
	}
}
