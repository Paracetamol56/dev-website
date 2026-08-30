import type { SelectOption } from '@melt-ui/svelte';

export type ColorMode = 'catppuccin' | 'random' | 'custom' | 'transparent';
export type ShapeOption = SelectOption<'square' | 'round'>;
export type FormatOption = SelectOption<'png' | 'jpg' | 'webp' | 'svg' | 'bmp' | 'ico'>;
