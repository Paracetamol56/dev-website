<script lang="ts">
	import { onMount } from 'svelte';
	import { get, writable, type Writable } from 'svelte/store';
	import { frameDocument, loadHighlightTheme, loadTheme } from './frameDocument';
	import { defaultSettings, type DeckSettings } from './settings';

	type Indices = { h: number; v: number; f?: number };

	export let content: string;
	export let slide: Writable<Indices> = writable({ h: 0, v: 0 });
	export let settings: DeckSettings = defaultSettings;
	export let title = 'Slides preview';

	let frame: HTMLIFrameElement;
	let ready = false;

	function post(message: unknown) {
		frame?.contentWindow?.postMessage(message, '*');
	}

	let applied = 0;
	async function configure(next: DeckSettings) {
		const request = ++applied;
		const [theme, highlightCss] = await Promise.all([
			loadTheme(next.theme),
			loadHighlightTheme(next.highlightTheme)
		]);
		if (request === applied) post({ type: 'settings', settings: next, theme, highlightCss });
	}

	function render(html: string) {
		post({ type: 'render', html, indices: get(slide) });
	}

	export function present() {
		frame.requestFullscreen?.();
		frame.focus();
	}

	onMount(() => {
		const receive = (event: MessageEvent) => {
			if (event.source !== frame.contentWindow) return;
			if (event.data?.type === 'ready') {
				ready = true;
			} else if (event.data?.type === 'slide') {
				const { h, v, f } = event.data.indices as Indices;
				const current = get(slide);
				if (current.h !== h || current.v !== v || current.f !== f) slide.set({ h, v, f });
			}
		};
		addEventListener('message', receive);
		return () => removeEventListener('message', receive);
	});

	$: if (ready) configure(settings).catch(console.error);
	$: if (ready) render(content);
	$: if (ready) post({ type: 'slide', indices: $slide });
</script>

<!-- The deck runs in its own document with an opaque origin: neither its styles nor its scripts reach the page -->
<iframe
	bind:this={frame}
	{title}
	srcdoc={frameDocument}
	sandbox="allow-scripts"
	allow="fullscreen"
	class="block h-full w-full border-0"
/>
