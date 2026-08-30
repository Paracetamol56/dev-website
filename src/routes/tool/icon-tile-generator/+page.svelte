<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import MeltSelect from '$lib/components/MeltSelect.svelte';
	import MeltRadioGroup from '$lib/components/MeltRadioGroup.svelte';
	import NumberInput from '$lib/components/NumberInput.svelte';
	import TextInput from '$lib/components/TextInput.svelte';
	import ColorSwatchPicker from '$lib/components/ColorSwatchPicker.svelte';
	import IconSearch from './IconSearch.svelte';
	import { writable, type Writable } from 'svelte/store';
	import type { SelectOption } from '@melt-ui/svelte';
	import { Dice5, Download, Shuffle } from 'lucide-svelte';
	import palette from '@catppuccin/palette';
	import { user } from '$lib/store';
	import { page } from '$app/stores';
	import { fetchIcon } from './icon-fetcher';
	import type { ColorMode, ShapeOption, FormatOption } from './types';

	type SourceOption = SelectOption<'lucide' | 'simpleicons'>;

	let sourceOptions: SourceOption[] = [
		{ value: 'lucide', label: 'Lucide' },
		{ value: 'simpleicons', label: 'Simple Icons' }
	];
	let source: Writable<SourceOption> = writable(sourceOptions[0]);

	// Define the icon sources with their properties
	const LUCIDE = {
		value: 'lucide',
		label: 'Lucide',
		slugStyle: 'kebab',
		browseUrl: 'https://lucide.dev/icons/'
	};

	const SIMPLE_ICONS = {
		value: 'simpleicons',
		label: 'Simple Icons',
		slugStyle: 'lower',
		browseUrl: 'https://simpleicons.org/'
	};

	// URL parameter handling
	$: {
		// This will run on both server and client
		$page.url.searchParams.forEach((value, key) => {
			switch (key) {
				case 'icon':
					iconName.set(value);
					break;
				case 'bgMode':
					bgMode.set(value as ColorMode);
					break;
				case 'fgMode':
					fgMode.set(value as ColorMode);
					break;
				case 'bgCatppuccin':
					bgCatppuccin.set(value);
					break;
				case 'fgCatppuccin':
					fgCatppuccin.set(value);
					break;
				case 'bgCustom':
					bgCustom.set(value);
					break;
				case 'fgCustom':
					fgCustom.set(value);
					break;
				case 'shape':
					const shapeOption = shapeOptions.find((option) => option.value === value);
					if (shapeOption) {
						shape.set(shapeOption);
					}
					break;
				case 'format':
					const formatOption = formatOptions.find((option) => option.value === value);
					if (formatOption) {
						format.set(formatOption);
					}
					break;
				case 'resolution':
					resolution.set(parseInt(value));
					break;
				case 'padding':
					padding.set(parseFloat(value));
					break;
				case 'strokeWidth':
					strokeWidth.set(parseFloat(value));
					break;
			}
		});
	}

	// Watch for changes and update URL (only on client side)
	$: if (typeof window !== 'undefined') {
		const params = new URLSearchParams();

		// Icon settings
		if ($iconName) params.set('icon', $iconName);

		// Color settings
		if ($bgMode !== 'catppuccin') params.set('bgMode', $bgMode);
		if ($fgMode !== 'catppuccin') params.set('fgMode', $fgMode);
		if ($bgCatppuccin !== 'base') params.set('bgCatppuccin', $bgCatppuccin);
		if ($fgCatppuccin !== 'text') params.set('fgCatppuccin', $fgCatppuccin);
		if ($bgCustom !== '#1e1e2e') params.set('bgCustom', $bgCustom);
		if ($fgCustom !== '#cdd6f4') params.set('fgCustom', $fgCustom);

		// Shape settings
		if ($shape.value !== 'square') params.set('shape', $shape.value);

		// Format settings
		if ($format.value !== 'png') params.set('format', $format.value);

		// Numeric settings
		if ($resolution !== 512) params.set('resolution', $resolution.toString());
		if ($padding !== 0.18) params.set('padding', $padding.toString());
		if ($strokeWidth !== 2.0) params.set('strokeWidth', $strokeWidth.toString());

		// Update URL
		const newUrl = `${window.location.pathname}?${params.toString()}`;
		history.replaceState({}, '', newUrl);
	}

	let iconName: Writable<string> = writable('camera');

	// Color settings
	type ColorMode = 'catppuccin' | 'random' | 'custom' | 'transparent';

	const CATPPUCCIN_COLORS = palette.variants[$user.flavour];

	let bgMode: Writable<ColorMode> = writable('catppuccin');
	let fgMode: Writable<ColorMode> = writable('catppuccin');

	let bgCatppuccin: Writable<string> = writable('base');
	let fgCatppuccin: Writable<string> = writable('text');

	let bgCustom: Writable<string> = writable('#1e1e2e');
	let fgCustom: Writable<string> = writable('#cdd6f4');

	function randomCatppuccinKey(): string {
		const keys = Object.keys(CATPPUCCIN_COLORS);
		return keys[Math.floor(Math.random() * keys.length)];
	}

	function generateRandomHexColor(): string {
		return (
			'#' +
			Math.floor(Math.random() * 16777215)
				.toString(16)
				.padStart(6, '0')
		);
	}

	function rollRandomColors() {
		if ($bgMode === 'random') {
			rollRandomBg();
		} else if ($bgMode === 'catppuccin') {
			bgCatppuccin.set(randomCatppuccinKey());
		} else if ($bgMode === 'custom') {
			bgCustom.set(generateRandomHexColor());
		}

		if ($fgMode === 'random') {
			rollRandomFg();
		} else if ($fgMode === 'catppuccin') {
			fgCatppuccin.set(randomCatppuccinKey());
		} else if ($fgMode === 'custom') {
			fgCustom.set(generateRandomHexColor());
		}
	}

	function rollRandomBg() {
		if (Math.random() > 0.5) {
			bgCatppuccin.set(randomCatppuccinKey());
		} else {
			bgCustom.set(generateRandomHexColor());
		}
	}

	function rollRandomFg() {
		if (Math.random() > 0.5) {
			fgCatppuccin.set(randomCatppuccinKey());
		} else {
			fgCustom.set(generateRandomHexColor());
		}
	}

	// --- Shape / format / resolution -------------------------------------------
	type ShapeOption = SelectOption<'square' | 'round'>;
	let shapeOptions: ShapeOption[] = [
		{ value: 'square', label: 'Square' },
		{ value: 'round', label: 'Round' }
	];
	let shape: Writable<ShapeOption> = writable(shapeOptions[0]);
	$: radius = $shape.value === 'round' ? 0.5 : 0.0;

	type FormatOption = SelectOption<'png' | 'jpg' | 'webp' | 'svg' | 'bmp' | 'ico'>;
	let formatOptions: FormatOption[] = [
		{ value: 'png', label: 'PNG' },
		{ value: 'jpg', label: 'JPG' },
		{ value: 'webp', label: 'WebP' },
		{ value: 'svg', label: 'SVG' },
		{ value: 'bmp', label: 'BMP' },
		{ value: 'ico', label: 'ICO' }
	];
	let format: Writable<FormatOption> = writable(formatOptions[0]);

	let resolution: Writable<number> = writable(512);
	let padding: Writable<number> = writable(0.18);
	let strokeWidth: Writable<number> = writable(2.0);

	// --- Preview / generation ----------------------------------------------------
	let previewSvg: string | null = null;
	let previewError: string | null = null;
	let isLoading = false;

	function triggerDownload(blob: Blob, filename: string) {
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = filename;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	async function generateAvatarSvg(iconName: string, options) {
		// Fetch the actual icon SVG
		let iconResult = await fetchIcon(iconName);

		// If we couldn't fetch the icon, use a placeholder
		if (!iconResult) {
			iconResult = {
				svg: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<circle cx="12" cy="12" r="10"></circle>
					<path d="M12 16v-4"></path>
					<path d="M12 8h.01"></path>
				</svg>`,
				source: 'lucide',
				renderMode: 'stroke',
				needsViewportTransform: false
			};
		}

		// For Simple Icons, we need to handle them differently since they have their own viewBox
		if (iconResult.source === 'simpleicons' && iconResult.needsViewportTransform) {
			// Extract the inner SVG content without the outer SVG tag
			const innerSvgMatch = iconResult.svg.match(/<svg[^>]*>([\s\S]*)<\/svg>/);
			let innerContent = iconResult.svg;
			if (innerSvgMatch && innerSvgMatch[1]) {
				innerContent = innerSvgMatch[1];
			}

			// Wrap the icon in a container with the specified background and styling
			const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${options.resolution}" height="${options.resolution}" viewBox="0 0 ${options.resolution} ${options.resolution}">
				<rect width="100%" height="100%" rx="${options.radius * options.resolution}" ry="${options.radius * options.resolution}" fill="${options.bg}" />
				<g transform="translate(${options.resolution * options.padding}, ${options.resolution * options.padding}) scale(${1 - 2 * options.padding})">
					<svg width="100%" height="100%" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
						${innerContent
							.replace(/stroke="[^"]*"/g, `stroke="${options.fg}"`)
							.replace(/fill="[^"]*"/g, `fill="${options.fg}"`)}
					</svg>
				</g>
			</svg>`;

			return { svg };
		}

		// Handle Lucide icons and other icons with standard transformation
		// Try to extract viewBox from the SVG
		let viewBoxWidth = 24;
		let viewBoxHeight = 24;
		const viewBoxMatch = iconResult.svg.match(/viewBox="([^"]*)"/);
		if (viewBoxMatch) {
			const viewBoxValues = viewBoxMatch[1].split(' ').map(Number);
			if (viewBoxValues.length === 4) {
				viewBoxWidth = viewBoxValues[2] - viewBoxValues[0];
				viewBoxHeight = viewBoxValues[3] - viewBoxValues[1];
			}
		} else {
			// Try to get width and height attributes
			const widthMatch = iconResult.svg.match(/width="([^"]*)"/);
			const heightMatch = iconResult.svg.match(/height="([^"]*)"/);
			if (widthMatch) viewBoxWidth = parseFloat(widthMatch[1]) || 24;
			if (heightMatch) viewBoxHeight = parseFloat(heightMatch[1]) || 24;
		}

		// Calculate the size and position for the icon within the container
		const containerSize = options.resolution * (1 - 2 * options.padding);
		const scale = Math.min(containerSize / viewBoxWidth, containerSize / viewBoxHeight);
		const iconWidth = viewBoxWidth * scale;
		const iconHeight = viewBoxHeight * scale;
		const iconX = (containerSize - iconWidth) / 2 + options.resolution * options.padding;
		const iconY = (containerSize - iconHeight) / 2 + options.resolution * options.padding;

		// Wrap the icon in a container with the specified background and styling
		const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${options.resolution}" height="${
			options.resolution
		}" viewBox="0 0 ${options.resolution} ${options.resolution}">
			<rect width="100%" height="100%" rx="${options.radius * options.resolution}" ry="${
			options.radius * options.resolution
		}" fill="${options.bg}" />
			<g transform="translate(${iconX}, ${iconY}) scale(${scale})">
				${iconResult.svg
					.replace(/stroke="[^"]*"/g, `stroke="${options.fg}"`)
					.replace(
						/fill="[^"]*"/g,
						iconResult.renderMode === 'fill' ? `fill="${options.fg}"` : 'fill="none"'
					)
					.replace(
						/stroke-width="[^"]*"/g,
						iconResult.renderMode === 'stroke' && options.strokeWidth
							? `stroke-width="${options.strokeWidth}"`
							: ''
					)}
			</g>
		</svg>`;

		return { svg };
	}

	async function generatePreview() {
		isLoading = true;
		previewError = null;
		try {
			const { svg } = await generateAvatarSvg($iconName, {
				resolution: $resolution,
				bg: resolvedBg,
				fg: resolvedFg,
				padding: $padding,
				radius: radius,
				strokeWidth: $strokeWidth
			});
			previewSvg = svg;
		} catch (e) {
			previewError = e instanceof Error ? e.message : 'Failed to fetch icon.';
			previewSvg = null;
		} finally {
			isLoading = false;
		}
	}

	$: resolvedBg =
		$bgMode === 'transparent'
			? 'transparent'
			: $bgMode === 'custom'
			? $bgCustom
			: CATPPUCCIN_COLORS[$bgCatppuccin]
			? CATPPUCCIN_COLORS[$bgCatppuccin].hex
			: $bgCustom;

	$: resolvedFg =
		$fgMode === 'custom'
			? $fgCustom
			: CATPPUCCIN_COLORS[$fgCatppuccin]
			? CATPPUCCIN_COLORS[$fgCatppuccin].hex
			: $fgCustom;

	$: if (
		$iconName ||
		resolvedBg ||
		resolvedFg ||
		$padding ||
		radius ||
		$resolution ||
		$strokeWidth ||
		$shape.value ||
		$bgMode ||
		$fgMode ||
		$bgCatppuccin ||
		$fgCatppuccin ||
		$bgCustom ||
		$fgCustom
	) {
		generatePreview();
	}

	async function downloadImage() {
		if (!previewSvg) return;

		const fmt = $format.value;

		if (fmt === 'svg') {
			const blob = new Blob([previewSvg], { type: 'image/svg+xml' });
			triggerDownload(blob, `${$iconName}.svg`);
			return;
		}

		// Rasterize via canvas for everything else.
		const img = new Image();
		const svgBlob = new Blob([previewSvg], { type: 'image/svg+xml' });
		const url = URL.createObjectURL(svgBlob);

		img.onload = () => {
			const canvas = document.createElement('canvas');
			canvas.width = $resolution;
			canvas.height = $resolution;
			const ctx = canvas.getContext('2d');
			if (!ctx) return;
			ctx.drawImage(img, 0, 0, $resolution, $resolution);
			URL.revokeObjectURL(url);

			const mime = fmt === 'jpg' ? 'image/jpeg' : fmt === 'bmp' ? 'image/bmp' : `image/${fmt}`;

			canvas.toBlob(
				(blob) => {
					if (blob) {
						triggerDownload(blob, `${$iconName}.${fmt}`);
					}
				},
				mime,
				0.92
			);
		};
		img.src = url;
	}
</script>

<svelte:head>
	<title>Icon tile generator - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-4xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				Icon tile generator
			</span>
		</h1>
		<p class="text-center text-ctp-subtext0 mb-8">
			Turn any Lucide or Simple Icons icon into a styled square image — perfect for avatars,
			favicons, or app icons.
		</p>
	</hgroup>

	<div class="flex flex-col gap-8">
		<!-- Icon Search -->
		<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust">
			<h2 class="text-2xl font-bold mb-4 text-ctp-text">Search Icon</h2>

			<IconSearch bind:iconName />
		</div>

		<!-- Preview -->
		<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust">
			<h2 class="text-2xl font-bold mb-4 text-ctp-text">Preview</h2>

			<div
				class="rounded-md overflow-hidden bg-ctp-crust flex items-center justify-center mx-auto"
				style="width: {$resolution}px; height: {$resolution}px;"
			>
				{#if isLoading}
					<p class="text-ctp-subtext0">Loading...</p>
				{:else if previewError}
					<p class="text-ctp-red text-sm px-4 text-center">{previewError}</p>
				{:else if previewSvg}
					{@html previewSvg}
				{:else}
					<p class="text-ctp-subtext0">Enter an icon name to preview</p>
				{/if}
			</div>

			<div class="mt-4 flex justify-end gap-2">
				<Button type="button" on:click={rollRandomColors}>
					<Shuffle size="16" />
					Random colors
				</Button>
				<Button type="button" on:click={downloadImage} disabled={!previewSvg}>
					<Download size="16" />
					Download
				</Button>
			</div>
		</div>

		<!-- Settings Panel -->
		<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust">
			<h2 class="text-2xl font-bold mb-4 text-ctp-text">Settings</h2>

			<div class="space-y-6">
				<!-- Background -->
				<div>
					<h3 class="text-sm font-semibold text-ctp-text mb-2">Background color</h3>
					<div class="mb-2">
						<MeltRadioGroup
							name="bgMode"
							bind:value={bgMode}
							options={['catppuccin', 'random', 'custom', 'transparent']}
						/>
					</div>

					<div class="flex items-center gap-3">
						{#if $bgMode === 'catppuccin'}
							<ColorSwatchPicker palette={CATPPUCCIN_COLORS} bind:selected={bgCatppuccin} />
						{:else if $bgMode === 'random'}
							<Button type="button" on:click={rollRandomBg}>
								<Dice5 size="16" />
								Roll
							</Button>
						{:else if $bgMode === 'custom'}
							<div class="flex items-center gap-3">
								<input
									type="color"
									bind:value={$bgCustom}
									class="w-8 h-8 rounded cursor-pointer border-2 border-ctp-surface0 bg-ctp-surface0"
								/>
								<TextInput bind:value={bgCustom} placeholder="#1e1e2e" class="w-24" />
							</div>
						{:else if $bgMode === 'transparent'}
							<!-- Transparent mode, no color picker needed -->
						{/if}
					</div>
				</div>

				<!-- Foreground -->
				<div>
					<h3 class="text-sm font-semibold text-ctp-text mb-2">Icon color</h3>
					<div class="mb-2">
						<MeltRadioGroup
							name="fgMode"
							bind:value={fgMode}
							options={['catppuccin', 'random', 'custom']}
						/>
					</div>

					<div class="flex items-center gap-3">
						{#if $fgMode === 'catppuccin'}
							<ColorSwatchPicker palette={CATPPUCCIN_COLORS} bind:selected={fgCatppuccin} />
						{:else if $fgMode === 'random'}
							<Button type="button" on:click={rollRandomFg}>
								<Dice5 size="16" />
								Roll
							</Button>
						{:else if $fgMode === 'custom'}
							<div class="flex items-center gap-3">
								<input
									type="color"
									bind:value={$fgCustom}
									class="w-8 h-8 rounded cursor-pointer border-2 border-ctp-surface0 bg-ctp-surface0"
								/>
								<TextInput bind:value={fgCustom} placeholder="#cdd6f4" class="w-24" />
							</div>
						{:else if $fgMode === 'transparent'}
							<!-- Transparent mode, no color picker needed -->
						{/if}
					</div>
				</div>

				<!-- Shape -->
				<div>
					<MeltSelect name="Shape" options={shapeOptions} bind:value={shape} />
				</div>

				<!-- Format & resolution -->
				<div class="grid grid-cols-2 gap-4">
					<div>
						<MeltSelect name="Format" options={formatOptions} bind:value={format} />
					</div>
					<div>
						<NumberInput
							label="Resolution (px)"
							bind:value={resolution}
							min={16}
							max={4096}
							step={16}
						/>
					</div>
				</div>

				<!-- Padding -->
				<div>
					<NumberInput label="Padding" bind:value={padding} min={0} max={0.45} step={0.01} />
				</div>

				<!-- Stroke width -->
				<div>
					<NumberInput
						label="Stroke width"
						bind:value={strokeWidth}
						min={0.5}
						max={4}
						step={0.1}
					/>
				</div>
			</div>
		</div>
	</div>
</section>
