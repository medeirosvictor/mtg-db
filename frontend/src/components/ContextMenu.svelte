<script lang="ts">
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';

  export let x: number = 0;
  export let y: number = 0;
  export let visible: boolean = false;

  interface MenuItem {
    label: string;
    icon?: string;
    action: () => void;
    checked?: boolean;
    separator?: boolean;
    disabled?: boolean;
  }

  export let items: MenuItem[] = [];

  const dispatch = createEventDispatcher();

  let menuEl: HTMLDivElement;

  function handleClickOutside(e: MouseEvent) {
    if (menuEl && !menuEl.contains(e.target as Node)) {
      close();
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      close();
    }
  }

  function close() {
    visible = false;
    dispatch('close');
  }

  function handleItemClick(item: MenuItem) {
    if (item.disabled) return;
    item.action();
    close();
  }

  onMount(() => {
    document.addEventListener('click', handleClickOutside);
    document.addEventListener('keydown', handleKeydown);
  });

  onDestroy(() => {
    document.removeEventListener('click', handleClickOutside);
    document.removeEventListener('keydown', handleKeydown);
  });

  $: adjustedX = Math.min(x, window.innerWidth - 220);
  $: adjustedY = Math.min(y, window.innerHeight - (items.length * 36 + 16));
</script>

{#if visible}
  <div
    class="fixed z-50 min-w-[200px] bg-bg-secondary border border-border rounded-lg p-1 shadow-lg"
    style="top: {adjustedY}px; left: {adjustedX}px;"
    bind:this={menuEl}
  >
    {#each items as item}
      {#if item.separator}
        <div class="h-px bg-border my-1"></div>
      {:else}
        <button
          class="flex items-center w-full px-3.5 py-2 border-none bg-transparent text-text-primary font-inherit text-sm cursor-pointer text-left gap-2 {item.disabled ? 'text-text-muted cursor-not-allowed' : 'hover:bg-bg-hover'} {item.checked ? 'text-accent' : ''}"
          class:cursor-not-allowed={item.disabled}
          class:text-text-muted={item.disabled}
          on:click={() => handleItemClick(item)}
        >
          {#if item.icon}
            <span class="w-[18px] text-center flex-shrink-0 text-sm">{item.icon}</span>
          {/if}
          <span class="flex-1">{item.label}</span>
          {#if item.checked}
            <span class="text-accent font-bold text-sm">✓</span>
          {/if}
        </button>
      {/if}
    {/each}
  </div>
{/if}
