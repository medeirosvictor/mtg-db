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

        // Notify parent to reload
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

<div class="modal-backdrop" on:click={() => dispatch('close')} on:keydown={(e) => e.key === 'Escape' && dispatch('close')}>
  <div class="modal" on:click|stopPropagation role="dialog" aria-modal="true">
    <div class="modal-header">
      <h2>📥 Import Deck</h2>
      <button class="close-btn" on:click={() => dispatch('close')}>×</button>
    </div>

    <div class="modal-body">
      {#if success}
        <div class="success">{success}</div>
        <div class="actions">
          <button class="btn btn-primary" on:click={() => dispatch('close')}>Done</button>
        </div>
      {:else}
        <div class="form-group">
          <label for="import-input">Paste a deck URL or card list</label>
          <textarea
            id="import-input"
            bind:value={input}
            placeholder="Paste one of:&#10;&#10;• Moxfield URL: https://moxfield.com/decks/abc123&#10;• Archidekt URL: https://archidekt.com/decks/12345&#10;• Card list:&#10;  1x Sol Ring&#10;  1x Dark Ritual&#10;  3x Forest"
            rows="10"
            disabled={loading}
          ></textarea>
          {#if inputHint}
            <p class="input-hint">{inputHint}</p>
          {/if}
        </div>

        <div class="form-group">
          <label>Import mode</label>
          <div class="mode-toggle">
            <button
              class="mode-btn"
              class:active={mode === 'merge'}
              on:click={() => mode = 'merge'}
              disabled={loading}
            >
              <strong>Merge</strong>
              <span>Add to existing cards</span>
            </button>
            <button
              class="mode-btn"
              class:active={mode === 'replace'}
              on:click={() => mode = 'replace'}
              disabled={loading}
            >
              <strong>Replace</strong>
              <span>Clear deck first</span>
            </button>
          </div>
        </div>

        {#if error}
          <div class="error">{error}</div>
        {/if}

        <div class="actions">
          <button type="button" class="btn btn-secondary" on:click={() => dispatch('close')} disabled={loading}>
            Cancel
          </button>
          <button class="btn btn-primary" on:click={handleImport} disabled={loading || !input.trim()}>
            {loading ? 'Importing...' : `Import (${mode})`}
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

  .form-group {
    margin-bottom: 16px;
  }

  label {
    display: block;
    font-size: 13px;
    font-weight: 600;
    color: var(--text-secondary);
    margin-bottom: 8px;
  }

  textarea {
    width: 100%;
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--text-primary);
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
    font-size: 13px;
    padding: 12px;
    resize: vertical;
    line-height: 1.5;
  }

  textarea:focus {
    outline: none;
    border-color: var(--accent);
  }

  textarea::placeholder {
    color: var(--text-muted);
    font-family: inherit;
  }

  .input-hint {
    font-size: 12px;
    color: var(--accent);
    margin-top: 6px;
  }

  .mode-toggle {
    display: flex;
    gap: 8px;
  }

  .mode-btn {
    flex: 1;
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 10px 14px;
    cursor: pointer;
    text-align: left;
    font-family: inherit;
    color: var(--text-secondary);
    transition: all 0.15s ease;
  }

  .mode-btn:hover:not(:disabled) {
    border-color: var(--accent);
  }

  .mode-btn.active {
    border-color: var(--accent);
    background: rgba(137, 180, 250, 0.1);
    color: var(--text-primary);
  }

  .mode-btn strong {
    display: block;
    font-size: 13px;
    margin-bottom: 2px;
  }

  .mode-btn span {
    font-size: 11px;
    color: var(--text-muted);
  }

  .mode-btn.active span {
    color: var(--accent);
  }

  .error {
    background: rgba(243, 139, 168, 0.1);
    border: 1px solid rgba(243, 139, 168, 0.3);
    color: var(--red);
    padding: 10px 14px;
    border-radius: var(--radius);
    font-size: 13px;
    margin-bottom: 16px;
  }

  .success {
    background: rgba(166, 227, 161, 0.1);
    border: 1px solid rgba(166, 227, 161, 0.3);
    color: var(--green);
    padding: 14px;
    border-radius: var(--radius);
    font-size: 14px;
    margin-bottom: 16px;
    text-align: center;
  }

  .actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }

  .btn {
    padding: 10px 18px;
    border-radius: var(--radius);
    font-size: 14px;
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
</style>
