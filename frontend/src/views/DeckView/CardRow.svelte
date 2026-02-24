<script lang="ts">
  import type { Card } from '../../../lib/types';
  import { createEventDispatcher } from 'svelte';

  export let card: Card;
  export let isNotFound: boolean = false;
  export let isBasicLand: boolean = false;
  export let isCommander: boolean = false;
  export let isFlipped: boolean = false;
  export let isEditing: boolean = false;
  export let editValue: string = '';
  export let editError: string = '';

  const dispatch = createEventDispatcher<{
    contextmenu: MouseEvent;
    dblclick: void;
    editSave: void;
    editCancel: void;
    editKeydown: KeyboardEvent;
    flip: void;
  }>();

  function getBadges(card: Card): string[] {
    const badges: string[] = [];
    const tags = card.tags || [];
    if (tags.includes('commander')) badges.push('commander');
    if (tags.includes('proxy')) badges.push('proxy');
    if (tags.includes('wishlist')) badges.push('wishlist');
    return badges;
  }
</script>

{#if isEditing}
  <div class="card-row editing-row">
    <div class="edit-container">
      <div class="edit-row">
        <input
          type="text"
          class="edit-input"
          bind:value={editValue}
          on:keydown={(e) => dispatch('editKeydown', e)}
          autofocus
        />
        <div class="edit-actions">
          <button class="edit-btn save" on:click={() => dispatch('editSave')} title="Save (Enter)">✓</button>
          <button class="edit-btn cancel" on:click={() => dispatch('editCancel')} title="Cancel (Esc)">✕</button>
        </div>
      </div>
      {#if editError}
        <div class="edit-error">{editError}</div>
      {/if}
    </div>
  </div>
{:else}
  <div
    class="card-row"
    class:basic-land={isBasicLand}
    class:is-commander={isCommander}
    class:is-not-found={isNotFound}
    on:contextmenu={(e) => dispatch('contextmenu', e)}
    on:dblclick={() => dispatch('dblclick')}
    role="button"
    tabindex="0"
  >
    <span class="col-qty">{card.quantity}×</span>
    <span class="col-name">
      {#if isNotFound}
        <span class="not-found-icon" title="Card not found on Scryfall">⚠️</span>
      {/if}
      {card.name}
      {#if card.foil}
        <span class="foil-tag">✨</span>
      {/if}
      {#if card.scryFall?.isDoubleFaced}
        <button
          class="flip-btn"
          on:click|stopPropagation={() => dispatch('flip')}
          title={isFlipped ? 'Show front face' : 'Show back face'}
        >🔄</button>
      {/if}
    </span>
    <span class="col-tags">
      {#each getBadges(card) as badge}
        <span class="card-badge card-badge-{badge}">
          {#if badge === 'commander'}👑{/if}
          {#if badge === 'proxy'}🖨️{/if}
          {#if badge === 'wishlist'}🛒{/if}
          {badge}
        </span>
      {/each}
    </span>
    <span class="col-price">
      {#if (card.tags || []).includes('proxy')}
        <span class="proxy-price" title="Proxy — price not tracked">—</span>
      {:else}
        {card.scryFall?.priceUsd ? '$' + card.scryFall.priceUsd : '-'}
      {/if}
    </span>
    <span class="col-set">
      {card.setCode || ''}
    </span>
  </div>
{/if}

<style>
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

  .card-row.basic-land {
    color: var(--text-muted);
  }

  .card-row.is-commander {
    background: rgba(203, 166, 247, 0.06);
  }

  .card-row.is-commander:hover {
    background: rgba(203, 166, 247, 0.12);
  }

  .card-row.is-not-found {
    background: rgba(243, 139, 168, 0.06);
  }

  .card-row.is-not-found:hover {
    background: rgba(243, 139, 168, 0.12);
  }

  .card-row.is-not-found .col-name {
    color: var(--red);
    font-weight: 600;
  }

  .not-found-icon {
    font-size: 12px;
    margin-right: 4px;
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

  .card-badge-commander {
    background: rgba(203, 166, 247, 0.15);
    color: var(--mauve);
  }

  .card-badge-proxy {
    background: rgba(249, 226, 175, 0.15);
    color: var(--yellow);
  }

  .card-badge-wishlist {
    background: rgba(137, 180, 250, 0.15);
    color: var(--accent);
  }

  .foil-tag {
    font-size: 11px;
  }

  .flip-btn {
    background: none;
    border: none;
    cursor: pointer;
    font-size: 12px;
    padding: 0 2px;
    opacity: 0.5;
    transition: opacity 0.15s ease;
    vertical-align: middle;
  }

  .flip-btn:hover {
    opacity: 1;
  }

  .col-price {
    width: 70px;
    flex-shrink: 0;
    text-align: right;
    color: var(--green);
    font-size: 12px;
    font-weight: 600;
  }

  .proxy-price {
    color: var(--text-muted);
    font-style: italic;
    font-weight: 400;
  }

  /* Inline editing */
  .editing-row {
    padding: 4px 16px;
  }

  .edit-container {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .edit-row {
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .edit-input {
    flex: 1;
    background: var(--bg-surface);
    border: 1px solid var(--accent);
    border-radius: var(--radius);
    color: var(--text-primary);
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
    font-size: 13px;
    padding: 6px 10px;
    outline: none;
  }

  .edit-input:focus {
    box-shadow: 0 0 0 2px rgba(137, 180, 250, 0.25);
  }

  .edit-actions {
    display: flex;
    gap: 4px;
  }

  .edit-btn {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    color: var(--text-secondary);
    width: 28px;
    height: 28px;
    border-radius: var(--radius);
    cursor: pointer;
    font-size: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .edit-btn.save:hover {
    background: rgba(166, 227, 161, 0.15);
    border-color: var(--green);
    color: var(--green);
  }

  .edit-btn.cancel:hover {
    background: rgba(243, 139, 168, 0.15);
    border-color: var(--red);
    color: var(--red);
  }

  .edit-error {
    font-size: 11px;
    color: var(--red);
    padding: 2px 0;
  }
</style>
