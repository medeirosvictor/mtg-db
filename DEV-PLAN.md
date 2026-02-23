# DEV-PLAN.md — MTG Desktop App

> A lightweight, fully local desktop app for tracking personal MTG Commander decks, collection, wishlists, prices, and proxy planning. Built with **Wails (Go)** + **Svelte** + **SQLite**.

---

## Principles

- **Plain text is king.** The existing `decks/*/deck.txt`, `info.md`, `wishlist.txt` files remain the source of truth. The app reads and writes them — never replaces them.
- **Offline-first.** The app works without internet after first sync. Scryfall data and images are cached locally.
- **No bloat.** No accounts, no cloud, no social features, no ads. Just your cards.
- **Proxy-aware.** First-class support for marking cards as proxy vs real. Price thresholds for auto-suggesting proxies.
- **Single binary.** Wails produces one executable (~10MB) using the OS native webview. No Electron. No browser dependency.

---

## Tech Stack

| Layer | Choice | Why |
|-------|--------|-----|
| Desktop framework | **Wails v2** | Native webview, single binary, Go backend with JS frontend bindings |
| Backend | **Go** | Fast, simple, great for file I/O, HTTP clients, SQLite, concurrency |
| Frontend | **Svelte + Vite** (not SvelteKit) | Tiny, fast, compiles away — perfect for a desktop UI with no server |
| Database | **SQLite** | Card metadata/price cache only — not the source of truth |
| Card data API | **Scryfall** (free, no auth) | Card metadata, images, prices, set info |
| Proxy art API | **MPC Autofill** (optional) | Community-uploaded proxy art for MakePlayingCards printing |
| Search | **Built-in fuzzy match** | `strings.Contains` / trigram in Go, or Fuse.js on frontend |

---

## Current State of the Repo

```
decks/
  abzan-desert/      100 cards  ✅ Owned     (Hazezon — Abzan Lands)
  avatar-ally/        102 cards  📋 Planned   (Tazri — 5C Allies)
  desert-dune/         99 cards  ✅ Owned     (Yuma — Naya Landfall)
  finalfantasy-voltron/133 cards 📋 Planned   (Cloud — Esper Voltron)
  jumpscare/          100 cards  ✅ Owned     (Arixmethes — Simic Big)
  lotr-aragorn/        99 cards  ✅ Owned     (Aragorn — 4C Humans)
  sultai-rogues/      122 cards  📦 Disassembled (Ukkima — Sultai Rogues)
  warhammer-spellslinger/106 cards ✅ Owned   (Lilah — Izzet Spells)
wishlists/
  master-purchase-list.txt       226 cards across planned decks
history/
  previous-order.txt
scripts/
  find-overlaps.sh               Bash — finds shared cards across decks
  validate-decks.sh              Bash — checks deck sizes
```

**Card format** (varies slightly between decks — parser must handle all):
```
1 Card Name                          ← qty + name (no "x")
1x Card Name                        ← qty + "x" + name
1 Card Name (SET) 123               ← with set code + collector number
1x Card Name (SET) 123 *F*          ← with foil marker
1 Card A / Card B (SET) 123         ← double-faced cards
```

---

## Phase 0 — Project Setup ✅

**Goal:** Wails project scaffolded, builds, opens a window, reads deck files.

- [x] Install Wails CLI, Go, Node, pnpm
- [x] `wails init -n mtg-db -t svelte-ts` (project lives at repo root, not nested)
- [x] Project structure:
  ```
  main.go
  app.go                    ← Wails app struct, methods bound to frontend
  internal/
    deck/                   ← Deck file parser (read/write .txt + info.md)
    config/                 ← App config (data directory path, preferences)
    db/                     ← (placeholder for Phase 1 SQLite)
  frontend/                 ← Svelte + Vite app
    src/
      App.svelte            ← Client-side router
      lib/                  ← Types, color utils
      views/                ← DeckList, DeckView
      components/           ← DeckCard, ColorPips, StatusBadge
  build/
  decks/                    ← Existing deck files (unchanged)
  wishlists/
  history/
  data/                     ← App-generated data (auto-created on launch)
    images/cache/           ← Scryfall images
    images/custom/          ← User-provided global images
    images/mpc/             ← MPC Autofill cached images
  ```
- [x] Deck parser in Go:
  - Parses all format variants (`1 Name`, `1x Name`, `1x Name (SET) 123`, `*F*`, DFCs, `{num}`, hyphenated collector numbers)
  - Reads `info.md` for commander, colors, status, strategy, universe
  - Reads `wishlist.txt`
  - Handles mixed formats within the same file
  - **14 unit tests + integration tests: 861 cards, 8 decks, 0 parse errors**
- [x] Bind `GetAllDecks()`, `GetDeck(slug)`, `ReloadDecks()` to frontend
- [x] Frontend: deck list dashboard with color pips, status badges, card counts, stats
- [x] Frontend: deck detail view with sorted card table and wishlist section
- [x] Dark theme (Catppuccin Mocha-inspired)
- [x] Builds and runs on Windows — **10MB binary**
- [x] pnpm for frontend package management

---

## Phase 1 — Core MVP (Deck Viewer + Prices)

**Goal:** Browse your decks visually with card images and prices. This is the "I can actually use this" milestone.

### 1A — Scryfall Integration + SQLite Cache

- [ ] Scryfall API client in Go:
  - `GET /cards/named?fuzzy=<name>` — single card lookup
  - `GET /cards/named?exact=<name>` — exact match
  - `POST /cards/collection` — bulk lookup (up to 75 cards per request)
  - Rate limiting: max 10 req/sec, 50-100ms between requests
- [ ] SQLite schema for card cache:
  ```sql
  CREATE TABLE cards (
    name TEXT PRIMARY KEY,
    scryfall_id TEXT,
    oracle_text TEXT,
    type_line TEXT,
    color_identity TEXT,   -- JSON array
    mana_cost TEXT,
    cmc REAL,
    set_code TEXT,
    collector_number TEXT,
    image_uri TEXT,         -- Scryfall image URL
    price_usd TEXT,
    price_usd_foil TEXT,
    price_eur TEXT,
    legalities TEXT,        -- JSON object
    updated_at DATETIME
  );
  ```
- [ ] On app launch / deck load: bulk-sync deck cards against Scryfall → cache in SQLite
- [ ] Price staleness: re-fetch prices if cache is older than 24 hours (configurable)
- [ ] Image download: fetch `image_uris.normal` (488×680, ~60KB each) → save to `data/images/cache/`

### 1B — Deck View

- [ ] Deck detail view with two modes:
  - **Grid view** — card images in a responsive grid (4-6 columns)
  - **List view** — compact table: name, type, mana cost, price, qty
- [ ] Per-card info on hover/click: oracle text, type line, mana cost, price, set, legality
- [ ] Deck summary header:
  - Commander name + image
  - Color identity icons
  - Total card count (highlight if ≠ 100)
  - Total deck price (USD)
  - Status badge (Owned / Planned / Disassembled)
- [ ] Category grouping: sort cards by type (Creatures, Instants, Sorceries, Artifacts, Enchantments, Lands, etc.)
- [ ] Mana curve visualization (simple bar chart)

### 1C — Deck List (Home Screen)

- [ ] Dashboard showing all decks as cards/tiles:
  - Deck name, commander image, color identity, card count, total price, status
- [ ] Click a deck → navigate to deck detail view
- [ ] Quick stats: total collection value, total decks, total unique cards

### 1D — Search

- [ ] Global search bar (always accessible, keyboard shortcut `Ctrl+K`)
- [ ] Fuzzy match across all decks and wishlists
- [ ] Results show: card name, which deck(s) it's in, quantity, price
- [ ] Fast — should feel instant, no loading spinners

---

## Phase 2 — Collection Management + Wishlists

**Goal:** Track what you actually own, plan purchases, manage proxies.

### 2A — Collection Tracking

- [ ] New data layer: `data/collection.json` (or SQLite table)
  ```json
  {
    "Sol Ring": { "owned": 3, "proxy": false },
    "Mana Crypt": { "owned": 0, "proxy": true }
  }
  ```
- [ ] Per-card "owned" quantity — distinct from "in deck" quantity
- [ ] Cross-deck awareness: "Sol Ring is in 5 decks, you own 3 → need 2 proxies"
- [ ] Proxy flag per card: mark as proxy vs real
- [ ] Bulk import: paste a list of owned cards to seed the collection
- [ ] Collection summary view: total cards owned, total value, proxy count

### 2B — Wishlist & Purchase Planning

- [ ] Wishlist view that aggregates:
  - Per-deck `wishlist.txt` files
  - `wishlists/master-purchase-list.txt`
- [ ] Sort by: price (cheapest first), deck, card name, color
- [ ] Tag cards as "buy real" vs "proxy print"
- [ ] Price threshold setting: auto-flag cards above $X as proxy candidates (default: $5)
- [ ] Purchase summary: "You need to buy 47 cards for $123.45 + proxy 12 cards"
- [ ] Export purchase list as text (copy-paste to TCGPlayer/Card Kingdom)

### 2C — Overlap Detection

- [ ] Built-in replacement for `find-overlaps.sh`
- [ ] Visual overlap matrix: which decks share which cards
- [ ] Warning badges on shared cards: "This card is in 3 decks but you own 1"

### 2D — Deck Validation

- [ ] Built-in replacement for `validate-decks.sh`
- [ ] Validate on load: card count, color identity vs commander, card legality
- [ ] Warnings (not errors) — this is casual Commander, not competitive
- [ ] Flag unknown cards (not found on Scryfall)

---

## Phase 3 — Deck Editing + Card Management

**Goal:** Full CRUD without touching text files by hand.

### 3A — Add / Remove / Move Cards

- [ ] Add card to deck: search by name → Scryfall autocomplete → add to `deck.txt`
- [ ] Remove card from deck
- [ ] Move card between decks (removes from source, adds to target, updates both `.txt` files)
- [ ] Move card to/from wishlist
- [ ] Undo/redo for edits

### 3B — Deck Import

- [ ] Paste a decklist in standard format (Moxfield/Archidekt/MTGO):
  ```
  1x Avenger of Zendikar (ZNR) 178
  1x Sol Ring (C21) 263
  3x Forest
  ```
- [ ] Parser validates each card against Scryfall
- [ ] Flags unknown cards with fuzzy suggestions ("Did you mean...?")
- [ ] Import from URL (Moxfield, Archidekt) via their public APIs — stretch goal

### 3C — Specific Printings

- [ ] When adding a card, show all available printings from Scryfall
- [ ] Display set symbol, collector number, card image for each printing
- [ ] Selected printing saved to `deck.txt` in `(SET) 123` format
- [ ] Bulk "upgrade printings" — choose preferred sets for your whole deck

---

## Phase 4 — Custom Images & Proxy Art

**Goal:** Full image customization — local files, drag-and-drop, MPC Autofill integration.

### 4A — Local Image Overrides

- [ ] Image resolution chain:
  ```
  1. decks/<deck>/images/<card-slug>.{jpg,png,webp}  → per-deck override
  2. data/images/custom/<card-slug>.{jpg,png,webp}    → global custom image
  3. data/images/mpc/<card-slug>--<id>.jpg            → MPC Autofill cached
  4. data/images/cache/<card-slug>.jpg                → Scryfall cached
  5. Scryfall API (live fetch if online)
  6. Placeholder (generic card back)
  ```
- [ ] Slug format: `avenger-of-zendikar.jpg` (lowercase, hyphens, strip punctuation)
- [ ] Drag & drop: drag an image onto a card → app copies it to the right folder
- [ ] File picker: "Use Local Image" button → native file dialog
- [ ] Manual: just drop a file in the folder — app detects it (file watcher or refresh)

### 4B — MPC Autofill Integration (Optional, Online)

- [ ] Settings: configure MPC Autofill backend URL (community instances)
- [ ] "Browse Proxy Art" button per card:
  1. `POST /2/editorSearch/` — search for community-uploaded art versions
  2. Display thumbnail grid from their image CDN
  3. User picks one → download and cache to `data/images/mpc/`
  4. Image used offline from then on
- [ ] Rate limiting: 1 request per 100ms (match their desktop tool behavior)
- [ ] Source attribution: show which community source contributed the image
- [ ] Browse card backs too (`/2/cardbacks/`)

---

## Phase 5 — Nice-to-Have Features

Everything below is post-MVP polish. Build if/when it's useful.

### UI / UX
- [ ] **Dark/light theme toggle** — dark by default (it's an MTG app, come on)
- [ ] **Keyboard-driven navigation** — `j/k` to move through cards, `Enter` to expand, `Esc` to go back
- [ ] **Card zoom** — click/hover a card for a large high-res popup
- [ ] **Deck color pie chart** — visual breakdown of color distribution
- [ ] **Mana curve chart** — bar chart of CMC distribution
- [ ] **Card category breakdown** — creatures vs spells vs lands pie chart
- [ ] **Drag & drop deck building** — drag cards between deck zones visually
- [ ] **Side-by-side deck comparison** — compare two decks visually
- [ ] **Print-friendly deck list** — export a clean formatted decklist for paper reference

### Data & Sync
- [ ] **MTGJson bulk import** — download full card database for instant offline lookups
- [ ] **Price history tracking** — store daily prices in SQLite, show price trend per card/deck over time
- [ ] **Price alerts** — "Sol Ring dropped below $2" (check on app launch)
- [ ] **Multi-format price display** — USD, EUR, foil vs non-foil toggle
- [ ] **Git integration** — since the repo is git-tracked, show deck change history (cards added/removed over time)
- [ ] **Backup/restore** — export all app data (collection, preferences, cache) as a single zip

### Collection
- [ ] **Binder view** — visual binder page layout (3x3 grid per page) for your full collection
- [ ] **Trade binder** — mark cards as "for trade" and generate a trade list
- [ ] **Inventory scanner** — paste TCGPlayer/Card Kingdom order confirmation → auto-add to collection
- [ ] **Stats dashboard** — most expensive card, average deck cost, total collection value over time

### Proxy Workflow
- [ ] **MPC order builder** — select cards to proxy → generate an MPC Autofill-compatible XML project file
- [ ] **Proxy print sheet** — lay out selected card images in a printable 3x3 grid PDF (for home printing)
- [ ] **Proxy cost estimator** — "Printing 108 cards on MPC ≈ $X at Y cardstock"

### Integration
- [ ] **Moxfield import** — paste a Moxfield deck URL → import the decklist
- [ ] **Archidekt import** — same as above
- [ ] **Scryfall syntax search** — support Scryfall's query syntax (e.g., `c:green cmc<3 t:creature`)
- [ ] **EDHRec integration** — show card recommendations for your commander (if they have a public API)

---

## API Reference

### Scryfall (free, no auth)

| Endpoint | Method | Use |
|----------|--------|-----|
| `/cards/named?fuzzy=<name>` | GET | Single card fuzzy lookup |
| `/cards/named?exact=<name>` | GET | Exact name lookup |
| `/cards/search?q=<query>` | GET | Full search with Scryfall syntax |
| `/cards/<id>` | GET | Card by Scryfall ID |
| `/cards/collection` | POST | Bulk lookup (up to 75 cards) |

- Docs: https://scryfall.com/docs/api
- Rate limit: 10 requests/second, be polite
- Card images: `card.image_uris.normal` (488×680)
- Prices included in card objects: `prices.usd`, `prices.usd_foil`, `prices.eur`

### MPC Autofill (community-hosted, no auth)

| Endpoint | Method | Use |
|----------|--------|-----|
| `/2/editorSearch/` | POST | Search community card art by name |
| `/2/exploreSearch/` | POST | Browse/filter full image database |
| `/2/cards/` | POST | Fetch card metadata by identifier (max 1000) |
| `/2/sources/` | GET | List image sources (community contributors) |
| `/2/DFCPairs/` | GET | Double-faced card pairings |
| `/2/cardbacks/` | POST | Search card back designs |
| `/2/languages/` | GET | Available languages |
| `/2/tags/` | GET | Card tags/categories |
| `/2/info/` | GET | Server metadata |

- Repo: https://github.com/chilli-axe/mpc-autofill
- Frontend: https://mpcautofill.github.io
- Backend: user-configured (each community hosts their own instance)
- Images served via Cloudflare CDN: `<CDN>/images/google_drive/{small|large}/<identifier>.jpg`
- Rate limit: 1 request per 100ms
- Card identifiers = Google Drive file IDs

---

## File Ownership

| File/Directory | Owner | App behavior |
|---------------|-------|-------------|
| `decks/*/deck.txt` | User (plain text) | Read + write (when editing) |
| `decks/*/info.md` | User (plain text) | Read + write |
| `decks/*/wishlist.txt` | User (plain text) | Read + write |
| `decks/*/images/` | User (manual) | Read only — user drops files here |
| `wishlists/` | User (plain text) | Read + write |
| `history/` | User (plain text) | Read only |
| `data/cards.db` | App (generated) | SQLite cache — can be deleted and rebuilt |
| `data/images/cache/` | App (generated) | Scryfall images — can be deleted and re-fetched |
| `data/images/custom/` | User (manual) | Global image overrides — user drops files here |
| `data/images/mpc/` | App (generated) | MPC Autofill cached art — can be deleted |
| `data/collection.json` | App (generated) | Collection/ownership data — important, back up |

---

## Milestones

| Milestone | What you get | Replaces |
|-----------|-------------|----------|
| **Phase 0 done** | App opens, reads your decks, shows a list | Manual file browsing |
| **Phase 1 done** | Visual deck browser with card images + prices | Scryfall manual lookups |
| **Phase 2 done** | Collection tracking, wishlists, proxy planning, overlap detection | `find-overlaps.sh`, `validate-decks.sh`, spreadsheets |
| **Phase 3 done** | Full deck editor, import, printing selection | Hand-editing `.txt` files |
| **Phase 4 done** | Custom art, MPC Autofill proxy art browser | Manually saving images |
| **Phase 5 done** | You've gone too far. Touch grass. Play some Commander. | Your social life |
