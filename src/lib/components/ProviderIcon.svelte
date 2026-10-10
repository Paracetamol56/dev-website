<script lang="ts">
	import { Mail } from 'lucide-svelte';
	import GitHub from './GitHub.svelte';
	import Google from './Google.svelte';

	export let provider: string;
	export let size: number = 16;
</script>

{#if provider === 'github'}
	<GitHub {size} />
{:else if provider === 'google'}
	<Google {size} />
{:else}
	<Mail {size} />
{/if}
