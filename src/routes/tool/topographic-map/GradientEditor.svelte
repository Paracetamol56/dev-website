<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import ColorPicker from '$lib/components/ColorPicker.svelte';
	import NumberInput from '$lib/components/NumberInput.svelte';
	import { nearestPaletteKey, resolveColor, type Palette, type PaletteColor } from '$lib/colors';
	import { Maximize2, RotateCcw, Trash2 } from 'lucide-svelte';
	import { derived, writable, type Writable } from 'svelte/store';
	import { DEFAULT_GRADIENT, gradientStops, type GradientStop } from './style';

	export let stops: Writable<GradientStop[]>;
	export let palette: Palette;
	/** Elevation range (min, max) of the exported area, to fit the bar to it */
	export let fit: () => Promise<[number, number]>;

	// Share of the bar within which a dragged stop snaps to sea level
	const SNAP = 0.02;

	const min = writable(-500);
	const max = writable(5000);
	$: span = Math.max(1, $max - $min || 1);
	let selected = 0;

	let bar: HTMLDivElement;
	// Reactive, so markers and gradient follow changes of the range
	$: position = (elevation: number) =>
		Math.min(100, Math.max(0, ((elevation - $min) / span) * 100));
	$: outOfRange = (elevation: number) => elevation < $min || elevation > $max;

	$: hexStops = gradientStops($stops.map((s) => ({ ...s, color: resolveColor(s.color, palette) })));
	$: background =
		hexStops.length > 1
			? `linear-gradient(to right, ${hexStops
					.map(([elevation, color]) => `${color} ${position(elevation)}%`)
					.join(', ')})`
			: hexStops[0]?.[1] ?? 'transparent';

	/** The gradient's colour at an elevation; new stops take the closest palette colour */
	function colorAt(elevation: number): string {
		const rgb = (hex: string) => [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16));
		const upper = hexStops.findIndex(([e]) => e >= elevation);
		if (upper <= 0) return hexStops[Math.max(upper, 0)][1];
		const [low, lowColor] = hexStops[upper - 1];
		const [high, highColor] = hexStops[upper];
		const t = (elevation - low) / (high - low);
		const [a, b] = [rgb(lowColor), rgb(highColor)];
		return `#${a
			.map((c, i) =>
				Math.round(c + (b[i] - c) * t)
					.toString(16)
					.padStart(2, '0')
			)
			.join('')}`;
	}

	function elevationAt(clientX: number) {
		const { left, width } = bar.getBoundingClientRect();
		const elevation = $min + ((clientX - left) / width) * span;
		const clamped = Math.min($max, Math.max($min, elevation));
		return Math.abs(clamped) <= span * SNAP && $min <= 0 && $max >= 0 ? 0 : Math.round(clamped);
	}

	function updateStop(index: number, change: Partial<GradientStop>) {
		stops.update((all) => all.map((stop, i) => (i === index ? { ...stop, ...change } : stop)));
	}

	function addStop(event: PointerEvent) {
		if (event.target !== bar) return;
		const elevation = elevationAt(event.clientX);
		stops.update((all) => [
			...all,
			{ elevation, color: nearestPaletteKey(colorAt(elevation), palette) }
		]);
		selected = $stops.length - 1;
	}

	function removeStop(index: number) {
		if ($stops.length <= 2) return;
		stops.update((all) => all.filter((_, i) => i !== index));
		selected = Math.min(selected, $stops.length - 1);
	}

	let dragging: number | null = null;
	function startDrag(event: PointerEvent, index: number) {
		selected = index;
		dragging = index;
		(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
	}
	function drag(event: PointerEvent) {
		if (dragging !== null) updateStop(dragging, { elevation: elevationAt(event.clientX) });
	}

	function onKeydown(event: KeyboardEvent, index: number) {
		const step = event.shiftKey ? 100 : 10;
		const moves: Record<string, number> = {
			ArrowLeft: -step,
			ArrowDown: -step,
			ArrowRight: step,
			ArrowUp: step
		};
		if (event.key in moves) {
			event.preventDefault();
			updateStop(index, { elevation: $stops[index].elevation + moves[event.key] });
		} else if (event.key === 'Delete' || event.key === 'Backspace') {
			event.preventDefault();
			removeStop(index);
		}
	}

	let fitting = false;
	let fitError: string | null = null;
	async function fitToFrame() {
		fitting = true;
		fitError = null;
		try {
			const [low, high] = await fit();
			min.set(Math.floor(low));
			max.set(Math.max(Math.ceil(high), Math.floor(low) + 1));
		} catch (e) {
			fitError =
				e instanceof Error && e.message !== 'Failed to fetch'
					? e.message
					: 'Could not measure the frame, check your connection.';
		} finally {
			fitting = false;
		}
	}

	// The selected stop's fields, as stores for the inputs
	function stopField<K extends keyof GradientStop>(
		index: number,
		key: K
	): Writable<GradientStop[K]> {
		const { subscribe } = derived(stops, (all) => all[index]?.[key]);
		return {
			subscribe,
			set: (value) => updateStop(index, { [key]: value }),
			update: (updater) => updateStop(index, { [key]: updater($stops[index][key]) })
		};
	}
	$: selectedColor = stopField(selected, 'color') as Writable<PaletteColor>;
	$: selectedElevation = stopField(selected, 'elevation');
	$: seaLevelShown = $min < 0 && $max > 0;
</script>

<div class="flex flex-col gap-4">
	<div class="flex flex-col gap-1">
		<!-- svelte-ignore a11y-no-static-element-interactions - stops are added with the pointer; the stops themselves are keyboard accessible -->
		<div
			bind:this={bar}
			class="relative h-10 cursor-copy rounded-md shadow-md shadow-ctp-crust"
			style="background: {background}"
			on:pointerdown={addStop}
			on:pointermove={drag}
			on:pointerup={() => (dragging = null)}
			title="Click to add a colour"
		>
			{#if seaLevelShown}
				<div
					class="pointer-events-none absolute inset-y-0 border-l border-dashed border-ctp-crust/70"
					style="left: {position(0)}%"
					aria-hidden="true"
				/>
			{/if}
			{#each $stops as stop, index}
				<button
					type="button"
					class="absolute top-1/2 h-12 w-4 -translate-x-1/2 -translate-y-1/2 cursor-ew-resize touch-none rounded-sm border-2 shadow-md shadow-ctp-crust focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve
						{index === selected ? 'z-10 border-ctp-text' : 'border-ctp-surface2'}
						{outOfRange(stop.elevation) ? 'border-dashed opacity-50' : ''}"
					style="left: {position(stop.elevation)}%; background-color: {resolveColor(
						stop.color,
						palette
					)}"
					aria-label="Colour at {stop.elevation} m{outOfRange(stop.elevation)
						? ', outside the range'
						: ''}"
					aria-pressed={index === selected}
					on:pointerdown|stopPropagation={(event) => startDrag(event, index)}
					on:pointermove={drag}
					on:pointerup={() => (dragging = null)}
					on:keydown={(event) => onKeydown(event, index)}
					on:focus={() => (selected = index)}
				/>
			{/each}
		</div>
		<div class="relative h-4 text-xs text-ctp-subtext0">
			<span class="absolute left-0">{$min.toLocaleString('en')} m</span>
			{#if seaLevelShown}
				<span class="absolute -translate-x-1/2" style="left: {position(0)}%">0 m</span>
			{/if}
			<span class="absolute right-0">{$max.toLocaleString('en')} m</span>
		</div>
	</div>

	<div class="flex flex-wrap items-end gap-3">
		<div class="w-28">
			<NumberInput label="Minimum (m)" value={min} min={-11000} max={9000} step={100} />
		</div>
		<div class="w-28">
			<NumberInput label="Maximum (m)" value={max} min={-11000} max={9000} step={100} />
		</div>
		<Button on:click={fitToFrame} disabled={fitting}>
			<Maximize2 size="16" />
			{fitting ? 'Measuring...' : 'Fit to frame'}
		</Button>
		{#if fitError}
			<p class="text-sm text-ctp-red" role="alert">{fitError}</p>
		{/if}
		<Button on:click={() => stops.set(DEFAULT_GRADIENT)}>
			<RotateCcw size="16" />
			Reset colours
		</Button>
	</div>

	{#if $stops[selected]}
		<div class="flex flex-wrap items-start gap-6 rounded-md bg-ctp-base p-4">
			<div class="flex flex-col gap-3">
				<span class="text-sm font-semibold text-ctp-text">Selected colour</span>
				<div class="w-32">
					<NumberInput
						label="Elevation (m)"
						value={selectedElevation}
						min={-11000}
						max={9000}
						step={10}
					/>
				</div>
				<Button on:click={() => removeStop(selected)} disabled={$stops.length <= 2}>
					<Trash2 size="16" />
					Remove
				</Button>
			</div>
			<ColorPicker
				label="Colour at {$stops[selected].elevation} m"
				value={selectedColor}
				{palette}
			/>
		</div>
	{/if}
	<p class="text-xs text-ctp-subtext0">
		Click the bar to add a colour, drag a marker to move it (it snaps to sea level), or select it
		and use the arrow keys. Two colours at the same height make a sharp edge, like the coastline.
	</p>
</div>
