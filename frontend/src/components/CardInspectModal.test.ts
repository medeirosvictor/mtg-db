import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import CardInspectModal from './CardInspectModal.svelte';
import type { Card } from '../lib/types';

// =====================
// Card Inspect Modal Tests
// =====================

function makeCard(overrides: Partial<Card> = {}): Card {
  return {
    quantity: 1,
    name: 'Lightning Bolt',
    scryFall: {
      oracleText: 'Lightning Bolt deals 3 damage to any target.',
      typeLine: 'Instant',
      manaCost: '{R}',
      cmc: 1,
      imageUri: 'https://example.com/bolt.jpg',
      priceUsd: '1.50',
      priceUsdFoil: '3.00',
    },
    tags: [],
    ...overrides,
  };
}

describe('CardInspectModal', () => {
  describe('rendering', () => {
    it('should render the card name in the header', () => {
      const card = makeCard();
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('Lightning Bolt');
    });

    it('should render the type line', () => {
      const card = makeCard();
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('Instant');
    });

    it('should render oracle text', () => {
      const card = makeCard();
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('Lightning Bolt deals 3 damage to any target.');
    });

    it('should render the card image', () => {
      const card = makeCard();
      const { container } = render(CardInspectModal, { props: { card } });
      const img = container.querySelector('img');
      expect(img).not.toBeNull();
      expect(img?.getAttribute('src')).toBe('https://example.com/bolt.jpg');
      expect(img?.getAttribute('alt')).toBe('Lightning Bolt');
    });

    it('should render price in USD', () => {
      const card = makeCard();
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('$1.50');
    });

    it('should render foil price', () => {
      const card = makeCard();
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('$3.00');
    });

    it('should render CMC', () => {
      const card = makeCard();
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('1'); // CMC value
    });

    it('should render quantity', () => {
      const card = makeCard({ quantity: 3 });
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('3×');
    });

    it('should render set code when present', () => {
      const card = makeCard({ setCode: 'M19', collectorNumber: '126' });
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('M19');
      expect(container.textContent).toContain('#126');
    });

    it('should show foil badge when card is foil', () => {
      const card = makeCard({ foil: true });
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('Foil');
    });

    it('should render placeholder when no image available', () => {
      const card = makeCard({
        scryFall: {
          oracleText: 'Some text',
          typeLine: 'Instant',
          manaCost: '{R}',
          cmc: 1,
        },
      });
      const { container } = render(CardInspectModal, { props: { card } });
      const img = container.querySelector('img');
      expect(img).toBeNull();
      // Should show the 2-letter placeholder
      expect(container.textContent).toContain('LI');
    });
  });

  describe('mana cost rendering', () => {
    it('should render mana pips for single color', () => {
      const card = makeCard();
      const { container } = render(CardInspectModal, { props: { card } });
      // Should have a mana pip with "R"
      const pips = container.querySelectorAll('span[title="R"]');
      expect(pips.length).toBeGreaterThan(0);
    });

    it('should render multiple mana pips', () => {
      const card = makeCard({
        scryFall: {
          ...makeCard().scryFall!,
          manaCost: '{2}{U}{U}',
        },
      });
      const { container } = render(CardInspectModal, { props: { card } });
      const pips = container.querySelectorAll('span[title="U"]');
      expect(pips.length).toBe(2);
    });

    it('should handle empty mana cost', () => {
      const card = makeCard({
        scryFall: {
          ...makeCard().scryFall!,
          manaCost: '',
        },
      });
      const { container } = render(CardInspectModal, { props: { card } });
      // Should not crash, just no pips
      expect(container.textContent).toContain('Lightning Bolt');
    });
  });

  describe('tags', () => {
    it('should render commander tag', () => {
      const card = makeCard({ tags: ['commander'] });
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('commander');
      expect(container.textContent).toContain('👑');
    });

    it('should render proxy tag and hide price', () => {
      const card = makeCard({ tags: ['proxy'] });
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('proxy');
      expect(container.textContent).toContain('🖨️');
      // Price should NOT be shown for proxies
      expect(container.textContent).not.toContain('$1.50');
    });

    it('should render wishlist tag', () => {
      const card = makeCard({ tags: ['wishlist'] });
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('wishlist');
      expect(container.textContent).toContain('🛒');
    });

    it('should render multiple tags', () => {
      const card = makeCard({ tags: ['commander', 'proxy'] });
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('commander');
      expect(container.textContent).toContain('proxy');
    });

    it('should show proxy notice for proxy cards', () => {
      const card = makeCard({ tags: ['proxy'] });
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('marked as a proxy');
    });
  });

  describe('double-faced cards', () => {
    it('should show flip button for DFC', () => {
      const card = makeCard({
        scryFall: {
          ...makeCard().scryFall!,
          isDoubleFaced: true,
          backImageUri: 'https://example.com/bolt-back.jpg',
        },
      });
      const { container } = render(CardInspectModal, { props: { card } });
      const flipBtn = Array.from(container.querySelectorAll('button')).find(
        (btn) => btn.textContent?.includes('Show Back')
      );
      expect(flipBtn).not.toBeUndefined();
    });

    it('should not show flip button for single-faced cards', () => {
      const card = makeCard();
      const { container } = render(CardInspectModal, { props: { card } });
      const flipBtn = Array.from(container.querySelectorAll('button')).find(
        (btn) => btn.textContent?.includes('Show Back')
      );
      expect(flipBtn).toBeUndefined();
    });

    it('should toggle to back image on flip', async () => {
      const card = makeCard({
        scryFall: {
          ...makeCard().scryFall!,
          isDoubleFaced: true,
          backImageUri: 'https://example.com/bolt-back.jpg',
        },
      });
      const { container } = render(CardInspectModal, { props: { card } });

      // Initially shows front image
      let img = container.querySelector('img');
      expect(img?.getAttribute('src')).toBe('https://example.com/bolt.jpg');

      // Click flip
      const flipBtn = Array.from(container.querySelectorAll('button')).find(
        (btn) => btn.textContent?.includes('Show Back')
      );
      if (flipBtn) {
        await fireEvent.click(flipBtn);
      }

      // Should now show back image
      img = container.querySelector('img');
      expect(img?.getAttribute('src')).toBe('https://example.com/bolt-back.jpg');
    });
  });

  describe('interactions', () => {
    it('should emit close on Escape key', async () => {
      const card = makeCard();
      const { component } = render(CardInspectModal, { props: { card } });

      let closed = false;
      component.$on('close', () => { closed = true; });

      await fireEvent.keyDown(window, { key: 'Escape' });
      expect(closed).toBe(true);
    });

    it('should emit close when clicking the backdrop', async () => {
      const card = makeCard();
      const { component, container } = render(CardInspectModal, { props: { card } });

      let closed = false;
      component.$on('close', () => { closed = true; });

      // Click the backdrop (outermost fixed div)
      const backdrop = container.querySelector('.fixed');
      if (backdrop) {
        await fireEvent.click(backdrop);
      }
      expect(closed).toBe(true);
    });

    it('should emit close when clicking the close button', async () => {
      const card = makeCard();
      const { component, container } = render(CardInspectModal, { props: { card } });

      let closed = false;
      component.$on('close', () => { closed = true; });

      // Find the × close button in the header
      const closeBtn = Array.from(container.querySelectorAll('button')).find(
        (btn) => btn.textContent?.trim() === '×'
      );
      if (closeBtn) {
        await fireEvent.click(closeBtn);
      }
      expect(closed).toBe(true);
    });

    it('should emit close when clicking the footer Close button', async () => {
      const card = makeCard();
      const { component, container } = render(CardInspectModal, { props: { card } });

      let closed = false;
      component.$on('close', () => { closed = true; });

      // Find the "Close" text button in the footer
      const closeBtn = Array.from(container.querySelectorAll('button')).find(
        (btn) => btn.textContent?.trim() === 'Close'
      );
      if (closeBtn) {
        await fireEvent.click(closeBtn);
      }
      expect(closed).toBe(true);
    });

    it('should NOT close when clicking inside the modal content', async () => {
      const card = makeCard();
      const { component, container } = render(CardInspectModal, { props: { card } });

      let closed = false;
      component.$on('close', () => { closed = true; });

      // Click on the inner modal dialog
      const dialog = container.querySelector('[role="dialog"], .bg-bg-secondary');
      if (dialog) {
        await fireEvent.click(dialog);
      }
      // Should NOT have closed (stopPropagation)
      expect(closed).toBe(false);
    });
  });

  describe('edge cases', () => {
    it('should handle card with no scryfall data', () => {
      const card: Card = { quantity: 1, name: 'Unknown Card' };
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('Unknown Card');
      expect(container.textContent).toContain('1×');
    });

    it('should handle card with empty tags array', () => {
      const card = makeCard({ tags: [] });
      const { container } = render(CardInspectModal, { props: { card } });
      // Should not show tags section label if no tags
      expect(container.textContent).toContain('Lightning Bolt');
    });

    it('should handle card with undefined tags', () => {
      const card = makeCard({ tags: undefined });
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('Lightning Bolt');
    });

    it('should format oracle text mana symbols as readable', () => {
      const card = makeCard({
        scryFall: {
          ...makeCard().scryFall!,
          oracleText: 'Pay {2}{W} to activate this ability.',
        },
      });
      const { container } = render(CardInspectModal, { props: { card } });
      expect(container.textContent).toContain('[2][W]');
    });
  });
});
