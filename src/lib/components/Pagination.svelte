<script lang="ts">
	import { createPagination, melt } from '@melt-ui/svelte';
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';
	import type { Writable } from 'svelte/store';

	export let page: Writable<number>;
	export let count;
	export let perPage = 10;

	const {
		elements: { root, pageTrigger, prevButton, nextButton },
		states: { pages, range }
	} = createPagination({
		page,
		count: count,
		perPage: perPage,
		siblingCount: 1
	});
</script>

<nav class="flex flex-col items-center gap-4" aria-label="pagination" use:melt={$root}>
	<p class="text-center text-ctp-subtext0">
		Showing items {$range.start} - {$range.end}
	</p>
	<div class="flex items-center gap-2">
		<button
			class="grid h-8 items-center rounded-md bg-ctp-mantle px-3 text-sm text-ctp-subtext0 shadow-md shadow-ctp-crust
      hover:opacity-75 disabled:cursor-not-allowed disabled:opacity-50 data-[selected]:bg-ctp-mauve
      data-[selected]:text-ctp-text"
			use:melt={$prevButton}
			data-umami-event="pagination"
			data-umami-event-type="previous"
		>
			<ChevronLeft class="square-4" />
		</button>
		{#each $pages as page (page.key)}
			{#if page.type === 'ellipsis'}
				<span>...</span>
			{:else}
				<button
					class="grid h-8 items-center rounded-md bg-ctp-mantle px-3 text-sm font-semibold text-ctp-subtext0 shadow-md shadow-ctp-crust
          hover:opacity-75 disabled:cursor-not-allowed disabled:opacity-50 data-[selected]:bg-ctp-mauve
          data-[selected]:text-ctp-base"
					use:melt={$pageTrigger(page)}
					data-umami-event="pagination"
					data-umami-event-type="page"
					data-umami-event-properties={`{ "page": ${page.value} }`}
				>
					{page.value}
				</button>
			{/if}
		{/each}
		<button
			class="grid h-8 items-center rounded-md bg-ctp-mantle px-3 text-sm text-ctp-subtext0 shadow-md shadow-ctp-crust
      hover:opacity-75 disabled:cursor-not-allowed disabled:opacity-50 data-[selected]:bg-ctp-mauve
      data-[selected]:text-ctp-text"
			use:melt={$nextButton}
			data-umami-event="pagination"
			data-umami-event-type="next"
		>
			<ChevronRight class="square-4" />
		</button>
	</div>
</nav>
