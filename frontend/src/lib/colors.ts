// Parse the color emoji string from info.md into individual MTG colors

export type MtgColor = 'W' | 'U' | 'B' | 'R' | 'G';

const emojiToColor: Record<string, MtgColor> = {
  '⚪': 'W',
  '🔵': 'U',
  '⚫': 'B',
  '🔴': 'R',
  '🟢': 'G',
};

const colorHex: Record<MtgColor, string> = {
  W: '#F9FAF4',
  U: '#0E68AB',
  B: '#150B00',
  R: '#D3202A',
  G: '#00733E',
};

export function parseColors(colorString: string): MtgColor[] {
  const colors: MtgColor[] = [];
  for (const [emoji, color] of Object.entries(emojiToColor)) {
    if (colorString.includes(emoji)) {
      colors.push(color);
    }
  }
  return colors;
}

export function getColorHex(color: MtgColor): string {
  return colorHex[color];
}
