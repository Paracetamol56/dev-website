<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Readable } from 'svelte/store';
	import type { Indices } from '../types';
	import type { OutlineSlide } from './outline';

	export let outline: OutlineSlide[];
	export let current: Readable<Indices>;

	const dispatch = createEventDispatcher<{ select: { indices: Indices; from: number } }>();

	const item = `flex w-full items-baseline gap-2 rounded-md px-2 py-1 text-left text-sm text-ctp-subtext1
		hover:bg-ctp-surface0 hover:text-ctp-text focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve
		aria-[current]:bg-ctp-mauve/20 aria-[current]:text-ctp-mauve`;
</script>

<nav aria-label="Outline" class="p-2">
	{#if outline.length}
		<ol class="flex flex-col gap-0.5">
			{#each outline as stack, h}
				<li>
					<button
						type="button"
						class="{item} font-semibold"
						aria-current={$current.h === h && !stack.children.length ? 'true' : undefined}
						on:click={() => dispatch('select', { indices: { h, v: 0 }, from: stack.from })}
					>
						<span class="w-6 shrink-0 text-xs tabular-nums text-ctp-overlay1">{h + 1}</span>
						<span class="truncate">{stack.title}</span>
					</button>
					{#if stack.children.length}
						<ol class="ml-4 flex flex-col gap-0.5 border-l border-ctp-surface1 pl-2">
							{#each stack.children as slide, v}
								<li>
									<button
										type="button"
										class={item}
										aria-current={$current.h === h && $current.v === v ? 'true' : undefined}
										on:click={() => dispatch('select', { indices: { h, v }, from: slide.from })}
									>
										<span class="w-6 shrink-0 text-xs tabular-nums text-ctp-overlay1">
											{h + 1}.{v + 1}
										</span>
										<span class="truncate">{slide.title}</span>
									</button>
								</li>
							{/each}
						</ol>
					{/if}
				</li>
			{/each}
		</ol>
	{:else}
		<p class="px-2 py-6 text-sm text-ctp-subtext0">
			Add a <code class="text-ctp-mauve">&lt;section&gt;</code> to create a slide.
		</p>
	{/if}
</nav>
