<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let loading: boolean;
  export let scryfallLoading: boolean;
  export let viewMode: 'list' | 'grid';

  const dispatch = createEventDispatcher<{
    import: void;
    export: void;
    sync: void;
    openFolder: void;
    viewChange: 'list' | 'grid';
  }>();
</script>

<div class="flex gap-3 mt-3">
  <button
    class="bg-bg-surface border-2 border-border text-text-secondary px-3 py-1.5 rounded text-xs font-inherit cursor-pointer hover:bg-bg-hover hover:border-accent hover:text-accent"
    on:click={() => dispatch('openFolder')}
    title="Open deck folder in file explorer"
  >
    📂 Open folder
  </button>
  <button
    class="bg-bg-surface border-2 border-border text-text-secondary px-3 py-1.5 rounded text-xs font-inherit cursor-pointer hover:bg-bg-hover hover:border-accent hover:text-accent"
    on:click={() => dispatch('import')}
    title="Import cards from URL or text"
  >
    📥 Import
  </button>
  <button
    class="bg-bg-surface border-2 border-border text-text-secondary px-3 py-1.5 rounded text-xs font-inherit cursor-pointer hover:bg-bg-hover hover:border-accent hover:text-accent"
    on:click={() => dispatch('export')}
    title="Export deck as text"
  >
    📤 Export
  </button>
  <button 
    class="bg-bg-surface border-2 border-border text-text-secondary px-3 py-1.5 rounded text-xs font-inherit cursor-pointer hover:bg-bg-hover hover:border-accent hover:text-accent disabled:opacity-50 disabled:cursor-not-allowed" 
    on:click={() => dispatch('sync')}
    disabled={loading || scryfallLoading}
    title="Fetch card data from Scryfall"
  >
    {loading || scryfallLoading ? 'Loading...' : '↻ Sync Scryfall'}
  </button>
  <div class="flex gap-0.5 bg-bg-surface border-2 border-border rounded p-0.5">
    <button 
      class="px-2.5 py-1 rounded text-sm {viewMode === 'list' ? 'bg-accent text-bg-primary' : 'bg-transparent text-text-muted hover:text-text-primary'}" 
      on:click={() => dispatch('viewChange', 'list')}
      title="List view"
    >☰</button>
    <button 
      class="px-2.5 py-1 rounded text-sm {viewMode === 'grid' ? 'bg-accent text-bg-primary' : 'bg-transparent text-text-muted hover:text-text-primary'}" 
      on:click={() => dispatch('viewChange', 'grid')}
      title="Grid view"
    >⊞</button>
  </div>
</div>
