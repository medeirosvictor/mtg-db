# DEV-PLAN.md — MTG Desktop App

---

## Core Principles

1. **Test-Driven Development (TDD).** Always write tests first. Verify tests FAIL (Red) before implementing the feature correctly (Green).
2. **Maintainability & Readability.** Code is read more than written. Prioritize clarity over cleverness.
3. **Single Responsibility.** Keep files short. Each file/function/module should have one clear purpose.
4. **Backend First.** Always implement backend functionality before frontend. Solidify the core logic in Go before building UI/UX in Svelte.

---

## Tech Stack

| Layer | Choice |
|-------|--------|
| Desktop framework | Wails v2 |
| Backend | Go |
| Frontend | Svelte + Vite |
| Database | SQLite |
| Card data API | Scryfall |
| Proxy art API | MPC Autofill (optional) |
| Search | Fuzzy match |

---

## Data Flow

**Plain text files are the source of truth:**
- `decks/*/deck.txt` — card list
- `decks/*/info.md` — metadata
- `decks/*/wishlist.txt` — per-deck wishlist

**App data** (`%APPDATA%/mtg-db/`):
- `config.json` — collection paths + preferences
- `cards.db` — SQLite card/price cache
- `images/cache/` — Scryfall card images

---

## Current Repo Structure

```
mtg-db/
  main.go
  go.mod, go.sum
  wails.json
  internal/
    app/            # Wails app handlers
    config/         # Config management
    deck/           # Deck parser
    deckimport/     # Deck import (Moxfield)
    db/             # SQLite cache
    scryfall/       # Scryfall API client
  frontend/
    src/
      App.svelte
      lib/          # Utilities, types, utils
      views/        # Page components
      components/   # Reusable components
  build/
```

---

## Card Format (parser supports)

```
1 Card Name                          ← qty + name
1x Card Name (SET) 123 *F*          ← with set code, collector number, foil
1 Card A / Card B (SET) 123         ← double-faced cards
```

---

## Implemented Features

| Feature | Status |
|---------|--------|
| Wails + Svelte + Go setup | ✅ Done |
| Deck parser (unit + integration tests) | ✅ Done |
| Dark theme (Catppuccin Mocha) | ✅ Done |
| Collection folder separation | ✅ Done |
| Config in `%APPDATA%` | ✅ Done |
| Deck list dashboard | ✅ Done |
| Color pips + status badges | ✅ Done |
| Deck detail view (grid + list) | ✅ Done |
| Card sorting (commander → land → alpha) | ✅ Done |
| Per-card inspect modal | ✅ Done |
| Deck sync with Scryfall | ✅ Done |
| SQLite card/price cache | ✅ Done |
| Image caching | ✅ Done |
| Global search (`Ctrl+K`) | ✅ Done |
| Edit deck title/strategy/status | ✅ Done |
| Import deck (paste) | ✅ Done |
| Export deck | ✅ Done |
| Basic deck editing (add/remove/qty) | ✅ Done |

---

## Remaining Work

### Sideboard Support (Fully Implemented ✅)

- [x] Add `#sideboard` tag support (Go + TS)
- [x] Filter main deck vs sideboard (Go + TS)
- [x] Validation: exactly 100 non-sideboard cards (Go + TS)
- [x] `ToggleCardTag` already supports sideboard via existing API
- [x] Display sideboard cards in separate section in UI (List + Grid views)
- [x] Context menu: "Move to Sideboard" / "Move to Main Deck"

---

### Phase 2 — Collection Management + Wishlists

#### 2A — Collection Tracking

- [ ] `collection.json` or SQLite table for owned quantities
- [ ] Per-card owned count (distinct from deck quantity)
- [ ] Cross-deck awareness: "Sol Ring in 5 decks, owned 3 → need 2 proxies"
- [ ] Proxy flag per card
- [ ] Bulk import owned cards
- [ ] Collection summary: total cards, value, proxy count

#### 2B — Wishlist & Purchase Planning

- [ ] Aggregate wishlists: per-deck + global
- [ ] Sort by: price, deck, name, color
- [ ] Tag: "buy real" vs "proxy print"
- [ ] Auto-flag cards above price threshold (default $5)
- [ ] Purchase summary: "Need 47 cards for $123.45 + proxy 12"
- [ ] Export purchase list as text

#### 2C — Overlap Detection

- [ ] Visual overlap matrix: shared cards across decks
- [ ] Warning badges: "This card is in 3 decks but you own 1"

#### 2D — Deck Validation

- [ ] Validate on load: card count, color identity, legality
- [ ] Flag unknown cards

---

### Phase 3 — Advanced Deck Editing + Card Management

#### 3A — Advanced Card Operations

- [ ] Add card: Scryfall autocomplete → add to `deck.txt`
- [ ] Move card between decks
- [ ] Move card to/from wishlist
- [ ] Undo/redo

#### 3B — Deck Import

- [ ] Import from URL (Moxfield, Archidekt)

#### 3C — Specific Printings

- [ ] Show all printings from Scryfall
- [ ] Display set symbol, collector number, image
- [ ] Bulk "upgrade printings"

---

### Phase 4 — Custom Images & Proxy Art

#### 4A — Local Image Overrides

- [ ] Resolution chain:
  ```
  1. <collection>/decks/<deck>/images/<card-slug>.{jpg,png,webp}
  2. %APPDATA%/mtg-db/images/custom/<card-slug>.{jpg,png,webp}
  3. %APPDATA%/mtg-db/images/mpc/<card-slug>--<id>.jpg
  4. %APPDATA%/mtg-db/images/cache/<card-slug>.jpg
  5. Scryfall API (live)
  6. Placeholder
  ```
- [ ] Slug format: lowercase, hyphens
- [ ] Drag & drop image onto card
- [ ] File picker for local images

#### 4B — MPC Autofill Integration

- [ ] Configure backend URL
- [ ] Browse proxy art → download to cache
- [ ] Rate limiting: 1 req/100ms
- [ ] Source attribution

---

### Phase 5 — Nice-to-Have

#### UI / UX
- [ ] Dark/light theme toggle
- [ ] Keyboard navigation (`j/k`, `Enter`, `Esc`)
- [ ] Card zoom popup
- [ ] Color pie chart
- [ ] Mana curve chart
- [ ] Category breakdown chart
- [ ] Drag & drop deck building
- [ ] Deck comparison
- [ ] Print-friendly deck list

#### Data & Sync
- [ ] MTGJson bulk import
- [ ] Price history tracking
- [ ] Price alerts
- [ ] Multi-format price display
- [ ] Git integration (deck change history)
- [ ] Backup/restore (zip export)

#### Collection
- [ ] Binder view
- [ ] Trade binder
- [ ] Inventory scanner
- [ ] Stats dashboard

#### Proxy Workflow
- [ ] MPC order builder
- [ ] Proxy print sheet (PDF)
- [ ] Proxy cost estimator

#### Integration
- [ ] Scryfall syntax search
- [ ] EDHRec integration

---

## API Reference

### Scryfall

| Endpoint | Method | Use |
|----------|--------|-----|
| `/cards/named?fuzzy=<name>` | GET | Fuzzy lookup |
| `/cards/named?exact=<name>` | GET | Exact lookup |
| `/cards/search?q=<query>` | GET | Full search |
| `/cards/<id>` | GET | By Scryfall ID |
| `/cards/collection` | POST | Bulk (max 75) |

- Rate limit: 10 req/sec
- Images: `card.image_uris.normal` (488×680)

### MPC Autofill

| Endpoint | Method | Use |
|----------|--------|-----|
| `/2/editorSearch/` | POST | Search art |
| `/2/exploreSearch/` | POST | Browse images |
| `/2/cards/` | POST | Fetch metadata |
| `/2/sources/` | GET | List sources |
| `/2/cardbacks/` | POST | Search backs |

- Rate limit: 1 req/100ms

---

## File Ownership

### Collection Folder (user-selected, plain text)

| File/Directory | Behavior |
|---------------|----------|
| `decks/*/deck.txt` | Read + write |
| `decks/*/info.md` | Read + write |
| `decks/*/wishlist.txt` | Read + write |
| `decks/*/images/` | Read only |
| `wishlists/` | Read + write |
| `history/` | Read only |

### App Data (`%APPDATA%/mtg-db/`)

| File/Directory | Behavior |
|---------------|----------|
| `config.json` | Generated |
| `cards.db` | Generated, rebuildable |
| `images/cache/` | Generated, rebuildable |
| `images/custom/` | User manual |
| `images/mpc/` | Generated, rebuildable |
| `collection.json` | Generated, back up |

---

## Milestones

| Phase | Deliverable |
|-------|-------------|
| 1 | ✅ Visual deck browser with images, prices, search |
| 2 | Collection tracking, wishlists, proxy planning, overlap detection |
| 3 | Advanced editor with Scryfall autocomplete, cross-deck moves, import |
| 4 | Custom art, MPC Autofill proxy art browser |
| 5 | Additional polish features |
