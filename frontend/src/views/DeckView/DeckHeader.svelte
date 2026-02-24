<script lang="ts">
  import ColorPips from '../../components/ColorPips.svelte';
  import StatusBadge from '../../components/StatusBadge.svelte';
  import type { Deck } from '../lib/types';

  export let deck: Deck;
  export let totalPrice: number;
  export let displayCommander: string;
</script>

<header class="deck-header">
  <div class="header-info">
    <div class="title-row">
      <h1>{deck.info.title}</h1>
      <StatusBadge status={deck.info.status.includes('Owned') ? 'Owned' : deck.info.status.includes('Planned') ? 'Planned' : 'Disassembled'} />
    </div>
    <div class="meta-row">
      <ColorPips colors={deck.info.colors} />
      <span class="commander">Commander: <strong>{displayCommander}</strong></span>
      <span class="card-count" class:warn={deck.cardCount !== 100}>
        {deck.cardCount} cards
      </span>
      <span class="total-price">
        ${totalPrice.toFixed(2)}
      </span>
    </div>
    {#if deck.info.strategy}
      <p class="strategy">{deck.info.strategy}</p>
    {/if}
  </div>
</header>

<style>
  .deck-header {
    margin-bottom: 24px;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--border);
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

  .total-price {
    font-size: 13px;
    color: var(--green);
    font-weight: 600;
  }

  .strategy {
    font-size: 13px;
    color: var(--text-secondary);
    line-height: 1.6;
    max-width: 800px;
  }
</style>
