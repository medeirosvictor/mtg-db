package scryfall

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
)

// DownloadImage downloads a card image from a URL and saves it to the cache directory
func DownloadImage(url, cacheDir, cardSlug string) (string, error) {
	if url == "" {
		return "", nil
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create a safe filename from the card name
	safeName := sanitizeFilename(cardSlug)
	filePath := filepath.Join(cacheDir, safeName+".jpg")

	// Check if already cached
	if _, err := os.Stat(filePath); err == nil {
		log.Printf("[Images] Using cached: %s", filePath)
		return filePath, nil // Already exists
	}

	// Download the image
	log.Printf("[Images] Downloading: %s -> %s", url, filePath)
	resp, err := httpClient.Get(url)
	if err != nil {
		log.Printf("[Images] Download failed: %v", err)
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Images] Download failed with status %d", resp.StatusCode)
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	// Copy the image data
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	log.Printf("[Images] Downloaded successfully: %s", filePath)
	return filePath, nil
}

// TODO: Phase 4A - Used for local image override resolution chain
// GetImagePath returns the path to a cached image, or empty string if not cached
func GetImagePath(cacheDir, cardSlug string) string {
	safeName := sanitizeFilename(cardSlug)
	filePath := filepath.Join(cacheDir, safeName+".jpg")

	if _, err := os.Stat(filePath); err == nil {
		return filePath
	}
	return ""
}

// sanitizeFilename creates a safe filename from a card name
func sanitizeFilename(name string) string {
	// Replace problematic characters
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, "*", "-")
	name = strings.ReplaceAll(name, "?", "-")
	name = strings.ReplaceAll(name, "\"", "-")
	name = strings.ReplaceAll(name, "<", "-")
	name = strings.ReplaceAll(name, ">", "-")
	name = strings.ReplaceAll(name, "|", "-")
	name = strings.ReplaceAll(name, "(", "-")
	name = strings.ReplaceAll(name, ")", "-")
	name = strings.ReplaceAll(name, "[", "-")
	name = strings.ReplaceAll(name, "]", "-")
	name = strings.ReplaceAll(name, "{", "-")
	name = strings.ReplaceAll(name, "}", "-")

	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Convert to lowercase
	name = strings.ToLower(name)

	// Limit length
	if len(name) > 100 {
		name = name[:100]
	}

	return name
}

// TODO: Phase 5 - Cache management UI
// ClearImageCache removes all cached images
func ClearImageCache(cacheDir string) error {
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			os.Remove(filepath.Join(cacheDir, entry.Name()))
		}
	}
	return nil
}
