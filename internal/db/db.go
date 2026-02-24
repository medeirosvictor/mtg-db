package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"app/internal/scryfall"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// CachedCard represents a card in our local cache
type CachedCard struct {
	Name            string
	OracleID        string
	OracleText      string
	TypeLine        string
	ManaCost        string
	CMC             float64
	Colors          string // JSON array
	ColorIdentity   string // JSON array
	SetCode         string
	SetName         string
	CollectorNumber string
	ImageURI        string
	BackImageURI    string // Back face image for DFCs
	IsDoubleFaced   bool
	PriceUSD        sql.NullString
	PriceUSDFoil    sql.NullString
	PriceEUR        sql.NullString
	PriceEURFoil    sql.NullString
	Legalities      string // JSON object
	UpdatedAt       time.Time
}

// Init opens or creates the SQLite database
func Init(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "cards.db")
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return fmt.Errorf("failed to set WAL mode: %w", err)
	}

	// Create tables
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Migrate: add DFC columns if missing
	if err := migrateAddDFCColumns(); err != nil {
		return fmt.Errorf("failed to migrate DFC columns: %w", err)
	}

	return nil
}

// migrateAddDFCColumns adds back_image_uri and is_double_faced columns if they don't exist.
func migrateAddDFCColumns() error {
	// Check if column exists by attempting a query
	_, err := db.Exec("SELECT back_image_uri FROM cards LIMIT 0")
	if err != nil {
		// Column doesn't exist, add it
		if _, err := db.Exec("ALTER TABLE cards ADD COLUMN back_image_uri TEXT DEFAULT ''"); err != nil {
			return err
		}
		if _, err := db.Exec("ALTER TABLE cards ADD COLUMN is_double_faced INTEGER DEFAULT 0"); err != nil {
			return err
		}
	}
	return nil
}

// createTables creates the necessary database tables
func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS cards (
		name TEXT PRIMARY KEY,
		oracle_id TEXT,
		oracle_text TEXT,
		type_line TEXT,
		mana_cost TEXT,
		cmc REAL,
		colors TEXT,
		color_identity TEXT,
		set_code TEXT,
		set_name TEXT,
		collector_number TEXT,
		image_uri TEXT,
		back_image_uri TEXT DEFAULT '',
		is_double_faced INTEGER DEFAULT 0,
		price_usd TEXT,
		price_usd_foil TEXT,
		price_eur TEXT,
		price_eur_foil TEXT,
		legalities TEXT,
		updated_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_cards_set ON cards(set_code);
	CREATE INDEX IF NOT EXISTS idx_cards_updated ON cards(updated_at);

	CREATE TABLE IF NOT EXISTS unmatched_cards (
		name TEXT PRIMARY KEY,
		deck_slug TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.Exec(schema)
	return err
}

// GetCard retrieves a card from the cache by name (case-insensitive).
// Also tries looking up by the front-face name for DFC cards.
func GetCard(name string) (*CachedCard, error) {
	card, err := getCardByName(name)
	if err != nil {
		return nil, err
	}
	if card != nil {
		return card, nil
	}
	// Try stripping DFC suffix: "Card A // Card B" → "Card A"
	stripped := scryfall.StripDFCName(name)
	if stripped != name {
		return getCardByName(stripped)
	}
	return nil, nil
}

func getCardByName(name string) (*CachedCard, error) {
	var card CachedCard
	err := db.QueryRow(`
		SELECT name, oracle_id, oracle_text, type_line, mana_cost, cmc, 
		       colors, color_identity, set_code, set_name, collector_number,
		       image_uri, back_image_uri, is_double_faced,
		       price_usd, price_usd_foil, price_eur, price_eur_foil,
		       legalities, updated_at
		FROM cards 
		WHERE LOWER(name) = LOWER(?)
	`, name).Scan(
		&card.Name, &card.OracleID, &card.OracleText, &card.TypeLine, &card.ManaCost, &card.CMC,
		&card.Colors, &card.ColorIdentity, &card.SetCode, &card.SetName, &card.CollectorNumber,
		&card.ImageURI, &card.BackImageURI, &card.IsDoubleFaced,
		&card.PriceUSD, &card.PriceUSDFoil, &card.PriceEUR, &card.PriceEURFoil,
		&card.Legalities, &card.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// UpsertCard inserts or updates a card in the cache
func UpsertCard(card *CachedCard) error {
	_, err := db.Exec(`
		INSERT INTO cards (name, oracle_id, oracle_text, type_line, mana_cost, cmc, 
		                  colors, color_identity, set_code, set_name, collector_number,
		                  image_uri, back_image_uri, is_double_faced,
		                  price_usd, price_usd_foil, price_eur, price_eur_foil,
		                  legalities, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			oracle_id = excluded.oracle_id,
			oracle_text = excluded.oracle_text,
			type_line = excluded.type_line,
			mana_cost = excluded.mana_cost,
			cmc = excluded.cmc,
			colors = excluded.colors,
			color_identity = excluded.color_identity,
			set_code = excluded.set_code,
			set_name = excluded.set_name,
			collector_number = excluded.collector_number,
			image_uri = excluded.image_uri,
			back_image_uri = excluded.back_image_uri,
			is_double_faced = excluded.is_double_faced,
			price_usd = excluded.price_usd,
			price_usd_foil = excluded.price_usd_foil,
			price_eur = excluded.price_eur,
			price_eur_foil = excluded.price_eur_foil,
			legalities = excluded.legalities,
			updated_at = excluded.updated_at
	`, card.Name, card.OracleID, card.OracleText, card.TypeLine, card.ManaCost, card.CMC,
		card.Colors, card.ColorIdentity, card.SetCode, card.SetName, card.CollectorNumber,
		card.ImageURI, card.BackImageURI, card.IsDoubleFaced,
		card.PriceUSD, card.PriceUSDFoil, card.PriceEUR, card.PriceEURFoil,
		card.Legalities, card.UpdatedAt)
	return err
}

// IsStale checks if a card's data is older than the given hours
func IsStale(name string, hours int) (bool, error) {
	var updatedAt time.Time
	err := db.QueryRow("SELECT updated_at FROM cards WHERE LOWER(name) = LOWER(?)", name).Scan(&updatedAt)
	if err == sql.ErrNoRows {
		return true, nil // Not in cache = treat as stale
	}
	if err != nil {
		return false, err
	}
	return time.Since(updatedAt) > time.Duration(hours)*time.Hour, nil
}

// FromScryfallCard converts a Scryfall card to a cached card.
// Handles double-faced cards by extracting front face data.
func FromScryfallCard(sc *scryfall.Card) *CachedCard {
	return &CachedCard{
		Name:            sc.Name,
		OracleID:        sc.OracleID,
		OracleText:      sc.FrontOracleText(),
		TypeLine:        sc.FrontTypeLine(),
		ManaCost:        sc.FrontManaCost(),
		CMC:             sc.CMC,
		Colors:          `"` + joinStrings(sc.Colors) + `"`,
		ColorIdentity:   `"` + joinStrings(sc.ColorIdentity) + `"`,
		SetCode:         sc.SetCode,
		SetName:         sc.SetName,
		CollectorNumber: sc.CollectorNumber,
		ImageURI:        sc.FrontFaceImage(),
		BackImageURI:    sc.BackFaceImage(),
		IsDoubleFaced:   sc.IsDoubleFaced(),
		PriceUSD:        nullString(sc.Prices.USD),
		PriceUSDFoil:    nullString(sc.Prices.USDFoil),
		PriceEUR:        nullString(sc.Prices.EUR),
		PriceEURFoil:    nullString(sc.Prices.EURFoil),
		Legalities:      mapToJSON(sc.Legalities),
		UpdatedAt:       time.Now(),
	}
}

func joinStrings(s []string) string {
	result := ""
	for i, v := range s {
		if i > 0 {
			result += ","
		}
		result += v
	}
	return result
}

func mapToJSON(m map[string]string) string {
	if m == nil {
		return "{}"
	}
	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ","
		}
		result += fmt.Sprintf("\"%s\":\"%s\"", k, v)
		first = false
	}
	result += "}"
	return result
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// ClearUnmatchedCards removes all unmatched card records for a given deck slug.
func ClearUnmatchedCards(deckSlug string) error {
	_, err := db.Exec("DELETE FROM unmatched_cards WHERE deck_slug = ?", deckSlug)
	return err
}

// AddUnmatchedCard records a card that couldn't be found on Scryfall
func AddUnmatchedCard(name, deckSlug string) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO unmatched_cards (name, deck_slug, created_at)
		VALUES (?, ?, ?)
	`, name, deckSlug, time.Now())
	return err
}

// GetUnmatchedCards retrieves all unmatched cards for a deck
func GetUnmatchedCards(deckSlug string) ([]string, error) {
	rows, err := db.Query("SELECT name FROM unmatched_cards WHERE deck_slug = ?", deckSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
