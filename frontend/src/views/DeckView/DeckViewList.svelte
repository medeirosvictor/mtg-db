<script lang="ts">
  import type { Card } from '../../lib/types';
  import { isBasicLand, cardToText } from '../../lib/cardUtils';
  import { UpdateCardText, ToggleCardTag } from '../../../wailsjs/go/app/App';
  import CardRow from './CardRow.svelte';
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

  // Selection state
  let selectedCards: Set<string> = new Set();
  let selectAll: boolean = false;

  // Inline editing state
  let editingCard: string | null = null;
  let editValue = '';
  let editError = '';

  // Context menu state
  let menuVisible = false;
  let menuX = 0;
  let menuY = 0;
  let menuItems: any[] = [];

  // Selection handlers
  export function toggleCardSelection(cardName: string) {
    if (selectedCards.has(cardName)) {
      selectedCards.delete(cardName);
    } else {
      selectedCards.add(cardName);
    }
    selectedCards = selectedCards;
    updateSelectAll();
  }

  export function toggleSelectAll() {
    if (selectAll) {
      selectedCards.clear();
    } else {
      cards.forEach(card => selectedCards.add(card.name));
    }
    selectedCards = selectedCards;
    selectAll = !selectAll;
  }

  function updateSelectAll() {
    if (cards.length === 0) {
      selectAll = false;
    } else {
      selectAll = cards.every(card => selectedCards.has(card.name));
    }
  }

  async function toggleTagOnSelected(tag: string) {
    if (selectedCards.size === 0) return;
    
    const cardNames = Array.from(selectedCards);
    for (const cardName of cardNames) {
      await ToggleCardTag(slug, cardName, tag);
    }
    
    // Clear selection after tagging
    selectedCards.clear();
    selectedCards = selectedCards;
    selectAll = false;
    
    dispatch('cardUpdated');
  }

  function showSelectionContextMenu(e: MouseEvent) {
    e.preventDefault();
    menuX = e.clientX;
    menuY = e.clientY;

    const hasProxy = Array.from(selectedCards).some(name => 
      cards.find(c => c.name === name)?.tags?.includes('proxy')
    );
    const hasWishlist = Array.from(selectedCards).some(name => 
      cards.find(c => c.name === name)?.tags?.includes('wishlist')
    );

    menuItems = [
      {
        label: `${selectedCards.size} card${selectedCards.size > 1 ? 's' : ''} selected`,
        icon: '📋',
        disabled: true,
        action: () => {},
      },
      { separator: true },
      {
        label: hasProxy ? 'Unmark as Proxy' : 'Mark as Proxy',
        icon: '🖨️',
        action: () => toggleTagOnSelected('proxy'),
      },
      {
        label: hasWishlist ? 'Unmark as Wishlisted' : 'Mark as Wishlisted',
        icon: '🛒',
        action: () => toggleTagOnSelected('wishlist'),
      },
      { separator: true },
      {
        label: 'Clear Selection',
        icon: '✕',
        action: () => {
          selectedCards.clear();
          selectedCards = selectedCards;
          selectAll = false;
        },
      },
    ];

    menuVisible = true;
  }

  function isNotFound(name: string): boolean {
    return notFoundCards.has(name.toLowerCase());
  }

  function startEditing(card: Card) {
    editingCard = card.name;
    editValue = cardToText(card);
    editError = '';
  }

  function cancelEditing() {
    editingCard = null;
    editValue = '';
    editError = '';
  }

  async function saveEditing(oldName: string) {
    if (!editValue.trim()) {
      editError = 'Card line cannot be empty';
      return;
    }

    try {
      const result = await UpdateCardText(slug, oldName, editValue);
      if (result === '') {
        editingCard = null;
        editValue = '';
        editError = '';
        dispatch('cardUpdated');
      } else {
        editError = result;
      }
    } catch (e) {
      editError = `Error: ${e}`;
    }
  }

  function handleEditKeydown(e: KeyboardEvent, oldName: string) {
    if (e.key === 'Enter') {
      e.preventDefault();
      saveEditing(oldName);
    } else if (e.key === 'Escape') {
      cancelEditing();
    }
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
      { separator: true },
      {
        label: 'Edit card text',
        icon: '✏️',
        action: () => startEditing(card),
      },
    ];

    menuVisible = true;
  }

  // Update selectAll when cards change
  $: if (cards) {
    updateSelectAll();
  }
</script>

<section>
  <h2 class="text-base font-semibold mb-3 text-text-secondary">Cards ({cards.length} unique, {deck?.cardCount} total)</h2>
  <div 
    class="bg-bg-secondary border border-border rounded-lg overflow-hidden"
    on:contextmenu={(e) => {
      if (selectedCards.size > 0) {
        showSelectionContextMenu(e);
      }
    }}
  >
    <div class="flex items-center px-4 py-2 bg-bg-surface text-xs font-semibold uppercase tracking-wide text-text-muted">
      <span class="w-8 flex-shrink-0 flex justify-center">
        <input 
          type="checkbox" 
          checked={selectAll} 
          on:change={toggleSelectAll}
          title={selectAll ? 'Deselect all' : 'Select all'}
          class="w-4 h-4 accent-accent"
        />
      </span>
      <span class="w-10 flex-shrink-0">#</span>
      <span class="flex-1 min-w-0">Card Name</span>
      <span class="w-48 flex-shrink-0">Tags</span>
      <span class="w-16 flex-shrink-0 text-right">Price</span>
      <span class="w-24 flex-shrink-0 text-right">Set</span>
    </div>
    {#each cards as card (card.name)}
      <CardRow
        {card}
        isNotFound={isNotFound(card.name)}
        isBasicLand={isBasicLand(card.name)}
        isCommander={(card.tags || []).includes('commander')}
        isFlipped={false}
        isEditing={editingCard === card.name}
        editValue={editingCard === card.name ? editValue : ''}
        editError={editingCard === card.name ? editError : ''}
        isSelected={selectedCards.has(card.name)}
        on:select={() => toggleCardSelection(card.name)}
        on:contextmenu={(e) => showCardContextMenu(e.detail, card)}
        on:dblclick={() => startEditing(card)}
        on:editSave={() => saveEditing(card.name)}
        on:editCancel={cancelEditing}
        on:editKeydown={(e) => handleEditKeydown(e.detail, card.name)}
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
