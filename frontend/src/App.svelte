<script lang="ts">
  import { onMount } from 'svelte';
  import DeckList from './views/DeckList/DeckList.svelte';
  import DeckView from './views/DeckView.svelte';
  import CollectionPicker from './views/CollectionPicker/CollectionPicker.svelte';
  import CollectionSwitcher from './components/CollectionSwitcher.svelte';
  import ThemeToggle from './components/ThemeToggle.svelte';
  import { GetAppState } from '../../wailsjs/go/app/App';
  import type { AppState, CollectionInfo } from './lib/types';

  let appState: AppState | null = null;
  let currentView = 'home';
  let currentSlug = '';
  let deckListKey = 0;

  onMount(async () => {
    try {
      appState = await GetAppState();
    } catch (e) {
      console.error('Failed to get app state:', e);
    }
  });

  function handleNavigate(event: CustomEvent<{ view: string; slug?: string }>) {
    const wasOnDeck = currentView === 'deck';
    currentView = event.detail.view;
    currentSlug = event.detail.slug || '';
    
    if (wasOnDeck && currentView === 'home') {
      deckListKey++;
    }
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
    <header class="flex items-center px-6 py-2 bg-bg-secondary border-b-2 border-border">
      <div class="flex items-center">
        <CollectionSwitcher 
          collections={appState.collections || []}
          activeLabel={appState.collectionLabel}
          activePath={appState.collectionPath}
          on:collectionChanged={handleCollectionChanged}
        />
      </div>
      <div class="flex-1 text-center">
        <span class="text-sm text-text-primary tracking-wide">🃏 MTG Collection</span>
      </div>
      <div class="flex items-center gap-3">
        <span class="text-[11px] text-text-muted truncate max-w-[250px]" title={appState.collectionPath}>{appState.collectionPath}</span>
        <ThemeToggle />
      </div>
    </header>
    {#key appState.collectionPath + deckListKey}
      <DeckList on:navigate={handleNavigate} />
    {/key}
  </div>
{:else if currentView === 'deck'}
  <div class="flex flex-col h-screen">
    <header class="flex items-center px-6 py-2 bg-bg-secondary border-b-2 border-border">
      <div class="flex items-center">
        <CollectionSwitcher 
          collections={appState.collections || []}
          activeLabel={appState.collectionLabel}
          activePath={appState.collectionPath}
          on:collectionChanged={handleCollectionChanged}
        />
      </div>
      <div class="flex-1 text-center">
        <button 
          class="bg-transparent border-none text-text-secondary font-inherit text-sm px-0 cursor-pointer hover:text-text-primary"
          on:click={() => currentView = 'home'}
        >🃏 MTG Collection</button>
        <span class="text-text-muted text-xs mx-1">/</span>
        <span class="text-sm text-text-primary">{currentSlug}</span>
      </div>
      <div class="flex items-center gap-3">
        <span class="text-[11px] text-text-muted truncate max-w-[250px]" title={appState.collectionPath}>{appState.collectionPath}</span>
        <ThemeToggle />
      </div>
    </header>
    {#key appState.collectionPath + currentSlug}
      <DeckView slug={currentSlug} on:navigate={handleNavigate} />
    {/key}
  </div>
{/if}
