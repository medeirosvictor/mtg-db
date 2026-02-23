package main

import (
	"app/internal/config"
	"app/internal/deck"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
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

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: config load failed: %v\n", err)
		return
	}
	a.config = cfg

	// If we have an active collection, validate it and load decks
	if a.config.HasActiveCollection() {
		if err := config.ValidateCollectionDir(a.config.ActiveCollectionPath()); err != nil {
			// Saved path is invalid — don't crash, frontend will handle it
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

// --- App State Methods ---

// AppState is sent to the frontend on startup to determine what view to show.
type AppState struct {
	HasCollection    bool              `json:"hasCollection"`
	CollectionPath   string            `json:"collectionPath"`
	CollectionLabel  string            `json:"collectionLabel"`
	CollectionValid  bool              `json:"collectionValid"`
	Collections      []CollectionInfo  `json:"collections"`
	NeedsSetup       bool              `json:"needsSetup"`
}

// CollectionInfo is a frontend-friendly view of a known collection.
type CollectionInfo struct {
	Path       string `json:"path"`
	Label      string `json:"label"`
	LastOpened string `json:"lastOpened"`
	IsActive   bool   `json:"isActive"`
	IsValid    bool   `json:"isValid"`
}

// GetAppState returns the current app state for the frontend.
func (a *App) GetAppState() AppState {
	if a.config == nil {
		return AppState{NeedsSetup: true}
	}

	state := AppState{
		HasCollection: a.config.HasActiveCollection(),
		NeedsSetup:    !a.config.HasActiveCollection(),
	}

	if state.HasCollection {
		state.CollectionPath = a.config.ActiveCollectionPath()
		err := config.ValidateCollectionDir(state.CollectionPath)
		state.CollectionValid = err == nil

		// Find label
		for _, c := range a.config.File.Collections {
			if c.Path == state.CollectionPath {
				state.CollectionLabel = c.Label
				break
			}
		}
	}

	// Build collections list
	for _, c := range a.config.File.Collections {
		err := config.ValidateCollectionDir(c.Path)
		state.Collections = append(state.Collections, CollectionInfo{
			Path:       c.Path,
			Label:      c.Label,
			LastOpened: c.LastOpened.Format("2006-01-02"),
			IsActive:   c.Path == a.config.ActiveCollectionPath(),
			IsValid:    err == nil,
		})
	}

	return state
}

// --- Collection Management Methods ---

// SelectCollectionFolder opens a native folder picker dialog and validates the selection.
// Returns an error message string (empty on success).
func (a *App) SelectCollectionFolder() string {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select your MTG collection folder",
	})
	if err != nil {
		return fmt.Sprintf("Failed to open folder picker: %v", err)
	}
	if dir == "" {
		return "" // User cancelled — not an error
	}

	// Validate
	if err := config.ValidateCollectionDir(dir); err != nil {
		return err.Error()
	}

	// Set as active
	if err := a.config.SetActiveCollection(dir); err != nil {
		return fmt.Sprintf("Failed to save config: %v", err)
	}

	a.loadDecks()
	return ""
}

// InitializeAndSelectFolder opens a native folder picker, creates collection skeleton,
// and sets it as active. Returns an error message string (empty on success).
func (a *App) InitializeAndSelectFolder() string {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Choose where to create your MTG collection",
	})
	if err != nil {
		return fmt.Sprintf("Failed to open folder picker: %v", err)
	}
	if dir == "" {
		return "" // User cancelled
	}

	// Check if it already has decks/
	if err := config.ValidateCollectionDir(dir); err == nil {
		// Already a valid collection, just use it
		if err := a.config.SetActiveCollection(dir); err != nil {
			return fmt.Sprintf("Failed to save config: %v", err)
		}
		a.loadDecks()
		return ""
	}

	// Initialize the skeleton
	if err := config.InitializeCollectionDir(dir); err != nil {
		return fmt.Sprintf("Failed to initialize collection: %v", err)
	}

	// It won't have any decks yet, but set it as active
	if err := a.config.SetActiveCollection(dir); err != nil {
		return fmt.Sprintf("Failed to save config: %v", err)
	}

	a.loadDecks()
	return ""
}

// SwitchCollection switches to a different known collection by path.
// Returns an error message string (empty on success).
func (a *App) SwitchCollection(path string) string {
	if err := config.ValidateCollectionDir(path); err != nil {
		return err.Error()
	}

	if err := a.config.SetActiveCollection(path); err != nil {
		return fmt.Sprintf("Failed to save config: %v", err)
	}

	a.loadDecks()
	return ""
}

// RenameCollection updates the label for a collection.
func (a *App) RenameCollection(path, label string) string {
	if err := a.config.SetCollectionLabel(path, label); err != nil {
		return err.Error()
	}
	return ""
}

// RemoveCollection removes a collection from the known list (does not delete files).
func (a *App) RemoveKnownCollection(path string) string {
	if err := a.config.RemoveCollection(path); err != nil {
		return err.Error()
	}
	a.loadDecks()
	return ""
}

// --- Deck Methods ---

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

// ToggleCardTag toggles a tag on a card in a deck and writes to disk.
// Returns the updated deck, or nil on error.
func (a *App) ToggleCardTag(slug string, cardName string, tag string) *deck.Deck {
	if a.config == nil || !a.config.HasActiveCollection() {
		return nil
	}

	deckDir := filepath.Join(a.config.DecksDir(), slug)
	deckPath := filepath.Join(deckDir, "deck.txt")

	// Find the deck in memory
	var deckIdx int = -1
	for i, d := range a.decks {
		if d.Slug == slug {
			deckIdx = i
			break
		}
	}
	if deckIdx == -1 {
		return nil
	}

	d := &a.decks[deckIdx]

	// Find the card and toggle the tag
	found := false
	for i := range d.Cards {
		if strings.EqualFold(d.Cards[i].Name, cardName) {
			found = true
			if d.Cards[i].HasTag(tag) {
				// Remove tag
				var newTags []string
				for _, t := range d.Cards[i].Tags {
					if t != tag {
						newTags = append(newTags, t)
					}
				}
				d.Cards[i].Tags = newTags
			} else {
				// For commander: enforce max 2
				if tag == deck.TagCommander {
					commanders := deck.GetCommanders(d.Cards)
					if len(commanders) >= 2 {
						// Already at max, don't add
						return d
					}
				}
				d.Cards[i].Tags = append(d.Cards[i].Tags, tag)
			}
			break
		}
	}
	if !found {
		return nil
	}

	// Write updated deck.txt
	if err := deck.WriteDeckFile(deckPath, d.Cards); err != nil {
		fmt.Fprintf(os.Stderr, "error writing deck file: %v\n", err)
		return nil
	}

	// Update commander in info from tags
	commanders := deck.GetCommanders(d.Cards)
	if len(commanders) > 0 {
		d.Info.Commander = strings.Join(commanders, " / ")
	} else {
		// If no commander tags, re-read from info.md
		info, err := deck.ParseInfoFile(filepath.Join(deckDir, "info.md"))
		if err == nil {
			d.Info.Commander = info.Commander
		}
	}

	// Recalculate card count
	d.CardCount = 0
	for _, c := range d.Cards {
		d.CardCount += c.Quantity
	}

	return d
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
