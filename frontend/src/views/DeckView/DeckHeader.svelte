<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ColorPips from '../../components/ColorPips.svelte';
  import StatusBadge from '../../components/StatusBadge.svelte';
  import type { Deck } from '../lib/types';

  export let deck: Deck;
  export let totalPrice: number;
  export let displayCommander: string;

  const dispatch = createEventDispatcher<{
    updateTitle: string;
    updateStrategy: string;
    updateStatus: string;
  }>();

  let editingTitle = false;
  let titleInput = '';
  let titleInputEl: HTMLInputElement;

  let editingStrategy = false;
  let strategyInput = '';
  let strategyInputEl: HTMLTextAreaElement;

  function startEditTitle() {
    titleInput = deck.info.title;
    editingTitle = true;
    setTimeout(() => titleInputEl?.focus(), 0);
  }

  function saveTitle() {
    if (titleInput.trim() && titleInput.trim() !== deck.info.title) {
      dispatch('updateTitle', titleInput.trim());
    }
    editingTitle = false;
  }

  function startEditStrategy() {
    strategyInput = deck.info.strategy || '';
    editingStrategy = true;
    setTimeout(() => strategyInputEl?.focus(), 0);
  }

  function saveStrategy() {
    if (strategyInput !== (deck.info.strategy || '')) {
      dispatch('updateStrategy', strategyInput);
    }
    editingStrategy = false;
  }

  function handleTitleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      saveTitle();
    } else if (e.key === 'Escape') {
      editingTitle = false;
    }
  }

  function handleStrategyKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      editingStrategy = false;
    }
  }

  function handleStatusChange(e: CustomEvent<string>) {
    dispatch('updateStatus', e.detail);
  }
</script>

<header class="mb-6 pb-4 border-b border-border">
  <div class="flex flex-col gap-2">
    <div class="flex items-center gap-3">
      {#if editingTitle}
        <input
          type="text"
          bind:this={titleInputEl}
          bind:value={titleInput}
          on:blur={saveTitle}
          on:keydown={handleTitleKeydown}
          class="text-2xl bg-bg-surface border-2 border-accent rounded px-2 py-1 text-text-primary outline-none"
        />
      {:else}
        <h1 
          class="text-2xl cursor-pointer hover:text-accent transition-colors"
          on:click={startEditTitle}
          title="Click to edit title"
        >
          {deck.info.title}
        </h1>
      {/if}
      <StatusBadge 
        status={deck.info.status.includes('Owned') ? 'Owned' : deck.info.status.includes('Planned') ? 'Planned' : 'Disassembled'} 
        on:change={handleStatusChange}
      />
    </div>
    <div class="flex items-center gap-4 flex-wrap">
      <ColorPips colors={deck.info.colors} />
      <span class="text-sm text-text-secondary">Commander: <strong class="text-accent">{displayCommander}</strong></span>
      <span class="text-sm {deck.cardCount !== 100 ? 'text-yellow' : 'text-green'}">
        {deck.cardCount} cards
      </span>
      <span class="text-sm text-green">
        ${totalPrice.toFixed(2)}
      </span>
    </div>
    {#if editingStrategy}
      <div class="mt-2">
        <textarea
          bind:this={strategyInputEl}
          bind:value={strategyInput}
          on:blur={saveStrategy}
          on:keydown={handleStrategyKeydown}
          class="w-full max-w-[800px] bg-bg-surface border-2 border-accent rounded px-3 py-2 text-sm text-text-secondary outline-none resize-y min-h-[60px]"
          placeholder="Add a strategy description..."
        ></textarea>
        <p class="text-xs text-text-muted mt-1">Press Esc to cancel</p>
      </div>
    {:else}
      <p 
        class="text-sm text-text-secondary leading-relaxed max-w-[800px] cursor-pointer hover:text-accent/70 transition-colors {deck.info.strategy ? '' : 'italic text-text-muted'}"
        on:click={startEditStrategy}
        title="Click to edit description"
      >
        {deck.info.strategy || 'Add a description...'}
      </p>
    {/if}
  </div>
</header>
