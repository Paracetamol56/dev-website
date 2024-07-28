<script lang="ts">
	import { page } from '$app/stores';  
	import type { PageData } from './$types';

	export let data: PageData;
</script>

<svelte:head>
	<title>{data.meta.title}</title>
	<meta property="og:type" content="article" />
	<meta property="og:title" content={data.meta.title} />
	<meta property="og:description" content={data.meta.description} />
	<meta property="og:url" content={$page.url.toString()} />
	<meta property="og:image" content={$page.url.toString() + '/og.png'} />
</svelte:head>

<article>
	<section class="container mx-auto mb-32">
		<hgroup>
			<h1 class="mb-8 text-6xl font-bold text-center">
				<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
					{data.meta.title}
				</span>
			</h1>
		</hgroup>

		<div class="flex gap-2 justify-center">
			{#each data.meta.tags as tag}
				<a href="/page?tag={tag}" class="font-semibold text-ctp-lavender">
					&num;{tag}
				</a>
			{/each}
		</div>
	</section>

	<div class="container mx-auto content">
		<svelte:component this={data.content} />
	</div>
</article>
