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

<div class="flex items-center gap-2 bg-bg-secondary border border-border rounded-lg px-3.5 py-2 mb-4 transition-all focus-within:border-accent focus-within:shadow-[0_0_0_2px_rgba(137,180,250,0.15)]">
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
      class="bg-transparent border-none text-text-muted cursor-pointer text-sm px-1.5 py-0.5 rounded hover:text-text-primary hover:bg-bg-hover flex-shrink-0 transition-colors" 
      on:click={() => { searchQuery = ''; searchInput?.focus(); }} 
      title="Clear search"
    >✕</button>
    <span class="text-xs text-text-muted flex-shrink-0 whitespace-nowrap">{count} / {total}</span>
  {/if}
</div>
