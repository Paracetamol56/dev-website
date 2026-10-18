// Place search uses OpenStreetMap's Nominatim

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
