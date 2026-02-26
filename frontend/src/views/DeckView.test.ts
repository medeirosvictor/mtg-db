import { describe, it, expect, vi } from 'vitest';

// =====================
// DeckView Default View Mode Test
// =====================
// Verify that the DeckView component defaults to grid view by checking the source.
// This is a simple static analysis test — no DOM rendering needed.

describe('DeckView defaults', () => {
  it('should default to grid view mode in source', async () => {
    // Read the actual source file to verify the default
    const fs = await import('fs');
    const path = await import('path');
    const source = fs.readFileSync(
      path.resolve(__dirname, 'DeckView.svelte'),
      'utf-8'
    );
    // Must contain the grid default, not list
    expect(source).toContain("let viewMode: 'list' | 'grid' = 'grid'");
    expect(source).not.toContain("let viewMode: 'list' | 'grid' = 'list'");
  });
});
