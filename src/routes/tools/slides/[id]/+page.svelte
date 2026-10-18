<script lang="ts">
	import { beforeNavigate, goto } from '$app/navigation';
	import {
		AlertCircle,
		ArrowLeft,
		Check,
		Code,
		Loader2,
		Pencil,
		Presentation,
		SlidersHorizontal,
		WifiOff
	} from 'lucide-svelte';
	import { onDestroy, onMount, tick } from 'svelte';
	import { writable } from 'svelte/store';
	import Slides, { type SlideEdit } from '$lib/components/deck/Slides.svelte';
	import type { DeckSettings } from '$lib/components/deck/settings';
	import Button from '$lib/components/ui/Button.svelte';
	import ResizeHandle from '$lib/components/ui/ResizeHandle.svelte';
	import { Tone } from '$lib/components/ui/styles';
	import { addToast } from '$lib/toast';
	import TitleDialog from '../TitleDialog.svelte';
	import { notifyError } from '../actions';
	import type { Indices } from '../types';
	import CodeEditor from './CodeEditor.svelte';
	import SettingsPanel from './SettingsPanel.svelte';
	import { DeckSocket } from './socket';
	import { peerColour, type Peer } from './presence';
	import type { PageData } from './$types';

	export let data: PageData;

	let title = data.deck.title;
	let settings: DeckSettings = data.deck.settings;
	let settingsTimer: ReturnType<typeof setTimeout> | undefined;
	let preview = '';
	let editingSlide = false;
	let previewTimer: ReturnType<typeof setTimeout>;
	let pending = false;
	let peers: Peer[] = [];
	let panel: 'code' | 'settings' = 'code';
	let editor: CodeEditor;
	let slides: Slides;

	const current = writable<Indices>({ h: 0, v: 0 });

	const layoutKey = 'slides:layout';
	const defaultCode = 0.5;
	const codeWidth = writable(defaultCode);
	let resizing = false;

	const socket = new DeckSocket(data.deck.id, () => editor?.syncedVersion() ?? data.deck.revision);
	const connection = socket.connection;

	socket.on({
		state: (message) => {
			title = message.title;
			if (settingsTimer) sendSettings();
			else settings = message.settings;
		},
		title: (message) => (title = message.title),
		settings: (message) => {
			if (!settingsTimer) settings = message.settings;
		},
		reset: (message) => {
			if (editor.hasPending()) {
				notifyError('Your last changes could not be kept, the deck was reloaded from the server');
			}
			editor.reset(message.doc, message.version);
		},
		deleted: () => {
			addToast({
				data: { title: 'Deck deleted', description: 'This deck was deleted', color: 'bg-ctp-red' }
			});
			goto('/tools/slides');
		},
		error: (message) => notifyError(message.error)
	});

	// The preview keeps its html while a text is edited in it, so its ids stay valid
	function refreshPreview() {
		clearTimeout(previewTimer);
		if (!editingSlide && editor) preview = editor.annotate();
	}

	function onChange() {
		clearTimeout(previewTimer);
		previewTimer = setTimeout(refreshPreview, 300);
	}

	async function onSelect(event: CustomEvent<number>) {
		panel = 'code';
		await tick();
		editor.reveal(event.detail);
	}

	function onEdit(event: CustomEvent<SlideEdit>) {
		editor.replace(event.detail);
	}

	function onEditing(event: CustomEvent<boolean>) {
		editingSlide = event.detail;
		if (!editingSlide) refreshPreview();
	}

	function sendSettings() {
		if (socket.send({ type: 'settings', settings })) {
			clearTimeout(settingsTimer);
			settingsTimer = undefined;
		}
	}

	function onSettings(event: CustomEvent<DeckSettings>) {
		settings = event.detail;
		clearTimeout(settingsTimer);
		settingsTimer = setTimeout(sendSettings, 300);
	}

	const renaming = writable(false);
	function rename(event: CustomEvent<string>) {
		if (socket.send({ type: 'title', title: event.detail })) renaming.set(false);
		else notifyError('The deck is offline, try again once it is connected');
	}

	$: unsaved = pending || settingsTimer !== undefined;

	beforeNavigate((navigation) => {
		if (!unsaved) return;
		if (navigation.willUnload) navigation.cancel();
		else if (!confirm('Your last changes are not saved yet. Leave anyway?')) navigation.cancel();
	});

	function readLayout(): { code?: unknown } | null {
		try {
			return JSON.parse(localStorage.getItem(layoutKey) ?? 'null');
		} catch {
			return null;
		}
	}

	function writeLayout() {
		try {
			localStorage.setItem(layoutKey, JSON.stringify({ code: $codeWidth }));
			return true;
		} catch {
			return false;
		}
	}

	onMount(() => {
		const saved = readLayout();
		if (typeof saved?.code === 'number') codeWidth.set(Math.min(0.8, Math.max(0.2, saved.code)));
		const unsubscribe = codeWidth.subscribe(writeLayout);
		refreshPreview();
		socket.connect();
		return unsubscribe;
	});

	onDestroy(() => {
		socket.close();
		clearTimeout(previewTimer);
		clearTimeout(settingsTimer);
	});
</script>

<svelte:head>
	<title>{title} - Slides - Mathéo Galuba</title>
</svelte:head>

<div class="flex flex-col gap-3 bg-ctp-mantle p-3 lg:h-[calc(100dvh-4.25rem)]">
	<header
		class="flex flex-wrap items-center gap-x-3 gap-y-2 rounded-md bg-ctp-base px-3 py-2 shadow-md shadow-ctp-crust"
	>
		<a
			href="/tools/slides"
			class="grid place-items-center rounded-md p-1.5 text-ctp-subtext0 hover:bg-ctp-surface0 hover:text-ctp-text"
			aria-label="Back to your decks"
		>
			<ArrowLeft size="18" />
		</a>
		<button
			type="button"
			class="group flex min-w-0 items-center gap-2 rounded-md px-2 py-1 hover:bg-ctp-surface0"
			on:click={() => renaming.set(true)}
		>
			<h1 class="truncate text-lg font-bold text-ctp-text">{title}</h1>
			<Pencil size="14" class="shrink-0 text-ctp-subtext0 opacity-0 group-hover:opacity-100" />
			<span class="sr-only">Rename</span>
		</button>

		<p class="flex items-center gap-1.5 text-sm text-ctp-subtext0" role="status">
			{#if $connection === 'refused'}
				<AlertCircle size="14" class="text-ctp-red" /> Disconnected
			{:else if $connection === 'closed'}
				<WifiOff size="14" class="text-ctp-yellow" /> Offline, reconnecting…
			{:else if $connection === 'connecting'}
				<Loader2 size="14" class="animate-spin" /> Connecting…
			{:else if unsaved}
				<Loader2 size="14" class="animate-spin" /> Saving…
			{:else}
				<Check size="14" class="text-ctp-green" /> Saved
			{/if}
		</p>

		<div class="ml-auto flex items-center gap-3">
			{#if peers.length}
				<ul class="flex -space-x-1.5" aria-label="Also editing">
					{#each peers as peer (peer.peer)}
						<li
							class="grid h-7 w-7 place-items-center rounded-full border-2 border-ctp-base text-xs font-bold text-ctp-crust"
							style="background-color: rgb(var(--ctp-{peerColour(peer.peer)}))"
							title="{peer.name} is editing"
						>
							{peer.name.slice(0, 1).toUpperCase()}
						</li>
					{/each}
				</ul>
			{/if}
			<Button tone={Tone.Accent} on:click={() => slides.present()}>
				<Presentation size="18" />
				Present
			</Button>
		</div>
	</header>

	<div
		class="grid min-h-0 flex-1 grid-cols-1 gap-3 lg:gap-0
			lg:grid-cols-[calc((100%_-_0.75rem)_*_var(--code))_0.75rem_minmax(0,1fr)]"
		style:--code={$codeWidth}
	>
		<section
			class="flex h-[60vh] flex-col overflow-hidden rounded-md bg-ctp-base shadow-md shadow-ctp-crust lg:h-auto"
		>
			<div role="tablist" aria-label="Editor" class="flex gap-1 border-b border-ctp-surface0 p-1.5">
				{#each [{ id: 'code', label: 'Code', icon: Code }, { id: 'settings', label: 'Settings', icon: SlidersHorizontal }] as tab}
					<button
						type="button"
						role="tab"
						aria-selected={panel === tab.id}
						class="flex items-center gap-1.5 rounded-md px-3 py-1 text-sm font-semibold text-ctp-subtext0
							hover:bg-ctp-surface0 hover:text-ctp-text aria-selected:bg-ctp-surface0 aria-selected:text-ctp-mauve"
						on:click={() => (panel = tab.id === 'code' ? 'code' : 'settings')}
					>
						<svelte:component this={tab.icon} size="14" />
						{tab.label}
					</button>
				{/each}
			</div>
			<div role="tabpanel" aria-label="Code" class="min-h-0 flex-1" class:hidden={panel !== 'code'}>
				<CodeEditor
					bind:this={editor}
					value={data.deck.content}
					version={data.deck.revision}
					{socket}
					on:change={onChange}
					on:pending={(event) => (pending = event.detail)}
					on:peers={(event) => (peers = event.detail)}
					on:cursor={(event) => slides.show(event.detail)}
				/>
			</div>
			<div
				role="tabpanel"
				aria-label="Settings"
				class="min-h-0 flex-1 overflow-y-auto"
				class:hidden={panel !== 'settings'}
			>
				<SettingsPanel {settings} on:change={onSettings} />
			</div>
		</section>
		<ResizeHandle
			value={codeWidth}
			min={0.2}
			max={0.8}
			defaultValue={defaultCode}
			label="Resize the code and the preview"
			class="hidden lg:flex"
			on:dragging={(event) => (resizing = event.detail)}
		/>
		<section
			aria-label="Preview"
			class="aspect-video overflow-hidden rounded-md bg-ctp-base shadow-md shadow-ctp-crust lg:aspect-auto"
			class:pointer-events-none={resizing}
		>
			<Slides
				bind:this={slides}
				content={preview}
				{settings}
				slide={current}
				on:select={onSelect}
				on:edit={onEdit}
				on:editing={onEditing}
			/>
		</section>
	</div>
</div>

<TitleDialog
	open={renaming}
	heading="Rename the deck"
	confirm="Rename"
	initial={title}
	on:submit={rename}
/>
