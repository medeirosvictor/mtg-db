<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let loading: boolean;
  export let scryfallLoading: boolean;
  export let viewMode: 'list' | 'grid';

  const dispatch = createEventDispatcher<{
    import: void;
    export: void;
    sync: void;
    viewChange: 'list' | 'grid';
  }>();
</script>

<div class="header-toolbar">
  <button
    class="toolbar-btn"
    on:click={() => dispatch('import')}
    title="Import cards from URL or text"
  >
    📥 Import
  </button>
  <button
    class="toolbar-btn"
    on:click={() => dispatch('export')}
    title="Export deck as text"
  >
    📤 Export
  </button>
  <button 
    class="toolbar-btn sync-btn" 
    on:click={() => dispatch('sync')}
    disabled={loading || scryfallLoading}
    title="Fetch card data from Scryfall"
  >
    {loading || scryfallLoading ? 'Loading...' : '↻ Sync Scryfall'}
  </button>
  <div class="view-toggle">
    <button 
      class="toggle-btn" 
      class:active={viewMode === 'list'} 
      on:click={() => dispatch('viewChange', 'list')}
      title="List view"
    >☰</button>
    <button 
      class="toggle-btn" 
      class:active={viewMode === 'grid'} 
      on:click={() => dispatch('viewChange', 'grid')}
      title="Grid view"
    >⊞</button>
  </div>
</div>

<style>
  .header-toolbar {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 12px;
  }

  .toolbar-btn {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    color: var(--text-secondary);
    padding: 6px 12px;
    border-radius: var(--radius);
    cursor: pointer;
    font-size: 12px;
    font-family: inherit;
    transition: all 0.15s ease;
  }

  .toolbar-btn:hover:not(:disabled) {
    background: var(--bg-hover);
    border-color: var(--accent);
    color: var(--accent);
  }

  .toolbar-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .view-toggle {
    display: flex;
    gap: 2px;
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 2px;
  }

  .toggle-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    padding: 6px 10px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
    font-family: inherit;
  }

  .toggle-btn:hover {
    color: var(--text-primary);
  }

  .toggle-btn.active {
    background: var(--accent);
    color: #11111b;
  }
</style>
