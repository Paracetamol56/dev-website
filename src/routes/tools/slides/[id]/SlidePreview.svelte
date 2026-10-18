<script lang="ts">
	import type { Readable, Writable } from 'svelte/store';
	import type { Indices } from '../types';
	import type { OutlineSection } from './outline';

	export let outline: Readable<OutlineSection[] | null>;
	export let current: Writable<Indices>;
</script>

<nav aria-label="Slides" class="flex gap-3 overflow-x-auto p-3 lg:h-full">
	{#each $outline ?? [] as section, index}
		<button
			type="button"
			aria-current={index === $current.h ? 'true' : undefined}
			class="flex aspect-video h-20 shrink-0 lg:h-full flex-col justify-between rounded-md bg-ctp-base p-2 text-left shadow-md shadow-ctp-crust
				transition-colors hover:bg-ctp-surface0 focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve
				aria-[current]:ring-2 aria-[current]:ring-ctp-mauve"
			on:click={() => current.set({ h: index, v: 0 })}
		>
			<span class="line-clamp-2 text-xs font-semibold text-ctp-text">
				{section.heading ?? 'Untitled'}
			</span>
			<span class="text-xs text-ctp-subtext0">{index + 1}</span>
		</button>
	{:else}
		<p class="px-1 py-6 text-sm text-ctp-subtext0">
			Add a <code class="text-ctp-mauve">&lt;section&gt;</code> to create a slide.
		</p>
	{/each}
</nav>
