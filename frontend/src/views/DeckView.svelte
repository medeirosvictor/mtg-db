<script lang="ts">
  import { onMount } from 'svelte';
  import type { Deck, Card } from '../lib/types';
  import { GetDeck, GetDeckBasic, ToggleCardTag, UpdateCardText } from '../../wailsjs/go/app/App';
  import ColorPips from '../components/ColorPips.svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import ContextMenu from '../components/ContextMenu.svelte';
  import ImportDeckModal from '../components/ImportDeckModal.svelte';
  import ExportDeckModal from '../components/ExportDeckModal.svelte';
  import { createEventDispatcher } from 'svelte';

  export let slug: string;

  const dispatch = createEventDispatcher();

  let deck: Deck | null = null;
  let loading = true;
  let scryfallLoading = false;
  let error = '';
  let viewMode: 'list' | 'grid' = 'list';

  // Not-found cards from Scryfall sync
  let notFoundCards: Set<string> = new Set();

  // Inline editing state
  let editingCard: string | null = null;
  let editValue = '';
  let editError = '';

  // DFC flip state: tracks which cards are showing back face
  // Using an object (record) instead of Set for Svelte reactivity
  let flippedCards: Record<string, boolean> = {};

  function toggleFlip(cardName: string) {
    if (flippedCards[cardName]) {
      delete flippedCards[cardName];
    } else {
      flippedCards[cardName] = true;
    }
    flippedCards = flippedCards; // trigger reactivity
  }

  function isFlipped(cardName: string): boolean {
    return !!flippedCards[cardName];
  }

  // Search state
  let searchQuery = '';
  let searchInput: HTMLInputElement;

  // Modal state
  let showImportModal = false;
  let showExportModal = false;

  // Context menu state
  let menuVisible = false;
  let menuX = 0;
  let menuY = 0;
  let menuItems: any[] = [];

  function handleGlobalKeydown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
      e.preventDefault();
      searchInput?.focus();
      searchInput?.select();
    }
    if (e.key === 'Escape' && searchQuery) {
      searchQuery = '';
      searchInput?.blur();
    }
  }

  function fuzzyMatch(text: string, query: string): boolean {
    const lower = text.toLowerCase();
    const words = query.toLowerCase().split(/\s+/).filter(w => w.length > 0);
    return words.every(word => lower.includes(word));
  }

  onMount(async () => {
    await loadDeck();
  });

  async function loadDeck() {
    try {
      loading = true;
      error = '';
      const result = await GetDeckBasic(slug);
      deck = result;
      if (!deck) {
        error = `Deck "${slug}" not found`;
      }
    } catch (e) {
      console.error('Failed to load deck:', e);
      error = `Failed to load deck: ${e}`;
    } finally {
      loading = false;
    }
  }

  async function syncScryfall() {
    try {
      scryfallLoading = true;
      error = '';
      notFoundCards = new Set();
      console.log('Syncing with Scryfall...');
      const result = await GetDeck(slug);
      if (result) {
        deck = result.deck;
        if (result.notFound && result.notFound.length > 0) {
          notFoundCards = new Set(result.notFound.map(n => n.toLowerCase()));
          console.log('Cards not found on Scryfall:', result.notFound);
        }
      }
      console.log('Scryfall sync complete');
    } catch (e) {
      console.error('Failed to sync Scryfall:', e);
      error = `Scryfall sync failed: ${e}`;
    } finally {
      scryfallLoading = false;
    }
  }

  function isNotFound(name: string): boolean {
    return notFoundCards.has(name.toLowerCase());
  }

  function isBasicLand(name: string): boolean {
    const basics = ['plains', 'island', 'swamp', 'mountain', 'forest'];
    return basics.includes(name.toLowerCase());
  }

  // Format a card back to its text line representation
  function cardToText(card: Card): string {
    let line = `${card.quantity}x ${card.name}`;
    if (card.setCode) {
      line += ` (${card.setCode})`;
      if (card.collectorNumber) {
        line += ` ${card.collectorNumber}`;
      }
    }
    if (card.foil) {
      line += ' *F*';
    }
    if (card.tags && card.tags.length > 0) {
      for (const tag of card.tags) {
        line += ` #${tag}`;
      }
    }
    return line;
  }

  function startEditing(card: Card) {
    editingCard = card.name;
    editValue = cardToText(card);
    editError = '';
  }

  function cancelEditing() {
    editingCard = null;
    editValue = '';
    editError = '';
  }

  async function saveEditing(oldName: string) {
    if (!editValue.trim()) {
      editError = 'Card line cannot be empty';
      return;
    }

    try {
      const result = await UpdateCardText(slug, oldName, editValue);
      if (result === '') {
        editingCard = null;
        editValue = '';
        editError = '';
        // Reload deck to get updated data
        const reloaded = await GetDeckBasic(slug);
        if (reloaded) {
          deck = reloaded;
        }
      } else {
        editError = result;
      }
    } catch (e) {
      editError = `Error: ${e}`;
    }
  }

  function handleEditKeydown(e: KeyboardEvent, oldName: string) {
    if (e.key === 'Enter') {
      e.preventDefault();
      saveEditing(oldName);
    } else if (e.key === 'Escape') {
      cancelEditing();
    }
  }

  $: allSortedCards = deck?.cards
    ? [...deck.cards].sort((a, b) => {
        const aCmd = (a.tags || []).includes('commander');
        const bCmd = (b.tags || []).includes('commander');
        if (aCmd !== bCmd) return aCmd ? -1 : 1;
        const aBasic = isBasicLand(a.name);
        const bBasic = isBasicLand(b.name);
        if (aBasic !== bBasic) return aBasic ? 1 : -1;
        return a.name.localeCompare(b.name);
      })
    : [];

  $: sortedCards = searchQuery.trim()
    ? allSortedCards.filter(c => {
        const haystack = [
          c.name,
          c.setCode || '',
          c.scryFall?.typeLine || '',
          c.scryFall?.oracleText || '',
          ...(c.tags || []),
        ].join(' ');
        return fuzzyMatch(haystack, searchQuery);
      })
    : allSortedCards;

  $: allWishlistCards = deck?.wishlist || [];
  $: wishlistCards = searchQuery.trim()
    ? allWishlistCards.filter(c => fuzzyMatch(c.name, searchQuery))
    : allWishlistCards;
  $: commanders = deck?.cards?.filter(c => (c.tags || []).includes('commander')) || [];
  $: commanderNames = commanders.map(c => c.name).join(' / ');
  $: displayCommander = commanderNames || deck?.info?.commander || 'None set';

  $: totalPrice = deck?.cards?.reduce((sum, card) => {
    const isProxy = (card.tags || []).includes('proxy');
    if (isProxy) return sum;
    const price = card.scryFall?.priceUsd ? parseFloat(card.scryFall.priceUsd) : 0;
    return sum + (price * card.quantity);
  }, 0) || 0;

  async function toggleTag(cardName: string, tag: string) {
    if (!deck) return;
    const updated = await ToggleCardTag(slug, cardName, tag);
    if (updated) {
      deck = updated;
    }
  }

  function showCardContextMenu(e: MouseEvent, card: Card, _context: 'deck' | 'wishlist') {
    e.preventDefault();
    menuX = e.clientX;
    menuY = e.clientY;

    const tags = card.tags || [];
    const isCommander = tags.includes('commander');
    const isProxy = tags.includes('proxy');
    const isWishlisted = tags.includes('wishlist');
    const commanderCount = deck?.cards?.filter(c => (c.tags || []).includes('commander')).length || 0;

    menuItems = [
      {
        label: isCommander ? 'Remove as Commander' : 'Set as Commander',
        icon: '👑',
        checked: isCommander,
        disabled: !isCommander && commanderCount >= 2,
        action: () => toggleTag(card.name, 'commander'),
      },
      { separator: true },
      {
        label: isProxy ? 'Unmark as Proxy' : 'Mark as Proxy',
        icon: '🖨️',
        checked: isProxy,
        action: () => toggleTag(card.name, 'proxy'),
      },
      {
        label: isWishlisted ? 'Unmark as Wishlisted' : 'Mark as Wishlisted',
        icon: '🛒',
        checked: isWishlisted,
        action: () => toggleTag(card.name, 'wishlist'),
      },
      { separator: true },
      {
        label: 'Edit card text',
        icon: '✏️',
        action: () => startEditing(card),
      },
    ];

    menuVisible = true;
  }

  function getCardBadges(card: Card): string[] {
    const badges: string[] = [];
    const tags = card.tags || [];
    if (tags.includes('commander')) badges.push('commander');
    if (tags.includes('proxy')) badges.push('proxy');
    if (tags.includes('wishlist')) badges.push('wishlist');
    return badges;
  }
</script>

<svelte:window on:keydown={handleGlobalKeydown} />

<div class="deck-view">
  {#if loading}
    <div class="loading">
      <div class="loading-spinner"></div>
      <p>Loading deck...</p>
      <p class="loading-hint">Fetching card data from Scryfall</p>
    </div>
  {:else if error}
    <div class="error-container">
      <div class="error">{error}</div>
      <button class="retry-btn" on:click={loadDeck}>Try Again</button>
    </div>
  {:else if deck}
    <header class="deck-header">
      <button class="back-btn" on:click={() => dispatch('navigate', { view: 'home' })}>
        ← Back
      </button>
      <div class="header-info">
        <div class="title-row">
          <h1>{deck.info.title}</h1>
          <StatusBadge status={deck.info.status.includes('Owned') ? 'Owned' : deck.info.status.includes('Planned') ? 'Planned' : 'Disassembled'} />
        </div>
        <div class="meta-row">
          <ColorPips colors={deck.info.colors} />
          <span class="commander">Commander: <strong>{displayCommander}</strong></span>
          <span class="card-count" class:warn={deck.cardCount !== 100}>
            {deck.cardCount} cards
          </span>
          <span class="total-price">
            ${totalPrice.toFixed(2)}
          </span>
        </div>
        {#if deck.info.strategy}
          <p class="strategy">{deck.info.strategy}</p>
        {/if}
      </div>
      <div class="header-toolbar">
        <button
          class="toolbar-btn"
          on:click={() => showImportModal = true}
          title="Import cards from URL or text"
        >
          📥 Import
        </button>
        <button
          class="toolbar-btn"
          on:click={() => showExportModal = true}
          title="Export deck as text"
        >
          📤 Export
        </button>
        <button 
          class="toolbar-btn sync-btn" 
          on:click={syncScryfall}
          disabled={loading || scryfallLoading}
          title="Fetch card data from Scryfall"
        >
          {loading || scryfallLoading ? 'Loading...' : '↻ Sync Scryfall'}
        </button>
        <div class="view-toggle">
          <button 
            class="toggle-btn" 
            class:active={viewMode === 'list'} 
            on:click={() => viewMode = 'list'}
            title="List view"
          >☰</button>
          <button 
            class="toggle-btn" 
            class:active={viewMode === 'grid'} 
            on:click={() => viewMode = 'grid'}
            title="Grid view"
          >⊞</button>
        </div>
      </div>
    </header>

    <div class="search-bar">
      <span class="search-icon">🔍</span>
      <input
        type="text"
        class="search-input"
        bind:this={searchInput}
        bind:value={searchQuery}
        placeholder="Filter cards by name, type, text, tags...  (Ctrl+F)"
      />
      {#if searchQuery}
        <button class="search-clear" on:click={() => { searchQuery = ''; searchInput?.focus(); }} title="Clear search">✕</button>
        <span class="search-count">{sortedCards.length} / {allSortedCards.length}</span>
      {/if}
    </div>

    {#if notFoundCards.size > 0}
      <div class="not-found-banner">
        ⚠️ {notFoundCards.size} card{notFoundCards.size > 1 ? 's' : ''} not found on Scryfall. Check card names highlighted in red below.
      </div>
    {/if}

    <div class="content">
      {#if viewMode === 'list'}
        <section class="card-list">
          <h2>Cards ({sortedCards.length} unique, {deck.cardCount} total)</h2>
          <div class="cards-table">
            <div class="table-header">
              <span class="col-qty">#</span>
              <span class="col-name">Card Name</span>
              <span class="col-tags">Tags</span>
              <span class="col-price">Price</span>
              <span class="col-set">Set</span>
            </div>
            {#each sortedCards as card (card.name)}
              {#if editingCard === card.name}
                <div class="card-row editing-row">
                  <div class="edit-container">
                    <div class="edit-row">
                      <input
                        type="text"
                        class="edit-input"
                        bind:value={editValue}
                        on:keydown={(e) => handleEditKeydown(e, card.name)}
                        autofocus
                      />
                      <div class="edit-actions">
                        <button class="edit-btn save" on:click={() => saveEditing(card.name)} title="Save (Enter)">✓</button>
                        <button class="edit-btn cancel" on:click={cancelEditing} title="Cancel (Esc)">✕</button>
                      </div>
                    </div>
                    {#if editError}
                      <div class="edit-error">{editError}</div>
                    {/if}
                  </div>
                </div>
              {:else}
                <div
                  class="card-row"
                  class:basic-land={isBasicLand(card.name)}
                  class:is-commander={(card.tags || []).includes('commander')}
                  class:is-not-found={isNotFound(card.name)}
                  on:contextmenu={(e) => showCardContextMenu(e, card, 'deck')}
                  on:dblclick={() => startEditing(card)}
                >
                  <span class="col-qty">{card.quantity}×</span>
                  <span class="col-name">
                    {#if isNotFound(card.name)}
                      <span class="not-found-icon" title="Card not found on Scryfall">⚠️</span>
                    {/if}
                    {card.name}
                    {#if card.foil}
                      <span class="foil-tag">✨</span>
                    {/if}
                    {#if card.scryFall?.isDoubleFaced}
                      <button
                        class="flip-btn"
                        on:click|stopPropagation={() => toggleFlip(card.name)}
                        title={isFlipped(card.name) ? 'Show front face' : 'Show back face'}
                      >🔄</button>
                    {/if}
                  </span>
                  <span class="col-tags">
                    {#each getCardBadges(card) as badge}
                      <span class="card-badge card-badge-{badge}">
                        {#if badge === 'commander'}👑{/if}
                        {#if badge === 'proxy'}🖨️{/if}
                        {#if badge === 'wishlist'}🛒{/if}
                        {badge}
                      </span>
                    {/each}
                  </span>
                  <span class="col-price">
                    {#if (card.tags || []).includes('proxy')}
                      <span class="proxy-price" title="Proxy — price not tracked">—</span>
                    {:else}
                      {card.scryFall?.priceUsd ? '$' + card.scryFall.priceUsd : '-'}
                    {/if}
                  </span>
                  <span class="col-set">
                    {card.setCode || ''}
                  </span>
                </div>
              {/if}
            {/each}
          </div>
        </section>
      {:else}
        <section class="card-grid">
          <h2>Cards ({sortedCards.length} unique, {deck.cardCount} total)</h2>
          <div class="cards-grid">
            {#each sortedCards as card (card.name)}
              <div
                class="grid-card"
                class:basic-land={isBasicLand(card.name)}
                class:is-commander={(card.tags || []).includes('commander')}
                class:is-not-found={isNotFound(card.name)}
                on:contextmenu={(e) => showCardContextMenu(e, card, 'deck')}
              >
                <div class="card-image">
                  {console.log(card)}
                  {#if (isFlipped(card.name) && card.scryFall?.backImageUri) || card.scryFall?.imageUri}
                    <img src={isFlipped(card.name) && card.scryFall?.backImageUri ? card.scryFall.backImageUri : card.scryFall?.imageUri} alt={card.name} loading="lazy" />
                  {:else}
                    <div class="card-placeholder" class:placeholder-not-found={isNotFound(card.name)}>
                      {#if isNotFound(card.name)}
                        <span class="not-found-icon-large">⚠️</span>
                      {:else}
                        {card.name.substring(0, 2).toUpperCase()}
                      {/if}
                    </div>
                  {/if}
                  {#if card.scryFall?.isDoubleFaced}
                    <button
                      class="flip-btn-grid"
                      on:click|stopPropagation={() => toggleFlip(card.name)}
                      title={isFlipped(card.name) ? 'Show front face' : 'Show back face'}
                    >🔄</button>
                  {/if}
                </div>
                <div class="card-details">
                  <span class="card-name" class:card-name-not-found={isNotFound(card.name)} title={card.name}>
                    {#if isNotFound(card.name)}<span class="not-found-icon-sm">⚠️</span>{/if}
                    {card.name}
                  </span>
                  <span class="card-qty">
                    {card.quantity}×
                    {#if (card.tags || []).includes('proxy')}
                      <span class="proxy-price">proxy</span>
                    {:else if card.scryFall?.priceUsd}
                      ${card.scryFall.priceUsd}
                    {/if}
                  </span>
                </div>
              </div>
            {/each}
          </div>
        </section>
      {/if}

      {#if wishlistCards.length > 0}
        <section class="card-list wishlist">
          <h2>Wishlist ({wishlistCards.length})</h2>
          <div class="cards-table">
            <div class="table-header">
              <span class="col-qty">#</span>
              <span class="col-name">Card Name</span>
              <span class="col-tags">Tags</span>
              <span class="col-set">Set</span>
            </div>
            {#each wishlistCards as card (card.name)}
              <div
                class="card-row wishlist-row"
                on:contextmenu={(e) => showCardContextMenu(e, card, 'wishlist')}
              >
                <span class="col-qty">{card.quantity}×</span>
                <span class="col-name">{card.name}</span>
                <span class="col-tags">
                  {#each getCardBadges(card) as badge}
                    <span class="card-badge card-badge-{badge}">
                      {#if badge === 'proxy'}🖨️{/if}
                      {#if badge === 'wishlist'}🛒{/if}
                      {badge}
                    </span>
                  {/each}
                </span>
                <span class="col-set">
                  {card.setCode || ''}
                </span>
              </div>
            {/each}
          </div>
        </section>
      {/if}
    </div>
  {/if}
</div>

{#if showImportModal}
  <ImportDeckModal
    {slug}
    on:close={() => showImportModal = false}
    on:imported={async () => {
      showImportModal = false;
      const reloaded = await GetDeckBasic(slug);
      if (reloaded) deck = reloaded;
    }}
  />
{/if}

{#if showExportModal}
  <ExportDeckModal
    {slug}
    deckTitle={deck?.info?.title || slug}
    on:close={() => showExportModal = false}
  />
{/if}

<ContextMenu
  bind:visible={menuVisible}
  x={menuX}
  y={menuY}
  items={menuItems}
/>

<style>
  .deck-view {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
  }

  .deck-header {
    margin-bottom: 24px;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--border);
  }

  .header-toolbar {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 12px;
  }

  .toolbar-btn {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    color: var(--text-secondary);
    padding: 6px 12px;
    border-radius: var(--radius);
    cursor: pointer;
    font-size: 12px;
    font-family: inherit;
    transition: all 0.15s ease;
  }

  .toolbar-btn:hover:not(:disabled) {
    background: var(--bg-hover);
    border-color: var(--accent);
    color: var(--accent);
  }

  .toolbar-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .view-toggle {
    display: flex;
    gap: 2px;
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 2px;
  }

  .toggle-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    padding: 6px 10px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
    font-family: inherit;
  }

  .toggle-btn:hover {
    color: var(--text-primary);
  }

  .toggle-btn.active {
    background: var(--accent);
    color: #11111b;
  }

  .back-btn {
    background: none;
    border: 1px solid var(--border);
    color: var(--text-secondary);
    padding: 6px 14px;
    border-radius: var(--radius);
    cursor: pointer;
    font-size: 13px;
    margin-bottom: 16px;
    font-family: inherit;
  }

  .back-btn:hover {
    background: var(--bg-surface);
    color: var(--text-primary);
  }

  .header-info {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .title-row {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  h1 {
    font-size: 24px;
    font-weight: 700;
  }

  .meta-row {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;
  }

  .commander {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .commander strong {
    color: var(--accent);
  }

  .card-count {
    font-size: 13px;
    color: var(--green);
    font-weight: 600;
  }

  .card-count.warn {
    color: var(--yellow);
  }

  .total-price {
    font-size: 13px;
    color: var(--green);
    font-weight: 600;
  }

  .strategy {
    font-size: 13px;
    color: var(--text-secondary);
    line-height: 1.6;
    max-width: 800px;
  }

  h2 {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 12px;
    color: var(--text-secondary);
  }

  .content {
    display: flex;
    flex-direction: column;
    gap: 32px;
  }

  /* Search bar */
  .search-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 8px 14px;
    margin-bottom: 16px;
    transition: border-color 0.15s ease;
  }

  .search-bar:focus-within {
    border-color: var(--accent);
    box-shadow: 0 0 0 2px rgba(137, 180, 250, 0.15);
  }

  .search-icon {
    font-size: 14px;
    flex-shrink: 0;
  }

  .search-input {
    flex: 1;
    background: none;
    border: none;
    color: var(--text-primary);
    font-family: inherit;
    font-size: 14px;
    outline: none;
  }

  .search-input::placeholder {
    color: var(--text-muted);
  }

  .search-clear {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 14px;
    padding: 2px 6px;
    border-radius: 4px;
    flex-shrink: 0;
  }

  .search-clear:hover {
    color: var(--text-primary);
    background: var(--bg-hover);
  }

  .search-count {
    font-size: 12px;
    color: var(--text-muted);
    flex-shrink: 0;
    white-space: nowrap;
  }

  /* Not-found banner */
  .not-found-banner {
    background: rgba(243, 139, 168, 0.1);
    border: 1px solid rgba(243, 139, 168, 0.3);
    color: var(--red);
    padding: 10px 16px;
    border-radius: var(--radius);
    font-size: 13px;
    margin-bottom: 16px;
  }

  .cards-table {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
  }

  .table-header {
    display: flex;
    padding: 8px 16px;
    background: var(--bg-surface);
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-muted);
  }

  .card-row {
    display: flex;
    padding: 6px 16px;
    border-bottom: 1px solid var(--border);
    font-size: 13px;
    transition: background 0.1s;
    align-items: center;
  }

  .card-row:last-child {
    border-bottom: none;
  }

  .card-row:hover {
    background: var(--bg-hover);
  }

  .card-row.basic-land {
    color: var(--text-muted);
  }

  .card-row.is-commander {
    background: rgba(203, 166, 247, 0.06);
  }

  .card-row.is-commander:hover {
    background: rgba(203, 166, 247, 0.12);
  }

  .card-row.is-not-found {
    background: rgba(243, 139, 168, 0.06);
  }

  .card-row.is-not-found:hover {
    background: rgba(243, 139, 168, 0.12);
  }

  .card-row.is-not-found .col-name {
    color: var(--red);
    font-weight: 600;
  }

  .card-row.wishlist-row {
    color: var(--text-secondary);
  }

  .not-found-icon {
    font-size: 12px;
    margin-right: 4px;
  }

  .col-qty {
    width: 40px;
    flex-shrink: 0;
    color: var(--text-muted);
  }

  .col-name {
    flex: 1;
    min-width: 0;
  }

  .col-tags {
    width: 200px;
    flex-shrink: 0;
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
  }

  .col-set {
    width: 120px;
    flex-shrink: 0;
    text-align: right;
    color: var(--text-muted);
    font-size: 12px;
  }

  .card-badge {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 1px 6px;
    border-radius: 3px;
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }

  .card-badge-commander {
    background: rgba(203, 166, 247, 0.15);
    color: var(--mauve);
  }

  .card-badge-proxy {
    background: rgba(249, 226, 175, 0.15);
    color: var(--yellow);
  }

  .card-badge-wishlist {
    background: rgba(137, 180, 250, 0.15);
    color: var(--accent);
  }

  .foil-tag {
    font-size: 11px;
  }

  /* DFC flip button — inline in list view */
  .flip-btn {
    background: none;
    border: none;
    cursor: pointer;
    font-size: 12px;
    padding: 0 2px;
    opacity: 0.5;
    transition: opacity 0.15s ease;
    vertical-align: middle;
  }

  .flip-btn:hover {
    opacity: 1;
  }

  /* DFC flip button — overlay in grid view */
  .flip-btn-grid {
    position: absolute;
    bottom: 6px;
    right: 6px;
    background: rgba(0, 0, 0, 0.6);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 50%;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    font-size: 14px;
    opacity: 0;
    transition: opacity 0.15s ease;
  }

  .flip-btn-grid:hover {
    background: rgba(0, 0, 0, 0.8);
    border-color: var(--accent);
  }

  .proxy-price {
    color: var(--text-muted);
    font-style: italic;
    font-weight: 400;
  }

  /* Inline editing */
  .editing-row {
    padding: 4px 16px;
  }

  .edit-container {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .edit-row {
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .edit-input {
    flex: 1;
    background: var(--bg-surface);
    border: 1px solid var(--accent);
    border-radius: var(--radius);
    color: var(--text-primary);
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
    font-size: 13px;
    padding: 6px 10px;
    outline: none;
  }

  .edit-input:focus {
    box-shadow: 0 0 0 2px rgba(137, 180, 250, 0.25);
  }

  .edit-actions {
    display: flex;
    gap: 4px;
  }

  .edit-btn {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    color: var(--text-secondary);
    width: 28px;
    height: 28px;
    border-radius: var(--radius);
    cursor: pointer;
    font-size: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .edit-btn.save:hover {
    background: rgba(166, 227, 161, 0.15);
    border-color: var(--green);
    color: var(--green);
  }

  .edit-btn.cancel:hover {
    background: rgba(243, 139, 168, 0.15);
    border-color: var(--red);
    color: var(--red);
  }

  .edit-error {
    font-size: 11px;
    color: var(--red);
    padding: 2px 0;
  }

  .loading, .error {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-secondary);
  }

  .loading-spinner {
    width: 40px;
    height: 40px;
    border: 3px solid var(--border);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 16px;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .loading-hint {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 8px;
  }

  .error-container {
    text-align: center;
    padding: 60px 20px;
  }

  .error {
    color: var(--red);
    background: rgba(243, 139, 168, 0.1);
    border: 1px solid rgba(243, 139, 168, 0.3);
    padding: 16px;
    border-radius: var(--radius);
    margin-bottom: 16px;
  }

  .retry-btn {
    background: var(--accent);
    color: #11111b;
    border: none;
    padding: 10px 20px;
    border-radius: var(--radius);
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    font-family: inherit;
  }

  .retry-btn:hover {
    background: var(--accent-hover);
  }

  .wishlist h2 {
    color: var(--orange);
  }

  .col-price {
    width: 70px;
    flex-shrink: 0;
    text-align: right;
    color: var(--green);
    font-size: 12px;
    font-weight: 600;
  }

  /* Grid View */
  .card-grid h2 {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 16px;
    color: var(--text-secondary);
  }

  .cards-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 16px;
  }

  .grid-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .grid-card:hover {
    border-color: var(--accent);
    transform: translateY(-2px);
  }

  .grid-card.basic-land {
    opacity: 0.7;
  }

  .grid-card.is-commander {
    border-color: var(--mauve);
  }

  .grid-card.is-not-found {
    border-color: var(--red);
  }

  .card-image {
    aspect-ratio: 488 / 680;
    background: var(--bg-surface);
    overflow: hidden;
    position: relative;
  }

  .grid-card:hover .flip-btn-grid {
    opacity: 1;
  }

  .card-image img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .card-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
    font-weight: 700;
    color: var(--text-muted);
    background: linear-gradient(135deg, var(--bg-surface) 0%, var(--bg-secondary) 100%);
  }

  .card-placeholder.placeholder-not-found {
    background: linear-gradient(135deg, rgba(243, 139, 168, 0.1) 0%, rgba(243, 139, 168, 0.05) 100%);
  }

  .not-found-icon-large {
    font-size: 36px;
  }

  .card-details {
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .card-name {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .card-name-not-found {
    color: var(--red);
  }

  .not-found-icon-sm {
    font-size: 10px;
    margin-right: 2px;
  }

  .card-qty {
    font-size: 10px;
    color: var(--text-muted);
  }
</style>
