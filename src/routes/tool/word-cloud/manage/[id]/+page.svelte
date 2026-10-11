<script lang="ts">
	import WordCloud from '../../WordCloud.svelte';
	import BarChart from './BarChart.svelte';
	import Table from './Table.svelte';
	import ParticipantPanel from './ParticipantPanel.svelte';
	import Activity from './Activity.svelte';
	import Sentiment from './Sentiment.svelte';
	import { Copy, Lock, LockOpen, Pencil, Presentation, Trash2 } from 'lucide-svelte';
	import MeltTooltip from '$lib/components/MeltTooltip.svelte';
	import type { PageData } from './$types';
	import { onDestroy, onMount } from 'svelte';
	import { user } from '$lib/store';
	import { followSession } from '../../socket';
	import {
		countWords,
		type Settings,
		type WordCloudSessionUser,
		type WordCloudWord
	} from '../../utils';
	import { deleteSession, duplicateSession, updateSession } from '../../actions';
	import Dialog from '../../Dialog.svelte';
	import Card from '../../Card.svelte';
	import SettingsFields from '../../SettingsFields.svelte';
	import Stats from './Stats.svelte';
	import TextInput from '$lib/components/TextInput.svelte';
	import Button from '$lib/components/Button.svelte';
	import { goto } from '$app/navigation';
	import { writable } from 'svelte/store';

	export let data: PageData;

	$: participants = new Map(
		[
			...new Set(
				[...data.session.words]
					.sort((a, b) => a.createdAt.localeCompare(b.createdAt))
					.map((word) => word.uuid)
			)
		].map((uuid, i) => [uuid, i + 1])
	);
	let participant: string | null = null;

	$: visibleWords = data.session.words.filter((word) => !word.hidden);
	$: distribution = countWords(visibleWords.map((word) => word.text));

	let stopFollowing = () => {};
	function applySession(
		session: WordCloudSessionUser,
		words?: WordCloudWord[],
		settings?: Settings
	) {
		data.session = {
			...data.session,
			...session,
			settings: { ...data.session.settings, ...session.settings, ...settings },
			words: words ?? data.session.words,
			closedAt: session.open ? null : data.session.closedAt ?? new Date().toISOString()
		};
	}

	onMount(() => {
		stopFollowing = followSession(
			data.session.id,
			(event) => {
				if (event.type === 'word')
					data.session.words = [...data.session.words, event.word as WordCloudWord];
				else if (event.type === 'session') {
					if (event.session) applySession(event.session, event.words, event.settings);
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
	const editName = writable('');
	const editDescription = writable('');
	let settings: Settings;
	let editError = '';
	function startEditing() {
		editName.set(data.session.name);
		editDescription.set(data.session.description);
		editError = '';
		editing.set(true);
	}
	async function saveEdit(e: Event) {
		e.preventDefault();
		const [name, description] = [$editName.trim(), $editDescription.trim()];
		if (name.length < 3 || name.length > 100) {
			editError = 'The title must be between 3 and 100 characters long';
			return;
		}
		if (description && (description.length < 10 || description.length > 1000)) {
			editError = 'The description must be empty or between 10 and 1000 characters long';
			return;
		}
		const session = await updateSession(data.session.id, { name, description, settings });
		if (session) {
			applySession(session);
			editing.set(false);
		}
	}
	onDestroy(() => stopFollowing());
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

	<div class="grid gap-8 lg:grid-cols-3">
		<div class="flex flex-col gap-4 lg:col-span-3">
			<Stats
				open={data.session.open}
				code={data.session.code}
				words={data.session.words}
				participants={participants.size}
				{distribution}
			/>
			<div class="flex flex-wrap gap-2">
				<Button link="/tool/word-cloud/manage/{data.session.id}/present">
					<Presentation size="16" />
					Present
				</Button>
				<Button on:click={startEditing}>
					<Pencil size="16" />
					Edit
				</Button>
				<MeltTooltip text="New empty session with the same title and description">
					<Button on:click={duplicate}>
						<Copy size="16" />
						Duplicate
					</Button>
				</MeltTooltip>
				{#if data.session.open}
					<Button color="bg-ctp-peach text-ctp-mantle" on:click={() => setOpen(false)}>
						<Lock size="16" />
						Close
					</Button>
				{:else}
					<Button color="bg-ctp-green text-ctp-mantle" on:click={() => setOpen(true)}>
						<LockOpen size="16" />
						Reopen
					</Button>
				{/if}
				<Button color="bg-ctp-red text-ctp-mantle" on:click={() => confirmDelete.set(true)}>
					<Trash2 size="16" />
					Delete
				</Button>
			</div>
		</div>

		<Card class={data.session.settings.sentiment ? 'lg:col-span-2' : 'lg:col-span-3'}>
			<WordCloud title="Word cloud" data={distribution} filename="word-cloud-{data.session.code}" />
		</Card>
		{#if data.session.settings.sentiment}
			<Card title="Mood">
				<Sentiment words={visibleWords} />
			</Card>
		{/if}

		<Card class="lg:col-span-2">
			<Activity words={visibleWords} />
		</Card>
		<Card title="Most frequent words">
			<BarChart data={distribution} />
		</Card>

		<Card title="Submissions" class="lg:col-span-3">
			<Table
				id={data.session.id}
				words={data.session.words}
				{participants}
				on:words={(event) => (data.session.words = event.detail)}
				on:participant={(event) => (participant = event.detail)}
			/>
		</Card>
	</div>
	<ParticipantPanel
		id={data.session.id}
		words={data.session.words}
		{participants}
		bind:participant
		on:words={(event) => (data.session.words = event.detail)}
		on:participant={(event) => (participant = event.detail)}
	/>
</section>

<Dialog open={confirmDelete} title="Delete this session?">
	<p class="mb-6 text-ctp-subtext0">
		“{data.session.name}” and its {data.session.words.length} submission{data.session.words
			.length === 1
			? ''
			: 's'} will be deleted for good.
	</p>
	<div class="flex justify-end gap-2">
		<Button color="bg-ctp-surface0 text-ctp-text" on:click={() => confirmDelete.set(false)}>
			Cancel
		</Button>
		<Button color="bg-ctp-red text-ctp-mantle" on:click={remove}>
			<Trash2 size="16" />
			Delete
		</Button>
	</div>
</Dialog>

<Dialog open={editing} title="Edit the session">
	<form class="flex flex-col gap-4" on:submit={saveEdit}>
		<TextInput label="Question or title" value={editName} />
		<div class="flex flex-col gap-1">
			<label for="edit-description" class="block text-sm font-semibold text-ctp-text">
				Description <small class="text-ctp-subtext0">(optional)</small>
			</label>
			<textarea
				id="edit-description"
				maxlength="1000"
				bind:value={$editDescription}
				class="h-28 w-full rounded-md bg-ctp-surface0 px-3 py-2 shadow-md shadow-ctp-crust focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			/>
		</div>
		<SettingsFields initial={data.session.settings} bind:settings />
		<p class="text-sm font-semibold text-ctp-red" aria-live="polite">{editError}</p>
		<div class="flex justify-end">
			<Button type="submit">
				<Pencil size="16" />
				<span>Save</span>
			</Button>
		</div>
	</form>
</Dialog>
