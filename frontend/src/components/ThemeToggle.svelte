<script lang="ts">
  import { onMount } from 'svelte';

  let theme: 'dark' | 'light' = 'dark';

  onMount(() => {
    const saved = localStorage.getItem('mtg-db-theme');
    if (saved === 'light' || saved === 'dark') {
      theme = saved;
    }
    document.documentElement.setAttribute('data-theme', theme);
  });

  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('mtg-db-theme', theme);
  }
</script>

<button
  class="flex items-center justify-center w-8 h-8 bg-bg-surface border-2 border-border rounded text-sm cursor-pointer hover:bg-bg-hover hover:border-accent transition-colors"
  on:click={toggleTheme}
  title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
>
  {#if theme === 'dark'}
    <span class="text-yellow">☀️</span>
  {:else}
    <span class="text-accent">🌙</span>
  {/if}
</button>
