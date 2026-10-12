<script lang="ts">
	import { createDialog, melt } from '@melt-ui/svelte';
	import { X } from 'lucide-svelte';
	import { writable, type Writable } from 'svelte/store';
	import { fade, fly } from 'svelte/transition';
	import { TextColor, TextSize, TextWeight } from './typography';

	export let open: Writable<boolean> = writable(false);
	export let title: string;
	export let description: string = '';

	const {
		elements: { overlay, content, title: heading, description: summary, close, portalled }
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
			class="fixed left-1/2 top-1/2 z-50 max-h-[90vh] w-[min(32rem,90vw)] -translate-x-1/2 -translate-y-1/2
				overflow-y-auto rounded-md bg-ctp-base p-6 shadow-md shadow-ctp-crust"
			transition:fly={{ duration: 150, y: 8 }}
		>
			<h2 use:melt={$heading} class="mb-4 pr-8 {TextSize.Xl} {TextWeight.Bold} {TextColor.Default}">
				{title}
			</h2>
			{#if description}
				<p use:melt={$summary} class="mb-6 {TextSize.Base} {TextColor.Muted}">{description}</p>
			{/if}
			<slot />
			{#if $$slots.footer}
				<div class="mt-6 flex justify-end gap-2">
					<slot name="footer" />
				</div>
			{/if}
			<button
				use:melt={$close}
				aria-label="Close"
				class="absolute right-4 top-4 grid place-items-center rounded-full p-1 {TextColor.Default}
					hover:bg-ctp-surface0 focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve"
			>
				<X size="16" />
			</button>
		</div>
	{/if}
</div>
