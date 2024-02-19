<script lang="ts">
	import Pagination from '$lib/components/Pagination.svelte';
	import { writable, type Writable } from 'svelte/store';
	import type { PageData } from './$types';
	import PageDisplay from './PageDisplay.svelte';

	export let data: PageData;
	const pageNumber: Writable<number> = writable(1);

	let scrollY: number;
	let section0: HTMLElement;
	let section1: HTMLElement;
	let section2: HTMLElement;
</script>

<svelte:window bind:scrollY />

<svelte:head>
	<title>Home - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32" bind:this={section0}>
	<hgroup>
		<h1 class="mb-8 text-6xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender"
				>Welcome</span
			> to my website
		</h1>
	</hgroup>
</section>

<section class="relative container mx-auto mb-32" bind:this={section1}>
	<span
		class="absolute top-0 left-1/2 transform -translate-x-1/2
					text-9xl font-black text-transparent text-outline-2 opacity-20 z-0"
		style="transform: translate(-50%, {(scrollY - section1?.offsetTop + 300) / 5}px);"
	>
		Pages
	</span>
	<div class=" flex flex-col items-center justify-center">
		<h2 class="mb-2 text-4xl font-bold">Pages</h2>
		<p class="text-ctp-subtext0">Pages with various content...</p>
		<hr class="my-8 h-1 w-6 border-none rounded-full bg-ctp-mauve" />
	</div>
	<div class="grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
		{#each data.pages.slice(($pageNumber - 1) * 6, $pageNumber * 6) as page}
			<PageDisplay {page} />
		{/each}
	</div>
	<div class="my-4 flex justify-center">
		<Pagination page={pageNumber} count={data.pages.length} perPage={6} />
	</div>
</section>

<section class="relative container mx-auto" bind:this={section2}>
	<span
		class="absolute top-0 left-1/2 transform -translate-x-1/2
					text-9xl font-black text-transparent text-outline-2 opacity-20 z-0"
		style="transform: translate(-50%, {(scrollY - section2?.offsetTop + 300) / 5}px);"
	>
		Tools
	</span>
	<div class=" flex flex-col items-center justify-center">
		<h2 class="mb-2 text-4xl font-bold">Tools</h2>
		<p class="text-ctp-subtext0">Tools to help you in your daily life...</p>
		<hr class="my-8 h-1 w-6 border-none rounded-full bg-ctp-mauve" />
	</div>
	<div class="grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
		{#each data.tools as tool}
			<PageDisplay page={tool} path="/tool" />
		{/each}
	</div>
</section>

<style lang="postcss">
	.text-outline-2 {
		-webkit-text-stroke-width: 2px;
		-webkit-text-stroke-color: theme('colors.ctp-text.DEFAULT');
	}
</style>
