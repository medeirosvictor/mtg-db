<script lang="ts">
  import { onMount } from 'svelte';
  import type { Deck, Card } from '../lib/types';
  import { 
    fuzzyMatch, 
    isBasicLand, 
    cardToText, 
    sortCards, 
    filterCards, 
    calculateTotalPrice, 
    getCommanderNames 
  } from '../lib/cardUtils';
  import { GetDeck, GetDeckBasic, ToggleCardTag, UpdateCardText, UpdateDeckInfo, UpdateDeckStatus } from '../../wailsjs/go/app/App';
  import DeckHeader from './DeckView/DeckHeader.svelte';
  import DeckToolbar from './DeckView/DeckToolbar.svelte';
  import DeckSearchBar from './DeckView/DeckSearchBar.svelte';
  import CardRow from './DeckView/CardRow.svelte';
  import CardGridItem from './DeckView/CardGridItem.svelte';
  import WishlistSection from './DeckView/WishlistSection.svelte';
  import ContextMenu from '../components/ContextMenu.svelte';
  import ImportDeckModal from '../components/ImportDeckModal.svelte';
  import ExportDeckModal from '../components/ExportDeckModal.svelte';
  import { createEventDispatcher } from 'svelte';

  export let slug: string;

  const dispatch = createEventDispatcher();

  let deck: Deck | null = null;
  let loading = true;
  let scryfallLoading = false;
  let error = '';
  let viewMode: 'list' | 'grid' = 'list';

  // Not-found cards from Scryfall sync
  let notFoundCards: Set<string> = new Set();

  // Inline editing state
  let editingCard: string | null = null;
  let editValue = '';
  let editError = '';

  // DFC flip state
  let flippedCards: Record<string, boolean> = {};

  // Selection state
  let selectedCards: Set<string> = new Set();
  let selectAll: boolean = false;

  // Search state
  let searchQuery = '';
  let searchBarComponent: DeckSearchBar;

  // Selection handlers
  function toggleCardSelection(cardName: string) {
    if (selectedCards.has(cardName)) {
      selectedCards.delete(cardName);
    } else {
      selectedCards.add(cardName);
    }
    selectedCards = selectedCards; // trigger reactivity
    updateSelectAll();
  }

  function toggleSelectAll() {
    if (selectAll) {
      selectedCards.clear();
    } else {
      sortedCards.forEach(card => selectedCards.add(card.name));
    }
    selectedCards = selectedCards;
    selectAll = !selectAll;
  }

  function updateSelectAll() {
    if (sortedCards.length === 0) {
      selectAll = false;
    } else {
      selectAll = sortedCards.every(card => selectedCards.has(card.name));
    }
  }

  async function toggleTagOnSelected(tag: string) {
    if (!deck || selectedCards.size === 0) return;
    
    const cardNames = Array.from(selectedCards);
    for (const cardName of cardNames) {
      await ToggleCardTag(slug, cardName, tag);
    }
    
    // Reload deck to get updated data
    const reloaded = await GetDeckBasic(slug);
    if (reloaded) {
      deck = reloaded;
    }
    
    // Clear selection after tagging
    selectedCards.clear();
    selectedCards = selectedCards;
    selectAll = false;
  }

  function showSelectionContextMenu(e: MouseEvent) {
    e.preventDefault();
    menuX = e.clientX;
    menuY = e.clientY;

    const hasProxy = Array.from(selectedCards).some(name => 
      sortedCards.find(c => c.name === name)?.tags?.includes('proxy')
    );
    const hasWishlist = Array.from(selectedCards).some(name => 
      sortedCards.find(c => c.name === name)?.tags?.includes('wishlist')
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

  // Modal state
  let showImportModal = false;
  let showExportModal = false;

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

  function handleGlobalKeydown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
      e.preventDefault();
      searchBarComponent?.focus();
    }
    if (e.key === 'Escape' && searchQuery) {
      searchQuery = '';
    }
  }

  onMount(async () => {
    await loadDeck();
  });

  // Check if any card has cached Scryfall data
  function hasCachedData(cards: Card[]): boolean {
    if (!cards || cards.length === 0) return false;
    return cards.some(card => card.scryFall?.imageUri || card.scryFall?.oracleText);
  }

  async function loadDeck() {
    try {
      loading = true;
      error = '';
      
      try {
        const result = await GetDeck(slug);
        if (result) {
          deck = result.deck;
          if (result.notFound && result.notFound.length > 0) {
            notFoundCards = new Set(result.notFound.map(n => n.toLowerCase()));
          }
          console.log('Deck loaded with Scryfall data');
          loading = false;
          return;
        }
      } catch (e) {
        console.log('GetDeck failed, falling back to GetDeckBasic:', e);
      }
      
      const result = await GetDeckBasic(slug);
      deck = result;
      if (!deck) {
        error = `Deck "${slug}" not found`;
      }
      
      if (deck && hasCachedData(deck.cards)) {
        console.log('Cached Scryfall data available - user can click Sync to load');
      }
    } catch (e) {
      console.error('Failed to load deck:', e);
      error = `Failed to load deck: ${e}`;
    } finally {
      loading = false;
    }
  }

  async function syncScryfall() {
    try {
      scryfallLoading = true;
      error = '';
      notFoundCards = new Set();
      console.log('Syncing with Scryfall...');
      const result = await GetDeck(slug);
      if (result) {
        deck = result.deck;
        if (result.notFound && result.notFound.length > 0) {
          notFoundCards = new Set(result.notFound.map(n => n.toLowerCase()));
          console.log('Cards not found on Scryfall:', result.notFound);
        }
      }
      console.log('Scryfall sync complete');
    } catch (e) {
      console.error('Failed to sync Scryfall:', e);
      error = `Scryfall sync failed: ${e}`;
    } finally {
      scryfallLoading = false;
    }
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
        const reloaded = await GetDeckBasic(slug);
        if (reloaded) {
          deck = reloaded;
        }
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

  // Derived state
  $: allSortedCards = deck?.cards ? sortCards(deck.cards) : [];
  $: sortedCards = filterCards(allSortedCards, searchQuery);

  $: allWishlistCards = deck?.wishlist || [];
  $: wishlistCards = filterCards(allWishlistCards, searchQuery);
  $: displayCommander = getCommanderNames(deck?.cards || []) || deck?.info?.commander || 'None set';

  $: totalPrice = calculateTotalPrice(deck?.cards || []);

  $: if (sortedCards) {
    updateSelectAll();
  }

  async function toggleTag(cardName: string, tag: string) {
    if (!deck) return;
    const updated = await ToggleCardTag(slug, cardName, tag);
    if (updated) {
      deck = updated;
    }
  }

  async function handleUpdateTitle(newTitle: string) {
    if (!deck) return;
    const result = await UpdateDeckInfo(slug, newTitle, deck.info.strategy || '');
    if (result === '') {
      deck.info.title = newTitle;
      deck = deck; // trigger reactivity
    } else {
      console.error('Failed to update title:', result);
    }
  }

  async function handleUpdateStrategy(newStrategy: string) {
    if (!deck) return;
    const result = await UpdateDeckInfo(slug, deck.info.title, newStrategy);
    if (result === '') {
      deck.info.strategy = newStrategy;
      deck = deck; // trigger reactivity
    } else {
      console.error('Failed to update strategy:', result);
    }
  }

  async function handleUpdateStatus(newStatus: string) {
    if (!deck) return;
    const result = await UpdateDeckStatus(slug, newStatus);
    if (result === '') {
      // Update the status in deck.info.status
      if (newStatus === 'Owned') {
        deck.info.status = '✅ Owned';
      } else if (newStatus === 'Planned') {
        deck.info.status = '📋 Planned';
      } else if (newStatus === 'Disassembled') {
        deck.info.status = '🔧 Disassembled';
      }
      deck = deck; // trigger reactivity
    } else {
      console.error('Failed to update status:', result);
    }
  }

  function showCardContextMenu(e: MouseEvent, card: Card) {
    e.preventDefault();
    menuX = e.clientX;
    menuY = e.clientY;

    const tags = card.tags || [];
    const isCommander = tags.includes('commander');
    const isProxy = tags.includes('proxy');
    const isWishlisted = tags.includes('wishlist');
    const commanderCount = deck?.cards?.filter(c => (c.tags || []).includes('commander')).length || 0;

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
</script>

<svelte:window on:keydown={handleGlobalKeydown} />

<div class="flex-1 overflow-y-auto p-6 lg:p-8">
  {#if loading}
    <div class="text-center py-16 text-text-secondary">
      <div class="w-10 h-10 border-3 border-border border-t-accent rounded-full animate-spin mx-auto mb-4"></div>
      <p>Loading deck...</p>
      <p class="text-xs text-text-muted mt-2">Fetching card data from Scryfall</p>
    </div>
  {:else if error}
    <div class="text-center py-16">
      <div class="bg-red/10 border border-red/30 text-red px-4 py-4 rounded-lg mb-4">{error}</div>
      <button 
        class="bg-accent text-bg-primary border-none px-5 py-2.5 rounded-lg text-sm font-semibold cursor-pointer hover:bg-accent-hover"
        on:click={loadDeck}
      >Try Again</button>
    </div>
  {:else if deck}
    <button 
      class="bg-transparent border border-border text-text-secondary px-3.5 py-1.5 rounded-lg text-sm mb-4 font-inherit cursor-pointer hover:bg-bg-surface hover:text-text-primary"
      on:click={() => dispatch('navigate', { view: 'home' })}
    >
      ← Back
    </button>

    <DeckHeader 
      {deck} 
      {totalPrice} 
      {displayCommander} 
      on:updateTitle={(e) => handleUpdateTitle(e.detail)}
      on:updateStrategy={(e) => handleUpdateStrategy(e.detail)}
      on:updateStatus={(e) => handleUpdateStatus(e.detail)}
    />
    
    <DeckToolbar 
      {loading} 
      {scryfallLoading} 
      {viewMode}
      on:import={() => showImportModal = true}
      on:export={() => showExportModal = true}
      on:sync={syncScryfall}
      on:viewChange={(e) => viewMode = e.detail}
    />

    <DeckSearchBar 
      bind:searchQuery 
      count={sortedCards.length}
      total={allSortedCards.length}
      bind:this={searchBarComponent}
    />

    {#if notFoundCards.size > 0}
      <div class="bg-red/10 border border-red/30 text-red px-4 py-2.5 rounded-lg text-sm mb-4">
        ⚠️ {notFoundCards.size} card{notFoundCards.size > 1 ? 's' : ''} not found on Scryfall. Check card names highlighted in red below.
      </div>
    {/if}

    <div class="flex flex-col gap-8">
      {#if viewMode === 'list'}
        <section>
          <h2 class="text-base font-semibold mb-3 text-text-secondary">Cards ({sortedCards.length} unique, {deck.cardCount} total)</h2>
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
            {#each sortedCards as card (card.name)}
              <CardRow
                {card}
                isNotFound={isNotFound(card.name)}
                isBasicLand={isBasicLand(card.name)}
                isCommander={(card.tags || []).includes('commander')}
                isFlipped={!!flippedCards[card.name]}
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
                on:flip={() => toggleFlip(card.name)}
              />
            {/each}
          </div>
        </section>
      {:else}
        <section>
          <h2 class="text-base font-semibold mb-4 text-text-secondary">Cards ({sortedCards.length} unique, {deck.cardCount} total)</h2>
          <div class="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-4">
            {#each sortedCards as card (card.name)}
              <CardGridItem
                {card}
                isNotFound={isNotFound(card.name)}
                isBasicLand={isBasicLand(card.name)}
                isCommander={(card.tags || []).includes('commander')}
                isFlipped={!!flippedCards[card.name]}
                on:contextmenu={(e) => showCardContextMenu(e.detail, card)}
                on:flip={() => toggleFlip(card.name)}
              />
            {/each}
          </div>
        </section>
      {/if}

      {#if wishlistCards.length > 0}
        <WishlistSection 
          cards={wishlistCards}
        />
      {/if}
    </div>
  {/if}
</div>

{#if showImportModal}
  <ImportDeckModal
    {slug}
    on:close={() => showImportModal = false}
    on:imported={async () => {
      showImportModal = false;
      const reloaded = await GetDeckBasic(slug);
      if (reloaded) deck = reloaded;
    }}
  />
{/if}

{#if showExportModal}
  <ExportDeckModal
    {slug}
    deckTitle={deck?.info?.title || slug}
    on:close={() => showExportModal = false}
  />
{/if}

<ContextMenu
  bind:visible={menuVisible}
  x={menuX}
  y={menuY}
  items={menuItems}
/>
