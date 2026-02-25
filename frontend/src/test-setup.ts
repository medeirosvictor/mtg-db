import '@testing-library/svelte';
import { vi } from 'vitest';

// Mock Wails runtime
vi.mock('../../wailsjs/go/app/App', () => ({
  GetAllDecks: vi.fn().mockResolvedValue([]),
  GetDeck: vi.fn().mockResolvedValue(null),
  GetDeckBasic: vi.fn().mockResolvedValue(null),
  ToggleCardTag: vi.fn().mockResolvedValue(null),
  UpdateCardText: vi.fn().mockResolvedValue(''),
  UpdateDeckInfo: vi.fn().mockResolvedValue(''),
  UpdateDeckStatus: vi.fn().mockResolvedValue(''),
  GetAppState: vi.fn().mockResolvedValue({
    hasCollection: true,
    collectionPath: '/test/path',
    collectionLabel: 'Test Collection',
    collectionValid: true,
    needsSetup: false,
    collections: [],
  }),
  SelectCollectionFolder: vi.fn().mockResolvedValue(''),
  SwitchCollection: vi.fn().mockResolvedValue(''),
  ImportDeck: vi.fn().mockResolvedValue({ cards: [], error: '' }),
  ExportDeck: vi.fn().mockResolvedValue(''),
  CreateDeckFromImport: vi.fn().mockResolvedValue('new-deck'),
}));

// Mock scrollIntoView
window.scrollIntoView = vi.fn();
