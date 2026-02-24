<script lang="ts">
  import type { Card } from '../../../lib/types';
  import { createEventDispatcher } from 'svelte';

  export let card: Card;
  export let isNotFound: boolean = false;
  export let isBasicLand: boolean = false;
  export let isCommander: boolean = false;
  export let isFlipped: boolean = false;

  const dispatch = createEventDispatcher<{
    contextmenu: MouseEvent;
    flip: void;
  }>();

  function getBadges(card: Card): string[] {
    const badges: string[] = [];
    const tags = card.tags || [];
    if (tags.includes('proxy')) badges.push('proxy');
    if (tags.includes('wishlist')) badges.push('wishlist');
    return badges;
  }

  $: imageUri = isFlipped && card.scryFall?.backImageUri 
    ? card.scryFall.backImageUri 
    : card.scryFall?.imageUri;
</script>

<div
  class="grid-card"
  class:basic-land={isBasicLand}
  class:is-commander={isCommander}
  class:is-not-found={isNotFound}
  on:contextmenu={(e) => dispatch('contextmenu', e)}
  role="button"
  tabindex="0"
>
  <div class="card-image">
    {#if imageUri}
      <img src={imageUri} alt={card.name} loading="lazy" />
    {:else}
      <div class="card-placeholder" class:placeholder-not-found={isNotFound}>
        {#if isNotFound}
          <span class="not-found-icon-large">⚠️</span>
        {:else}
          {card.name.substring(0, 2).toUpperCase()}
        {/if}
      </div>
    {/if}
    {#if card.scryFall?.isDoubleFaced}
      <button
        class="flip-btn-grid"
        on:click|stopPropagation={() => dispatch('flip')}
        title={isFlipped ? 'Show front face' : 'Show back face'}
      >🔄</button>
    {/if}
  </div>
  <div class="card-details">
    <span class="card-name" class:card-name-not-found={isNotFound} title={card.name}>
      {#if isNotFound}<span class="not-found-icon-sm">⚠️</span>{/if}
      {card.name}
    </span>
    <span class="card-qty">
      {card.quantity}×
      {#if (card.tags || []).includes('proxy')}
        <span class="proxy-price">proxy</span>
      {:else if card.scryFall?.priceUsd}
        ${card.scryFall.priceUsd}
      {/if}
    </span>
  </div>
</div>

<style>
  .grid-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .grid-card:hover {
    border-color: var(--accent);
    transform: translateY(-2px);
  }

  .grid-card:hover .flip-btn-grid {
    opacity: 1;
  }

  .grid-card.basic-land {
    opacity: 0.7;
  }

  .grid-card.is-commander {
    border-color: var(--mauve);
  }

  .grid-card.is-not-found {
    border-color: var(--red);
  }

  .card-image {
    aspect-ratio: 488 / 680;
    background: var(--bg-surface);
    overflow: hidden;
    position: relative;
  }

  .card-image img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .flip-btn-grid {
    position: absolute;
    bottom: 6px;
    right: 6px;
    background: rgba(0, 0, 0, 0.6);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 50%;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    font-size: 14px;
    opacity: 0;
    transition: opacity 0.15s ease;
  }

  .flip-btn-grid:hover {
    background: rgba(0, 0, 0, 0.8);
    border-color: var(--accent);
  }

  .card-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
    font-weight: 700;
    color: var(--text-muted);
    background: linear-gradient(135deg, var(--bg-surface) 0%, var(--bg-secondary) 100%);
  }

  .card-placeholder.placeholder-not-found {
    background: linear-gradient(135deg, rgba(243, 139, 168, 0.1) 0%, rgba(243, 139, 168, 0.05) 100%);
  }

  .not-found-icon-large {
    font-size: 36px;
  }

  .card-details {
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .card-name {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .card-name-not-found {
    color: var(--red);
  }

  .not-found-icon-sm {
    font-size: 10px;
    margin-right: 2px;
  }

  .card-qty {
    font-size: 10px;
    color: var(--text-muted);
  }

  .proxy-price {
    color: var(--text-muted);
    font-style: italic;
  }
</style>
