<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import MeltRadioGroup from '$lib/components/MeltRadioGroup.svelte';
	import MeltSelect from '$lib/components/MeltSelect.svelte';
	import TextInput from '$lib/components/TextInput.svelte';
	import type { SelectOption } from '@melt-ui/svelte';
	import MeltTooltip from '$lib/components/MeltTooltip.svelte';
	import { RectangleHorizontal, RectangleVertical, Search } from 'lucide-svelte';
	import { onDestroy } from 'svelte';
	import { writable } from 'svelte/store';
	import ExportPanel from './ExportPanel.svelte';
	import TopoMap from './TopoMap.svelte';
	import { computeContours, renderSvg } from './contours';
	import { fetchRaster, findPlace, loadDem, type Place } from './elevation';
	import { DEFAULT_STYLE, contourSpacing, resolveStyle, type OutputStyle } from './style';
	import { user } from '$lib/store';
	import catppuccin from '@catppuccin/palette';

	// Opening the page shows this place without querying Nominatim
	let place: Place = { name: 'Mont Blanc', lat: 45.8326, lon: 6.8652 };
	const query = writable(place.name);

	let loading = false;
	let error: string | null = null;
	let controller: AbortController | null = null;

	async function search() {
		const text = $query.trim();
		if (!text) return;
		controller?.abort();
		const current = (controller = new AbortController());
		loading = true;
		error = null;
		try {
			place = await findPlace(text, current.signal);
		} catch (e) {
			if (current.signal.aborted) return;
			error = e instanceof Error ? e.message : 'Could not find this place.';
		} finally {
			if (controller === current) loading = false;
		}
	}

	onDestroy(() => controller?.abort());

	// Output shapes in landscape, as width / height; the orientation button turns them to portrait
	const frameOptions: SelectOption<string>[] = [
		{ value: '1', label: 'Square 1:1' },
		{ value: String(5 / 4), label: '5:4 (8×10 in)' },
		{ value: String(4 / 3), label: '4:3' },
		{ value: String(7 / 5), label: '7:5 (5×7 in)' },
		{ value: String(3 / 2), label: '3:2 (photo)' },
		{ value: String(Math.SQRT2), label: 'A4, A3… (ISO paper)' },
		{ value: String(11 / 8.5), label: 'US Letter' },
		{ value: String(14 / 8.5), label: 'US Legal' },
		{ value: String(16 / 10), label: '16:10' },
		{ value: String(16 / 9), label: '16:9 (widescreen)' },
		{ value: '2', label: '2:1' },
		{ value: String(64 / 27), label: '21:9 (ultrawide)' },
		{ value: '3', label: '3:1 (panorama)' }
	];
	const frame = writable(frameOptions[5]);
	let portrait = true;
	$: landscapeRatio = Number($frame.value);
	$: ratio = portrait ? 1 / landscapeRatio : landscapeRatio;

	/** "My location" becomes the current place, like a search result */
	function showLocation(event: CustomEvent<{ lat: number; lon: number }>) {
		const { lat, lon } = event.detail;
		error = null;
		place = { name: 'Your location', lat, lon };
	}

	let topoMap: TopoMap;
	let style: OutputStyle = DEFAULT_STYLE;
	// Palette colours follow the user's Catppuccin flavour; drawing uses plain hex colours
	$: palette = catppuccin.variants[$user.flavour];
	$: resolved = resolveStyle(style, palette);

	/** Lowest and highest elevation inside the frame, from a quick low-resolution grid. */
	async function frameRange(): Promise<[number, number]> {
		const frame = topoMap.frameGeometry();
		if (!frame) throw new Error('The map is not ready yet.');
		const raster = await fetchRaster(await loadDem(), frame.bounds, 256);
		let low = Infinity;
		let high = -Infinity;
		for (const value of raster.data) {
			if (value < low) low = value;
			if (value > high) high = value;
		}
		if (low > high) throw new Error('No elevation data here.');
		return [low, high];
	}
	const view = writable<string>('output');

	/** The full-size SVG of the frame, matching what the map previews. */
	async function render(): Promise<string> {
		const frame = topoMap.frameGeometry();
		if (!frame) throw new Error('The map is not ready yet.');
		const [interval, majorEvery] = contourSpacing(resolved, topoMap.zoom());
		// Elevation detail grows with the image size, within what the data offers
		const maxWidth = Math.min(Math.max(resolved.width, 512), 2048);
		const [contours, features] = await Promise.all([
			loadDem()
				.then((dem) => fetchRaster(dem, frame.bounds, maxWidth))
				.then((raster) => computeContours(raster, interval)),
			resolved.water || resolved.roads || resolved.buildings ? topoMap.frameFeatures() : undefined
		]);
		return renderSvg(contours, resolved, majorEvery, frame.corners, features);
	}
</script>

<svelte:head>
	<title>Topographic map - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-4xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				Topographic map
			</span>
		</h1>
		<p class="text-center text-ctp-subtext0 mb-8">Create topographic maps with contour lines.</p>
	</hgroup>

	<div class="flex flex-col gap-8">
		<form
			class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-wrap items-end gap-4"
			on:submit|preventDefault={search}
		>
			<div class="relative flex-1 min-w-[16rem]">
				<Search class="absolute left-3 bottom-2 text-ctp-subtext0" size="16" />
				<TextInput
					label="Place or coordinates"
					value={query}
					placeholder="Mont Blanc, or 45.8326, 6.8652"
					class="pl-9"
				/>
			</div>
			<Button type="submit" disabled={loading}>{loading ? 'Searching...' : 'Show'}</Button>
		</form>

		<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-col gap-4">
			<div aria-live="polite" class="text-sm">
				{#if error}
					<p class="text-ctp-red">{error}</p>
				{:else}
					<p class="font-semibold text-ctp-text">{place.name}</p>
					<p class="text-ctp-subtext0">{place.lat.toFixed(4)}, {place.lon.toFixed(4)}</p>
				{/if}
			</div>

			<TopoMap
				bind:this={topoMap}
				{place}
				style={resolved}
				{palette}
				view={$view === 'relief' ? 'relief' : 'output'}
				{ratio}
				on:locate={showLocation}
			/>

			<div class="flex flex-wrap items-end justify-between gap-4">
				<div class="flex items-end gap-2">
					<MeltSelect name="Output shape" options={frameOptions} value={frame} />
					<MeltTooltip text={portrait ? 'Switch to landscape' : 'Switch to portrait'}>
						<button
							type="button"
							class="grid h-8 w-10 place-items-center rounded-md bg-ctp-surface0 shadow-md shadow-ctp-crust hover:bg-ctp-surface1 focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve disabled:opacity-40"
							aria-label={portrait
								? 'Portrait, switch to landscape'
								: 'Landscape, switch to portrait'}
							disabled={landscapeRatio === 1}
							on:click={() => (portrait = !portrait)}
						>
							{#if portrait}
								<RectangleVertical size="18" />
							{:else}
								<RectangleHorizontal size="18" />
							{/if}
						</button>
					</MeltTooltip>
				</div>
				<fieldset>
					<legend class="text-sm font-semibold text-ctp-text mb-2">Map view</legend>
					<MeltRadioGroup name="view" options={['output', 'relief']} value={view} />
				</fieldset>
			</div>

			<p class="text-xs text-ctp-subtext0">
				Place search: Nominatim, © <a
					class="underline"
					href="https://www.openstreetmap.org/copyright"
					target="_blank"
					rel="noreferrer">OpenStreetMap contributors</a
				>. Drag to move, scroll or pinch to zoom. The frame shows the exported image.
			</p>
		</div>

		<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-col gap-4">
			<h2 class="text-2xl font-bold text-ctp-text">Settings</h2>
			<ExportPanel bind:style {palette} {render} {frameRange} />
		</div>
	</div>
</section>
