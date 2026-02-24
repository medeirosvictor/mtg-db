<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { AddCards } from '../../wailsjs/go/main/App';

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

<div class="modal-backdrop" on:click={() => dispatch('close')} on:keydown={(e) => e.key === 'Enter' && dispatch('close')}>
  <div class="modal" on:click|stopPropagation role="dialog" aria-modal="true">
    <div class="modal-header">
      <h2>Add Cards</h2>
      <button class="close-btn" on:click={() => dispatch('close')}>×</button>
    </div>

    <form on:submit|preventDefault={handleSubmit}>
      <div class="form-group">
        <label for="card-input">Enter cards (one per line)</label>
        <textarea
          id="card-input"
          bind:value={cardInput}
          placeholder="1x Sol Ring&#10;1x Dark Ritual&#10;3x Forest"
          rows="8"
          disabled={loading}
        ></textarea>
        <p class="hint">Format: 1x Card Name or 1 Card Name</p>
      </div>

      {#if error}
        <div class="error">{error}</div>
      {/if}

      <div class="actions">
        <button type="button" class="btn btn-secondary" on:click={() => dispatch('close')} disabled={loading}>
          Cancel
        </button>
        <button type="submit" class="btn btn-primary" disabled={loading || !cardInput.trim()}>
          {loading ? 'Adding...' : 'Add Cards'}
        </button>
      </div>
    </form>
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
    max-width: 480px;
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

  form {
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
    font-family: inherit;
    font-size: 14px;
    padding: 12px;
    resize: vertical;
  }

  textarea:focus {
    outline: none;
    border-color: var(--accent);
  }

  textarea::placeholder {
    color: var(--text-muted);
  }

  .hint {
    font-size: 12px;
    color: var(--text-muted);
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
