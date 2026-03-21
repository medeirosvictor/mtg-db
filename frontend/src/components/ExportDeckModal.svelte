<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { ExportDeck } from '../../wailsjs/go/app/App';

  export let slug: string;
  export let deckTitle: string = '';

  const dispatch = createEventDispatcher();

  let deckText = '';
  let loading = true;
  let copied = false;
  let copyTimeout: ReturnType<typeof setTimeout>;

  onMount(async () => {
    try {
      deckText = await ExportDeck(slug);
    } catch (e) {
      deckText = `Error loading deck: ${e}`;
    } finally {
      loading = false;
    }
  });

  async function copyToClipboard() {
    try {
      await navigator.clipboard.writeText(deckText);
      copied = true;
      clearTimeout(copyTimeout);
      copyTimeout = setTimeout(() => { copied = false; }, 2000);
    } catch (e) {
      const textarea = document.createElement('textarea');
      textarea.value = deckText;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
      copied = true;
      clearTimeout(copyTimeout);
      copyTimeout = setTimeout(() => { copied = false; }, 2000);
    }
  }

  function downloadTxt() {
    const filename = slug + '.txt';
    const blob = new Blob([deckText], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      dispatch('close');
    }
  }

  $: lineCount = deckText.split('\n').filter(l => l.trim().length > 0).length;
</script>

<svelte:window on:keydown={handleKeydown} />

<div 
  class="fixed inset-0 bg-black/60 flex items-center justify-center z-50"
  on:click={() => dispatch('close')} 
  on:keydown={(e) => e.key === 'Escape' && dispatch('close')}
>
  <div 
    class="bg-bg-secondary border-2 border-border rounded w-full max-w-[560px] shadow-2xl" 
    on:click|stopPropagation 
    role="dialog" 
    aria-modal="true"
  >
    <div class="flex items-center justify-between px-5 py-4 border-b border-border">
      <h2 class="text-lg m-0">📤 Export Deck</h2>
      <button 
        class="bg-transparent border-none text-text-muted text-2xl cursor-pointer p-0 leading-none hover:text-text-primary"
        on:click={() => dispatch('close')}
      >×</button>
    </div>

    <div class="p-5">
      {#if loading}
        <div class="text-center py-10 text-text-muted">Loading deck...</div>
      {:else}
        <div class="flex items-center justify-between mb-3">
          {#if deckTitle}
            <span class="text-sm text-text-primary">{deckTitle}</span>
          {/if}
          <span class="text-xs text-text-muted">{lineCount} cards</span>
        </div>

        <div class="mb-4">
          <textarea
            class="w-full bg-bg-surface border border-border rounded text-text-primary font-mono text-xs px-3 py-2.5 resize-y leading-relaxed focus:outline-none focus:border-accent"
            readonly
            value={deckText}
            rows="16"
          ></textarea>
        </div>

        <div class="flex gap-2.5 justify-end">
          <button 
            class="px-4 py-2.5 bg-bg-surface text-text-primary border border-border rounded text-sm hover:bg-bg-hover"
            on:click={() => dispatch('close')}
          >
            Close
          </button>
          <button 
            class="px-4 py-2.5 bg-bg-surface text-text-primary border border-border rounded text-sm hover:bg-bg-hover hover:border-accent"
            on:click={downloadTxt} 
            title="Download as .txt file"
          >
            💾 Download .txt
          </button>
          <button 
            class="px-4 py-2.5 bg-accent text-bg-primary rounded text-sm hover:bg-accent-hover"
            on:click={copyToClipboard} 
            title="Copy to clipboard"
          >
            {copied ? '✓ Copied!' : '📋 Copy to Clipboard'}
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>
