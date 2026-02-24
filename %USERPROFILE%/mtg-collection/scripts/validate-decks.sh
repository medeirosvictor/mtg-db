#!/bin/bash
# Validate deck sizes — Commander decks should have exactly 100 cards

YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "=== Deck Size Validation ==="
echo ""

for deck_dir in decks/*/; do
    deck_file="$deck_dir/deck.txt"
    [ -f "$deck_file" ] || continue

    name=$(basename "$deck_dir")
    total=0

    while IFS= read -r line; do
        # Extract quantity (number at start of line)
        qty=$(echo "$line" | grep -oE '^[0-9]+' | head -1)
        [ -n "$qty" ] && total=$((total + qty))
    done < "$deck_file"

    if [ "$total" -eq 100 ]; then
        printf "${GREEN}✓ %-30s %3d cards${NC}\n" "$name" "$total"
    elif [ "$total" -gt 100 ]; then
        printf "${RED}✗ %-30s %3d cards (over by %d)${NC}\n" "$name" "$total" $((total - 100))
    else
        printf "${YELLOW}⚠ %-30s %3d cards (short by %d)${NC}\n" "$name" "$total" $((100 - total))
    fi
done

echo ""
echo "=== Wishlist Sizes ==="
for wl in decks/*/wishlist.txt wishlists/*.txt; do
    [ -f "$wl" ] || continue
    total=0
    while IFS= read -r line; do
        qty=$(echo "$line" | grep -oE '^[0-9]+' | head -1)
        [ -n "$qty" ] && total=$((total + qty))
    done < "$wl"
    printf "  %-45s %3d cards\n" "$wl" "$total"
done
