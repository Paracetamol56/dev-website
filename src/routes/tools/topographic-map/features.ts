import type { GeoJSONFeature, Map as MapLibreMap } from 'maplibre-gl';
import type { Bounds } from './elevation';

// Water bodies and roads come from OpenStreetMap, as vector tiles served by OpenFreeMap
// (OpenMapTiles schema). The map draws them live, and the export reads the same tiles back.

export const OSM_TILES_URL = 'https://tiles.openfreemap.org/planet';
export const OSM_ATTRIBUTION =
	'<a href="https://openfreemap.org" target="_blank" rel="noreferrer">OpenFreeMap</a> © <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noreferrer">OpenStreetMap contributors</a>';

/** A line or ring as [lon, lat] points */
export type Line = [number, number][];

export type RoadClass = 'major' | 'minor' | 'path';
export type RoadDetail = 'major' | 'minor' | 'all';

export type MapFeatures = {
	/** Lakes, sea, wide rivers: each polygon is a list of rings (outer and holes) */
	water: Line[][];
	/** Rivers, streams and canals */
	waterways: Line[];
	roads: { class: RoadClass; line: Line }[];
	/** Building footprints, as polygons */
	buildings: Line[][];
};

export const WATERWAY_CLASSES = ['river', 'stream', 'canal'];

/** OpenMapTiles `transportation` classes for each of our road classes */
export const ROAD_CLASSES: Record<RoadClass, string[]> = {
	major: ['motorway', 'trunk', 'primary', 'secondary'],
	minor: ['tertiary', 'minor', 'service'],
	path: ['track', 'path']
};

/** Road classes drawn at each detail level; local roads leave out service roads */
export function roadClassesFor(detail: RoadDetail): Record<RoadClass, string[]> {
	return {
		major: ROAD_CLASSES.major,
		minor:
			detail === 'major' ? [] : detail === 'minor' ? ['tertiary', 'minor'] : ROAD_CLASSES.minor,
		path: detail === 'all' ? ROAD_CLASSES.path : []
	};
}

function lines(feature: GeoJSONFeature): Line[] {
	const { geometry } = feature;
	if (geometry.type === 'LineString') return [geometry.coordinates as Line];
	if (geometry.type === 'MultiLineString') return geometry.coordinates as Line[];
	return [];
}

function polygons(feature: GeoJSONFeature): Line[][] {
	const { geometry } = feature;
	if (geometry.type === 'Polygon') return [geometry.coordinates as Line[]];
	if (geometry.type === 'MultiPolygon') return geometry.coordinates as Line[][];
	return [];
}

/**
 * Water, roads and buildings of the loaded map tiles that touch `bounds`, exactly as the map draws them.
 * Features come back cut at tile edges, which does not show since areas are filled without
 * an outline and lines simply meet.
 */
export function collectFeatures(
	map: MapLibreMap,
	bounds: Bounds,
	layers: { water: boolean; roads: RoadDetail | null; buildings: boolean }
): MapFeatures {
	const features: MapFeatures = { water: [], waterways: [], roads: [], buildings: [] };
	const [west, south, east, north] = bounds;
	// Bounding-box overlap, so a long straight segment crossing the frame still counts
	const touches = (coordinates: Line) => {
		let minLon = Infinity,
			maxLon = -Infinity,
			minLat = Infinity,
			maxLat = -Infinity;
		for (const [lon, lat] of coordinates) {
			minLon = Math.min(minLon, lon);
			maxLon = Math.max(maxLon, lon);
			minLat = Math.min(minLat, lat);
			maxLat = Math.max(maxLat, lat);
		}
		return minLon <= east && maxLon >= west && minLat <= north && maxLat >= south;
	};
	const query = (sourceLayer: string) => map.querySourceFeatures('osm', { sourceLayer });

	if (layers.water) {
		for (const feature of query('water')) {
			features.water.push(...polygons(feature).filter((rings) => rings.some(touches)));
		}
		for (const feature of query('waterway')) {
			if (!WATERWAY_CLASSES.includes(feature.properties.class)) continue;
			features.waterways.push(...lines(feature).filter(touches));
		}
	}

	if (layers.roads) {
		const classes = roadClassesFor(layers.roads);
		for (const feature of query('transportation')) {
			const roadClass = (Object.keys(classes) as RoadClass[]).find((c) =>
				classes[c].includes(feature.properties.class)
			);
			if (!roadClass) continue;
			for (const line of lines(feature).filter(touches))
				features.roads.push({ class: roadClass, line });
		}
	}

	if (layers.buildings) {
		for (const feature of query('building')) {
			features.buildings.push(...polygons(feature).filter((rings) => rings.some(touches)));
		}
	}
	return features;
}
