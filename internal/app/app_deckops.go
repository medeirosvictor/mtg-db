package app

import (
	"app/internal/deck"
	"app/internal/deckimport"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// UpdateDeckInfo updates the deck's title and strategy in info.md.
func (a *App) UpdateDeckInfo(slug string, title string, strategy string) string {
	if a.config == nil || !a.config.HasActiveCollection() {
		return "No active collection"
	}

	deckDir := filepath.Join(a.config.DecksDir(), slug)
	infoPath := filepath.Join(deckDir, "info.md")

	var deckIdx int = -1
	for i, d := range a.decks {
		if d.Slug == slug {
			deckIdx = i
			break
		}
	}

	info := deck.DeckInfo{
		Title:    title,
		Strategy: strategy,
	}
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

	if deckIdx >= 0 {
		a.decks[deckIdx].Info.Title = title
		a.decks[deckIdx].Info.Strategy = strategy
	}

	return ""
}

// UpdateDeckStatus updates the deck's status in info.md.
func (a *App) UpdateDeckStatus(slug string, status string) string {
	if a.config == nil || !a.config.HasActiveCollection() {
		return "No active collection"
	}

	deckDir := filepath.Join(a.config.DecksDir(), slug)
	infoPath := filepath.Join(deckDir, "info.md")

	var deckIdx int = -1
	for i, d := range a.decks {
		if d.Slug == slug {
			deckIdx = i
			break
		}
	}

	info := deck.DeckInfo{}
	if deckIdx >= 0 {
		d := &a.decks[deckIdx]
		info.Title = d.Info.Title
		info.Strategy = d.Info.Strategy
		info.Colors = d.Info.Colors
		info.Commander = d.Info.Commander
		info.Universe = d.Info.Universe
	}

	// Format status with emoji
	info.Status = "📋 " + status
	if status == "Owned" {
		info.Status = "✅ Owned"
	} else if status == "Planned" {
		info.Status = "📋 Planned"
	} else if status == "Disassembled" {
		info.Status = "🔧 Disassembled"
	}

	if err := deck.WriteInfoFile(infoPath, info); err != nil {
		return fmt.Sprintf("Failed to write info file: %v", err)
	}

	if deckIdx >= 0 {
		a.decks[deckIdx].Info.Status = info.Status
	}

	return ""
}

// ImportDeck imports cards into a deck from a URL or raw text.
func (a *App) ImportDeck(slug string, input string, mode string) *deckimport.ImportResult {
	if a.config == nil || !a.config.HasActiveCollection() {
		return &deckimport.ImportResult{Error: "No active collection"}
	}

	result := deckimport.DetectAndImport(input)
	if result.Error != "" {
		return &result
	}
	if len(result.Cards) == 0 {
		result.Error = "No cards found in the input"
		return &result
	}

	var deckIdx int = -1
	for i, d := range a.decks {
		if d.Slug == slug {
			deckIdx = i
			break
		}
	}
	if deckIdx == -1 {
		result.Error = "Deck not found"
		return &result
	}

	d := &a.decks[deckIdx]
	deckPath := filepath.Join(a.config.DecksDir(), slug, "deck.txt")

	if mode == "replace" {
		d.Cards = nil
		for _, ic := range result.Cards {
			d.Cards = deck.AddCard(d.Cards, ic.Name, ic.Quantity)
		}
	} else {
		for _, ic := range result.Cards {
			d.Cards = deck.AddCard(d.Cards, ic.Name, ic.Quantity)
		}
	}

	if err := deck.WriteDeckFile(deckPath, d.Cards); err != nil {
		result.Error = fmt.Sprintf("Failed to write deck file: %v", err)
		return &result
	}

	d.CardCount = 0
	for _, c := range d.Cards {
		d.CardCount += c.Quantity
	}

	return &result
}

// ExportDeck returns the deck as formatted text lines.
func (a *App) ExportDeck(slug string) string {
	for i := range a.decks {
		if a.decks[i].Slug == slug {
			d := &a.decks[i]
			var lines []string
			for _, c := range d.Cards {
				lines = append(lines, deck.FormatCardLine(c))
			}
			return strings.Join(lines, "\n")
		}
	}
	return ""
}

// CreateDeckFromImport creates a new deck folder from an import.
func (a *App) CreateDeckFromImport(title string, input string) string {
	if a.config == nil || !a.config.HasActiveCollection() {
		return "error:No active collection"
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return "error:Deck title is required"
	}

	slug := generateSlug(title)
	if slug == "" {
		return "error:Could not generate a valid folder name from the title"
	}

	deckDir := filepath.Join(a.config.DecksDir(), slug)

	if _, err := os.Stat(deckDir); err == nil {
		return "error:A deck with that name already exists"
	}

	result := deckimport.DetectAndImport(input)
	if result.Error != "" {
		return "error:" + result.Error
	}
	if len(result.Cards) == 0 {
		return "error:No cards found in the input"
	}

	if err := os.MkdirAll(deckDir, 0755); err != nil {
		return fmt.Sprintf("error:Failed to create deck folder: %v", err)
	}

	var cards []deck.Card
	for _, ic := range result.Cards {
		cards = deck.AddCard(cards, ic.Name, ic.Quantity)
	}

	deckPath := filepath.Join(deckDir, "deck.txt")
	if err := deck.WriteDeckFile(deckPath, cards); err != nil {
		return fmt.Sprintf("error:Failed to write deck file: %v", err)
	}

	displayTitle := title

	info := deck.DeckInfo{
		Title:  displayTitle,
		Status: "📋 Planned",
	}
	infoPath := filepath.Join(deckDir, "info.md")
	if err := deck.WriteInfoFile(infoPath, info); err != nil {
		log.Printf("[Import] Warning: failed to write info.md: %v", err)
	}

	a.loadDecks()

	return slug
}
