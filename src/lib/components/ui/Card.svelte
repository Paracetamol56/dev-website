<script lang="ts">
	export let as: 'div' | 'section' | 'article' | 'li' = 'div';
	export let interactive: boolean = false;
	export let padded: boolean = true;
	let className = '';
	export { className as class };
</script>

<svelte:element
	this={as}
	class="rounded-md bg-ctp-mantle shadow-md shadow-ctp-crust {padded ? 'p-6' : ''}
		{interactive ? 'border-2 border-transparent transition-colors hover:border-ctp-mauve' : ''}
		{className}"
	{...$$restProps}
>
	<slot />
</svelte:element>
