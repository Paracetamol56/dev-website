<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import api from '$lib/api';
	import { user } from '$lib/store';
	import Hero from '$lib/components/ui/Hero.svelte';
	import Loader from '$lib/components/ui/Loader.svelte';

	let allowed = false;

	onMount(async () => {
		if ($user.id === null) {
			goto('/');
			return;
		}
		try {
			const response = await api.callWithAuth('get', `/users/${$user.id}`);
			allowed = response.data.role === 'admin';
			if ($user.role !== (response.data.role ?? null)) {
				$user = { ...$user, role: response.data.role ?? null };
			}
		} catch (error) {
			console.error(error);
		}
		if (!allowed) goto('/');
	});
</script>

<svelte:head>
	<title>Admin - Mathéo Galuba</title>
	<meta name="robots" content="noindex, nofollow" />
</svelte:head>

{#if allowed}
	<Hero class="mb-10" highlight="Admin" />
	<section class="container mx-auto mb-16 max-w-6xl">
		<slot />
	</section>
{:else}
	<div class="flex justify-center py-32"><Loader /></div>
{/if}
