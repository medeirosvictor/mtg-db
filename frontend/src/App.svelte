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
  <div class="flex flex-col h-screen">
    <header class="flex items-center justify-between px-6 py-3 bg-bg-secondary border-b border-border">
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
  <div class="flex flex-col h-screen">
    <header class="flex items-center justify-between px-6 py-3 bg-bg-secondary border-b border-border">
      <CollectionSwitcher 
        collections={appState.collections || []}
        activeLabel={appState.collectionLabel}
        activePath={appState.collectionPath}
        on:collectionChanged={handleCollectionChanged}
      />
      <button 
        class="bg-transparent border-none text-text-secondary font-inherit text-sm px-3 py-1.5 rounded hover:bg-bg-hover hover:text-text-primary cursor-pointer"
        on:click={() => currentView = 'home'}
      >← Back</button>
    </header>
    {#key appState.collectionPath + currentSlug}
      <DeckView slug={currentSlug} on:navigate={handleNavigate} />
    {/key}
  </div>
{/if}
