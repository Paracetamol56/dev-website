<script lang="ts">
	import Hero from '$lib/components/ui/Hero.svelte';
	import PageDisplay from '$lib/components/PageDisplay.svelte';
	import { page } from '$app/stores';
	import type { PageData } from './$types';
	import { onMount } from 'svelte';

	export let data: PageData;

	const sorted = [...data.pages].sort(
		(a, b) => new Date(b.release).getTime() - new Date(a.release).getTime()
	);
	const counts = new Map<string, number>();
	for (const entry of sorted) {
		for (const name of entry.tags) counts.set(name, (counts.get(name) ?? 0) + 1);
	}
	const ranked = [...counts.entries()].sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]));

	// The page is prerendered, so the query string is only known once mounted
	let mounted = false;
	onMount(() => (mounted = true));
	$: tag = mounted ? $page.url.searchParams.get('tag') : null;
	$: tags = ranked.filter(([name, count]) => count > 1 || name === tag);
	$: pages = tag === null ? sorted : sorted.filter((p) => p.tags.includes(tag as string));
</script>

<svelte:head>
	<title>Pages {tag ? `- #${tag}` : ''} - Mathéo Galuba</title>
</svelte:head>

<Hero
	class="mb-16"
	highlight="Pages"
	description="{pages.length} {pages.length === 1 ? 'page' : 'pages'}{tag ? ` tagged #${tag}` : ''}"
>
	<ul class="mx-auto flex max-w-4xl flex-wrap justify-center gap-2">
		{#each tags as [name, count] (name)}
			<li>
				<a
					class="flex items-center gap-1.5 rounded-full bg-ctp-surface0 px-3 py-1 text-sm font-semibold text-ctp-lavender
								transition-colors hover:bg-ctp-surface1
								aria-[current]:bg-ctp-mauve aria-[current]:text-ctp-base"
					href={tag === name ? '/pages' : `/pages?tag=${name}`}
					aria-current={tag === name ? 'true' : undefined}
					data-sveltekit-noscroll
				>
					#{name}
					<span class="opacity-60">{count}</span>
				</a>
			</li>
		{/each}
	</ul>
</Hero>

<section class="relative container mx-auto mb-32">
	{#if pages.length === 0}
		<p class="my-8 font-semibold text-lg text-center">No page with this tag.</p>
	{:else}
		<div class="grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
			{#each pages as entry (entry.slug)}
				<PageDisplay page={entry} />
			{/each}
		</div>
	{/if}
</section>
