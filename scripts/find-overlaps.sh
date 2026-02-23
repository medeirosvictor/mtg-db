#!/bin/bash
# Find cards that appear in multiple decks or in both a deck and the purchase list
# Ignores basic lands

BASICS="^[0-9]+x (Plains|Island|Swamp|Mountain|Forest)$"

echo "=== Cards in Multiple Decks ==="
echo ""

# Extract card names from all decks (strip quantity, set codes, trailing whitespace)
tmp_dir=$(mktemp -d)

for deck_dir in decks/*/; do
    deck_file="$deck_dir/deck.txt"
    [ -f "$deck_file" ] || continue
    name=$(basename "$deck_dir")

    # Normalize: lowercase, strip qty prefix, strip set codes like (SET) 123, trim
    sed -E 's/^[0-9]+x?\s+//; s/\s*\([A-Z0-9]+\)\s*[0-9]*\s*(\*F\*)?//g; s/\s*\{[0-9]+\}//g; s/[[:space:]]*$//' "$deck_file" \
        | tr '[:upper:]' '[:lower:]' \
        | grep -viE '^(plains|island|swamp|mountain|forest)$' \
        | sort -u \
        > "$tmp_dir/$name"
done

# Find overlaps between all pairs
found=0
decks=("$tmp_dir"/*)
for ((i=0; i<${#decks[@]}; i++)); do
    for ((j=i+1; j<${#decks[@]}; j++)); do
        overlap=$(comm -12 "${decks[$i]}" "${decks[$j]}")
        if [ -n "$overlap" ]; then
            name_a=$(basename "${decks[$i]}")
            name_b=$(basename "${decks[$j]}")
            count=$(echo "$overlap" | wc -l | tr -d ' ')
            echo "[$name_a] ↔ [$name_b] — $count shared cards:"
            echo "$overlap" | sed 's/^/  /'
            echo ""
            found=1
        fi
    done
done

[ "$found" -eq 0 ] && echo "No overlaps found between decks."

echo ""
echo "=== Cards in Purchase List AND a Deck ==="
echo ""

# Build purchase list
if [ -f wishlists/master-purchase-list.txt ]; then
    sed -E 's/^[0-9]+x?\s+//; s/\s*\([A-Z0-9]+\)\s*[0-9]*//g; s/\s*\{[0-9]+\}//g; s/[[:space:]]*$//' wishlists/master-purchase-list.txt \
        | tr '[:upper:]' '[:lower:]' \
        | grep -viE '^(plains|island|swamp|mountain|forest)$' \
        | sort -u \
        > "$tmp_dir/purchase"

    for deck_file in "$tmp_dir"/*; do
        name=$(basename "$deck_file")
        [ "$name" = "purchase" ] && continue
        overlap=$(comm -12 "$deck_file" "$tmp_dir/purchase")
        if [ -n "$overlap" ]; then
            count=$(echo "$overlap" | wc -l | tr -d ' ')
            echo "[$name] has $count cards also in purchase list:"
            echo "$overlap" | sed 's/^/  /'
            echo ""
        fi
    done
fi

rm -rf "$tmp_dir"
