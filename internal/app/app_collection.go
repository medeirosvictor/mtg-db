package app

import (
	"app/internal/config"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// SelectCollectionFolder opens a native folder picker dialog and validates the selection.
func (a *App) SelectCollectionFolder() string {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select your MTG collection folder",
	})
	if err != nil {
		return fmt.Sprintf("Failed to open folder picker: %v", err)
	}
	if dir == "" {
		return "" // User cancelled — not an error
	}

	// Validate
	if err := config.ValidateCollectionDir(dir); err != nil {
		return err.Error()
	}

	// Set as active
	if err := a.config.SetActiveCollection(dir); err != nil {
		return fmt.Sprintf("Failed to save config: %v", err)
	}

	a.loadDecks()
	return ""
}

// InitializeAndSelectFolder opens a native folder picker, creates collection skeleton.
func (a *App) InitializeAndSelectFolder() string {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Choose where to create your MTG collection",
	})
	if err != nil {
		return fmt.Sprintf("Failed to open folder picker: %v", err)
	}
	if dir == "" {
		return "" // User cancelled
	}

	// Check if it already has decks/
	if err := config.ValidateCollectionDir(dir); err == nil {
		if err := a.config.SetActiveCollection(dir); err != nil {
			return fmt.Sprintf("Failed to save config: %v", err)
		}
		a.loadDecks()
		return ""
	}

	// Initialize the skeleton
	if err := config.InitializeCollectionDir(dir); err != nil {
		return fmt.Sprintf("Failed to initialize collection: %v", err)
	}

	if err := a.config.SetActiveCollection(dir); err != nil {
		return fmt.Sprintf("Failed to save config: %v", err)
	}

	a.loadDecks()
	return ""
}

// SwitchCollection switches to a different known collection by path.
func (a *App) SwitchCollection(path string) string {
	if err := config.ValidateCollectionDir(path); err != nil {
		return err.Error()
	}

	if err := a.config.SetActiveCollection(path); err != nil {
		return fmt.Sprintf("Failed to save config: %v", err)
	}

	a.loadDecks()
	return ""
}

// RenameCollection updates the label for a collection.
func (a *App) RenameCollection(path, label string) string {
	if err := a.config.SetCollectionLabel(path, label); err != nil {
		return err.Error()
	}
	return ""
}

// OpenDeckFolder opens a deck's directory in the OS file explorer.
func (a *App) OpenDeckFolder(slug string) string {
	if a.config == nil || !a.config.HasActiveCollection() {
		return "No active collection"
	}

	deckDir := filepath.Join(a.config.DecksDir(), slug)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", deckDir)
	case "darwin":
		cmd = exec.Command("open", deckDir)
	default:
		cmd = exec.Command("xdg-open", deckDir)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Sprintf("Failed to open folder: %v", err)
	}
	return ""
}

// RemoveKnownCollection removes a collection from the known list.
func (a *App) RemoveKnownCollection(path string) string {
	if err := a.config.RemoveCollection(path); err != nil {
		return err.Error()
	}
	a.loadDecks()
	return ""
}
