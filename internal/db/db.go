package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"app/internal/scryfall"

	_ "github.com/mattn/go-sqlite3"
)

// DB Interface allows for dependency injection and mocking
type DB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Close() error
}

// database wraps sql.DB with a mutex for thread safety
type database struct {
	*sql.DB
	mu sync.RWMutex
}

// Global database instance - protected by mutex for thread safety
var (
	db   *database
	dbMu sync.RWMutex
)

// SetDB allows injecting a custom DB implementation (useful for testing)
func SetDB(custom DB) {
	dbMu.Lock()
	defer dbMu.Unlock()
	if custom == nil {
		return
	}
	// If it's already a *database, use it directly
	if d, ok := custom.(*database); ok {
		db = d
		return
	}
	// Wrap other implementations
	db = &database{DB: custom.(*sql.DB)}
}

// GetDB returns the current database instance (thread-safe)
func GetDB() *database {
	dbMu.RLock()
	defer dbMu.RUnlock()
	return db
}

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
	dbMu.Lock()
	defer dbMu.Unlock()

	// Already initialized
	if db != nil {
		return nil
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "cards.db")
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return fmt.Errorf("failed to set WAL mode: %w", err)
	}

	// Configure connection pool for better concurrency
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	db = &database{DB: sqlDB}

	// Create tables
	if err := createTablesLocked(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Migrate: add DFC columns if missing
	if err := migrateAddDFCColumnsLocked(); err != nil {
		return fmt.Errorf("failed to migrate DFC columns: %w", err)
	}

	return nil
}

// migrateAddDFCColumns adds back_image_uri and is_double_faced columns if they don't exist.
func migrateAddDFCColumnsLocked() error {
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
func createTablesLocked() error {
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
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil || db.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

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
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil || db.DB == nil {
		return fmt.Errorf("database not initialized")
	}

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
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil || db.DB == nil {
		return false, fmt.Errorf("database not initialized")
	}

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
		Legalities:      MapToJSON(sc.Legalities),
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

// MapToJSON converts a map to JSON string - fixed version with proper escaping
func MapToJSON(m map[string]string) string {
	if m == nil {
		return "{}"
	}
	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ","
		}
		// Escape special characters in values
		escapedV := escapeJSON(v)
		result += fmt.Sprintf("\"%s\":\"%s\"", escapeJSON(k), escapedV)
		first = false
	}
	result += "}"
	return result
}

// escapeJSON escapes special characters in JSON strings
func escapeJSON(s string) string {
	result := ""
	for _, c := range s {
		switch c {
		case '"':
			result += "\\\""
		case '\\':
			result += "\\\\"
		case '\n':
			result += "\\n"
		case '\r':
			result += "\\r"
		case '\t':
			result += "\\t"
		default:
			result += string(c)
		}
	}
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
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil || db.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.Exec("DELETE FROM unmatched_cards WHERE deck_slug = ?", deckSlug)
	return err
}

// AddUnmatchedCard records a card that couldn't be found on Scryfall
func AddUnmatchedCard(name, deckSlug string) error {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil || db.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.Exec(`
		INSERT OR REPLACE INTO unmatched_cards (name, deck_slug, created_at)
		VALUES (?, ?, ?)
	`, name, deckSlug, time.Now())
	return err
}

// GetUnmatchedCards retrieves all unmatched cards for a deck
func GetUnmatchedCards(deckSlug string) ([]string, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil || db.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

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
	dbMu.Lock()
	defer dbMu.Unlock()

	if db != nil && db.DB != nil {
		return db.DB.Close()
	}
	return nil
}

// Reset closes the current database and resets the global to nil
// Useful for testing to ensure clean state
func Reset() {
	dbMu.Lock()
	defer dbMu.Unlock()

	if db != nil && db.DB != nil {
		db.DB.Close()
	}
	db = nil
}
