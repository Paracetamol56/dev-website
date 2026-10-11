<script lang="ts">
	import MeltTooltip from '$lib/components/MeltTooltip.svelte';
	import type { Map as MapLibreMap } from 'maplibre-gl';
	import {
		LoaderCircle,
		LocateFixed,
		Minus,
		Navigation2,
		Plus,
		RotateCcw,
		RotateCw
	} from 'lucide-svelte';
	import { createEventDispatcher, onDestroy } from 'svelte';

	export let map: MapLibreMap;

	const ROTATION_STEP = 15;
	const dispatch = createEventDispatcher<{ locate: { lat: number; lon: number } }>();

	let bearing = map.getBearing();
	const onRotate = () => (bearing = map.getBearing());
	map.on('rotate', onRotate);
	onDestroy(() => map.off('rotate', onRotate));

	// The map bearing is the compass direction at the top of the screen
	$: heading = ((Math.round(bearing) % 360) + 360) % 360;

	let locating = false;
	let locateError: string | null = null;

	function locate() {
		if (!('geolocation' in navigator)) {
			locateError = 'Location is not available in this browser.';
			return;
		}
		locating = true;
		locateError = null;
		navigator.geolocation.getCurrentPosition(
			({ coords }) => {
				locating = false;
				dispatch('locate', { lat: coords.latitude, lon: coords.longitude });
			},
			(error) => {
				locating = false;
				locateError =
					error.code === error.PERMISSION_DENIED
						? 'Location access was denied.'
						: 'Could not find your location.';
			},
			{ enableHighAccuracy: true, timeout: 10000 }
		);
	}

	const rotate = (degrees: number) => map.easeTo({ bearing: map.getBearing() + degrees });
	const button =
		'grid square-8 place-items-center text-ctp-text hover:bg-ctp-surface0 focus:outline-none focus-visible:bg-ctp-surface0 focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ctp-mauve disabled:opacity-50';
	const group =
		'flex flex-col overflow-hidden rounded-md bg-ctp-mantle/90 shadow-md shadow-ctp-crust';
</script>

<div class="absolute right-2 top-2 flex flex-col items-end gap-2">
	<div class={group}>
		<MeltTooltip text="Zoom in">
			<button type="button" class={button} aria-label="Zoom in" on:click={() => map.zoomIn()}>
				<Plus size="16" />
			</button>
		</MeltTooltip>
		<MeltTooltip text="Zoom out">
			<button type="button" class={button} aria-label="Zoom out" on:click={() => map.zoomOut()}>
				<Minus size="16" />
			</button>
		</MeltTooltip>
	</div>

	<div class={group}>
		<MeltTooltip text="Rotate left">
			<button
				type="button"
				class={button}
				aria-label="Rotate left"
				on:click={() => rotate(-ROTATION_STEP)}
			>
				<RotateCcw size="16" />
			</button>
		</MeltTooltip>
		<MeltTooltip text="Reset north (heading {heading}°)">
			<button
				type="button"
				class={button}
				aria-label="Reset north, current heading {heading} degrees"
				on:click={() => map.resetNorth()}
			>
				<!-- Points to north as the map turns -->
				<Navigation2
					size="16"
					class="fill-ctp-red stroke-ctp-red"
					style="transform: rotate({-bearing}deg)"
				/>
			</button>
		</MeltTooltip>
		<MeltTooltip text="Rotate right">
			<button
				type="button"
				class={button}
				aria-label="Rotate right"
				on:click={() => rotate(ROTATION_STEP)}
			>
				<RotateCw size="16" />
			</button>
		</MeltTooltip>
	</div>

	<div class={group}>
		<MeltTooltip text="Go to my location">
			<button
				type="button"
				class={button}
				aria-label="Go to my location"
				disabled={locating}
				on:click={locate}
			>
				{#if locating}
					<LoaderCircle size="16" class="animate-spin" />
				{:else}
					<LocateFixed size="16" />
				{/if}
			</button>
		</MeltTooltip>
	</div>

	{#if locateError}
		<p
			class="max-w-[12rem] rounded-md bg-ctp-mantle/90 px-2 py-1 text-xs text-ctp-red shadow-md shadow-ctp-crust"
			role="alert"
		>
			{locateError}
		</p>
	{/if}
</div>
