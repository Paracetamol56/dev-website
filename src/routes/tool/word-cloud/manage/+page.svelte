<script lang="ts">
	import { Archive, Plus } from 'lucide-svelte';
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
				archived = res.data;
			})
			.catch((err) => {
				console.error(err);
			});
	}
</script>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-6xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				Your sessions
			</span>
		</h1>
	</hgroup>
</section>

<section class="container mx-auto">
	<div class="mb-8 flex justify-left gap-4">
		<h2 class="text-2xl font-bold text-ctp-lavender">Open sessions</h2>
		<Button link="/tool/word-cloud/new">
			<span>New session</span>
			<Plus size="18" />
		</Button>
	</div>

	<div class="mb-8 grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
		{#each data.sessions as session}
			<div class="p-8 bg-ctp-mantle rounded-md shadow-md shadow-ctp-crust">
				<div class="flex justify-start items-baseline mb-4">
					<div class="relative square-2 mr-2 bg-ctp-green rounded-full">
						<span
							class="animate-ping absolute top-0.5 right-0.5 block square-1 rounded-full ring-2 ring-ctp-green bg-ctp-green"
						/>
					</div>
					<p class="text-ctp-subtext0 text-sm">Open</p>
					<p class="ml-auto text-ctp-subtext0 text-sm">
						{session.submitions} submition{session.submitions == 0 ? '' : 's'}
					</p>
				</div>
				<a href="/tool/word-cloud/manage/{session.id}">
					<h4 class="mb-4 text-2xl font-bold hover:opacity-75 transition-opacity">
						{session.name}
					</h4>
				</a>
				<p class="text-ctp-subtext0">{session.description}</p>
			</div>
		{/each}
	</div>

	<div class="mb-8">
		<h2 class="text-2xl font-bold text-ctp-lavender">Closed sessions</h2>
	</div>
	{#if archived === undefined}
		<Button on:click={loadArchived}>
			<span>Load closed sessions</span>
			<Archive size="18" />
		</Button>
	{:else if archived.length > 0}
		<div class="mb-8 grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
			{#each archived as session}
				<div class="p-8 bg-ctp-mantle rounded-md shadow-md shadow-ctp-crust">
					<div class="flex justify-start items-baseline mb-4">
						<div class="relative square-2 mr-2 bg-ctp-red rounded-full" />
						<p class="text-ctp-subtext0 text-sm">Closed</p>
						<p class="ml-auto text-ctp-subtext0 text-sm">
							{session.submitions} submition{session.submitions == 0 ? '' : 's'}
						</p>
					</div>
					<a href="/tool/word-cloud/manage/{session.id}">
						<h4 class="mb-4 text-2xl font-bold hover:opacity-75 transition-opacity">
							{session.name}
						</h4>
					</a>
					<p class="text-ctp-subtext0">{session.description}</p>
				</div>
			{/each}
		</div>
	{:else}
		<div class="mb-8 grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
			<p>You</p>
		</div>
	{/if}
</section>
