package deck

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Card represents a single card entry parsed from a deck file.
type Card struct {
	Quantity        int           `json:"quantity"`
	Name            string        `json:"name"`
	SetCode         string        `json:"setCode,omitempty"`
	CollectorNumber string        `json:"collectorNumber,omitempty"`
	Foil            bool          `json:"foil,omitempty"`
	Tags            []string      `json:"tags,omitempty"`
	Scryfall        *ScryfallData `json:"scryFall,omitempty"`
}

// ScryfallData holds additional card data from the Scryfall API
type ScryfallData struct {
	OracleText    string   `json:"oracleText,omitempty"`
	TypeLine      string   `json:"typeLine,omitempty"`
	ManaCost      string   `json:"manaCost,omitempty"`
	CMC           float64  `json:"cmc,omitempty"`
	ImageURI      string   `json:"imageUri,omitempty"`
	BackImageURI  string   `json:"backImageUri,omitempty"`
	IsDoubleFaced bool     `json:"isDoubleFaced,omitempty"`
	PriceUSD      string   `json:"priceUsd,omitempty"`
	PriceUSDFoil  string   `json:"priceUsdFoil,omitempty"`
	ColorIdentity string   `json:"colorIdentity,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

// Tag constants.
const (
	TagCommander = "commander"
	TagProxy     = "proxy"
	TagWishlist  = "wishlist"
	TagSideboard = "sideboard"
)

// HasTag returns true if the card has the given tag.
func (c Card) HasTag(tag string) bool {
	for _, t := range c.Tags {
		if t == tag {
			return true
		}
	}
	return false
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

// tagRegex matches #tag tokens at the end of a line.
var tagRegex = regexp.MustCompile(`\s+#([a-zA-Z][a-zA-Z0-9_-]*)`)

// extractTags strips #tag tokens from a line and returns the cleaned line + tags.
func extractTags(line string) (string, []string) {
	var tags []string
	matches := tagRegex.FindAllStringSubmatch(line, -1)
	for _, m := range matches {
		tags = append(tags, strings.ToLower(m[1]))
	}
	cleaned := tagRegex.ReplaceAllString(line, "")
	return strings.TrimSpace(cleaned), tags
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

	// Extract tags before parsing the rest
	cleaned, tags := extractTags(line)

	matches := cardLineRegex.FindStringSubmatch(cleaned)
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
	foil := strings.Contains(cleaned, "*F*")

	return Card{
		Quantity:        qty,
		Name:            name,
		SetCode:         setCode,
		CollectorNumber: collectorNum,
		Foil:            foil,
		Tags:            tags,
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

// WriteInfoFile writes the deck metadata back to info.md.
func WriteInfoFile(path string, info DeckInfo) error {
	var lines []string
	if info.Title != "" {
		lines = append(lines, "# "+info.Title)
		lines = append(lines, "")
	}
	if info.Status != "" {
		lines = append(lines, "- **Status:** "+info.Status)
	}
	if info.Colors != "" {
		lines = append(lines, "- **Colors:** "+info.Colors)
	}
	if info.Commander != "" {
		lines = append(lines, "- **Commander:** "+info.Commander)
	}
	if info.Strategy != "" {
		lines = append(lines, "- **Strategy:** "+info.Strategy)
	}
	if info.Universe != "" {
		lines = append(lines, "- **Universe:** "+info.Universe)
	}
	lines = append(lines, "") // trailing newline

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

// FormatCardLine formats a Card back to a text line for writing to deck.txt.
func FormatCardLine(c Card) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%d", c.Quantity))

	// Use "1x" format for consistency
	sb.WriteString("x ")
	sb.WriteString(c.Name)

	if c.SetCode != "" {
		sb.WriteString(" (")
		sb.WriteString(c.SetCode)
		sb.WriteString(")")
		if c.CollectorNumber != "" {
			sb.WriteString(" ")
			sb.WriteString(c.CollectorNumber)
		}
	}

	if c.Foil {
		sb.WriteString(" *F*")
	}

	// Append tags
	for _, tag := range c.Tags {
		sb.WriteString(" #")
		sb.WriteString(tag)
	}

	return sb.String()
}

// WriteDeckFile writes all cards to a deck.txt file.
func WriteDeckFile(path string, cards []Card) error {
	var lines []string
	for _, c := range cards {
		lines = append(lines, FormatCardLine(c))
	}
	// Trailing newline
	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0o644)
}

// TODO: Phase 3A - Granular tag control (currently ToggleCardTag in app_cardops.go does inline toggling)
// SetCardTag adds or removes a tag on a card in a deck file.
// Returns the updated list of cards.
func SetCardTag(cards []Card, cardName string, tag string, enabled bool) []Card {
	for i := range cards {
		if strings.EqualFold(cards[i].Name, cardName) {
			if enabled {
				if !cards[i].HasTag(tag) {
					cards[i].Tags = append(cards[i].Tags, tag)
				}
			} else {
				var newTags []string
				for _, t := range cards[i].Tags {
					if t != tag {
						newTags = append(newTags, t)
					}
				}
				cards[i].Tags = newTags
			}
		}
	}
	return cards
}

// AddCard adds a card to the deck. If card already exists, increments quantity.
// Returns the updated list of cards.
func AddCard(cards []Card, name string, quantity int) []Card {
	if quantity <= 0 {
		return cards
	}

	// Check if card already exists (case-insensitive)
	for i := range cards {
		if strings.EqualFold(cards[i].Name, name) {
			cards[i].Quantity += quantity
			return cards
		}
	}

	// Card doesn't exist, add new
	newCard := Card{
		Quantity: quantity,
		Name:     name,
	}
	return append(cards, newCard)
}

// RemoveCard removes a card from the deck by name (case-insensitive).
// Returns the updated list of cards.
func RemoveCard(cards []Card, name string) []Card {
	var result []Card
	for _, c := range cards {
		if !strings.EqualFold(c.Name, name) {
			result = append(result, c)
		}
	}
	return result
}

// UpdateCardQty updates the quantity of a card. If quantity <= 0, removes the card.
// Returns the updated list of cards.
func UpdateCardQty(cards []Card, name string, quantity int) []Card {
	if quantity <= 0 {
		return RemoveCard(cards, name)
	}

	for i := range cards {
		if strings.EqualFold(cards[i].Name, name) {
			cards[i].Quantity = quantity
			return cards
		}
	}

	// Card doesn't exist, add new
	newCard := Card{
		Quantity: quantity,
		Name:     name,
	}
	return append(cards, newCard)
}

// GetCommanders returns the names of cards tagged as commander in a deck.
func GetCommanders(cards []Card) []string {
	var commanders []string
	for _, c := range cards {
		if c.HasTag(TagCommander) {
			commanders = append(commanders, c.Name)
		}
	}
	return commanders
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

	// Derive commander from #commander tags if present, otherwise use info.md
	commanders := GetCommanders(cards)
	if len(commanders) > 0 {
		sort.Strings(commanders)
		info.Commander = strings.Join(commanders, " / ")
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

// TODO: Phase 2D - Server-side deck validation
// FilterMainDeck returns cards that are NOT in the sideboard.
func FilterMainDeck(cards []Card) []Card {
	var result []Card
	for _, c := range cards {
		if !c.HasTag(TagSideboard) {
			result = append(result, c)
		}
	}
	return result
}

// TODO: Phase 2D - Server-side deck validation
// FilterSideboard returns only cards that are in the sideboard.
func FilterSideboard(cards []Card) []Card {
	var result []Card
	for _, c := range cards {
		if c.HasTag(TagSideboard) {
			result = append(result, c)
		}
	}
	return result
}

// DeckValidationResult holds the result of validating a deck.
type DeckValidationResult struct {
	IsValid       bool     `json:"isValid"`
	Errors        []string `json:"errors"`
	WarningCount int      `json:"warningCount"`
}

// TODO: Phase 2D - Server-side deck validation
// ValidateDeck checks if a deck is valid for Commander format.
// It validates that the main deck (non-sideboard) has exactly 100 cards.
// Commander cards COUNT toward the 100 card limit.
func ValidateDeck(cards []Card) DeckValidationResult {
	mainDeck := FilterMainDeck(cards)

	// Count ALL non-sideboard cards (including commanders - they count toward 100)
	totalCount := 0
	for _, c := range mainDeck {
		totalCount += c.Quantity
	}

	var errors []string

	if totalCount < 100 {
		errors = append(errors, fmt.Sprintf("Deck must have exactly 100 cards (currently %d)", totalCount))
	} else if totalCount > 100 {
		errors = append(errors, fmt.Sprintf("Deck must have exactly 100 cards (currently %d)", totalCount))
	}

	return DeckValidationResult{
		IsValid:       len(errors) == 0,
		Errors:        errors,
		WarningCount:  0,
	}
}

// TODO: Phase 2D - Server-side deck validation
// GetNonSideboardCount returns the total count of non-sideboard cards.
func GetNonSideboardCount(cards []Card) int {
	mainDeck := FilterMainDeck(cards)
	count := 0
	for _, c := range mainDeck {
		count += c.Quantity
	}
	return count
}
