<script lang="ts">
	import { goto } from '$app/navigation';
	import { createDropdownMenu, melt } from '@melt-ui/svelte';
	import { Copy, Lock, LockOpen, MoreHorizontal, Trash2 } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { writable } from 'svelte/store';
	import { fly } from 'svelte/transition';
	import Dialog from '../Dialog.svelte';
	import { deleteSession, duplicateSession, updateSession } from '../actions';
	import type { ManagePageData } from '../utils';

	export let session: ManagePageData;

	const dispatch = createEventDispatcher<{ change: void }>();
	const {
		elements: { trigger, menu, item },
		states: { open: menuOpen }
	} = createDropdownMenu({ forceVisible: true, positioning: { placement: 'bottom-end' } });

	const date = (value: string | null) => (value ? new Date(value).toLocaleDateString() : '');

	async function setOpen(open: boolean) {
		if (await updateSession(session.id, { open })) dispatch('change');
	}
	async function duplicate() {
		const id = await duplicateSession(session.id);
		if (id) goto(`/tool/word-cloud/manage/${id}`);
	}
	const confirmDelete = writable(false);
	async function remove() {
		confirmDelete.set(false);
		if (await deleteSession(session.id)) dispatch('change');
	}

	const menuItem =
		'flex cursor-pointer items-center gap-2 rounded-md px-3 py-1.5 text-sm data-[highlighted]:bg-ctp-mauve/25 data-[highlighted]:text-ctp-mauve';
</script>

<div class="relative flex flex-col gap-3 rounded-md bg-ctp-mantle p-6 shadow-md shadow-ctp-crust">
	<div class="flex items-center gap-2 pr-8 text-sm text-ctp-subtext0">
		<span class="square-2 rounded-full {session.open ? 'bg-ctp-green' : 'bg-ctp-red'}" />
		<span>{session.open ? 'Open' : 'Closed'}</span>
		<span class="font-mono font-semibold text-ctp-mauve">{session.code}</span>
	</div>
	<a
		href="/tool/word-cloud/manage/{session.id}"
		class="text-xl font-bold text-ctp-text hover:text-ctp-mauve focus:outline-none focus-visible:underline"
	>
		{session.name}
	</a>
	{#if session.description}
		<p class="line-clamp-3 text-ctp-subtext0">{session.description}</p>
	{/if}
	<p class="mt-auto text-xs text-ctp-subtext0">
		{session.submissions} submission{session.submissions === 1 ? '' : 's'} · created
		{date(session.createdAt)}{session.closedAt ? ` · closed ${date(session.closedAt)}` : ''}
	</p>

	<button
		use:melt={$trigger}
		aria-label="Actions for {session.name}"
		class="absolute right-4 top-4 grid place-items-center rounded-md p-1 hover:bg-ctp-surface0"
	>
		<MoreHorizontal size="18" />
	</button>
	{#if $menuOpen}
		<div
			use:melt={$menu}
			class="z-20 flex min-w-[10rem] flex-col rounded-md bg-ctp-surface0 p-1 shadow-md shadow-ctp-crust"
			transition:fly={{ duration: 150, y: -4 }}
		>
			<div use:melt={$item} class={menuItem} on:m-click={duplicate}>
				<Copy size="14" /> Duplicate
			</div>
			{#if session.open}
				<div use:melt={$item} class={menuItem} on:m-click={() => setOpen(false)}>
					<Lock size="14" /> Close
				</div>
			{:else}
				<div use:melt={$item} class={menuItem} on:m-click={() => setOpen(true)}>
					<LockOpen size="14" /> Reopen
				</div>
			{/if}
			<div
				use:melt={$item}
				class="{menuItem} text-ctp-red"
				on:m-click={() => confirmDelete.set(true)}
			>
				<Trash2 size="14" /> Delete
			</div>
		</div>
	{/if}
</div>

<Dialog open={confirmDelete} title="Delete this session?">
	<p class="mb-6 text-ctp-subtext0">
		“{session.name}” and its {session.submissions} submission{session.submissions === 1 ? '' : 's'} will
		be deleted for good.
	</p>
	<div class="flex justify-end gap-2">
		<button
			class="rounded-md bg-ctp-surface0 px-3 py-1 font-semibold hover:bg-ctp-surface1"
			on:click={() => confirmDelete.set(false)}>Cancel</button
		>
		<button
			class="flex items-center gap-1 rounded-md bg-ctp-red px-3 py-1 font-semibold text-ctp-mantle hover:opacity-80"
			on:click={remove}
		>
			<Trash2 size="16" />
			Delete
		</button>
	</div>
</Dialog>
