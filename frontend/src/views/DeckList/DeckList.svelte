<script lang="ts">
  import { onMount } from 'svelte';
  import type { DeckSummary } from '../../lib/types';
  import { fuzzyMatch } from '../../lib/cardUtils';
  import { GetAllDecks, OpenDeckFolder } from '../../../wailsjs/go/app/App';
  import DeckCard from '../../components/DeckCard.svelte';
  import ContextMenu from '../../components/ContextMenu.svelte';
  import ImportNewDeckModal from '../../components/ImportNewDeckModal.svelte';
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  let decks: DeckSummary[] = [];
  let loading = true;
  let error = '';

  let showImportModal = false;

  let searchQuery = '';
  let searchInput: HTMLInputElement;

  // Context menu state
  let menuVisible = false;
  let menuX = 0;
  let menuY = 0;
  let menuItems: any[] = [];

  function showDeckContextMenu(e: MouseEvent, slug: string) {
    menuX = e.clientX;
    menuY = e.clientY;
    menuItems = [
      {
        label: 'Open in file explorer',
        icon: '📂',
        action: () => OpenDeckFolder(slug),
      },
    ];
    menuVisible = true;
  }

  onMount(async () => {
    try {
      const result = await GetAllDecks();
      decks = result || [];
    } catch (e) {
      error = `Failed to load decks: ${e}`;
    } finally {
      loading = false;
    }
  });

  function handleKeydown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
      e.preventDefault();
      searchInput?.focus();
      searchInput?.select();
    }
    if (e.key === 'Escape' && searchQuery) {
      searchQuery = '';
      searchInput?.blur();
    }
  }

  $: filteredDecks = searchQuery.trim()
    ? decks.filter(d => {
        const haystack = [d.title, d.commander, d.colors, d.status, d.universe || '', d.slug].join(' ');
        return fuzzyMatch(haystack, searchQuery);
      })
    : decks;

  $: totalCards = decks.reduce((sum, d) => sum + d.cardCount, 0);
  $: ownedDecks = decks.filter(d => d.status === 'Owned').length;
  $: plannedDecks = decks.filter(d => d.status === 'Planned').length;
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="flex-1 overflow-y-auto p-6 lg:p-8">
  {#if !loading}
    <div class="flex justify-between items-center mb-5">
      <div class="flex gap-5">
        {#if decks.length > 0}
          <div class="flex gap-5">
            <div class="flex flex-col items-center gap-0.5">
              <span class="text-lg font-mono text-accent">{decks.length}</span>
              <span class="text-[10px] text-text-muted uppercase tracking-wide">Decks</span>
            </div>
            <div class="flex flex-col items-center gap-0.5">
              <span class="text-lg font-mono text-accent">{totalCards}</span>
              <span class="text-[10px] text-text-muted uppercase tracking-wide">Cards</span>
            </div>
            <div class="flex flex-col items-center gap-0.5">
              <span class="text-lg font-mono text-green">{ownedDecks}</span>
              <span class="text-[10px] text-text-muted uppercase tracking-wide">Owned</span>
            </div>
            <div class="flex flex-col items-center gap-0.5">
              <span class="text-lg font-mono text-yellow">{plannedDecks}</span>
              <span class="text-[10px] text-text-muted uppercase tracking-wide">Planned</span>
            </div>
          </div>
        {/if}
      </div>
      <button
        class="whitespace-nowrap bg-bg-surface border-2 border-border text-text-secondary px-4 py-2 rounded text-sm hover:bg-bg-hover hover:border-accent hover:text-accent"
        on:click={() => showImportModal = true}
        title="Import a new deck from URL or card list"
      >
        📥 Import Deck
      </button>
    </div>
  {/if}

  {#if !loading && decks.length > 0}
    <div class="flex items-center gap-2 bg-bg-secondary border-2 border-border rounded px-3.5 py-2 mb-5 focus-within:border-accent">
      <span class="text-sm flex-shrink-0">🔍</span>
      <input
        type="text"
        class="flex-1 bg-transparent border-none text-text-primary font-inherit text-sm outline-none placeholder:text-text-muted"
        bind:this={searchInput}
        bind:value={searchQuery}
        placeholder="Search decks by name, commander, colors, status...  (Ctrl+F)"
      />
      {#if searchQuery}
        <button
          class="bg-transparent border-none text-text-muted cursor-pointer text-sm px-1.5 py-0.5 rounded hover:text-text-primary hover:bg-bg-hover flex-shrink-0"
          on:click={() => { searchQuery = ''; searchInput?.focus(); }}
          title="Clear search"
        >✕</button>
        <span class="text-xs text-text-muted flex-shrink-0 whitespace-nowrap">{filteredDecks.length} / {decks.length}</span>
      {/if}
    </div>
  {/if}

  {#if loading}
    <div class="text-center py-16 text-text-secondary">Loading decks...</div>
  {:else if error}
    <div class="text-center py-16 text-red">{error}</div>
  {:else if decks.length === 0}
    <div class="text-center py-16 text-text-secondary">
      <p>No decks found.</p>
      <p class="text-sm text-text-muted mt-2">Make sure the <code class="bg-bg-surface px-1.5 py-0.5 rounded text-xs">decks/</code> directory exists with deck folders.</p>
    </div>
  {:else if filteredDecks.length === 0}
    <div class="text-center py-16 text-text-secondary">
      <p>No decks match "<strong class="text-text-primary">{searchQuery}</strong>"</p>
      <p class="text-sm text-text-muted mt-2">Try a different search term or press <kbd class="bg-bg-surface border border-border rounded px-1.5 py-0.5 text-xs font-inherit">Esc</kbd> to clear.</p>
    </div>
  {:else}
    <div class="grid grid-cols-[repeat(auto-fill,minmax(320px,1fr))] gap-4">
      {#each filteredDecks as deck (deck.slug)}
        <DeckCard
          {deck}
          onClick={() => dispatch('navigate', { view: 'deck', slug: deck.slug })}
          on:contextmenu={(e) => showDeckContextMenu(e.detail, deck.slug)}
        />
      {/each}
    </div>
  {/if}
</div>

<ContextMenu
  bind:visible={menuVisible}
  x={menuX}
  y={menuY}
  items={menuItems}
/>

{#if showImportModal}
  <ImportNewDeckModal
    on:close={() => showImportModal = false}
    on:created={async (e) => {
      showImportModal = false;
      const result = await GetAllDecks();
      decks = result || [];
      dispatch('navigate', { view: 'deck', slug: e.detail.slug });
    }}
  />
{/if}
