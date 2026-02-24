<script lang="ts">
  import ColorPips from '../../components/ColorPips.svelte';
  import StatusBadge from '../../components/StatusBadge.svelte';
  import type { Deck } from '../lib/types';

  export let deck: Deck;
  export let totalPrice: number;
  export let displayCommander: string;
</script>

<header class="mb-6 pb-4 border-b border-border">
  <div class="flex flex-col gap-2">
    <div class="flex items-center gap-3">
      <h1 class="text-2xl font-bold">{deck.info.title}</h1>
      <StatusBadge status={deck.info.status.includes('Owned') ? 'Owned' : deck.info.status.includes('Planned') ? 'Planned' : 'Disassembled'} />
    </div>
    <div class="flex items-center gap-4 flex-wrap">
      <ColorPips colors={deck.info.colors} />
      <span class="text-sm text-text-secondary">Commander: <strong class="text-accent">{displayCommander}</strong></span>
      <span class="text-sm {deck.cardCount !== 100 ? 'text-yellow' : 'text-green'} font-semibold">
        {deck.cardCount} cards
      </span>
      <span class="text-sm text-green font-semibold">
        ${totalPrice.toFixed(2)}
      </span>
    </div>
    {#if deck.info.strategy}
      <p class="text-sm text-text-secondary leading-relaxed max-w-[800px]">{deck.info.strategy}</p>
    {/if}
  </div>
</header>
