package app

import (
	"app/internal/db"
	"app/internal/deck"
	"app/internal/scryfall"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// syncDeckCards fetches Scryfall data for all cards in a deck.
func (a *App) syncDeckCards(cards []deck.Card, deckSlug string) []string {
	if a.config == nil {
		log.Println("[Sync] No config, skipping Scryfall sync")
		return nil
	}

	proxySet := make(map[string]bool)
	var names []string
	for i := range cards {
		names = append(names, cards[i].Name)
		if cards[i].HasTag(deck.TagProxy) {
			proxySet[strings.ToLower(cards[i].Name)] = true
		}
	}

	if len(names) == 0 {
		return nil
	}

	log.Printf("[Sync] Starting sync for %d cards (%d proxies)", len(names), len(proxySet))

	client := scryfall.NewClient()
	db.ClearUnmatchedCards(deckSlug)

	var toFetch []string
	for _, name := range names {
		cached, err := db.GetCard(name)
		if err == nil && cached == nil {
			toFetch = append(toFetch, name)
		} else if err == nil && cached != nil {
			stale, _ := db.IsStale(name, 24)
			if stale {
				toFetch = append(toFetch, name)
			}
		}
	}

	log.Printf("[Sync] Cards to fetch from Scryfall: %d", len(toFetch))

	var notFoundNames []string
	if len(toFetch) > 0 {
		log.Println("[Sync] Calling Scryfall API...")
		scCards, notFound, _, err := client.FetchCardsByNames(toFetch)
		if err != nil {
			log.Printf("[Sync] Scryfall API error: %v", err)
		} else {
			log.Printf("[Sync] Got %d cards from Scryfall, %d not found", len(scCards), len(notFound))

			for _, sc := range scCards {
				cached := db.FromScryfallCard(sc)
				db.UpsertCard(cached)

				// Download front face image
				frontImg := sc.FrontFaceImage()
				if frontImg != "" {
					cacheDir := filepath.Join(a.config.AppDataDir, "images", "cache")
					imgSlug := sc.Name + "-" + sc.SetCode
					_, err := scryfall.DownloadImage(frontImg, cacheDir, imgSlug)
					if err != nil {
						log.Printf("[Sync] Image download failed for %s: %v", sc.Name, err)
					}
				}

				// Download back face image for DFCs
				backImg := sc.BackFaceImage()
				if backImg != "" {
					cacheDir := filepath.Join(a.config.AppDataDir, "images", "cache")
					imgSlug := sc.Name + "-" + sc.SetCode + "-back"
					_, err := scryfall.DownloadImage(backImg, cacheDir, imgSlug)
					if err != nil {
						log.Printf("[Sync] Back image download failed for %s: %v", sc.Name, err)
					}
				}
			}

			for _, name := range notFound {
				log.Printf("[Sync] Card not found: %s", name)
				db.AddUnmatchedCard(name, deckSlug)
				notFoundNames = append(notFoundNames, name)
			}
		}
	} else {
		log.Println("[Sync] All cards already cached")
	}

	cachedCount := 0
	for i, name := range names {
		cached, err := db.GetCard(name)
		if err == nil && cached != nil {
			isProxy := proxySet[strings.ToLower(name)]
			sfData := &deck.ScryfallData{
				OracleText:    cached.OracleText,
				TypeLine:      cached.TypeLine,
				ManaCost:      cached.ManaCost,
				CMC:           cached.CMC,
				ImageURI:      cached.ImageURI,
				BackImageURI:  cached.BackImageURI,
				IsDoubleFaced: cached.IsDoubleFaced,
				ColorIdentity: cached.ColorIdentity,
			}
			if !isProxy {
				sfData.PriceUSD = cached.PriceUSD.String
				sfData.PriceUSDFoil = cached.PriceUSDFoil.String
			}
			cards[i].Scryfall = sfData
			cachedCount++
		}
	}

	log.Printf("[Sync] Loaded Scryfall data for %d/%d cards", cachedCount, len(names))

	// Cache commander image to deck folder
	a.cacheCommanderImage(cards, deckSlug)

	return notFoundNames
}

// cacheCommanderImage saves the commander's image to the deck's images/ folder.
func (a *App) cacheCommanderImage(cards []deck.Card, deckSlug string) {
	if a.config == nil || !a.config.HasActiveCollection() {
		return
	}

	// Find the first commander with an image
	for _, c := range cards {
		if !c.HasTag(deck.TagCommander) {
			continue
		}
		if c.Scryfall == nil || c.Scryfall.ImageURI == "" {
			continue
		}

		deckImagesDir := filepath.Join(a.config.DecksDir(), deckSlug, "images")
		destPath := filepath.Join(deckImagesDir, "commander.jpg")

		// Skip if already cached locally
		if fileExists(destPath) {
			log.Printf("[Sync] Commander image already cached: %s", destPath)
			return
		}

		// Download from Scryfall URL
		_, err := scryfall.DownloadImage(c.Scryfall.ImageURI, deckImagesDir, "commander")
		if err != nil {
			log.Printf("[Sync] Commander image cache failed for %s: %v", c.Name, err)
		} else {
			log.Printf("[Sync] Cached commander image for %s → %s", c.Name, destPath)
		}
		return
	}
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
