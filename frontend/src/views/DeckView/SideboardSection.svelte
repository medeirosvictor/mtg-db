<script lang="ts">
  import type { Card } from '../../lib/types';

  export let cards: Card[];
</script>

<section>
  <h2 class="text-base font-semibold mb-3 text-purple-400">Sideboard ({cards.length} cards)</h2>
  <div class="bg-bg-secondary border border-purple-500/30 rounded-lg overflow-hidden">
    <div class="flex items-center px-4 py-2 bg-bg-surface text-xs font-semibold uppercase tracking-wide text-text-muted">
      <span class="w-10 flex-shrink-0">#</span>
      <span class="flex-1 min-w-0">Card Name</span>
      <span class="w-48 flex-shrink-0">Tags</span>
      <span class="w-24 flex-shrink-0 text-right">Set</span>
    </div>
    {#each cards as card (card.name)}
      <div
        class="flex items-center px-4 py-1.5 border-b border-border last:border-b-0 text-sm hover:bg-bg-hover transition-colors text-text-secondary"
        role="button"
        tabindex="0"
      >
        <span class="w-10 flex-shrink-0 text-text-muted">{card.quantity}×</span>
        <span class="flex-1 min-w-0">{card.name}</span>
        <span class="w-48 flex-shrink-0 flex gap-1 flex-wrap">
          {#each card.tags || [] as tag}
            <span class="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wide
              {tag === 'commander' ? 'bg-yellow/15 text-yellow' : ''}
              {tag === 'proxy' ? 'bg-yellow/15 text-yellow' : ''}
              {tag === 'sideboard' ? 'bg-purple-500/15 text-purple-400' : ''}
              {tag === 'wishlist' ? 'bg-accent/15 text-accent' : ''}
            ">
              {#if tag === 'commander'}👑{/if}
              {#if tag === 'proxy'}🖨️{/if}
              {#if tag === 'sideboard'}📋{/if}
              {#if tag === 'wishlist'}🛒{/if}
              {tag}
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
