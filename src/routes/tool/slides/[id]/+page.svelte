<script lang="ts">
	import { basicSetup } from 'codemirror';
	import { indentWithTab } from '@codemirror/commands';
	import { EditorView, keymap } from '@codemirror/view';
	import { ChangeSet, EditorState, type Extension, StateEffect } from '@codemirror/state';
	import { collab, receiveUpdates, type Update } from '@codemirror/collab';
	import { html } from '@codemirror/lang-html';
	import { melt, createScrollArea, createToolbar } from '@melt-ui/svelte';
	import type { PageData } from './$types';
	import Slides from '$lib/components/deck/Slides.svelte';
	import { user } from '$lib/store';
	import { variants } from '@catppuccin/palette';
	import Button from '$lib/components/Button.svelte';
	import { Presentation } from 'lucide-svelte';
	import { onDestroy, onMount } from 'svelte';
	import { DebouncedOutlineGenerator } from './outlineGenerator';
	import SlidePreview from './SlidePreview.svelte';
	import { type Writable, writable } from 'svelte/store';
	import peerExtension from './peerExtension';

	export let data: PageData;
	let value: string = data.slideDeck?.content || '<!-- Your slides go here -->';
	let editorView: EditorView;
	let currentSlide: Writable<{ h: number; v: number; f?: number }> = writable({ h: 0, v: 0 });

	const ws = new WebSocket(`ws://localhost:8080/api/slides/${data.slideDeck?.id}/ws`);
	ws.onopen = () => {
		console.log('Connected to WebSocket');
	};
	ws.onclose = () => {
		console.log('Disconnected from WebSocket');
	};

	// Dynamically change CodeMirror theme based on the current theme
	$: loadTheme($user.flavour);

	async function loadTheme(flavor: keyof typeof variants) {
		try {
			let newTheme: any;
			const themeModule = await import('@catppuccin/codemirror');
			switch (flavor) {
				case 'latte':
					newTheme = themeModule.catppuccinLatte;
					break;
				case 'frappe':
					newTheme = themeModule.catppuccinFrappe;
					break;
				case 'macchiato':
					newTheme = themeModule.catppuccinMacchiato;
					break;
				case 'mocha':
				default:
					newTheme = themeModule.catppuccinMocha;
					break;
			}

			if (!editorView) return;
			editorView.dispatch({
				effects: StateEffect.reconfigure.of(getExtensions(ws, newTheme))
			});
		} catch (error) {
			console.error('Failed to load theme:', error);
		}
	}

	function getExtensions(ws: WebSocket, theme?: Extension): Extension[] {
		const ext: Extension[] = [
			basicSetup,
			EditorView.lineWrapping,
			keymap.of([indentWithTab]),
			html(),
			peerExtension(data.slideDeck.revision, ws),
		];

		if (theme) ext.push(theme);

		return ext;
	}

	/* const outlineGenerator = new DebouncedOutlineGenerator(300);
	$: outlineGenerator.execute(value);
	outlineGenerator.store?.subscribe((value) => {
		if (value) {
			console.log('Outline generated:', value);
		}
	});
  */

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

	onMount(() => {
		editorView = new EditorView({
			parent: document.getElementById('editor')!,
			state: EditorState.create({
				doc: value,
				extensions: getExtensions(ws)
			}),
			dispatch(tr) {
				editorView.update([tr]);
			}
		});
	});

	onDestroy(() => {
		// outlineGenerator.cleanup();
    ws.close();
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
				<div id="editor">
					<!-- CodeMirror editor will be mounted here -->
				</div>
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
		<!-- {#if outlineGenerator.store !== null}
			<SlidePreview outline={outlineGenerator.store} current={currentSlide} />
		{/if} -->
	</div>
</div>

<style lang="postcss">
	:global(body) {
		@apply bg-ctp-mantle !important;
	}
</style>
