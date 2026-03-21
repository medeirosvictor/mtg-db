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

<section>
  <h2 class="text-base mb-3 text-orange">Wishlist ({cards.length})</h2>
  <div class="bg-bg-secondary border-2 border-border rounded overflow-hidden">
    <div class="flex items-center px-4 py-2 bg-bg-surface text-xs uppercase tracking-wide text-text-muted">
      <span class="w-10 flex-shrink-0">#</span>
      <span class="flex-1 min-w-0">Card Name</span>
      <span class="w-48 flex-shrink-0">Tags</span>
      <span class="w-24 flex-shrink-0 text-right">Set</span>
    </div>
    {#each cards as card (card.name)}
      <div
        class="flex items-center px-4 py-1.5 border-b border-border last:border-b-0 text-sm hover:bg-bg-hover text-text-secondary"
        on:contextmenu={(e) => dispatch('contextmenu', e)}
        role="button"
        tabindex="0"
      >
        <span class="w-10 flex-shrink-0 text-text-muted">{card.quantity}×</span>
        <span class="flex-1 min-w-0">{card.name}</span>
        <span class="w-48 flex-shrink-0 flex gap-1 flex-wrap">
          {#each getBadges(card) as badge}
            <span class="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] uppercase tracking-wide
              {badge === 'proxy' ? 'bg-yellow/15 text-yellow' : ''}
              {badge === 'wishlist' ? 'bg-accent/15 text-accent' : ''}
            ">
              {#if badge === 'proxy'}🖨️{/if}
              {#if badge === 'wishlist'}🛒{/if}
              {badge}
            </span>
          {/each}
        </span>
        <span class="w-24 flex-shrink-0 text-right text-text-muted text-xs">
          {card.setCode || ''}
        </span>
      </div>
    {/each}
  </div>
</section>
