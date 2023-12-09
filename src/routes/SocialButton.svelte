<script lang="ts">
	import { createTooltip, melt } from '@melt-ui/svelte';
	import { fade } from 'svelte/transition';

	export let href: string;
	export let tooltip: string;

	const {
		elements: { trigger, content, arrow },
		states: { open }
	} = createTooltip({
		positioning: {
			placement: 'top'
		},
		openDelay: 500,
		closeOnPointerDown: false,
		forceVisible: true
	});
</script>

<a {href} target="_blank" rel="noopener noreferrer" use:melt={$trigger}>
	<slot />
</a>

{#if $open}
	<div
		use:melt={$content}
		transition:fade={{ duration: 100 }}
		class="z-10 rounded-lg bg-ctp-text shadow"
	>
		<div use:melt={$arrow} />
		<p class="px-4 py-1 text-ctp-mantle">{tooltip}</p>
	</div>
{/if}
