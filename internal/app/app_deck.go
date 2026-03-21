package app

import "app/internal/deck"

// DeckSummary is a lightweight view of a deck for the dashboard.
type DeckSummary struct {
	Slug              string `json:"slug"`
	Title             string `json:"title"`
	Commander         string `json:"commander"`
	CommanderImageUri string `json:"commanderImageUri,omitempty"`
	Colors            string `json:"colors"`
	Status            string `json:"status"`
	CardCount         int    `json:"cardCount"`
	Universe          string `json:"universe,omitempty"`
}

// GetAllDecks returns summaries of all loaded decks (thread-safe).
func (a *App) GetAllDecks() []DeckSummary {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.getAllDecksUnsafe()
}

// SyncResult is returned by GetDeck to include not-found card names.
type SyncResult struct {
	Deck     *deck.Deck `json:"deck"`
	NotFound []string   `json:"notFound"`
}

// GetDeck returns the full deck data for a given slug with Scryfall sync (thread-safe).
func (a *App) GetDeck(slug string) *SyncResult {
	a.mu.RLock()
	d := a.getDeckUnsafe(slug)
	a.mu.RUnlock()

	if d == nil {
		return nil
	}

	notFound := a.syncDeckCards(d.Cards, slug)

	return &SyncResult{
		Deck:     d,
		NotFound: notFound,
	}
}

// GetDeckBasic returns deck data without Scryfall sync (faster, thread-safe).
func (a *App) GetDeckBasic(slug string) *deck.Deck {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.getDeckUnsafe(slug)
}

// getDeckUnsafe returns a pointer to a deck by slug (caller must hold read lock)
func (a *App) getDeckUnsafe(slug string) *deck.Deck {
	for i := range a.decks {
		if a.decks[i].Slug == slug {
			return &a.decks[i]
		}
	}
	return nil
}
