import type { Card } from './types';

// Basic land names (including Wastes from Future Sight)
const BASIC_LANDS = ['plains', 'island', 'swamp', 'mountain', 'forest', 'wastes'];

// Tag constants (must match Go backend)
const TAG_SIDEBOARD = 'sideboard';
const TAG_COMMANDER = 'commander';

/**
 * Check if a card is in the sideboard (not main deck).
 */
export function isSideboard(card: Card): boolean {
  return (card.tags || []).includes(TAG_SIDEBOARD);
}

/**
 * Filter cards that are NOT in the sideboard (main deck only).
 */
export function filterMainDeck(cards: Card[]): Card[] {
  return cards.filter(c => !isSideboard(c));
}

/**
 * Filter cards that ARE in the sideboard.
 */
export function filterSideboard(cards: Card[]): Card[] {
  return cards.filter(c => isSideboard(c));
}

/**
 * Fuzzy match text against a query.
 * Returns true if all query words appear in the text.
 */
export function fuzzyMatch(text: string, query: string): boolean {
  const lower = text.toLowerCase();
  const words = query.toLowerCase().split(/\s+/).filter(w => w.length > 0);
  return words.every(word => lower.includes(word));
}

/**
 * Check if a card name is a basic land.
 */
export function isBasicLand(name: string): boolean {
  return BASIC_LANDS.includes(name.toLowerCase());
}

/**
 * Convert a card to its text line representation (for editing).
 * Format: "4x Lightning Bolt (M10) 88 *F* #commander"
 */
export function cardToText(card: Card): string {
  let line = `${card.quantity}x ${card.name}`;
  if (card.setCode) {
    line += ` (${card.setCode})`;
    if (card.collectorNumber) {
      line += ` ${card.collectorNumber}`;
    }
  }
  if (card.foil) {
    line += ' *F*';
  }
  if (card.tags && card.tags.length > 0) {
    for (const tag of card.tags) {
      line += ` #${tag}`;
    }
  }
  return line;
}

/**
 * Sort cards: commanders first, then non-sideboard non-basic lands,
 * then non-sideboard basic lands, then sideboard cards (at the end).
 */
export function sortCards(cards: Card[]): Card[] {
  return [...cards].sort((a, b) => {
    const aCmd = (a.tags || []).includes(TAG_COMMANDER);
    const bCmd = (b.tags || []).includes(TAG_COMMANDER);
    if (aCmd !== bCmd) return aCmd ? -1 : 1;

    const aSide = isSideboard(a);
    const bSide = isSideboard(b);
    if (aSide !== bSide) return aSide ? 1 : -1;
    
    const aBasic = isBasicLand(a.name);
    const bBasic = isBasicLand(b.name);
    if (aBasic !== bBasic) return aBasic ? 1 : -1;
    
    return a.name.localeCompare(b.name);
  });
}

/**
 * Filter cards by search query using fuzzy matching.
 * Searches: name, set code, type line, oracle text, tags
 */
export function filterCards(cards: Card[], query: string): Card[] {
  if (!query.trim()) return cards;
  
  return cards.filter(c => {
    const haystack = [
      c.name,
      c.setCode || '',
      c.scryFall?.typeLine || '',
      c.scryFall?.oracleText || '',
      ...(c.tags || []),
    ].join(' ');
    return fuzzyMatch(haystack, query);
  });
}

/**
 * Calculate total price of cards (excluding proxies and sideboard).
 */
export function calculateTotalPrice(cards: Card[]): number {
  return cards.reduce((sum, card) => {
    // Skip proxies
    const isProxy = (card.tags || []).includes('proxy');
    if (isProxy) return sum;
    // Skip sideboard cards (they're not part of the main deck price)
    if (isSideboard(card)) return sum;
    const price = card.scryFall?.priceUsd ? parseFloat(card.scryFall.priceUsd) : 0;
    return sum + (price * card.quantity);
  }, 0);
}

/**
 * Get commander names from cards.
 */
export function getCommanderNames(cards: Card[]): string {
  const commanders = cards.filter(c => (c.tags || []).includes('commander'));
  return commanders.map(c => c.name).join(' / ');
}
