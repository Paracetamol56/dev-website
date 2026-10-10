import type mlcontour from 'maplibre-contour';

// Place search uses OpenStreetMap's Nominatim. OpenStreetMap has no elevation service, so
// heights come from the AWS Terrain Tiles: web-mercator PNG tiles in the "terrarium" encoding,
// where each pixel stores one elevation.

export const DEM_URL = 'https://s3.amazonaws.com/elevation-tiles-prod/terrarium/{z}/{x}/{y}.png';
export const DEM_MAX_ZOOM = 14;

const TILE_SIZE = 256;

export type DemSource = InstanceType<typeof mlcontour.DemSource>;

/** [west, south, east, north] in degrees */
export type Bounds = [number, number, number, number];

/** Elevations in metres, row-major from the north-west corner. */
export type Raster = {
	width: number;
	height: number;
	data: Float32Array;
	bounds: Bounds;
	/** Web-mercator zoom of the grid, and its top-left corner in global pixels at that zoom */
	zoom: number;
	left: number;
	top: number;
};

let demSource: Promise<DemSource> | null = null;

/** The DEM tile source shared by the map and the export, so each tile is fetched once. */
export function loadDem(): Promise<DemSource> {
	demSource ??= import('maplibre-contour').then(
		({ default: mlcontour }) =>
			new mlcontour.DemSource({
				url: DEM_URL,
				encoding: 'terrarium',
				maxzoom: DEM_MAX_ZOOM,
				worker: true,
				// Contour tiles need several elevation tiles each; give slow connections time
				timeoutMs: 30000
			})
	);
	return demSource;
}

/** Position of a point in the raster's pixel grid (0, 0 is the grid's top-left corner). */
export function rasterPosition(raster: Raster, lat: number, lon: number): [number, number] {
	const { x, y } = project(lat, lon, raster.zoom);
	return [x * TILE_SIZE - raster.left, y * TILE_SIZE - raster.top];
}

/** Web-mercator position in tiles at the given zoom. */
function project(lat: number, lon: number, zoom: number) {
	const tiles = 2 ** zoom;
	const sin = Math.sin((lat * Math.PI) / 180);
	return {
		x: ((lon + 180) / 360) * tiles,
		y: (0.5 - Math.log((1 + sin) / (1 - sin)) / (4 * Math.PI)) * tiles
	};
}

export type Place = {
	name: string;
	lat: number;
	lon: number;
	/** [west, south, east, north], when the place has an extent (a city, a park...) */
	bounds?: [number, number, number, number];
};

const COORDINATES = /^\s*(-?\d+(?:\.\d+)?)\s*[,\s]\s*(-?\d+(?:\.\d+)?)\s*$/;

/** Resolves "lat, lon" directly, or searches the place name with Nominatim. */
export async function findPlace(query: string, signal?: AbortSignal): Promise<Place> {
	const match = query.match(COORDINATES);
	if (match) {
		const lat = Number(match[1]);
		const lon = Number(match[2]);
		if (Math.abs(lat) > 85 || Math.abs(lon) > 180) {
			throw new Error('Coordinates must be a latitude within ±85° and a longitude within ±180°.');
		}
		return { name: `${lat}, ${lon}`, lat, lon };
	}

	const params = new URLSearchParams({ q: query, format: 'jsonv2', limit: '1' });
	const response = await fetch(`https://nominatim.openstreetmap.org/search?${params}`, { signal });
	if (!response.ok) throw new Error(`Place search failed (${response.status}).`);
	const [result] = await response.json();
	if (!result) throw new Error(`No place found for “${query}”.`);
	// Nominatim's bounding box is [south, north, west, east]
	const [south, north, west, east] = (result.boundingbox ?? []).map(Number);
	return {
		name: result.display_name,
		lat: Number(result.lat),
		lon: Number(result.lon),
		bounds: result.boundingbox ? [west, south, east, north] : undefined
	};
}

/** Elevation in metres at a point, read from the (cached) DEM tile at the given zoom. */
export async function elevationAt(dem: DemSource, lat: number, lon: number, zoom: number) {
	const z = Math.min(Math.max(Math.floor(zoom), 0), DEM_MAX_ZOOM);
	const tiles = 2 ** z;
	const position = project(lat, lon, z);
	const x = ((position.x % tiles) + tiles) % tiles;
	const y = position.y;
	if (y < 0 || y >= tiles) return null;

	const tile = await dem.getDemTile(z, Math.floor(x), Math.floor(y));
	const px = Math.floor((x % 1) * tile.width);
	const py = Math.floor((y % 1) * tile.height);
	return tile.data[py * tile.width + px];
}

/**
 * Elevations covering `bounds`, at the most detailed zoom that keeps the grid at most
 * `maxWidth` pixels wide. Pixels beyond the poles are NaN.
 */
export async function fetchRaster(
	dem: DemSource,
	bounds: Bounds,
	maxWidth = 1024
): Promise<Raster> {
	const [west, south, east, north] = bounds;
	const widthAtZoom0 = (project(0, east, 0).x - project(0, west, 0).x) * TILE_SIZE;
	const zoom = Math.min(Math.max(Math.floor(Math.log2(maxWidth / widthAtZoom0)), 0), DEM_MAX_ZOOM);
	const tiles = 2 ** zoom;

	const nw = project(north, west, zoom);
	const se = project(south, east, zoom);
	const left = Math.floor(nw.x * TILE_SIZE);
	const top = Math.floor(nw.y * TILE_SIZE);
	const width = Math.max(2, Math.ceil(se.x * TILE_SIZE) - left);
	const height = Math.max(2, Math.ceil(se.y * TILE_SIZE) - top);
	const data = new Float32Array(width * height).fill(NaN);

	const requests: Promise<void>[] = [];
	for (
		let ty = Math.floor(top / TILE_SIZE);
		ty <= Math.floor((top + height - 1) / TILE_SIZE);
		ty++
	) {
		if (ty < 0 || ty >= tiles) continue;
		for (
			let tx = Math.floor(left / TILE_SIZE);
			tx <= Math.floor((left + width - 1) / TILE_SIZE);
			tx++
		) {
			requests.push(
				dem.getDemTile(zoom, ((tx % tiles) + tiles) % tiles, ty).then((tile) => {
					const originX = tx * TILE_SIZE - left;
					const originY = ty * TILE_SIZE - top;
					for (let py = Math.max(0, -originY); py < Math.min(TILE_SIZE, height - originY); py++) {
						for (let px = Math.max(0, -originX); px < Math.min(TILE_SIZE, width - originX); px++) {
							data[(originY + py) * width + originX + px] = tile.data[py * tile.width + px];
						}
					}
				})
			);
		}
	}
	await Promise.all(requests);
	return { width, height, data, bounds, zoom, left, top };
}
