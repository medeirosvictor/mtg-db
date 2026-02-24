<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { CreateDeckFromImport } from '../../wailsjs/go/app/App';

  const dispatch = createEventDispatcher();

  let title = '';
  let input = '';
  let loading = false;
  let error = '';

  async function handleImport() {
    if (!title.trim() || !input.trim()) return;

    loading = true;
    error = '';

    try {
      const result = await CreateDeckFromImport(title.trim(), input.trim());
      if (result.startsWith('error:')) {
        error = result.slice(6);
      } else {
        dispatch('created', { slug: result });
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
      <h2 class="text-lg font-semibold m-0">📥 Import New Deck</h2>
      <button 
        class="bg-transparent border-none text-text-muted text-2xl cursor-pointer p-0 leading-none hover:text-text-primary"
        on:click={() => dispatch('close')}
      >×</button>
    </div>

    <div class="p-5">
      <div class="mb-4">
        <label for="deck-title" class="block text-sm font-semibold text-text-secondary mb-2">Deck title</label>
        <input
          id="deck-title"
          type="text"
          class="w-full bg-bg-surface border border-border rounded-lg text-text-primary font-inherit text-sm px-3 py-2.5 focus:outline-none focus:border-accent placeholder:text-text-muted"
          bind:value={title}
          placeholder="e.g. Simic Landfall"
          disabled={loading}
          autofocus
        />
        <p class="text-xs text-text-muted mt-1.5">Used as the deck folder name and display title</p>
      </div>

      <div class="mb-4">
        <label for="import-input" class="block text-sm font-semibold text-text-secondary mb-2">Deck URL or card list</label>
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
          disabled={loading || !title.trim() || !input.trim()}
        >
          {loading ? 'Importing...' : 'Create Deck'}
        </button>
      </div>
    </div>
  </div>
</div>
