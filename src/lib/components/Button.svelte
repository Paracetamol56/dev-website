<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { goto } from '$app/navigation';

	export let type: 'button' | 'submit' | 'reset' = 'button';
	export let disabled: boolean = false;
	export let link: string | null = null;
	export let data: Record<string, any> = {};
	export let color = 'bg-ctp-mauve text-ctp-mantle';
	$: dataProps = Object.entries(data).reduce(
		(acc, [key, value]) => ({ ...acc, [`data-${key}`]: value }),
		{}
	);

	const dispatch = createEventDispatcher();

	function handleClick() {
		if (disabled) return;
		if (link) {
			goto(link);
		}
		dispatch('click');
	}
</script>

<button
	class="flex items-center gap-1 rounded-md px-3 py-1 {color}
        font-semibold
        shadow-md shadow-ctp-crust transition-opacity
        {disabled ? 'opacity-50 cursor-not-allowed' : 'hover:opacity-80 active:opacity-60'}"
	on:click={handleClick}
	{type}
	{disabled}
	{...dataProps}
>
	<slot />
</button>
