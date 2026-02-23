<script lang="ts">
  import { onMount } from 'svelte';
  import type { Deck, Card } from '../lib/types';
  import { GetDeck } from '../../wailsjs/go/main/App';
  import ColorPips from '../components/ColorPips.svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import { createEventDispatcher } from 'svelte';

  export let slug: string;

  const dispatch = createEventDispatcher();

  let deck: Deck | null = null;
  let loading = true;
  let error = '';

  onMount(async () => {
    try {
      const result = await GetDeck(slug);
      deck = result;
      if (!deck) {
        error = `Deck "${slug}" not found`;
      }
    } catch (e) {
      error = `Failed to load deck: ${e}`;
    } finally {
      loading = false;
    }
  });

  // Group cards by type heuristic (basic categorization by name patterns)
  // This will be replaced by Scryfall type data in Phase 1
  function isBasicLand(name: string): boolean {
    const basics = ['plains', 'island', 'swamp', 'mountain', 'forest'];
    return basics.includes(name.toLowerCase());
  }

  // Sort: non-basics first, then alphabetical
  $: sortedCards = deck?.cards
    ? [...deck.cards].sort((a, b) => {
        const aBasic = isBasicLand(a.name);
        const bBasic = isBasicLand(b.name);
        if (aBasic !== bBasic) return aBasic ? 1 : -1;
        return a.name.localeCompare(b.name);
      })
    : [];

  $: wishlistCards = deck?.wishlist || [];
</script>

<div class="deck-view">
  {#if loading}
    <div class="loading">Loading deck...</div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if deck}
    <header class="deck-header">
      <button class="back-btn" on:click={() => dispatch('navigate', { view: 'home' })}>
        ← Back
      </button>
      <div class="header-info">
        <div class="title-row">
          <h1>{deck.info.title}</h1>
          <StatusBadge status={deck.info.status.includes('Owned') ? 'Owned' : deck.info.status.includes('Planned') ? 'Planned' : 'Disassembled'} />
        </div>
        <div class="meta-row">
          <ColorPips colors={deck.info.colors} />
          <span class="commander">Commander: <strong>{deck.info.commander}</strong></span>
          <span class="card-count" class:warn={deck.cardCount !== 100}>
            {deck.cardCount} cards
          </span>
        </div>
        {#if deck.info.strategy}
          <p class="strategy">{deck.info.strategy}</p>
        {/if}
      </div>
    </header>

    <div class="content">
      <section class="card-list">
        <h2>Cards ({sortedCards.length} unique, {deck.cardCount} total)</h2>
        <div class="cards-table">
          <div class="table-header">
            <span class="col-qty">#</span>
            <span class="col-name">Card Name</span>
            <span class="col-set">Set</span>
          </div>
          {#each sortedCards as card}
            <div class="card-row" class:basic-land={isBasicLand(card.name)}>
              <span class="col-qty">{card.quantity}×</span>
              <span class="col-name">
                {card.name}
                {#if card.foil}
                  <span class="foil-tag">✨</span>
                {/if}
              </span>
              <span class="col-set">
                {#if card.setCode}
                  {card.setCode}
                  {#if card.collectorNumber}
                    #{card.collectorNumber}
                  {/if}
                {/if}
              </span>
            </div>
          {/each}
        </div>
      </section>

      {#if wishlistCards.length > 0}
        <section class="card-list wishlist">
          <h2>Wishlist ({wishlistCards.length})</h2>
          <div class="cards-table">
            <div class="table-header">
              <span class="col-qty">#</span>
              <span class="col-name">Card Name</span>
              <span class="col-set">Set</span>
            </div>
            {#each wishlistCards as card}
              <div class="card-row wishlist-row">
                <span class="col-qty">{card.quantity}×</span>
                <span class="col-name">{card.name}</span>
                <span class="col-set">
                  {#if card.setCode}
                    {card.setCode}
                  {/if}
                </span>
              </div>
            {/each}
          </div>
        </section>
      {/if}
    </div>
  {/if}
</div>

<style>
  .deck-view {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
  }

  .deck-header {
    margin-bottom: 24px;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--border);
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

  .header-info {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .title-row {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  h1 {
    font-size: 24px;
    font-weight: 700;
  }

  .meta-row {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;
  }

  .commander {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .commander strong {
    color: var(--accent);
  }

  .card-count {
    font-size: 13px;
    color: var(--green);
    font-weight: 600;
  }

  .card-count.warn {
    color: var(--yellow);
  }

  .strategy {
    font-size: 13px;
    color: var(--text-secondary);
    line-height: 1.6;
    max-width: 800px;
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

  .card-row {
    display: flex;
    padding: 6px 16px;
    border-bottom: 1px solid var(--border);
    font-size: 13px;
    transition: background 0.1s;
  }

  .card-row:last-child {
    border-bottom: none;
  }

  .card-row:hover {
    background: var(--bg-hover);
  }

  .card-row.basic-land {
    color: var(--text-muted);
  }

  .card-row.wishlist-row {
    color: var(--text-secondary);
  }

  .col-qty {
    width: 40px;
    flex-shrink: 0;
    color: var(--text-muted);
  }

  .col-name {
    flex: 1;
  }

  .col-set {
    width: 120px;
    flex-shrink: 0;
    text-align: right;
    color: var(--text-muted);
    font-size: 12px;
  }

  .foil-tag {
    font-size: 11px;
  }

  .loading, .error {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-secondary);
  }

  .error {
    color: var(--red);
  }

  .wishlist h2 {
    color: var(--orange);
  }
</style>
