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

  // Clamp position so menu doesn't overflow the window
  $: adjustedX = Math.min(x, window.innerWidth - 220);
  $: adjustedY = Math.min(y, window.innerHeight - (items.length * 36 + 16));
</script>

{#if visible}
  <div
    class="context-menu"
    style="top: {adjustedY}px; left: {adjustedX}px;"
    bind:this={menuEl}
  >
    {#each items as item}
      {#if item.separator}
        <div class="separator"></div>
      {:else}
        <button
          class="menu-item"
          class:disabled={item.disabled}
          class:checked={item.checked}
          on:click={() => handleItemClick(item)}
        >
          {#if item.icon}
            <span class="item-icon">{item.icon}</span>
          {/if}
          <span class="item-label">{item.label}</span>
          {#if item.checked}
            <span class="item-check">✓</span>
          {/if}
        </button>
      {/if}
    {/each}
  </div>
{/if}

<style>
  .context-menu {
    position: fixed;
    z-index: 1000;
    min-width: 200px;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 4px 0;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
  }

  .menu-item {
    display: flex;
    align-items: center;
    width: 100%;
    padding: 8px 14px;
    border: none;
    background: none;
    color: var(--text-primary);
    font-family: inherit;
    font-size: 13px;
    cursor: pointer;
    text-align: left;
    gap: 8px;
  }

  .menu-item:hover:not(.disabled) {
    background: var(--bg-hover);
  }

  .menu-item.disabled {
    color: var(--text-muted);
    cursor: not-allowed;
  }

  .menu-item.checked {
    color: var(--accent);
  }

  .item-icon {
    width: 18px;
    text-align: center;
    flex-shrink: 0;
    font-size: 14px;
  }

  .item-label {
    flex: 1;
  }

  .item-check {
    color: var(--accent);
    font-weight: 700;
    font-size: 14px;
  }

  .separator {
    height: 1px;
    background: var(--border);
    margin: 4px 0;
  }
</style>
