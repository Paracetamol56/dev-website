<script lang="ts">
	import CodeMirror from 'svelte-codemirror-editor';
	import { html } from '@codemirror/lang-html';
	import { melt, createScrollArea, createToolbar } from '@melt-ui/svelte';
	import type { PageData } from './$types';
	import Slides from '$lib/components/deck/Slides.svelte';
	import { user } from '$lib/store';
	import { variants } from '@catppuccin/palette';
	import Button from '$lib/components/Button.svelte';
	import { Presentation } from 'lucide-svelte';
	import { onDestroy } from 'svelte';
	import { DebouncedOutlineGenerator } from './outlineGenerator';
	import SlidePreview from './SlidePreview.svelte';
	import { type Writable, writable } from 'svelte/store';

	export let data: PageData;
	let value: string = data.deck?.html || '<!-- Your slides go here -->';
	let currentSlide: Writable<{ h: number; v: number; f?: number }> = writable({ h: 0, v: 0 });

	// Dynamically change CodeMirror theme based on the current theme
	let currentTheme: any = null;
	$: {
		loadTheme($user.flavour);
	}

	async function loadTheme(flavor: keyof typeof variants) {
		try {
			let themeModule;

			switch (flavor) {
				case 'latte':
					themeModule = await import('@catppuccin/codemirror');
					currentTheme = themeModule.catppuccinLatte;
					break;
				case 'frappe':
					themeModule = await import('@catppuccin/codemirror');
					currentTheme = themeModule.catppuccinFrappe;
					break;
				case 'macchiato':
					themeModule = await import('@catppuccin/codemirror');
					currentTheme = themeModule.catppuccinMacchiato;
					break;
				case 'mocha':
				default:
					themeModule = await import('@catppuccin/codemirror');
					currentTheme = themeModule.catppuccinMocha;
					break;
			}
		} catch (error) {
			console.error('Failed to load theme:', error);
			// Fallback to mocha theme
			const fallbackModule = await import('@catppuccin/codemirror');
			currentTheme = fallbackModule.catppuccinMocha;
		}
	}

	// Get the AST from the editor
	const outlineGenerator = new DebouncedOutlineGenerator(2000);
	$: outlineGenerator.execute(value);
	outlineGenerator.store?.subscribe((value) => {
		if (value) {
			console.log('Outline generated:', value);
		}
	});

	const {
		elements: { root, content, viewport, corner, scrollbarY, thumbY, thumbX, scrollbarX }
	} = createScrollArea({
		type: 'auto',
		dir: 'ltr'
	});

	const {
		elements: { root: toolbarRoot, button, link, separator },
		builders: { createToolbarGroup }
	} = createToolbar();

	onDestroy(() => {
		outlineGenerator.cleanup();
	});
</script>

<div
	class="h-[calc(100dvh-4.25rem)] max-h-screen grid grid-cols-2 grid-rows-[auto_1fr_auto] gap-3 p-3"
>
	<div class="bg-ctp-base text-ctp-text rounded-md shadow-md shadow-ctp-crust bg-ctp-crust/50 p-4">
		<Button disabled><Presentation size="18" /><span>Present</span></Button>
	</div>
	<div
		use:melt={$root}
		class="col-start-1 row-start-2 h-full overflow-hidden bg-ctp-base rounded-md shadow-md shadow-ctp-crust"
	>
		<div use:melt={$viewport} class="h-full">
			<div use:melt={$content} class="h-full">
				<CodeMirror bind:value lang={html()} theme={currentTheme} class="h-full" />
			</div>
		</div>
		<div
			use:melt={$scrollbarY}
			class="flex h-full w-3.5 touch-none select-none border-l border-l-transparent p-1 transition-colors"
		>
			<div use:melt={$thumbY} class="relative flex-1 rounded-full bg-ctp-surface1" />
		</div>
		<div
			use:melt={$scrollbarX}
			class="flex h-3.5 w-full touch-none select-none border-t border-t-transparent p-1"
		>
			<div use:melt={$thumbX} class="relative rounded-full bg-ctp-surface1" />
		</div>
		<div use:melt={$corner} />
	</div>
	<div
		class="row-span-2 col-start-2 row-start-1 rounded-md overflow-hidden bg-ctp-base shadow-md shadow-ctp-crust"
	>
		<Slides slide={currentSlide}>
			{@html value}
		</Slides>
	</div>
	<div class="col-span-2 row-start-3 rounded-md shadow-md shadow-ctp-crust bg-ctp-crust/50">
		{#if outlineGenerator.store !== null}
			<SlidePreview outline={outlineGenerator.store} current={currentSlide} />
		{/if}
	</div>
</div>

<style lang="postcss">
	:global(body) {
		@apply bg-ctp-mantle !important;
	}
</style>
