<script lang="ts">
	export let title = '';
	let className = '';
	export { className as class };
</script>

<section
	class="flex flex-col gap-4 rounded-md bg-ctp-mantle p-6 shadow-md shadow-ctp-crust {className}"
>
	{#if title || $$slots.actions}
		<div class="flex flex-wrap items-center justify-between gap-2">
			{#if title}<h2 class="text-xl font-bold text-ctp-text">{title}</h2>{/if}
			<slot name="actions" />
		</div>
	{/if}
	<slot />
</section>
