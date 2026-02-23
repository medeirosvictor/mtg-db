# Improvement Ideas — mtg-db app

The goal: a lightweight, fast, local-first app to manage casual Commander decks. Not Moxfield. Not Archidekt. Just something clean that does the job without the bloat.

---

## Core Concept

A terminal UI or small local web app that reads/writes the same plain-text deck files in this repo. The `.txt` files stay the source of truth — the app is just a nice interface on top. No accounts, no cloud, no login walls.

## Feature Ideas

### 1. Fuzzy Card Search (across your collection)

Search all your decks and wishlists instantly with fuzzy matching.

- Type `aven zen` → finds "Avenger of Zendikar" in `abzan-desert` and `desert-dune`
- Show which decks a card is in, how many copies you own total
- Fast — should feel instant even with 1000+ cards
- Libraries: `fzf` for terminal, or a simple search bar in a web UI using something like Fuse.js

### 2. Scryfall Integration

Query [Scryfall API](https://scryfall.com/docs/api) when adding cards or browsing:

- `GET https://api.scryfall.com/cards/named?fuzzy=avenger+of+zen` — fuzzy name lookup
- `GET https://api.scryfall.com/cards/search?q=name:avenger` — full search
- Pull: card image, oracle text, color identity, type line, legality
- **Rate limit:** Scryfall asks for 50-100ms between requests, no auth needed
- Cache results locally (SQLite or JSON files) so repeated lookups are instant

### 3. Price Lookup

Scryfall already includes prices in card objects:

```json
"prices": {
  "usd": "1.23",
  "usd_foil": "4.56",
  "eur": "1.10"
}
```

- Show per-card price when viewing a deck
- Show total deck cost estimate
- Flag expensive cards as proxy candidates (e.g. anything over $5 or a configurable threshold)
- Could also pull from [MTGJson](https://mtgjson.com/) for bulk price data

### 4. Add Cards with Moxfield/Archidekt Syntax

Support pasting in the standard format:

```
1x Avenger of Zendikar (ZNR) 178
1x Sol Ring (C21) 263
3x Forest
```

Parser should handle:
- `1x` or `1` prefix (quantity)
- Optional `(SET) number` for specific printings
- Case-insensitive, trim whitespace
- Validate against Scryfall on paste (flag unknown cards)
- Auto-suggest the correct name if fuzzy match is close

### 5. Deck Management

- List all decks with card count, status (owned/planned/disassembled), total price
- View a deck with card images (pulled from Scryfall) or compact text list
- Move cards between decks, wishlist, and collection
- Validate: flag decks that aren't exactly 100 cards
- Detect duplicates across decks (the `find-overlaps.sh` logic, but built in)

### 6. Wishlist / Purchase Planning

- Per-deck wishlists (already have the files)
- Master purchase list that aggregates all deck wishlists
- Sort by price to plan purchases (cheapest first, or batch by store)
- Mark cards as "proxy" vs "buy real" with a simple tag
- Track purchase history with dates

---

## Tech Stack Options

### Option A: Terminal UI (TUI)

Best if you want to stay in the terminal and keep it truly minimal.

- **Language:** Python or Go
- **TUI framework:** [Textual](https://textual.textualize.io/) (Python) or [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Go)
- **Search:** built-in fuzzy matching
- **Data:** reads/writes the same .txt files directly
- **Pros:** fast, no browser, fits the terminal workflow, portable
- **Cons:** less visual (no card images inline, though iTerm2/Kitty can do images)

### Option B: Local Web App

Best if you want card images and a more visual layout.

- **Language:** TypeScript
- **Framework:** something lean — plain HTML + [htmx](https://htmx.org/) or a small Svelte/Solid app
- **Backend:** Bun or Deno serving local files, or even fully static with JS reading from a local API
- **Data:** same .txt files, or optionally a SQLite db synced from them
- **Pros:** card images, drag-and-drop, better for deck building visually
- **Cons:** more setup, browser dependency

### Option C: Hybrid — CLI + Simple Web Viewer

- CLI for quick operations: `mtg add desert-dune "1x Exploration"`, `mtg search "avenger"`, `mtg price lotr-aragorn`
- Web UI only for visual deck browsing / card image gallery
- Best of both worlds

---

## Suggested MVP (Start Here)

If building this incrementally, I'd start with:

1. **CLI tool** that reads the existing `.txt` files
2. **`mtg search <query>`** — fuzzy search across all decks
3. **`mtg add <deck> "1x Card Name"`** — add a card with Scryfall validation
4. **`mtg deck <name>`** — show deck list with card count and total price
5. **`mtg validate`** — check all decks for 100-card count
6. **`mtg overlaps`** — find shared cards

Then layer on the web viewer or TUI later.

---

## API Reference

### Scryfall (free, no auth)

| Endpoint | Use |
|----------|-----|
| `GET /cards/named?fuzzy=<name>` | Single card fuzzy lookup |
| `GET /cards/named?exact=<name>` | Exact name lookup |
| `GET /cards/search?q=<query>` | Full search with Scryfall syntax |
| `GET /cards/<id>` | Card by Scryfall ID |
| `GET /cards/collection` | Bulk lookup (POST, up to 75 cards) |

- Docs: https://scryfall.com/docs/api
- Rate limit: 10 requests/second, be nice
- Card images: `card.image_uris.normal` (488×680)

### MTGJson (bulk data, free)

- Full card database download: https://mtgjson.com/downloads/all-files/
- Good for offline price lookups and bulk validation
- Updated daily

---

## UX Principles

- **Plain text stays king.** The `.txt` files are always the source of truth. The app reads and writes them. You can always edit by hand.
- **Instant search.** No loading spinners. Fuzzy match should feel like spotlight/fzf.
- **No bloat.** No social features, no deck sharing, no ads, no premium tiers. Just your cards.
- **Proxy-aware.** First-class support for marking cards as "proxy" vs "real." Price thresholds for auto-suggesting proxies.
- **Offline-first.** Cache Scryfall data locally. Should work without internet after first sync.
