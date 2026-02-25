import { describe, it, expect } from 'vitest';
import {
  isBasicLand,
  cardToText,
  sortCards,
  filterCards,
  calculateTotalPrice,
  getCommanderNames,
  fuzzyMatch,
} from '../lib/cardUtils';
import type { Card } from '../lib/types';

describe('cardUtils', () => {
  describe('isBasicLand', () => {
    it('should identify basic lands', () => {
      expect(isBasicLand('Forest')).toBe(true);
      expect(isBasicLand('Island')).toBe(true);
      expect(isBasicLand('Mountain')).toBe(true);
      expect(isBasicLand('Swamp')).toBe(true);
      expect(isBasicLand('Plains')).toBe(true);
      expect(isBasicLand('Wastes')).toBe(true);
    });

    it('should reject non-basic lands', () => {
      expect(isBasicLand('Overgrown Tomb')).toBe(false);
      expect(isBasicLand('Tropical Island')).toBe(false);
      expect(isBasicLand('Lightning Bolt')).toBe(false);
    });

    it('should be case insensitive', () => {
      expect(isBasicLand('FOREST')).toBe(true);
      expect(isBasicLand('forest')).toBe(true);
      expect(isBasicLand('FoReSt')).toBe(true);
    });
  });

  describe('cardToText', () => {
    it('should format card with quantity', () => {
      const card: Card = { quantity: 3, name: 'Lightning Bolt' };
      expect(cardToText(card)).toBe('3x Lightning Bolt');
    });

    it('should include set code', () => {
      const card: Card = { quantity: 1, name: 'Sol Ring', setCode: 'M19' };
      expect(cardToText(card)).toBe('1x Sol Ring (M19)');
    });

    it('should include collector number', () => {
      const card: Card = { quantity: 1, name: 'Sol Ring', setCode: 'M19', collectorNumber: '123' };
      expect(cardToText(card)).toBe('1x Sol Ring (M19) 123');
    });

    it('should include foil marker', () => {
      const card: Card = { quantity: 1, name: 'Sol Ring', foil: true };
      expect(cardToText(card)).toBe('1x Sol Ring *F*');
    });

    it('should include tags', () => {
      const card: Card = { quantity: 1, name: 'Sol Ring', tags: ['commander', 'proxy'] };
      expect(cardToText(card)).toBe('1x Sol Ring #commander #proxy');
    });
  });

  describe('sortCards', () => {
    it('should sort cards by name', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Zephyr' },
        { quantity: 1, name: 'Alpha' },
        { quantity: 1, name: 'Beta' },
      ];
      const sorted = sortCards(cards);
      expect(sorted[0].name).toBe('Alpha');
      expect(sorted[1].name).toBe('Beta');
      expect(sorted[2].name).toBe('Zephyr');
    });

    it('should sort non-basic lands before basic lands', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Forest' },
        { quantity: 1, name: 'Lightning Bolt' },
        { quantity: 1, name: 'Island' },
      ];
      const sorted = sortCards(cards);
      // Sorted: non-basic first, then basic lands alphabetically
      expect(sorted[0].name).toBe('Lightning Bolt');
      expect(sorted[1].name).toBe('Forest');
      expect(sorted[2].name).toBe('Island');
    });
  });

  describe('filterCards', () => {
    const cards: Card[] = [
      { quantity: 1, name: 'Lightning Bolt', tags: ['commander'] },
      { quantity: 2, name: 'Counterspell', tags: ['proxy'] },
      { quantity: 3, name: 'Forest' },
      { quantity: 1, name: 'Sol Ring', tags: ['commander', 'proxy'] },
    ];

    it('should filter by name', () => {
      const result = filterCards(cards, 'bolt');
      expect(result.length).toBe(1);
      expect(result[0].name).toBe('Lightning Bolt');
    });

    it('should filter by tag', () => {
      const result = filterCards(cards, 'commander');
      expect(result.length).toBe(2);
    });

    it('should be case insensitive', () => {
      const result = filterCards(cards, 'LIGHTNING');
      expect(result.length).toBe(1);
    });

    it('should return all cards for empty query', () => {
      const result = filterCards(cards, '');
      expect(result.length).toBe(4);
    });
  });

  describe('calculateTotalPrice', () => {
    it('should calculate total price from cards', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Card1', scryFall: { priceUsd: '10.00' } },
        { quantity: 2, name: 'Card2', scryFall: { priceUsd: '5.50' } },
      ];
      const total = calculateTotalPrice(cards);
      expect(total).toBe(21.00); // 10 + 2*5.50
    });

    it('should handle missing prices', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Card1', scryFall: { priceUsd: '10.00' } },
        { quantity: 1, name: 'Card2' },
      ];
      const total = calculateTotalPrice(cards);
      expect(total).toBe(10.00);
    });

    it('should return 0 for empty deck', () => {
      const total = calculateTotalPrice([]);
      expect(total).toBe(0);
    });
  });

  describe('getCommanderNames', () => {
    it('should return commander names from tags', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Teferi, Hero of Dominaria', tags: ['commander'] },
        { quantity: 1, name: 'Lightning Bolt' },
      ];
      expect(getCommanderNames(cards)).toBe('Teferi, Hero of Dominaria');
    });

    it('should return empty string for no commanders', () => {
      const cards: Card[] = [
        { quantity: 1, name: 'Lightning Bolt' },
      ];
      expect(getCommanderNames(cards)).toBe('');
    });
  });

  describe('fuzzyMatch', () => {
    it('should match exact text', () => {
      expect(fuzzyMatch('Lightning Bolt', 'Lightning')).toBe(true);
      expect(fuzzyMatch('Lightning Bolt', 'Bolt')).toBe(true);
    });

    it('should be case insensitive', () => {
      expect(fuzzyMatch('Lightning Bolt', 'LIGHTNING')).toBe(true);
      expect(fuzzyMatch('Lightning Bolt', 'bolt')).toBe(true);
    });

    it('should match multiple words', () => {
      expect(fuzzyMatch('Lightning Bolt', 'lightning bolt')).toBe(true);
      expect(fuzzyMatch('Lightning Bolt', 'lightning bolt ')).toBe(true);
    });

    it('should not match partial words incorrectly', () => {
      // "light" should match "Lightning" but not "Flight"
      expect(fuzzyMatch('Lightning Bolt', 'light')).toBe(true);
    });
  });
});
