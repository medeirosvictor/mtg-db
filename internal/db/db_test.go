package db

import (
	"app/internal/scryfall"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestDB creates a temp directory and initializes a fresh database.
func setupTestDB(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "db-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	// Reset the global db to nil before each test
	if db != nil {
		db.Close()
		db = nil
	}
	if err := Init(tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Init failed: %v", err)
	}
	return tmpDir
}

func cleanupTestDB(t *testing.T, dir string) {
	t.Helper()
	Close()
	os.RemoveAll(dir)
}

func makeCachedCard(name string) *CachedCard {
	return &CachedCard{
		Name:            name,
		OracleID:        "oracle-" + name,
		OracleText:      name + " does something.",
		TypeLine:        "Instant",
		ManaCost:        "{R}",
		CMC:             1.0,
		Colors:          `"R"`,
		ColorIdentity:   `"R"`,
		SetCode:         "M19",
		SetName:         "Core Set 2019",
		CollectorNumber: "123",
		ImageURI:        "https://example.com/" + name + ".jpg",
		BackImageURI:    "",
		IsDoubleFaced:   false,
		PriceUSD:        sql.NullString{String: "1.50", Valid: true},
		PriceUSDFoil:    sql.NullString{String: "3.00", Valid: true},
		PriceEUR:        sql.NullString{String: "1.20", Valid: true},
		PriceEURFoil:    sql.NullString{String: "2.50", Valid: true},
		Legalities:      `{"commander":"legal","standard":"not_legal"}`,
		UpdatedAt:       time.Now(),
	}
}

// =====================
// CRITICAL: Init / Schema Tests
// =====================

func TestInit_CreatesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "db-init-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if db != nil {
		db.Close()
		db = nil
	}

	subDir := filepath.Join(tmpDir, "nested", "path")
	if err := Init(subDir); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	// DB file should exist
	dbPath := filepath.Join(subDir, "cards.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("cards.db was not created")
	}
}

func TestInit_CreatesTablesAndIndexes(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	// Verify cards table exists
	var tableName string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='cards'").Scan(&tableName)
	if err != nil {
		t.Fatalf("cards table not found: %v", err)
	}

	// Verify unmatched_cards table exists
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='unmatched_cards'").Scan(&tableName)
	if err != nil {
		t.Fatalf("unmatched_cards table not found: %v", err)
	}

	// Verify indexes
	var indexName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_cards_set'").Scan(&indexName)
	if err != nil {
		t.Fatalf("idx_cards_set index not found: %v", err)
	}
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_cards_updated'").Scan(&indexName)
	if err != nil {
		t.Fatalf("idx_cards_updated index not found: %v", err)
	}
}

func TestInit_WALMode(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode;").Scan(&journalMode); err != nil {
		t.Fatal(err)
	}
	if journalMode != "wal" {
		t.Errorf("journal_mode = %s, want wal", journalMode)
	}
}

func TestInit_DFCMigrationIdempotent(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	// Running migration again should not fail
	if err := migrateAddDFCColumns(); err != nil {
		t.Fatalf("second migration call failed: %v", err)
	}

	// Verify columns exist by inserting a DFC card
	card := makeCachedCard("Test DFC")
	card.BackImageURI = "https://example.com/back.jpg"
	card.IsDoubleFaced = true
	if err := UpsertCard(card); err != nil {
		t.Fatalf("insert DFC card failed: %v", err)
	}

	got, err := GetCard("Test DFC")
	if err != nil {
		t.Fatal(err)
	}
	if got.BackImageURI != "https://example.com/back.jpg" {
		t.Errorf("BackImageURI = %q, want back.jpg URL", got.BackImageURI)
	}
	if !got.IsDoubleFaced {
		t.Error("IsDoubleFaced should be true")
	}
}

// =====================
// CRITICAL: UpsertCard + GetCard Tests
// =====================

func TestUpsertCard_InsertAndRetrieve(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	card := makeCachedCard("Lightning Bolt")
	if err := UpsertCard(card); err != nil {
		t.Fatalf("UpsertCard failed: %v", err)
	}

	got, err := GetCard("Lightning Bolt")
	if err != nil {
		t.Fatalf("GetCard failed: %v", err)
	}
	if got == nil {
		t.Fatal("GetCard returned nil")
	}

	if got.Name != "Lightning Bolt" {
		t.Errorf("Name = %q, want %q", got.Name, "Lightning Bolt")
	}
	if got.OracleText != "Lightning Bolt does something." {
		t.Errorf("OracleText = %q", got.OracleText)
	}
	if got.TypeLine != "Instant" {
		t.Errorf("TypeLine = %q", got.TypeLine)
	}
	if got.ManaCost != "{R}" {
		t.Errorf("ManaCost = %q", got.ManaCost)
	}
	if got.CMC != 1.0 {
		t.Errorf("CMC = %f", got.CMC)
	}
	if got.SetCode != "M19" {
		t.Errorf("SetCode = %q", got.SetCode)
	}
	if got.ImageURI != "https://example.com/Lightning Bolt.jpg" {
		t.Errorf("ImageURI = %q", got.ImageURI)
	}
	if !got.PriceUSD.Valid || got.PriceUSD.String != "1.50" {
		t.Errorf("PriceUSD = %v", got.PriceUSD)
	}
}

func TestUpsertCard_UpdateExisting(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	card := makeCachedCard("Sol Ring")
	card.PriceUSD = sql.NullString{String: "1.00", Valid: true}
	UpsertCard(card)

	// Update with new price
	card.PriceUSD = sql.NullString{String: "2.00", Valid: true}
	card.SetCode = "C21"
	if err := UpsertCard(card); err != nil {
		t.Fatalf("UpsertCard update failed: %v", err)
	}

	got, _ := GetCard("Sol Ring")
	if got.PriceUSD.String != "2.00" {
		t.Errorf("PriceUSD after update = %q, want 2.00", got.PriceUSD.String)
	}
	if got.SetCode != "C21" {
		t.Errorf("SetCode after update = %q, want C21", got.SetCode)
	}
}

func TestGetCard_NotFound(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	got, err := GetCard("Nonexistent Card")
	if err != nil {
		t.Fatalf("GetCard error: %v", err)
	}
	if got != nil {
		t.Error("expected nil for missing card")
	}
}

func TestGetCard_CaseInsensitive(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	card := makeCachedCard("Lightning Bolt")
	UpsertCard(card)

	// Should match case-insensitively
	got, _ := GetCard("lightning bolt")
	if got == nil {
		t.Fatal("case-insensitive lookup returned nil")
	}
	if got.Name != "Lightning Bolt" {
		t.Errorf("Name = %q, want Lightning Bolt", got.Name)
	}

	got2, _ := GetCard("LIGHTNING BOLT")
	if got2 == nil {
		t.Fatal("uppercase lookup returned nil")
	}
}

func TestGetCard_DFCNameStripping(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	// Insert with front-face name only (as Scryfall returns it)
	card := makeCachedCard("Delver of Secrets")
	card.IsDoubleFaced = true
	card.BackImageURI = "https://example.com/insectile.jpg"
	UpsertCard(card)

	// Lookup with full DFC name should find it via stripping
	got, err := GetCard("Delver of Secrets // Insectile Aberration")
	if err != nil {
		t.Fatalf("GetCard error: %v", err)
	}
	if got == nil {
		t.Fatal("DFC name lookup returned nil")
	}
	if got.Name != "Delver of Secrets" {
		t.Errorf("Name = %q, want Delver of Secrets", got.Name)
	}
}

func TestGetCard_DFCSlashVariant(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	card := makeCachedCard("Card A")
	UpsertCard(card)

	// Lookup with single-slash DFC variant
	got, _ := GetCard("Card A / Card B")
	if got == nil {
		t.Fatal("single-slash DFC lookup returned nil")
	}
}

// =====================
// CRITICAL: UpsertCard Null Price Handling
// =====================

func TestUpsertCard_NullPrices(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	card := makeCachedCard("Priceless Card")
	card.PriceUSD = sql.NullString{Valid: false}
	card.PriceUSDFoil = sql.NullString{Valid: false}
	card.PriceEUR = sql.NullString{Valid: false}
	card.PriceEURFoil = sql.NullString{Valid: false}
	UpsertCard(card)

	got, _ := GetCard("Priceless Card")
	if got == nil {
		t.Fatal("GetCard returned nil")
	}
	if got.PriceUSD.Valid {
		t.Error("PriceUSD should be NULL")
	}
	if got.PriceUSDFoil.Valid {
		t.Error("PriceUSDFoil should be NULL")
	}
}

// =====================
// CRITICAL: IsStale Tests
// =====================

func TestIsStale_MissingCard(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	stale, err := IsStale("Missing Card", 24)
	if err != nil {
		t.Fatalf("IsStale error: %v", err)
	}
	if !stale {
		t.Error("missing card should be treated as stale")
	}
}

func TestIsStale_FreshCard(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	card := makeCachedCard("Fresh Card")
	card.UpdatedAt = time.Now()
	UpsertCard(card)

	stale, err := IsStale("Fresh Card", 24)
	if err != nil {
		t.Fatalf("IsStale error: %v", err)
	}
	if stale {
		t.Error("card updated just now should not be stale")
	}
}

func TestIsStale_OldCard(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	card := makeCachedCard("Old Card")
	card.UpdatedAt = time.Now().Add(-48 * time.Hour)
	UpsertCard(card)

	stale, err := IsStale("Old Card", 24)
	if err != nil {
		t.Fatalf("IsStale error: %v", err)
	}
	if !stale {
		t.Error("card updated 48 hours ago should be stale with 24h threshold")
	}
}

func TestIsStale_CaseInsensitive(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	card := makeCachedCard("Sol Ring")
	card.UpdatedAt = time.Now()
	UpsertCard(card)

	stale, _ := IsStale("sol ring", 24)
	if stale {
		t.Error("case-insensitive stale check should find the card")
	}
}

// =====================
// CRITICAL: FromScryfallCard Conversion Tests
// =====================

func TestFromScryfallCard_SingleFaced(t *testing.T) {
	sc := &scryfall.Card{
		Name:            "Lightning Bolt",
		OracleID:        "abc-123",
		OracleText:      "Deal 3 damage.",
		TypeLine:        "Instant",
		ManaCost:        "{R}",
		CMC:             1.0,
		Colors:          []string{"R"},
		ColorIdentity:   []string{"R"},
		SetCode:         "m19",
		SetName:         "Core Set 2019",
		CollectorNumber: "126",
		ImageURIs:       scryfall.ImageURISet{Normal: "https://img.scryfall.com/bolt.jpg"},
		Legalities:      map[string]string{"commander": "legal"},
	}
	sc.Prices.USD = "1.50"
	sc.Prices.USDFoil = "3.00"

	cached := FromScryfallCard(sc)

	if cached.Name != "Lightning Bolt" {
		t.Errorf("Name = %q", cached.Name)
	}
	if cached.OracleText != "Deal 3 damage." {
		t.Errorf("OracleText = %q", cached.OracleText)
	}
	if cached.ImageURI != "https://img.scryfall.com/bolt.jpg" {
		t.Errorf("ImageURI = %q", cached.ImageURI)
	}
	if cached.BackImageURI != "" {
		t.Errorf("BackImageURI should be empty, got %q", cached.BackImageURI)
	}
	if cached.IsDoubleFaced {
		t.Error("should not be double-faced")
	}
	if !cached.PriceUSD.Valid || cached.PriceUSD.String != "1.50" {
		t.Errorf("PriceUSD = %v", cached.PriceUSD)
	}
}

func TestFromScryfallCard_DoubleFaced(t *testing.T) {
	sc := &scryfall.Card{
		Name:     "Delver of Secrets // Insectile Aberration",
		OracleID: "delver-id",
		CMC:      1.0,
		CardFaces: []scryfall.CardFace{
			{
				Name:       "Delver of Secrets",
				ManaCost:   "{U}",
				TypeLine:   "Creature — Human Wizard",
				OracleText: "At the beginning of your upkeep, look...",
				ImageURIs:  scryfall.ImageURISet{Normal: "https://img.scryfall.com/delver-front.jpg"},
			},
			{
				Name:       "Insectile Aberration",
				TypeLine:   "Creature — Human Insect",
				OracleText: "Flying",
				ImageURIs:  scryfall.ImageURISet{Normal: "https://img.scryfall.com/delver-back.jpg"},
			},
		},
		Colors:        []string{"U"},
		ColorIdentity: []string{"U"},
		SetCode:       "isd",
	}

	cached := FromScryfallCard(sc)

	if !cached.IsDoubleFaced {
		t.Error("should be double-faced")
	}
	if cached.ImageURI != "https://img.scryfall.com/delver-front.jpg" {
		t.Errorf("ImageURI = %q, want front face", cached.ImageURI)
	}
	if cached.BackImageURI != "https://img.scryfall.com/delver-back.jpg" {
		t.Errorf("BackImageURI = %q, want back face", cached.BackImageURI)
	}
	if cached.OracleText != "At the beginning of your upkeep, look..." {
		t.Errorf("OracleText should be front face text, got %q", cached.OracleText)
	}
	if cached.ManaCost != "{U}" {
		t.Errorf("ManaCost = %q, want {U}", cached.ManaCost)
	}
	if cached.TypeLine != "Creature — Human Wizard" {
		t.Errorf("TypeLine = %q, want front face type", cached.TypeLine)
	}
}

func TestFromScryfallCard_EmptyPrices(t *testing.T) {
	sc := &scryfall.Card{
		Name: "Token Card",
	}
	// Prices are empty strings by default

	cached := FromScryfallCard(sc)

	if cached.PriceUSD.Valid {
		t.Error("PriceUSD should not be valid for empty price")
	}
	if cached.PriceEUR.Valid {
		t.Error("PriceEUR should not be valid for empty price")
	}
}

func TestFromScryfallCard_NilLegalities(t *testing.T) {
	sc := &scryfall.Card{
		Name:       "Test",
		Legalities: nil,
	}
	cached := FromScryfallCard(sc)
	if cached.Legalities != "{}" {
		t.Errorf("Legalities = %q, want {}", cached.Legalities)
	}
}

// =====================
// CRITICAL: Unmatched Cards CRUD
// =====================

func TestUnmatchedCards_AddAndRetrieve(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	if err := AddUnmatchedCard("Fake Card", "my-deck"); err != nil {
		t.Fatal(err)
	}
	if err := AddUnmatchedCard("Another Fake", "my-deck"); err != nil {
		t.Fatal(err)
	}

	names, err := GetUnmatchedCards("my-deck")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("got %d unmatched, want 2", len(names))
	}
}

func TestUnmatchedCards_IsolatedByDeck(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	AddUnmatchedCard("Card A", "deck-1")
	AddUnmatchedCard("Card B", "deck-2")

	names1, _ := GetUnmatchedCards("deck-1")
	names2, _ := GetUnmatchedCards("deck-2")

	if len(names1) != 1 || names1[0] != "Card A" {
		t.Errorf("deck-1 unmatched = %v, want [Card A]", names1)
	}
	if len(names2) != 1 || names2[0] != "Card B" {
		t.Errorf("deck-2 unmatched = %v, want [Card B]", names2)
	}
}

func TestClearUnmatchedCards(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	AddUnmatchedCard("Card A", "my-deck")
	AddUnmatchedCard("Card B", "my-deck")
	AddUnmatchedCard("Card C", "other-deck")

	if err := ClearUnmatchedCards("my-deck"); err != nil {
		t.Fatal(err)
	}

	names, _ := GetUnmatchedCards("my-deck")
	if len(names) != 0 {
		t.Errorf("my-deck should have 0 unmatched after clear, got %d", len(names))
	}

	// other-deck should be unaffected
	names2, _ := GetUnmatchedCards("other-deck")
	if len(names2) != 1 {
		t.Errorf("other-deck should still have 1 unmatched, got %d", len(names2))
	}
}

func TestUnmatchedCards_UpsertOnDuplicate(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	// Add same card twice — should not error (INSERT OR REPLACE)
	AddUnmatchedCard("Duplicate", "my-deck")
	if err := AddUnmatchedCard("Duplicate", "my-deck"); err != nil {
		t.Fatalf("duplicate insert failed: %v", err)
	}

	names, _ := GetUnmatchedCards("my-deck")
	if len(names) != 1 {
		t.Errorf("got %d entries for duplicate, want 1", len(names))
	}
}

func TestGetUnmatchedCards_EmptyDeck(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	names, err := GetUnmatchedCards("nonexistent-deck")
	if err != nil {
		t.Fatal(err)
	}
	if names != nil && len(names) != 0 {
		t.Errorf("expected empty result, got %v", names)
	}
}

// =====================
// MEDIUM: Close and Re-Init Lifecycle
// =====================

func TestClose_AndReInit(t *testing.T) {
	dir := setupTestDB(t)

	card := makeCachedCard("Persistent Card")
	UpsertCard(card)

	// Close
	if err := Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Re-init on the same directory
	if err := Init(dir); err != nil {
		t.Fatalf("Re-Init failed: %v", err)
	}

	// Data should persist
	got, err := GetCard("Persistent Card")
	if err != nil {
		t.Fatalf("GetCard after re-init failed: %v", err)
	}
	if got == nil {
		t.Fatal("card should persist after Close + Re-Init")
	}

	cleanupTestDB(t, dir)
}

// =====================
// MEDIUM: Helper Function Tests
// =====================

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"R"}, "R"},
		{[]string{"W", "U", "B"}, "W,U,B"},
	}
	for _, tt := range tests {
		got := joinStrings(tt.input)
		if got != tt.want {
			t.Errorf("joinStrings(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMapToJSON(t *testing.T) {
	// nil map
	if got := mapToJSON(nil); got != "{}" {
		t.Errorf("mapToJSON(nil) = %q, want {}", got)
	}

	// empty map
	if got := mapToJSON(map[string]string{}); got != "{}" {
		t.Errorf("mapToJSON({}) = %q, want {}", got)
	}

	// single entry
	got := mapToJSON(map[string]string{"commander": "legal"})
	if got != `{"commander":"legal"}` {
		t.Errorf("mapToJSON = %q", got)
	}
}

func TestNullString(t *testing.T) {
	valid := nullString("1.50")
	if !valid.Valid || valid.String != "1.50" {
		t.Errorf("nullString(1.50) = %v", valid)
	}

	invalid := nullString("")
	if invalid.Valid {
		t.Error("nullString('') should not be valid")
	}
}
