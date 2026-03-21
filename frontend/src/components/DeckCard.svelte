<script lang="ts">
  import type { DeckSummary } from '../lib/types';
  import ColorPips from './ColorPips.svelte';
  import StatusBadge from './StatusBadge.svelte';
  import { createEventDispatcher } from 'svelte';

  export let deck: DeckSummary;
  export let onClick: () => void;

  const dispatch = createEventDispatcher<{
    contextmenu: MouseEvent;
  }>();

  $: cardCountClass = deck.cardCount === 100 ? 'text-green' : 'text-yellow';
  $: hasImage = !!deck.commanderImageUri;
</script>

<button
  class="deck-card w-full text-left bg-bg-secondary border-2 border-border rounded p-5 hover:bg-bg-hover hover:border-accent cursor-pointer text-text-primary font-inherit relative"
  on:click={onClick}
  on:contextmenu|preventDefault={(e) => dispatch('contextmenu', e)}
>
  <!-- Image bg in its own clipping container so it doesn't overflow the card -->
  {#if hasImage}
    <div class="absolute inset-0 rounded overflow-hidden pointer-events-none">
      <div
        class="absolute inset-0 bg-cover bg-top bg-no-repeat"
        style="background-image: url({deck.commanderImageUri});"
      ></div>
    </div>
  {/if}

  <div class="relative z-10">
    <div class="flex justify-between items-start gap-2 mb-3">
      <h3 class="text-base leading-snug">{deck.title}</h3>
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
        <span class="text-lg font-mono {cardCountClass}">{deck.cardCount}</span>
        <span class="text-[10px] text-text-muted uppercase tracking-wide">cards</span>
      </div>
    </div>

    {#if deck.universe}
      <div class="mt-2 pt-2 border-t border-border">
        <span class="text-xs text-text-muted">{deck.universe}</span>
      </div>
    {/if}
  </div>
</button>

<style>
  .deck-card .absolute .absolute {
    opacity: 0.08;
  }
  .deck-card:hover .absolute .absolute {
    opacity: 0.14;
  }
  :global([data-theme="light"]) .deck-card .absolute .absolute {
    opacity: 0.06;
  }
  :global([data-theme="light"]) .deck-card:hover .absolute .absolute {
    opacity: 0.10;
  }
</style>
