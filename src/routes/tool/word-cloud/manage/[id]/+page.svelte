<script lang="ts">
	import WordCloud from '../../WordCloud.svelte';
	import BarChart from './BarChart.svelte';
	import Table from './Table.svelte';
	import { createDialog, melt } from '@melt-ui/svelte';
	import { Copy, Lock, LockOpen, Pencil, QrCode, Trash2, X } from 'lucide-svelte';
	import MeltTooltip from '$lib/components/MeltTooltip.svelte';
	import { fade, fly } from 'svelte/transition';
	import QRCode from 'qrcode';
	import type { PageData } from './$types';
	import { onDestroy, onMount } from 'svelte';
	import { user } from '$lib/store';
	import { followSession } from '../../socket';
	import type { WordCloudSessionUser, WordCloudWord } from '../../utils';
	import { deleteSession, duplicateSession, updateSession } from '../../actions';
	import Dialog from '../../Dialog.svelte';
	import Button from '$lib/components/Button.svelte';
	import { goto } from '$app/navigation';
	import { writable } from 'svelte/store';

	export let data: PageData;

	function addWord(word: WordCloudWord) {
		data.session.words = [...data.session.words, word];
		const found = data.distribution.find((entry) => entry.text === word.text);
		if (found) found.occurence++;
		else data.distribution.push({ text: word.text, occurence: 1 });
		data.distribution = data.distribution.sort((a, b) => b.occurence - a.occurence);
	}

	let stopFollowing = () => {};
	function applySession(session: WordCloudSessionUser) {
		data.session = {
			...data.session,
			...session,
			closedAt: session.open ? null : data.session.closedAt ?? new Date()
		};
	}

	onMount(() => {
		stopFollowing = followSession(
			data.session.id,
			(event) => {
				if (event.type === 'word') addWord(event.word);
				else if (event.type === 'session') {
					if (event.session) applySession(event.session);
					else goto('/tool/word-cloud/manage');
				}
			},
			$user.accessToken
		);
	});

	async function setOpen(open: boolean) {
		const session = await updateSession(data.session.id, { open });
		if (session) applySession(session);
	}

	async function duplicate() {
		const id = await duplicateSession(data.session.id);
		if (id) goto(`/tool/word-cloud/manage/${id}`);
	}

	const confirmDelete = writable(false);
	async function remove() {
		if (await deleteSession(data.session.id)) goto('/tool/word-cloud/manage');
	}

	const editing = writable(false);
	let editName = '';
	let editDescription = '';
	let editError = '';
	function startEditing() {
		editName = data.session.name;
		editDescription = data.session.description;
		editError = '';
		editing.set(true);
	}
	async function saveEdit(e: Event) {
		e.preventDefault();
		const [name, description] = [editName.trim(), editDescription.trim()];
		if (name.length < 3 || name.length > 100) {
			editError = 'The title must be between 3 and 100 characters long';
			return;
		}
		if (description && (description.length < 10 || description.length > 1000)) {
			editError = 'The description must be empty or between 10 and 1000 characters long';
			return;
		}
		const session = await updateSession(data.session.id, { name, description });
		if (session) {
			applySession(session);
			editing.set(false);
		}
	}
	onDestroy(() => stopFollowing());

	let qrWidthAvailable: number;
	let canvas: HTMLCanvasElement;

	$: {
		if (canvas !== undefined) {
			const url = `${window.location.origin}/tool/word-cloud?code=${data.session.code}`;
			QRCode.toCanvas(
				canvas,
				url,
				{
					errorCorrectionLevel: 'H',
					scale: 3 + qrWidthAvailable / 175
				},
				(error) => {
					if (error) console.error(error);
				}
			);
		}
	}

	const humanReadableDate = (date: Date | null) => {
		if (date === null) return '';
		return date.toLocaleString();
	};

	const {
		elements: { trigger, overlay, content, title, close, portalled },
		states: { open }
	} = createDialog({
		forceVisible: true
	});
</script>

<svelte:head>
	<title>{data.session.name} - Word cloud - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-4xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				{data.session.name}
			</span>
		</h1>
		{#if data.session.description}
			<p class="text-center text-ctp-subtext0 mb-8">{data.session.description}</p>
		{/if}
	</hgroup>

	<div class="flex flex-col gap-8">
		<div
			class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-wrap items-center gap-x-8 gap-y-4"
		>
			<div class="flex items-center gap-2">
				{#if data.session.open}
					<span class="relative square-2 rounded-full bg-ctp-green">
						<span class="absolute inset-0 animate-ping rounded-full bg-ctp-green" />
					</span>
					<span class="font-semibold">Open</span>
				{:else}
					<span class="square-2 rounded-full bg-ctp-red" />
					<span class="font-semibold">Closed</span>
				{/if}
			</div>
			<dl class="flex flex-wrap gap-x-8 gap-y-2 text-sm">
				<div>
					<dt class="text-ctp-subtext0">Code</dt>
					<dd class="font-mono text-lg font-bold text-ctp-mauve">{data.session.code}</dd>
				</div>
				<div>
					<dt class="text-ctp-subtext0">Submissions</dt>
					<dd class="text-lg font-semibold">{data.session.words.length}</dd>
				</div>
				<div>
					<dt class="text-ctp-subtext0">Unique words</dt>
					<dd class="text-lg font-semibold">{data.distribution.length}</dd>
				</div>
				<div>
					<dt class="text-ctp-subtext0">Created</dt>
					<dd class="text-lg font-semibold">
						{humanReadableDate(new Date(data.session.createdAt))}
					</dd>
				</div>
				{#if !data.session.open && data.session.closedAt}
					<div>
						<dt class="text-ctp-subtext0">Closed</dt>
						<dd class="text-lg font-semibold">
							{humanReadableDate(new Date(data.session.closedAt))}
						</dd>
					</div>
				{/if}
			</dl>
			<div class="ml-auto flex flex-wrap gap-2">
				{#if data.session.open}
					<MeltTooltip text="Show the code and QR code to the audience">
						<button
							class="flex items-center gap-1 rounded-md px-3 py-1 font-semibold text-ctp-mantle shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60 bg-ctp-mauve"
							use:melt={$trigger}
						>
							<QrCode size="16" />
							Join info
						</button>
					</MeltTooltip>
				{/if}
				<button
					class="flex items-center gap-1 rounded-md px-3 py-1 font-semibold text-ctp-mantle shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60 bg-ctp-mauve"
					on:click={startEditing}
				>
					<Pencil size="16" />
					Edit
				</button>
				<MeltTooltip text="New empty session with the same title and description">
					<button
						class="flex items-center gap-1 rounded-md px-3 py-1 font-semibold text-ctp-mantle shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60 bg-ctp-mauve"
						on:click={duplicate}
					>
						<Copy size="16" />
						Duplicate
					</button>
				</MeltTooltip>
				{#if data.session.open}
					<button
						class="flex items-center gap-1 rounded-md px-3 py-1 font-semibold text-ctp-mantle shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60 bg-ctp-peach"
						on:click={() => setOpen(false)}
					>
						<Lock size="16" />
						Close
					</button>
				{:else}
					<button
						class="flex items-center gap-1 rounded-md px-3 py-1 font-semibold text-ctp-mantle shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60 bg-ctp-green"
						on:click={() => setOpen(true)}
					>
						<LockOpen size="16" />
						Reopen
					</button>
				{/if}
				<button
					class="flex items-center gap-1 rounded-md px-3 py-1 font-semibold text-ctp-mantle shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60 bg-ctp-red"
					on:click={() => confirmDelete.set(true)}
				>
					<Trash2 size="16" />
					Delete
				</button>
			</div>
		</div>

		<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-col gap-4">
			<h2 class="text-2xl font-bold text-ctp-text">Word cloud</h2>
			<WordCloud data={data.distribution} filename="word-cloud-{data.session.code}" />
		</div>

		<div class="bg-ctp-mantle p-6 rounded-md shadow-md shadow-ctp-crust flex flex-col gap-4">
			<h2 class="text-2xl font-bold text-ctp-text">Most frequent words</h2>
			<BarChart data={data.distribution} />
		</div>

		<Table data={data.session.words} id={data.session.id} />
	</div>
</section>

<div use:melt={$portalled}>
	{#if $open}
		<div
			use:melt={$overlay}
			class="fixed inset-0 z-30 bg-black/50"
			transition:fade={{ duration: 200 }}
		/>
		<div
			class="fixed left-[50%] top-[50%] z-50 h-[90vh] w-[90vw]
            translate-x-[-50%] translate-y-[-50%] rounded-md bg-ctp-base
            p-6 shadow-md overflow-auto"
			transition:fly={{ duration: 200, y: 10 }}
			use:melt={$content}
		>
			<h2 use:melt={$title} class="mt-4 mb-8 text-center text-4xl font-bold">Join the session !</h2>
			<div class="w-full flex flex-col items-center" bind:clientWidth={qrWidthAvailable}>
				<canvas bind:this={canvas} />
				<a
					class="my-8 text-3xl font-medium"
					href="{window.location.origin}/tool/word-cloud?code={data.session.code}"
					target="_blank"
				>
					{window.location.origin}/tool/word-cloud
				</a>
				<p class="my-8 text-4xl font-bold">
					Code: <span class="text-ctp-mauve">{data.session.code}</span>
				</p>
			</div>

			<button
				use:melt={$close}
				aria-label="close"
				class="absolute right-4 top-4 inline-flex h-6 w-6 appearance-none
                items-center justify-center rounded-full p-1 text-base
                hover:bg-ctp-mauve hover:text-ctp-base transition-colors"
			>
				<X class="square-4" />
			</button>
		</div>
	{/if}
</div>

<Dialog open={confirmDelete} title="Delete this session?">
	<p class="mb-6 text-ctp-subtext0">
		“{data.session.name}” and its {data.session.words.length} submission{data.session.words
			.length === 1
			? ''
			: 's'} will be deleted for good.
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

<Dialog open={editing} title="Edit the session">
	<form class="flex flex-col gap-4" on:submit={saveEdit}>
		<div>
			<label for="edit-name" class="mb-2 block text-sm font-semibold">Question or title</label>
			<input
				id="edit-name"
				maxlength="100"
				bind:value={editName}
				class="h-8 w-full rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			/>
		</div>
		<div>
			<label for="edit-description" class="mb-2 block text-sm font-semibold">
				Description <small class="text-ctp-subtext0">(optional)</small>
			</label>
			<textarea
				id="edit-description"
				maxlength="1000"
				bind:value={editDescription}
				class="h-28 w-full rounded-md bg-ctp-surface0 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			/>
		</div>
		<p class="text-sm font-semibold text-ctp-red" aria-live="polite">{editError}</p>
		<div class="flex justify-end">
			<Button type="submit">
				<Pencil size="16" />
				<span>Save</span>
			</Button>
		</div>
	</form>
</Dialog>
