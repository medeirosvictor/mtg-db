package scryfall

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// =====================
// CRITICAL: StripDFCName Tests
// =====================

func TestStripDFCName(t *testing.T) {
	tests := []struct {
		name string
		input string
		want  string
	}{
		{"single-faced card", "Lightning Bolt", "Lightning Bolt"},
		{"DFC with double slash", "Delver of Secrets // Insectile Aberration", "Delver of Secrets"},
		{"DFC with single slash", "Card A / Card B", "Card A"},
		{"no extra spacing", "A//B", "A//B"},                         // no space around // → not a DFC separator
		{"single slash no space", "A/B", "A/B"},                      // no space around / → not a DFC separator
		{"triple part name", "Part A // Part B // Part C", "Part A"}, // only first // matters
		{"leading spaces preserved", "  Card A // Card B", "Card A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripDFCName(tt.input)
			if got != tt.want {
				t.Errorf("StripDFCName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// =====================
// CRITICAL: Card Methods Tests
// =====================

func TestCard_IsDoubleFaced(t *testing.T) {
	single := &Card{Name: "Bolt"}
	if single.IsDoubleFaced() {
		t.Error("single-faced card should not be DFC")
	}

	dfc := &Card{
		Name: "Delver",
		CardFaces: []CardFace{
			{Name: "Delver of Secrets"},
			{Name: "Insectile Aberration"},
		},
	}
	if !dfc.IsDoubleFaced() {
		t.Error("two-faced card should be DFC")
	}
}

func TestCard_FrontFaceImage(t *testing.T) {
	// Card with top-level image
	card := &Card{
		ImageURIs: ImageURISet{Normal: "https://img/front.jpg"},
	}
	if got := card.FrontFaceImage(); got != "https://img/front.jpg" {
		t.Errorf("FrontFaceImage = %q, want top-level image", got)
	}

	// DFC without top-level image — should use first face
	dfc := &Card{
		CardFaces: []CardFace{
			{ImageURIs: ImageURISet{Normal: "https://img/face0.jpg"}},
			{ImageURIs: ImageURISet{Normal: "https://img/face1.jpg"}},
		},
	}
	if got := dfc.FrontFaceImage(); got != "https://img/face0.jpg" {
		t.Errorf("FrontFaceImage = %q, want first face image", got)
	}

	// No images at all
	empty := &Card{}
	if got := empty.FrontFaceImage(); got != "" {
		t.Errorf("FrontFaceImage = %q, want empty", got)
	}
}

func TestCard_BackFaceImage(t *testing.T) {
	dfc := &Card{
		CardFaces: []CardFace{
			{ImageURIs: ImageURISet{Normal: "https://img/face0.jpg"}},
			{ImageURIs: ImageURISet{Normal: "https://img/face1.jpg"}},
		},
	}
	if got := dfc.BackFaceImage(); got != "https://img/face1.jpg" {
		t.Errorf("BackFaceImage = %q", got)
	}

	single := &Card{}
	if got := single.BackFaceImage(); got != "" {
		t.Errorf("single-faced BackFaceImage = %q, want empty", got)
	}
}

func TestCard_FrontOracleText(t *testing.T) {
	// Top-level oracle text
	card := &Card{OracleText: "Deal 3 damage."}
	if got := card.FrontOracleText(); got != "Deal 3 damage." {
		t.Errorf("got %q", got)
	}

	// DFC fallback to first face
	dfc := &Card{
		CardFaces: []CardFace{
			{OracleText: "Face 0 text"},
			{OracleText: "Face 1 text"},
		},
	}
	if got := dfc.FrontOracleText(); got != "Face 0 text" {
		t.Errorf("got %q, want face 0 text", got)
	}
}

func TestCard_FrontManaCost(t *testing.T) {
	card := &Card{ManaCost: "{R}"}
	if got := card.FrontManaCost(); got != "{R}" {
		t.Errorf("got %q", got)
	}

	dfc := &Card{
		CardFaces: []CardFace{
			{ManaCost: "{U}"},
			{ManaCost: ""},
		},
	}
	if got := dfc.FrontManaCost(); got != "{U}" {
		t.Errorf("got %q", got)
	}
}

func TestCard_FrontTypeLine(t *testing.T) {
	// With card faces, use first face
	dfc := &Card{
		TypeLine: "Legendary Creature",
		CardFaces: []CardFace{
			{TypeLine: "Creature — Human Wizard"},
			{TypeLine: "Creature — Insect"},
		},
	}
	if got := dfc.FrontTypeLine(); got != "Creature — Human Wizard" {
		t.Errorf("got %q, want first face type line", got)
	}

	// Without card faces, use top-level
	single := &Card{TypeLine: "Instant"}
	if got := single.FrontTypeLine(); got != "Instant" {
		t.Errorf("got %q", got)
	}
}

// =====================
// CRITICAL: HTTP Client Tests (using httptest)
// =====================

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := &Client{
		http:    server.Client(),
		baseURL: server.URL,
	}
	return server, client
}

func TestFetchCardByName_Success(t *testing.T) {
	want := &Card{
		Name:     "Lightning Bolt",
		ManaCost: "{R}",
		CMC:      1.0,
	}
	resp := CardSearchResponse{
		Object: "card",
		Card:   want,
	}

	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cards/named" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("fuzzy") == "" {
			t.Error("missing fuzzy parameter")
		}
		if r.Header.Get("User-Agent") != UserAgent {
			t.Errorf("User-Agent = %q, want %q", r.Header.Get("User-Agent"), UserAgent)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	got, err := client.FetchCardByName("Lightning Bolt")
	if err != nil {
		t.Fatalf("FetchCardByName error: %v", err)
	}
	if got == nil {
		t.Fatal("expected card, got nil")
	}
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
	if got.ManaCost != want.ManaCost {
		t.Errorf("ManaCost = %q, want %q", got.ManaCost, want.ManaCost)
	}
}

func TestFetchCardByName_NotFound(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	got, err := client.FetchCardByName("Fake Card")
	if err != nil {
		t.Fatalf("unexpected error for 404: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil card for 404, got %+v", got)
	}
}

func TestFetchChunk_BulkLookup(t *testing.T) {
	wantCards := []*Card{
		{Name: "Sol Ring", CMC: 1},
		{Name: "Lightning Bolt", CMC: 1},
	}
	notFoundIds := []CollectionIdentifier{
		{Name: "Fake Card"},
	}

	server, client := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing Content-Type header")
		}

		var req CollectionRequest
		json.NewDecoder(r.Body).Decode(&req)

		if len(req.Identifiers) != 3 {
			t.Errorf("expected 3 identifiers, got %d", len(req.Identifiers))
		}

		resp := CollectionResponse{
			Object:   "list",
			Data:     wantCards,
			NotFound: notFoundIds,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	names := []string{"Sol Ring", "Lightning Bolt", "Fake Card"}
	cards, notFound, err := client.fetchChunk(names)
	if err != nil {
		t.Fatalf("fetchChunk error: %v", err)
	}
	if len(cards) != 2 {
		t.Errorf("got %d cards, want 2", len(cards))
	}
	if len(notFound) != 1 || notFound[0] != "Fake Card" {
		t.Errorf("notFound = %v, want [Fake Card]", notFound)
	}
}

// =====================
// CRITICAL: FetchCardsByNames Name Mapping Logic
// =====================
// These tests verify the DFC deduplication and name mapping logic
// without making actual HTTP calls.

func TestFetchCardsByNames_EmptyInput(t *testing.T) {
	client := NewClient()
	cards, notFound, nameMap, err := client.FetchCardsByNames([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cards != nil || notFound != nil || nameMap != nil {
		t.Error("empty input should return all nils")
	}
}

// Test DFC deduplication logic in isolation
func TestDFCNameDeduplication(t *testing.T) {
	names := []string{
		"Lightning Bolt",
		"Delver of Secrets // Insectile Aberration",
		"Delver of Secrets",
		"Card A / Card B",
	}

	// Simulate the deduplication logic from FetchCardsByNames
	strippedToOriginal := make(map[string][]string)
	var apiNames []string
	seen := make(map[string]bool)

	for _, name := range names {
		stripped := StripDFCName(name)
		key := toLower(stripped)
		strippedToOriginal[key] = append(strippedToOriginal[key], name)
		if !seen[key] {
			seen[key] = true
			apiNames = append(apiNames, stripped)
		}
	}

	// "Delver of Secrets // Insectile Aberration" and "Delver of Secrets" should deduplicate
	if len(apiNames) != 3 {
		t.Errorf("expected 3 unique API names, got %d: %v", len(apiNames), apiNames)
	}

	// Both Delver variants should map to the same key
	delverOriginals := strippedToOriginal[toLower("Delver of Secrets")]
	if len(delverOriginals) != 2 {
		t.Errorf("expected 2 originals for Delver, got %d", len(delverOriginals))
	}
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i, c := range []byte(s) {
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// Test canonical name mapping logic
func TestCanonicalNameMapping(t *testing.T) {
	// Simulate results returned by Scryfall
	allCards := []*Card{
		{Name: "Delver of Secrets // Insectile Aberration"},
		{Name: "Lightning Bolt"},
		{Name: "Card A // Card B"},
	}

	// Original deck names
	originalNames := []string{
		"Delver of Secrets // Insectile Aberration",
		"Delver of Secrets",
		"Lightning Bolt",
		"Card A / Card B",
	}

	// Rebuild the mapping logic
	strippedToOriginal := make(map[string][]string)
	for _, name := range originalNames {
		stripped := StripDFCName(name)
		key := toLower(stripped)
		strippedToOriginal[key] = append(strippedToOriginal[key], name)
	}

	canonicalLower := make(map[string]string)
	for _, card := range allCards {
		frontLower := toLower(StripDFCName(card.Name))
		canonicalLower[frontLower] = card.Name
	}

	nameMap := make(map[string]string)
	for _, originals := range strippedToOriginal {
		for _, orig := range originals {
			key := toLower(StripDFCName(orig))
			if canonical, ok := canonicalLower[key]; ok {
				nameMap[orig] = canonical
			}
		}
	}

	// Both Delver variants should map to Scryfall's canonical name
	if nameMap["Delver of Secrets // Insectile Aberration"] != "Delver of Secrets // Insectile Aberration" {
		t.Errorf("full DFC name mapping = %q", nameMap["Delver of Secrets // Insectile Aberration"])
	}
	if nameMap["Delver of Secrets"] != "Delver of Secrets // Insectile Aberration" {
		t.Errorf("front-only name mapping = %q", nameMap["Delver of Secrets"])
	}

	// Single-slash variant should map to canonical double-slash
	if nameMap["Card A / Card B"] != "Card A // Card B" {
		t.Errorf("single-slash mapping = %q", nameMap["Card A / Card B"])
	}

	if nameMap["Lightning Bolt"] != "Lightning Bolt" {
		t.Errorf("simple name mapping = %q", nameMap["Lightning Bolt"])
	}
}

// Test not-found mapping back to original names
func TestNotFoundNameMapping(t *testing.T) {
	originalNames := []string{
		"Fake Card A // Fake Card B",
		"Another Fake",
	}

	strippedToOriginal := make(map[string][]string)
	for _, name := range originalNames {
		stripped := StripDFCName(name)
		key := toLower(stripped)
		strippedToOriginal[key] = append(strippedToOriginal[key], name)
	}

	// Simulate: "Fake Card A" not found by Scryfall
	notFoundStripped := []string{"Fake Card A"}

	var notFound []string
	for _, stripped := range notFoundStripped {
		key := toLower(stripped)
		if originals, ok := strippedToOriginal[key]; ok {
			notFound = append(notFound, originals...)
		} else {
			notFound = append(notFound, stripped)
		}
	}

	if len(notFound) != 1 {
		t.Fatalf("expected 1 not-found, got %d", len(notFound))
	}
	if notFound[0] != "Fake Card A // Fake Card B" {
		t.Errorf("not-found = %q, want original DFC name", notFound[0])
	}
}

// =====================
// MEDIUM: Chunking Logic
// =====================

func TestBulkLookupChunking(t *testing.T) {
	// Generate 200 card names to test chunking at 75
	var names []string
	for i := 0; i < 200; i++ {
		names = append(names, "Card "+string(rune('A'+i%26))+string(rune('0'+i/26)))
	}

	// Verify chunking produces correct chunk sizes
	var chunks []int
	for i := 0; i < len(names); i += BulkLookupLimit {
		end := i + BulkLookupLimit
		if end > len(names) {
			end = len(names)
		}
		chunks = append(chunks, end-i)
	}

	if len(chunks) != 3 {
		t.Errorf("expected 3 chunks for 200 cards, got %d", len(chunks))
	}
	if chunks[0] != 75 {
		t.Errorf("chunk 0 = %d, want 75", chunks[0])
	}
	if chunks[1] != 75 {
		t.Errorf("chunk 1 = %d, want 75", chunks[1])
	}
	if chunks[2] != 50 {
		t.Errorf("chunk 2 = %d, want 50", chunks[2])
	}
}
