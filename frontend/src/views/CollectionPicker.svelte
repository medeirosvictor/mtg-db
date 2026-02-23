<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { CollectionInfo } from '../lib/types';
  import { SelectCollectionFolder, InitializeAndSelectFolder, SwitchCollection, RemoveKnownCollection } from '../../wailsjs/go/main/App';

  export let collections: CollectionInfo[] = [];
  export let invalidPath: string = '';

  const dispatch = createEventDispatcher();

  let error = '';
  let loading = false;

  async function selectExisting() {
    loading = true;
    error = '';
    try {
      const result = await SelectCollectionFolder();
      if (result === '') {
        // Success or user cancelled — check if we now have a collection
        dispatch('collectionChanged');
      } else {
        error = result;
      }
    } catch (e) {
      error = `Error: ${e}`;
    } finally {
      loading = false;
    }
  }

  async function initializeNew() {
    loading = true;
    error = '';
    try {
      const result = await InitializeAndSelectFolder();
      if (result === '') {
        dispatch('collectionChanged');
      } else {
        error = result;
      }
    } catch (e) {
      error = `Error: ${e}`;
    } finally {
      loading = false;
    }
  }

  async function switchTo(path: string) {
    loading = true;
    error = '';
    try {
      const result = await SwitchCollection(path);
      if (result === '') {
        dispatch('collectionChanged');
      } else {
        error = result;
      }
    } catch (e) {
      error = `Error: ${e}`;
    } finally {
      loading = false;
    }
  }

  async function removeCollection(path: string) {
    try {
      const result = await RemoveKnownCollection(path);
      if (result === '') {
        dispatch('collectionChanged');
      } else {
        error = result;
      }
    } catch (e) {
      error = `Error: ${e}`;
    }
  }

  $: validCollections = (collections || []).filter(c => c.isValid);
</script>

<div class="picker">
  <div class="picker-content">
    <div class="logo">🃏</div>
    <h1>MTG Collection Manager</h1>
    <p class="subtitle">Choose a collection folder to get started</p>

    {#if invalidPath}
      <div class="warning-box">
        <span class="warning-icon">⚠️</span>
        <div>
          <strong>Collection not found</strong>
          <p>The previously saved path no longer exists or is invalid:</p>
          <code>{invalidPath}</code>
        </div>
      </div>
    {/if}

    {#if error}
      <div class="error-box">
        <span class="error-icon">❌</span>
        <div>
          <strong>Invalid folder</strong>
          <p>{error}</p>
          <p class="hint">Expected a folder with a <code>decks/</code> subfolder containing deck folders with <code>deck.txt</code> files.</p>
        </div>
      </div>
    {/if}

    <div class="actions">
      <button class="btn btn-primary" on:click={selectExisting} disabled={loading}>
        <span class="btn-icon">📂</span>
        Open Collection Folder
      </button>
      <button class="btn btn-secondary" on:click={initializeNew} disabled={loading}>
        <span class="btn-icon">✨</span>
        Create New Collection
      </button>
    </div>

    {#if validCollections.length > 0}
      <div class="recent">
        <h3>Recent Collections</h3>
        <div class="collection-list">
          {#each validCollections as col}
            <div class="collection-item">
              <button class="collection-btn" on:click={() => switchTo(col.path)}>
                <span class="collection-label">{col.label}</span>
                <span class="collection-path">{col.path}</span>
              </button>
              <button
                class="remove-btn"
                title="Remove from list"
                on:click|stopPropagation={() => removeCollection(col.path)}
              >×</button>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <div class="help-text">
      <p>A collection folder should look like:</p>
      <pre>my-collection/
  decks/
    my-deck/
      deck.txt
      info.md
    another-deck/
      deck.txt</pre>
    </div>
  </div>
</div>

<style>
  .picker {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    padding: 32px;
  }

  .picker-content {
    max-width: 540px;
    width: 100%;
    text-align: center;
  }

  .logo {
    font-size: 64px;
    margin-bottom: 16px;
  }

  h1 {
    font-size: 28px;
    font-weight: 700;
    margin-bottom: 8px;
  }

  .subtitle {
    color: var(--text-secondary);
    font-size: 15px;
    margin-bottom: 32px;
  }

  .warning-box, .error-box {
    display: flex;
    gap: 12px;
    text-align: left;
    padding: 16px;
    border-radius: var(--radius);
    margin-bottom: 24px;
  }

  .warning-box {
    background: rgba(249, 226, 175, 0.08);
    border: 1px solid rgba(249, 226, 175, 0.25);
  }

  .error-box {
    background: rgba(243, 139, 168, 0.08);
    border: 1px solid rgba(243, 139, 168, 0.25);
  }

  .warning-icon, .error-icon {
    font-size: 20px;
    flex-shrink: 0;
  }

  .warning-box strong {
    color: var(--yellow);
  }

  .error-box strong {
    color: var(--red);
  }

  .warning-box p, .error-box p {
    font-size: 13px;
    color: var(--text-secondary);
    margin-top: 4px;
  }

  .warning-box code, .error-box code {
    background: var(--bg-surface);
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 12px;
    display: inline-block;
    margin-top: 4px;
    word-break: break-all;
  }

  .hint {
    font-size: 12px !important;
    color: var(--text-muted) !important;
    margin-top: 8px !important;
  }

  .actions {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 32px;
  }

  .btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    padding: 14px 24px;
    border-radius: var(--radius);
    border: none;
    font-size: 15px;
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
  }

  .btn-primary:hover:not(:disabled) {
    background: var(--accent-hover);
    transform: translateY(-1px);
  }

  .btn-secondary {
    background: var(--bg-surface);
    color: var(--text-primary);
    border: 1px solid var(--border);
  }

  .btn-secondary:hover:not(:disabled) {
    background: var(--bg-hover);
    border-color: var(--accent);
    transform: translateY(-1px);
  }

  .btn-icon {
    font-size: 18px;
  }

  .recent {
    text-align: left;
    margin-bottom: 32px;
  }

  .recent h3 {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 12px;
  }

  .collection-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .collection-item {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .collection-btn {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    padding: 12px 16px;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    cursor: pointer;
    font-family: inherit;
    color: var(--text-primary);
    transition: all 0.15s ease;
  }

  .collection-btn:hover {
    background: var(--bg-hover);
    border-color: var(--accent);
  }

  .collection-label {
    font-size: 14px;
    font-weight: 600;
  }

  .collection-path {
    font-size: 11px;
    color: var(--text-muted);
    margin-top: 2px;
    word-break: break-all;
  }

  .remove-btn {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: none;
    border: 1px solid transparent;
    border-radius: var(--radius);
    color: var(--text-muted);
    cursor: pointer;
    font-size: 18px;
    font-family: inherit;
    flex-shrink: 0;
  }

  .remove-btn:hover {
    background: rgba(243, 139, 168, 0.1);
    border-color: rgba(243, 139, 168, 0.3);
    color: var(--red);
  }

  .help-text {
    text-align: left;
    padding: 16px;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
  }

  .help-text p {
    font-size: 12px;
    color: var(--text-muted);
    margin-bottom: 8px;
  }

  .help-text pre {
    font-size: 12px;
    color: var(--text-secondary);
    background: var(--bg-surface);
    padding: 12px;
    border-radius: 6px;
    overflow-x: auto;
    line-height: 1.6;
  }
</style>
