<script lang="ts">
	import { melt, createTooltip } from '@melt-ui/svelte';
	import { display as tex } from 'mathlifier';
	import { fade } from 'svelte/transition';

	export let text: string | undefined = undefined;
	export let html: string | undefined = undefined;
	export let math: string | undefined = undefined;

	$: {
		if (text === undefined && html === undefined && math === undefined) {
			throw new Error('At least one of text, html, or math must be defined');
		}
		if (text !== undefined && (html !== undefined || math !== undefined)) {
			throw new Error('Only one of text, html, or math can be defined');
		}
	}

	const {
		elements: { trigger, content, arrow },
		states: { open }
	} = createTooltip({
		positioning: {
			placement: 'top'
		},
		openDelay: 0,
		closeDelay: 0,
		closeOnPointerDown: false,
		forceVisible: true
	});
</script>

<div use:melt={$trigger}>
	<slot />
</div>

{#if $open}
	<div
		use:melt={$content}
		transition:fade={{ duration: 100 }}
		class="z-10 rounded-lg bg-ctp-text shadow overflow-hidden"
	>
		<div use:melt={$arrow} />
		{#if html}
			<div class="px-4 py-1">
				{@html html}
			</div>
		{:else if math}
			<div class="px-4 py-1 text-ctp-mantle">
				{@html tex(math, {overflowAuto: false})}
			</div>
		{:else if text}
			<p class="px-4 py-1 text-ctp-mantle">{text}</p>
		{/if}
	</div>
{/if}
