<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { completeOAuth } from '$lib/oauth';

	onMount(() => completeOAuth('github', $page.url.searchParams));
</script>
