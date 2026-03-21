// Mirrors the Go types from app.go and internal/deck

export interface DeckSummary {
  slug: string;
  title: string;
  commander: string;
  commanderImageUri?: string;
  colors: string;
  status: string;
  cardCount: number;
  universe?: string;
}

export interface ScryFallData {
  oracleText?: string;
  typeLine?: string;
  manaCost?: string;
  cmc?: number;
  imageUri?: string;
  backImageUri?: string;
  isDoubleFaced?: boolean;
  priceUsd?: string;
  priceUsdFoil?: string;
  colorIdentity?: string;
}

export interface Card {
  quantity: number;
  name: string;
  setCode?: string;
  collectorNumber?: string;
  foil?: boolean;
  tags?: string[];
  scryFall?: ScryFallData;
}

export interface DeckInfo {
  title: string;
  status: string;
  colors: string;
  commander: string;
  strategy: string;
  universe?: string;
}

export interface Deck {
  slug: string;
  info: DeckInfo;
  cards: Card[];
  wishlist: Card[];
  cardCount: number;
}

export interface SyncResult {
  deck: Deck;
  notFound: string[] | null;
}

export interface ImportedCard {
  quantity: number;
  name: string;
}

export interface ImportResult {
  cards: ImportedCard[];
  deckName: string;
  source: string;
  error?: string;
}

export interface CollectionInfo {
  path: string;
  label: string;
  lastOpened: string;
  isActive: boolean;
  isValid: boolean;
}

export interface AppState {
  hasCollection: boolean;
  collectionPath: string;
  collectionLabel: string;
  collectionValid: boolean;
  collections: CollectionInfo[] | null;
  needsSetup: boolean;
}
