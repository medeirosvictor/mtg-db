package deck

import (
	"fmt"
	"testing"
)

// =====================
// Sideboard Tag Tests
// =====================

func TestTagSideboard_Constant(t *testing.T) {
	// Sideboard tag constant should be defined
	if TagSideboard != "sideboard" {
		t.Errorf("TagSideboard = %q, want %q", TagSideboard, "sideboard")
	}
}

func TestCard_HasTag_Sideboard(t *testing.T) {
	card := Card{
		Name:  "Lightning Bolt",
		Tags:  []string{"sideboard"},
	}

	if !card.HasTag("sideboard") {
		t.Error("Card with sideboard tag should have sideboard tag")
	}
}

func TestCard_HasTag_Sideboard_NotPresent(t *testing.T) {
	card := Card{
		Name:  "Lightning Bolt",
		Tags:  []string{"commander"},
	}

	if card.HasTag("sideboard") {
		t.Error("Card without sideboard tag should NOT have sideboard tag")
	}
}

func TestExtractTags_Sideboard(t *testing.T) {
	line := "1x Lightning Bolt #sideboard"
	cleaned, tags := extractTags(line)

	if cleaned != "1x Lightning Bolt" {
		t.Errorf("cleaned = %q, want %q", cleaned, "1x Lightning Bolt")
	}

	if len(tags) != 1 || tags[0] != "sideboard" {
		t.Errorf("tags = %v, want [sideboard]", tags)
	}
}

func TestExtractTags_MultipleTagsIncludingSideboard(t *testing.T) {
	line := "1x Sol Ring #commander #sideboard"
	cleaned, tags := extractTags(line)

	if cleaned != "1x Sol Ring" {
		t.Errorf("cleaned = %q, want %q", cleaned, "1x Sol Ring")
	}

	if len(tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(tags))
	}

	hasCommander := false
	hasSideboard := false
	for _, tag := range tags {
		if tag == "commander" {
			hasCommander = true
		}
		if tag == "sideboard" {
			hasSideboard = true
		}
	}

	if !hasCommander {
		t.Error("should have commander tag")
	}
	if !hasSideboard {
		t.Error("should have sideboard tag")
	}
}

// =====================
// Sideboard Filtering Tests
// =====================

func TestFilterSideboard_Cards(t *testing.T) {
	cards := []Card{
		{Name: "Lightning Bolt", Tags: []string{}},
		{Name: "Sol Ring", Tags: []string{"sideboard"}},
		{Name: "Counterspell", Tags: []string{"sideboard"}},
		{Name: "Forest", Tags: []string{}},
	}

	mainDeck := FilterMainDeck(cards)
	sideboard := FilterSideboard(cards)

	if len(mainDeck) != 2 {
		t.Errorf("main deck length = %d, want 2", len(mainDeck))
	}

	if len(sideboard) != 2 {
		t.Errorf("sideboard length = %d, want 2", len(sideboard))
	}

	// Verify main deck contains correct cards
	mainNames := make(map[string]bool)
	for _, c := range mainDeck {
		mainNames[c.Name] = true
	}
	if !mainNames["Lightning Bolt"] {
		t.Error("main deck should contain Lightning Bolt")
	}
	if !mainNames["Forest"] {
		t.Error("main deck should contain Forest")
	}

	// Verify sideboard contains correct cards
	sideNames := make(map[string]bool)
	for _, c := range sideboard {
		sideNames[c.Name] = true
	}
	if !sideNames["Sol Ring"] {
		t.Error("sideboard should contain Sol Ring")
	}
	if !sideNames["Counterspell"] {
		t.Error("sideboard should contain Counterspell")
	}
}

func TestFilterSideboard_Empty(t *testing.T) {
	cards := []Card{
		{Name: "Lightning Bolt", Tags: []string{}},
	}

	mainDeck := FilterMainDeck(cards)
	sideboard := FilterSideboard(cards)

	if len(mainDeck) != 1 {
		t.Errorf("main deck length = %d, want 1", len(mainDeck))
	}

	if len(sideboard) != 0 {
		t.Errorf("sideboard length = %d, want 0", len(sideboard))
	}
}

func TestFilterSideboard_AllSideboard(t *testing.T) {
	cards := []Card{
		{Name: "Lightning Bolt", Tags: []string{"sideboard"}},
		{Name: "Counterspell", Tags: []string{"sideboard"}},
	}

	mainDeck := FilterMainDeck(cards)
	sideboard := FilterSideboard(cards)

	if len(mainDeck) != 0 {
		t.Errorf("main deck length = %d, want 0", len(mainDeck))
	}

	if len(sideboard) != 2 {
		t.Errorf("sideboard length = %d, want 2", len(sideboard))
	}
}

// =====================
// Deck Validation Tests
// =====================

func TestValidateDeck_Valid100Cards(t *testing.T) {
	cards := make([]Card, 100)
	for i := 0; i < 100; i++ {
		cards[i] = Card{Name: "Card", Quantity: 1}
	}

	result := ValidateDeck(cards)

	if !result.IsValid {
		t.Error("Deck with exactly 100 non-sideboard cards should be valid")
	}
}

func TestValidateDeck_Valid99Cards(t *testing.T) {
	cards := make([]Card, 99)
	for i := 0; i < 99; i++ {
		cards[i] = Card{Name: "Card", Quantity: 1}
	}

	result := ValidateDeck(cards)

	if result.IsValid {
		t.Error("Deck with 99 non-sideboard cards should NOT be valid")
	}

	if len(result.Errors) == 0 {
		t.Error("Should have at least one error")
	}
}

func TestValidateDeck_Valid101Cards(t *testing.T) {
	cards := make([]Card, 101)
	for i := 0; i < 101; i++ {
		cards[i] = Card{Name: "Card", Quantity: 1}
	}

	result := ValidateDeck(cards)

	if result.IsValid {
		t.Error("Deck with 101 non-sideboard cards should NOT be valid")
	}
}

func TestValidateDeck_SideboardDoesNotCount(t *testing.T) {
	// 100 main deck + 15 sideboard = valid
	cards := make([]Card, 115)
	for i := 0; i < 100; i++ {
		cards[i] = Card{Name: "Main Card", Quantity: 1}
	}
	for i := 100; i < 115; i++ {
		cards[i] = Card{Name: "Sideboard Card", Tags: []string{"sideboard"}, Quantity: 1}
	}

	result := ValidateDeck(cards)

	if !result.IsValid {
		t.Error("Deck with 100 main + 15 sideboard should be valid")
	}
}

func TestValidateDeck_CommanderWithSideboard(t *testing.T) {
	// 99 regular + 1 commander + 1 sideboard = 100 main (valid)
	// Commander DOES count toward 100
	cards := []Card{
		{Name: "Atraxa", Tags: []string{"commander"}, Quantity: 1},
	}
	// 99 regular cards
	for i := 0; i < 99; i++ {
		cards = append(cards, Card{Name: fmt.Sprintf("Card %d", i), Quantity: 1})
	}
	// 1 sideboard card (doesn't count toward 100)
	cards = append(cards, Card{Name: "Sideboard Card", Tags: []string{"sideboard"}, Quantity: 1})

	result := ValidateDeck(cards)

	if !result.IsValid {
		t.Error("Deck with 99 main + 1 commander + 1 sideboard should be valid")
	}
}

func TestValidateDeck_CommanderCountsToward100(t *testing.T) {
	// 100 regular + 1 commander = 101 (invalid)
	cards := []Card{
		{Name: "Atraxa", Tags: []string{"commander"}, Quantity: 1},
	}
	for i := 0; i < 100; i++ {
		cards = append(cards, Card{Name: fmt.Sprintf("Card %d", i), Quantity: 1})
	}

	result := ValidateDeck(cards)

	if result.IsValid {
		t.Error("Deck with 100 regular + 1 commander should be INVALID (101 total)")
	}
	if len(result.Errors) == 0 {
		t.Error("Should have an error message")
	}
}

// =====================
// SetCardTag with Sideboard
// =====================

func TestSetCardTag_AddSideboard(t *testing.T) {
	cards := []Card{
		{Name: "Lightning Bolt", Tags: []string{}},
	}

	result := SetCardTag(cards, "Lightning Bolt", "sideboard", true)

	if len(result) != 1 {
		t.Fatalf("len(result) = %d, want 1", len(result))
	}

	if !result[0].HasTag("sideboard") {
		t.Error("Card should have sideboard tag after adding")
	}
}

func TestSetCardTag_RemoveSideboard(t *testing.T) {
	cards := []Card{
		{Name: "Lightning Bolt", Tags: []string{"sideboard"}},
	}

	result := SetCardTag(cards, "Lightning Bolt", "sideboard", false)

	if len(result) != 1 {
		t.Fatalf("len(result) = %d, want 1", len(result))
	}

	if result[0].HasTag("sideboard") {
		t.Error("Card should NOT have sideboard tag after removing")
	}
}

func TestSetCardTag_ToggleSideboardAndCommander(t *testing.T) {
	// A card can be both commander and sideboard (e.g., partner commanders)
	cards := []Card{
		{Name: "Ludevic, Necro-Alchemist", Tags: []string{"commander"}},
	}

	// Add sideboard tag while keeping commander
	result := SetCardTag(cards, "Ludevic, Necro-Alchemist", "sideboard", true)

	if !result[0].HasTag("commander") {
		t.Error("Card should still have commander tag")
	}
	if !result[0].HasTag("sideboard") {
		t.Error("Card should have sideboard tag added")
	}
}
