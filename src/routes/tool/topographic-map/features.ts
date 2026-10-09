import type { Bounds } from './elevation';

// Water bodies and roads come from OpenStreetMap through the Overpass API.

// Public instances are often overloaded, so a busy or failing one falls through to the next
const OVERPASS_URLS = [
	'https://overpass-api.de/api/interpreter',
	'https://overpass.private.coffee/api/interpreter'
];
const RETRYABLE = new Set([429, 502, 503, 504]);

async function queryOverpass(query: string, signal?: AbortSignal): Promise<Response> {
	let failure = 'Could not reach OpenStreetMap.';
	for (const url of OVERPASS_URLS) {
		try {
			const response = await fetch(url, {
				method: 'POST',
				body: new URLSearchParams({ data: query }),
				signal
			});
			if (response.ok) return response;
			failure = RETRYABLE.has(response.status)
				? 'OpenStreetMap servers are busy, try again in a minute.'
				: `Could not load roads and water (${response.status}).`;
			if (!RETRYABLE.has(response.status)) break;
		} catch (e) {
			if (signal?.aborted) throw e;
		}
	}
	throw new Error(failure);
}

/** A line or ring as [lon, lat] points */
export type Line = [number, number][];

export type RoadClass = 'major' | 'minor' | 'path';
export type RoadDetail = 'major' | 'minor' | 'all';

export type MapFeatures = {
	/** Lakes, reservoirs, wide rivers: each polygon is a list of rings (outer and holes) */
	water: Line[][];
	/** Rivers, streams and canals */
	waterways: Line[];
	coastline: Line[];
	roads: { class: RoadClass; line: Line }[];
};

export type FeatureLayers = { water: boolean; roads: RoadDetail | null };

const ROAD_CLASSES: Record<string, RoadClass> = {
	motorway: 'major',
	motorway_link: 'major',
	trunk: 'major',
	trunk_link: 'major',
	primary: 'major',
	primary_link: 'major',
	secondary: 'major',
	secondary_link: 'major',
	tertiary: 'minor',
	tertiary_link: 'minor',
	unclassified: 'minor',
	residential: 'minor',
	living_street: 'minor',
	service: 'minor',
	track: 'path',
	path: 'path',
	footway: 'path',
	bridleway: 'path',
	cycleway: 'path',
	steps: 'path'
};

/** Highway values included at each detail level */
const ROAD_FILTER: Record<RoadDetail, string> = {
	major: '^(motorway|trunk|primary|secondary)(_link)?$',
	minor:
		'^((motorway|trunk|primary|secondary|tertiary)(_link)?|unclassified|residential|living_street)$',
	all: `^(${Object.keys(ROAD_CLASSES).join('|')})$`
};

type OverpassPoint = { lat: number; lon: number };
type OverpassElement =
	| { type: 'way'; tags?: Record<string, string>; geometry?: OverpassPoint[] }
	| {
			type: 'relation';
			tags?: Record<string, string>;
			members?: { type: string; role: string; geometry?: OverpassPoint[] }[];
	  };

const toLine = (geometry: OverpassPoint[] = []): Line => geometry.map((p) => [p.lon, p.lat]);
const samePoint = (a: [number, number], b: [number, number]) => a[0] === b[0] && a[1] === b[1];

/** Joins the ways of a multipolygon into closed rings, matching their end points. */
function assembleRings(ways: Line[]): Line[] {
	const remaining = ways.filter((way) => way.length > 1).map((way) => [...way]);
	const rings: Line[] = [];
	while (remaining.length) {
		const ring = remaining.pop()!;
		let extended = true;
		while (!samePoint(ring[0], ring[ring.length - 1]) && extended) {
			extended = false;
			const end = ring[ring.length - 1];
			for (let i = 0; i < remaining.length; i++) {
				const way = remaining[i];
				if (samePoint(way[0], end)) ring.push(...way.slice(1));
				else if (samePoint(way[way.length - 1], end)) ring.push(...way.reverse().slice(1));
				else continue;
				remaining.splice(i, 1);
				extended = true;
				break;
			}
		}
		rings.push(ring);
	}
	return rings;
}

export async function fetchFeatures(
	bounds: Bounds,
	layers: FeatureLayers,
	signal?: AbortSignal
): Promise<MapFeatures> {
	const features: MapFeatures = { water: [], waterways: [], coastline: [], roads: [] };
	if (!layers.water && !layers.roads) return features;

	const [west, south, east, north] = bounds;
	const box = `(${south},${west},${north},${east})`;
	const queries: string[] = [];
	if (layers.water) {
		queries.push(
			`way["natural"="water"]${box};`,
			`relation["natural"="water"]${box};`,
			`way["waterway"~"^(river|stream|canal)$"]${box};`,
			`way["natural"="coastline"]${box};`
		);
	}
	if (layers.roads) queries.push(`way["highway"~"${ROAD_FILTER[layers.roads]}"]${box};`);

	const query = `[out:json][timeout:25];(${queries.join('')});out geom;`;
	const response = await queryOverpass(query, signal);
	const { elements } = (await response.json()) as { elements: OverpassElement[] };

	for (const element of elements) {
		const tags = element.tags ?? {};
		if (element.type === 'relation') {
			const members = element.members ?? [];
			const rings = (role: string) =>
				assembleRings(
					members.filter((m) => m.type === 'way' && m.role === role).map((m) => toLine(m.geometry))
				);
			features.water.push([...rings('outer'), ...rings('inner')]);
		} else if (tags.natural === 'water') {
			features.water.push([toLine(element.geometry)]);
		} else if (tags.natural === 'coastline') {
			features.coastline.push(toLine(element.geometry));
		} else if (tags.waterway) {
			features.waterways.push(toLine(element.geometry));
		} else if (tags.highway && ROAD_CLASSES[tags.highway]) {
			features.roads.push({ class: ROAD_CLASSES[tags.highway], line: toLine(element.geometry) });
		}
	}
	return features;
}
