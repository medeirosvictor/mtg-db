package scryfall

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	BaseURL         = "https://api.scryfall.com"
	BulkLookupLimit = 75
	RateLimitMs     = 100              // milliseconds between requests
	RequestTimeout  = 10 * time.Second // timeout for each request
	UserAgent       = "mtg-db/1.0"     // Required by Scryfall API
)

var (
	lastRequest  time.Time
	rateLimiter  chan struct{} = make(chan struct{}, 1)
	lastReqMu    sync.Mutex    // Protects lastRequest
)

func init() {
	// Initialize the rate limiter so it's not blocking
	rateLimiter <- struct{}{}
}

// ImageURISet represents a set of image URIs at various sizes.
type ImageURISet struct {
	Small      string `json:"small"`
	Normal     string `json:"normal"`
	Large      string `json:"large"`
	PNG        string `json:"png"`
	ArtCrop    string `json:"art_crop"`
	BorderCrop string `json:"border_crop"`
}

// CardFace represents one face of a double-faced card.
type CardFace struct {
	Name       string      `json:"name"`
	ManaCost   string      `json:"mana_cost"`
	TypeLine   string      `json:"type_line"`
	OracleText string      `json:"oracle_text"`
	Colors     []string    `json:"colors"`
	ImageURIs  ImageURISet `json:"image_uris,omitempty"`
}

// Card represents the essential card data from Scryfall.
type Card struct {
	Name            string   `json:"name"`
	Layout          string   `json:"layout"`
	OracleID        string   `json:"oracle_id"`
	OracleText      string   `json:"oracle_text,omitempty"`
	TypeLine        string   `json:"type_line,omitempty"`
	ManaCost        string   `json:"mana_cost,omitempty"`
	CMC             float64  `json:"cmc"`
	Colors          []string `json:"colors"`
	ColorIdentity   []string `json:"color_identity"`
	SetCode         string   `json:"set_code"`
	SetName         string   `json:"set_name"`
	CollectorNumber string   `json:"collector_number"`
	Prices          struct {
		USD     string `json:"usd"`
		USDFoil string `json:"usd_foil"`
		EUR     string `json:"eur"`
		EURFoil string `json:"eur_foil"`
	} `json:"prices"`
	ImageURIs      ImageURISet       `json:"image_uris,omitempty"`
	CardFaces      []CardFace        `json:"card_faces,omitempty"`
	Legalities     map[string]string `json:"legalities"`
	Reserved       bool              `json:"reserved"`
	FoundInBooster bool              `json:"found_in_booster"`
}

// IsDoubleFaced returns true if this card has two faces (transform, modal_dfc, etc.)
func (c *Card) IsDoubleFaced() bool {
	return len(c.CardFaces) >= 2
}

// FrontFaceImage returns the front face image URI, falling back through card_faces[0].
func (c *Card) FrontFaceImage() string {
	if c.ImageURIs.Normal != "" {
		return c.ImageURIs.Normal
	}
	if len(c.CardFaces) > 0 && c.CardFaces[0].ImageURIs.Normal != "" {
		return c.CardFaces[0].ImageURIs.Normal
	}
	return ""
}

// BackFaceImage returns the back face image URI, or empty string if single-faced.
func (c *Card) BackFaceImage() string {
	if len(c.CardFaces) >= 2 {
		return c.CardFaces[1].ImageURIs.Normal
	}
	return ""
}

// FrontOracleText returns oracle text for the front face.
func (c *Card) FrontOracleText() string {
	if c.OracleText != "" {
		return c.OracleText
	}
	if len(c.CardFaces) > 0 {
		return c.CardFaces[0].OracleText
	}
	return ""
}

// FrontManaCost returns mana cost for the front face.
func (c *Card) FrontManaCost() string {
	if c.ManaCost != "" {
		return c.ManaCost
	}
	if len(c.CardFaces) > 0 {
		return c.CardFaces[0].ManaCost
	}
	return ""
}

// FrontTypeLine returns the type line for the front face.
func (c *Card) FrontTypeLine() string {
	if len(c.CardFaces) > 0 {
		return c.CardFaces[0].TypeLine
	}
	return c.TypeLine
}

// StripDFCName returns just the front face name from a "Front // Back" style name.
func StripDFCName(name string) string {
	if idx := strings.Index(name, " // "); idx > 0 {
		return strings.TrimSpace(name[:idx])
	}
	if idx := strings.Index(name, " / "); idx > 0 {
		return strings.TrimSpace(name[:idx])
	}
	return name
}

// CardSearchResponse represents the response from /cards/named
type CardSearchResponse struct {
	Object    string `json:"object"`
	Card      *Card  `json:"data"`
	SearchURI string `json:"search_uri"`
}

// CollectionRequest represents a bulk lookup request
type CollectionRequest struct {
	Identifiers []CollectionIdentifier `json:"identifiers"`
}

// CollectionIdentifier identifies a card in bulk lookup
type CollectionIdentifier struct {
	ID              string `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	Set             string `json:"set,omitempty"`
	CollectorNumber string `json:"collector_number,omitempty"`
}

// CollectionResponse represents the response from /cards/collection
type CollectionResponse struct {
	Object   string                 `json:"object"`
	NotFound []CollectionIdentifier `json:"not_found"`
	Data     []*Card                `json:"data"`
}

// Client wraps HTTP operations for Scryfall API
type Client struct {
	http    *http.Client
	baseURL string // overrides BaseURL const; used in tests
}

// NewClient creates a new Scryfall API client
func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

func (c *Client) getBaseURL() string {
	if c.baseURL != "" {
		return c.baseURL
	}
	return BaseURL
}

// waitForRateLimit ensures we don't exceed Scryfall's rate limits
func (c *Client) waitForRateLimit() {
	<-rateLimiter

	lastReqMu.Lock()
	elapsed := time.Since(lastRequest)
	if elapsed < RateLimitMs*time.Millisecond {
		time.Sleep(RateLimitMs*time.Millisecond - elapsed)
	}
	lastRequest = time.Now()
	lastReqMu.Unlock()

	rateLimiter <- struct{}{}
}

// TODO: Phase 3A - Single card lookup for Scryfall autocomplete
// FetchCardByName fetches a single card by name (fuzzy match)
func (c *Client) FetchCardByName(name string) (*Card, error) {
	c.waitForRateLimit()

	reqURL := fmt.Sprintf("%s/cards/named?fuzzy=%s", c.getBaseURL(), url.QueryEscape(name))
	log.Printf("[Scryfall] GET %s", reqURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		log.Printf("[Scryfall] Request failed: %v", err)
		return nil, fmt.Errorf("failed to fetch card: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[Scryfall] Response status: %d", resp.StatusCode)

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Card not found
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("scryfall API error (%d): %s", resp.StatusCode, string(body))
	}

	var result CardSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[Scryfall] Decode error: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("[Scryfall] Found card: %s", result.Card.Name)
	return result.Card, nil
}

// TODO: Phase 3A - Exact name lookup for Scryfall autocomplete
// FetchCardByExactName fetches a single card by exact name match
func (c *Client) FetchCardByExactName(name string) (*Card, error) {
	c.waitForRateLimit()

	reqURL := fmt.Sprintf("%s/cards/named?exact=%s", c.getBaseURL(), url.QueryEscape(name))
	log.Printf("[Scryfall] GET %s", reqURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		log.Printf("[Scryfall] Request failed: %v", err)
		return nil, fmt.Errorf("failed to fetch card: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[Scryfall] Response status: %d", resp.StatusCode)

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("scryfall API error (%d): %s", resp.StatusCode, string(body))
	}

	var result CardSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[Scryfall] Decode error: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("[Scryfall] Found card: %s", result.Card.Name)
	return result.Card, nil
}

// FetchCardsByNames fetches multiple cards by name (bulk lookup, max 75).
// Names containing " // " or " / " are automatically stripped to the front face name
// for the API call. The returned Card objects use Scryfall's canonical full name.
// The originalNames map maps the original deck name → Scryfall canonical name.
func (c *Client) FetchCardsByNames(names []string) ([]*Card, []string, map[string]string, error) {
	if len(names) == 0 {
		return nil, nil, nil, nil
	}

	// Build mapping: stripped name → list of original names
	// This handles DFC names like "Card A // Card B" → send "Card A" to Scryfall
	strippedToOriginal := make(map[string][]string) // lowered stripped → original names
	var apiNames []string
	seen := make(map[string]bool)

	for _, name := range names {
		stripped := StripDFCName(name)
		key := strings.ToLower(stripped)
		strippedToOriginal[key] = append(strippedToOriginal[key], name)
		if !seen[key] {
			seen[key] = true
			apiNames = append(apiNames, stripped)
		}
	}

	// Process in chunks of 75
	var allCards []*Card
	var notFoundStripped []string

	for i := 0; i < len(apiNames); i += BulkLookupLimit {
		end := i + BulkLookupLimit
		if end > len(apiNames) {
			end = len(apiNames)
		}
		chunk := apiNames[i:end]

		cards, missing, err := c.fetchChunk(chunk)
		if err != nil {
			return nil, nil, nil, err
		}
		allCards = append(allCards, cards...)
		notFoundStripped = append(notFoundStripped, missing...)
	}

	// Build original→canonical name map from results
	nameMap := make(map[string]string) // original name → scryfall canonical name
	canonicalLower := make(map[string]string) // lowered front face → canonical full name
	for _, card := range allCards {
		frontLower := strings.ToLower(StripDFCName(card.Name))
		canonicalLower[frontLower] = card.Name
	}
	for _, originals := range strippedToOriginal {
		for _, orig := range originals {
			key := strings.ToLower(StripDFCName(orig))
			if canonical, ok := canonicalLower[key]; ok {
				nameMap[orig] = canonical
			}
		}
	}

	// Map not-found stripped names back to original names
	var notFound []string
	for _, stripped := range notFoundStripped {
		key := strings.ToLower(stripped)
		if originals, ok := strippedToOriginal[key]; ok {
			notFound = append(notFound, originals...)
		} else {
			notFound = append(notFound, stripped)
		}
	}

	return allCards, notFound, nameMap, nil
}

// fetchChunk handles a single chunk of card lookups
func (c *Client) fetchChunk(names []string) ([]*Card, []string, error) {
	c.waitForRateLimit()

	log.Printf("[Scryfall] Bulk lookup for %d cards, first 3: %v", len(names), names[:min(3, len(names))])

	identifiers := make([]CollectionIdentifier, len(names))
	for i, name := range names {
		identifiers[i] = CollectionIdentifier{Name: name}
	}

	reqBody := CollectionRequest{
		Identifiers: identifiers,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("[Scryfall] Identifiers count: %d", len(identifiers))

	req, err := http.NewRequest("POST", c.getBaseURL()+"/cards/collection", bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		log.Printf("[Scryfall] Request failed: %v", err)
		return nil, nil, fmt.Errorf("failed to fetch cards: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[Scryfall] Bulk lookup response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("[Scryfall] Error response body: %s", string(respBody))
		return nil, nil, fmt.Errorf("scryfall API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result CollectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Collect not found identifiers
	var notFound []string
	for _, nf := range result.NotFound {
		notFound = append(notFound, nf.Name)
	}

	return result.Data, notFound, nil
}
