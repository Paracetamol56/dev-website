<script lang="ts">
	import type { Page } from '$lib/page';

	export let page: Page;
	export let path: string = "/page";
</script>

<div class="p-8 bg-ctp-crust/50 backdrop-blur-sm rounded-md shadow-md shadow-ctp-crust z-10">
	<div class="mb-2 flex justify-left flex-wrap gap-x-2 items-center">
		{#each page.tags.slice(0, 4) as tag}
			<a href="{path}?tag={tag}">
				<span class="text-sm font-semibold text-ctp-lavender">#{tag}</span>
			</a>
		{/each}
		{#if page.tags.length > 4}
			<span class="text-sm font-semibold text-ctp-lavender">...</span>
		{/if}
	</div>
	<a href="{path}/{page.slug}">
		<h4 class="mb-4 text-2xl font-bold hover:opacity-75 transition-opacity">
			{page.title}
		</h4>
		<p class="text-ctp-subtext0">{page.description}</p>
	</a>
</div>
