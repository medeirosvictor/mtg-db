package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Collection represents a known collection folder.
type Collection struct {
	Path       string    `json:"path"`
	Label      string    `json:"label"`
	LastOpened time.Time `json:"lastOpened"`
}

// Preferences holds user-configurable settings.
type Preferences struct {
	ProxyThreshold     float64 `json:"proxyThreshold"`
	PriceStalenessHours int    `json:"priceStalenessHours"`
}

// ConfigFile is the JSON structure persisted to disk.
type ConfigFile struct {
	ActiveCollection string       `json:"activeCollection"`
	Collections      []Collection `json:"collections"`
	Preferences      Preferences  `json:"preferences"`
}

// Config holds the runtime app configuration.
type Config struct {
	// AppDataDir is the OS-specific app data directory (%APPDATA%/mtg-db, etc.)
	AppDataDir string

	// ConfigFilePath is the full path to config.json
	ConfigFilePath string

	// File is the loaded (or default) config file data
	File ConfigFile
}

// DefaultPreferences returns sensible defaults.
func DefaultPreferences() Preferences {
	return Preferences{
		ProxyThreshold:      5.00,
		PriceStalenessHours: 24,
	}
}

// appDataDir returns the platform-specific app data directory.
func appDataDir() (string, error) {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("APPDATA")
		if base == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, "Library", "Application Support")
	default: // linux, etc.
		base = os.Getenv("XDG_DATA_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, ".local", "share")
		}
	}
	return filepath.Join(base, "mtg-db"), nil
}

// Load reads the config from the OS app data directory.
// If no config exists, returns a Config with defaults (no active collection).
func Load() (*Config, error) {
	dir, err := appDataDir()
	if err != nil {
		return nil, fmt.Errorf("determine app data dir: %w", err)
	}

	// Ensure app data directories exist
	dirs := []string{
		dir,
		filepath.Join(dir, "images", "cache"),
		filepath.Join(dir, "images", "custom"),
		filepath.Join(dir, "images", "mpc"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, fmt.Errorf("create dir %s: %w", d, err)
		}
	}

	cfgPath := filepath.Join(dir, "config.json")
	cfg := &Config{
		AppDataDir:     dir,
		ConfigFilePath: cfgPath,
		File: ConfigFile{
			Preferences: DefaultPreferences(),
		},
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			// First launch — no config yet
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, &cfg.File); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

// Save writes the config to disk.
func (c *Config) Save() error {
	data, err := json.MarshalIndent(c.File, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(c.ConfigFilePath, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// ActiveCollectionPath returns the path of the currently active collection, or "".
func (c *Config) ActiveCollectionPath() string {
	return c.File.ActiveCollection
}

// SetActiveCollection sets the active collection path, adds it to known collections
// if not already present, updates LastOpened, and saves.
func (c *Config) SetActiveCollection(path string) error {
	c.File.ActiveCollection = path

	// Update or add to collections list
	found := false
	for i := range c.File.Collections {
		if c.File.Collections[i].Path == path {
			c.File.Collections[i].LastOpened = time.Now()
			found = true
			break
		}
	}
	if !found {
		c.File.Collections = append(c.File.Collections, Collection{
			Path:       path,
			Label:      filepath.Base(path),
			LastOpened: time.Now(),
		})
	}

	return c.Save()
}

// SetCollectionLabel updates the label for a collection path.
func (c *Config) SetCollectionLabel(path, label string) error {
	for i := range c.File.Collections {
		if c.File.Collections[i].Path == path {
			c.File.Collections[i].Label = label
			return c.Save()
		}
	}
	return fmt.Errorf("collection not found: %s", path)
}

// RemoveCollection removes a collection from the known list.
// If it was active, clears activeCollection.
func (c *Config) RemoveCollection(path string) error {
	var filtered []Collection
	for _, col := range c.File.Collections {
		if col.Path != path {
			filtered = append(filtered, col)
		}
	}
	c.File.Collections = filtered
	if c.File.ActiveCollection == path {
		c.File.ActiveCollection = ""
		if len(filtered) > 0 {
			c.File.ActiveCollection = filtered[0].Path
		}
	}
	return c.Save()
}

// HasActiveCollection returns true if there's a valid active collection configured.
func (c *Config) HasActiveCollection() bool {
	return c.File.ActiveCollection != ""
}

// --- Directory helpers (for app-generated data in AppDataDir) ---

// DecksDir returns the decks/ path inside the active collection.
func (c *Config) DecksDir() string {
	return filepath.Join(c.File.ActiveCollection, "decks")
}

// TODO: Phase 2B - Wishlist & purchase planning
// WishlistsDir returns the wishlists/ path inside the active collection.
func (c *Config) WishlistsDir() string {
	return filepath.Join(c.File.ActiveCollection, "wishlists")
}

// ImageCacheDir returns the path to the Scryfall image cache (in app data).
func (c *Config) ImageCacheDir() string {
	return filepath.Join(c.AppDataDir, "images", "cache")
}

// TODO: Phase 4A - Local image overrides
// CustomImagesDir returns the path to user-provided global images (in app data).
func (c *Config) CustomImagesDir() string {
	return filepath.Join(c.AppDataDir, "images", "custom")
}

// TODO: Phase 4B - MPC Autofill proxy art
// MPCImagesDir returns the path to MPC Autofill cached images (in app data).
func (c *Config) MPCImagesDir() string {
	return filepath.Join(c.AppDataDir, "images", "mpc")
}

// TODO: Phase 2A - Direct DB path access for collection management
// DBPath returns the path to the SQLite database (in app data).
func (c *Config) DBPath() string {
	return filepath.Join(c.AppDataDir, "cards.db")
}

// ValidateCollectionDir checks that a path is a valid collection folder.
// Returns nil if valid, or an error describing what's wrong.
func ValidateCollectionDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("folder does not exist: %s", path)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", path)
	}

	decksDir := filepath.Join(path, "decks")
	info, err = os.Stat(decksDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("no decks/ subfolder found in %s", path)
	}

	// Check for at least one deck subfolder with deck.txt
	entries, err := os.ReadDir(decksDir)
	if err != nil {
		return fmt.Errorf("cannot read decks/: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		deckFile := filepath.Join(decksDir, entry.Name(), "deck.txt")
		if _, err := os.Stat(deckFile); err == nil {
			return nil // Found at least one valid deck
		}
	}

	return fmt.Errorf("no decks found in %s/decks/ (expected subfolders with deck.txt)", path)
}

// InitializeCollectionDir creates the skeleton structure for a new collection.
func InitializeCollectionDir(path string) error {
	dirs := []string{
		filepath.Join(path, "decks"),
		filepath.Join(path, "wishlists"),
		filepath.Join(path, "history"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", d, err)
		}
	}
	return nil
}
