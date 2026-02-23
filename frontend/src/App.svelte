<script lang="ts">
  import DeckList from './views/DeckList.svelte';
  import DeckView from './views/DeckView.svelte';

  let currentView = 'home';
  let currentSlug = '';

  function handleNavigate(event: CustomEvent<{ view: string; slug?: string }>) {
    currentView = event.detail.view;
    currentSlug = event.detail.slug || '';
  }
</script>

{#if currentView === 'home'}
  <DeckList on:navigate={handleNavigate} />
{:else if currentView === 'deck'}
  <DeckView slug={currentSlug} on:navigate={handleNavigate} />
{/if}
