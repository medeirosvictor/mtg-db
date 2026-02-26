package app

import (
	"encoding/json"
	"testing"
)

// =====================
// CRITICAL: AppState JSON Serialization Tests
// =====================

func TestAppState_HasCollectionExported(t *testing.T) {
	// Verify that HasCollection is exported and serializes to JSON.
	// This was a bug: the field was lowercase "hasCollection" (unexported),
	// which meant it was always omitted from JSON output.
	state := AppState{
		HasCollection:   true,
		CollectionPath:  "/some/path",
		CollectionLabel: "My Collection",
		CollectionValid: true,
		NeedsSetup:      false,
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	val, ok := m["hasCollection"]
	if !ok {
		t.Fatal("hasCollection key missing from JSON output — field may be unexported")
	}
	if val != true {
		t.Errorf("hasCollection = %v, want true", val)
	}
}

func TestAppState_HasCollectionFalse(t *testing.T) {
	state := AppState{
		HasCollection: false,
		NeedsSetup:    true,
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	val, ok := m["hasCollection"]
	if !ok {
		t.Fatal("hasCollection key missing from JSON output")
	}
	if val != false {
		t.Errorf("hasCollection = %v, want false", val)
	}
}

func TestAppState_AllFieldsSerialized(t *testing.T) {
	state := AppState{
		HasCollection:   true,
		CollectionPath:  "/test/path",
		CollectionLabel: "Test",
		CollectionValid: true,
		NeedsSetup:      false,
		Collections: []CollectionInfo{
			{
				Path:       "/test/path",
				Label:      "Test",
				LastOpened: "2026-02-26",
				IsActive:   true,
				IsValid:    true,
			},
		},
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedKeys := []string{
		"hasCollection",
		"collectionPath",
		"collectionLabel",
		"collectionValid",
		"collections",
		"needsSetup",
	}

	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("expected key %q missing from JSON output", key)
		}
	}
}

func TestCollectionInfo_Serialization(t *testing.T) {
	info := CollectionInfo{
		Path:       "/my/collection",
		Label:      "Commander Decks",
		LastOpened: "2026-01-15",
		IsActive:   true,
		IsValid:    true,
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedKeys := []string{"path", "label", "lastOpened", "isActive", "isValid"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("expected key %q missing from JSON output", key)
		}
	}

	if m["path"] != "/my/collection" {
		t.Errorf("path = %v, want /my/collection", m["path"])
	}
	if m["isActive"] != true {
		t.Errorf("isActive = %v, want true", m["isActive"])
	}
}
