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
    inspect: void;
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
  class="bg-bg-secondary border rounded-lg overflow-hidden cursor-pointer transition-all hover:border-accent hover:-translate-y-0.5
    {isBasicLand ? 'opacity-70' : ''}
    {isCommander ? 'border-mauve' : ''}
    {isNotFound ? 'border-red' : 'border-border'}"
  on:contextmenu={(e) => dispatch('contextmenu', e)}
  on:click={() => dispatch('inspect')}
  role="button"
  tabindex="0"
>
  <div class="aspect-[488/680] bg-bg-surface overflow-hidden relative">
    {#if imageUri}
      <img src={imageUri} alt={card.name} loading="lazy" class="w-full h-full object-cover" />
    {:else}
      <div class="w-full h-full flex items-center justify-center text-2xl font-bold text-text-muted {isNotFound ? 'bg-red/10' : 'bg-gradient-to-br from-bg-surface to-bg-secondary'}">
        {#if isNotFound}
          <span class="text-4xl">⚠️</span>
        {:else}
          {card.name.substring(0, 2).toUpperCase()}
        {/if}
      </div>
    {/if}
    {#if card.scryFall?.isDoubleFaced}
      <button
        class="absolute bottom-1.5 right-1.5 bg-black/60 border border-white/20 rounded-full w-7 h-7 flex items-center justify-center cursor-pointer text-sm opacity-0 hover:bg-black/80 hover:border-accent transition-all"
        class:opacity-100={true}
        on:click|stopPropagation={() => dispatch('flip')}
        title={isFlipped ? 'Show front face' : 'Show back face'}
      >🔄</button>
    {/if}
  </div>
  <div class="p-2 flex flex-col gap-0.5">
    <span 
      class="text-xs font-semibold text-text-primary whitespace-nowrap overflow-hidden text-ellipsis {isNotFound ? 'text-red' : ''}" 
      title={card.name}
    >
      {#if isNotFound}<span class="text-[10px] mr-0.5">⚠️</span>{/if}
      {card.name}
    </span>
    <span class="text-xs text-text-muted">
      {card.quantity}×
      {#if (card.tags || []).includes('proxy')}
        <span class="italic">proxy</span>
      {:else if card.scryFall?.priceUsd}
        ${card.scryFall.priceUsd}
      {/if}
    </span>
  </div>
</div>
