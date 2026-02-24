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

<div class="flex justify-end gap-3 mt-3">
  <button
    class="bg-bg-surface border border-border text-text-secondary px-3 py-1.5 rounded-lg text-xs font-inherit cursor-pointer hover:bg-bg-hover hover:border-accent hover:text-accent transition-all"
    on:click={() => dispatch('import')}
    title="Import cards from URL or text"
  >
    📥 Import
  </button>
  <button
    class="bg-bg-surface border border-border text-text-secondary px-3 py-1.5 rounded-lg text-xs font-inherit cursor-pointer hover:bg-bg-hover hover:border-accent hover:text-accent transition-all"
    on:click={() => dispatch('export')}
    title="Export deck as text"
  >
    📤 Export
  </button>
  <button 
    class="bg-bg-surface border border-border text-text-secondary px-3 py-1.5 rounded-lg text-xs font-inherit cursor-pointer hover:bg-bg-hover hover:border-accent hover:text-accent transition-all disabled:opacity-50 disabled:cursor-not-allowed" 
    on:click={() => dispatch('sync')}
    disabled={loading || scryfallLoading}
    title="Fetch card data from Scryfall"
  >
    {loading || scryfallLoading ? 'Loading...' : '↻ Sync Scryfall'}
  </button>
  <div class="flex gap-0.5 bg-bg-surface border border-border rounded-lg p-0.5">
    <button 
      class="px-2.5 py-1 rounded text-sm transition-all {viewMode === 'list' ? 'bg-accent text-bg-primary shadow-sm' : 'bg-transparent text-text-muted hover:text-text-primary'}" 
      on:click={() => dispatch('viewChange', 'list')}
      title="List view"
    >☰</button>
    <button 
      class="px-2.5 py-1 rounded text-sm transition-all {viewMode === 'grid' ? 'bg-accent text-bg-primary shadow-sm' : 'bg-transparent text-text-muted hover:text-text-primary'}" 
      on:click={() => dispatch('viewChange', 'grid')}
      title="Grid view"
    >⊞</button>
  </div>
</div>
