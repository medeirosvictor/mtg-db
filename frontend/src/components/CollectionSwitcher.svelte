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

<div class="switcher">
  <button class="switcher-btn" on:click|stopPropagation={toggleDropdown} disabled={loading}>
    <span class="switcher-icon">📁</span>
    <span class="switcher-label">{activeLabel || 'Collection'}</span>
    <span class="switcher-chevron" class:open>▾</span>
  </button>

  {#if open}
    <div class="dropdown" role="menu" tabindex="-1" on:click|stopPropagation on:keydown|stopPropagation>
      {#if otherCollections.length > 0}
        <div class="dropdown-section">
          <div class="dropdown-heading">Switch Collection</div>
          {#each otherCollections as col}
            <button class="dropdown-item" on:click={() => switchTo(col.path)}>
              <span class="item-label">{col.label}</span>
              <span class="item-path">{col.path}</span>
            </button>
          {/each}
        </div>
        <div class="dropdown-divider"></div>
      {/if}
      <button class="dropdown-item action-item" on:click={openFolder}>
        <span class="item-icon">📂</span>
        <span class="item-label">Open another folder…</span>
      </button>
    </div>
  {/if}
</div>

<style>
  .switcher {
    position: relative;
  }

  .switcher-btn {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--text-primary);
    cursor: pointer;
    font-family: inherit;
    font-size: 13px;
    transition: all 0.15s ease;
  }

  .switcher-btn:hover:not(:disabled) {
    background: var(--bg-hover);
    border-color: var(--accent);
  }

  .switcher-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .switcher-icon {
    font-size: 14px;
  }

  .switcher-label {
    font-weight: 600;
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .switcher-chevron {
    font-size: 10px;
    color: var(--text-muted);
    transition: transform 0.15s ease;
  }

  .switcher-chevron.open {
    transform: rotate(180deg);
  }

  .dropdown {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    min-width: 280px;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
    z-index: 100;
    padding: 4px 0;
  }

  .dropdown-section {
    padding: 4px 0;
  }

  .dropdown-heading {
    padding: 6px 14px;
    font-size: 11px;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .dropdown-item {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    width: 100%;
    padding: 8px 14px;
    border: none;
    background: none;
    color: var(--text-primary);
    font-family: inherit;
    font-size: 13px;
    cursor: pointer;
    text-align: left;
  }

  .dropdown-item:hover {
    background: var(--bg-hover);
  }

  .dropdown-item.action-item {
    flex-direction: row;
    align-items: center;
    gap: 8px;
  }

  .item-label {
    font-weight: 500;
  }

  .item-path {
    font-size: 11px;
    color: var(--text-muted);
    margin-top: 1px;
    word-break: break-all;
  }

  .item-icon {
    font-size: 14px;
  }

  .dropdown-divider {
    height: 1px;
    background: var(--border);
    margin: 4px 0;
  }
</style>
