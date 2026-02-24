<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { CollectionInfo } from '../lib/types';
  import { SelectCollectionFolder, SwitchCollection } from '../../wailsjs/go/app/App';

  export let collections: CollectionInfo[] = [];
  export let activeLabel: string = '';
  export let activePath: string = '';

  const dispatch = createEventDispatcher();

  let open = false;
  let loading = false;

  function toggleDropdown() {
    open = !open;
  }

  function closeDropdown() {
    open = false;
  }

  async function switchTo(path: string) {
    if (path === activePath) {
      open = false;
      return;
    }
    loading = true;
    try {
      const result = await SwitchCollection(path);
      if (result === '') {
        dispatch('collectionChanged');
      }
    } catch (e) {
      console.error('Switch failed:', e);
    } finally {
      loading = false;
      open = false;
    }
  }

  async function openFolder() {
    loading = true;
    try {
      const result = await SelectCollectionFolder();
      if (result === '') {
        dispatch('collectionChanged');
      }
    } catch (e) {
      console.error('Open folder failed:', e);
    } finally {
      loading = false;
      open = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && open) {
      closeDropdown();
    }
  }

  function handleClickOutside(e: MouseEvent) {
    const target = e.target as HTMLElement;
    if (!target.closest('.switcher')) {
      closeDropdown();
    }
  }

  $: otherCollections = (collections || []).filter(c => c.path !== activePath && c.isValid);
</script>

<svelte:window on:click={handleClickOutside} on:keydown={handleKeydown} />

<div class="relative">
  <button 
    class="flex items-center gap-2 px-3 py-1.5 bg-bg-surface border border-border rounded-lg text-text-primary font-inherit text-sm hover:bg-bg-hover hover:border-accent transition-all disabled:opacity-50 disabled:cursor-not-allowed"
    on:click|stopPropagation={toggleDropdown} 
    disabled={loading}
  >
    <span class="text-sm">📁</span>
    <span class="font-semibold max-w-[200px] overflow-hidden text-ellipsis whitespace-nowrap">{activeLabel || 'Collection'}</span>
    <span class="text-[10px] text-text-muted transition-transform {open ? 'rotate-180' : ''}">▾</span>
  </button>

  {#if open}
    <div 
      class="absolute top-full left-0 mt-1 min-w-[280px] bg-bg-secondary border border-border rounded-lg shadow-lg z-10 p-1" 
      role="menu" 
      tabindex="-1" 
      on:click|stopPropagation 
      on:keydown|stopPropagation
    >
      {#if otherCollections.length > 0}
        <div class="py-1">
          <div class="px-3.5 py-1.5 text-[11px] font-semibold text-text-muted uppercase tracking-wide">Switch Collection</div>
          {#each otherCollections as col}
            <button 
              class="flex flex-col items-start w-full px-3.5 py-2 border-none bg-transparent text-text-primary font-inherit text-sm cursor-pointer text-left hover:bg-bg-hover"
              on:click={() => switchTo(col.path)}
            >
              <span class="font-medium">{col.label}</span>
              <span class="text-xs text-text-muted mt-0.5 break-all">{col.path}</span>
            </button>
          {/each}
        </div>
        <div class="h-px bg-border my-1"></div>
      {/if}
      <button 
        class="flex items-center gap-2 w-full px-3.5 py-2 border-none bg-transparent text-text-primary font-inherit text-sm cursor-pointer text-left hover:bg-bg-hover"
        on:click={openFolder}
      >
        <span class="text-sm">📂</span>
        <span class="font-medium">Open another folder…</span>
      </button>
    </div>
  {/if}
</div>
