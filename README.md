# mtg-db 🃏

Personal MTG Commander collection manager — a lightweight desktop app for tracking decks, wishlists, prices, and proxy planning. Fully local, no bloat.

Built with **Wails (Go)** + **Svelte** + **SQLite**.

## Quick Start

```bash
# Prerequisites: Go 1.23+, Node 18+, pnpm, Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Development (hot reload)
wails dev

# Build production binary
wails build
# → build/bin/mtg-db.exe (~10MB)
```

## Structure

```
decks/<deck-name>/
  deck.txt            # Decklist (format: "1x Card Name" or "1 Card Name (SET) 123")
  info.md             # Commander, colors, strategy notes
  wishlist.txt        # Upgrade ideas for this deck
  images/             # Per-deck image overrides (optional)
wishlists/
  master-purchase-list.txt
history/
  previous-order.txt
data/                 # App-generated (gitignored)
  cards.db            # SQLite cache (Scryfall card data + prices)
  images/cache/       # Scryfall images
  images/custom/      # User-provided global image overrides
  images/mpc/         # MPC Autofill cached proxy art
internal/             # Go backend packages
  deck/               # Deck file parser
  config/             # App configuration
  db/                 # SQLite (coming in Phase 1)
frontend/             # Svelte + Vite frontend
```

## Decks

| Folder | Deck | Status |
|--------|------|--------|
| `warhammer-spellslinger` | Warhammer 40K — Izzet Spellslinger | ✅ Owned |
| `abzan-desert` | Abzan Desert — Lands Matter | ✅ Owned |
| `jumpscare` | Duskmourn — Simic Big Creatures | ✅ Owned |
| `desert-dune` | Thunderjunction — Naya Plants/Landfall | ✅ Owned |
| `lotr-aragorn` | LOTR — 4-Color Human Tribal | ✅ Owned |
| `sultai-rogues` | Sultai Rogues — Unblockable | 📦 Disassembled |
| `avatar-ally` | Avatar TLA — 5-Color Allies | 📋 Planned |
| `finalfantasy-voltron` | Final Fantasy — Esper Voltron/Ninjas | 📋 Planned |

## Card Format

Plain text, always the source of truth:
```
1 Card Name
1x Card Name
1x Card Name (SET) 123
1 Card Name (SET) 23s *F*
```

## Development Plan

See [DEV-PLAN.md](DEV-PLAN.md) for the full roadmap (Phases 0–5).
