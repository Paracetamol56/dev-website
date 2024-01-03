<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import { user } from '$lib/store';
	import { variants } from '@catppuccin/palette';
	import { melt, createRadioGroup, type CreateRadioGroupProps } from '@melt-ui/svelte';
	import { Check, Palette, Save } from 'lucide-svelte';
	import type { Writable } from 'svelte/store';
	import type { UserSettings } from './userSettings';
	import api from '$lib/api';
	import { addToast } from '../+layout.svelte';

	export let userSettings: Writable<UserSettings | null>;
	export let markSavedState: (saved: boolean | null) => void;
	let flavour: keyof typeof variants;

	$: {
		if ($userSettings !== null) {
			flavour = $userSettings.flavour as keyof typeof variants;
		}
	}

	const handleSave = async () => {
		api
			.callWithAuth('patch', `/users/${$user.id}`, { flavour })
			.then((res) => {
				if (res.status === 200) {
					if ($userSettings !== null) $userSettings.flavour = flavour;
					markSavedState(true);
					$user.flavour = flavour;
				}
			})
			.catch((err) => {
				console.error(err);
				addToast({
					data: {
						title: 'Internal Error',
						description: 'Could not save the theme preference',
						color: 'bg-ctp-red'
					}
				});
			});
	};

	const variantsClasses = ['ctp-latte', 'ctp-frappe', 'ctp-macchiato', 'ctp-mocha'];

	const onFlavourChange: CreateRadioGroupProps['onValueChange'] = ({ curr, next }) => {
		if (curr === next) return curr;
		if (!(next in variants)) return curr;
		if ($userSettings?.flavour !== next) markSavedState(false);
		else markSavedState(null);
		flavour = next as keyof typeof variants;
		return next;
	};

	const {
		elements: { root, item },
		helpers: { isChecked }
	} = createRadioGroup({
		onValueChange: onFlavourChange,
		defaultValue: $user.flavour
	});
</script>

<form class="grid grid-cols-1 gap-x-8 gap-y-6" on:submit={handleSave}>
	<fieldset class="max-w-xl">
		<label for="flavour" class="flex items-center gap-1 mb-2 text-sm font-semibold">
			<Palette size="16" />
			<span>Flavour</span>
		</label>
		<div class="grid grid-cols-1 sm:grid-cols-2 rounded-lg bg-ctp-base gap-3" use:melt={$root}>
			{#each Object.keys(variants) as variant, i}
				<div
					class="{variantsClasses[
						i
					]} bg-ctp-base aspect-video relative cursor-pointer scroll-my-2 rounded-md py-2 pl-4 pr-4 ring-2
                    {$isChecked(variant)
						? 'ring-ctp-mauve text-ctp-mauve'
						: 'ring-ctp-crust text-ctp-text'}
                    hover:ring-ctp-mauve hover:text-ctp-mauve"
					use:melt={$item(variant)}
				>
					{#if $isChecked(variant)}
						<div
							class="check absolute left-2 top-1/2 z-10 text-ctp-mauve transform -translate-y-1/2"
						>
							<Check class="square-4" />
						</div>
					{/if}

					<div class="pl-4 h-full">
						<p class="mb-2 text-left capitalize text-lg font-semibold">
							{variant}
						</p>
						<div class="h-4 w-2/3 mb-4 rounded-full bg-ctp-mantle" />
						<div class="h-6 w-1/3 mb-4 rounded-full bg-ctp-mauve" />
						<div class="flex items-end w-full gap-1">
							<div class="flex-1 h-2 rounded-full bg-ctp-rosewater" />
							<div class="flex-1 h-2 rounded-full bg-ctp-flamingo" />
							<div class="flex-1 h-2 rounded-full bg-ctp-pink" />
							<div class="flex-1 h-2 rounded-full bg-ctp-mauve" />
							<div class="flex-1 h-2 rounded-full bg-ctp-red" />
							<div class="flex-1 h-2 rounded-full bg-ctp-maroon" />
							<div class="flex-1 h-2 rounded-full bg-ctp-peach" />
							<div class="flex-1 h-2 rounded-full bg-ctp-yellow" />
							<div class="flex-1 h-2 rounded-full bg-ctp-green" />
							<div class="flex-1 h-2 rounded-full bg-ctp-teal" />
							<div class="flex-1 h-2 rounded-full bg-ctp-sky" />
							<div class="flex-1 h-2 rounded-full bg-ctp-sapphire" />
							<div class="flex-1 h-2 rounded-full bg-ctp-blue" />
							<div class="flex-1 h-2 rounded-full bg-ctp-lavender" />
						</div>
					</div>
				</div>
			{/each}
		</div>
		<small class="text-center">
			<img
				src={$user.flavour === 'latte' ? '/img/catppuccin-light.png' : '/img/catppuccin-dark.png'}
				alt="Catppuccin logo"
				class="inline-block rounded-full square-4 mr-1"
			/>
			Powered by
			<a class="text-ctp-blue" href="https://github.com/catppuccin/catppuccin" target="_blank"
				>Catppuccin</a
			>
		</small>
	</fieldset>
	<div class="flex justify-start">
		<Button type="submit">
			<span>Save</span>
			<Save size="16" />
		</Button>
	</div>
</form>
