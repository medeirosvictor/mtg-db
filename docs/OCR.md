# Card Scanning via Camera — Implementation Plan

## Goal

Point your camera at a physical MTG card → app identifies it → adds 1x to the current deck. Everything runs locally. No cloud APIs.

## Scope

- Auto-detect: scan continuously, add after 3 consistent readings
- Quantity is always 1x (no picker)
- Modal stays open after adding (batch mode)
- Card name list bundled locally, refreshable from Settings
- Supports Windows, macOS, and Linux

---

## Architecture

### Frontend (Svelte)

```
ScanCardModal.svelte
  ├── getUserMedia() → live camera stream
  ├── Canvas: crop top 15% of frame (name region)
  ├── Every 200ms: send cropped frame as base64 to backend
  ├── Show live detected text
  └── On 3 consistent detections (≈600ms): auto-add card, 1s cooldown
```

### Backend (Go)

```
App.RecognizeCardText(imageBase64 string) string
  ├── Decode base64 → image
  ├── Windows  → PowerShell + Windows.Media.Ocr
  ├── macOS    → Swift + Vision framework
  └── Linux    → Tesseract CLI (apt/brew install)

App.MatchCardName(text string) []MatchResult
  └── Fuzzy match against bundled oracle-cards names
      ├── Levenshtein distance + prefix scoring
      └── Return top 3 matches with confidence scores

App.AddCards(slug string, cardLines string) string
  └── Already exists, no changes needed
```

---

## OCR Strategy by Platform

### Windows
- Use `Windows.Media.Ocr` via embedded PowerShell snippet
- GPU-accelerated via DirectML
- ~50ms per frame, ~93% accuracy on printed text
- No system deps beyond Windows 10+

**Go → PowerShell flow:**
1. Write cropped frame to `%TEMP%\mtg-scan-<uuid>.png`
2. Run PowerShell script that calls `Windows.Media.Ocr`
3. Parse stdout for OCR text
4. Delete temp file

### macOS
- Use `Vision` framework via embedded Swift helper
- GPU-accelerated via CoreML
- ~50ms per frame, ~93% accuracy

**Go → Swift flow:**
1. Write cropped frame to `/tmp/mtg-scan-<uuid>.png`
2. Run Swift binary that calls Vision
3. Parse stdout for OCR text
4. Delete temp file

**Swift helper** (`cmd/card-ocr/main.swift`):
- Compile once: `swiftc -o card-ocr cmd/card-ocr/main.swift`
- Bundle binary with app

### Linux
- Use `tesseract` CLI
- User must install: `apt install tesseract-ocr` or `brew install tesseract`
- ~300ms per frame, ~85% accuracy

**Go → CLI flow:**
1. Write cropped frame to `/tmp/mtg-scan-<uuid>.png`
2. Run `tesseract <img> stdout -l eng --psm 7`
3. Parse stdout for OCR text
4. Delete temp file

---

## Auto-Detect Loop

```
Frame capture: every 200ms (5 FPS)
    ↓
Crop top 15% (name region) via canvas
    ↓
Send base64 to App.RecognizeCardText()
    ↓
Fuzzy match → top candidate
    ↓
If same card detected 3 frames in a row (≈600ms):
    → Show "Adding: <Card>" flash (300ms)
    → Call App.AddCards(slug, "1x <Card>")
    → 1 second cooldown before next scan
```

---

## Card Name List

- Source: Scryfall bulk data `oracle-cards.json`
- Store in `%APPDATA%/mtg-db/card-names.json`
- Load once at startup, keep in memory
- Download button in Settings with last-updated timestamp
- Format: flat array of strings (`["Lightning Bolt", "Sol Ring", ...]`)

**Matching algorithm:**
1. Normalize: lowercase, strip non-alphanumeric
2. Score: Levenshtein distance + prefix match bonus
3. Return top 3 results above confidence threshold

---

## File Structure

```
mtg-db/
  cmd/
    card-ocr/           ← macOS Swift OCR helper
      main.swift
      card-ocr           ← compiled binary
  internal/
    app/
      app_ocr.go         ← RecognizeCardText, MatchCardName
      app_settings.go     ← Settings view model (card db refresh)
    ocr/
      ocr_windows.go
      ocr_macos.go
      ocr_linux.go
      ocr.go              ← platform dispatcher
      names.go             ← card name matching
  data/                   ← %APPDATA%/mtg-db/
    card-names.json
  frontend/
    src/
      views/Settings/     ← Settings view (refresh db button, last updated)
      views/DeckView/
        ScanCardModal.svelte
```

---

## Effort Breakdown

| Task | Hours | Notes |
|---|---|---|
| Bundle Scryfall oracle-cards.json + load at startup | 1 | Download script, store path |
| `MatchCardName` Go endpoint + fuzzy matching | 2 | Levenshtein, scoring |
| macOS Swift OCR helper (compile, test) | 2 | Vision framework |
| Windows PowerShell OCR wrapper | 1 | Test on Windows |
| Linux Tesseract fallback | 1 | Document install deps |
| Platform dispatcher (`ocr.go`) | 1 | runtime detection |
| `ScanCardModal.svelte` — camera + preview | 3 | getUserMedia, canvas crop |
| Auto-detect loop (debounce, cooldown, flash UI) | 2 | 3-frame consistency |
| Settings view — refresh card db | 1 | Last updated, download button |
| Wire into DeckView toolbar | 30 min | |
| Test on Windows | 1 | Lighting, angles, foils |
| Test on macOS (if available) | 1 | |
| **Total** | **~16 hrs** | |

---

## Detection Confidence

| Source | Per-frame | With 3-frame debounce |
|---|---|---|
| Windows.Media.Ocr | ~50ms | Very fast |
| Vision (macOS) | ~50ms | Very fast |
| Tesseract | ~300ms | Acceptable |

Tesseract on Linux will feel slower than Windows/Mac. Consider adding a quality indicator ("Hold steady...") to set expectations.

---

## Settings Screen

```
Card Database
  Last updated: 2026-03-21
  [Refresh card database] button  ← downloads latest oracle-cards.json
  Size: 2.1 MB
```

---

## Open Questions

- [x] Batch scanning — yes, modal stays open
- [x] Auto-detect — yes, 3-frame consistency
- [x] Quantity — always 1x
- [x] Offline list — bundled locally, refreshable from Settings
- [ ] Language detection — Scryfall has non-English cards; should OCR attempt `en` only, or try to detect language?
- [ ] DFC detection — if user scans a transform card, do we capture the back face name too?

---

## Next Steps (Next Session)

1. Download `oracle-cards.json` from Scryfall
2. Build `MatchCardName` endpoint with fuzzy matching
3. Implement macOS Swift helper (compile `card-ocr`)
4. Implement Windows PowerShell wrapper
5. Test all three platforms
6. Build `ScanCardModal` camera UI
7. Wire debounce loop
