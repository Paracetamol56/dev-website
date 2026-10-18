<script lang="ts">
	import { goto } from '$app/navigation';
	import { Copy, Eye, LogOut, MoreHorizontal, PenLine, Trash2 } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { writable } from 'svelte/store';
	import Button from '$lib/components/ui/Button.svelte';
	import Dialog from '$lib/components/ui/Dialog.svelte';
	import DropdownMenu from '$lib/components/ui/DropdownMenu.svelte';
	import { Tone } from '$lib/components/ui/styles';
	import { user } from '$lib/store';
	import { deleteDeck, duplicateDeck, removeMember } from './actions';
	import type { DeckSummary } from './types';

	export let deck: DeckSummary;

	const dispatch = createEventDispatcher<{ change: void }>();
	const confirmDelete = writable(false);
	const confirmLeave = writable(false);
	let busy = false;

	const date = (value: string) =>
		new Date(value).toLocaleDateString(undefined, {
			day: 'numeric',
			month: 'short',
			year: 'numeric'
		});

	async function duplicate() {
		if (await duplicateDeck(deck.id)) dispatch('change');
	}

	async function remove() {
		busy = true;
		const deleted = await deleteDeck(deck.id);
		busy = false;
		if (deleted) {
			confirmDelete.set(false);
			dispatch('change');
		}
	}

	async function leave() {
		if (!$user.id) return;
		busy = true;
		const left = await removeMember(deck.id, $user.id);
		busy = false;
		if (left) {
			confirmLeave.set(false);
			dispatch('change');
		}
	}

	const open = () => goto(`/tools/slides/${deck.id}`);
	$: opening = deck.outline[0]?.title;
	$: items =
		deck.role === 'owner'
			? [
					{ label: 'Edit', icon: PenLine, onSelect: open },
					{ label: 'Duplicate', icon: Copy, onSelect: duplicate },
					{
						label: 'Delete',
						icon: Trash2,
						tone: Tone.Danger,
						onSelect: () => confirmDelete.set(true)
					}
			  ]
			: [
					deck.role === 'editor'
						? { label: 'Edit', icon: PenLine, onSelect: open }
						: { label: 'View', icon: Eye, onSelect: open },
					...(deck.role === 'editor'
						? [{ label: 'Duplicate', icon: Copy, onSelect: duplicate }]
						: []),
					{
						label: 'Leave',
						icon: LogOut,
						tone: Tone.Danger,
						onSelect: () => confirmLeave.set(true)
					}
			  ];
</script>

<article
	class="group relative flex h-full flex-col overflow-hidden rounded-md bg-ctp-mantle shadow-md shadow-ctp-crust"
>
	<a
		href="/tools/slides/{deck.id}"
		class="flex aspect-video flex-col items-center justify-center gap-2 overflow-hidden bg-ctp-crust p-6 text-center
			focus:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ctp-mauve"
	>
		<span
			class="line-clamp-3 text-2xl font-bold text-ctp-text transition-colors group-hover:text-ctp-mauve"
		>
			{deck.title}
		</span>
		{#if opening && opening !== deck.title}
			<span class="line-clamp-1 text-sm text-ctp-subtext0">{opening}</span>
		{/if}
	</a>
	<div class="flex items-center gap-2 p-4">
		<p class="min-w-0 flex-1 truncate text-sm text-ctp-subtext0">
			{#if deck.role !== 'owner'}
				<span
					class="mr-1 rounded-full bg-ctp-surface0 px-2 py-0.5 text-xs font-semibold text-ctp-subtext1"
				>
					{deck.role === 'editor' ? 'Can edit' : 'Can view'}
				</span>
				{deck.ownerName ?? 'Someone'} ·
			{/if}
			{deck.outline.length} slide{deck.outline.length === 1 ? '' : 's'} · Edited {date(
				deck.updatedAt
			)}
		</p>
		<DropdownMenu label="Actions for {deck.title}" {items} class="p-1 hover:bg-ctp-surface0">
			<MoreHorizontal size="18" />
		</DropdownMenu>
	</div>
</article>

<Dialog open={confirmDelete} title="Delete this deck?">
	<p class="mb-6 text-ctp-subtext0">“{deck.title}” will be deleted for good.</p>
	<div class="flex justify-end gap-2">
		<Button tone={Tone.Neutral} on:click={() => confirmDelete.set(false)}>Cancel</Button>
		<Button tone={Tone.Danger} loading={busy} on:click={remove}>
			<Trash2 size="16" />
			Delete
		</Button>
	</div>
</Dialog>

<Dialog open={confirmLeave} title="Leave this deck?">
	<p class="mb-6 text-ctp-subtext0">
		“{deck.title}” will leave your list. You will need a new link from its owner to come back.
	</p>
	<div class="flex justify-end gap-2">
		<Button tone={Tone.Neutral} on:click={() => confirmLeave.set(false)}>Cancel</Button>
		<Button tone={Tone.Danger} loading={busy} on:click={leave}>
			<LogOut size="16" />
			Leave
		</Button>
	</div>
</Dialog>
