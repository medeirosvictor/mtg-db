<script lang="ts">
  import type { Card } from '../../lib/types';
  import { ToggleCardTag } from '../../../wailsjs/go/app/App';
  import CardGridItem from './CardGridItem.svelte';
  import WishlistSection from './WishlistSection.svelte';
  import ContextMenu from '../../components/ContextMenu.svelte';
  import { createEventDispatcher } from 'svelte';

  export let deck: any;
  export let cards: Card[];
  export let wishlistCards: Card[];
  export let notFoundCards: Set<string>;
  export let slug: string;

  const dispatch = createEventDispatcher<{
    cardUpdated: void;
  }>();

  // DFC flip state
  let flippedCards: Record<string, boolean> = {};

  // Context menu state
  let menuVisible = false;
  let menuX = 0;
  let menuY = 0;
  let menuItems: any[] = [];

  function toggleFlip(cardName: string) {
    if (flippedCards[cardName]) {
      const { [cardName]: _, ...rest } = flippedCards;
      flippedCards = rest;
    } else {
      flippedCards = { ...flippedCards, [cardName]: true };
    }
  }

  function isNotFound(name: string): boolean {
    return notFoundCards.has(name.toLowerCase());
  }

  async function toggleTag(cardName: string, tag: string) {
    await ToggleCardTag(slug, cardName, tag);
    dispatch('cardUpdated');
  }

  function showCardContextMenu(e: MouseEvent, card: Card) {
    e.preventDefault();
    menuX = e.clientX;
    menuY = e.clientY;

    const tags = card.tags || [];
    const isCommander = tags.includes('commander');
    const isProxy = tags.includes('proxy');
    const isWishlisted = tags.includes('wishlist');
    const commanderCount = deck?.cards?.filter((c: Card) => (c.tags || []).includes('commander')).length || 0;

    menuItems = [
      {
        label: isCommander ? 'Remove as Commander' : 'Set as Commander',
        icon: '👑',
        checked: isCommander,
        disabled: !isCommander && commanderCount >= 2,
        action: () => toggleTag(card.name, 'commander'),
      },
      { separator: true },
      {
        label: isProxy ? 'Unmark as Proxy' : 'Mark as Proxy',
        icon: '🖨️',
        checked: isProxy,
        action: () => toggleTag(card.name, 'proxy'),
      },
      {
        label: isWishlisted ? 'Unmark as Wishlisted' : 'Mark as Wishlisted',
        icon: '🛒',
        checked: isWishlisted,
        action: () => toggleTag(card.name, 'wishlist'),
      },
    ];

    menuVisible = true;
  }
</script>

<section>
  <h2 class="text-base font-semibold mb-4 text-text-secondary">Cards ({cards.length} unique, {deck?.cardCount} total)</h2>
  <div class="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-4">
    {#each cards as card (card.name)}
      <CardGridItem
        {card}
        isNotFound={isNotFound(card.name)}
        isBasicLand={false}
        isCommander={(card.tags || []).includes('commander')}
        isFlipped={!!flippedCards[card.name]}
        on:contextmenu={(e) => showCardContextMenu(e.detail, card)}
        on:flip={() => toggleFlip(card.name)}
      />
    {/each}
  </div>
</section>

{#if wishlistCards.length > 0}
  <WishlistSection 
    cards={wishlistCards}
  />
{/if}

<ContextMenu
  bind:visible={menuVisible}
  x={menuX}
  y={menuY}
  items={menuItems}
/>
