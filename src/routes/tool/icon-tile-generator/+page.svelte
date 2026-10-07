<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import IconSearch from './IconSearch.svelte';
	import TileSettings from './TileSettings.svelte';
	import { writable, type Writable } from 'svelte/store';
	import { Download, Shuffle } from 'lucide-svelte';
	import palette from '@catppuccin/palette';
	import { user } from '$lib/store';
	import { page } from '$app/stores';
	import { fetchIcon } from './icon-fetcher';
	import { randomColorPair, resolveColor, type TileColor } from './colors';
	import { FORMATS, SHAPES, type Format, type IconSource, type Shape } from './types';

	const CATPPUCCIN_COLORS = palette.variants[$user.flavour];

	const DEFAULTS = {
		bg: 'base',
		fg: 'text',
		shape: 'square' as Shape,
		format: 'png' as Format,
		resolution: 512,
		padding: 0.18,
		strokeWidth: 2.0
	};

	let iconName: Writable<string> = writable('camera');
	let iconSource: Writable<IconSource | ''> = writable('');
	let bg: Writable<TileColor> = writable(DEFAULTS.bg);
	let fg: Writable<TileColor> = writable(DEFAULTS.fg);
	let shape: Writable<Shape> = writable(DEFAULTS.shape);
	let format: Writable<Format> = writable(DEFAULTS.format);
	let resolution: Writable<number> = writable(DEFAULTS.resolution);
	let padding: Writable<number> = writable(DEFAULTS.padding);
	let strokeWidth: Writable<number> = writable(DEFAULTS.strokeWidth);

	// Links made before `bg`/`fg` existed stored a mode plus one value per mode
	function legacyColor(params: URLSearchParams, prefix: 'bg' | 'fg'): TileColor | null {
		const mode = params.get(`${prefix}Mode`);
		if (mode === 'transparent') return 'transparent';
		if (mode === 'custom') return params.get(`${prefix}Custom`);
		return params.get(`${prefix}Catppuccin`);
	}

	// URL parameter handling
	$: {
		// This will run on both server and client
		const params = $page.url.searchParams;
		const icon = params.get('icon');
		if (icon) iconName.set(icon);
		const source = params.get('source');
		if (source === 'lucide' || source === 'simpleicons') iconSource.set(source);

		const bgParam = params.get('bg') ?? legacyColor(params, 'bg');
		if (bgParam) bg.set(bgParam);
		const fgParam = params.get('fg') ?? legacyColor(params, 'fg');
		if (fgParam && fgParam !== 'transparent') fg.set(fgParam);

		const shapeParam = params.get('shape');
		if (SHAPES.includes(shapeParam as Shape)) shape.set(shapeParam as Shape);
		const formatParam = params.get('format');
		if (FORMATS.includes(formatParam as Format)) format.set(formatParam as Format);

		const resolutionParam = parseInt(params.get('resolution') ?? '');
		if (!isNaN(resolutionParam)) resolution.set(resolutionParam);
		const paddingParam = parseFloat(params.get('padding') ?? '');
		if (!isNaN(paddingParam)) padding.set(paddingParam);
		const strokeWidthParam = parseFloat(params.get('strokeWidth') ?? '');
		if (!isNaN(strokeWidthParam)) strokeWidth.set(strokeWidthParam);
	}

	// Watch for changes and update URL (only on client side)
	$: if (typeof window !== 'undefined') {
		const params = new URLSearchParams();

		if ($iconName) params.set('icon', $iconName);
		if ($iconSource) params.set('source', $iconSource);
		if ($bg !== DEFAULTS.bg) params.set('bg', $bg);
		if ($fg !== DEFAULTS.fg) params.set('fg', $fg);
		if ($shape !== DEFAULTS.shape) params.set('shape', $shape);
		if ($format !== DEFAULTS.format) params.set('format', $format);
		if ($resolution !== DEFAULTS.resolution) params.set('resolution', $resolution.toString());
		if ($padding !== DEFAULTS.padding) params.set('padding', $padding.toString());
		if ($strokeWidth !== DEFAULTS.strokeWidth) params.set('strokeWidth', $strokeWidth.toString());

		const newUrl = `${window.location.pathname}?${params.toString()}`;
		history.replaceState({}, '', newUrl);
	}

	$: radius = $shape === 'round' ? 0.5 : 0.0;

	// A transparent background stays transparent
	function rollRandomColors() {
		const pair = randomColorPair($bg === 'transparent');
		bg.set(pair.bg);
		fg.set(pair.fg);
	}

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

	async function generateAvatarSvg(iconName: string, source: IconSource | '', options) {
		// Fetch the actual icon SVG
		let iconResult = await fetchIcon(iconName, source);

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
					<svg width="100%" height="100%" viewBox="0 0 24 24" fill="${options.fg}" xmlns="http://www.w3.org/2000/svg">
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
			const { svg } = await generateAvatarSvg($iconName, $iconSource, {
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

	$: resolvedBg = resolveColor($bg, CATPPUCCIN_COLORS);
	$: resolvedFg = resolveColor($fg, CATPPUCCIN_COLORS);

	$: if (
		$iconName ||
		$iconSource ||
		resolvedBg ||
		resolvedFg ||
		$padding ||
		radius ||
		$resolution ||
		$strokeWidth
	) {
		generatePreview();
	}

	async function downloadImage() {
		if (!previewSvg) return;

		const fmt = $format;

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

			<IconSearch {iconName} {iconSource} />
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

			<TileSettings
				palette={CATPPUCCIN_COLORS}
				{bg}
				{fg}
				{shape}
				{format}
				{resolution}
				{padding}
				{strokeWidth}
				iconSource={$iconSource}
			/>
		</div>
	</div>
</section>
