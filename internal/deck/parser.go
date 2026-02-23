package deck

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Card represents a single card entry parsed from a deck file.
type Card struct {
	Quantity        int    `json:"quantity"`
	Name            string `json:"name"`
	SetCode         string `json:"setCode,omitempty"`
	CollectorNumber string `json:"collectorNumber,omitempty"`
	Foil            bool   `json:"foil,omitempty"`
}

// DeckInfo represents metadata parsed from info.md.
type DeckInfo struct {
	Title     string `json:"title"`
	Status    string `json:"status"`
	Colors    string `json:"colors"`
	Commander string `json:"commander"`
	Strategy  string `json:"strategy"`
	Universe  string `json:"universe,omitempty"`
}

// Deck represents a full deck with its cards, info, and wishlist.
type Deck struct {
	Slug      string   `json:"slug"`
	Info      DeckInfo `json:"info"`
	Cards     []Card   `json:"cards"`
	Wishlist  []Card   `json:"wishlist"`
	CardCount int      `json:"cardCount"`
}

// cardLineRegex handles all known formats:
//
//	1 Card Name
//	1x Card Name
//	1x Card Name (SET) 123
//	1x Card Name (SET) 123 *F*
//	1x Card Name (SET) 23s *F*   (non-numeric collector numbers)
//	1 Card Name (PLST) CMA-30    (hyphenated collector numbers)
//	1 Card A / Card B (SET) 123
//
// Also handles {collector_num} variant from some files.
var cardLineRegex = regexp.MustCompile(
	`^(\d+)x?\s+(.+?)` + // qty + name (non-greedy)
		`(?:\s+\(([A-Za-z0-9]+)\)\s*([A-Za-z0-9-]+)?)?` + // optional (SET) collector#
		`(?:\s+\{(\d+)\})?` + // optional {num} variant
		`(?:\s+\*F\*)?` + // optional foil marker
		`\s*$`, // trailing whitespace
)

// ParseCardLine parses a single line from a deck or wishlist file.
func ParseCardLine(line string) (Card, error) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
		return Card{}, fmt.Errorf("skip: empty or comment")
	}

	matches := cardLineRegex.FindStringSubmatch(line)
	if matches == nil {
		return Card{}, fmt.Errorf("could not parse line: %q", line)
	}

	qty, err := strconv.Atoi(matches[1])
	if err != nil {
		return Card{}, fmt.Errorf("invalid quantity in %q: %w", line, err)
	}

	name := strings.TrimSpace(matches[2])
	setCode := strings.ToUpper(strings.TrimSpace(matches[3]))
	collectorNum := strings.TrimSpace(matches[4])
	if collectorNum == "" {
		collectorNum = strings.TrimSpace(matches[5]) // {num} variant
	}
	foil := strings.Contains(line, "*F*")

	return Card{
		Quantity:        qty,
		Name:            name,
		SetCode:         setCode,
		CollectorNumber: collectorNum,
		Foil:            foil,
	}, nil
}

// ParseDeckFile reads a deck.txt or wishlist.txt and returns all parsed cards.
func ParseDeckFile(path string) ([]Card, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var cards []Card
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		card, err := ParseCardLine(scanner.Text())
		if err != nil {
			continue // skip blanks, comments, unparseable lines
		}
		cards = append(cards, card)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}
	return cards, nil
}

// infoLineRegex parses "- **Key:** Value" lines from info.md.
var infoLineRegex = regexp.MustCompile(`^-\s+\*\*(.+?):\*\*\s*(.+)$`)

// ParseInfoFile reads an info.md and returns deck metadata.
func ParseInfoFile(path string) (DeckInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return DeckInfo{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	info := DeckInfo{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Title from the # heading
		if strings.HasPrefix(line, "# ") {
			info.Title = strings.TrimPrefix(line, "# ")
			continue
		}

		matches := infoLineRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		key := strings.TrimSpace(matches[1])
		value := strings.TrimSpace(matches[2])

		switch strings.ToLower(key) {
		case "status":
			info.Status = value
		case "colors":
			info.Colors = value
		case "commander":
			info.Commander = value
		case "strategy":
			info.Strategy = value
		case "universe":
			info.Universe = value
		}
	}
	if err := scanner.Err(); err != nil {
		return DeckInfo{}, fmt.Errorf("scan %s: %w", path, err)
	}
	return info, nil
}

// LoadDeck loads a complete deck from a directory (deck.txt, info.md, wishlist.txt).
func LoadDeck(dir string) (Deck, error) {
	slug := filepath.Base(dir)

	// Parse info.md
	info, err := ParseInfoFile(filepath.Join(dir, "info.md"))
	if err != nil {
		info = DeckInfo{Title: slug} // fallback: use folder name
	}

	// Parse deck.txt
	cards, err := ParseDeckFile(filepath.Join(dir, "deck.txt"))
	if err != nil {
		return Deck{}, fmt.Errorf("load deck %s: %w", slug, err)
	}

	// Count total cards
	cardCount := 0
	for _, c := range cards {
		cardCount += c.Quantity
	}

	// Parse wishlist.txt (optional)
	wishlist, _ := ParseDeckFile(filepath.Join(dir, "wishlist.txt"))

	return Deck{
		Slug:      slug,
		Info:      info,
		Cards:     cards,
		Wishlist:  wishlist,
		CardCount: cardCount,
	}, nil
}

// LoadAllDecks scans a directory for deck subdirectories and loads them all.
func LoadAllDecks(decksDir string) ([]Deck, error) {
	entries, err := os.ReadDir(decksDir)
	if err != nil {
		return nil, fmt.Errorf("read decks dir %s: %w", decksDir, err)
	}

	var decks []Deck
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		deckDir := filepath.Join(decksDir, entry.Name())
		// Only load if deck.txt exists
		if _, err := os.Stat(filepath.Join(deckDir, "deck.txt")); err != nil {
			continue
		}
		deck, err := LoadDeck(deckDir)
		if err != nil {
			continue // skip broken decks
		}
		decks = append(decks, deck)
	}
	return decks, nil
}
