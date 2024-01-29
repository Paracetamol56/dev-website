<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import api from '$lib/api';

	onMount(() => {
		const token: string = $page.url.searchParams.get('token') ?? '';
		const redirect: string = $page.url.searchParams.get('redirect') ?? '/';

		api.verify(token, redirect);
	});
</script>
