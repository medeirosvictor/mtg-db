<script lang="ts">
  import { onMount } from 'svelte';
  import DeckList from './views/DeckList.svelte';
  import DeckView from './views/DeckView.svelte';
  import CollectionPicker from './views/CollectionPicker.svelte';
  import CollectionSwitcher from './components/CollectionSwitcher.svelte';
  import { GetAppState } from '../../wailsjs/go/app/App';
  import type { AppState, CollectionInfo } from './lib/types';

  let appState: AppState | null = null;
  let currentView = 'home';
  let currentSlug = '';

  onMount(async () => {
    try {
      appState = await GetAppState();
    } catch (e) {
      console.error('Failed to get app state:', e);
    }
  });

  function handleNavigate(event: CustomEvent<{ view: string; slug?: string }>) {
    currentView = event.detail.view;
    currentSlug = event.detail.slug || '';
  }

  async function handleCollectionChanged() {
    appState = await GetAppState();
  }
</script>

{#if !appState || appState.needsSetup || !appState.collectionValid}
  <CollectionPicker 
    collections={appState?.collections || []} 
    invalidPath={appState && !appState.collectionValid ? appState.collectionPath : ''}
    on:collectionChanged={handleCollectionChanged}
  />
{:else if currentView === 'home'}
  <div class="app-layout">
    <header class="app-header">
      <CollectionSwitcher 
        collections={appState.collections || []}
        activeLabel={appState.collectionLabel}
        activePath={appState.collectionPath}
        on:collectionChanged={handleCollectionChanged}
      />
    </header>
    {#key appState.collectionPath}
      <DeckList on:navigate={handleNavigate} />
    {/key}
  </div>
{:else if currentView === 'deck'}
  <div class="app-layout">
    <header class="app-header">
      <CollectionSwitcher 
        collections={appState.collections || []}
        activeLabel={appState.collectionLabel}
        activePath={appState.collectionPath}
        on:collectionChanged={handleCollectionChanged}
      />
      <button class="back-btn" on:click={() => currentView = 'home'}>← Back</button>
    </header>
    {#key appState.collectionPath + currentSlug}
      <DeckView slug={currentSlug} on:navigate={handleNavigate} />
    {/key}
  </div>
{/if}

<style>
  .app-layout {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }

  .app-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 24px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border);
  }

  .back-btn {
    background: none;
    border: none;
    color: var(--text-secondary);
    font-family: inherit;
    font-size: 14px;
    cursor: pointer;
    padding: 6px 12px;
    border-radius: var(--radius);
  }

  .back-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }
</style>
