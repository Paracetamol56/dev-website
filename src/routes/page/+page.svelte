<script lang="ts">
  import { Eraser } from 'lucide-svelte';
	import PageDisplay from '../PageDisplay.svelte';
  import type { PageData } from './$types';

  export let data: PageData;
</script>

<svelte:head>
	<title>Pages {data.tag ? `- #${data.tag}` : ''} - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-6xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
        Pages
      </span>
		</h1>
    {#if data.tag}
      <div class="flex flex-col items-center">
        <h2 class="mb-2 text-xl font-semibold text-ctp-lavender">#{data.tag}</h2>
        <a href="/page" class="text-sm font-semibold text-ctp-lavender flex gap-2 items-center"><Eraser size="16" /><span>Clear filter</span></a>
      </div>
    {/if}
	</hgroup>
</section>

<section class="relative container mx-auto mb-32">
	<div class="grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
		{#each data.pages as page}
			<PageDisplay {page} />
		{/each}
	</div>
	<!--<div class="my-4 flex justify-center">
		<Pagination page={pageNumber} count={data.pages.length} perPage={6} />
	</div>-->
</section>