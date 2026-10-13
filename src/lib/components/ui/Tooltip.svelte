<script lang="ts">
	import { createTooltip, melt } from '@melt-ui/svelte';
	import { display as tex } from 'mathlifier';
	import { fade } from 'svelte/transition';
	import { TextColor } from './typography';

	export let text: string | undefined = undefined;
	export let html: string | undefined = undefined;
	export let math: string | undefined = undefined;
	export let placement: 'top' | 'bottom' | 'left' | 'right' = 'top';

	const {
		elements: { trigger, content, arrow },
		states: { open }
	} = createTooltip({
		positioning: { placement },
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
		class="z-50 overflow-hidden rounded-lg bg-ctp-text px-4 py-1 shadow {TextColor.Inverse}"
	>
		<div use:melt={$arrow} />
		{#if html}
			{@html html}
		{:else if math}
			{@html tex(math, { overflowAuto: false })}
		{:else if text}
			<p>{text}</p>
		{/if}
	</div>
{/if}
