import { describe, it, expect } from 'vitest';
import {
  filterMainDeck,
  filterSideboard,
  sortCards,
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
