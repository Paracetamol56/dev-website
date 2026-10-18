<script context="module" lang="ts">
	export type SlideEdit = {
		id: number;
		html: string | null;
		align: string | null;
		tag: string | null;
	};
	// Source ids of the slide on screen and of its top-level section
	export type CurrentSlide = { slide: number | null; stack: number | null };
</script>

<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { get, writable, type Writable } from 'svelte/store';
	import { frameDocument, loadHighlightTheme, loadTheme } from './frameDocument';
	import { defaultSettings, type DeckSettings } from './settings';
	import SlideToolbar, { type FormatCommand, type TextFormat } from './SlideToolbar.svelte';

	type Indices = { h: number; v: number; f?: number };
	type Rect = { top: number; left: number; bottom: number; right: number };

	export let content: string;
	export let slide: Writable<Indices> = writable({ h: 0, v: 0 });
	export let settings: DeckSettings = defaultSettings;
	export let title = 'Slides preview';
	export let readonly = false;

	const dispatch = createEventDispatcher<{ select: number; edit: SlideEdit; editing: boolean }>();

	let frame: HTMLIFrameElement;
	let ready = false;
	let currentSlide: CurrentSlide = { slide: null, stack: null };
	let editing: { id: number; rect: Rect; format: TextFormat } | null = null;
	let toolbar: HTMLDivElement;
	let width = 0;
	let toolbarWidth = 0;
	let toolbarHeight = 0;

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

	export function show(id: number | null) {
		post({ type: 'show', id });
	}

	export function current(): CurrentSlide {
		return currentSlide;
	}

	export function startEditing(id: number) {
		post({ type: 'start', id });
	}

	export function commit() {
		if (editing) post({ type: 'commit' });
	}

	function onFormat(event: CustomEvent<FormatCommand>) {
		post({ type: 'format', ...event.detail });
	}

	function onPointerDown(event: PointerEvent) {
		if (editing && !toolbar?.contains(event.target as Node)) commit();
	}

	function onFullscreen() {
		post({ type: 'present', on: document.fullscreenElement === frame });
	}

	$: toolbarTop = editing
		? editing.rect.top - toolbarHeight - 10 < 4
			? editing.rect.bottom + 10
			: editing.rect.top - toolbarHeight - 10
		: 0;
	$: toolbarLeft = editing ? Math.max(4, Math.min(editing.rect.left, width - toolbarWidth - 4)) : 0;

	onMount(() => {
		const receive = (event: MessageEvent) => {
			if (event.source !== frame.contentWindow) return;
			if (event.data?.type === 'ready') {
				ready = true;
			} else if (event.data?.type === 'select') {
				dispatch('select', event.data.id);
			} else if (event.data?.type === 'edit') {
				const { id, html, align, tag } = event.data as SlideEdit;
				dispatch('edit', { id, html, align, tag });
			} else if (event.data?.type === 'editing') {
				const wasEditing = editing !== null;
				editing = event.data.id === null ? null : event.data;
				if (wasEditing !== (editing !== null)) dispatch('editing', editing !== null);
			} else if (event.data?.type === 'slide') {
				currentSlide = { slide: event.data.slide ?? null, stack: event.data.stack ?? null };
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
	$: if (ready) post({ type: 'readonly', on: readonly });
</script>

<svelte:window on:pointerdown={onPointerDown} />
<svelte:document on:fullscreenchange={onFullscreen} />

<div class="relative h-full w-full" bind:clientWidth={width}>
	<!-- The deck runs in its own document with an opaque origin: neither its styles nor its scripts reach the page -->
	<iframe
		bind:this={frame}
		{title}
		srcdoc={frameDocument}
		sandbox="allow-scripts"
		allow="fullscreen"
		class="block h-full w-full border-0"
	/>
	{#if editing}
		<div
			bind:this={toolbar}
			bind:offsetWidth={toolbarWidth}
			bind:offsetHeight={toolbarHeight}
			class="absolute z-10"
			style:top="{toolbarTop}px"
			style:left="{toolbarLeft}px"
		>
			<SlideToolbar format={editing.format} on:format={onFormat} />
		</div>
	{/if}
</div>
