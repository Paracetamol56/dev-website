<script lang="ts">
	import { Archive, Plus } from 'lucide-svelte';
	import SessionCard from './SessionCard.svelte';
	import type { PageData } from './$types';
	import Button from '$lib/components/Button.svelte';
	import api from '$lib/api';
	import { user } from '$lib/store';
	import type { ManagePageData } from '../utils';

	export let data: PageData;
	let archived: ManagePageData[];

	function loadArchived() {
		api
			.callWithAuth('GET', `/word-cloud?user=${$user.id}&status=closed`)
			.then((res) => {
				archived = res.data as ManagePageData[];
			})
			.catch((err) => {
				console.error(err);
			});
	}
</script>

<svelte:head>
	<title>Your word clouds - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-4xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				Your word clouds
			</span>
		</h1>
		<p class="text-center text-ctp-subtext0 mb-8">Open a session to follow its answers live.</p>
	</hgroup>

	<div class="flex flex-col gap-8">
		<div class="flex flex-col gap-4">
			<div class="flex flex-wrap items-center justify-between gap-4">
				<h2 class="text-2xl font-bold text-ctp-text">Open sessions</h2>
				<Button link="/tool/word-cloud/new">
					<Plus size="18" />
					<span>New session</span>
				</Button>
			</div>
			{#if data.sessions.length === 0}
				<p class="text-ctp-subtext0">You don't have any open session.</p>
			{:else}
				<div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
					{#each data.sessions as session (session.id)}
						<SessionCard {session} open />
					{/each}
				</div>
			{/if}
		</div>

		<div class="flex flex-col gap-4">
			<h2 class="text-2xl font-bold text-ctp-text">Closed sessions</h2>
			{#if archived === undefined}
				<div>
					<Button on:click={loadArchived}>
						<Archive size="18" />
						<span>Show closed sessions</span>
					</Button>
				</div>
			{:else if archived.length === 0}
				<p class="text-ctp-subtext0">You don't have any closed session.</p>
			{:else}
				<div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
					{#each archived as session (session.id)}
						<SessionCard {session} open={false} />
					{/each}
				</div>
			{/if}
		</div>
	</div>
</section>
