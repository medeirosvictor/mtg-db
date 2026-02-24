<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { ImportDeck } from '../../wailsjs/go/app/App';

  export let slug: string;

  const dispatch = createEventDispatcher();

  let input = '';
  let mode: 'merge' | 'replace' = 'merge';
  let loading = false;
  let error = '';
  let success = '';

  async function handleImport() {
    if (!input.trim()) return;

    loading = true;
    error = '';
    success = '';

    try {
      const result = await ImportDeck(slug, input.trim(), mode);
      if (result.error) {
        error = result.error;
      } else {
        const count = (result.cards || []).reduce((sum: number, c: any) => sum + c.quantity, 0);
        const sourceLabel = result.source === 'moxfield' ? 'Moxfield' : result.source === 'archidekt' ? 'Archidekt' : 'text';
        const nameLabel = result.deckName ? ` "${result.deckName}"` : '';
        success = `Imported ${count} cards from ${sourceLabel}${nameLabel} (${mode === 'replace' ? 'replaced' : 'merged'})`;

        dispatch('imported');
      }
    } catch (e) {
      error = `Import failed: ${e}`;
    } finally {
      loading = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      dispatch('close');
    }
  }

  function detectInputType(text: string): string {
    const trimmed = text.trim();
    if (/moxfield\.com\/decks\//.test(trimmed)) return '🔗 Moxfield URL detected';
    if (/archidekt\.com\/decks\//.test(trimmed)) return '🔗 Archidekt URL detected';
    const lines = trimmed.split('\n').filter(l => l.trim().length > 0);
    if (lines.length > 0) return `📝 ${lines.length} line${lines.length > 1 ? 's' : ''} of card text`;
    return '';
  }

  $: inputHint = detectInputType(input);
</script>

<svelte:window on:keydown={handleKeydown} />

<div 
  class="fixed inset-0 bg-black/60 flex items-center justify-center z-50"
  on:click={() => dispatch('close')} 
  on:keydown={(e) => e.key === 'Escape' && dispatch('close')}
>
  <div 
    class="bg-bg-secondary border border-border rounded-lg w-full max-w-[560px] shadow-2xl" 
    on:click|stopPropagation 
    role="dialog" 
    aria-modal="true"
  >
    <div class="flex items-center justify-between px-5 py-4 border-b border-border">
      <h2 class="text-lg font-semibold m-0">📥 Import Deck</h2>
      <button 
        class="bg-transparent border-none text-text-muted text-2xl cursor-pointer p-0 leading-none hover:text-text-primary"
        on:click={() => dispatch('close')}
      >×</button>
    </div>

    <div class="p-5">
      {#if success}
        <div class="bg-green/10 border border-green/30 text-green px-4 py-3.5 rounded-lg text-sm text-center mb-4">{success}</div>
        <div class="flex justify-end">
          <button 
            class="px-4 py-2.5 bg-accent text-bg-primary rounded-lg text-sm font-semibold hover:bg-accent-hover transition-colors"
            on:click={() => dispatch('close')}
          >Done</button>
        </div>
      {:else}
        <div class="mb-4">
          <label for="import-input" class="block text-sm font-semibold text-text-secondary mb-2">Paste a deck URL or card list</label>
          <textarea
            id="import-input"
            class="w-full bg-bg-surface border border-border rounded-lg text-text-primary font-mono text-xs px-3 py-2.5 resize-y leading-relaxed focus:outline-none focus:border-accent placeholder:text-text-muted placeholder:font-inherit"
            bind:value={input}
            placeholder="Paste one of:&#10;&#10;• Moxfield URL: https://moxfield.com/decks/abc123&#10;• Archidekt URL: https://archidekt.com/decks/12345&#10;• Card list:&#10;  1x Sol Ring&#10;  1x Dark Ritual&#10;  3x Forest"
            rows="10"
            disabled={loading}
          ></textarea>
          {#if inputHint}
            <p class="text-xs text-accent mt-1.5">{inputHint}</p>
          {/if}
        </div>

        <div class="mb-4">
          <label class="block text-sm font-semibold text-text-secondary mb-2">Import mode</label>
          <div class="flex gap-2">
            <button
              class="flex-1 bg-bg-surface border rounded-lg px-3.5 py-2.5 cursor-pointer text-left font-inherit transition-all
                {mode === 'merge' ? 'border-accent bg-accent/10 text-text-primary' : 'border-border text-text-secondary hover:border-accent'}"
              on:click={() => mode = 'merge'}
              disabled={loading}
            >
              <strong class="block text-sm">Merge</strong>
              <span class="text-xs {mode === 'merge' ? 'text-accent' : 'text-text-muted'}">Add to existing cards</span>
            </button>
            <button
              class="flex-1 bg-bg-surface border rounded-lg px-3.5 py-2.5 cursor-pointer text-left font-inherit transition-all
                {mode === 'replace' ? 'border-accent bg-accent/10 text-text-primary' : 'border-border text-text-secondary hover:border-accent'}"
              on:click={() => mode = 'replace'}
              disabled={loading}
            >
              <strong class="block text-sm">Replace</strong>
              <span class="text-xs {mode === 'replace' ? 'text-accent' : 'text-text-muted'}">Clear deck first</span>
            </button>
          </div>
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
            class="px-4 py-2.5 bg-accent text-bg-primary rounded-lg text-sm font-semibold hover:bg-accent-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            on:click={handleImport} 
            disabled={loading || !input.trim()}
          >
            {loading ? 'Importing...' : `Import (${mode})`}
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>
