<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import MeltSelect from '$lib/components/MeltSelect.svelte';
	import TextInput from '$lib/components/TextInput.svelte';
	import type { SelectOption } from '@melt-ui/svelte';
	import { Frame, Search } from 'lucide-svelte';
	import { onDestroy } from 'svelte';
	import { writable } from 'svelte/store';
	import ExportPanel from './ExportPanel.svelte';
	import TopoMap from './TopoMap.svelte';
	import { fetchRaster, findPlace, loadDem, type Place, type Raster } from './elevation';

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

	// Output shapes, as width / height
	const frameOptions: SelectOption<string>[] = [
		{ value: '1', label: 'Square' },
		{ value: String(3 / 2), label: 'Landscape 3:2' },
		{ value: String(2 / 3), label: 'Portrait 2:3' },
		{ value: String(Math.SQRT2), label: 'A4 landscape' },
		{ value: String(Math.SQRT1_2), label: 'A4 portrait' },
		{ value: String(16 / 9), label: 'Widescreen 16:9' }
	];
	const frame = writable(frameOptions[0]);

	let topoMap: TopoMap;
	let raster: Raster | null = null;
	let capturing = false;
	let captureError: string | null = null;

	async function previewSelection() {
		const bounds = topoMap.frameBounds();
		if (!bounds) return;
		capturing = true;
		captureError = null;
		try {
			raster = await fetchRaster(await loadDem(), bounds);
		} catch (e) {
			captureError = e instanceof Error ? e.message : 'Could not load elevation data.';
		} finally {
			capturing = false;
		}
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

			<TopoMap bind:this={topoMap} {place} ratio={Number($frame.value)} />

			<div class="flex flex-wrap items-end justify-between gap-4">
				<MeltSelect name="Output shape" options={frameOptions} value={frame} />
				<div class="flex items-center gap-3">
					{#if captureError}
						<p class="text-sm text-ctp-red">{captureError}</p>
					{/if}
					<Button on:click={previewSelection} disabled={capturing}>
						<Frame size="16" />
						{capturing ? 'Loading...' : 'Preview selection'}
					</Button>
				</div>
			</div>

			<p class="text-xs text-ctp-subtext0">
				Place search: Nominatim, © <a
					class="underline"
					href="https://www.openstreetmap.org/copyright"
					target="_blank"
					rel="noreferrer">OpenStreetMap contributors</a
				>. Drag to move, scroll or pinch to zoom; the frame marks the area to export.
			</p>
		</div>

		{#if raster}
			<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-col gap-4">
				<h2 class="text-2xl font-bold text-ctp-text">Output</h2>
				<ExportPanel {raster} />
			</div>
		{/if}
	</div>
</section>
