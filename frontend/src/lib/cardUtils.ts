import type { Card } from './types';

// Basic land names
const BASIC_LANDS = ['plains', 'island', 'swamp', 'mountain', 'forest'];

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
 * Sort cards: commanders first, then non-basic lands, then basic lands, then alphabetical.
 */
export function sortCards(cards: Card[]): Card[] {
  return [...cards].sort((a, b) => {
    const aCmd = (a.tags || []).includes('commander');
    const bCmd = (b.tags || []).includes('commander');
    if (aCmd !== bCmd) return aCmd ? -1 : 1;
    
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
 * Calculate total price of cards (excluding proxies).
 */
export function calculateTotalPrice(cards: Card[]): number {
  return cards.reduce((sum, card) => {
    const isProxy = (card.tags || []).includes('proxy');
    if (isProxy) return sum;
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
