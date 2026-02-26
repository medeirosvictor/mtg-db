package app

import (
	"app/internal/deck"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReloadDecks re-reads all deck files from disk.
func (a *App) ReloadDecks() []DeckSummary {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.loadDecksUnsafe()
	return a.getAllDecksUnsafe()
}

// ToggleCardTag toggles a tag on a card in a deck and writes to disk.
func (a *App) ToggleCardTag(slug string, cardName string, tag string) *deck.Deck {
	if a.config == nil || !a.config.HasActiveCollection() {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	deckDir := filepath.Join(a.config.DecksDir(), slug)
	deckPath := filepath.Join(deckDir, "deck.txt")

	deckIdx := a.getDeckIndexUnsafe(slug)
	if deckIdx == -1 {
		return nil
	}

	d := &a.decks[deckIdx]

	found := false
	for i := range d.Cards {
		if strings.EqualFold(d.Cards[i].Name, cardName) {
			found = true
			if d.Cards[i].HasTag(tag) {
				var newTags []string
				for _, t := range d.Cards[i].Tags {
					if t != tag {
						newTags = append(newTags, t)
					}
				}
				d.Cards[i].Tags = newTags
			} else {
				if tag == deck.TagCommander {
					commanders := deck.GetCommanders(d.Cards)
					if len(commanders) >= 2 {
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

	if err := deck.WriteDeckFile(deckPath, d.Cards); err != nil {
		fmt.Fprintf(os.Stderr, "error writing deck file: %v\n", err)
		return nil
	}

	commanders := deck.GetCommanders(d.Cards)
	if len(commanders) > 0 {
		d.Info.Commander = strings.Join(commanders, " / ")
	} else {
		info, err := deck.ParseInfoFile(filepath.Join(deckDir, "info.md"))
		if err == nil {
			d.Info.Commander = info.Commander
		}
	}

	d.CardCount = 0
	for _, c := range d.Cards {
		d.CardCount += c.Quantity
	}

	return d
}

// AddCards adds one or more cards to a deck.
func (a *App) AddCards(slug string, cardLines string) string {
	if a.config == nil || !a.config.HasActiveCollection() {
		return "No active collection"
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	deckDir := filepath.Join(a.config.DecksDir(), slug)
	deckPath := filepath.Join(deckDir, "deck.txt")

	deckIdx := a.getDeckIndexUnsafe(slug)
	if deckIdx == -1 {
		return "Deck not found"
	}

	d := &a.decks[deckIdx]

	lines := strings.Split(cardLines, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		card, err := deck.ParseCardLine(line)
		if err != nil {
			continue
		}

		d.Cards = deck.AddCard(d.Cards, card.Name, card.Quantity)
	}

	if err := deck.WriteDeckFile(deckPath, d.Cards); err != nil {
		return fmt.Sprintf("Failed to write deck file: %v", err)
	}

	d.CardCount = 0
	for _, c := range d.Cards {
		d.CardCount += c.Quantity
	}

	return ""
}

// UpdateCardQuantity updates the quantity of a card in a deck.
func (a *App) UpdateCardQuantity(slug string, cardName string, quantity int) string {
	if a.config == nil || !a.config.HasActiveCollection() {
		return "No active collection"
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	deckDir := filepath.Join(a.config.DecksDir(), slug)
	deckPath := filepath.Join(deckDir, "deck.txt")

	deckIdx := a.getDeckIndexUnsafe(slug)
	if deckIdx == -1 {
		return "Deck not found"
	}

	d := &a.decks[deckIdx]
	d.Cards = deck.UpdateCardQty(d.Cards, cardName, quantity)

	if err := deck.WriteDeckFile(deckPath, d.Cards); err != nil {
		return fmt.Sprintf("Failed to write deck file: %v", err)
	}

	d.CardCount = 0
	for _, c := range d.Cards {
		d.CardCount += c.Quantity
	}

	return ""
}

// RemoveCard removes a card from a deck by name.
func (a *App) RemoveCard(slug string, cardName string) string {
	return a.UpdateCardQuantity(slug, cardName, 0)
}

// UpdateCardText replaces a card's raw line in deck.txt.
func (a *App) UpdateCardText(slug string, oldName string, newLine string) string {
	if a.config == nil || !a.config.HasActiveCollection() {
		return "No active collection"
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	deckDir := filepath.Join(a.config.DecksDir(), slug)
	deckPath := filepath.Join(deckDir, "deck.txt")

	deckIdx := a.getDeckIndexUnsafe(slug)
	if deckIdx == -1 {
		return "Deck not found"
	}

	d := &a.decks[deckIdx]

	newCard, err := deck.ParseCardLine(newLine)
	if err != nil {
		return fmt.Sprintf("Invalid card line: %v", err)
	}

	found := false
	for i := range d.Cards {
		if strings.EqualFold(d.Cards[i].Name, oldName) {
			if strings.EqualFold(d.Cards[i].Name, newCard.Name) {
				newCard.Scryfall = d.Cards[i].Scryfall
			}
			d.Cards[i] = newCard
			found = true
			break
		}
	}
	if !found {
		return fmt.Sprintf("Card %q not found in deck", oldName)
	}

	if err := deck.WriteDeckFile(deckPath, d.Cards); err != nil {
		return fmt.Sprintf("Failed to write deck file: %v", err)
	}

	d.CardCount = 0
	for _, c := range d.Cards {
		d.CardCount += c.Quantity
	}

	commanders := deck.GetCommanders(d.Cards)
	if len(commanders) > 0 {
		d.Info.Commander = strings.Join(commanders, " / ")
	}

	return ""
}

// Helper functions that assume lock is already held
func (a *App) loadDecksUnsafe() {
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

func (a *App) getDeckIndexUnsafe(slug string) int {
	for i := range a.decks {
		if a.decks[i].Slug == slug {
			return i
		}
	}
	return -1
}

func (a *App) getAllDecksUnsafe() []DeckSummary {
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
