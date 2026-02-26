<script lang="ts">
  import { onMount } from 'svelte';
  import type { Deck, Card } from '../lib/types';
  import { 
    sortCards, 
    filterCards, 
    calculateTotalPrice, 
    getCommanderNames 
  } from '../lib/cardUtils';
  import { GetDeck, GetDeckBasic, UpdateDeckInfo, UpdateDeckStatus } from '../../wailsjs/go/app/App';
  import DeckHeader from './DeckView/DeckHeader.svelte';
  import DeckToolbar from './DeckView/DeckToolbar.svelte';
  import DeckSearchBar from './DeckView/DeckSearchBar.svelte';
  import DeckViewList from './DeckView/DeckViewList.svelte';
  import DeckViewGrid from './DeckView/DeckViewGrid.svelte';
  import ImportDeckModal from '../components/ImportDeckModal.svelte';
  import ExportDeckModal from '../components/ExportDeckModal.svelte';
  import { createEventDispatcher } from 'svelte';

  export let slug: string;

  const dispatch = createEventDispatcher();

  let deck: Deck | null = null;
  let loading = true;
  let scryfallLoading = false;
  let error = '';
  let viewMode: 'list' | 'grid' = 'grid';

  // Not-found cards from Scryfall sync
  let notFoundCards: Set<string> = new Set();

  // Search state
  let searchQuery = '';
  let searchBarComponent: DeckSearchBar;

  // Modal state
  let showImportModal = false;
  let showExportModal = false;

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

  async function handleCardUpdated() {
    const reloaded = await GetDeckBasic(slug);
    if (reloaded) {
      deck = reloaded;
    }
  }

  // Derived state
  $: allSortedCards = deck?.cards ? sortCards(deck.cards) : [];
  $: sortedCards = filterCards(allSortedCards, searchQuery);

  $: allWishlistCards = deck?.wishlist || [];
  $: wishlistCards = filterCards(allWishlistCards, searchQuery);
  $: displayCommander = getCommanderNames(deck?.cards || []) || deck?.info?.commander || 'None set';

  $: totalPrice = calculateTotalPrice(deck?.cards || []);

  async function handleUpdateTitle(newTitle: string) {
    if (!deck) return;
    const result = await UpdateDeckInfo(slug, newTitle, deck.info.strategy || '');
    if (result === '') {
      deck.info.title = newTitle;
      deck = deck;
    } else {
      console.error('Failed to update title:', result);
    }
  }

  async function handleUpdateStrategy(newStrategy: string) {
    if (!deck) return;
    const result = await UpdateDeckInfo(slug, deck.info.title, newStrategy);
    if (result === '') {
      deck.info.strategy = newStrategy;
      deck = deck;
    } else {
      console.error('Failed to update strategy:', result);
    }
  }

  async function handleUpdateStatus(newStatus: string) {
    if (!deck) return;
    const result = await UpdateDeckStatus(slug, newStatus);
    if (result === '') {
      if (newStatus === 'Owned') {
        deck.info.status = '✅ Owned';
      } else if (newStatus === 'Planned') {
        deck.info.status = '📋 Planned';
      } else if (newStatus === 'Disassembled') {
        deck.info.status = '🔧 Disassembled';
      }
      deck = deck;
    } else {
      console.error('Failed to update status:', result);
    }
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
        <DeckViewList
          {deck}
          cards={sortedCards}
          {wishlistCards}
          {notFoundCards}
          {slug}
          on:cardUpdated={handleCardUpdated}
        />
      {:else}
        <DeckViewGrid
          {deck}
          cards={sortedCards}
          {wishlistCards}
          {notFoundCards}
          {slug}
          on:cardUpdated={handleCardUpdated}
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