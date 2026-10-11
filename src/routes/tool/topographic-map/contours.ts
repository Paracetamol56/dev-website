import { rasterPosition, type Raster } from './elevation';
import type { Line, MapFeatures } from './features';
import { ROAD_STYLE, gradientStops, type FrameCorners, type OutputStyle } from './style';

/** Grids narrower than this are upsampled first, so overzoomed areas give smooth lines. */
const MIN_GRID_WIDTH = 800;
/**
 * Real resolution of the elevation data behind the tiles, in metres. Finer tiles are
 * interpolated from it and come out as small terraces, which trace as staircase lines.
 */
const SOURCE_RESOLUTION = 40;
/** Largest deviation, in grid units, allowed when dropping nearly straight points */
const SIMPLIFY_TOLERANCE = 0.25;
/** Tighter tolerance for the final pass, which only removes points left in line by smoothing */
const FINAL_TOLERANCE = 0.08;
const SMOOTHING_PASSES = 2;

export type Contours = {
	/** Coordinate space of the lines: 0..width, 0..height */
	width: number;
	height: number;
	/** Metres between two contour lines */
	interval: number;
	lines: { elevation: number; points: number[] }[];
	/** Position of a point in the same coordinate space as the lines */
	project: (lat: number, lon: number) => [number, number];
	/** The elevations the lines were traced from; one raster pixel spans `factor` units */
	raster: Raster;
	factor: number;
};

/** Contour lines every `interval` metres. */
export async function computeContours(raster: Raster, interval: number): Promise<Contours> {
	const { default: mlcontour } = await import('maplibre-contour');

	// Contours are traced on a slightly blurred copy, which removes the terraces; the raster
	// itself stays untouched for the altitude background
	let tile = mlcontour.HeightTile.fromRawDem({
		...raster,
		data: blur(raster.data, raster.width, raster.height, blurRadius(raster))
	});
	const factor = Math.max(1, Math.min(4, Math.ceil(MIN_GRID_WIDTH / raster.width)));
	if (factor > 1) tile = tile.subsamplePixelCenters(factor);
	tile = tile.averagePixelCentersToGrid().materialize(1);

	const width = tile.width - 1;
	const height = tile.height - 1;
	const isolines = mlcontour.generateIsolines(interval, tile, width, 0);
	const lines = Object.entries(isolines).flatMap(([elevation, segments]) =>
		// Lines of one or two points are noise at this scale
		segments
			.filter((points) => points.length >= 6)
			.map((points) => ({
				elevation: Number(elevation),
				points: simplify(smooth(simplify(points)), FINAL_TOLERANCE)
			}))
	);
	// Grid point x is at x / factor raster pixels (see averagePixelCentersToGrid)
	const project = (lat: number, lon: number) => {
		const [x, y] = rasterPosition(raster, lat, lon);
		return [x * factor, y * factor] as [number, number];
	};
	return { width, height, interval, lines, project, raster, factor };
}

/** Gaussian blur radius (standard deviation, in raster pixels) matching the data's real resolution. */
function blurRadius(raster: Raster): number {
	const [, south, , north] = raster.bounds;
	const latitude = ((south + north) / 2) * (Math.PI / 180);
	const metresPerPixel = (40075016.686 * Math.cos(latitude)) / (256 * 2 ** raster.zoom);
	return Math.min(3, Math.max(0.5, SOURCE_RESOLUTION / metresPerPixel));
}

/** Separable Gaussian blur; missing values (NaN) are left out instead of spreading. */
function blur(data: Float32Array, width: number, height: number, sigma: number): Float32Array {
	const radius = Math.ceil(sigma * 2.5);
	const kernel = Array.from({ length: radius * 2 + 1 }, (_, i) =>
		Math.exp(-((i - radius) ** 2) / (2 * sigma * sigma))
	);
	const pass = (input: Float32Array, horizontal: boolean) => {
		const output = new Float32Array(input.length);
		for (let y = 0; y < height; y++) {
			for (let x = 0; x < width; x++) {
				let sum = 0;
				let weights = 0;
				for (let k = -radius; k <= radius; k++) {
					const sx = horizontal ? x + k : x;
					const sy = horizontal ? y : y + k;
					if (sx < 0 || sx >= width || sy < 0 || sy >= height) continue;
					const value = input[sy * width + sx];
					if (Number.isNaN(value)) continue;
					sum += value * kernel[k + radius];
					weights += kernel[k + radius];
				}
				output[y * width + x] =
					Number.isNaN(input[y * width + x]) || !weights ? NaN : sum / weights;
			}
		}
		return output;
	};
	return pass(pass(data, true), false);
}

/** Douglas–Peucker: drops points within `tolerance` of the line through their neighbours. */
function simplify(points: number[], tolerance = SIMPLIFY_TOLERANCE): number[] {
	const count = points.length / 2;
	if (count < 3) return points;
	const keep = new Uint8Array(count);
	keep[0] = keep[count - 1] = 1;
	const stack: [number, number][] = [[0, count - 1]];
	while (stack.length) {
		const [first, last] = stack.pop()!;
		const [ax, ay, bx, by] = [
			points[first * 2],
			points[first * 2 + 1],
			points[last * 2],
			points[last * 2 + 1]
		];
		const length = Math.hypot(bx - ax, by - ay);
		let farthest = -1;
		let distance = tolerance;
		for (let i = first + 1; i < last; i++) {
			const [px, py] = [points[i * 2], points[i * 2 + 1]];
			const d = length
				? Math.abs((bx - ax) * (ay - py) - (ax - px) * (by - ay)) / length
				: Math.hypot(px - ax, py - ay);
			if (d > distance) {
				distance = d;
				farthest = i;
			}
		}
		if (farthest >= 0) {
			keep[farthest] = 1;
			stack.push([first, farthest], [farthest, last]);
		}
	}
	return points.filter((_, i) => keep[i >> 1]);
}

/** Chaikin corner cutting: rounds corners; open lines keep their ends, closed rings stay closed. */
function smooth(points: number[]): number[] {
	let current = points;
	for (let pass = 0; pass < SMOOTHING_PASSES; pass++) {
		const count = current.length / 2;
		if (count < 3) break;
		const closed =
			current[0] === current[current.length - 2] && current[1] === current[current.length - 1];
		const next: number[] = closed ? [] : [current[0], current[1]];
		for (let i = 0; i < count - 1; i++) {
			const [ax, ay, bx, by] = current.slice(i * 2, i * 2 + 4);
			next.push(
				0.75 * ax + 0.25 * bx,
				0.75 * ay + 0.25 * by,
				0.25 * ax + 0.75 * bx,
				0.25 * ay + 0.75 * by
			);
		}
		if (closed) next.push(next[0], next[1]);
		else next.push(current[current.length - 2], current[current.length - 1]);
		current = next;
	}
	return current;
}

const format = (n: number) => n.toFixed(1);

/** The raster coloured by altitude with the gradient, as a PNG data URL (no data stays transparent). */
function elevationImage(raster: Raster, gradient: OutputStyle['gradient']): string {
	const stops = gradientStops(gradient).map(
		([elevation, hex]) =>
			[elevation, [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16))] as const
	);
	const canvas = document.createElement('canvas');
	canvas.width = raster.width;
	canvas.height = raster.height;
	const context = canvas.getContext('2d')!;
	const image = context.createImageData(raster.width, raster.height);
	raster.data.forEach((elevation, i) => {
		if (Number.isNaN(elevation)) return;
		if (stops.length === 1) {
			image.data.set([...stops[0][1], 255], i * 4);
			return;
		}
		let stop = 1;
		while (stop < stops.length - 1 && elevation > stops[stop][0]) stop++;
		const [low, lowColor] = stops[stop - 1];
		const [high, highColor] = stops[stop];
		const t = Math.min(1, Math.max(0, (elevation - low) / (high - low)));
		for (let c = 0; c < 3; c++) {
			image.data[i * 4 + c] = Math.round(lowColor[c] + (highColor[c] - lowColor[c]) * t);
		}
		image.data[i * 4 + 3] = 255;
	});
	context.putImageData(image, 0, 0);
	return canvas.toDataURL('image/png');
}

function linePath(points: number[]) {
	let d = `M${format(points[0])} ${format(points[1])}L`;
	for (let i = 2; i < points.length; i += 2) d += `${format(points[i])} ${format(points[i + 1])} `;
	return d.trimEnd();
}

/**
 * The output image of the frame: water, buildings, contour lines, then roads, as in the map preview.
 * The drawing is turned so the frame's top edge is horizontal, then cropped to the frame.
 */
export function renderSvg(
	contours: Contours,
	style: OutputStyle,
	majorEvery: number,
	frame: FrameCorners,
	features?: MapFeatures
): string {
	// The map rotates in this same projected space, so the frame is an exact rectangle here
	const [topLeft, topRight, , bottomLeft] = frame.map(([lon, lat]) => contours.project(lat, lon));
	const frameWidth = Math.hypot(topRight[0] - topLeft[0], topRight[1] - topLeft[1]);
	const frameHeight = Math.hypot(bottomLeft[0] - topLeft[0], bottomLeft[1] - topLeft[1]);
	const angle = (Math.atan2(topRight[1] - topLeft[1], topRight[0] - topLeft[0]) * 180) / Math.PI;

	const scale = style.width / frameWidth;
	const height = Math.round(frameHeight * scale);
	const majorInterval = contours.interval * majorEvery;
	const isMajor = (elevation: number) => majorEvery > 1 && elevation % majorInterval === 0;
	// Widths are given in output pixels, the drawing is in grid units
	const stroke = (width: number) => (Math.max(0, width) / scale).toFixed(3);

	const projected = (line: Line) => line.flatMap(([lon, lat]) => contours.project(lat, lon));
	const paths = (lines: Line[]) =>
		lines
			.filter((l) => l.length > 1)
			.map((l) => linePath(projected(l)))
			.join('');
	const contourPaths = (lines: Contours['lines']) => lines.map((l) => linePath(l.points)).join('');

	const parts = [
		`<svg xmlns="http://www.w3.org/2000/svg" width="${
			style.width
		}" height="${height}" viewBox="0 0 ${format(frameWidth)} ${format(frameHeight)}">`,
		style.backgroundMode === 'color'
			? `<rect width="100%" height="100%" fill="${style.background}"/>`
			: '',
		`<g transform="rotate(${(-angle).toFixed(3)}) translate(${format(-topLeft[0])} ${format(
			-topLeft[1]
		)})">`
	];

	if (style.backgroundMode === 'elevation') {
		// Raster pixel i spans grid units [i, i + 1) × factor, starting at the grid origin
		const { raster, factor } = contours;
		parts.push(
			`<image href="${elevationImage(raster, style.gradient)}" width="${
				raster.width * factor
			}" height="${
				raster.height * factor
			}" preserveAspectRatio="none" image-rendering="optimizeQuality"/>`
		);
	}

	if (style.water && features) {
		// One path per polygon: even-odd keeps its islands as holes, while the copies of an area
		// in neighbouring tiles overlap without cancelling out. Opaque and without outline, so the
		// pieces join seamlessly.
		const areas = features.water
			.map((rings) =>
				rings
					.filter((r) => r.length > 2)
					.map((r) => `${linePath(projected(r))}Z`)
					.join('')
			)
			.filter(Boolean);
		const rivers = paths(features.waterways);
		if (areas.length) {
			parts.push(`<g fill="${style.waterColor}" fill-rule="evenodd">`);
			for (const d of areas) parts.push(`<path d="${d}"/>`);
			parts.push('</g>');
		}
		if (rivers) {
			parts.push(
				`<path fill="none" stroke="${style.waterColor}" stroke-width="${stroke(
					style.waterWidth
				)}" stroke-linecap="round" stroke-linejoin="round" d="${rivers}"/>`
			);
		}
	}

	if (style.buildings && features?.buildings.length) {
		parts.push(`<g fill="${style.buildingColor}">`);
		for (const rings of features.buildings) {
			const d = rings
				.filter((r) => r.length > 2)
				.map((r) => `${linePath(projected(r))}Z`)
				.join('');
			if (d) parts.push(`<path d="${d}"/>`);
		}
		parts.push('</g>');
	}

	const minor = contours.lines.filter((line) => !isMajor(line.elevation));
	const major = contours.lines.filter((line) => isMajor(line.elevation));
	parts.push(
		`<g fill="none" stroke="${style.color}" stroke-linecap="round" stroke-linejoin="round">`
	);
	if (minor.length)
		parts.push(`<path stroke-width="${stroke(style.minorWidth)}" d="${contourPaths(minor)}"/>`);
	if (major.length)
		parts.push(`<path stroke-width="${stroke(style.majorWidth)}" d="${contourPaths(major)}"/>`);
	parts.push('</g>');

	if (style.roads && features) {
		parts.push(
			`<g fill="none" stroke="${style.roadColor}" stroke-linecap="round" stroke-linejoin="round">`
		);
		for (const roadClass of ['path', 'minor', 'major'] as const) {
			const d = paths(features.roads.filter((road) => road.class === roadClass).map((r) => r.line));
			if (!d) continue;
			const { width, dash } = ROAD_STYLE[roadClass];
			const lineWidth = style.roadWidth * width;
			const dasharray = dash
				? ` stroke-dasharray="${dash.map((d) => stroke(d * lineWidth)).join(' ')}"`
				: '';
			parts.push(`<path stroke-width="${stroke(lineWidth)}"${dasharray} d="${d}"/>`);
		}
		parts.push('</g>');
	}

	parts.push('</g></svg>');
	return parts.join('');
}
