package app

import "app/internal/config"

// AppState is sent to the frontend on startup to determine what view to show.
type AppState struct {
	hasCollection   bool             `json:"hasCollection"`
	CollectionPath  string           `json:"collectionPath"`
	CollectionLabel string           `json:"collectionLabel"`
	CollectionValid bool             `json:"collectionValid"`
	Collections     []CollectionInfo `json:"collections"`
	NeedsSetup      bool             `json:"needsSetup"`
}

// CollectionInfo is a frontend-friendly view of a known collection.
type CollectionInfo struct {
	Path       string `json:"path"`
	Label      string `json:"label"`
	LastOpened string `json:"lastOpened"`
	IsActive   bool   `json:"isActive"`
	IsValid    bool   `json:"isValid"`
}

// GetAppState returns the current app state for the frontend.
func (a *App) GetAppState() AppState {
	if a.config == nil {
		return AppState{NeedsSetup: true}
	}

	state := AppState{
		hasCollection: a.config.HasActiveCollection(),
		NeedsSetup:    !a.config.HasActiveCollection(),
	}

	if state.hasCollection {
		state.CollectionPath = a.config.ActiveCollectionPath()
		err := config.ValidateCollectionDir(state.CollectionPath)
		state.CollectionValid = err == nil

		// Find label
		for _, c := range a.config.File.Collections {
			if c.Path == state.CollectionPath {
				state.CollectionLabel = c.Label
				break
			}
		}
	}

	// Build collections list
	for _, c := range a.config.File.Collections {
		err := config.ValidateCollectionDir(c.Path)
		state.Collections = append(state.Collections, CollectionInfo{
			Path:       c.Path,
			Label:      c.Label,
			LastOpened: c.LastOpened.Format("2006-01-02"),
			IsActive:   c.Path == a.config.ActiveCollectionPath(),
			IsValid:    err == nil,
		})
	}

	return state
}
