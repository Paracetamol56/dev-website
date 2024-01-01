<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import { user } from '$lib/store';
	import axios from 'axios';
	import { Archive, Download, Trash2, UserX, X } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { addToast } from '../+layout.svelte';
	import { create } from 'd3';
	import { createDialog, melt } from '@melt-ui/svelte';
	import { fade, fly } from 'svelte/transition';
	import { goto } from '$app/navigation';
	import type { Writable } from 'svelte/store';
	import type { UserSettings } from './userSettings';
	import api from '$lib/api';

	export let userSettings: Writable<UserSettings | null>;

	let createdAt: Date;
	let lastLogin: Date;

	$: {
		if ($userSettings) {
			createdAt = new Date($userSettings.createdAt);
			lastLogin = new Date($userSettings.lastLogin);
		}
	}

	const handleExport = () => {
		axios.get('/api/exports', {
			headers: {
				Authorization: `Bearer ${$user?.accessToken}`
			}
		}).then((response) => {
			const element = document.createElement('a');
			const file = new Blob([JSON.stringify(response.data)], { type: 'text/plain' });
			element.href = URL.createObjectURL(file);
			element.download = 'data.json';
			document.body.appendChild(element); // Required for this to work in FireFox
			element.click();
		});
	};

	let confirmation = '';
	let confirmationError = '';
	const validateConfirmation = () => {
		if (confirmation !== 'DELETE') {
			confirmationError = 'Invalid confirmation';
			return false;
		}
		confirmationError = '';
		return true;
	};

	const handleDelete = () => {
		if (!validateConfirmation()) {
			addToast({
				data: {
					title: 'Invalid form',
					description: 'Please check your inputs',
					color: 'bg-ctp-red'
				}
			});
			return;
		}

		api.callWithAuth('delete', `/users/${$user.id}`)
			.then(res => {
				api.logout();
				addToast({
					data: {
						title: 'Success',
						description: 'Your account has been deleted',
						color: 'bg-ctp-green'
					}
				});
				goto('/');
			})
			.catch((err) => {
				console.error(err);
				addToast({
					data: {
						title: 'Error',
						description: 'Failed load user data',
						color: 'bg-ctp-red'
					}
				});
				goto('/');
			});
	};

	const {
		elements: { trigger, overlay, content, title, description, close, portalled },
		states: { open }
	} = createDialog({
		forceVisible: true
	});
</script>

<div class="grid gap-x-8 gap-y-6">
	<div>
		<p>
			Account created on <span class="font-semibold"
				>{createdAt.toLocaleDateString('en-US', {
					year: 'numeric',
					month: 'long',
					day: 'numeric'
				})}</span
			>
		</p>
		<p>
			Last login on <span class="font-semibold"
				>{lastLogin.toLocaleDateString('en-US', {
					year: 'numeric',
					month: 'long',
					day: 'numeric'
				})}</span
			>
		</p>
	</div>
	<div>
		<h4 class="mb-2 flex items-center gap-1 text-base font-semibold">
			<Archive size="18" /> Export your data
		</h4>
		<div class="flex justify-start">
			<Button on:click={handleExport}>
				<span>Export</span>
				<Download size="16" />
			</Button>
		</div>
	</div>
	<div>
		<h4 class="mb-2 flex items-center gap-1 text-ctp-red text-base font-semibold">
			<UserX size="18" /> Delete your account
		</h4>
		<div class="flex justify-start">
			<button
				use:melt={$trigger}
				class="flex items-center gap-1 rounded-md bg-ctp-red px-3 py-1
					font-semibold text-ctp-mantle
					shadow-md shadow-ctp-crust transition-opacity
					hover:opacity-80 active:opacity-60"
			>
				<span>Delete</span>
				<Trash2 size="16" />
			</button>
		</div>
	</div>
</div>

<div class="" use:melt={$portalled}>
	{#if $open}
		<div
			use:melt={$overlay}
			class="fixed inset-0 z-10 bg-black/50"
			transition:fade={{ duration: 200 }}
		/>
		<div
			class="fixed left-[50%] top-[50%] z-50 max-h-[85vh] w-[90vw]
            max-w-[450px] translate-x-[-50%] translate-y-[-50%] rounded-lg bg-ctp-base
            p-6 shadow-md"
			transition:fly={{ duration: 200, y: 10 }}
			use:melt={$content}
		>
			<h2 use:melt={$title} class="m-0 text-lg font-medium text-ctp-text">Delete your account</h2>
			<p use:melt={$description} class="mb-5 mt-2 leading-normal text-ctp-text">
				You are about to delete your account. This action is irreversible.<br />
				<span class="font-semibold">Are you sure you want to continue?</span>
			</p>

		<form on:submit={handleDelete}>
				<fieldset class="">
					<label class="block mb-2 text-ctp-text" for="confirmation"
						>Type <span class="font-semibold pointer-events-none">DELETE</span> to confirm</label
					>
					<input
						class="h-8 rounded-md px-3 leading-none bg-ctp-surface0
                  focus:ring-2 focus:ring-ctp-mauve focus:outline-none {confirmationError != ''
							? 'ring-2 ring-ctp-red'
							: ''}"
						id="confirmation"
						bind:value={confirmation}
						on:blur={validateConfirmation}
						autocomplete="off"
						on:paste={(e) => e.preventDefault()}
						on:drop={(e) => e.preventDefault()}
						type="text"
					/>
				</fieldset>
				<p class="mb-4 text-sm text-ctp-red">{confirmationError}</p>
				<div class="mt-6 flex justify-end gap-4">
					<button
						use:melt={$close}
						aria-label="close"
						class="flex items-center gap-1 rounded-md border-2 border-ctp-mauve px-3 py-1
							font-semibold text-ctp-mauve
							shadow-md shadow-ctp-crust transition-opacity
							hover:opacity-80 active:opacity-60"
					>
						<span>Cancel</span>
						<X size="16" />
					</button>
					<Button type="submit">
						<span>Confirm</span>
						<Trash2 size="16" />
					</Button>
				</div>
			</form>
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
