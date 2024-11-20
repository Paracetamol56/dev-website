<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import api from '$lib/api';
	import Button from '$lib/components/Button.svelte';
	import { ArrowRightToLine } from 'lucide-svelte';
	import { addToast } from '../../+layout.svelte';
	import { createPinInput, melt } from '@melt-ui/svelte';
	import { user } from '$lib/store';

	let codeError: string = '';

	const {
		elements: { root, input },
		states: { value }
	} = createPinInput({
		placeholder: '•',
		defaultValue: $page.url.searchParams.get('code')?.split('') ?? []
	});

	const validateCode = (code: string[]) => {
		if (code.length !== 5 || code.some((char) => char === '')) {
			codeError = 'Code must be 5 characters long';
			return false;
		}
		// Every character must be a number, lowercase letter or uppercase letter
		if (!code.every((char) => /[a-zA-Z0-9]/.test(char))) {
			codeError = 'Code must only contain numbers and letters';
			return false;
		}
		codeError = '';
		return true;
	};

	const handleSubmit = (e: Event) => {
		e.preventDefault();
		if (validateCode($value)) {
			api
				.call('GET', `/word-cloud?code=${$value.join('')}`)
				.then((res) => {
					if (res.data.open) {
						goto(`/tool/word-cloud/${res.data.id}`);
					} else {
						codeError = 'This session is closed';
						addToast({
							data: {
								title: 'Error',
								description: 'The code you provided does not match any open session',
								color: 'bg-ctp-red'
							}
						});
					}
				})
				.catch((error) => {
					console.error(error);
					if (error.response.status === 404) {
						codeError = 'Session not found';
						addToast({
							data: {
								title: 'Error',
								description: 'The code you provided does not match any open session',
								color: 'bg-ctp-red'
							}
						});
					} else {
						codeError = 'An error occured';
						addToast({
							data: {
								title: 'Error',
								description: 'An error occured while joining the session',
								color: 'bg-ctp-red'
							}
						});
					}
				});
		}
	};
</script>

<svelte:head>
	<title>Word cloud - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-6xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
				Word cloud
			</span>
		</h1>
	</hgroup>
</section>

<section class="container mx-auto">
	<form on:submit={handleSubmit}>
		<div class="flex flex-col gap-8 items-center">
			<div class="w-fit">
				<label for="code" class="mb-2 text-sm font-semibold"> Session code </label>
				<div use:melt={$root} class="flex items-center gap-2">
					{#each Array.from({ length: 5 }) as _}
						<input
							id="code"
							autocomplete="off"
							type="text"
							maxlength="1"
							class="rounded-md bg-ctp-surface0 text-center text-lg text-ctp-text square-12
                  shadow-md shadow-ctp-crust focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
							on:keydown={(e) => {
								if (e.key === 'Enter') {
									handleSubmit(e);
								}
							}}
							use:melt={$input()}
						/>
					{/each}
				</div>
				<p class="text-left text-sm font-semibold text-ctp-red">{codeError}</p>
			</div>
			<Button type="submit">
				<span>Join</span>
				<ArrowRightToLine size="18" />
			</Button>
		</div>
	</form>

	<div class="mt-8 flex justify-center gap-4">
		<a class="font-semibold hover:text-ctp-blue transition-colors" href="/tool/word-cloud/new"
			>Create a new session</a
		>
		{#if $user.id !== null}
			<a class="font-semibold hover:text-ctp-blue transition-colors" href="/tool/word-cloud/manage"
				>View your sessions</a
			>
		{/if}
	</div>
</section>
