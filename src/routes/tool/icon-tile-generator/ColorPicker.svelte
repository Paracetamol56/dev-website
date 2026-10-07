<script lang="ts">
	import MeltTooltip from '$lib/components/MeltTooltip.svelte';
	import { Dice5 } from 'lucide-svelte';
	import PaletteWheel from './PaletteWheel.svelte';
	import { writable, type Writable } from 'svelte/store';
	import { parseHex, randomPaletteKey, resolveColor, type Palette, type TileColor } from './colors';

	export let label: string;
	export let value: Writable<TileColor>;
	export let palette: Palette;
	export let allowTransparent = false;

	const id = `color-${label.toLowerCase().replace(/\W+/g, '-')}`;

	$: resolved = resolveColor($value, palette);
	$: transparent = $value === 'transparent';

	// The hex field always shows the colour in use; only complete, valid hex values are applied
	const hexText = writable('');
	$: hexText.set(transparent ? '' : resolved);
	function applyHex(text: string) {
		const hex = parseHex(text);
		if (hex && hex !== resolved) value.set(hex);
	}
	$: applyHex($hexText);
</script>

<fieldset>
	<legend class="text-sm font-semibold text-ctp-text mb-3">{label}</legend>

	<div class="flex flex-wrap items-center gap-6">
		<PaletteWheel {label} {palette} {value} />

		<div class="flex flex-col gap-2">
			<span class="text-xs font-semibold text-ctp-subtext0">Custom colour</span>
			<div class="flex items-center gap-2">
				<label
					class="relative block square-8 shrink-0 rounded-md border-2 border-ctp-surface0 overflow-hidden cursor-pointer
						focus-within:ring-2 focus-within:ring-ctp-mauve
						{transparent
						? 'bg-[conic-gradient(#888_25%,#ccc_0_50%,#888_0_75%,#ccc_0)] bg-[length:12px_12px]'
						: ''}"
					style={transparent ? '' : `background-color: ${resolved}`}
					title="Pick a custom colour"
				>
					<span class="sr-only">{label}: pick a custom colour</span>
					<input
						type="color"
						class="absolute inset-0 opacity-0 cursor-pointer"
						value={transparent ? '#000000' : resolved}
						on:input={(e) => value.set(e.currentTarget.value)}
					/>
				</label>

				<input
					{id}
					type="text"
					bind:value={$hexText}
					placeholder={transparent ? 'none' : '#rrggbb'}
					aria-label="{label} hex code"
					maxlength="7"
					spellcheck="false"
					class="h-8 w-28 rounded-md bg-ctp-surface0 px-3 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-ctp-mauve shadow-md shadow-ctp-crust"
				/>
			</div>

			<div class="flex items-center gap-2">
				{#if allowTransparent}
					<button
						type="button"
						aria-pressed={transparent}
						class="h-8 rounded-md px-3 text-sm bg-ctp-surface0 hover:bg-ctp-surface1 focus:outline-none focus:ring-2 focus:ring-ctp-mauve
							{transparent ? 'ring-2 ring-ctp-mauve text-ctp-mauve' : ''}"
						on:click={() => value.set(transparent ? 'base' : 'transparent')}
					>
						Transparent
					</button>
				{/if}
				<MeltTooltip text="Random palette colour">
					<button
						type="button"
						aria-label="Random {label.toLowerCase()} from the palette"
						class="grid place-items-center square-8 rounded-md bg-ctp-surface0 hover:bg-ctp-surface1 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
						on:click={() => value.set(randomPaletteKey(palette, $value))}
					>
						<Dice5 size="16" />
					</button>
				</MeltTooltip>
			</div>
		</div>
	</div>
</fieldset>
