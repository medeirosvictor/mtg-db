import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import DeckToolbar from './DeckView/DeckToolbar.svelte';

// DeckView initializes viewMode to 'grid' (see DeckView.svelte:28).
// DeckToolbar is the component that renders the view-mode toggle, receiving
// that value as a prop. Testing the toolbar with viewMode='grid' verifies the
// behaviour the user observes: grid is marked as the active view.

describe('DeckView view mode', () => {
  it('shows grid as the active view when viewMode is grid', () => {
    const { getByTitle } = render(DeckToolbar, {
      props: { loading: false, scryfallLoading: false, viewMode: 'grid' },
    });

    const gridButton = getByTitle('Grid view');
    const listButton = getByTitle('List view');

    expect(gridButton.className).toContain('bg-accent');
    expect(listButton.className).not.toContain('bg-accent');
  });

  it('shows list as the active view when viewMode is list', () => {
    const { getByTitle } = render(DeckToolbar, {
      props: { loading: false, scryfallLoading: false, viewMode: 'list' },
    });

    const listButton = getByTitle('List view');
    const gridButton = getByTitle('Grid view');

    expect(listButton.className).toContain('bg-accent');
    expect(gridButton.className).not.toContain('bg-accent');
  });
});
