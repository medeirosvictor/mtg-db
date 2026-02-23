# mtg-db 🃏

Casual MTG collection tracker — commander decks, wishlists, and purchase history. Proxy-friendly, no competitive stuff.

## Structure

```
decks/<deck-name>/
  deck.txt            # Decklist (format: "1x Card Name")
  wishlist.txt        # Upgrade ideas for this deck
  info.md             # Commander, colors, strategy notes
wishlists/
  master-purchase-list.txt
history/
  previous-order.txt
scripts/
  validate-decks.sh   # Check deck sizes
  find-overlaps.sh    # Find shared cards across decks
agent-feedback/
  improvement-ideas.md
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

```
1x Card Name
1x Card Name (SET) 123
```
