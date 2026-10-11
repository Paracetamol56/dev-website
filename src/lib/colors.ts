import type { AlphaColor, Color, Labels } from '@catppuccin/palette';

export type Palette = Labels<Color, AlphaColor>;

/**
 * A colour picked by the user: a Catppuccin palette key (follows the user's flavour),
 * a `#rrggbb` hex, or `transparent`.
 */
export type PaletteColor = string;

const HEX = /^#?([0-9a-f]{6})$/i;

export function parseHex(text: string): string | null {
	const match = text.trim().match(HEX);
	return match ? `#${match[1].toLowerCase()}` : null;
}

export function isPaletteKey(color: PaletteColor, palette: Palette): color is keyof Palette {
	return Object.hasOwn(palette, color);
}

export function resolveColor(color: PaletteColor, palette: Palette): string {
	if (color === 'transparent') return color;
	if (isPaletteKey(color, palette)) return palette[color].hex;
	return parseHex(color) ?? palette.text.hex;
}

export function randomPaletteKey(palette: Palette, except?: PaletteColor): keyof Palette {
	const keys = (Object.keys(palette) as (keyof Palette)[]).filter((key) => key !== except);
	return keys[Math.floor(Math.random() * keys.length)];
}

export const ACCENTS: (keyof Palette)[] = [
	'rosewater',
	'flamingo',
	'pink',
	'mauve',
	'red',
	'maroon',
	'peach',
	'yellow',
	'green',
	'teal',
	'sky',
	'sapphire',
	'blue',
	'lavender'
];

/** From the text colour down to the deepest background. */
export const NEUTRALS: (keyof Palette)[] = [
	'text',
	'subtext1',
	'subtext0',
	'overlay2',
	'overlay1',
	'overlay0',
	'surface2',
	'surface1',
	'surface0',
	'base',
	'mantle',
	'crust'
];

/** Hue of a `#rrggbb` colour, in degrees. */
export function hue(hex: string): number {
	const [r, g, b] = [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255);
	const max = Math.max(r, g, b);
	const delta = max - Math.min(r, g, b);
	if (delta === 0) return 0;
	const h =
		max === r ? ((g - b) / delta) % 6 : max === g ? (b - r) / delta + 2 : (r - g) / delta + 4;
	return (h * 60 + 360) % 360;
}

const rgb = (hex: string) => [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16));

/** The palette colour closest to a `#rrggbb` hex. */
export function nearestPaletteKey(hex: string, palette: Palette): keyof Palette {
	const [r, g, b] = rgb(hex);
	let nearest: keyof Palette = 'text';
	let best = Infinity;
	for (const key of Object.keys(palette) as (keyof Palette)[]) {
		const [pr, pg, pb] = rgb(palette[key].hex);
		const distance = (r - pr) ** 2 + (g - pg) ** 2 + (b - pb) ** 2;
		if (distance < best) {
			best = distance;
			nearest = key;
		}
	}
	return nearest;
}
