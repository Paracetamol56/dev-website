import { resolveColor, type Palette, type PaletteColor } from '$lib/colors';
import type { RoadDetail } from './features';

/** The export frame's corners as [lon, lat]: top-left, top-right, bottom-right, bottom-left */
export type FrameCorners = [[number, number], [number, number], [number, number], [number, number]];

/** A colour of the gradient used for the altitude background, at an elevation in metres */
export type GradientStop = { elevation: number; color: PaletteColor };

/**
 * Everything that defines the output image; the map previews it and the export renders it.
 * Colours are palette keys (following the user's flavour) or hex; see `resolveStyle`.
 */
export type OutputStyle = {
	/** Metres between contour lines, or `'auto'` to follow the map zoom */
	interval: number | 'auto';
	/** Every n-th contour line is a major (thicker) one; 1 for none */
	majorEvery: number;
	color: PaletteColor;
	minorWidth: number;
	majorWidth: number;
	/** Plain colour, colours by altitude (hypsometric tints), or nothing */
	backgroundMode: 'color' | 'elevation' | 'transparent';
	background: PaletteColor;
	/** Colours by altitude for the 'elevation' background and the relief view */
	gradient: GradientStop[];
	water: boolean;
	waterColor: PaletteColor;
	/** Width of rivers and streams; lakes and sea are filled */
	waterWidth: number;
	roads: boolean;
	roadDetail: RoadDetail;
	roadColor: PaletteColor;
	/** Width of major roads; local roads and paths are thinner, paths dashed */
	roadWidth: number;
	buildings: boolean;
	buildingColor: PaletteColor;
	/** Width of the exported image in pixels; line widths are relative to it */
	width: number;
};

/**
 * Default hypsometric tints, from the Catppuccin palette so they follow the user's flavour.
 * The tiles store the sea surface as 0 m, so two stops at 0 m give a sharp coastline: water at
 * and below sea level, land just above.
 */
export const DEFAULT_GRADIENT: GradientStop[] = [
	{ elevation: -500, color: 'blue' },
	{ elevation: 0, color: 'sky' },
	{ elevation: 0, color: 'green' },
	{ elevation: 600, color: 'yellow' },
	{ elevation: 1500, color: 'peach' },
	{ elevation: 2500, color: 'maroon' },
	{ elevation: 3500, color: 'flamingo' },
	{ elevation: 4500, color: 'rosewater' }
];

export const DEFAULT_STYLE: OutputStyle = {
	interval: 'auto',
	majorEvery: 5,
	color: 'text',
	minorWidth: 1,
	majorWidth: 2.5,
	backgroundMode: 'color',
	background: 'base',
	gradient: DEFAULT_GRADIENT,
	water: false,
	waterColor: 'sapphire',
	waterWidth: 2,
	roads: false,
	roadDetail: 'minor',
	roadColor: 'red',
	roadWidth: 3,
	buildings: false,
	buildingColor: 'overlay0',
	width: 2000
};

/** The style with every colour as hex, for drawing. */
export function resolveStyle(style: OutputStyle, palette: Palette): OutputStyle {
	const hex = (color: PaletteColor) => resolveColor(color, palette);
	return {
		...style,
		color: hex(style.color),
		background: hex(style.background),
		gradient: style.gradient.map((stop) => ({ ...stop, color: hex(stop.color) })),
		waterColor: hex(style.waterColor),
		roadColor: hex(style.roadColor),
		buildingColor: hex(style.buildingColor)
	};
}

/**
 * Gradient stops sorted by elevation, as [elevation, colour]. Stops at the same elevation make
 * a sharp edge; they are nudged apart because colour ramps need strictly increasing stops.
 */
export function gradientStops(gradient: GradientStop[]): [number, string][] {
	const sorted = [...gradient].sort((a, b) => a.elevation - b.elevation);
	const stops: [number, string][] = [];
	for (const { elevation, color } of sorted) {
		const previous = stops[stops.length - 1]?.[0] ?? -Infinity;
		stops.push([Math.max(elevation, previous + 0.01), color]);
	}
	return stops;
}

/** Relative widths and dash pattern (in multiples of the line width) of each road class */
export const ROAD_STYLE = {
	major: { width: 1, dash: null },
	minor: { width: 0.6, dash: null },
	path: { width: 0.4, dash: [3.75, 3] }
} as const;

// With 'auto', the spacing follows the map zoom: [minor, major] metres from that zoom onwards
const AUTO_CONTOURS: Record<number, [number, number]> = {
	0: [500, 2000],
	9: [200, 1000],
	11: [100, 500],
	12: [50, 250],
	13: [20, 100],
	15: [10, 50]
};

/** maplibre-contour thresholds: zoom → [minor, major] (or [minor] without major lines) */
export function contourThresholds(style: OutputStyle): Record<number, number[]> {
	if (style.interval === 'auto') return AUTO_CONTOURS;
	const major = style.majorEvery > 1 ? [style.interval * style.majorEvery] : [];
	return { 0: [style.interval, ...major] };
}

/** The [minor interval, majorEvery] actually drawn at a map zoom. */
export function contourSpacing(style: OutputStyle, zoom: number): [number, number] {
	if (style.interval !== 'auto') return [style.interval, style.majorEvery];
	const key = Math.max(
		...Object.keys(AUTO_CONTOURS)
			.map(Number)
			.filter((z) => z <= Math.floor(zoom))
	);
	const [minor, major] = AUTO_CONTOURS[key];
	return [minor, major / minor];
}
