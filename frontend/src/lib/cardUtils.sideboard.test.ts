import { describe, it, expect } from 'vitest';
import {
  filterMainDeck,
  filterSideboard,
  sortCards,
  isValidDeck,
  getNonSideboardCount,
} from '../lib/cardUtils';
import type { Card } from '../lib/types';

describe('cardUtils - Sideboard Support', () => {
  describe('filterMainDeck', () => {
    it('should filter out sideboard cards', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Lightning Bolt' },
        { quantity: 1, name: 'Sol Ring', tags: ['sideboard'] },
        { quantity: 1, name: 'Counterspell', tags: ['sideboard'] },
        { quantity: 1, name: 'Forest' },
      ];

      const mainDeck = filterMainDeck(cards);

      expect(mainDeck.length).toBe(2);
      expect(mainDeck.map(c => c.name)).toEqual(['Lightning Bolt', 'Forest']);
    });

    it('should return empty array when all cards are sideboard', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Sol Ring', tags: ['sideboard'] },
        { quantity: 1, name: 'Counterspell', tags: ['sideboard'] },
      ];

      const mainDeck = filterMainDeck(cards);

      expect(mainDeck.length).toBe(0);
    });

    it('should handle empty array', () => {
      const mainDeck = filterMainDeck([]);
      expect(mainDeck.length).toBe(0);
    });
  });

  describe('filterSideboard', () => {
    it('should filter only sideboard cards', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Lightning Bolt' },
        { quantity: 1, name: 'Sol Ring', tags: ['sideboard'] },
        { quantity: 1, name: 'Counterspell', tags: ['sideboard'] },
        { quantity: 1, name: 'Forest' },
      ];

      const sideboard = filterSideboard(cards);

      expect(sideboard.length).toBe(2);
      expect(sideboard.map(c => c.name)).toEqual(['Sol Ring', 'Counterspell']);
    });

    it('should return empty array when no cards are sideboard', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Lightning Bolt' },
        { quantity: 1, name: 'Forest' },
      ];

      const sideboard = filterSideboard(cards);

      expect(sideboard.length).toBe(0);
    });

    it('should handle cards with multiple tags including sideboard', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Sol Ring', tags: ['commander', 'sideboard'] },
      ];

      const sideboard = filterSideboard(cards);

      expect(sideboard.length).toBe(1);
      expect(sideboard[0].name).toBe('Sol Ring');
    });
  });

  describe('getNonSideboardCount', () => {
    it('should count only non-sideboard cards', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Lightning Bolt' },
        { quantity: 2, name: 'Sol Ring', tags: ['sideboard'] },
        { quantity: 1, name: 'Counterspell', tags: ['sideboard'] },
        { quantity: 3, name: 'Forest' },
      ];

      const count = getNonSideboardCount(cards);

      // 1 + 3 = 4 (sideboard cards don't count, even with quantity > 1)
      expect(count).toBe(4);
    });

    it('should return 0 for empty deck', () => {
      const count = getNonSideboardCount([]);
      expect(count).toBe(0);
    });

    it('should return 0 for all sideboard cards', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Sol Ring', tags: ['sideboard'] },
        { quantity: 10, name: 'Counterspell', tags: ['sideboard'] },
      ];

      const count = getNonSideboardCount(cards);

      expect(count).toBe(0);
    });
  });

  describe('isValidDeck', () => {
    it('should return valid for exactly 100 non-sideboard cards', () => {
      const cards: Card[] = Array.from({ length: 100 }, (_, i) => ({
        quantity: 1,
        name: `Card ${i}`,
      }));

      const result = isValidDeck(cards);

      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should return invalid for 99 cards', () => {
      const cards: Card[] = Array.from({ length: 99 }, (_, i) => ({
        quantity: 1,
        name: `Card ${i}`,
      }));

      const result = isValidDeck(cards);

      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Deck must have exactly 100 cards (currently 99)');
    });

    it('should return invalid for 101 cards', () => {
      const cards: Card[] = Array.from({ length: 101 }, (_, i) => ({
        quantity: 1,
        name: `Card ${i}`,
      }));

      const result = isValidDeck(cards);

      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Deck must have exactly 100 cards (currently 101)');
    });

    it('should ignore sideboard cards in count', () => {
      // 100 main deck + 15 sideboard = valid
      const cards: Card[] = [
        ...Array.from({ length: 100 }, (_, i) => ({
          quantity: 1,
          name: `Main Card ${i}`,
        })),
        ...Array.from({ length: 15 }, (_, i) => ({
          quantity: 1,
          name: `Sideboard Card ${i}`,
          tags: ['sideboard'] as const,
        })),
      ];

      const result = isValidDeck(cards);

      expect(result.valid).toBe(true);
    });

    it('should validate 99 main + sideboard = invalid', () => {
      const cards: Card[] = [
        ...Array.from({ length: 99 }, (_, i) => ({
          quantity: 1,
          name: `Main Card ${i}`,
        })),
        ...Array.from({ length: 20 }, (_, i) => ({
          quantity: 1,
          name: `Sideboard Card ${i}`,
          tags: ['sideboard'] as const,
        })),
      ];

      const result = isValidDeck(cards);

      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Deck must have exactly 100 cards (currently 99)');
    });

    it('should handle commander tag (commander does not count toward 100)', () => {
      // Commander is extra, so 1 commander + 99 regular = 100 valid
      const cards: Card[] = [
        { quantity: 1, name: 'Atraxa, Precipitous Vanguard', tags: ['commander'] },
        ...Array.from({ length: 99 }, (_, i) => ({
          quantity: 1,
          name: `Card ${i}`,
        })),
      ];

      const result = isValidDeck(cards);

      expect(result.valid).toBe(true);
    });
  });

  describe('sortCards with sideboard', () => {
    it('should sort: commanders first, then main deck, then sideboard, then basic lands', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Forest', tags: [] },
        { quantity: 1, name: 'Lightning Bolt', tags: ['sideboard'] },
        { quantity: 1, name: 'Atraxa', tags: ['commander'] },
        { quantity: 1, name: 'Island', tags: [] },
        { quantity: 1, name: 'Sol Ring', tags: ['sideboard'] },
        { quantity: 1, name: 'Counterspell', tags: [] },
      ];

      const sorted = sortCards(cards);

      // Order: commander (Atraxa), non-sideboard non-basic (Counterspell), non-sideboard basic (Forest, Island), sideboard (Lightning Bolt, Sol Ring)
      expect(sorted[0].name).toBe('Atraxa');
      expect(sorted[1].name).toBe('Counterspell');
      // Basic lands should come after non-basic
      const basicLands = sorted.filter(c => c.name === 'Forest' || c.name === 'Island');
      expect(basicLands.length).toBe(2);
      // Sideboard at the end
      const sideboard = sorted.filter(c => c.tags?.includes('sideboard'));
      expect(sideboard.length).toBe(2);
      expect(sorted[sorted.length - 1].tags).toContain('sideboard');
    });
  });
});
