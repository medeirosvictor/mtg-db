<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { AddCards } from '../../wailsjs/go/app/App';

  export let slug: string;

  const dispatch = createEventDispatcher();

  let cardInput = '';
  let loading = false;
  let error = '';

  async function handleSubmit() {
    if (!cardInput.trim()) return;

    loading = true;
    error = '';

    try {
      const result = await AddCards(slug, cardInput);
      if (result === '') {
        dispatch('cardsAdded');
        dispatch('close');
      } else {
        error = result;
      }
    } catch (e) {
      error = `Error: ${e}`;
    } finally {
      loading = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      dispatch('close');
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div 
  class="fixed inset-0 bg-black/60 flex items-center justify-center z-50"
  on:click={() => dispatch('close')} 
  on:keydown={(e) => e.key === 'Enter' && dispatch('close')}
>
  <div 
    class="bg-bg-secondary border border-border rounded-lg w-full max-w-[480px] shadow-2xl" 
    on:click|stopPropagation 
    role="dialog" 
    aria-modal="true"
  >
    <div class="flex items-center justify-between px-5 py-4 border-b border-border">
      <h2 class="text-lg font-semibold m-0">Add Cards</h2>
      <button 
        class="bg-transparent border-none text-text-muted text-2xl cursor-pointer p-0 leading-none hover:text-text-primary"
        on:click={() => dispatch('close')}
      >×</button>
    </div>

    <form on:submit|preventDefault={handleSubmit} class="p-5">
      <div class="mb-4">
        <label for="card-input" class="block text-sm font-semibold text-text-secondary mb-2">Enter cards (one per line)</label>
        <textarea
          id="card-input"
          class="w-full bg-bg-surface border border-border rounded-lg text-text-primary font-inherit text-sm px-3 py-2.5 resize-y focus:outline-none focus:border-accent placeholder:text-text-muted"
          bind:value={cardInput}
          placeholder="1x Sol Ring&#10;1x Dark Ritual&#10;3x Forest"
          rows="8"
          disabled={loading}
        ></textarea>
        <p class="text-xs text-text-muted mt-1.5">Format: 1x Card Name or 1 Card Name</p>
      </div>

      {#if error}
        <div class="bg-red/10 border border-red/30 text-red px-3.5 py-2.5 rounded-lg text-sm mb-4">{error}</div>
      {/if}

      <div class="flex gap-3 justify-end">
        <button 
          type="button" 
          class="px-4 py-2.5 bg-bg-surface text-text-primary border border-border rounded-lg text-sm font-semibold hover:bg-bg-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          on:click={() => dispatch('close')} 
          disabled={loading}
        >
          Cancel
        </button>
        <button 
          type="submit" 
          class="px-4 py-2.5 bg-accent text-bg-primary rounded-lg text-sm font-semibold hover:bg-accent-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          disabled={loading || !cardInput.trim()}
        >
          {loading ? 'Adding...' : 'Add Cards'}
        </button>
      </div>
    </form>
  </div>
</div>
