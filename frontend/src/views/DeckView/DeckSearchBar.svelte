<script lang="ts">
  export let searchQuery: string;
  export let count: number;
  export let total: number;

  let searchInput: HTMLInputElement;

  export function focus() {
    searchInput?.focus();
    searchInput?.select();
  }
</script>

<div class="search-bar">
  <span class="search-icon">🔍</span>
  <input
    type="text"
    class="search-input"
    bind:this={searchInput}
    bind:value={searchQuery}
    placeholder="Filter cards by name, type, text, tags...  (Ctrl+F)"
  />
  {#if searchQuery}
    <button class="search-clear" on:click={() => { searchQuery = ''; searchInput?.focus(); }} title="Clear search">✕</button>
    <span class="search-count">{count} / {total}</span>
  {/if}
</div>

<style>
  .search-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 8px 14px;
    margin-bottom: 16px;
    transition: border-color 0.15s ease;
  }

  .search-bar:focus-within {
    border-color: var(--accent);
    box-shadow: 0 0 0 2px rgba(137, 180, 250, 0.15);
  }

  .search-icon {
    font-size: 14px;
    flex-shrink: 0;
  }

  .search-input {
    flex: 1;
    background: none;
    border: none;
    color: var(--text-primary);
    font-family: inherit;
    font-size: 14px;
    outline: none;
  }

  .search-input::placeholder {
    color: var(--text-muted);
  }

  .search-clear {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 14px;
    padding: 2px 6px;
    border-radius: 4px;
    flex-shrink: 0;
  }

  .search-clear:hover {
    color: var(--text-primary);
    background: var(--bg-hover);
  }

  .search-count {
    font-size: 12px;
    color: var(--text-muted);
    flex-shrink: 0;
    white-space: nowrap;
  }
</style>
