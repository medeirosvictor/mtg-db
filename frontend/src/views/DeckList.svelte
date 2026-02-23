<script lang="ts">
  import { onMount } from 'svelte';
  import type { DeckSummary } from '../lib/types';
  import { GetAllDecks } from '../../wailsjs/go/main/App';
  import DeckCard from '../components/DeckCard.svelte';
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  let decks: DeckSummary[] = [];
  let loading = true;
  let error = '';

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

  // Stats
  $: totalCards = decks.reduce((sum, d) => sum + d.cardCount, 0);
  $: ownedDecks = decks.filter(d => d.status === 'Owned').length;
  $: plannedDecks = decks.filter(d => d.status === 'Planned').length;
</script>

<div class="deck-list">
  <header class="page-header">
    <div class="header-content">
      <h1>🃏 MTG Collection</h1>
      <p class="subtitle">Commander Decks & Collection Manager</p>
    </div>
    {#if !loading && decks.length > 0}
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
  </header>

  {#if loading}
    <div class="loading">Loading decks...</div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if decks.length === 0}
    <div class="empty">
      <p>No decks found.</p>
      <p class="hint">Make sure the <code>decks/</code> directory exists with deck folders.</p>
    </div>
  {:else}
    <div class="grid">
      {#each decks as deck (deck.slug)}
        <DeckCard {deck} onClick={() => dispatch('navigate', { view: 'deck', slug: deck.slug })} />
      {/each}
    </div>
  {/if}
</div>

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
    margin-bottom: 32px;
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
