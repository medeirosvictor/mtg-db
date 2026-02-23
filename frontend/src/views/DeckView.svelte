<script lang="ts">
  import { onMount } from 'svelte';
  import type { Deck, Card } from '../lib/types';
  import { GetDeck, ToggleCardTag } from '../../wailsjs/go/main/App';
  import ColorPips from '../components/ColorPips.svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import ContextMenu from '../components/ContextMenu.svelte';
  import { createEventDispatcher } from 'svelte';

  export let slug: string;

  const dispatch = createEventDispatcher();

  let deck: Deck | null = null;
  let loading = true;
  let error = '';

  // Context menu state
  let menuVisible = false;
  let menuX = 0;
  let menuY = 0;
  let menuItems: any[] = [];
  let menuContext: 'deck' | 'wishlist' = 'deck';

  onMount(async () => {
    await loadDeck();
  });

  async function loadDeck() {
    try {
      loading = true;
      const result = await GetDeck(slug);
      deck = result;
      if (!deck) {
        error = `Deck "${slug}" not found`;
      }
    } catch (e) {
      error = `Failed to load deck: ${e}`;
    } finally {
      loading = false;
    }
  }

  function isBasicLand(name: string): boolean {
    const basics = ['plains', 'island', 'swamp', 'mountain', 'forest'];
    return basics.includes(name.toLowerCase());
  }

  $: sortedCards = deck?.cards
    ? [...deck.cards].sort((a, b) => {
        // Commanders first
        const aCmd = (a.tags || []).includes('commander');
        const bCmd = (b.tags || []).includes('commander');
        if (aCmd !== bCmd) return aCmd ? -1 : 1;
        // Then non-basics, then basics
        const aBasic = isBasicLand(a.name);
        const bBasic = isBasicLand(b.name);
        if (aBasic !== bBasic) return aBasic ? 1 : -1;
        return a.name.localeCompare(b.name);
      })
    : [];

  $: wishlistCards = deck?.wishlist || [];
  $: commanders = deck?.cards?.filter(c => (c.tags || []).includes('commander')) || [];
  $: commanderNames = commanders.map(c => c.name).join(' / ');
  $: displayCommander = commanderNames || deck?.info?.commander || 'None set';

  async function toggleTag(cardName: string, tag: string) {
    if (!deck) return;
    const updated = await ToggleCardTag(slug, cardName, tag);
    if (updated) {
      deck = updated;
    }
  }

  function showCardContextMenu(e: MouseEvent, card: Card, context: 'deck' | 'wishlist') {
    e.preventDefault();
    menuX = e.clientX;
    menuY = e.clientY;
    menuContext = context;

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

<div class="deck-view">
  {#if loading}
    <div class="loading">Loading deck...</div>
  {:else if error}
    <div class="error">{error}</div>
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
        </div>
        {#if deck.info.strategy}
          <p class="strategy">{deck.info.strategy}</p>
        {/if}
      </div>
    </header>

    <div class="content">
      <section class="card-list">
        <h2>Cards ({sortedCards.length} unique, {deck.cardCount} total)</h2>
        <div class="cards-table">
          <div class="table-header">
            <span class="col-qty">#</span>
            <span class="col-name">Card Name</span>
            <span class="col-tags">Tags</span>
            <span class="col-set">Set</span>
          </div>
          {#each sortedCards as card (card.name)}
            <div
              class="card-row"
              class:basic-land={isBasicLand(card.name)}
              class:is-commander={(card.tags || []).includes('commander')}
              on:contextmenu={(e) => showCardContextMenu(e, card, 'deck')}
            >
              <span class="col-qty">{card.quantity}×</span>
              <span class="col-name">
                {card.name}
                {#if card.foil}
                  <span class="foil-tag">✨</span>
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
              <span class="col-set">
                {#if card.setCode}
                  {card.setCode}
                  {#if card.collectorNumber}
                    #{card.collectorNumber}
                  {/if}
                {/if}
              </span>
            </div>
          {/each}
        </div>
      </section>

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
                  {#if card.setCode}
                    {card.setCode}
                  {/if}
                </span>
              </div>
            {/each}
          </div>
        </section>
      {/if}
    </div>
  {/if}
</div>

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

  .card-row.wishlist-row {
    color: var(--text-secondary);
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

  .loading, .error {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-secondary);
  }

  .error {
    color: var(--red);
  }

  .wishlist h2 {
    color: var(--orange);
  }
</style>
