<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import { theme, user } from '$lib/stores';
	import { variants } from '@catppuccin/palette';
	import {
		melt,
		createRadioGroup,
		type CreateRadioGroupProps
	} from '@melt-ui/svelte';
	import axios from 'axios';
	import { Check, Palette, Save } from 'lucide-svelte';
	import { onMount } from 'svelte';

	let flavour: keyof typeof variants;

	onMount(() => {
		flavour = $theme as keyof typeof variants;
	});

	const handleSave = async () => {
		axios
			.patch(`/api/user/${$user?.id}`, {
				flavour: flavour || $user?.flavour
			})
			.then((res) => {
				$theme = flavour;
			})
			.catch((err) => {
				console.error(err);
			});
	};

	const variantsClasses = ['ctp-latte', 'ctp-frappe', 'ctp-macchiato', 'ctp-mocha'];

	const onFlavourChange: CreateRadioGroupProps['onValueChange'] = ({ curr, next }) => {
		if (!next) {
			return curr;
		}
		flavour = next as keyof typeof variants;
		return next;
	};

	const {
		elements: { root, item, hiddenInput },
		helpers: { isChecked }
	} = createRadioGroup({
		onValueChange: onFlavourChange,
		defaultValue: $theme
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
				src={$theme === 'latte' ? '/img/catppuccin-light.png' : '/img/catppuccin-dark.png'}
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
