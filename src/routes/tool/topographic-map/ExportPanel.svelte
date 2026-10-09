<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import MeltCheckbox from '$lib/components/MeltCheckbox.svelte';
	import MeltSelect from '$lib/components/MeltSelect.svelte';
	import NumberInput from '$lib/components/NumberInput.svelte';
	import { downloadBlob } from '$lib/download';
	import type { SelectOption } from '@melt-ui/svelte';
	import { Download } from 'lucide-svelte';
	import { onDestroy } from 'svelte';
	import { writable } from 'svelte/store';
	import { computeContours, renderSvg, type Contours } from './contours';
	import type { Bounds, Raster } from './elevation';
	import { fetchFeatures, type MapFeatures, type RoadDetail } from './features';

	export let raster: Raster;

	const intervalOptions: SelectOption<string>[] = [
		{ value: 'auto', label: 'Auto' },
		...[5, 10, 20, 50, 100, 200, 500].map((m) => ({ value: String(m), label: `${m} m` }))
	];
	const interval = writable(intervalOptions[0]);
	const majorEvery = writable(5);
	const minorWidth = writable(1);
	const majorWidth = writable(2.5);
	const outputWidth = writable(2000);
	const transparent = writable(false);
	let color = '#1e1e2e';
	let background = '#ffffff';

	const showWater = writable(false);
	const waterWidth = writable(2);
	let waterColor = '#1e66f5';
	const showRoads = writable(false);
	const roadWidth = writable(3);
	let roadColor = '#d20f39';
	const roadDetailOptions: SelectOption<string>[] = [
		{ value: 'major', label: 'Major roads' },
		{ value: 'minor', label: 'Major and local roads' },
		{ value: 'all', label: 'All roads, tracks and paths' }
	];
	const roadDetail = writable(roadDetailOptions[1]);

	let contours: Contours | null = null;
	let error: string | null = null;
	let computing = false;
	let request = 0;

	async function update(raster: Raster, intervalValue: string) {
		const current = ++request;
		computing = true;
		error = null;
		try {
			const result = await computeContours(
				raster,
				intervalValue === 'auto' ? 'auto' : Number(intervalValue)
			);
			if (current === request) contours = result;
		} catch (e) {
			if (current === request)
				error = e instanceof Error ? e.message : 'Could not compute contours.';
		} finally {
			if (current === request) computing = false;
		}
	}
	$: update(raster, $interval.value);

	// Map features are fetched once per region and only re-fetched when more data is needed
	const ROAD_DETAIL_RANK: Record<RoadDetail, number> = { major: 0, minor: 1, all: 2 };
	let features: MapFeatures | null = null;
	let loaded: { bounds: Bounds; water: boolean; roads: RoadDetail | null } | null = null;
	let featuresError: string | null = null;
	let loadingFeatures = false;
	let featureController: AbortController | null = null;

	async function updateFeatures(bounds: Bounds, water: boolean, roads: RoadDetail | null) {
		const covered =
			loaded?.bounds === bounds &&
			(!water || loaded.water) &&
			(!roads || (loaded.roads && ROAD_DETAIL_RANK[loaded.roads] >= ROAD_DETAIL_RANK[roads]));
		if (covered || (!water && !roads)) return;

		featureController?.abort();
		const controller = (featureController = new AbortController());
		loadingFeatures = true;
		featuresError = null;
		try {
			// Keep what was already loaded for this region when adding a layer
			const keep = loaded?.bounds === bounds ? loaded : null;
			const request = {
				water: water || !!keep?.water,
				roads:
					keep?.roads && (!roads || ROAD_DETAIL_RANK[keep.roads] > ROAD_DETAIL_RANK[roads])
						? keep.roads
						: roads
			};
			features = await fetchFeatures(bounds, request, controller.signal);
			loaded = { bounds, ...request };
		} catch (e) {
			if (controller.signal.aborted) return;
			featuresError = e instanceof Error ? e.message : 'Could not load roads and water.';
		} finally {
			if (featureController === controller) loadingFeatures = false;
		}
	}
	$: selectedRoadDetail = $roadDetail.value as RoadDetail;
	$: updateFeatures(raster.bounds, $showWater, $showRoads ? selectedRoadDetail : null);

	// Only draw the roads of the selected detail, even if more were loaded
	$: visibleFeatures = features && {
		...features,
		roads: features.roads.filter(
			(road) =>
				selectedRoadDetail === 'all' ||
				road.class === 'major' ||
				(selectedRoadDetail === 'minor' && road.class === 'minor')
		)
	};

	$: width = Math.min(Math.max(Math.round($outputWidth) || 0, 64), 8192);
	$: svg = contours
		? renderSvg(
				contours,
				{
					width,
					color,
					background: $transparent ? null : background,
					minorWidth: Math.max(0, $minorWidth),
					majorWidth: Math.max(0, $majorWidth),
					majorEvery: Math.max(1, Math.round($majorEvery)),
					water: $showWater ? { color: waterColor, width: Math.max(0, $waterWidth) } : null,
					roads: $showRoads ? { color: roadColor, width: Math.max(0, $roadWidth) } : null
				},
				visibleFeatures ?? undefined
		  )
		: null;
	$: height = contours ? Math.round((contours.height * width) / contours.width) : 0;
	$: kilobytes = svg ? Math.round(new Blob([svg]).size / 1024) : 0;

	let previewUrl: string | null = null;
	$: {
		if (previewUrl) URL.revokeObjectURL(previewUrl);
		previewUrl = svg ? URL.createObjectURL(new Blob([svg], { type: 'image/svg+xml' })) : null;
	}
	onDestroy(() => previewUrl && URL.revokeObjectURL(previewUrl));

	function downloadSvg() {
		if (svg) downloadBlob(new Blob([svg], { type: 'image/svg+xml' }), 'topographic-map.svg');
	}

	function downloadPng() {
		if (!previewUrl) return;
		const image = new Image();
		image.onload = () => {
			const canvas = document.createElement('canvas');
			canvas.width = width;
			canvas.height = height;
			canvas.getContext('2d')?.drawImage(image, 0, 0, width, height);
			canvas.toBlob((blob) => blob && downloadBlob(blob, 'topographic-map.png'), 'image/png');
		};
		image.src = previewUrl;
	}
</script>

<div class="flex flex-col gap-6">
	<div class="grid gap-4 sm:grid-cols-3 lg:grid-cols-6 items-end">
		<MeltSelect name="Contour interval" options={intervalOptions} value={interval} />
		<NumberInput label="Major line every" value={majorEvery} min={1} max={20} />
		<NumberInput label="Minor width (px)" value={minorWidth} min={0} max={20} step={0.5} />
		<NumberInput label="Major width (px)" value={majorWidth} min={0} max={20} step={0.5} />
		<NumberInput label="Image width (px)" value={outputWidth} min={64} max={8192} step={100} />
		<div class="flex flex-wrap items-center gap-3">
			<label class="flex items-center gap-2 text-sm font-semibold">
				<input
					type="color"
					bind:value={color}
					class="square-8 cursor-pointer rounded bg-transparent"
				/>
				Lines
			</label>
			<label class="flex items-center gap-2 text-sm font-semibold">
				<input
					type="color"
					bind:value={background}
					disabled={$transparent}
					class="square-8 cursor-pointer rounded bg-transparent disabled:opacity-40"
				/>
				Background
			</label>
			<MeltCheckbox name="export-transparent" label="Transparent" checked={transparent} />
		</div>
	</div>

	<fieldset class="flex flex-col gap-3">
		<legend class="text-sm font-semibold text-ctp-text mb-2">Layers from OpenStreetMap</legend>
		<div class="grid gap-4 sm:grid-cols-2">
			<div class="flex flex-wrap items-end gap-4 rounded-md bg-ctp-base p-3">
				<div class="flex h-8 items-center">
					<MeltCheckbox name="layer-water" label="Water" checked={showWater} />
				</div>
				<label class="flex items-center gap-2 text-sm font-semibold">
					<input
						type="color"
						bind:value={waterColor}
						class="square-8 cursor-pointer rounded bg-transparent"
					/>
					Colour
				</label>
				<div class="w-32">
					<NumberInput label="River width (px)" value={waterWidth} min={0} max={20} step={0.5} />
				</div>
			</div>
			<div class="flex flex-wrap items-end gap-4 rounded-md bg-ctp-base p-3">
				<div class="flex h-8 items-center">
					<MeltCheckbox name="layer-roads" label="Roads" checked={showRoads} />
				</div>
				<label class="flex items-center gap-2 text-sm font-semibold">
					<input
						type="color"
						bind:value={roadColor}
						class="square-8 cursor-pointer rounded bg-transparent"
					/>
					Colour
				</label>
				<div class="w-32">
					<NumberInput label="Road width (px)" value={roadWidth} min={0} max={20} step={0.5} />
				</div>
				<MeltSelect name="Detail" options={roadDetailOptions} value={roadDetail} />
			</div>
		</div>
		{#if loadingFeatures}
			<p class="text-sm text-ctp-subtext0">Loading roads and water...</p>
		{:else if featuresError}
			<p class="text-sm text-ctp-red">{featuresError}</p>
		{/if}
	</fieldset>

	<p class="text-sm text-ctp-subtext0" aria-live="polite">
		{#if error}
			<span class="text-ctp-red">{error}</span>
		{:else if contours}
			Contours every {contours.interval} m · {contours.lines.length} lines · {width} × {height} px ·
			SVG
			{kilobytes} KB
		{:else if computing}
			Computing contours...
		{/if}
	</p>

	{#if previewUrl}
		<div
			class="rounded-md overflow-hidden {$transparent
				? 'bg-[conic-gradient(#888_25%,#ccc_0_50%,#888_0_75%,#ccc_0)] bg-[length:16px_16px]'
				: ''} transition-opacity {computing ? 'opacity-50' : ''}"
		>
			<img
				src={previewUrl}
				alt="Contour lines of the selected region"
				class="block w-full h-auto"
			/>
		</div>
	{/if}

	<div class="flex flex-wrap justify-end gap-2">
		<Button on:click={downloadSvg} disabled={!svg}>
			<Download size="16" />
			Download SVG
		</Button>
		<Button on:click={downloadPng} disabled={!svg}>
			<Download size="16" />
			Download PNG
		</Button>
	</div>
</div>
