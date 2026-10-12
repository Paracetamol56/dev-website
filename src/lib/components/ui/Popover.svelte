<script lang="ts">
	import { createPopover, melt } from '@melt-ui/svelte';
	import { writable, type Writable } from 'svelte/store';
	import { fade } from 'svelte/transition';
	import { TextColor, TextSize } from './typography';

	export let open: Writable<boolean> = writable(false);
	export let label: string;
	export let placement: 'top' | 'bottom' | 'left' | 'right' = 'bottom';
	let className = '';
	export { className as class };

	const {
		elements: { trigger, content, arrow }
	} = createPopover({ open, forceVisible: true, positioning: { placement } });
</script>

<button
	use:melt={$trigger}
	type="button"
	aria-label={label}
	class="inline-flex rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve {className}"
>
	<slot name="trigger" />
</button>

{#if $open}
	<div
		use:melt={$content}
		transition:fade={{ duration: 100 }}
		class="z-50 rounded-md bg-ctp-base p-4 shadow-md shadow-ctp-crust {TextSize.Base} {TextColor.Default}"
	>
		<div use:melt={$arrow} />
		<slot />
	</div>
{/if}
