<script lang="ts">
	import Pagination from '$lib/components/Pagination.svelte';
	import axios from 'axios';
	import { writable, type Writable } from 'svelte/store';

	let loading: boolean = true;
	let data: {
		pages: {
			title: string;
			description: string;
			slug: string;
			categories: string[];
		}[];
		tools: {
			title: string;
			description: string;
			path: string;
		}[];
	} = {
		pages: [],
		tools: [
			{
				title: 'Word cloud',
				description: 'Interractive word cloud to know public opinion about a subject',
				path: 'word-cloud'
			}
		]
	};
	let pageNumber: Writable<number> = writable(1);

	axios
		.get('/api/content')
		.then((res) => {
			data.pages = res.data;
			loading = false;
		})
		.catch((err) => {
			console.error(err);
			loading = false;
		});

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
		{#if loading}
			{#each Array(6) as _}
				<div
					class="p-8 bg-ctp-crust/50 backdrop-blur-sm rounded-md shadow-md shadow-ctp-crust z-10"
				>
					<div class="mb-2 flex justify-left flex-wrap gap-x-2 items-center">
						<div class="text-sm font-semibold bg-ctp-lavender/20 w-1/4 rounded-full">&nbsp;</div>
						<div class="text-sm font-semibold bg-ctp-lavender/20 w-1/4 rounded-full">&nbsp;</div>
					</div>
					<div class="mb-4 text-2xl bg-ctp-text/20 font-bold w-2/3 rounded-full">&nbsp;</div>
					<div class="bg-ctp-subtext0/20 rounded-full">&nbsp;</div>
				</div>
			{/each}
		{:else}
			{#each data.pages as page}
				<div
					class="p-8 bg-ctp-crust/50 backdrop-blur-sm rounded-md shadow-md shadow-ctp-crust z-10"
				>
					<div class="mb-2 flex justify-left flex-wrap gap-x-2 items-center">
						{#each page.categories as category}
							<a href="/">
								<span class="text-sm font-semibold text-ctp-lavender">#{category}</span>
							</a>
						{/each}
					</div>
					<a href="/page/{page.slug}">
						<h4 class="mb-4 text-2xl font-bold hover:opacity-75 transition-opacity">
							{page.title}
						</h4>
						<p class="text-ctp-subtext0">{page.description}</p>
					</a>
				</div>
			{/each}
		{/if}
	</div>
	<!--<div class="my-4 flex justify-center">
		<Pagination page={pageNumber} count={data.pages.length} perPage={6} />
	</div>-->
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
			<div class="p-8 bg-ctp-crust/50 backdrop-blur-sm rounded-md shadow-md shadow-ctp-crust z-10">
				<div class="mb-2 flex justify-left flex-wrap gap-x-2 items-center">
					<!-- {#each page.categories as category}
						<a href="/">
							<span class="text-sm font-semibold text-ctp-lavender">#{category}</span>
						</a>
					{/each} -->
				</div>
				<a href={tool.path}>
					<h4 class="mb-4 text-2xl font-bold hover:opacity-75 transition-opacity">{tool.title}</h4>
					<p class="text-ctp-subtext0">{tool.description}</p>
				</a>
			</div>
		{/each}
	</div>
</section>

<style lang="postcss">
	.text-outline-2 {
		-webkit-text-stroke-width: 2px;
		-webkit-text-stroke-color: theme('colors.ctp-text.DEFAULT');
	}
</style>
