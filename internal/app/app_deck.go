package app

import "app/internal/deck"

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

// SyncResult is returned by GetDeck to include not-found card names.
type SyncResult struct {
	Deck     *deck.Deck `json:"deck"`
	NotFound []string   `json:"notFound"`
}

// GetDeck returns the full deck data for a given slug with Scryfall sync.
func (a *App) GetDeck(slug string) *SyncResult {
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

	notFound := a.syncDeckCards(d.Cards, slug)

	return &SyncResult{
		Deck:     d,
		NotFound: notFound,
	}
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
