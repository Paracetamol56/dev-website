<script lang="ts">
	import { createDialog, melt } from '@melt-ui/svelte';
	import { X } from 'lucide-svelte';
	import type { Writable } from 'svelte/store';
	import { fade, fly } from 'svelte/transition';

	export let open: Writable<boolean>;
	export let title: string;

	const {
		elements: { overlay, content, title: heading, close, portalled }
	} = createDialog({ open, forceVisible: true });
</script>

<div use:melt={$portalled}>
	{#if $open}
		<div
			use:melt={$overlay}
			class="fixed inset-0 z-30 bg-ctp-crust/70"
			transition:fade={{ duration: 150 }}
		/>
		<div
			use:melt={$content}
			class="fixed left-1/2 top-1/2 z-50 w-[min(32rem,90vw)] -translate-x-1/2 -translate-y-1/2
				rounded-md bg-ctp-base p-6 shadow-md shadow-ctp-crust"
			transition:fly={{ duration: 150, y: 8 }}
		>
			<h2 use:melt={$heading} class="mb-4 pr-8 text-xl font-bold text-ctp-text">{title}</h2>
			<slot />
			<button
				use:melt={$close}
				aria-label="Close"
				class="absolute right-4 top-4 grid place-items-center rounded-full p-1 hover:bg-ctp-surface0"
			>
				<X size="16" />
			</button>
		</div>
	{/if}
</div>
