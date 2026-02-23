// Mirrors the Go types from app.go and internal/deck

export interface DeckSummary {
  slug: string;
  title: string;
  commander: string;
  colors: string;
  status: string;
  cardCount: number;
  universe?: string;
}

export interface Card {
  quantity: number;
  name: string;
  setCode?: string;
  collectorNumber?: string;
  foil?: boolean;
  tags?: string[];
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
