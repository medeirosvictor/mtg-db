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

<div class="flex items-center gap-2 bg-bg-secondary border-2 border-border rounded px-3.5 py-2 my-3 focus-within:border-accent">
  <span class="text-sm flex-shrink-0">🔍</span>
  <input
    type="text"
    class="flex-1 bg-transparent border-none text-text-primary font-inherit text-sm outline-none placeholder:text-text-muted"
    bind:this={searchInput}
    bind:value={searchQuery}
    placeholder="Filter cards by name, type, text, tags...  (Ctrl+F)"
  />
  {#if searchQuery}
    <button 
      class="bg-transparent border-none text-text-muted cursor-pointer text-sm px-1.5 py-0.5 rounded hover:text-text-primary hover:bg-bg-hover flex-shrink-0" 
      on:click={() => { searchQuery = ''; searchInput?.focus(); }} 
      title="Clear search"
    >✕</button>
    <span class="text-xs text-text-muted flex-shrink-0 whitespace-nowrap">{count} / {total}</span>
  {/if}
</div>
