<script lang="ts">
	import type { Page } from '$lib/page';
	import { ArrowRight, Lock } from 'lucide-svelte';
	import { user } from '$lib/store';

	export let page: Omit<Page, 'release'> & { release?: Date | string };
	export let path: string = '/page';
	export let requiresAuth: boolean = false;

	$: release = page.release ? new Date(page.release) : null;
</script>

<div
	class="group relative z-10 flex h-full flex-col rounded-md border-2 border-transparent bg-ctp-crust/50 p-6
				shadow-md shadow-ctp-crust backdrop-blur-sm transition duration-200
				hover:-translate-y-1 hover:border-ctp-mauve focus-within:border-ctp-mauve"
>
	<div class="mb-3 flex flex-wrap items-center gap-x-2 gap-y-1">
		{#each page.tags.slice(0, 4) as tag}
			<a
				class="relative z-10 text-sm font-semibold text-ctp-lavender hover:text-ctp-mauve transition-colors"
				href="{path}?tag={tag}"
			>
				#{tag}
			</a>
		{/each}
		{#if page.tags.length > 4}
			<span class="text-sm font-semibold text-ctp-lavender">...</span>
		{/if}
	</div>
	<h4 class="mb-3 text-2xl font-bold">
		<a
			class="outline-none after:absolute after:inset-0 after:rounded-md group-hover:text-ctp-mauve transition-colors"
			href="{path}/{page.slug}"
		>
			{page.title}
		</a>
	</h4>
	<p class="mb-6 text-ctp-subtext0">{page.description}</p>
	<div class="mt-auto flex items-center gap-3 text-sm font-semibold text-ctp-subtext0">
		{#if release && !isNaN(release.getTime())}
			<time datetime={release.toISOString()}>
				{release.toLocaleDateString('en-GB', {
					day: 'numeric',
					month: 'short',
					year: 'numeric',
					timeZone: 'UTC'
				})}
			</time>
		{/if}
		{#if requiresAuth && $user.id === null}
			<span class="flex items-center gap-1">
				<Lock size="14" />
				Login required
			</span>
		{/if}
		<ArrowRight
			size="18"
			class="ml-auto text-ctp-mauve opacity-0 -translate-x-2 transition duration-200 group-hover:opacity-100 group-hover:translate-x-0"
		/>
	</div>
</div>
