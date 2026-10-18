<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { Plus, Presentation } from 'lucide-svelte';
	import { writable } from 'svelte/store';
	import Button from '$lib/components/ui/Button.svelte';
	import Hero from '$lib/components/ui/Hero.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import TextInput from '$lib/components/ui/TextInput.svelte';
	import DeckCard from './DeckCard.svelte';
	import TitleDialog from './TitleDialog.svelte';
	import { createDeck } from './actions';
	import { perPage } from './types';
	import type { PageData } from './$types';

	export let data: PageData;

	const creating = writable(false);
	let busy = false;

	async function create(event: CustomEvent<string>) {
		busy = true;
		const deck = await createDeck(event.detail);
		busy = false;
		if (deck) {
			creating.set(false);
			goto(`/tools/slides/${deck.id}`);
		}
	}

	function navigate(page: number, search: string) {
		const query = new URLSearchParams();
		if (page > 1) query.set('page', String(page));
		if (search) query.set('search', search);
		goto(`/tools/slides${query.size ? `?${query}` : ''}`, {
			keepFocus: true,
			noScroll: true,
			replaceState: true
		});
	}

	const search = writable(data.search);
	let timer: ReturnType<typeof setTimeout>;
	function onSearch() {
		clearTimeout(timer);
		timer = setTimeout(() => navigate(1, $search.trim()), 300);
	}

	const page = writable(data.decks.page);
	$: page.set(data.decks.page);
	$: if ($page !== data.decks.page) navigate($page, data.search);
</script>

<svelte:head>
	<title>Slides - Mathéo Galuba</title>
</svelte:head>

<Hero
	class="mb-16"
	highlight="Slides"
	description="Write reveal.js presentations in HTML and see them come to life as you type."
>
	<div class="flex justify-center">
		<Button on:click={() => creating.set(true)}>
			<Plus size="18" />
			New deck
		</Button>
	</div>
</Hero>

<section class="container mx-auto mb-32">
	<div class="mb-8 flex flex-wrap items-end justify-between gap-4">
		<h2 class="text-2xl font-bold text-ctp-text">
			Your decks
			<span class="text-base font-semibold text-ctp-subtext0">{data.decks.total}</span>
		</h2>
		<TextInput
			type="search"
			label="Search"
			hideLabel
			placeholder="Search by title"
			value={search}
			class="w-64 max-w-full"
			on:input={onSearch}
		/>
	</div>

	{#if data.decks.items.length === 0}
		<div class="flex flex-col items-center gap-4 rounded-md bg-ctp-mantle p-12 text-center">
			<Presentation size="32" class="text-ctp-mauve" />
			{#if data.search}
				<p class="text-ctp-subtext0">No deck matches “{data.search}”.</p>
			{:else}
				<p class="text-ctp-subtext0">You don't have any slide deck yet.</p>
				<Button on:click={() => creating.set(true)}>
					<Plus size="18" />
					Create your first deck
				</Button>
			{/if}
		</div>
	{:else}
		<div class="mb-12 grid grid-cols-1 gap-8 sm:grid-cols-2 xl:grid-cols-3">
			{#each data.decks.items as deck (deck.id)}
				<DeckCard {deck} on:change={() => invalidateAll()} />
			{/each}
		</div>
		{#if data.decks.total > perPage}
			<Pagination {page} count={data.decks.total} {perPage} />
		{/if}
	{/if}
</section>

<TitleDialog
	open={creating}
	heading="New deck"
	confirm="Create"
	loading={busy}
	on:submit={create}
/>
