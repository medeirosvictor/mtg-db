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
  export let isSelected: boolean = false;

  const dispatch = createEventDispatcher<{
    contextmenu: MouseEvent;
    dblclick: void;
    editSave: void;
    editCancel: void;
    editKeydown: KeyboardEvent;
    flip: void;
    select: void;
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
  <div class="py-1 px-4">
    <div class="flex flex-col gap-1">
      <div class="flex gap-1.5 items-center">
        <input
          type="text"
          class="flex-1 bg-bg-surface border border-accent rounded-lg text-text-primary font-mono text-sm px-2.5 py-1.5 outline-none focus:ring-2 focus:ring-accent/25"
          bind:value={editValue}
          on:keydown={(e) => dispatch('editKeydown', e)}
          autofocus
        />
        <div class="flex gap-1">
          <button 
            class="w-7 h-7 bg-bg-surface border border-border text-text-secondary rounded-lg cursor-pointer text-sm flex items-center justify-center hover:bg-green/15 hover:border-green hover:text-green"
            on:click={() => dispatch('editSave')} 
            title="Save (Enter)"
          >✓</button>
          <button 
            class="w-7 h-7 bg-bg-surface border border-border text-text-secondary rounded-lg cursor-pointer text-sm flex items-center justify-center hover:bg-red/15 hover:border-red hover:text-red"
            on:click={() => dispatch('editCancel')} 
            title="Cancel (Esc)"
          >✕</button>
        </div>
      </div>
      {#if editError}
        <div class="text-xs text-red">{editError}</div>
      {/if}
    </div>
  </div>
{:else}
  <div
    class="flex items-center px-4 py-1.5 border-b border-border last:border-b-0 transition-colors duration-100
      {isBasicLand ? 'text-text-muted' : ''}
      {isCommander ? 'bg-mauve/5 hover:bg-mauve/10' : ''}
      {isNotFound ? 'bg-red/5 hover:bg-red/10' : ''}
      {isSelected ? 'bg-accent/10 hover:bg-accent/15' : ''}
      {!isCommander && !isNotFound && !isSelected ? 'hover:bg-bg-hover' : ''}"
    on:contextmenu={(e) => dispatch('contextmenu', e)}
    on:dblclick={() => dispatch('dblclick')}
    role="button"
    tabindex="0"
  >
    <span class="w-8 flex-shrink-0 flex justify-center">
      <input 
        type="checkbox" 
        checked={isSelected} 
        on:change={() => dispatch('select')}
        on:click|stopPropagation
        class="w-4 h-4 accent-accent cursor-pointer"
      />
    </span>
    <span class="w-10 flex-shrink-0 text-text-muted">{card.quantity}×</span>
    <span class="flex-1 min-w-0 {isNotFound ? 'text-red font-semibold' : ''}">
      {#if isNotFound}
        <span class="text-xs mr-1" title="Card not found on Scryfall">⚠️</span>
      {/if}
      {card.name}
      {#if card.foil}
        <span class="text-xs ml-1">✨</span>
      {/if}
      {#if card.scryFall?.isDoubleFaced}
        <button
          class="ml-1 bg-transparent border-none cursor-pointer text-xs p-0.5 opacity-50 hover:opacity-100 transition-opacity"
          on:click|stopPropagation={() => dispatch('flip')}
          title={isFlipped ? 'Show front face' : 'Show back face'}
        >🔄</button>
      {/if}
    </span>
    <span class="w-48 flex-shrink-0 flex gap-1 flex-wrap">
      {#each getBadges(card) as badge}
        <span class="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wide
          {badge === 'commander' ? 'bg-mauve/15 text-mauve' : ''}
          {badge === 'proxy' ? 'bg-yellow/15 text-yellow' : ''}
          {badge === 'wishlist' ? 'bg-accent/15 text-accent' : ''}
        ">
          {#if badge === 'commander'}👑{/if}
          {#if badge === 'proxy'}🖨️{/if}
          {#if badge === 'wishlist'}🛒{/if}
          {badge}
        </span>
      {/each}
    </span>
    <span class="w-16 flex-shrink-0 text-right text-green text-xs font-semibold">
      {#if (card.tags || []).includes('proxy')}
        <span class="text-text-muted italic font-normal" title="Proxy — price not tracked">—</span>
      {:else}
        {card.scryFall?.priceUsd ? '$' + card.scryFall.priceUsd : '-'}
      {/if}
    </span>
    <span class="w-24 flex-shrink-0 text-right text-text-muted text-xs">
      {card.setCode || ''}
    </span>
  </div>
{/if}
