<script lang="ts">
	import MeltRadioGroup from '$lib/components/MeltRadioGroup.svelte';
	import MeltSelect from '$lib/components/MeltSelect.svelte';
	import NumberInput from '$lib/components/NumberInput.svelte';
	import type { SelectOption } from '@melt-ui/svelte';
	import { get, type Writable } from 'svelte/store';
	import ColorPicker from '$lib/components/ColorPicker.svelte';
	import type { Palette, PaletteColor } from '$lib/colors';
	import { FORMATS, SHAPES, type Format, type IconSource, type Shape } from './types';

	export let palette: Palette;
	export let bg: Writable<PaletteColor>;
	export let fg: Writable<PaletteColor>;
	export let shape: Writable<Shape>;
	export let format: Writable<Format>;
	export let resolution: Writable<number>;
	export let padding: Writable<number>;
	export let strokeWidth: Writable<number>;
	export let iconSource: IconSource | '';

	const formatOptions = FORMATS.map((value) => ({ value, label: value.toUpperCase() }));

	// MeltRadioGroup works on plain strings
	const shapeValue = shape as Writable<string>;

	// MeltSelect works on { value, label } options: present the format store in that shape
	const formatOption: Writable<SelectOption<string>> = {
		subscribe: (run) =>
			format.subscribe((value) =>
				run(formatOptions.find((option) => option.value === value) ?? formatOptions[0])
			),
		set: (option) => format.set(option.value as Format),
		update: (updater) => formatOption.set(updater(get(formatOption)))
	};
</script>

<div class="flex flex-col gap-8">
	<div class="grid gap-8 lg:grid-cols-2">
		<ColorPicker label="Background" value={bg} {palette} allowTransparent />
		<ColorPicker label="Icon" value={fg} {palette} />
	</div>

	<!-- Columns follow the number of fields, since stroke width only applies to Lucide -->
	<div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-none lg:grid-flow-col lg:auto-cols-fr">
		<fieldset class="flex flex-col gap-1">
			<legend class="text-sm font-semibold text-ctp-text mb-1">Shape</legend>
			<div class="flex h-8 items-center">
				<MeltRadioGroup name="shape" options={[...SHAPES]} value={shapeValue} />
			</div>
		</fieldset>
		<NumberInput label="Resolution (px)" value={resolution} min={16} max={4096} step={16} />
		<NumberInput label="Padding" value={padding} min={0} max={0.45} step={0.01} />
		{#if iconSource !== 'simpleicons'}
			<NumberInput label="Stroke width" value={strokeWidth} min={0.5} max={4} step={0.1} />
		{/if}
		<MeltSelect name="Format" options={formatOptions} value={formatOption} />
	</div>
</div>
