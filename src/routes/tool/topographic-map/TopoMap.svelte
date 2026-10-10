<script lang="ts">
	import 'maplibre-gl/dist/maplibre-gl.css';
	import type { Map as MapLibreMap, Marker, VectorTileSource } from 'maplibre-gl';
	import { onDestroy, onMount } from 'svelte';
	import {
		DEM_MAX_ZOOM,
		elevationAt,
		loadDem,
		type Bounds,
		type DemSource,
		type Place
	} from './elevation';
	import {
		OSM_ATTRIBUTION,
		OSM_TILES_URL,
		WATERWAY_CLASSES,
		collectFeatures,
		roadClassesFor,
		type MapFeatures
	} from './features';
	import type { Palette } from '$lib/colors';
	import MapControls from './MapControls.svelte';
	import {
		gradientStops,
		ROAD_STYLE,
		contourThresholds,
		type FrameCorners,
		type OutputStyle
	} from './style';

	export let place: Place;
	/** Width / height of the export frame drawn over the map */
	export let ratio: number;
	/** The output style, previewed live inside the frame */
	export let style: OutputStyle;
	/** 'output' shows the map as the exported image, 'relief' as coloured terrain */
	export let view: 'output' | 'relief';
	/** The user's Catppuccin flavour, for the map's own colours (marker, labels, relief view) */
	export let palette: Palette;

	// Legend of the relief view: the gradient, labelled at each distinct stop elevation
	$: legend = gradientStops(style.gradient);
	$: legendLabels = [...new Set(style.gradient.map((stop) => stop.elevation))].sort(
		(a, b) => a - b
	);
	const legendPosition = (elevation: number, from: number, to: number) =>
		((elevation - from) / (to - from || 1)) * 100;

	let container: HTMLDivElement;
	let map: MapLibreMap | null = null;
	let marker: Marker | null = null;
	let dem: DemSource | null = null;
	let error: string | null = null;

	// Everything outside the frame is dimmed with the flavour's darkest colour
	$: frameDim = `rgb(${[1, 3, 5]
		.map((i) => parseInt(palette.crust.hex.slice(i, i + 2), 16))
		.join(' ')} / 0.45)`;

	// The export frame fills most of the map while keeping its ratio
	const FRAME_FILL = 0.8;
	let mapWidth = 0;
	let mapHeight = 0;
	$: frameWidth = Math.min(mapWidth * FRAME_FILL, mapHeight * FRAME_FILL * ratio);
	$: frameHeight = frameWidth / ratio;
	$: frameLeft = (mapWidth - frameWidth) / 2;
	$: frameTop = (mapHeight - frameHeight) / 2;

	/**
	 * The export frame: its corners as [lon, lat] (top-left, top-right, bottom-right,
	 * bottom-left on screen, so rotated with the map) and the bounds enclosing them.
	 */
	export function frameGeometry(): { corners: FrameCorners; bounds: Bounds } | null {
		if (!map) return null;
		const right = frameLeft + frameWidth;
		const bottom = frameTop + frameHeight;
		const corners = (
			[
				[frameLeft, frameTop],
				[right, frameTop],
				[right, bottom],
				[frameLeft, bottom]
			] as const
		).map(([x, y]) => map!.unproject([x, y]).toArray()) as FrameCorners;
		const lons = corners.map(([lon]) => lon);
		const lats = corners.map(([, lat]) => lat);
		return {
			corners,
			bounds: [Math.min(...lons), Math.min(...lats), Math.max(...lons), Math.max(...lats)]
		};
	}

	export function zoom(): number {
		return map?.getZoom() ?? 0;
	}

	/** Water and roads inside the frame, as drawn, once the map has finished loading tiles. */
	export async function frameFeatures(): Promise<MapFeatures> {
		const frame = frameGeometry();
		if (!map || !frame) return { water: [], waterways: [], roads: [], buildings: [] };
		if (!map.areTilesLoaded()) {
			// 'idle' only fires after a render, so request one; a stuck tile must not block the export
			const idle = new Promise((resolve) => map!.once('idle', resolve));
			map.triggerRepaint();
			await Promise.race([idle, new Promise((resolve) => setTimeout(resolve, 10000))]);
		}
		return collectFeatures(map, frame.bounds, {
			water: style.water,
			roads: style.roads ? style.roadDetail : null,
			buildings: style.buildings
		});
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
		const contourUrl = contourTilesUrl(style);
		shownContourUrl = contourUrl;

		try {
			map = new maplibregl.Map({
				container,
				center: [place.lon, place.lat],
				zoom: 12,
				maxZoom: 17,
				// The map can turn under the export frame, but stays flat
				maxPitch: 0,
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
						contours: { type: 'vector', tiles: [contourUrl], maxzoom: 15 },
						osm: { type: 'vector', url: OSM_TILES_URL, attribution: OSM_ATTRIBUTION }
					},
					// Paint values are set by applyStyle once the map has loaded
					layers: [
						{ id: 'background', type: 'background' },
						{
							id: 'relief',
							type: 'color-relief',
							source: 'dem',
							paint: {
								'color-relief-color': [
									'interpolate',
									['linear'],
									['elevation'],
									...gradientStops(style.gradient).flat()
								]
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
						{ id: 'water', type: 'fill', source: 'osm', 'source-layer': 'water' },
						{
							id: 'waterway',
							type: 'line',
							source: 'osm',
							'source-layer': 'waterway',
							filter: ['in', ['get', 'class'], ['literal', WATERWAY_CLASSES]],
							layout: { 'line-cap': 'round', 'line-join': 'round' }
						},
						{ id: 'building', type: 'fill', source: 'osm', 'source-layer': 'building' },
						...(['minor', 'major'] as const).map((kind) => ({
							id: `contour-${kind}`,
							type: 'line' as const,
							source: 'contours',
							'source-layer': 'contours',
							filter: (kind === 'major'
								? ['>=', ['get', 'level'], 1]
								: ['==', ['get', 'level'], 0]) as any,
							layout: { 'line-cap': 'round' as const, 'line-join': 'round' as const }
						})),
						{
							id: 'contour-labels',
							type: 'symbol',
							source: 'contours',
							'source-layer': 'contours',
							filter: ['>', ['get', 'level'], 0],
							layout: {
								'symbol-placement': 'line',
								'text-field': ['concat', ['number-format', ['get', 'ele'], {}], ' m'],
								'text-font': ['Noto Sans Regular'],
								'text-size': 10
							},
							paint: { 'text-halo-width': 1 }
						},
						...(['path', 'minor', 'major'] as const).map((roadClass) => ({
							id: `road-${roadClass}`,
							type: 'line' as const,
							source: 'osm',
							'source-layer': 'transportation',
							layout: { 'line-cap': 'round' as const, 'line-join': 'round' as const }
						}))
					]
				}
			});
		} catch {
			error = 'This browser cannot display the map (WebGL is required).';
			return;
		}

		// Paint as soon as the style is ready, without waiting for the first tiles
		if (map.isStyleLoaded()) loaded = true;
		else map.once('style.load', () => (loaded = true));
		map.addControl(new maplibregl.ScaleControl({ unit: 'metric' }), 'bottom-right');
		markerClass = maplibregl.Marker;
		placeMarker(palette.red.hex);

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

	let loaded = false;
	let shownContourUrl = '';

	function contourTilesUrl(style: OutputStyle) {
		return dem!.contourProtocolUrl({
			thresholds: contourThresholds(style),
			elevationKey: 'ele',
			levelKey: 'level',
			contourLayer: 'contours',
			overzoom: 1
		});
	}

	function setVisible(id: string, visible: boolean) {
		map!.setLayoutProperty(id, 'visibility', visible ? 'visible' : 'none');
	}

	/** Draws the output style; widths are scaled so the frame looks like the exported image. */
	function applyStyle(
		style: OutputStyle,
		view: 'output' | 'relief',
		scale: number,
		palette: Palette
	) {
		if (!map || !dem) return;
		const output = view === 'output';

		const contourUrl = contourTilesUrl(style);
		if (contourUrl !== shownContourUrl) {
			shownContourUrl = contourUrl;
			(map.getSource('contours') as VectorTileSource).setTiles([contourUrl]);
		}

		map.setPaintProperty(
			'background',
			'background-color',
			output ? style.background : palette.base.hex
		);
		// Contour labels contrast with the relief in every flavour
		map.setPaintProperty('contour-labels', 'text-color', palette.crust.hex);
		map.setPaintProperty('contour-labels', 'text-halo-color', palette.text.hex);
		map.setPaintProperty(
			'background',
			'background-opacity',
			output && style.backgroundMode !== 'color' ? 0 : 1
		);
		setVisible('relief', !output || style.backgroundMode === 'elevation');
		setVisible('hillshade', !output);
		setVisible('contour-labels', !output);

		// A colour ramp needs two stops; a single colour is repeated
		const stops = gradientStops(style.gradient);
		const ramp =
			stops.length > 1
				? stops
				: [
						stops[0] ?? [0, palette.base.hex],
						[(stops[0]?.[0] ?? 0) + 1, stops[0]?.[1] ?? palette.base.hex]
				  ];
		map.setPaintProperty('relief', 'color-relief-color', [
			'interpolate',
			['linear'],
			['elevation'],
			...ramp.flat()
		]);

		setVisible('water', style.water);
		setVisible('waterway', style.water);
		map.setPaintProperty('water', 'fill-color', style.waterColor);
		map.setPaintProperty('waterway', 'line-color', style.waterColor);
		map.setPaintProperty('waterway', 'line-width', style.waterWidth * scale);

		for (const [kind, width] of [
			['minor', style.minorWidth],
			['major', style.majorWidth]
		] as const) {
			map.setPaintProperty(`contour-${kind}`, 'line-color', style.color);
			map.setPaintProperty(`contour-${kind}`, 'line-width', Math.max(0, width) * scale);
		}

		setVisible('building', style.buildings);
		map.setPaintProperty('building', 'fill-color', style.buildingColor);

		const classes = roadClassesFor(style.roadDetail);
		for (const roadClass of ['path', 'minor', 'major'] as const) {
			const id = `road-${roadClass}`;
			setVisible(id, style.roads && classes[roadClass].length > 0);
			map.setFilter(id, ['in', ['get', 'class'], ['literal', classes[roadClass]]]);
			map.setPaintProperty(id, 'line-color', style.roadColor);
			map.setPaintProperty(id, 'line-width', style.roadWidth * ROAD_STYLE[roadClass].width * scale);
			const dash = ROAD_STYLE[roadClass].dash;
			if (dash) map.setPaintProperty(id, 'line-dasharray', [...dash]);
		}
	}

	$: previewScale = style.width > 0 ? frameWidth / style.width : 1;
	$: if (loaded) applyStyle(style, view, previewScale, palette);

	// MapLibre markers cannot change colour, so the marker is recreated for a new flavour
	let markerClass: typeof Marker | null = null;
	let markerColor = '';
	function placeMarker(color: string) {
		if (!map || !markerClass || color === markerColor) return;
		marker?.remove();
		markerColor = color;
		marker = new markerClass({ color }).setLngLat([shown.lon, shown.lat]).addTo(map);
	}
	$: placeMarker(palette.red.hex);

	// Searches move the map; the initial place is handled when the map is created
	let shown = place;
	$: if (place !== shown) {
		shown = place;
		goTo(place);
	}
</script>

<div
	class="relative h-[70vh] min-h-[24rem] w-full overflow-hidden rounded-md {view === 'output' &&
	style.backgroundMode === 'transparent'
		? 'bg-[conic-gradient(#888_25%,#ccc_0_50%,#888_0_75%,#ccc_0)] bg-[length:16px_16px]'
		: 'bg-ctp-crust'}"
	bind:clientWidth={mapWidth}
	bind:clientHeight={mapHeight}
>
	<!-- MapLibre makes this element position: relative, so it is sized rather than inset -->
	<div bind:this={container} class="h-full w-full" />

	{#if frameWidth > 0}
		<!-- Export frame; the shadow dims everything outside it -->
		<div
			class="pointer-events-none absolute rounded-sm border-2 border-ctp-text/90"
			style="left: {frameLeft}px; top: {frameTop}px; width: {frameWidth}px; height: {frameHeight}px; box-shadow: 0 0 0 100vmax {frameDim};"
			aria-hidden="true"
		/>
	{/if}

	{#if error}
		<p class="absolute inset-0 grid place-items-center p-4 text-center text-ctp-red">{error}</p>
	{:else}
		{#if loaded && map}
			<MapControls {map} on:locate />
		{/if}

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
			<div class:hidden={view !== 'relief'}>
				<div
					class="h-2 w-56 rounded-sm"
					style="background: linear-gradient(to right, {legend
						.map(
							([elevation, color]) =>
								`${color} ${legendPosition(elevation, legend[0][0], legend[legend.length - 1][0])}%`
						)
						.join(', ')})"
				/>
				<div class="relative mt-0.5 h-3 w-56 text-[10px] text-ctp-subtext0">
					{#each legendLabels as elevation, i}
						<span
							class="absolute whitespace-nowrap {i === 0
								? ''
								: i === legendLabels.length - 1
								? '-translate-x-full'
								: '-translate-x-1/2'}"
							style="left: {legendPosition(
								elevation,
								legendLabels[0],
								legendLabels[legendLabels.length - 1]
							)}%"
						>
							{elevation === 0 ? 'sea' : elevation.toLocaleString('en')}
						</span>
					{/each}
				</div>
			</div>
		</div>
	{/if}
</div>

<style lang="postcss">
	/* MapLibre's own scale bar and attribution, in the site's theme */
	:global(.maplibregl-ctrl.maplibregl-ctrl-scale) {
		@apply rounded-sm border-ctp-text bg-ctp-mantle/90 px-1 text-[10px] text-ctp-text;
	}
	:global(.maplibregl-ctrl.maplibregl-ctrl-attrib) {
		@apply rounded-md bg-ctp-mantle/90 text-ctp-subtext0;
	}
	:global(.maplibregl-ctrl-attrib a) {
		@apply text-ctp-subtext0 underline;
	}
	:global(.maplibregl-ctrl-attrib-button) {
		@apply invert;
	}
</style>
