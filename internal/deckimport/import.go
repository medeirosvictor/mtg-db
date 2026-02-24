package deckimport

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// ImportedCard represents a card parsed from an import source.
type ImportedCard struct {
	Quantity int    `json:"quantity"`
	Name     string `json:"name"`
}

// ImportResult is the result of an import operation.
type ImportResult struct {
	Cards     []ImportedCard `json:"cards"`
	DeckName  string         `json:"deckName"`
	Source    string         `json:"source"` // "moxfield", "archidekt", "text"
	Error     string         `json:"error,omitempty"`
}

// --- Moxfield ---

// Moxfield public API: GET https://api2.moxfield.com/v3/decks/all/<deckId>
// The deck ID is the last path segment of a moxfield.com/decks/<id> URL.

var moxfieldURLRegex = regexp.MustCompile(`(?:https?://)?(?:www\.)?moxfield\.com/decks/([A-Za-z0-9_-]+)`)

type moxfieldDeck struct {
	Name       string                       `json:"name"`
	Mainboard  map[string]moxfieldCardEntry  `json:"mainboard"`
	Commanders map[string]moxfieldCardEntry  `json:"commanders"`
	Sideboard  map[string]moxfieldCardEntry  `json:"sideboard"`
}

type moxfieldCardEntry struct {
	Quantity int          `json:"quantity"`
	Card     moxfieldCard `json:"card"`
}

type moxfieldCard struct {
	Name string `json:"name"`
}

func ImportFromMoxfield(url string) ImportResult {
	matches := moxfieldURLRegex.FindStringSubmatch(url)
	if matches == nil || len(matches) < 2 {
		return ImportResult{Error: "Invalid Moxfield URL. Expected: moxfield.com/decks/<id>"}
	}
	deckID := matches[1]

	apiURL := fmt.Sprintf("https://api2.moxfield.com/v3/decks/all/%s", deckID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return ImportResult{Error: fmt.Sprintf("Failed to create request: %v", err)}
	}
	req.Header.Set("User-Agent", "mtg-db/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return ImportResult{Error: fmt.Sprintf("Failed to fetch from Moxfield: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ImportResult{Error: fmt.Sprintf("Moxfield API error (%d): %s", resp.StatusCode, string(body))}
	}

	var deck moxfieldDeck
	if err := json.NewDecoder(resp.Body).Decode(&deck); err != nil {
		return ImportResult{Error: fmt.Sprintf("Failed to parse Moxfield response: %v", err)}
	}

	var cards []ImportedCard

	// Commanders first
	for _, entry := range deck.Commanders {
		cards = append(cards, ImportedCard{Quantity: entry.Quantity, Name: entry.Card.Name})
	}
	// Mainboard
	for _, entry := range deck.Mainboard {
		cards = append(cards, ImportedCard{Quantity: entry.Quantity, Name: entry.Card.Name})
	}
	// Sideboard
	for _, entry := range deck.Sideboard {
		cards = append(cards, ImportedCard{Quantity: entry.Quantity, Name: entry.Card.Name})
	}

	return ImportResult{
		Cards:    cards,
		DeckName: deck.Name,
		Source:   "moxfield",
	}
}

// --- Archidekt ---

// Archidekt public API: GET https://archidekt.com/api/decks/<id>/
// The deck ID is a numeric value from archidekt.com/decks/<id>/...

var archidektURLRegex = regexp.MustCompile(`(?:https?://)?(?:www\.)?archidekt\.com/decks/(\d+)`)

type archidektDeck struct {
	Name  string            `json:"name"`
	Cards []archidektEntry  `json:"cards"`
}

type archidektEntry struct {
	Quantity int           `json:"quantity"`
	Card     archidektCard `json:"card"`
}

type archidektCard struct {
	OracleCard archidektOracle `json:"oracleCard"`
}

type archidektOracle struct {
	Name string `json:"name"`
}

func ImportFromArchidekt(url string) ImportResult {
	matches := archidektURLRegex.FindStringSubmatch(url)
	if matches == nil || len(matches) < 2 {
		return ImportResult{Error: "Invalid Archidekt URL. Expected: archidekt.com/decks/<id>"}
	}
	deckID := matches[1]

	apiURL := fmt.Sprintf("https://archidekt.com/api/decks/%s/", deckID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return ImportResult{Error: fmt.Sprintf("Failed to create request: %v", err)}
	}
	req.Header.Set("User-Agent", "mtg-db/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return ImportResult{Error: fmt.Sprintf("Failed to fetch from Archidekt: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ImportResult{Error: fmt.Sprintf("Archidekt API error (%d): %s", resp.StatusCode, string(body))}
	}

	var deck archidektDeck
	if err := json.NewDecoder(resp.Body).Decode(&deck); err != nil {
		return ImportResult{Error: fmt.Sprintf("Failed to parse Archidekt response: %v", err)}
	}

	var cards []ImportedCard
	for _, entry := range deck.Cards {
		name := entry.Card.OracleCard.Name
		if name == "" {
			continue
		}
		cards = append(cards, ImportedCard{Quantity: entry.Quantity, Name: name})
	}

	return ImportResult{
		Cards:    cards,
		DeckName: deck.Name,
		Source:   "archidekt",
	}
}

// --- Text import ---

func ImportFromText(text string) ImportResult {
	lines := strings.Split(text, "\n")
	var cards []ImportedCard

	// Simple regex: optional qty (with optional x), then card name
	lineRegex := regexp.MustCompile(`^\s*(\d+)x?\s+(.+?)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		matches := lineRegex.FindStringSubmatch(line)
		if matches != nil {
			qty := 1
			fmt.Sscanf(matches[1], "%d", &qty)
			name := strings.TrimSpace(matches[2])
			// Strip set code and collector number if present: "Card Name (SET) 123"
			if idx := strings.Index(name, "("); idx > 0 {
				name = strings.TrimSpace(name[:idx])
			}
			// Strip foil marker
			name = strings.TrimSuffix(name, "*F*")
			name = strings.TrimSpace(name)
			// Strip tags
			tagIdx := strings.Index(name, "#")
			if tagIdx > 0 {
				name = strings.TrimSpace(name[:tagIdx])
			}
			if name != "" {
				cards = append(cards, ImportedCard{Quantity: qty, Name: name})
			}
		} else {
			// Treat as card name with qty 1
			name := line
			if idx := strings.Index(name, "("); idx > 0 {
				name = strings.TrimSpace(name[:idx])
			}
			name = strings.TrimSuffix(name, "*F*")
			name = strings.TrimSpace(name)
			tagIdx := strings.Index(name, "#")
			if tagIdx > 0 {
				name = strings.TrimSpace(name[:tagIdx])
			}
			if name != "" {
				cards = append(cards, ImportedCard{Quantity: 1, Name: name})
			}
		}
	}

	return ImportResult{
		Cards:  cards,
		Source: "text",
	}
}

// DetectAndImport auto-detects the input type and imports accordingly.
func DetectAndImport(input string) ImportResult {
	input = strings.TrimSpace(input)

	if moxfieldURLRegex.MatchString(input) {
		return ImportFromMoxfield(input)
	}
	if archidektURLRegex.MatchString(input) {
		return ImportFromArchidekt(input)
	}

	return ImportFromText(input)
}
