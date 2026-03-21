<script lang="ts">
  import type { Card } from '../lib/types';
  import { createEventDispatcher } from 'svelte';

  export let card: Card;

  const dispatch = createEventDispatcher();

  let showBack = false;

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      dispatch('close');
    }
  }

  // Mana cost symbols → rendered pips
  // Parses "{W}{U}{3}" etc. into displayable tokens
  function parseManaCost(raw: string): string[] {
    if (!raw) return [];
    const matches = raw.match(/\{[^}]+\}/g);
    return matches || [];
  }

  function manaSymbolClass(symbol: string): string {
    const s = symbol.replace(/[{}]/g, '').toUpperCase();
    switch (s) {
      case 'W': return 'bg-[#F9FAF4] text-[#1a1a1a]';
      case 'U': return 'bg-[#0E68AB] text-white';
      case 'B': return 'bg-[#150B00] text-[#ccc] border border-[#555]';
      case 'R': return 'bg-[#D3202A] text-white';
      case 'G': return 'bg-[#00733E] text-white';
      case 'C': return 'bg-[#9ca3af] text-[#1a1a1a]';
      default:  return 'bg-[#6c7086] text-white'; // generic/colorless numbers
    }
  }

  function manaSymbolLabel(symbol: string): string {
    return symbol.replace(/[{}]/g, '');
  }

  function getCardType(typeLine: string): string {
    if (!typeLine) return '';
    // Just the main type before the dash
    const dash = typeLine.indexOf('—');
    return dash > 0 ? typeLine.substring(0, dash).trim() : typeLine;
  }

  function getCardSubtype(typeLine: string): string {
    if (!typeLine) return '';
    const dash = typeLine.indexOf('—');
    return dash > 0 ? typeLine.substring(dash + 1).trim() : '';
  }

  function formatOracleText(text: string): string {
    if (!text) return '';
    // Replace mana symbols with readable text
    return text.replace(/\{([^}]+)\}/g, '[$1]');
  }

  $: isDFC = card.scryFall?.isDoubleFaced;
  $: frontImage = card.scryFall?.imageUri;
  $: backImage = card.scryFall?.backImageUri;
  $: displayImage = showBack && backImage ? backImage : frontImage;
  $: manaCost = parseManaCost(card.scryFall?.manaCost || '');
  $: typeLine = card.scryFall?.typeLine || '';
  $: oracleText = card.scryFall?.oracleText || '';
  $: cmc = card.scryFall?.cmc;
  $: priceUsd = card.scryFall?.priceUsd;
  $: priceUsdFoil = card.scryFall?.priceUsdFoil;
  $: tags = card.tags || [];
  $: isProxy = tags.includes('proxy');
  $: isCommander = tags.includes('commander');
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- svelte-ignore a11y-click-events-have-key-events -->
<div 
  class="fixed inset-0 bg-black/70 flex items-center justify-center z-50 backdrop-blur-sm"
  on:click={() => dispatch('close')}
>
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div 
    class="bg-bg-secondary border-2 border-border rounded w-full max-w-[720px] shadow-2xl flex flex-col max-h-[90vh]"
    on:click|stopPropagation
  >
    <!-- Header -->
    <div class="flex items-center justify-between px-5 py-3 border-b border-border flex-shrink-0">
      <div class="flex items-center gap-2.5">
        {#if isCommander}
          <span title="Commander">👑</span>
        {/if}
        <h2 class="text-lg m-0 text-text-primary">{card.name}</h2>
        {#if manaCost.length > 0}
          <div class="flex gap-0.5 items-center ml-1">
            {#each manaCost as symbol}
              <span 
                class="w-5 h-5 rounded-full text-[10px] flex items-center justify-center flex-shrink-0 {manaSymbolClass(symbol)}"
                title={manaSymbolLabel(symbol)}
              >
                {manaSymbolLabel(symbol)}
              </span>
            {/each}
          </div>
        {/if}
      </div>
      <button 
        class="bg-transparent border-none text-text-muted text-2xl cursor-pointer p-0 leading-none hover:text-text-primary"
        on:click={() => dispatch('close')}
      >×</button>
    </div>

    <!-- Body -->
    <div class="flex gap-6 p-5 overflow-y-auto">
      <!-- Card image -->
      <div class="flex-shrink-0 flex flex-col items-center gap-2">
        <div class="w-[260px] rounded overflow-hidden bg-bg-surface shadow-lg">
          {#if displayImage}
            <img 
              src={displayImage} 
              alt={card.name} 
              class="w-full h-auto block"
            />
          {:else}
            <div class="aspect-[488/680] flex items-center justify-center text-4xl text-text-muted bg-gradient-to-br from-bg-surface to-bg-secondary">
              {card.name.substring(0, 2).toUpperCase()}
            </div>
          {/if}
        </div>
        {#if isDFC && backImage}
          <button
            class="bg-bg-surface border border-border text-text-secondary px-3 py-1.5 rounded text-xs cursor-pointer hover:bg-bg-hover hover:text-text-primary hover:border-accent"
            on:click={() => showBack = !showBack}
          >
            🔄 {showBack ? 'Show Front' : 'Show Back'}
          </button>
        {/if}
      </div>

      <!-- Card details -->
      <div class="flex-1 flex flex-col gap-4 min-w-0">
        <!-- Type line -->
        {#if typeLine}
          <div>
            <span class="text-[11px] uppercase tracking-wider text-text-muted">Type</span>
            <p class="text-sm text-text-primary mt-0.5">{typeLine}</p>
          </div>
        {/if}

        <!-- Oracle text -->
        {#if oracleText}
          <div>
            <span class="text-[11px] uppercase tracking-wider text-text-muted">Oracle Text</span>
            <div class="mt-1 text-sm text-text-secondary leading-relaxed whitespace-pre-line bg-bg-surface rounded px-3 py-2.5 border border-border">
              {formatOracleText(oracleText)}
            </div>
          </div>
        {/if}

        <!-- Stats row -->
        <div class="flex flex-wrap gap-4">
          {#if cmc !== undefined && cmc !== null}
            <div>
              <span class="text-[11px] uppercase tracking-wider text-text-muted">CMC</span>
              <p class="text-sm text-text-primary mt-0.5">{cmc}</p>
            </div>
          {/if}
          <div>
            <span class="text-[11px] uppercase tracking-wider text-text-muted">Quantity</span>
            <p class="text-sm text-text-primary mt-0.5">{card.quantity}×</p>
          </div>
          {#if card.setCode}
            <div>
              <span class="text-[11px] uppercase tracking-wider text-text-muted">Set</span>
              <p class="text-sm text-text-primary mt-0.5 uppercase">{card.setCode}{#if card.collectorNumber} #{card.collectorNumber}{/if}</p>
            </div>
          {/if}
          {#if card.foil}
            <div>
              <span class="text-[11px] uppercase tracking-wider text-text-muted">Finish</span>
              <p class="text-sm text-text-primary mt-0.5">✨ Foil</p>
            </div>
          {/if}
        </div>

        <!-- Price -->
        {#if !isProxy && (priceUsd || priceUsdFoil)}
          <div>
            <span class="text-[11px] uppercase tracking-wider text-text-muted">Price</span>
            <div class="flex gap-4 mt-0.5">
              {#if priceUsd}
                <p class="text-sm text-green">${priceUsd} <span class="text-text-muted font-normal">USD</span></p>
              {/if}
              {#if priceUsdFoil}
                <p class="text-sm text-yellow">${priceUsdFoil} <span class="text-text-muted font-normal">foil</span></p>
              {/if}
            </div>
          </div>
        {/if}

        <!-- Tags -->
        {#if tags.length > 0}
          <div>
            <span class="text-[11px] uppercase tracking-wider text-text-muted">Tags</span>
            <div class="flex gap-1.5 mt-1 flex-wrap">
              {#each tags as tag}
                <span class="inline-block px-2 py-0.5 rounded text-[11px]
                  {tag === 'commander' ? 'bg-yellow/20 text-yellow border border-yellow/30' : ''}
                  {tag === 'proxy' ? 'bg-orange/20 text-orange border border-orange/30' : ''}
                  {tag === 'wishlist' ? 'bg-accent/20 text-accent border border-accent/30' : ''}
                  {!['commander', 'proxy', 'wishlist'].includes(tag) ? 'bg-bg-surface text-text-secondary border border-border' : ''}
                ">
                  {tag === 'commander' ? '👑 ' : ''}{tag === 'proxy' ? '🖨️ ' : ''}{tag === 'wishlist' ? '🛒 ' : ''}{tag}
                </span>
              {/each}
            </div>
          </div>
        {/if}

        {#if isProxy}
          <div class="bg-orange/10 border border-orange/30 text-orange px-3 py-2 rounded text-xs">
            🖨️ This card is marked as a proxy — price is excluded from deck total.
          </div>
        {/if}
      </div>
    </div>

    <!-- Footer -->
    <div class="flex justify-end px-5 py-3 border-t border-border flex-shrink-0">
      <button 
        class="px-4 py-2 bg-bg-surface text-text-primary border border-border rounded text-sm hover:bg-bg-hover"
        on:click={() => dispatch('close')}
      >
        Close
      </button>
    </div>
  </div>
</div>
