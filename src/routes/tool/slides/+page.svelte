<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import { Plus } from 'lucide-svelte';
	import type { PageData } from './$types';
	import MeltPagination from '$lib/components/MeltPagination.svelte';
	import { writable } from 'svelte/store';
	import { goto } from '$app/navigation';

	export let data: PageData;

	const currentPage = writable(1);
	currentPage.subscribe((value) => {
		goto(`/tool/slides?page=${value}`, { replaceState: true });
	});
</script>

<svelte:head>
	<title>Slides - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-6xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				Slides
			</span>
		</h1>
	</hgroup>
</section>

<section class="container mx-auto">
	<div class="mb-8 flex justify-left gap-4">
		<h2 class="text-2xl font-bold text-ctp-lavender">Recent decks</h2>
		<Button link="/tool/codecarbon-dashboard/project/new">
			<span>New deck</span>
			<Plus size="18" />
		</Button>
	</div>
	{#if data.slides.total_items === 0}
		<p>You don't have any slide deck.</p>
	{:else}
		<div class="mb-8 grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
			{#each data.slides.items as slide (slide.id)}
				<div class="p-8 bg-ctp-mantle rounded-md shadow-md shadow-ctp-crust">
					<div class="flex gap-2 justify-end items-baseline mb-4">
						<a href="/tool/slides/{slide.id}" class="flex-1">
							<h4 class="text-2xl font-bold hover:opacity-75 transition-opacity">
								{slide.title}
							</h4>
						</a>
					</div>
				</div>
			{/each}
		</div>
		<MeltPagination page={currentPage} count={data.slides.total_items} perPage={12} />
	{/if}
</section>
