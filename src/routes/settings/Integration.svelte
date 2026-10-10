<script lang="ts">
	import { Check, KeyRound, Plus } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import type { Writable } from 'svelte/store';
	import type { UserSettings } from './userSettings';
	import api from '$lib/api';
	import Button from '$lib/components/Button.svelte';
	import ProviderIcon from '$lib/components/ProviderIcon.svelte';
	import { getProviders, providerLabels, startOAuth, type OAuthProvider } from '$lib/oauth';
	import { addToast } from '../+layout.svelte';
	import { addPasskey, isCancellation, passkeysSupported } from '$lib/passkey';

	export let userSettings: Writable<UserSettings | null>;

	let providers: OAuthProvider[] = [];
	onMount(async () => (providers = await getProviders()));

	$: identities = $userSettings?.identities ?? [];
	$: names = [
		...new Set([
			'email',
			...providers.map((provider) => provider.name),
			...identities.map((identity) => identity.provider)
		])
	];
	$: rows = names.map((name) => ({
		name,
		identity: identities.find((identity) => identity.provider === name),
		provider: providers.find((provider) => provider.name === name)
	}));

	const formatDate = (date: string) =>
		new Date(date).toLocaleDateString(undefined, { dateStyle: 'medium' });

	$: passkeys = $userSettings?.passkeys ?? [];
	let passkeyName = '';

	const showError = (description: string) =>
		addToast({ data: { title: 'Error', description, color: 'bg-ctp-red' } });

	const createPasskey = async () => {
		if (!$userSettings) return;
		try {
			$userSettings = await addPasskey($userSettings.id, passkeyName.trim());
			passkeyName = '';
		} catch (error: any) {
			if (isCancellation(error)) return;
			console.error(error);
			showError(error.response?.data?.error ?? 'Failed to add the passkey.');
		}
	};

	const removePasskey = (id: string) => {
		api
			.callWithAuth('DELETE', `/users/${$userSettings?.id}/passkeys/${id}`)
			.then((response) => ($userSettings = response.data))
			.catch((error) => {
				console.error(error);
				showError('Failed to remove the passkey.');
			});
	};

	const unlink = (provider: string) => {
		api
			.callWithAuth('DELETE', `/users/${$userSettings?.id}/identities/${provider}`)
			.then((response) => ($userSettings = response.data))
			.catch((error) => {
				console.error(error);
				showError(`Failed to disconnect ${providerLabels[provider] ?? provider}.`);
			});
	};
</script>

{#if $userSettings !== null}
	<p class="mb-4">
		These are the ways you can log in. They all lead to the same account because they share your
		email address, <strong>{$userSettings.email}</strong>.
	</p>

	<ul class="flex flex-col gap-3">
		{#each rows as { name, identity, provider }}
			<li
				class="flex flex-wrap items-center gap-4 rounded-lg bg-ctp-mantle p-4 shadow-md shadow-ctp-crust"
			>
				<span
					class="flex shrink-0 items-center justify-center rounded-md bg-ctp-surface0 square-10"
				>
					<ProviderIcon provider={name} size={20} />
				</span>
				<div class="min-w-0 flex-1">
					<p class="flex flex-wrap items-center gap-2 font-semibold">
						{providerLabels[name] ?? name}
						{#if name === 'email'}
							<span class="text-xs font-normal text-ctp-overlay1">Default</span>
						{:else if identity}
							<span
								class="flex items-center gap-1 rounded-full bg-ctp-green/15 px-2 py-0.5 text-xs text-ctp-green"
							>
								<Check size="12" />
								Connected
							</span>
						{/if}
					</p>
					<p class="text-sm text-ctp-subtext0">
						{#if name === 'email'}
							Magic link sent to {$userSettings.email}
							{#if identity}· Last used {formatDate(identity.lastLogin)}{/if}
						{:else if identity}
							{#if identity.profileUrl}
								<a class="text-ctp-blue" href={identity.profileUrl} target="_blank" rel="noreferrer"
									>{identity.username ?? identity.email}</a
								>
							{:else}
								{identity.email}
							{/if}
							· Connected {formatDate(identity.linkedAt)} · Last used {formatDate(
								identity.lastLogin
							)}
						{:else}
							Not connected
						{/if}
					</p>
				</div>
				{#if identity && name !== 'email'}
					<button
						type="button"
						class="rounded-md px-3 py-1 text-sm font-semibold text-ctp-red transition-colors hover:bg-ctp-surface0"
						on:click={() => unlink(name)}
					>
						Disconnect
					</button>
				{:else if provider}
					<Button on:click={() => provider && startOAuth(provider, '/settings')}>
						<span>Connect</span>
					</Button>
				{/if}
			</li>
		{/each}

		<li class="rounded-lg bg-ctp-mantle p-4 shadow-md shadow-ctp-crust">
			<div class="flex flex-wrap items-center gap-4">
				<span
					class="flex shrink-0 items-center justify-center rounded-md bg-ctp-surface0 square-10"
				>
					<KeyRound size="20" />
				</span>
				<div class="min-w-0 flex-1">
					<p class="flex flex-wrap items-center gap-2 font-semibold">
						Passkeys
						{#if passkeys.length > 0}
							<span
								class="flex items-center gap-1 rounded-full bg-ctp-green/15 px-2 py-0.5 text-xs text-ctp-green"
							>
								<Check size="12" />
								{passkeys.length} registered
							</span>
						{/if}
					</p>
					<p class="text-sm text-ctp-subtext0">
						Log in with your fingerprint, face or device PIN. Each passkey only works on this
						website.
					</p>
				</div>
			</div>

			{#if passkeys.length > 0}
				<ul class="mt-4 flex flex-col divide-y divide-ctp-surface0 border-t border-ctp-surface0">
					{#each passkeys as passkey}
						<li class="flex flex-wrap items-center gap-x-4 gap-y-1 py-2">
							<span class="font-semibold">{passkey.name}</span>
							<span class="text-sm text-ctp-subtext0">
								Added {formatDate(passkey.createdAt)} ·
								{passkey.lastUsed ? `Last used ${formatDate(passkey.lastUsed)}` : 'Never used'}
							</span>
							<button
								type="button"
								class="ml-auto rounded-md px-3 py-1 text-sm font-semibold text-ctp-red transition-colors hover:bg-ctp-surface0"
								on:click={() => removePasskey(passkey.id)}
							>
								Remove
							</button>
						</li>
					{/each}
				</ul>
			{/if}

			{#if passkeysSupported()}
				<form class="mt-4 flex flex-wrap gap-2" on:submit|preventDefault={createPasskey}>
					<input
						type="text"
						maxlength="50"
						placeholder="Name, e.g. Laptop"
						aria-label="Passkey name"
						class="h-8 min-w-0 flex-1 rounded-md bg-ctp-surface0 px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve sm:max-w-xs"
						bind:value={passkeyName}
					/>
					<Button type="submit">
						<Plus size="16" />
						<span>Add a passkey</span>
					</Button>
				</form>
			{:else}
				<p class="mt-4 text-sm text-ctp-subtext0">This browser does not support passkeys.</p>
			{/if}
		</li>
	</ul>
{/if}
