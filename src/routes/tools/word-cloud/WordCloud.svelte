<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import { ACCENTS } from '$lib/colors';
	import { downloadBlob } from '$lib/download';
	import { user } from '$lib/store';
	import catppuccin from '@catppuccin/palette';
	import cloud from 'd3-cloud';
	import { FileCode, FileImage } from 'lucide-svelte';
	import { onDestroy } from 'svelte';

	export let data: { text: string; occurence: number }[];
	export let exportable = true;
	export let title = '';
	export let filename = 'word-cloud';

	const FONT = 'Inter, sans-serif';
	const RELAYOUT_DELAY = 300;

	let width = 0;
	$: height = Math.round((width * 9) / 16);
	$: palette = catppuccin.variants[$user.flavour];

	type PlacedWord = cloud.Word & { color: string };
	let words: PlacedWord[] = [];
	let dropped = 0;

	const hash = (text: string) => [...text].reduce((h, c) => (h * 31 + c.charCodeAt(0)) >>> 0, 7);
	$: colorOf = (text: string) => palette[ACCENTS[hash(text) % ACCENTS.length]].hex;

	// Seeded, so a new word does not reshuffle the whole cloud
	function seededRandom() {
		let seed = 42;
		return () => {
			seed = (seed * 16807) % 2147483647;
			return (seed - 1) / 2147483646;
		};
	}

	const MAX_ATTEMPTS = 4;

	function layout(entries: typeof data, width: number, height: number, attempt = 0) {
		if (!width || !height || entries.length === 0) {
			words = [];
			dropped = 0;
			return;
		}
		const counts = entries.map((entry) => Math.sqrt(entry.occurence));
		const [low, high] = [Math.min(...counts), Math.max(...counts)];
		// The more words, the less room each one gets; every retry shrinks the words further
		const shrink = 0.8 ** attempt;
		const share = Math.sqrt((width * height) / entries.length);
		const smallest = Math.max(10, Math.min(height * 0.04, share * 0.35) * shrink);
		const largest = Math.max(smallest, Math.min(height * 0.2, share * 0.9) * shrink);
		const size = (count: number) =>
			high === low
				? smallest + (largest - smallest) * 0.4
				: smallest + ((count - low) / (high - low)) * (largest - smallest);

		cloud()
			.size([width, height])
			.words(entries.map((entry, i) => ({ text: entry.text, size: size(counts[i]) })))
			.padding(3)
			.rotate(0)
			.font(FONT)
			.fontWeight('bold')
			.fontSize((word) => word.size!)
			.random(seededRandom())
			.on('end', (placed) => {
				if (placed.length < entries.length && attempt < MAX_ATTEMPTS - 1) {
					layout(entries, width, height, attempt + 1);
					return;
				}
				laidOut = true;
				words = placed.map((word) => ({ ...word, color: colorOf(word.text!) }));
				dropped = entries.length - placed.length;
			})
			.start();
	}

	let timer: ReturnType<typeof setTimeout>;
	let laidOut = false;
	$: {
		clearTimeout(timer);
		const [entries, w, h] = [data, width, height];
		timer = setTimeout(() => layout(entries, w, h), laidOut ? RELAYOUT_DELAY : 0);
	}
	$: recolor(colorOf);
	function recolor(color: (text: string) => string) {
		words = words.map((word) => ({ ...word, color: color(word.text!) }));
	}
	onDestroy(() => clearTimeout(timer));

	const escape = (text: string) =>
		text.replace(/[&<>"]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c]!));

	function svgMarkup() {
		const texts = words
			.map(
				(word) =>
					`<text transform="translate(${word.x} ${word.y})" font-size="${word.size}" fill="${
						word.color
					}">${escape(word.text!)}</text>`
			)
			.join('');
		return `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="${
			-width / 2
		} ${
			-height / 2
		} ${width} ${height}" text-anchor="middle" font-family="${FONT}" font-weight="bold">${texts}</svg>`;
	}

	function exportSvg() {
		downloadBlob(new Blob([svgMarkup()], { type: 'image/svg+xml' }), `${filename}.svg`);
	}

	async function exportPng() {
		const url = URL.createObjectURL(new Blob([svgMarkup()], { type: 'image/svg+xml' }));
		try {
			const image = new Image();
			image.src = url;
			await image.decode();
			const canvas = document.createElement('canvas');
			canvas.width = width * 2;
			canvas.height = height * 2;
			canvas.getContext('2d')?.drawImage(image, 0, 0, canvas.width, canvas.height);
			canvas.toBlob((blob) => blob && downloadBlob(blob, `${filename}.png`), 'image/png');
		} finally {
			URL.revokeObjectURL(url);
		}
	}
</script>

<div class="flex flex-col gap-2">
	{#if exportable}
		<div class="flex flex-wrap items-center justify-between gap-2">
			{#if title}<h2 class="text-xl font-bold text-ctp-text">{title}</h2>{/if}
			<div class="ml-auto flex gap-2">
				<Button on:click={exportSvg} disabled={!words.length}>
					<FileCode size="16" />
					<span>SVG</span>
				</Button>
				<Button on:click={exportPng} disabled={!words.length}>
					<FileImage size="16" />
					<span>PNG</span>
				</Button>
			</div>
		</div>
	{:else if title}
		<h2 class="text-xl font-bold text-ctp-text">{title}</h2>
	{/if}
	<div class="w-full" bind:clientWidth={width}>
		<svg
			class="block max-w-full"
			{width}
			{height}
			viewBox="{-width / 2} {-height / 2} {width} {height}"
			text-anchor="middle"
			font-family={FONT}
			font-weight="bold"
			role="img"
			aria-label="Word cloud of {data.length} words"
		>
			{#each words as word (word.text)}
				<text transform="translate({word.x} {word.y})" font-size={word.size} fill={word.color}
					>{word.text}</text
				>
			{/each}
		</svg>
		{#if data.length === 0}
			<p class="py-8 text-center text-ctp-subtext0">No words yet</p>
		{:else if dropped > 0}
			<p class="text-sm text-ctp-subtext0" aria-live="polite">
				{dropped} less frequent word{dropped === 1 ? '' : 's'} did not fit
			</p>
		{/if}
	</div>
</div>
