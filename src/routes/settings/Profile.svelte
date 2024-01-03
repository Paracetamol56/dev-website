<script lang="ts">
	import { user } from '$lib/store';
	import axios from 'axios';
	import { AtSign, Save, User } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { addToast } from '../+layout.svelte';
	import Button from '$lib/components/Button.svelte';
	import type { Writable } from 'svelte/store';
	import type { UserSettings } from './userSettings';
	import api from '$lib/api';

	export let userSettings: Writable<UserSettings | null>;
	export let markSavedState: (saved: boolean | null) => void;

	let name: string;
	let nameError: string = '';
	let email: string;
	let emailError: string = '';

	onMount(() => {
		if ($userSettings !== null) {
			name = $userSettings.name;
			email = $userSettings.email;
		}
	});

	const validateName = () => {
		if (name.length === 0) {
			nameError = 'Name is required';
			return false;
		}
		if (name.length < 2) {
			nameError = 'Name must be at least 2 characters long';
			return false;
		}
		if (name.length > 100) {
			nameError = 'Name must be less than 100 characters long';
			return false;
		}
		nameError = '';
		markSavedState(false);
		return true;
	};
	const validateEmail = () => {
		if (email.length === 0) {
			emailError = 'Email is required';
			return false;
		}
		if (!/\S+@\S+\.\S+/.test(email)) {
			emailError = 'Email address is invalid';
			return false;
		}
		emailError = '';
		markSavedState(false);
		return true;
	};

	const handleSave = async () => {
		if (!validateName() || !validateEmail()) {
			addToast({
				data: {
					title: 'Invalid form',
					description: 'Please check your inputs',
					color: 'bg-ctp-red'
				}
			});
			return;
		}

		api
			.callWithAuth('patch', `/users/${$user.id}`, {
				name
			})
			.then((res) => {
				if (res.status === 200) {
					if ($userSettings !== null) $userSettings.name = name;
					markSavedState(true);
				}
			})
			.catch((err) => {
				console.error(err);
			});
	};
</script>

<form class="grid gap-x-8 gap-y-6" on:submit={handleSave}>
	<fieldset class="max-w-[15rem]">
		<label for="name" class="flex items-center gap-1 mb-2 text-sm font-semibold"
			><User size="16" /><span>Name</span></label
		>
		<input
			type="text"
			id="name"
			class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0 shadow-md shadow-ctp-crust px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			name="name"
			placeholder="Name"
			bind:value={name}
			on:blur={() => validateName()}
		/>
		<span class="text-left text-sm font-semibold text-ctp-red">{nameError}</span>
	</fieldset>
	<fieldset class="max-w-[15rem]">
		<label for="email" class="flex items-center gap-1 mb-2 text-sm font-semibold"
			><AtSign size="16" /><span>Email</span></label
		>
		<input
			type="email"
			disabled
			id="email"
			class="opacity-75 flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0 shadow-md shadow-ctp-crust px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			name="email"
			placeholder="Email"
			bind:value={email}
			on:blur={() => validateEmail()}
		/>
		<span class="text-left text-sm font-semibold text-ctp-red">{emailError}</span>
	</fieldset>
	<div class="flex justify-start">
		<Button type="submit">
			<span>Save</span>
			<Save size="16" />
		</Button>
	</div>
</form>
