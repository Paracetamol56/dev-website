<script lang="ts">
	import { Plus } from 'lucide-svelte';
	import type { PageData } from '../$types';

	export let data: PageData;
</script>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-6xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender"
				>Your sessions</span
			>
		</h1>
	</hgroup>
</section>

<section class="container mx-auto">
	<div class="mb-8 flex justify-center">
		<a href="/word-cloud/new">
			<button
				class="flex justify-center items-center rounded-md bg-ctp-mauve px-3 py-1 font-medium
      text-ctp-surface0 hover:opacity-75 active:opacity-50 transition-opacity"
			>
				New session
				<Plus stroke-width="3" />
			</button>
		</a>
	</div>

	<div class="grid grid-cols-1 gap-8 md:grid-cols-2 xl:grid-cols-3">
		{#each data.sessions as session}
			<div class="p-8 bg-ctp-mantle rounded-md shadow-md shadow-ctp-crust">
				<div class="flex justify-start items-baseline mb-4">
					{#if session.open}
						<div class="relative square-2 mr-2 bg-ctp-green rounded-full">
							<span
								class="animate-ping absolute top-0.5 right-0.5 block square-1 rounded-full ring-2 ring-ctp-green bg-ctp-green"
							/>
						</div>
						<p class="text-ctp-subtext0 text-sm">Open</p>
					{:else}
						<div class="relative square-2 mr-2 bg-ctp-red rounded-full" />
						<p class="text-ctp-subtext0 text-sm">Closed</p>
					{/if}
					<p class="ml-auto text-ctp-subtext0 text-sm">
						{session.submitions} submition{session.submitions <= 1 ? '' : 's'}
					</p>
				</div>
				<a href="/word-cloud/{session.id}">
					<h4 class="mb-4 text-2xl font-bold hover:opacity-75 transition-opacity">
						{session.name}
					</h4>
				</a>
				<p class="text-ctp-subtext0">{session.description}</p>
			</div>
		{/each}
	</div>
</section>
