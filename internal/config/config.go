package config

import (
	"os"
	"path/filepath"
)

// Config holds app-level configuration.
type Config struct {
	// RootDir is the base directory of the mtg-db repo (where decks/ lives).
	RootDir string

	// DataDir is where app-generated data lives (SQLite, image cache, etc.).
	DataDir string
}

// New creates a Config based on the given root directory.
// If rootDir is empty, it defaults to the current working directory.
func New(rootDir string) (*Config, error) {
	if rootDir == "" {
		var err error
		rootDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	dataDir := filepath.Join(rootDir, "data")

	// Ensure data directories exist
	dirs := []string{
		dataDir,
		filepath.Join(dataDir, "images", "cache"),
		filepath.Join(dataDir, "images", "custom"),
		filepath.Join(dataDir, "images", "mpc"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	return &Config{
		RootDir: rootDir,
		DataDir: dataDir,
	}, nil
}

// DecksDir returns the path to the decks directory.
func (c *Config) DecksDir() string {
	return filepath.Join(c.RootDir, "decks")
}

// WishlistsDir returns the path to the wishlists directory.
func (c *Config) WishlistsDir() string {
	return filepath.Join(c.RootDir, "wishlists")
}

// ImageCacheDir returns the path to the Scryfall image cache.
func (c *Config) ImageCacheDir() string {
	return filepath.Join(c.DataDir, "images", "cache")
}

// CustomImagesDir returns the path to user-provided global images.
func (c *Config) CustomImagesDir() string {
	return filepath.Join(c.DataDir, "images", "custom")
}

// MPCImagesDir returns the path to MPC Autofill cached images.
func (c *Config) MPCImagesDir() string {
	return filepath.Join(c.DataDir, "images", "mpc")
}

// DBPath returns the path to the SQLite database.
func (c *Config) DBPath() string {
	return filepath.Join(c.DataDir, "cards.db")
}
