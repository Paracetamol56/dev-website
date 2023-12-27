<script lang="ts">
	import { melt, createTooltip } from '@melt-ui/svelte';
	import { tex } from 'mathlifier';
	import { fade } from 'svelte/transition';

	export let text: string;

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
		class="z-10 rounded-lg bg-ctp-text shadow"
	>
		<div use:melt={$arrow} />
		<p class="px-4 py-1 text-ctp-mantle">{text}</p>
	</div>
{/if}
