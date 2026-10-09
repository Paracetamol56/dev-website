<script lang="ts">
	import 'maplibre-gl/dist/maplibre-gl.css';
	import type { Map as MapLibreMap, Marker } from 'maplibre-gl';
	import { onDestroy, onMount } from 'svelte';
	import {
		DEM_MAX_ZOOM,
		elevationAt,
		loadDem,
		type Bounds,
		type DemSource,
		type Place
	} from './elevation';

	export let place: Place;
	/** Width / height of the export frame drawn over the map */
	export let ratio: number;

	// Hypsometric tints, shared by the relief layer and the legend. The tiles store the sea
	// surface as 0 m, so 0 m and below is water and land starts just above it.
	const RELIEF: [number, string][] = [
		[-500, '#1f4e8c'],
		[0, '#8fc1e3'],
		[0.5, '#4d8a4a'],
		[300, '#a9c46c'],
		[1000, '#efe3a3'],
		[2000, '#c79a5b'],
		[3000, '#8c5a3c'],
		[4500, '#f4f1ec']
	];
	const LEGEND = RELIEF.filter(([elevation]) => elevation !== 0.5);

	// [minor, major] contour spacing in metres, from the given zoom level onwards
	const CONTOURS = { 9: [200, 1000], 11: [100, 500], 12: [50, 250], 13: [20, 100], 15: [10, 50] };

	let container: HTMLDivElement;
	let map: MapLibreMap | null = null;
	let marker: Marker | null = null;
	let dem: DemSource | null = null;
	let error: string | null = null;

	// The export frame fills most of the map while keeping its ratio
	const FRAME_FILL = 0.8;
	let mapWidth = 0;
	let mapHeight = 0;
	$: frameWidth = Math.min(mapWidth * FRAME_FILL, mapHeight * FRAME_FILL * ratio);
	$: frameHeight = frameWidth / ratio;
	$: frameLeft = (mapWidth - frameWidth) / 2;
	$: frameTop = (mapHeight - frameHeight) / 2;

	/** Geographic bounds of the export frame. */
	export function frameBounds(): Bounds | null {
		if (!map) return null;
		const northWest = map.unproject([frameLeft, frameTop]);
		const southEast = map.unproject([frameLeft + frameWidth, frameTop + frameHeight]);
		return [northWest.lng, southEast.lat, southEast.lng, northWest.lat];
	}

	let markerElevation: number | null = null;
	let cursorElevation: number | null = null;
	let cursorRequest = 0;

	const formatElevation = (value: number | null) =>
		value === null ? '–' : `${Math.round(value).toLocaleString('en')} m`;

	async function updateMarkerElevation(place: Place) {
		if (!dem) return;
		markerElevation = null;
		try {
			markerElevation = await elevationAt(dem, place.lat, place.lon, DEM_MAX_ZOOM);
		} catch {
			markerElevation = null;
		}
	}

	function goTo(place: Place) {
		if (!map || !marker) return;
		marker.setLngLat([place.lon, place.lat]);
		if (place.bounds) {
			map.fitBounds(place.bounds, { padding: 40, maxZoom: 13 });
		} else {
			map.flyTo({ center: [place.lon, place.lat], zoom: 12 });
		}
		updateMarkerElevation(place);
	}

	onMount(async () => {
		// MapLibre needs the browser (WebGL, workers), so it is only loaded on the client
		const [maplibregl, loadedDem, { default: workerUrl }] = await Promise.all([
			import('maplibre-gl'),
			loadDem(),
			// MapLibre locates its worker relative to its own file, which bundling breaks
			import('maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url')
		]);
		maplibregl.setWorkerUrl(workerUrl);

		dem = loadedDem;
		dem.setupMaplibre(maplibregl);

		try {
			map = new maplibregl.Map({
				container,
				center: [place.lon, place.lat],
				zoom: 12,
				maxZoom: 17,
				// The export frame is axis-aligned, so the map stays north-up and flat
				dragRotate: false,
				pitchWithRotate: false,
				touchPitch: false,
				attributionControl: { compact: true },
				style: {
					version: 8,
					glyphs: 'https://tiles.openfreemap.org/fonts/{fontstack}/{range}.pbf',
					sources: {
						dem: {
							type: 'raster-dem',
							encoding: 'terrarium',
							tiles: [dem.sharedDemProtocolUrl],
							tileSize: 256,
							maxzoom: DEM_MAX_ZOOM,
							attribution:
								'<a href="https://registry.opendata.aws/terrain-tiles/" target="_blank" rel="noreferrer">Terrain Tiles</a> (Mapzen, AWS Open Data)'
						},
						contours: {
							type: 'vector',
							tiles: [
								dem.contourProtocolUrl({
									thresholds: CONTOURS,
									elevationKey: 'ele',
									levelKey: 'level',
									contourLayer: 'contours',
									overzoom: 1
								})
							],
							maxzoom: 15
						}
					},
					layers: [
						{ id: 'background', type: 'background', paint: { 'background-color': '#1e1e2e' } },
						{
							id: 'relief',
							type: 'color-relief',
							source: 'dem',
							paint: {
								'color-relief-color': ['interpolate', ['linear'], ['elevation'], ...RELIEF.flat()]
							}
						},
						{
							id: 'hillshade',
							type: 'hillshade',
							source: 'dem',
							paint: {
								'hillshade-exaggeration': 0.35,
								'hillshade-shadow-color': 'rgba(0, 0, 0, 0.4)',
								'hillshade-highlight-color': 'rgba(255, 255, 255, 0.2)'
							}
						},
						{
							id: 'contour-lines',
							type: 'line',
							source: 'contours',
							'source-layer': 'contours',
							minzoom: 9,
							paint: {
								'line-color': 'rgba(30, 30, 46, 0.55)',
								'line-width': ['match', ['get', 'level'], 1, 1.1, 0.5]
							}
						},
						{
							id: 'contour-labels',
							type: 'symbol',
							source: 'contours',
							'source-layer': 'contours',
							minzoom: 9,
							filter: ['>', ['get', 'level'], 0],
							layout: {
								'symbol-placement': 'line',
								'text-field': ['concat', ['number-format', ['get', 'ele'], {}], ' m'],
								'text-font': ['Noto Sans Regular'],
								'text-size': 10
							},
							paint: {
								'text-color': '#1e1e2e',
								'text-halo-color': 'rgba(255, 255, 255, 0.7)',
								'text-halo-width': 1
							}
						}
					]
				}
			});
		} catch {
			error = 'This browser cannot display the map (WebGL is required).';
			return;
		}

		map.touchZoomRotate.disableRotation();
		map.keyboard.disableRotation();
		map.addControl(new maplibregl.NavigationControl({ showCompass: false }), 'top-right');
		map.addControl(new maplibregl.ScaleControl({ unit: 'metric' }), 'bottom-right');
		marker = new maplibregl.Marker({ color: '#d20f39' })
			.setLngLat([place.lon, place.lat])
			.addTo(map);

		map.on('mousemove', async (event) => {
			const request = ++cursorRequest;
			const { lat, lng } = event.lngLat.wrap();
			try {
				const value = await elevationAt(dem!, lat, lng, map!.getZoom());
				if (request === cursorRequest) cursorElevation = value;
			} catch {
				if (request === cursorRequest) cursorElevation = null;
			}
		});
		map.on('mouseout', () => {
			cursorRequest++;
			cursorElevation = null;
		});

		updateMarkerElevation(place);
	});

	onDestroy(() => map?.remove());

	// Searches move the map; the initial place is handled when the map is created
	let shown = place;
	$: if (place !== shown) {
		shown = place;
		goTo(place);
	}
</script>

<div
	class="relative h-[70vh] min-h-[24rem] w-full overflow-hidden rounded-md bg-ctp-crust"
	bind:clientWidth={mapWidth}
	bind:clientHeight={mapHeight}
>
	<!-- MapLibre makes this element position: relative, so it is sized rather than inset -->
	<div bind:this={container} class="h-full w-full" />

	{#if frameWidth > 0}
		<!-- Export frame; the shadow dims everything outside it -->
		<div
			class="pointer-events-none absolute rounded-sm border-2 border-white/90"
			style="left: {frameLeft}px; top: {frameTop}px; width: {frameWidth}px; height: {frameHeight}px; box-shadow: 0 0 0 100vmax rgba(17, 17, 27, 0.45);"
			aria-hidden="true"
		/>
	{/if}

	{#if error}
		<p class="absolute inset-0 grid place-items-center p-4 text-center text-ctp-red">{error}</p>
	{:else}
		<div
			class="pointer-events-none absolute left-2 top-2 flex flex-col gap-2 rounded-md bg-ctp-mantle/90 p-2 text-xs shadow-md shadow-ctp-crust"
		>
			<p>
				<span class="text-ctp-subtext0">Marker</span>
				<span class="font-semibold text-ctp-text">{formatElevation(markerElevation)}</span>
				{#if cursorElevation !== null}
					<span class="ml-2 text-ctp-subtext0">Cursor</span>
					<span class="font-semibold text-ctp-text">{formatElevation(cursorElevation)}</span>
				{/if}
			</p>
			<div>
				<div
					class="h-2 w-56 rounded-sm"
					style="background: linear-gradient(to right, {LEGEND.map(([, color]) => color).join(
						', '
					)})"
				/>
				<div class="mt-0.5 flex w-56 justify-between text-[10px] text-ctp-subtext0">
					{#each LEGEND as [elevation]}
						<span>{elevation === 0 ? 'sea' : elevation.toLocaleString('en')}</span>
					{/each}
				</div>
			</div>
		</div>
	{/if}
</div>
