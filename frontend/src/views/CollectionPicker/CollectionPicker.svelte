<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { CollectionInfo } from '../../lib/types';
  import { SelectCollectionFolder, InitializeAndSelectFolder, SwitchCollection, RemoveKnownCollection } from '../../../wailsjs/go/app/App';

  export let collections: CollectionInfo[] = [];
  export let invalidPath: string = '';

  const dispatch = createEventDispatcher();

  let error = '';
  let loading = false;

  async function selectExisting() {
    loading = true;
    error = '';
    try {
      const result = await SelectCollectionFolder();
      if (result === '') {
        dispatch('collectionChanged');
      } else {
        error = result;
      }
    } catch (e) {
      error = `Error: ${e}`;
    } finally {
      loading = false;
    }
  }

  async function initializeNew() {
    loading = true;
    error = '';
    try {
      const result = await InitializeAndSelectFolder();
      if (result === '') {
        dispatch('collectionChanged');
      } else {
        error = result;
      }
    } catch (e) {
      error = `Error: ${e}`;
    } finally {
      loading = false;
    }
  }

  async function switchTo(path: string) {
    loading = true;
    error = '';
    try {
      const result = await SwitchCollection(path);
      if (result === '') {
        dispatch('collectionChanged');
      } else {
        error = result;
      }
    } catch (e) {
      error = `Error: ${e}`;
    } finally {
      loading = false;
    }
  }

  async function removeCollection(path: string) {
    try {
      const result = await RemoveKnownCollection(path);
      if (result === '') {
        dispatch('collectionChanged');
      } else {
        error = result;
      }
    } catch (e) {
      error = `Error: ${e}`;
    }
  }

  $: validCollections = (collections || []).filter(c => c.isValid);
</script>

<div class="flex items-center justify-center min-h-screen p-8">
  <div class="max-w-[540px] w-full text-center">
    <div class="text-6xl mb-4">🃏</div>
    <h1 class="text-[28px] font-bold mb-2">MTG Collection Manager</h1>
    <p class="text-text-secondary mb-8">Choose a collection folder to get started</p>

    {#if invalidPath}
      <div class="flex gap-3 text-left p-4 bg-yellow/10 border border-yellow/25 rounded-lg mb-6">
        <span class="text-xl flex-shrink-0">⚠️</span>
        <div>
          <strong class="text-yellow">Collection not found</strong>
          <p class="text-sm text-text-secondary mt-1">The previously saved path no longer exists or is invalid:</p>
          <code class="bg-bg-surface px-1.5 py-0.5 rounded text-xs inline-block mt-1 break-all">{invalidPath}</code>
        </div>
      </div>
    {/if}

    {#if error}
      <div class="flex gap-3 text-left p-4 bg-red/10 border border-red/25 rounded-lg mb-6">
        <span class="text-xl flex-shrink-0">❌</span>
        <div>
          <strong class="text-red">Invalid folder</strong>
          <p class="text-sm text-text-secondary mt-1">{error}</p>
          <p class="text-xs text-text-muted mt-2">Expected a folder with a <code>decks/</code> subfolder containing deck folders with <code>deck.txt</code> files.</p>
        </div>
      </div>
    {/if}

    <div class="flex flex-col gap-3 mb-8">
      <button
        class="flex items-center justify-center gap-2.5 px-6 py-3.5 bg-accent text-bg-primary rounded-lg text-[15px] font-semibold hover:bg-accent-hover transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        on:click={selectExisting}
        disabled={loading}
      >
        <span class="text-lg">📂</span>
        Open Collection Folder
      </button>
      <button
        class="flex items-center justify-center gap-2.5 px-6 py-3.5 bg-bg-surface text-text-primary border border-border rounded-lg text-[15px] font-semibold hover:bg-bg-hover hover:border-accent transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        on:click={initializeNew}
        disabled={loading}
      >
        <span class="text-lg">✨</span>
        Create New Collection
      </button>
    </div>

    {#if validCollections.length > 0}
      <div class="text-left mb-8">
        <h3 class="text-sm font-semibold text-text-muted uppercase tracking-wide mb-3">Recent Collections</h3>
        <div class="flex flex-col gap-2">
          {#each validCollections as col}
            <div class="flex items-center gap-1">
              <button
                class="flex-1 flex flex-col items-start p-3 bg-bg-secondary border border-border rounded-lg hover:bg-bg-hover hover:border-accent transition-all cursor-pointer"
                on:click={() => switchTo(col.path)}
              >
                <span class="text-sm font-semibold">{col.label}</span>
                <span class="text-xs text-text-muted mt-0.5 break-all">{col.path}</span>
              </button>
              <button
                class="w-8 h-8 flex items-center justify-center bg-transparent border border-transparent rounded-lg text-text-muted hover:bg-red/10 hover:border-red/30 hover:text-red transition-all text-lg"
                title="Remove from list"
                on:click|stopPropagation={() => removeCollection(col.path)}
              >×</button>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <div class="text-left p-4 bg-bg-secondary border border-border rounded-lg">
      <p class="text-xs text-text-muted mb-2">A collection folder should look like:</p>
      <pre class="text-xs text-text-secondary bg-bg-surface p-3 rounded-md overflow-x-auto leading-relaxed">my-collection/
  decks/
    my-deck/
      deck.txt
      info.md
    another-deck/
      deck.txt</pre>
    </div>
  </div>
</div>
