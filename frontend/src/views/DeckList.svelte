<script lang="ts">
  import { onMount } from 'svelte';
  import type { DeckSummary } from '../lib/types';
  import { GetAllDecks } from '../../wailsjs/go/app/App';
  import DeckCard from '../components/DeckCard.svelte';
  import ImportNewDeckModal from '../components/ImportNewDeckModal.svelte';
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  let decks: DeckSummary[] = [];
  let loading = true;
  let error = '';

  // Modal state
  let showImportModal = false;

  // Search state
  let searchQuery = '';
  let searchInput: HTMLInputElement;

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

  function fuzzyMatch(text: string, query: string): boolean {
    const lower = text.toLowerCase();
    const q = query.toLowerCase();
    // Simple substring match — check if all words in query appear in text
    const words = q.split(/\s+/).filter(w => w.length > 0);
    return words.every(word => lower.includes(word));
  }

  $: filteredDecks = searchQuery.trim()
    ? decks.filter(d => {
        const haystack = [
          d.title,
          d.commander,
          d.colors,
          d.status,
          d.universe || '',
          d.slug,
        ].join(' ');
        return fuzzyMatch(haystack, searchQuery);
      })
    : decks;

  // Stats (based on all decks, not filtered)
  $: totalCards = decks.reduce((sum, d) => sum + d.cardCount, 0);
  $: ownedDecks = decks.filter(d => d.status === 'Owned').length;
  $: plannedDecks = decks.filter(d => d.status === 'Planned').length;
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="deck-list">
  <header class="page-header">
    <div class="header-content">
      <h1>🃏 MTG Collection</h1>
      <p class="subtitle">Commander Decks & Collection Manager</p>
    </div>
    {#if !loading}
      <div class="header-actions">
        {#if decks.length > 0}
          <div class="stats">
            <div class="stat">
              <span class="stat-value">{decks.length}</span>
              <span class="stat-label">Decks</span>
            </div>
            <div class="stat">
              <span class="stat-value">{totalCards}</span>
              <span class="stat-label">Cards</span>
            </div>
            <div class="stat">
              <span class="stat-value">{ownedDecks}</span>
              <span class="stat-label">Owned</span>
            </div>
            <div class="stat">
              <span class="stat-value">{plannedDecks}</span>
              <span class="stat-label">Planned</span>
            </div>
          </div>
        {/if}
        <button class="import-btn" on:click={() => showImportModal = true} title="Import a new deck from URL or card list">
          📥 Import Deck
        </button>
      </div>
    {/if}
  </header>

  {#if !loading && decks.length > 0}
    <div class="search-bar">
      <span class="search-icon">🔍</span>
      <input
        type="text"
        class="search-input"
        bind:this={searchInput}
        bind:value={searchQuery}
        placeholder="Search decks by name, commander, colors, status...  (Ctrl+F)"
      />
      {#if searchQuery}
        <button class="search-clear" on:click={() => { searchQuery = ''; searchInput?.focus(); }} title="Clear search">✕</button>
        <span class="search-count">{filteredDecks.length} / {decks.length}</span>
      {/if}
    </div>
  {/if}

  {#if loading}
    <div class="loading">Loading decks...</div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if decks.length === 0}
    <div class="empty">
      <p>No decks found.</p>
      <p class="hint">Make sure the <code>decks/</code> directory exists with deck folders.</p>
    </div>
  {:else if filteredDecks.length === 0}
    <div class="empty">
      <p>No decks match "<strong>{searchQuery}</strong>"</p>
      <p class="hint">Try a different search term or press <kbd>Esc</kbd> to clear.</p>
    </div>
  {:else}
    <div class="grid">
      {#each filteredDecks as deck (deck.slug)}
        <DeckCard {deck} onClick={() => dispatch('navigate', { view: 'deck', slug: deck.slug })} />
      {/each}
    </div>
  {/if}
</div>

{#if showImportModal}
  <ImportNewDeckModal
    on:close={() => showImportModal = false}
    on:created={async (e) => {
      showImportModal = false;
      // Reload decks and navigate to the new deck
      const result = await GetAllDecks();
      decks = result || [];
      dispatch('navigate', { view: 'deck', slug: e.detail.slug });
    }}
  />
{/if}

<style>
  .deck-list {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    margin-bottom: 24px;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--border);
  }

  h1 {
    font-size: 28px;
    font-weight: 700;
  }

  .subtitle {
    color: var(--text-secondary);
    font-size: 14px;
    margin-top: 4px;
  }

  .header-actions {
    display: flex;
    align-items: flex-end;
    gap: 20px;
  }

  .import-btn {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    color: var(--text-secondary);
    padding: 8px 16px;
    border-radius: var(--radius);
    cursor: pointer;
    font-size: 13px;
    font-weight: 600;
    font-family: inherit;
    transition: all 0.15s ease;
    white-space: nowrap;
  }

  .import-btn:hover {
    background: var(--bg-hover);
    border-color: var(--accent);
    color: var(--accent);
  }

  .stats {
    display: flex;
    gap: 24px;
  }

  .stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
  }

  .stat-value {
    font-size: 22px;
    font-weight: 700;
    color: var(--accent);
  }

  .stat-label {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  /* Search bar */
  .search-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 8px 14px;
    margin-bottom: 20px;
    transition: border-color 0.15s ease;
  }

  .search-bar:focus-within {
    border-color: var(--accent);
    box-shadow: 0 0 0 2px rgba(137, 180, 250, 0.15);
  }

  .search-icon {
    font-size: 14px;
    flex-shrink: 0;
  }

  .search-input {
    flex: 1;
    background: none;
    border: none;
    color: var(--text-primary);
    font-family: inherit;
    font-size: 14px;
    outline: none;
  }

  .search-input::placeholder {
    color: var(--text-muted);
  }

  .search-clear {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 14px;
    padding: 2px 6px;
    border-radius: 4px;
    flex-shrink: 0;
  }

  .search-clear:hover {
    color: var(--text-primary);
    background: var(--bg-hover);
  }

  .search-count {
    font-size: 12px;
    color: var(--text-muted);
    flex-shrink: 0;
    white-space: nowrap;
  }

  kbd {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 1px 6px;
    font-size: 12px;
    font-family: inherit;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 16px;
  }

  .loading, .error, .empty {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-secondary);
  }

  .error {
    color: var(--red);
  }

  .empty strong {
    color: var(--text-primary);
  }

  .hint {
    margin-top: 8px;
    font-size: 13px;
    color: var(--text-muted);
  }

  code {
    background: var(--bg-surface);
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 12px;
  }
</style>
