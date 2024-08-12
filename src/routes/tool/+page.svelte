<script lang="ts">
	import { Eraser } from 'lucide-svelte';
	import PageDisplay from '../PageDisplay.svelte';
	import type { PageData } from './$types';
	import { Tool } from '$lib/page';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	export let data: PageData;
	let tag: string | null = null;
	let tools: Tool[] = [];
	onMount(() => {
		tag = $page.url.searchParams.get('tag');
		tools = data.tools;
		if (tag !== null) {
			tools = tools.filter((p) => p.tags.includes(tag as string));
		}
	});
</script>

<svelte:head>
	<title>Tools {tag ? `- #${tag}` : ''} - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-6xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				Tools
			</span>
		</h1>
		{#if tag}
			<div class="flex flex-col items-center">
				<h2 class="mb-2 text-xl font-semibold text-ctp-lavender">#{tag}</h2>
				<a href="/tool" class="text-sm font-semibold text-ctp-lavender flex gap-2 items-center">
					<Eraser size="16" />
					<span>Clear filter</span>
				</a>
			</div>
		{/if}
	</hgroup>
</section>

<section class="relative container mx-auto mb-32">
	<div class="grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
		{#each tools as tool}
			<PageDisplay page={tool} path="/tool" />
		{/each}
	</div>
	<!--<div class="my-4 flex justify-center">
		<Pagination page={pageNumber} count={data.pages.length} perPage={6} />
	</div>-->
</section>
