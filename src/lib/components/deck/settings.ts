export type DeckSettings = {
	theme: string;
	transition: string;
	transitionSpeed: string;
	backgroundTransition: string;
	controls: boolean;
	progress: boolean;
	slideNumber: boolean;
	center: boolean;
	loop: boolean;
	autoAnimate: boolean;
	autoAnimateEasing: string;
	autoAnimateDuration: number;
	width: number;
	height: number;
	markdown: boolean;
	highlight: boolean;
	highlightTheme: string;
};

export type Option = { value: string; label: string };

export const defaultSettings: DeckSettings = {
	theme: 'black',
	transition: 'slide',
	transitionSpeed: 'default',
	backgroundTransition: 'fade',
	controls: true,
	progress: true,
	slideNumber: false,
	center: true,
	loop: false,
	autoAnimate: true,
	autoAnimateEasing: 'ease',
	autoAnimateDuration: 1,
	width: 960,
	height: 700,
	markdown: true,
	highlight: true,
	highlightTheme: 'monokai'
};

export const themes: Option[] = [
	{ value: 'black', label: 'Black' },
	{ value: 'white', label: 'White' },
	{ value: 'black-contrast', label: 'Black, high contrast' },
	{ value: 'white-contrast', label: 'White, high contrast' },
	{ value: 'league', label: 'League' },
	{ value: 'beige', label: 'Beige' },
	{ value: 'night', label: 'Night' },
	{ value: 'serif', label: 'Serif' },
	{ value: 'simple', label: 'Simple' },
	{ value: 'solarized', label: 'Solarized' },
	{ value: 'moon', label: 'Moon' },
	{ value: 'dracula', label: 'Dracula' },
	{ value: 'sky', label: 'Sky' },
	{ value: 'blood', label: 'Blood' }
];

export const transitions: Option[] = [
	{ value: 'none', label: 'None' },
	{ value: 'fade', label: 'Fade' },
	{ value: 'slide', label: 'Slide' },
	{ value: 'convex', label: 'Convex' },
	{ value: 'concave', label: 'Concave' },
	{ value: 'zoom', label: 'Zoom' }
];

export const speeds: Option[] = [
	{ value: 'fast', label: 'Fast' },
	{ value: 'default', label: 'Normal' },
	{ value: 'slow', label: 'Slow' }
];

export const easings: Option[] = [
	{ value: 'ease', label: 'Ease' },
	{ value: 'ease-in', label: 'Ease in' },
	{ value: 'ease-out', label: 'Ease out' },
	{ value: 'ease-in-out', label: 'Ease in and out' },
	{ value: 'linear', label: 'Linear' }
];

export const sizes = [
	{ width: 960, height: 700, label: 'Standard (960 × 700)' },
	{ width: 1024, height: 768, label: '4:3 (1024 × 768)' },
	{ width: 1280, height: 720, label: '16:9 (1280 × 720)' },
	{ width: 1280, height: 800, label: '16:10 (1280 × 800)' }
];

export const highlightThemes: Option[] = [
	{ value: 'monokai', label: 'Monokai' },
	{ value: 'zenburn', label: 'Zenburn' }
];
