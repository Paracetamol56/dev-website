<script lang="ts">
	import { createPagination, melt } from '@melt-ui/svelte';
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';
	import type { Writable } from 'svelte/store';
	import { TextAlign, TextColor, TextSize, TextWeight } from './typography';

	export let page: Writable<number>;
	export let count: number;
	export let perPage: number = 10;

	const {
		elements: { root, pageTrigger, prevButton, nextButton },
		states: { pages, range },
		options
	} = createPagination({
		page,
		count,
		perPage,
		siblingCount: 1
	});
	$: options.count.set(count);
	$: options.perPage.set(perPage);

	const button = `grid h-8 items-center rounded-md bg-ctp-mantle px-3 shadow-md shadow-ctp-crust
		${TextSize.Sm} ${TextWeight.Semibold} ${TextColor.Muted}
		hover:opacity-75 focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve
		disabled:cursor-not-allowed disabled:opacity-50
		data-[selected]:bg-ctp-mauve data-[selected]:text-ctp-mantle`;
</script>

<nav use:melt={$root} aria-label="pagination" class="flex flex-col items-center gap-4">
	<p class="{TextAlign.Center} {TextSize.Base} {TextColor.Muted}">
		Showing items {$range.start + 1} - {$range.end}
	</p>
	<div class="flex items-center gap-2">
		<button
			use:melt={$prevButton}
			aria-label="Previous page"
			class={button}
			data-umami-event="pagination"
			data-umami-event-type="previous"
		>
			<ChevronLeft class="square-4" />
		</button>
		{#each $pages as item (item.key)}
			{#if item.type === 'ellipsis'}
				<span class={TextColor.Muted}>...</span>
			{:else}
				<button
					use:melt={$pageTrigger(item)}
					class={button}
					data-umami-event="pagination"
					data-umami-event-type="page"
					data-umami-event-properties={`{ "page": ${item.value} }`}
				>
					{item.value}
				</button>
			{/if}
		{/each}
		<button
			use:melt={$nextButton}
			aria-label="Next page"
			class={button}
			data-umami-event="pagination"
			data-umami-event-type="next"
		>
			<ChevronRight class="square-4" />
		</button>
	</div>
</nav>
