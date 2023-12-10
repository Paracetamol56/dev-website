<script lang="ts">
	import { user } from '$lib/stores';
	import { goto } from '$app/navigation';
	import { addToast } from '../+layout.svelte';
	import { createAccordion, melt } from '@melt-ui/svelte';
	import { onMount } from 'svelte';
	import { slide } from 'svelte/transition';
	import Integration from './Integration.svelte';
	import Appearance from './Appearance.svelte';
	import Profile from './Profile.svelte';
	import Account from './Account.svelte';

	onMount(() => {
		if (!$user) {
			addToast({
				data: {
					title: 'You are not logged in',
					description: 'Please log in to continue',
					color: 'bg-ctp-peach'
				}
			});

			goto('/');
		}
	});

	const items = [
		{
			id: 'profile',
			title: 'Profile',
			component: Profile
		},
		{
			id: 'account',
			title: 'Account',
			component: Account
		},
		{
			id: 'appearance',
			title: 'Appearance',
			component: Appearance
		},
		{
			id: 'integration',
			title: 'Integrations',
			component: Integration
		}
	];

	const {
		elements: { content, item, trigger, root },
		helpers: { isSelected }
	} = createAccordion({
		defaultValue: 'profile'
	});
</script>

<hgroup>
	<h1 class="mb-8 text-6xl font-bold text-center">
		<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
			Settings
		</span>
	</h1>
</hgroup>

<section class="container mx-auto">
	<div {...root}>
		{#each items as { id, title, component }, i}
			<div
				use:melt={$item(id)}
				class="overflow-hidden transition-colors first:rounded-t-xl last:rounded-b-xl shadow-md shadow-ctp-crust"
			>
				<h2 class="flex">
					<button
						use:melt={$trigger(id)}
						class="flex flex-1 cursor-pointer items-center justify-between
            bg-ctp-mantle px-5 py-5 text-lg font-semibold leading-none
            transition-colors hover:bg-ctp-crust focus:!ring-0
            focus-visible:text-magnum-800"
					>
						{title}
					</button>
				</h2>

				{#if $isSelected(id)}
					<div
						class="overflow-hidden bg-ctp-base px-5 py-4"
						use:melt={$content(id)}
						transition:slide={{ duration: 200 }}
					>
						<svelte:component this={component} />
					</div>
				{/if}
			</div>
		{/each}
	</div>
</section>
