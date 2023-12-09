<script lang="ts">
	import { onMount } from 'svelte';

	onMount(() => {
		document.querySelectorAll('pre').forEach((element: Element) => {
			element.removeAttribute('style');
		});
	});
</script>

<div
	class="flex p-4 my-8 xl:mx-32 2xl:mx-64 flex-col overflow-x-auto bg-ctp-mantle rounded-md shadow-md shadow-ctp-crust data-[orientation=vertical]:flex-row'}"
>
	<slot />
</div>
