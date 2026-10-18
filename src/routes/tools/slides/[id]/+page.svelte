<script lang="ts">
	import { beforeNavigate, goto } from '$app/navigation';
	import { Code, ListTree, PanelLeftClose, SlidersHorizontal } from 'lucide-svelte';
	import { onDestroy, onMount, tick } from 'svelte';
	import { writable } from 'svelte/store';
	import Slides, { type SlideEdit } from '$lib/components/deck/Slides.svelte';
	import type { DeckSettings } from '$lib/components/deck/settings';
	import ResizeHandle from '$lib/components/ui/ResizeHandle.svelte';
	import { addToast } from '$lib/toast';
	import TitleDialog from '../TitleDialog.svelte';
	import { notifyError } from '../actions';
	import type { Indices } from '../types';
	import CodeEditor from './CodeEditor.svelte';
	import DeckHeader from './DeckHeader.svelte';
	import type { BlockKind } from './blocks';
	import OutlinePanel from './OutlinePanel.svelte';
	import SettingsPanel from './SettingsPanel.svelte';
	import { buildOutline, type OutlineSlide } from './outline';
	import { DeckSocket } from './socket';
	import type { Peer } from './presence';
	import type { PageData } from './$types';
	import Button from '$lib/components/ui/Button.svelte';

	export let data: PageData;

	let title = data.deck.title;
	let settings: DeckSettings = data.deck.settings;
	let settingsTimer: ReturnType<typeof setTimeout> | undefined;
	let preview = '';
	let editingSlide = false;
	let outline: OutlineSlide[] = buildOutline(data.deck.content);
	let previewTimer: ReturnType<typeof setTimeout>;
	let pending = false;
	let peers: Peer[] = [];
	type Panel = 'code' | 'outline' | 'settings';
	let panel: Panel = 'code';
	const tabs = [
		{ id: 'code', label: 'Code', icon: Code },
		{ id: 'outline', label: 'Outline', icon: ListTree },
		{ id: 'settings', label: 'Settings', icon: SlidersHorizontal }
	] as const;

	function openPanel(id: Panel) {
		panel = id;
		panelOpen.set(true);
	}
	let editor: CodeEditor;
	let slides: Slides;

	const current = writable<Indices>({ h: 0, v: 0 });

	const layoutKey = 'slides:layout';
	const defaultCode = 0.5;
	const codeWidth = writable(defaultCode);
	const panelOpen = writable(true);
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
		if (!editor) return;
		outline = buildOutline(editor.text());
		if (!editingSlide) preview = editor.annotate();
	}

	function onOutline(event: CustomEvent<{ indices: Indices; from: number }>) {
		current.set(event.detail.indices);
		editor.moveTo(event.detail.from);
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

	async function onInsert(event: CustomEvent<BlockKind | 'image'>) {
		if (event.detail === 'image') {
			addToast({
				data: {
					title: 'Coming soon',
					description: 'Adding images is not available yet',
					color: 'bg-ctp-blue'
				}
			});
			return;
		}
		slides.commit();
		const focus = editor.insert(event.detail, slides.current());
		refreshPreview();
		if (focus === null) {
			openPanel('code');
			return;
		}
		const id = editor.elementAt(focus);
		await tick();
		if (id !== null) slides.startEditing(id);
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

	function readLayout(): { code?: unknown; open?: unknown } | null {
		try {
			return JSON.parse(localStorage.getItem(layoutKey) ?? 'null');
		} catch {
			return null;
		}
	}

	function writeLayout() {
		try {
			localStorage.setItem(layoutKey, JSON.stringify({ code: $codeWidth, open: $panelOpen }));
			return true;
		} catch {
			return false;
		}
	}

	onMount(() => {
		const saved = readLayout();
		if (typeof saved?.code === 'number') codeWidth.set(Math.min(0.8, Math.max(0.2, saved.code)));
		if (typeof saved?.open === 'boolean') panelOpen.set(saved.open);
		const unsubscribe = [codeWidth.subscribe(writeLayout), panelOpen.subscribe(writeLayout)];
		refreshPreview();
		socket.connect();
		return () => unsubscribe.forEach((stop) => stop());
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

<div class="flex flex-col gap-3 bg-ctp-mantle p-3 lg:h-dvh">
	<DeckHeader
		{title}
		{connection}
		{unsaved}
		{peers}
		on:rename={() => renaming.set(true)}
		on:present={() => slides.present()}
		on:insert={onInsert}
		on:undo={() => editor.undo()}
		on:redo={() => editor.redo()}
	/>

	<div
		class="grid min-h-0 flex-1 grid-cols-1 gap-3 lg:gap-0
			{$panelOpen
			? 'lg:grid-cols-[calc((100%_-_0.75rem)_*_var(--code))_0.75rem_minmax(0,1fr)]'
			: 'lg:grid-cols-[auto_minmax(0,1fr)] lg:gap-3'}"
		style:--code={$codeWidth}
	>
		{#if !$panelOpen}
			<nav
				aria-label="Editor panels"
				class="hidden flex-col gap-1 rounded-md bg-ctp-base p-1 shadow-md shadow-ctp-crust lg:flex"
			>
				{#each tabs as tab}
					<button
						type="button"
						class="flex flex-col items-center gap-2 rounded-md px-1.5 py-3 text-sm font-semibold text-ctp-subtext0
							hover:bg-ctp-surface0 hover:text-ctp-mauve focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve"
						title="Show {tab.label.toLowerCase()}"
						on:click={() => openPanel(tab.id)}
					>
						<svelte:component this={tab.icon} size="14" />
						<span class="rotate-180 [writing-mode:vertical-rl]">{tab.label}</span>
					</button>
				{/each}
			</nav>
		{/if}
		<section
			class="flex h-[60vh] flex-col overflow-hidden rounded-md bg-ctp-base shadow-md shadow-ctp-crust lg:h-auto"
			class:lg:hidden={!$panelOpen}
		>
			<div class="flex items-center gap-1 border-b border-ctp-surface0 p-1.5">
				<button
					type="button"
					class="hidden place-items-center rounded-md p-1.5 text-ctp-subtext0 hover:bg-ctp-surface0 hover:text-ctp-text lg:grid"
					aria-label="Hide the code panel"
					title="Hide the code panel"
					on:click={() => panelOpen.set(false)}
				>
					<PanelLeftClose size="14" />
				</button>
				<div role="tablist" aria-label="Editor" class="flex gap-1">
					{#each tabs as tab}
						<button
							type="button"
							role="tab"
							aria-selected={panel === tab.id}
							class="flex items-center gap-1.5 rounded-md px-3 py-1 text-sm font-semibold text-ctp-subtext0
							hover:bg-ctp-surface0 hover:text-ctp-text aria-selected:bg-ctp-surface0 aria-selected:text-ctp-mauve"
							on:click={() => (panel = tab.id)}
						>
							<svelte:component this={tab.icon} size="14" />
							{tab.label}
						</button>
					{/each}
				</div>
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
				aria-label="Outline"
				class="min-h-0 flex-1 overflow-y-auto"
				class:hidden={panel !== 'outline'}
			>
				<OutlinePanel {outline} {current} on:select={onOutline} />
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
			class="hidden {$panelOpen ? 'lg:flex' : ''}"
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
