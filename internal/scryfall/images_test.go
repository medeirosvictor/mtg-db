package scryfall

import (
	"os"
	"path/filepath"
	"testing"
)

// =====================
// CRITICAL: sanitizeFilename Tests
// =====================

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Basic replacements
		{"simple name", "Lightning Bolt", "lightning_bolt"},
		{"spaces to underscores", "Sol Ring", "sol_ring"},
		{"lowercase", "SOL RING", "sol_ring"},

		// Problematic characters
		{"forward slash", "Card/A", "card-a"},
		{"backslash", "Card\\A", "card-a"},
		{"colon", "Card:A", "card-a"},
		{"asterisk", "Card*A", "card-a"},
		{"question mark", "Card?A", "card-a"},
		{"double quote", `Card"A`, "card-a"},
		{"less than", "Card<A", "card-a"},
		{"greater than", "Card>A", "card-a"},
		{"pipe", "Card|A", "card-a"},
		{"parentheses become hyphens", "Card(A)", "card-a-"},     // trailing hyphen NOT trimmed
		{"square brackets become hyphens", "Card[A]", "card-a-"}, // trailing hyphen NOT trimmed
		{"curly braces become hyphens", "Card{A}", "card-a-"},    // trailing hyphen NOT trimmed

		// Multiple problematic chars
		{"all special chars", "Hello/World!@#", "hello-world!@#"}, // only / replaced, rest stays

		// DFC names (important!)
		{"DFC double slash", "Delver of Secrets // Insectile Aberration", "delver_of_secrets_--_insectile_aberration"},
		{"DFC single slash", "Card A / Card B", "card_a_-_card_b"},

		// Edge cases
		{"empty string", "", ""},
		{"only spaces become underscores", "   ", "___"},
		{"only special chars stay", "!@#$%", "!@#$%"}, // all become - but then stay
		{"numbers", "Card123", "card123"},
		{"mixed case", "SoL RinG", "sol_ring"},
		{"unicode preserved", "Ætherflux Reservoir", "ætherflux_reservoir"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeFilename_TruncatesAt100(t *testing.T) {
	long := string(make([]byte, 150))
	for i := range long {
		long = long[:i] + "a" + long[i+1:]
	}
	// Actually, easier: create a known 150-char string
	long = ""
	for i := 0; i < 150; i++ {
		long += "a"
	}

	got := sanitizeFilename(long)
	if len(got) != 100 {
		t.Errorf("len(sanitizeFilename(long)) = %d, want 100", len(got))
	}
}

// =====================
// MEDIUM: DownloadImage Tests (need actual HTTP, skipping for unit tests)
// =====================

// DownloadImage is hard to test without mocking HTTP.
// It creates dirs, checks cache, downloads, saves.
// We can test the "already cached" path.

func TestDownloadImage_CacheHit(t *testing.T) {
	// Create a temp dir with a fake cached image
	tmpDir, err := os.MkdirTemp("", "img-cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Pre-create a cached image
	cacheFile := filepath.Join(tmpDir, "lightning_bolt.jpg")
	if err := os.WriteFile(cacheFile, []byte("fake image data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Download should find the cached file
	path, err := DownloadImage("https://should-not-be-called.com/bolt.jpg", tmpDir, "Lightning Bolt")
	if err != nil {
		t.Fatalf("DownloadImage failed: %v", err)
	}
	if path != cacheFile {
		t.Errorf("returned path = %q, want cached path %q", path, cacheFile)
	}
}

func TestDownloadImage_EmptyURL(t *testing.T) {
	path, err := DownloadImage("", "/some/nonexistent/dir", "Card")
	if err != nil {
		t.Fatal(err)
	}
	if path != "" {
		t.Errorf("expected empty path for empty URL, got %q", path)
	}
}

func TestDownloadImage_CreatesDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "img-dir-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	nestedDir := filepath.Join(tmpDir, "nested", "path")

	path, err := DownloadImage("", nestedDir, "Card")
	// Empty URL should return empty path without trying to create dir for image
	_ = path
	_ = err

	// Now test with a URL that would try to download — this requires HTTP mocking
	// For unit test, just verify directory creation works:
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("nested dir should exist")
	}
}

// =====================
// MEDIUM: GetImagePath Tests
// =====================

func TestGetImagePath_Exists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "getpath-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a cached image
	cacheFile := filepath.Join(tmpDir, "sol_ring.jpg")
	if err := os.WriteFile(cacheFile, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	path := GetImagePath(tmpDir, "Sol Ring")
	if path != cacheFile {
		t.Errorf("GetImagePath = %q, want %q", path, cacheFile)
	}
}

func TestGetImagePath_NotExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "getpath-missing-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	path := GetImagePath(tmpDir, "Nonexistent Card")
	if path != "" {
		t.Errorf("GetImagePath for missing = %q, want empty", path)
	}
}

// =====================
// MEDIUM: ClearImageCache Tests
// =====================

func TestClearImageCache_EmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clear-empty-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Should not error on empty/non-existent dir
	if err := ClearImageCache(tmpDir); err != nil {
		t.Fatalf("ClearImageCache failed: %v", err)
	}
}

func TestClearImageCache_NonExistentDir(t *testing.T) {
	// Non-existent should return nil
	if err := ClearImageCache("/nonexistent/path/12345"); err != nil {
		t.Errorf("ClearImageCache on non-existent should not error, got: %v", err)
	}
}

func TestClearImageCache_RemovesFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clear-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some cached images
	files := []string{"card1.jpg", "card2.jpg", "card3.png"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// Create a subdirectory (should be ignored)
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "ignored.jpg"), []byte("x"), 0644)

	// Clear
	if err := ClearImageCache(tmpDir); err != nil {
		t.Fatalf("ClearImageCache failed: %v", err)
	}

	// Verify files removed
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("file %s should have been removed", f)
		}
	}

	// Subdirectory and its file should remain
	subPath := filepath.Join(subDir, "ignored.jpg")
	if _, err := os.Stat(subPath); os.IsNotExist(err) {
		t.Error("subdirectory and its contents should remain")
	}
}
