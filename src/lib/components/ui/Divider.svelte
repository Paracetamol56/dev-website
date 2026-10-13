<script lang="ts">
	export let vertical: boolean = false;
	let className = '';
	export { className as class };
</script>

{#if vertical}
	<span
		role="separator"
		aria-orientation="vertical"
		class="w-0.5 self-stretch bg-ctp-surface0 {className}"
	/>
{:else}
	<hr class="border-0 border-t-2 border-ctp-surface0 {className}" />
{/if}
