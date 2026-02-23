<script lang="ts">
  import type { DeckSummary } from '../lib/types';
  import ColorPips from './ColorPips.svelte';
  import StatusBadge from './StatusBadge.svelte';

  export let deck: DeckSummary;
  export let onClick: () => void;

  $: cardCountClass = deck.cardCount === 100 ? 'count-ok' : 'count-warn';
</script>

<button class="deck-card" on:click={onClick}>
  <div class="deck-header">
    <h3 class="deck-title">{deck.title}</h3>
    <StatusBadge status={deck.status} />
  </div>

  <div class="deck-commander">
    <span class="label">Commander</span>
    <span class="value">{deck.commander}</span>
  </div>

  <div class="deck-meta">
    <div class="meta-item">
      <ColorPips colors={deck.colors} />
    </div>
    <div class="meta-item">
      <span class="count {cardCountClass}">{deck.cardCount}</span>
      <span class="label">cards</span>
    </div>
  </div>

  {#if deck.universe}
    <div class="deck-universe">
      <span class="label">{deck.universe}</span>
    </div>
  {/if}
</button>

<style>
  .deck-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 20px;
    cursor: pointer;
    transition: all 0.15s ease;
    text-align: left;
    width: 100%;
    color: var(--text-primary);
    font-family: inherit;
    font-size: inherit;
  }

  .deck-card:hover {
    background: var(--bg-hover);
    border-color: var(--accent);
    transform: translateY(-2px);
  }

  .deck-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 12px;
    gap: 8px;
  }

  .deck-title {
    font-size: 16px;
    font-weight: 600;
    line-height: 1.3;
  }

  .deck-commander {
    margin-bottom: 12px;
  }

  .deck-commander .value {
    display: block;
    font-size: 13px;
    color: var(--accent);
    margin-top: 2px;
  }

  .deck-meta {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .meta-item {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .label {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .count {
    font-size: 18px;
    font-weight: 700;
  }

  .count-ok {
    color: var(--green);
  }

  .count-warn {
    color: var(--yellow);
  }

  .deck-universe {
    margin-top: 8px;
    padding-top: 8px;
    border-top: 1px solid var(--border);
  }
</style>
