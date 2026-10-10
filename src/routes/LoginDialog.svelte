<script lang="ts">
	import '../app.postcss';
	import Button from '$lib/components/Button.svelte';
	import { createDialog, melt } from '@melt-ui/svelte';
	import { KeyRound, Send, X } from 'lucide-svelte';
	import { isCancellation, loginWithPasskey, passkeysSupported } from '$lib/passkey';
	import { addToast } from './+layout.svelte';
	import { writable, type Writable } from 'svelte/store';
	import { fade, fly } from 'svelte/transition';
	import api from '$lib/api';
	import ProviderIcon from '$lib/components/ProviderIcon.svelte';
	import { getProviders, providerLabels, startOAuth, type OAuthProvider } from '$lib/oauth';
	import { onMount } from 'svelte';

	const dialogOpen: Writable<boolean> = writable(false);
	const {
		elements: { trigger, overlay, content, title, description, close, portalled },
		states: { open }
	} = createDialog({
		forceVisible: true,
		open: dialogOpen
	});

	let email = '';
	let emailError = '';

	let providers: OAuthProvider[] = [];
	onMount(async () => (providers = await getProviders()));

	function validateEmail(): boolean {
		if (email.length == 0) {
			emailError = 'Email is required';
			return false;
		}
		if (!email.match(/^[^\s@]+@[^\s@]+\.[^\s@]+$/)) {
			emailError = 'Invalid email';
			return false;
		}
		emailError = '';
		return true;
	}

	const handlePasskey = async () => {
		try {
			await loginWithPasskey();
			dialogOpen.set(false);
		} catch (error: any) {
			if (isCancellation(error)) return;
			console.error(error);
			addToast({
				data: {
					title: 'Passkey login failed',
					description: error.response?.data?.error ?? 'An error occured, please try again later',
					color: 'bg-ctp-red'
				}
			});
		}
	};

	const handleSubmit = (event: Event) => {
		event.preventDefault();

		if (!validateEmail()) {
			addToast({
				data: {
					title: 'Invalid form',
					description: 'Please check your inputs',
					color: 'bg-ctp-red'
				}
			});
			return;
		}

		dialogOpen.set(false);

		api.login(email);
	};
</script>

<!-- Button -->
<button
	class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1
  font-semibold text-ctp-mantle
  shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60"
	use:melt={$trigger}
	data-umami-event="login-dialog"
>
	Login
</button>

<!-- Dialog -->
<div use:melt={$portalled}>
	{#if $open}
		<div
			use:melt={$overlay}
			class="fixed inset-0 z-30 bg-black/50"
			transition:fade={{ duration: 200 }}
		/>
		<div
			class="fixed left-[50%] top-[50%] z-50 max-h-[85vh] w-[90vw]
            max-w-[450px] translate-x-[-50%] translate-y-[-50%] rounded-lg bg-ctp-base
            p-6 shadow-md"
			transition:fly={{ duration: 200, y: 10 }}
			use:melt={$content}
		>
			<h2 use:melt={$title} class="m-0 text-lg font-medium text-ctp-text">Login</h2>

			{#if providers.length > 0 || passkeysSupported()}
				<div class="mt-6 flex flex-wrap justify-center gap-4">
					{#if passkeysSupported()}
						<Button on:click={handlePasskey} data={{ 'umami-event': 'login-passkey' }}>
							<span>Continue with a passkey</span>
							<KeyRound size="16" />
						</Button>
					{/if}
					{#each providers as provider}
						<Button
							on:click={() => startOAuth(provider, window.location.pathname)}
							data={{
								'umami-event': 'login-oauth',
								'umami-event-provider': provider.name
							}}
						>
							<span>Continue with {providerLabels[provider.name] ?? provider.name}</span>
							<ProviderIcon provider={provider.name} />
						</Button>
					{/each}
				</div>
				<div class="my-4 flex justify-between items-center">
					<span class="w-full h-px bg-ctp-text opacity-20" />
					<strong class="mx-4">OR</strong>
					<span class="w-full h-px bg-ctp-text opacity-20" />
				</div>
			{/if}

			<p use:melt={$description} class="mb-5 mt-2 leading-normal text-ctp-text">
				This is a passwordless authentication. I only need your email to send you a magic link.
				<br />
				<strong> Don't forget to check your spam folder! </strong>
			</p>

			<form on:submit={handleSubmit} class="mb-5">
				<fieldset class="flex items-center gap-5">
					<label class="text-right text-ctp-text" for="name"> Email </label>
					<input
						class="h-8 flex-1 rounded-md px-3 leading-none bg-ctp-surface0
                  focus:ring-2 focus:ring-ctp-mauve focus:outline-none {emailError != ''
							? 'ring-2 ring-ctp-red'
							: ''}"
						id="email"
						bind:value={email}
						on:blur={validateEmail}
						type="email"
					/>
				</fieldset>
				<p class="mb-4 text-right text-sm text-ctp-red">{emailError}</p>
				<div class="mt-6 flex justify-center gap-4">
					<Button
						type="submit"
						data={{
							'umami-event': 'login-email'
						}}
					>
						<span>Send the magic link</span>
						<Send size="16" />
					</Button>
				</div>
			</form>

			<p class="text-right">
				<small>
					Check the <a class="text-ctp-mauve" href="/page/privacy-policy">privacy policy</a>.
				</small>
			</p>

			<button
				use:melt={$close}
				aria-label="close"
				class="absolute right-4 top-4 inline-flex h-6 w-6 appearance-none
                items-center justify-center rounded-full p-1 text-base
                hover:bg-ctp-mauve hover:text-ctp-base transition-colors"
			>
				<X class="square-4" />
			</button>
		</div>
	{/if}
</div>
