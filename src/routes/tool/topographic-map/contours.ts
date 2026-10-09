import { tickStep } from 'd3';
import { rasterPosition, type Raster } from './elevation';
import type { Line, MapFeatures } from './features';

/** Grids narrower than this are upsampled first, so overzoomed areas give smooth lines. */
const MIN_GRID_WIDTH = 800;

export type Contours = {
	/** Coordinate space of the lines: 0..width, 0..height */
	width: number;
	height: number;
	/** Metres between two contour lines */
	interval: number;
	lines: { elevation: number; points: number[] }[];
	/** Position of a point in the same coordinate space as the lines */
	project: (lat: number, lon: number) => [number, number];
};

export type ContourStyle = {
	/** Width of the output image in pixels; the height follows the region's shape */
	width: number;
	color: string;
	/** `null` for a transparent background */
	background: string | null;
	minorWidth: number;
	majorWidth: number;
	/** Every n-th contour line is a major (thicker) one */
	majorEvery: number;
	/** Lakes are filled, rivers and coastline stroked with `width` */
	water: { color: string; width: number } | null;
	/** `width` is for major roads; minor roads and paths are thinner, paths dashed */
	roads: { color: string; width: number } | null;
};

/** Contour lines every `interval` metres, or an interval giving about 25 lines with `'auto'`. */
export async function computeContours(
	raster: Raster,
	interval: number | 'auto'
): Promise<Contours> {
	const { default: mlcontour } = await import('maplibre-contour');

	let tile = mlcontour.HeightTile.fromRawDem(raster);
	const factor = Math.max(1, Math.min(4, Math.ceil(MIN_GRID_WIDTH / raster.width)));
	if (factor > 1) tile = tile.subsamplePixelCenters(factor);
	tile = tile.averagePixelCentersToGrid().materialize(1);

	if (interval === 'auto') {
		let min = Infinity;
		let max = -Infinity;
		for (const value of raster.data) {
			if (value < min) min = value;
			if (value > max) max = value;
		}
		interval = min < max ? Math.max(1, tickStep(min, max, 25)) : 10;
	}

	const width = tile.width - 1;
	const height = tile.height - 1;
	const isolines = mlcontour.generateIsolines(interval, tile, width, 0);
	const lines = Object.entries(isolines).flatMap(([elevation, segments]) =>
		// Lines of one or two points are noise at this scale
		segments
			.filter((points) => points.length >= 6)
			.map((points) => ({ elevation: Number(elevation), points }))
	);
	// Grid point x is at x / factor raster pixels (see averagePixelCentersToGrid)
	const project = (lat: number, lon: number) => {
		const [x, y] = rasterPosition(raster, lat, lon);
		return [x * factor, y * factor] as [number, number];
	};
	return { width, height, interval, lines, project };
}

const format = (n: number) => n.toFixed(1);

function linePath(points: number[]) {
	let d = `M${format(points[0])} ${format(points[1])}L`;
	for (let i = 2; i < points.length; i += 2) d += `${format(points[i])} ${format(points[i + 1])} `;
	return d.trimEnd();
}

export function renderSvg(contours: Contours, style: ContourStyle, features?: MapFeatures): string {
	const scale = style.width / contours.width;
	const height = Math.round(contours.height * scale);
	const majorInterval = contours.interval * style.majorEvery;
	const isMajor = (elevation: number) => style.majorEvery > 1 && elevation % majorInterval === 0;
	// Widths are given in output pixels, the drawing is in grid units
	const stroke = (width: number) => (width / scale).toFixed(3);

	const projected = (line: Line) => line.flatMap(([lon, lat]) => contours.project(lat, lon));
	const paths = (lines: Line[]) =>
		lines
			.filter((l) => l.length > 1)
			.map((l) => linePath(projected(l)))
			.join('');
	const contourPaths = (lines: Contours['lines']) => lines.map((l) => linePath(l.points)).join('');

	const minor = contours.lines.filter((line) => !isMajor(line.elevation));
	const major = contours.lines.filter((line) => isMajor(line.elevation));
	const parts = [
		`<svg xmlns="http://www.w3.org/2000/svg" width="${style.width}" height="${height}" viewBox="0 0 ${contours.width} ${contours.height}">`,
		style.background ? `<rect width="100%" height="100%" fill="${style.background}"/>` : ''
	];

	if (style.water && features) {
		const { color, width } = style.water;
		const areas = features.water
			.map((rings) =>
				rings
					.filter((r) => r.length > 2)
					.map((r) => `${linePath(projected(r))}Z`)
					.join('')
			)
			.join('');
		const lines = paths([...features.waterways, ...features.coastline]);
		if (areas)
			parts.push(
				`<path fill="${color}" fill-opacity="0.4" stroke="${color}" stroke-width="${stroke(
					width / 2
				)}" fill-rule="evenodd" d="${areas}"/>`
			);
		if (lines) {
			parts.push(
				`<path fill="none" stroke="${color}" stroke-width="${stroke(
					width
				)}" stroke-linecap="round" stroke-linejoin="round" d="${lines}"/>`
			);
		}
	}

	parts.push(
		`<g fill="none" stroke="${style.color}" stroke-linecap="round" stroke-linejoin="round">`
	);
	if (minor.length)
		parts.push(`<path stroke-width="${stroke(style.minorWidth)}" d="${contourPaths(minor)}"/>`);
	if (major.length)
		parts.push(`<path stroke-width="${stroke(style.majorWidth)}" d="${contourPaths(major)}"/>`);
	parts.push('</g>');

	if (style.roads && features) {
		const { color, width } = style.roads;
		const byClass = (roadClass: string) =>
			paths(features.roads.filter((road) => road.class === roadClass).map((road) => road.line));
		const layers: [string, number, string][] = [
			['path', width * 0.4, ` stroke-dasharray="${stroke(width * 1.5)} ${stroke(width * 1.2)}"`],
			['minor', width * 0.6, ''],
			['major', width, '']
		];
		parts.push(`<g fill="none" stroke="${color}" stroke-linecap="round" stroke-linejoin="round">`);
		for (const [roadClass, roadWidth, extra] of layers) {
			const d = byClass(roadClass);
			if (d) parts.push(`<path stroke-width="${stroke(roadWidth)}"${extra} d="${d}"/>`);
		}
		parts.push('</g>');
	}

	parts.push('</svg>');
	return parts.join('');
}
