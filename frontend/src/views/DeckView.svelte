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
  import { GetDeck, GetDeckBasic, ToggleCardTag, UpdateCardText } from '../../wailsjs/go/app/App';
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
      // Use object spread to create a new object without the key
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
    // Check if at least one card has Scryfall data cached
    return cards.some(card => card.scryFall?.imageUri || card.scryFall?.oracleText);
  }

  async function loadDeck() {
    try {
      loading = true;
      error = '';
      
      // First, try to get the full deck with cached Scryfall data
      // This will return cached data if available, or fetch fresh if not
      try {
        const result = await GetDeck(slug);
        if (result) {
          deck = result.deck;
          if (result.notFound && result.notFound.length > 0) {
            notFoundCards = new Set(result.notFound.map(n => n.toLowerCase()));
          }
          // Successfully loaded with Scryfall data
          console.log('Deck loaded with Scryfall data');
          loading = false;
          return;
        }
      } catch (e) {
        // GetDeck failed, fall back to GetDeckBasic
        console.log('GetDeck failed, falling back to GetDeckBasic:', e);
      }
      
      // Fallback: Load basic deck without Scryfall data
      const result = await GetDeckBasic(slug);
      deck = result;
      if (!deck) {
        error = `Deck "${slug}" not found`;
      }
      
      // Check if we have cached data that could be loaded
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

  // Update selectAll when sorted cards change
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

<div class="deck-view">
  {#if loading}
    <div class="loading">
      <div class="loading-spinner"></div>
      <p>Loading deck...</p>
      <p class="loading-hint">Fetching card data from Scryfall</p>
    </div>
  {:else if error}
    <div class="error-container">
      <div class="error">{error}</div>
      <button class="retry-btn" on:click={loadDeck}>Try Again</button>
    </div>
  {:else if deck}
    <button class="back-btn" on:click={() => dispatch('navigate', { view: 'home' })}>
      ← Back
    </button>

    <DeckHeader {deck} {totalPrice} {displayCommander} />
    
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
      <div class="not-found-banner">
        ⚠️ {notFoundCards.size} card{notFoundCards.size > 1 ? 's' : ''} not found on Scryfall. Check card names highlighted in red below.
      </div>
    {/if}

    <div class="content">
      {#if viewMode === 'list'}
        <section class="card-list">
          <h2>Cards ({sortedCards.length} unique, {deck.cardCount} total)</h2>
          <div 
            class="cards-table"
            on:contextmenu={(e) => {
              if (selectedCards.size > 0) {
                showSelectionContextMenu(e);
              }
            }}
          >
            <div class="table-header">
              <span class="col-select">
                <input 
                  type="checkbox" 
                  checked={selectAll} 
                  on:change={toggleSelectAll}
                  title={selectAll ? 'Deselect all' : 'Select all'}
                />
              </span>
              <span class="col-qty">#</span>
              <span class="col-name">Card Name</span>
              <span class="col-tags">Tags</span>
              <span class="col-price">Price</span>
              <span class="col-set">Set</span>
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
        <section class="card-grid">
          <h2>Cards ({sortedCards.length} unique, {deck.cardCount} total)</h2>
          <div class="cards-grid">
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

<style>
  .deck-view {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
  }

  .back-btn {
    background: none;
    border: 1px solid var(--border);
    color: var(--text-secondary);
    padding: 6px 14px;
    border-radius: var(--radius);
    cursor: pointer;
    font-size: 13px;
    margin-bottom: 16px;
    font-family: inherit;
  }

  .back-btn:hover {
    background: var(--bg-surface);
    color: var(--text-primary);
  }

  h2 {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 12px;
    color: var(--text-secondary);
  }

  .content {
    display: flex;
    flex-direction: column;
    gap: 32px;
  }

  /* Table header (used in list view) */
  .cards-table {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
  }

  .table-header {
    display: flex;
    padding: 8px 16px;
    background: var(--bg-surface);
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-muted);
  }

  .col-qty {
    width: 40px;
    flex-shrink: 0;
  }

  .col-select {
    width: 32px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .col-select input[type="checkbox"] {
    width: 16px;
    height: 16px;
    cursor: pointer;
    accent-color: var(--accent);
  }

  .col-name {
    flex: 1;
    min-width: 0;
  }

  .col-tags {
    width: 200px;
    flex-shrink: 0;
  }

  .col-set {
    width: 120px;
    flex-shrink: 0;
    text-align: right;
  }

  .col-price {
    width: 70px;
    flex-shrink: 0;
    text-align: right;
  }

  /* Not-found banner */
  .not-found-banner {
    background: rgba(243, 139, 168, 0.1);
    border: 1px solid rgba(243, 139, 168, 0.3);
    color: var(--red);
    padding: 10px 16px;
    border-radius: var(--radius);
    font-size: 13px;
    margin-bottom: 16px;
  }

  /* Grid View */
  .card-grid h2 {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 16px;
    color: var(--text-secondary);
  }

  .cards-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 16px;
  }

  /* Loading/Error states */
  .loading, .error {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-secondary);
  }

  .loading-spinner {
    width: 40px;
    height: 40px;
    border: 3px solid var(--border);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 16px;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .loading-hint {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 8px;
  }

  .error-container {
    text-align: center;
    padding: 60px 20px;
  }

  .error {
    color: var(--red);
    background: rgba(243, 139, 168, 0.1);
    border: 1px solid rgba(243, 139, 168, 0.3);
    padding: 16px;
    border-radius: var(--radius);
    margin-bottom: 16px;
  }

  .retry-btn {
    background: var(--accent);
    color: #11111b;
    border: none;
    padding: 10px 20px;
    border-radius: var(--radius);
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    font-family: inherit;
  }

  .retry-btn:hover {
    background: var(--accent-hover);
  }
</style>
