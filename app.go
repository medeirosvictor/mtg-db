package main

import (
	"app/internal/config"
	"app/internal/db"
	"app/internal/deck"
	"app/internal/scryfall"
	"context"
	"fmt"
	"log"
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
	HasCollection   bool             `json:"hasCollection"`
	CollectionPath  string           `json:"collectionPath"`
	CollectionLabel string           `json:"collectionLabel"`
	CollectionValid bool             `json:"collectionValid"`
	Collections     []CollectionInfo `json:"collections"`
	NeedsSetup      bool             `json:"needsSetup"`
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
// If syncScryfall is true, it will fetch card data from Scryfall (may be slow on first load).
// Set syncScryfall to false for faster initial load.
func (a *App) GetDeck(slug string) *deck.Deck {
	var d *deck.Deck
	for i := range a.decks {
		if a.decks[i].Slug == slug {
			d = &a.decks[i]
			break
		}
	}
	if d == nil {
		return nil
	}

	// Sync cards with Scryfall
	a.syncDeckCards(d.Cards)

	return d
}

// GetDeckBasic returns deck data without Scryfall sync (faster).
func (a *App) GetDeckBasic(slug string) *deck.Deck {
	for i := range a.decks {
		if a.decks[i].Slug == slug {
			return &a.decks[i]
		}
	}
	return nil
}

// syncDeckCards fetches Scryfall data for all cards in a deck
func (a *App) syncDeckCards(cards []deck.Card) map[string]*deck.ScryfallData {
	if a.config == nil {
		log.Println("[Sync] No config, skipping Scryfall sync")
		return nil
	}

	result := make(map[string]*deck.ScryfallData)
	var names []string
	for i := range cards {
		names = append(names, cards[i].Name)
	}

	if len(names) == 0 {
		return result
	}

	log.Printf("[Sync] Starting sync for %d cards", len(names))

	// Create Scryfall client
	client := scryfall.NewClient()

	// Check which cards need fetching (not in cache or stale)
	var toFetch []string
	for _, name := range names {
		cached, err := db.GetCard(name)
		if err == nil && cached == nil {
			toFetch = append(toFetch, name)
		} else if err == nil && cached != nil {
			// Check if stale (>24 hours)
			stale, _ := db.IsStale(name, 24)
			if stale {
				toFetch = append(toFetch, name)
			}
		}
	}

	log.Printf("[Sync] Cards to fetch from Scryfall: %d", len(toFetch))

	// Fetch missing/stale cards from Scryfall
	if len(toFetch) > 0 {
		log.Println("[Sync] Calling Scryfall API...")
		scCards, notFound, err := client.FetchCardsByNames(toFetch)
		if err != nil {
			log.Printf("[Sync] Scryfall API error: %v", err)
		} else {
			log.Printf("[Sync] Got %d cards from Scryfall, %d not found", len(scCards), len(notFound))

			// Cache successful lookups
			for _, sc := range scCards {
				cached := db.FromScryfallCard(sc)
				db.UpsertCard(cached)

				// Download image
				if sc.ImageURIs.Normal != "" {
					cacheDir := filepath.Join(a.config.AppDataDir, "images", "cache")
					slug := sc.Name + "-" + sc.SetCode
					_, err := scryfall.DownloadImage(sc.ImageURIs.Normal, cacheDir, slug)
					if err != nil {
						log.Printf("[Sync] Image download failed for %s: %v", sc.Name, err)
					}
				}
			}

			// Record unmatched cards
			for _, name := range notFound {
				log.Printf("[Sync] Card not found: %s", name)
				db.AddUnmatchedCard(name, "")
			}
		}
	} else {
		log.Println("[Sync] All cards already cached")
	}

	// Load all cards from cache and populate Scryfall data
	cachedCount := 0
	for i, name := range names {
		cached, err := db.GetCard(name)
		if err == nil && cached != nil {
			sfData := &deck.ScryfallData{
				OracleText:    cached.OracleText,
				TypeLine:      cached.TypeLine,
				ManaCost:      cached.ManaCost,
				CMC:           cached.CMC,
				ImageURI:      cached.ImageURI,
				PriceUSD:      cached.PriceUSD.String,
				PriceUSDFoil:  cached.PriceUSDFoil.String,
				ColorIdentity: cached.ColorIdentity,
			}
			result[name] = sfData
			// Also update the card in the slice
			cards[i].Scryfall = sfData
			cachedCount++
		}
	}

	log.Printf("[Sync] Loaded Scryfall data for %d/%d cards", cachedCount, len(names))

	return result
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

// AddCards adds one or more cards to a deck.
// cardLines should be in the format "1x Card Name" or "1 Card Name" (one per line)
// Returns the updated deck, or error message.
func (a *App) AddCards(slug string, cardLines string) string {
	if a.config == nil || !a.config.HasActiveCollection() {
		return "No active collection"
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
		return "Deck not found"
	}

	d := &a.decks[deckIdx]

	// Parse each line and add cards
	lines := strings.Split(cardLines, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		card, err := deck.ParseCardLine(line)
		if err != nil {
			continue // skip invalid lines
		}

		d.Cards = deck.AddCard(d.Cards, card.Name, card.Quantity)
	}

	// Write to file
	if err := deck.WriteDeckFile(deckPath, d.Cards); err != nil {
		return fmt.Sprintf("Failed to write deck file: %v", err)
	}

	// Recalculate card count
	d.CardCount = 0
	for _, c := range d.Cards {
		d.CardCount += c.Quantity
	}

	return ""
}

// UpdateCardQuantity updates the quantity of a card in a deck.
// If quantity is 0, removes the card.
// Returns the updated deck, or error message.
func (a *App) UpdateCardQuantity(slug string, cardName string, quantity int) string {
	if a.config == nil || !a.config.HasActiveCollection() {
		return "No active collection"
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
		return "Deck not found"
	}

	d := &a.decks[deckIdx]

	// Update quantity
	d.Cards = deck.UpdateCardQty(d.Cards, cardName, quantity)

	// Write to file
	if err := deck.WriteDeckFile(deckPath, d.Cards); err != nil {
		return fmt.Sprintf("Failed to write deck file: %v", err)
	}

	// Recalculate card count
	d.CardCount = 0
	for _, c := range d.Cards {
		d.CardCount += c.Quantity
	}

	return ""
}

// RemoveCard removes a card from a deck by name.
// Returns the updated deck, or error message.
func (a *App) RemoveCard(slug string, cardName string) string {
	return a.UpdateCardQuantity(slug, cardName, 0)
}

// UpdateDeckInfo updates the deck's title and strategy in info.md.
// Returns error message string (empty on success).
func (a *App) UpdateDeckInfo(slug string, title string, strategy string) string {
	if a.config == nil || !a.config.HasActiveCollection() {
		return "No active collection"
	}

	deckDir := filepath.Join(a.config.DecksDir(), slug)
	infoPath := filepath.Join(deckDir, "info.md")

	// Find the deck in memory to update it
	var deckIdx int = -1
	for i, d := range a.decks {
		if d.Slug == slug {
			deckIdx = i
			break
		}
	}

	// Write to info.md
	info := deck.DeckInfo{
		Title:    title,
		Strategy: strategy,
	}
	// Preserve existing fields if deck is loaded
	if deckIdx >= 0 {
		d := &a.decks[deckIdx]
		info.Status = d.Info.Status
		info.Colors = d.Info.Colors
		info.Commander = d.Info.Commander
		info.Universe = d.Info.Universe
	}

	if err := deck.WriteInfoFile(infoPath, info); err != nil {
		return fmt.Sprintf("Failed to write info file: %v", err)
	}

	// Update in-memory deck
	if deckIdx >= 0 {
		a.decks[deckIdx].Info.Title = title
		a.decks[deckIdx].Info.Strategy = strategy
	}

	return ""
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
