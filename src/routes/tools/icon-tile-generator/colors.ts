import { ACCENTS, type Palette, type PaletteColor } from '$lib/colors';

// Neutrals that contrast well with accents
const RANDOM_NEUTRALS: (keyof Palette)[] = ['text', 'surface0', 'base', 'mantle', 'crust'];

const pick = <T>(items: T[]) => items[Math.floor(Math.random() * items.length)];

/** A readable background/icon pair: one accent colour on one neutral colour. */
export function randomColorPair(transparentBg: boolean): { bg: PaletteColor; fg: PaletteColor } {
	if (transparentBg) return { bg: 'transparent', fg: pick(ACCENTS) };
	return Math.random() < 0.5
		? { bg: pick(RANDOM_NEUTRALS), fg: pick(ACCENTS) }
		: { bg: pick(ACCENTS), fg: pick(RANDOM_NEUTRALS) };
}
