<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import ColorPicker from '$lib/components/ColorPicker.svelte';
	import MeltRadioGroup from '$lib/components/MeltRadioGroup.svelte';
	import MeltSelect from '$lib/components/MeltSelect.svelte';
	import MeltSlider from '$lib/components/MeltSlider.svelte';
	import NumberInput from '$lib/components/NumberInput.svelte';
	import type { Palette } from '$lib/colors';
	import { downloadBlob } from '$lib/download';
	import type { SelectOption } from '@melt-ui/svelte';
	import { Download } from 'lucide-svelte';
	import { writable } from 'svelte/store';
	import GradientEditor from './GradientEditor.svelte';
	import LayerSection from './LayerSection.svelte';
	import type { RoadDetail } from './features';
	import { DEFAULT_STYLE, type OutputStyle } from './style';

	/** Edited here, previewed by the map */
	export let style: OutputStyle;
	export let palette: Palette;
	/** Renders the full-size SVG of the frame */
	export let render: () => Promise<string>;
	/** Elevation range (min, max) of the frame, for the gradient editor */
	export let frameRange: () => Promise<[number, number]>;

	// The interval slider steps through these; 'auto' follows the map zoom
	const INTERVALS = ['auto', 5, 10, 20, 50, 100, 200, 500] as const;
	const intervalLabel = (index: number) =>
		INTERVALS[index] === 'auto' ? 'Auto (follows zoom)' : `${INTERVALS[index]} m`;

	const roadDetailOptions: SelectOption<string>[] = [
		{ value: 'major', label: 'Major roads' },
		{ value: 'minor', label: 'Major and local roads' },
		{ value: 'all', label: 'All roads, tracks and paths' }
	];

	const interval = writable([0]);
	const majorEvery = writable(DEFAULT_STYLE.majorEvery);
	const minorWidth = writable(DEFAULT_STYLE.minorWidth);
	const majorWidth = writable(DEFAULT_STYLE.majorWidth);
	const color = writable(DEFAULT_STYLE.color);
	const water = writable(DEFAULT_STYLE.water);
	const waterColor = writable(DEFAULT_STYLE.waterColor);
	const waterWidth = writable(DEFAULT_STYLE.waterWidth);
	const roads = writable(DEFAULT_STYLE.roads);
	const roadColor = writable(DEFAULT_STYLE.roadColor);
	const roadWidth = writable(DEFAULT_STYLE.roadWidth);
	const roadDetail = writable(roadDetailOptions[1]);
	const buildings = writable(DEFAULT_STYLE.buildings);
	const buildingColor = writable(DEFAULT_STYLE.buildingColor);
	const backgroundMode = writable<string>(DEFAULT_STYLE.backgroundMode);
	const background = writable(DEFAULT_STYLE.background);
	const gradient = writable(DEFAULT_STYLE.gradient);
	const width = writable(DEFAULT_STYLE.width);

	const positive = (value: number) => Math.max(0, Number(value) || 0);
	$: style = {
		interval: INTERVALS[$interval[0]] ?? 'auto',
		majorEvery: Math.max(1, Math.round($majorEvery) || 1),
		color: $color,
		minorWidth: positive($minorWidth),
		majorWidth: positive($majorWidth),
		backgroundMode: $backgroundMode as OutputStyle['backgroundMode'],
		background: $background,
		gradient: $gradient,
		water: $water,
		waterColor: $waterColor,
		waterWidth: positive($waterWidth),
		roads: $roads,
		roadDetail: $roadDetail.value as RoadDetail,
		roadColor: $roadColor,
		roadWidth: positive($roadWidth),
		buildings: $buildings,
		buildingColor: $buildingColor,
		width: Math.min(Math.max(Math.round($width) || 0, 64), 8192)
	};

	let exporting: 'svg' | 'png' | null = null;
	let error: string | null = null;

	async function download(format: 'svg' | 'png') {
		exporting = format;
		error = null;
		try {
			const svg = await render();
			if (format === 'svg') {
				downloadBlob(new Blob([svg], { type: 'image/svg+xml' }), 'topographic-map.svg');
			} else {
				downloadBlob(await rasterize(svg), 'topographic-map.png');
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not create the image.';
		} finally {
			exporting = null;
		}
	}

	async function rasterize(svg: string): Promise<Blob> {
		const url = URL.createObjectURL(new Blob([svg], { type: 'image/svg+xml' }));
		try {
			const image = new Image();
			image.src = url;
			await image.decode();
			const canvas = document.createElement('canvas');
			canvas.width = image.naturalWidth;
			canvas.height = image.naturalHeight;
			canvas.getContext('2d')?.drawImage(image, 0, 0);
			return await new Promise((resolve, reject) =>
				canvas.toBlob((blob) => (blob ? resolve(blob) : reject(new Error('PNG export failed.'))))
			);
		} finally {
			URL.revokeObjectURL(url);
		}
	}
</script>

<div class="flex flex-col gap-10">
	<section class="flex flex-col gap-4">
		<h3 class="text-lg font-semibold text-ctp-text">Contour lines</h3>
		<div class="flex flex-wrap items-start gap-8">
			<div class="flex min-w-[16rem] flex-1 flex-col gap-6">
				<MeltSlider
					name="Interval"
					value={interval}
					min={0}
					max={INTERVALS.length - 1}
					step={1}
					defaultValue={0}
					format={intervalLabel}
					ticks={INTERVALS.map((value) => (value === 'auto' ? 'Auto' : String(value)))}
				/>
				<div class="grid grid-cols-3 gap-4">
					<NumberInput label="Major every" value={majorEvery} min={1} max={20} />
					<NumberInput label="Minor width" value={minorWidth} min={0} max={20} step={0.5} />
					<NumberInput label="Major width" value={majorWidth} min={0} max={20} step={0.5} />
				</div>
				<p class="text-xs text-ctp-subtext0">
					Widths are in pixels of the exported image. Every n-th line is a thicker major line.
				</p>
			</div>
			<ColorPicker label="Line colour" value={color} {palette} />
		</div>
	</section>

	<section class="flex flex-col gap-4">
		<h3 class="text-lg font-semibold text-ctp-text">Layers from OpenStreetMap</h3>
		<div class="grid gap-4 xl:grid-cols-3">
			<LayerSection name="water" label="Water" enabled={water}>
				<ColorPicker label="Water colour" value={waterColor} {palette} />
				<div class="w-36">
					<NumberInput label="River width (px)" value={waterWidth} min={0} max={20} step={0.5} />
				</div>
			</LayerSection>
			<LayerSection name="roads" label="Roads" enabled={roads}>
				<ColorPicker label="Road colour" value={roadColor} {palette} />
				<div class="flex flex-wrap items-end gap-4">
					<div class="w-36">
						<NumberInput label="Road width (px)" value={roadWidth} min={0} max={20} step={0.5} />
					</div>
					<MeltSelect name="Detail" options={roadDetailOptions} value={roadDetail} />
				</div>
			</LayerSection>
			<LayerSection name="buildings" label="Buildings" enabled={buildings}>
				<ColorPicker label="Building colour" value={buildingColor} {palette} />
				<p class="text-xs text-ctp-subtext0">Buildings appear when zoomed in on a town.</p>
			</LayerSection>
		</div>
	</section>

	<section class="flex flex-col gap-4">
		<h3 class="text-lg font-semibold text-ctp-text">Background</h3>
		<MeltRadioGroup
			name="background"
			options={['color', 'elevation', 'transparent']}
			value={backgroundMode}
		/>
		{#if $backgroundMode === 'color'}
			<ColorPicker label="Background colour" value={background} {palette} />
		{:else if $backgroundMode === 'elevation'}
			<GradientEditor stops={gradient} {palette} fit={frameRange} />
		{/if}
	</section>

	<section class="flex flex-col gap-4">
		<h3 class="text-lg font-semibold text-ctp-text">Image</h3>
		<div class="flex flex-wrap items-end gap-4">
			<div class="w-40">
				<NumberInput label="Width (px)" value={width} min={64} max={8192} step={100} />
			</div>
			<div class="ml-auto flex flex-wrap items-center gap-2">
				{#if error}
					<p class="text-sm text-ctp-red">{error}</p>
				{/if}
				<Button on:click={() => download('svg')} disabled={exporting !== null}>
					<Download size="16" />
					{exporting === 'svg' ? 'Preparing...' : 'Download SVG'}
				</Button>
				<Button on:click={() => download('png')} disabled={exporting !== null}>
					<Download size="16" />
					{exporting === 'png' ? 'Preparing...' : 'Download PNG'}
				</Button>
			</div>
		</div>
	</section>
</div>
