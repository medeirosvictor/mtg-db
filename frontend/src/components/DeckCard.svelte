<script lang="ts">
  import type { DeckSummary } from '../lib/types';
  import ColorPips from './ColorPips.svelte';
  import StatusBadge from './StatusBadge.svelte';

  export let deck: DeckSummary;
  export let onClick: () => void;

  $: cardCountClass = deck.cardCount === 100 ? 'text-green' : 'text-yellow';
</script>

<button 
  class="w-full text-left bg-bg-secondary border border-border rounded-xl p-5 hover:bg-bg-hover hover:border-accent transition-all cursor-pointer text-text-primary font-inherit"
  on:click={onClick}
>
  <div class="flex justify-between items-start gap-2 mb-3">
    <h3 class="text-base font-semibold leading-snug">{deck.title}</h3>
    <StatusBadge status={deck.status} />
  </div>

  <div class="mb-3">
    <span class="text-[11px] text-text-muted uppercase tracking-wide">Commander</span>
    <span class="block text-sm text-accent mt-0.5">{deck.commander}</span>
  </div>

  <div class="flex justify-between items-center">
    <div class="flex items-center gap-1.5">
      <ColorPips colors={deck.colors} />
    </div>
    <div class="flex items-center gap-1.5">
      <span class="text-lg font-bold {cardCountClass}">{deck.cardCount}</span>
      <span class="text-[11px] text-text-muted uppercase tracking-wide">cards</span>
    </div>
  </div>

  {#if deck.universe}
    <div class="mt-2 pt-2 border-t border-border">
      <span class="text-xs text-text-muted">{deck.universe}</span>
    </div>
  {/if}
</button>
