<script lang="ts">
	import Pagination from '$lib/components/Pagination.svelte';
	import { writable, type Writable } from 'svelte/store';
	import type { PageData } from './$types';
	import PageDisplay from './PageDisplay.svelte';
	import { ArrowRight } from 'lucide-svelte';

	export let data: PageData;
	const pageNumber: Writable<number> = writable(1);
	const TOPICS = 8;

	let scrollY: number;
	let section1: HTMLElement;
	let section2: HTMLElement;

	const drift = (section: HTMLElement | undefined, scroll: number) =>
		section ? Math.max(-30, Math.min(30, (scroll - section.offsetTop + 300) / 5)) : 0;

	const counts = new Map<string, number>();
	for (const page of data.pages) {
		for (const tag of page.tags) counts.set(tag, (counts.get(tag) ?? 0) + 1);
	}
	const topics = [...counts.entries()]
		.sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
		.slice(0, TOPICS)
		.map(([tag]) => tag);
</script>

<svelte:window bind:scrollY />

<svelte:head>
	<title>Home - Mathéo Galuba</title>
</svelte:head>

<section class="relative isolate container mx-auto mb-32 pt-8 text-center">
	<div
		class="pointer-events-none absolute left-1/2 top-0 -z-10 h-72 w-[min(48rem,100%)] -translate-x-1/2
					rounded-full bg-gradient-to-r from-ctp-mauve/25 to-ctp-blue/20 blur-3xl"
	/>
	<h1 class="mb-6 text-5xl font-bold sm:text-6xl lg:text-7xl">
		<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender"
			>Welcome</span
		> to my website
	</h1>
	<p class="mx-auto mb-8 max-w-2xl text-lg text-ctp-subtext0">
		Write-ups, experiments and small tools from my side projects because I don't know what to do
		with them.
	</p>
	{#if topics.length > 0}
		<ul class="mx-auto flex max-w-2xl flex-wrap justify-center gap-2">
			{#each topics as tag}
				<li>
					<a
						class="block rounded-full bg-ctp-surface0 px-3 py-1 text-sm font-semibold text-ctp-lavender
									transition-colors hover:bg-ctp-surface1"
						href="/page?tag={tag}"
					>
						#{tag}
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</section>

<section class="relative container mx-auto mb-32" bind:this={section1}>
	<span
		class="pointer-events-none absolute top-0 left-1/2
					text-8xl sm:text-9xl font-black text-transparent text-outline-2 opacity-20 z-0"
		style="transform: translate(-50%, {drift(section1, scrollY)}px);"
	>
		Pages
	</span>
	<div class="relative flex flex-col items-center justify-center">
		<h2 class="mb-2 text-4xl font-bold">Pages</h2>
		<p class="text-ctp-subtext0">Pages with various content...</p>
		<hr class="my-8 h-1 w-6 border-none rounded-full bg-ctp-mauve" />
	</div>
	{#if data.pages.length === 0}
		<div class="my-8 font-semibold text-lg text-center">
			<p>Coming soon...</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
			{#each data.pages.slice(($pageNumber - 1) * 6, $pageNumber * 6) as page (page.slug)}
				<PageDisplay {page} />
			{/each}
		</div>
		<div class="mt-8 flex flex-col items-center gap-6">
			<Pagination page={pageNumber} count={data.pages.length} perPage={6} />
			<a
				class="flex items-center gap-2 font-semibold text-ctp-mauve hover:opacity-75 transition-opacity"
				href="/page"
			>
				All {data.pages.length} pages
				<ArrowRight size="18" />
			</a>
		</div>
	{/if}
</section>

<section class="relative container mx-auto" bind:this={section2}>
	<span
		class="pointer-events-none absolute top-0 left-1/2
					text-8xl sm:text-9xl font-black text-transparent text-outline-2 opacity-20 z-0"
		style="transform: translate(-50%, {drift(section2, scrollY)}px);"
	>
		Tools
	</span>
	<div class="relative flex flex-col items-center justify-center">
		<h2 class="mb-2 text-4xl font-bold">Tools</h2>
		<p class="text-ctp-subtext0">Tools to help you in your daily life...</p>
		<hr class="my-8 h-1 w-6 border-none rounded-full bg-ctp-mauve" />
	</div>
	{#if data.tools.length === 0}
		<div class="my-8 font-semibold text-lg text-center">
			<p>Coming soon...</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
			{#each data.tools as tool (tool.slug)}
				<PageDisplay page={tool} path="/tool" requiresAuth={tool.requiresAuth} />
			{/each}
		</div>
	{/if}
</section>

<style lang="postcss">
	.text-outline-2 {
		-webkit-text-stroke-width: 2px;
		-webkit-text-stroke-color: theme('colors.ctp-text.DEFAULT');
	}
</style>
