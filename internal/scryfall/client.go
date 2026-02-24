package scryfall

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	lastRequest time.Time
	rateLimiter chan struct{} = make(chan struct{}, 1)
)

func init() {
	// Initialize the rate limiter so it's not blocking
	rateLimiter <- struct{}{}
}

// Card represents the essential card data from Scryfall.
type Card struct {
	Name            string   `json:"name"`
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
	ImageURIs struct {
		Small      string `json:"small"`
		Normal     string `json:"normal"`
		Large      string `json:"large"`
		PNG        string `json:"png"`
		ArtCrop    string `json:"art_crop"`
		BorderCrop string `json:"border_crop"`
	} `json:"image_uris,omitempty"`
	Legalities     map[string]string `json:"legalities"`
	Reserved       bool              `json:"reserved"`
	FoundInBooster bool              `json:"found_in_booster"`
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
	http *http.Client
}

// NewClient creates a new Scryfall API client
func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// waitForRateLimit ensures we don't exceed Scryfall's rate limits
func (c *Client) waitForRateLimit() {
	<-rateLimiter
	elapsed := time.Since(lastRequest)
	if elapsed < RateLimitMs*time.Millisecond {
		time.Sleep(RateLimitMs*time.Millisecond - elapsed)
	}
	lastRequest = time.Now()
	rateLimiter <- struct{}{}
}

// FetchCardByName fetches a single card by name (fuzzy match)
func (c *Client) FetchCardByName(name string) (*Card, error) {
	c.waitForRateLimit()

	url := fmt.Sprintf("%s/cards/named?fuzzy=%s", BaseURL, name)
	log.Printf("[Scryfall] GET %s", url)
	req, err := http.NewRequest("GET", url, nil)
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

// FetchCardByExactName fetches a single card by exact name match
func (c *Client) FetchCardByExactName(name string) (*Card, error) {
	c.waitForRateLimit()

	url := fmt.Sprintf("%s/cards/named?exact=%s", BaseURL, name)
	log.Printf("[Scryfall] GET %s", url)
	req, err := http.NewRequest("GET", url, nil)
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

// FetchCardsByNames fetches multiple cards by name (bulk lookup, max 75)
func (c *Client) FetchCardsByNames(names []string) ([]*Card, []string, error) {
	if len(names) == 0 {
		return nil, nil, nil
	}

	// Process in chunks of 75
	var allCards []*Card
	var notFound []string

	for i := 0; i < len(names); i += BulkLookupLimit {
		end := i + BulkLookupLimit
		if end > len(names) {
			end = len(names)
		}
		chunk := names[i:end]

		cards, missing, err := c.fetchChunk(chunk)
		if err != nil {
			return nil, nil, err
		}
		allCards = append(allCards, cards...)
		notFound = append(notFound, missing...)
	}

	return allCards, notFound, nil
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

	log.Printf("[Scryfall] Request body: %s", string(body))
	log.Printf("[Scryfall] Identifiers count: %d", len(identifiers))

	req, err := http.NewRequest("POST", BaseURL+"/cards/collection", bytes.NewReader(body))
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
