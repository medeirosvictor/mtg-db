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
      // Fallback for older browsers
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

<div class="modal-backdrop" on:click={() => dispatch('close')} on:keydown={(e) => e.key === 'Escape' && dispatch('close')}>
  <div class="modal" on:click|stopPropagation role="dialog" aria-modal="true">
    <div class="modal-header">
      <h2>📤 Export Deck</h2>
      <button class="close-btn" on:click={() => dispatch('close')}>×</button>
    </div>

    <div class="modal-body">
      {#if loading}
        <div class="loading">Loading deck...</div>
      {:else}
        <div class="deck-info">
          {#if deckTitle}
            <span class="deck-name">{deckTitle}</span>
          {/if}
          <span class="line-count">{lineCount} cards</span>
        </div>

        <div class="text-container">
          <textarea
            class="deck-text"
            readonly
            value={deckText}
            rows="16"
          ></textarea>
        </div>

        <div class="actions">
          <button class="btn btn-secondary" on:click={() => dispatch('close')}>
            Close
          </button>
          <button class="btn btn-action" on:click={downloadTxt} title="Download as .txt file">
            💾 Download .txt
          </button>
          <button class="btn btn-primary" on:click={copyToClipboard} title="Copy to clipboard">
            {copied ? '✓ Copied!' : '📋 Copy to Clipboard'}
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    width: 100%;
    max-width: 560px;
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4);
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 1px solid var(--border);
  }

  .modal-header h2 {
    font-size: 18px;
    font-weight: 600;
    margin: 0;
  }

  .close-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 24px;
    cursor: pointer;
    padding: 0;
    line-height: 1;
  }

  .close-btn:hover {
    color: var(--text-primary);
  }

  .modal-body {
    padding: 20px;
  }

  .loading {
    text-align: center;
    padding: 40px;
    color: var(--text-muted);
  }

  .deck-info {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
  }

  .deck-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .line-count {
    font-size: 12px;
    color: var(--text-muted);
  }

  .text-container {
    margin-bottom: 16px;
  }

  .deck-text {
    width: 100%;
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--text-primary);
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
    font-size: 12px;
    padding: 12px;
    resize: vertical;
    line-height: 1.6;
  }

  .deck-text:focus {
    outline: none;
    border-color: var(--accent);
  }

  .actions {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
  }

  .btn {
    padding: 10px 16px;
    border-radius: var(--radius);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    font-family: inherit;
    transition: all 0.15s ease;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background: var(--accent);
    color: #11111b;
    border: none;
  }

  .btn-primary:hover:not(:disabled) {
    background: var(--accent-hover);
  }

  .btn-secondary {
    background: var(--bg-surface);
    color: var(--text-primary);
    border: 1px solid var(--border);
  }

  .btn-secondary:hover:not(:disabled) {
    background: var(--bg-hover);
  }

  .btn-action {
    background: var(--bg-surface);
    color: var(--text-primary);
    border: 1px solid var(--border);
  }

  .btn-action:hover:not(:disabled) {
    background: var(--bg-hover);
    border-color: var(--accent);
  }
</style>
