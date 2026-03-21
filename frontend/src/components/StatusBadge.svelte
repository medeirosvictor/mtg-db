<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let status: string;

  const dispatch = createEventDispatcher();

  let open = false;

  const statuses = ['Owned', 'Planned', 'Disassembled'];

  $: badgeClass = {
    'Owned': 'bg-green text-bg-primary',
    'Planned': 'bg-yellow text-bg-primary',
    'Disassembled': 'bg-text-muted text-bg-primary',
  }[status] || 'bg-text-muted text-bg-primary';

  function toggleDropdown() {
    open = !open;
  }

  function selectStatus(newStatus: string) {
    dispatch('change', newStatus);
    open = false;
  }

  function handleClickOutside(e: MouseEvent) {
    const target = e.target as HTMLElement;
    if (!target.closest('.status-dropdown')) {
      open = false;
    }
  }
</script>

<svelte:window on:click={handleClickOutside} />

<div class="relative status-dropdown">
  <button 
    class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] uppercase tracking-wide cursor-pointer {badgeClass} hover:opacity-80"
    on:click|stopPropagation={toggleDropdown}
  >
    {status}
    <span class="text-[8px]">▾</span>
  </button>

  {#if open}
    <div class="absolute top-full left-0 mt-1 bg-bg-secondary border border-border rounded shadow-lg z-50 py-1 min-w-[120px]">
      {#each statuses as s}
        <button 
          class="w-full text-left px-3 py-1.5 text-xs uppercase tracking-wide hover:bg-bg-hover
            {s === status ? 'text-accent' : 'text-text-primary'}"
          on:click|stopPropagation={() => selectStatus(s)}
        >
          {s}
          {#if s === status} ✓{/if}
        </button>
      {/each}
    </div>
  {/if}
</div>
