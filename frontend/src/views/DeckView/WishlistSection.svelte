<script lang="ts">
  import type { Card } from '../../../lib/types';
  import { createEventDispatcher } from 'svelte';

  export let cards: Card[];

  const dispatch = createEventDispatcher<{
    contextmenu: MouseEvent;
  }>();

  function getBadges(card: Card): string[] {
    const badges: string[] = [];
    const tags = card.tags || [];
    if (tags.includes('proxy')) badges.push('proxy');
    if (tags.includes('wishlist')) badges.push('wishlist');
    return badges;
  }
</script>

<section class="card-list wishlist">
  <h2>Wishlist ({cards.length})</h2>
  <div class="cards-table">
    <div class="table-header">
      <span class="col-qty">#</span>
      <span class="col-name">Card Name</span>
      <span class="col-tags">Tags</span>
      <span class="col-set">Set</span>
    </div>
    {#each cards as card (card.name)}
      <div
        class="card-row wishlist-row"
        on:contextmenu={(e) => dispatch('contextmenu', e)}
        role="button"
        tabindex="0"
      >
        <span class="col-qty">{card.quantity}×</span>
        <span class="col-name">{card.name}</span>
        <span class="col-tags">
          {#each getBadges(card) as badge}
            <span class="card-badge card-badge-{badge}">
              {#if badge === 'proxy'}🖨️{/if}
              {#if badge === 'wishlist'}🛒{/if}
              {badge}
            </span>
          {/each}
        </span>
        <span class="col-set">
          {card.setCode || ''}
        </span>
      </div>
    {/each}
  </div>
</section>

<style>
  .wishlist h2 {
    color: var(--orange);
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
    align-items: center;
  }

  .card-row:last-child {
    border-bottom: none;
  }

  .card-row:hover {
    background: var(--bg-hover);
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
    min-width: 0;
  }

  .col-tags {
    width: 200px;
    flex-shrink: 0;
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
  }

  .col-set {
    width: 120px;
    flex-shrink: 0;
    text-align: right;
    color: var(--text-muted);
    font-size: 12px;
  }

  .card-badge {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 1px 6px;
    border-radius: 3px;
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  .card-badge-proxy {
    background: rgba(249, 226, 175, 0.15);
    color: var(--yellow);
  }

  .card-badge-wishlist {
    background: rgba(137, 180, 250, 0.15);
    color: var(--accent);
  }
</style>
