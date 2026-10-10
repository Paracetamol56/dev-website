<script lang="ts">
	import type { Writable } from 'svelte/store';
	import {
		ACCENTS,
		NEUTRALS,
		hue,
		resolveColor,
		type Palette,
		type PaletteColor
	} from '$lib/colors';

	export let label: string;
	export let palette: Palette;
	export let value: Writable<PaletteColor>;

	const SIZE = 144;
	const CENTER = SIZE / 2;
	const OUTER = 70;
	const INNER = 44;
	const GAP = 5; // px of space between segments
	const CORNER = 3; // px corner radius, drawn as a round-joined stroke

	const capitalize = (key: string) => key.charAt(0).toUpperCase() + key.slice(1);

	// Order by real hue so the wheel reads correctly in every flavour
	$: accents = [...ACCENTS].sort((a, b) => hue(palette[a].hex) - hue(palette[b].hex));
	$: step = 360 / accents.length;

	function point(radius: number, degrees: number) {
		const rad = (degrees * Math.PI) / 180;
		return `${(CENTER + radius * Math.sin(rad)).toFixed(2)} ${(
			CENTER -
			radius * Math.cos(rad)
		).toFixed(2)}`;
	}

	// Angle that moves a point at `radius` sideways by `px`, keeping the gaps a constant width
	const inset = (px: number, radius: number) => (Math.asin(px / radius) * 180) / Math.PI;

	// The path is shrunk by the corner radius; its rounded stroke grows it back to full size
	function segment(index: number) {
		const outer = OUTER - CORNER;
		const inner = INNER + CORNER;
		const side = GAP / 2 + CORNER;
		const start = index * step;
		const end = (index + 1) * step;
		return [
			`M ${point(outer, start + inset(side, outer))}`,
			`A ${outer} ${outer} 0 0 1 ${point(outer, end - inset(side, outer))}`,
			`L ${point(inner, end - inset(side, inner))}`,
			`A ${inner} ${inner} 0 0 0 ${point(inner, start + inset(side, inner))}`,
			'Z'
		].join(' ');
	}

	// The selected segment moves slightly outwards
	function offset(index: number) {
		const rad = ((index + 0.5) * step * Math.PI) / 180;
		return `translate(${(3 * Math.sin(rad)).toFixed(2)} ${(-3 * Math.cos(rad)).toFixed(2)})`;
	}

	function onKeydown(event: KeyboardEvent, key: string) {
		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			value.set(key);
		}
	}

	let hovered: keyof Palette | null = null;
	$: resolved = resolveColor($value, palette);
	$: shown = hovered ?? ($value in palette ? ($value as keyof Palette) : null);
	$: caption = shown
		? `${capitalize(shown)} · ${palette[shown].hex}`
		: $value === 'transparent'
		? 'Transparent'
		: `Custom · ${resolved}`;
</script>

<div class="flex flex-col items-center gap-2">
	<div class="flex items-center gap-3">
		<svg
			width={SIZE}
			height={SIZE}
			viewBox="0 0 {SIZE} {SIZE}"
			class="overflow-visible"
			role="group"
			aria-label="{label}: palette colours"
		>
			<defs>
				<pattern
					id="checker-{label.toLowerCase()}"
					width="12"
					height="12"
					patternUnits="userSpaceOnUse"
				>
					<rect width="12" height="12" fill="#ccc" />
					<rect width="6" height="6" fill="#888" />
					<rect x="6" y="6" width="6" height="6" fill="#888" />
				</pattern>
			</defs>

			{#each accents as key, index (key)}
				{@const selected = $value === key}
				<g class="group transition-transform" transform={selected ? offset(index) : undefined}>
					<!-- Selection / keyboard focus outline, following the rounded corners -->
					<path
						d={segment(index)}
						stroke-width={2 * CORNER + 4}
						stroke-linejoin="round"
						class="pointer-events-none {selected
							? 'fill-ctp-text stroke-ctp-text'
							: 'fill-ctp-subtext0 stroke-ctp-subtext0 opacity-0 group-has-[:focus-visible]:opacity-100'}"
					/>
					<path
						d={segment(index)}
						fill={palette[key].hex}
						stroke={palette[key].hex}
						stroke-width={2 * CORNER}
						stroke-linejoin="round"
						class="cursor-pointer outline-none"
						role="button"
						tabindex="0"
						aria-label={capitalize(key)}
						aria-pressed={selected}
						on:click={() => value.set(key)}
						on:keydown={(e) => onKeydown(e, key)}
						on:mouseenter={() => (hovered = key)}
						on:mouseleave={() => (hovered = null)}
						on:focus={() => (hovered = key)}
						on:blur={() => (hovered = null)}
					/>
				</g>
			{/each}

			<!-- Current colour -->
			<circle
				cx={CENTER}
				cy={CENTER}
				r={INNER - 8}
				fill={resolved === 'transparent' ? `url(#checker-${label.toLowerCase()})` : resolved}
				class="stroke-ctp-surface0"
				stroke-width="2"
			/>
		</svg>

		<div class="grid grid-cols-2 gap-1" role="group" aria-label="{label}: palette greys">
			{#each NEUTRALS as key (key)}
				{@const selected = $value === key}
				<button
					type="button"
					class="block square-5 rounded outline-none transition-transform hover:scale-110 ring-offset-ctp-mantle
						focus-visible:ring-2 focus-visible:ring-ctp-subtext0
						{selected ? 'ring-2 ring-ctp-mauve' : ''}"
					style="background-color: {palette[key].hex}"
					aria-label={capitalize(key)}
					aria-pressed={selected}
					on:click={() => value.set(key)}
					on:mouseenter={() => (hovered = key)}
					on:mouseleave={() => (hovered = null)}
					on:focus={() => (hovered = key)}
					on:blur={() => (hovered = null)}
				/>
			{/each}
		</div>
	</div>

	<p class="text-sm text-ctp-subtext0 font-mono" aria-live="polite">{caption}</p>
</div>
