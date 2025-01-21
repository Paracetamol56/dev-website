<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { goto } from '$app/navigation';

	export let type: 'button' | 'submit' | 'reset' = 'button';
	export let disabled: boolean = false;
	export let link: string | null = null;
	export let data: Record<string, any> = {};
	$: dataProps = Object.entries(data).reduce((acc, [key, value]) => ({...acc, [`data-${key}`]: value}), {});

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
	class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1
        font-semibold text-ctp-mantle
        shadow-md shadow-ctp-crust transition-opacity
        {disabled ? 'opacity-50 cursor-not-allowed' : 'hover:opacity-80 active:opacity-60'}"
	on:click={handleClick}
	{type}
	{...dataProps}
>
	<slot />
</button>
