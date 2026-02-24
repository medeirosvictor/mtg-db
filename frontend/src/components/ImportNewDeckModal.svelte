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
        // result is the new deck slug
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

<div class="modal-backdrop" on:click={() => dispatch('close')} on:keydown={(e) => e.key === 'Escape' && dispatch('close')}>
  <div class="modal" on:click|stopPropagation role="dialog" aria-modal="true">
    <div class="modal-header">
      <h2>📥 Import New Deck</h2>
      <button class="close-btn" on:click={() => dispatch('close')}>×</button>
    </div>

    <div class="modal-body">
      <div class="form-group">
        <label for="deck-title">Deck title</label>
        <input
          id="deck-title"
          type="text"
          class="text-input"
          bind:value={title}
          placeholder="e.g. Simic Landfall"
          disabled={loading}
          autofocus
        />
        <p class="hint">Used as the deck folder name and display title</p>
      </div>

      <div class="form-group">
        <label for="import-input">Deck URL or card list</label>
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

      {#if error}
        <div class="error">{error}</div>
      {/if}

      <div class="actions">
        <button type="button" class="btn btn-secondary" on:click={() => dispatch('close')} disabled={loading}>
          Cancel
        </button>
        <button class="btn btn-primary" on:click={handleImport} disabled={loading || !title.trim() || !input.trim()}>
          {loading ? 'Importing...' : 'Create Deck'}
        </button>
      </div>
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

  .text-input {
    width: 100%;
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--text-primary);
    font-family: inherit;
    font-size: 14px;
    padding: 10px 12px;
  }

  .text-input:focus {
    outline: none;
    border-color: var(--accent);
  }

  .text-input::placeholder {
    color: var(--text-muted);
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

  .hint {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 6px;
  }

  .input-hint {
    font-size: 12px;
    color: var(--accent);
    margin-top: 6px;
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
