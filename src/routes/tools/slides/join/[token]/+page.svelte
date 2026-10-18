<script lang="ts">
	import axios from 'axios';
	import { Loader2, Presentation } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import api from '$lib/api';
	import LoginDialog from '$lib/components/LoginDialog.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { user } from '$lib/store';

	let joining = false;
	let error = '';

	async function join(token: string) {
		joining = true;
		try {
			const joined = (await api.callWithAuth('POST', '/slides/join', { token })).data as {
				id: string;
			};
			await goto(`/tools/slides/${joined.id}`, { replaceState: true });
		} catch (err) {
			const status = axios.isAxiosError(err) ? err.response?.status : undefined;
			error =
				status === 404 || status === 400
					? 'This link does not work anymore. Ask the owner of the deck for a new one.'
					: 'The deck could not be opened, try again in a moment.';
			joining = false;
		}
	}

	$: if ($user.accessToken && !joining && !error) join($page.params.token ?? '');
</script>

<svelte:head>
	<title>Join a deck - Slides - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32 flex justify-center">
	<div
		class="flex w-full max-w-md flex-col items-center gap-4 rounded-md bg-ctp-mantle p-8 text-center shadow-md shadow-ctp-crust"
	>
		<Presentation size="32" class="text-ctp-mauve" />
		<h1 class="text-2xl font-bold text-ctp-text">Join a slide deck</h1>
		{#if error}
			<p class="text-ctp-subtext0" role="alert">{error}</p>
			<Button href="/tools/slides">Go to your decks</Button>
		{:else if $user.accessToken}
			<p class="flex items-center gap-2 text-ctp-subtext0" role="status">
				<Loader2 size="16" class="animate-spin" /> Opening the deck…
			</p>
		{:else}
			<p class="text-ctp-subtext0">Someone shared a slide deck with you. Log in to open it.</p>
			<LoginDialog />
		{/if}
	</div>
</section>
