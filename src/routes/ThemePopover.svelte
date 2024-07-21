<script lang="ts">
	import {
		createPopover,
		createRadioGroup,
		melt,
		type CreateRadioGroupProps
	} from '@melt-ui/svelte';
	import { fade } from 'svelte/transition';
	import { Check, Palette, X } from 'lucide-svelte';
	import { variants } from '@catppuccin/palette';
	import { user } from '$lib/store';
	import { addToast } from './+layout.svelte';
	import api from '$lib/api';

	type Theme = keyof typeof variants;

	const onThemeChange: CreateRadioGroupProps['onValueChange'] = ({ curr, next }) => {
		if (curr === next) return curr;
		if (!(next in variants)) return curr;
		$user.flavour = next as keyof typeof variants;
		// API call to persist the theme on the user's profile if logged in
		if ($user.accessToken) {
			api
				.callWithAuth('patch', `/users/${$user.id}`, { flavour: next })
				.then((res) => {
					if (res.status === 200) {
						return next;
					}
					addToast({
						data: {
							title: 'Internal Error',
							description: 'Could not persist theme preference',
							color: 'bg-ctp-red'
						}
					});
					return curr;
				})
				.catch((err) => {
					addToast({
						data: {
							title: 'Internal Error',
							description: 'Could not persist theme preference',
							color: 'bg-ctp-red'
						}
					});
					return curr;
				});
		}
		return next;
	};

	const {
		elements: { trigger, content, arrow, close },
		states: { open }
	} = createPopover({
		forceVisible: true
	});

	const {
		elements: { root, item, hiddenInput },
		helpers: { isChecked }
	} = createRadioGroup({
		defaultValue: $user.flavour,
		onValueChange: onThemeChange
	});

	// Just to force tailwind to include the classes
	const variantsClasses = ['ctp-latte', 'ctp-frappe', 'ctp-macchiato', 'ctp-mocha'];
</script>

<button
	type="button"
	class="inline-flex square-8 items-center justify-center rounded-full bg-transparent p-0
        text-sm font-medium text-ctp-text transition-colors hover:bg-ctp-mauve/20"
	use:melt={$trigger}
	aria-label="Update dimensions"
>
	<Palette size="16" stroke-width="3" />
	<span class="sr-only">Open theme popover</span>
</button>

{#if $open}
	<div
		use:melt={$content}
		transition:fade={{ duration: 100 }}
		class="z-30 w-60 rounded-[4px] bg-ctp-base p-5 shadow-md shadow-ctp-crust"
	>
		<div use:melt={$arrow} />
		<div class="flex flex-col gap-2.5">
			<p class="mb-2 font-semibold text-ctp-text">Theme</p>

			<div
				use:melt={$root}
				class="flex flex-col gap-3 data-[orientation=horizontal]:flex-row"
				aria-label="View density"
			>
				{#each Object.keys(variants) as variant, i}
					<div class="flex max-h-full flex-col gap-0 overflow-y-auto">
						<button
							use:melt={$item(variant)}
							class="{variantsClasses[
								i
							]} bg-ctp-base relative cursor-pointer scroll-my-2 rounded-md py-2 pl-4 pr-4 text-ctp-text
                    {$isChecked(variant) ? 'bg-ctp-mauve/25 text-ctp-mauve' : ''}
                    hover:bg-ctp-mauve/25 hover:text-ctp-mauve"
							data-umami-event="update-theme"
							data-umami-event-properties={`{ "theme": "${variant}" }`}
						>
							{#if $isChecked(variant)}
								<div
									class="check absolute left-2 top-1/2 z-10 text-ctp-mauve transform -translate-y-1/2"
								>
									<Check class="square-4" />
								</div>
							{/if}
							<div class="pl-4">
								<p class="text-left capitalize font-medium">
									{variant}
								</p>
								<div class="flex w-full justify-between">
									<div class="square-2 rounded-full bg-ctp-rosewater" />
									<div class="square-2 rounded-full bg-ctp-flamingo" />
									<div class="square-2 rounded-full bg-ctp-pink" />
									<div class="square-2 rounded-full bg-ctp-mauve" />
									<div class="square-2 rounded-full bg-ctp-red" />
									<div class="square-2 rounded-full bg-ctp-maroon" />
									<div class="square-2 rounded-full bg-ctp-peach" />
									<div class="square-2 rounded-full bg-ctp-yellow" />
									<div class="square-2 rounded-full bg-ctp-green" />
									<div class="square-2 rounded-full bg-ctp-teal" />
									<div class="square-2 rounded-full bg-ctp-sky" />
									<div class="square-2 rounded-full bg-ctp-sapphire" />
									<div class="square-2 rounded-full bg-ctp-blue" />
									<div class="square-2 rounded-full bg-ctp-lavender" />
									<!-- 
                  <div class="square-3 bg-ctp-text"></div>
                  <div class="square-3 bg-ctp-subtext1"></div>
                  <div class="square-3 bg-ctp-subtext0"></div>
                  <div class="square-3 bg-ctp-overlay2"></div>
                  <div class="square-3 bg-ctp-overlay1"></div>
                  <div class="square-3 bg-ctp-overlay0"></div>
                  <div class="square-3 bg-ctp-surface2"></div>
                  <div class="square-3 bg-ctp-surface1"></div>
                  <div class="square-3 bg-ctp-surface0"></div>
                  <div class="square-3 bg-ctp-base"></div>
                  <div class="square-3 bg-ctp-mantle"></div>
                  <div class="square-3 bg-ctp-crust"></div>
                  -->
								</div>
							</div>
						</button>
					</div>
				{/each}
				<input name="line-height" use:melt={$hiddenInput} />
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
		</div>
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
