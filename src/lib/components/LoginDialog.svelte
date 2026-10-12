<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import { Tone } from '$lib/components/ui/styles';
	import { createDialog, melt } from '@melt-ui/svelte';
	import { KeyRound, Send, X } from 'lucide-svelte';
	import { isCancellation, loginWithPasskey, passkeysSupported } from '$lib/passkey';
	import { addToast } from '$lib/toast';
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
				<div class="mt-6 flex flex-col gap-3">
					{#if passkeysSupported()}
						<Button
							class="w-full gap-2 !py-2"
							tone={Tone.Neutral}
							on:click={handlePasskey}
							data-umami-event="login-passkey"
						>
							<KeyRound size="18" />
							<span>Continue with a passkey</span>
						</Button>
					{/if}
					{#each providers as provider}
						<Button
							class="w-full gap-2 !py-2"
							tone={Tone.Neutral}
							on:click={() => startOAuth(provider, window.location.pathname)}
							data-umami-event="login-oauth"
							data-umami-event-provider={provider.name}
						>
							<ProviderIcon provider={provider.name} size={18} />
							<span>Continue with {providerLabels[provider.name] ?? provider.name}</span>
						</Button>
					{/each}
				</div>
				<div class="my-5 flex items-center gap-4 text-sm font-semibold text-ctp-subtext0">
					<span class="h-px flex-1 bg-ctp-surface1" />
					or
					<span class="h-px flex-1 bg-ctp-surface1" />
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
				<Button class="w-full gap-2 !py-2" type="submit" data-umami-event="login-email">
					<Send size="18" />
					<span>Send the magic link</span>
				</Button>
			</form>

			<p class="text-right">
				<small>
					Check the <a class="text-ctp-mauve" href="/pages/privacy-policy">privacy policy</a>.
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
