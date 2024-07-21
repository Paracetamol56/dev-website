<script lang="ts">
	import { createPopover, melt } from '@melt-ui/svelte';
	import { fade } from 'svelte/transition';
	import { Cog, LogOut, User, X } from 'lucide-svelte';
	import { user } from '$lib/store';
	import { addToast } from './+layout.svelte';

	const logOut = () => {
		$user = {
			accessToken: null,
			refreshToken: null,
			id: null,
			email: null,
			flavour: $user.flavour
		};
		addToast({
			data: {
				title: 'Logged out',
				description: 'You have been logged out.',
				color: 'bg-ctp-green'
			}
		});
		location.reload();
	};

	const {
		elements: { trigger, content, arrow, close },
		states: { open }
	} = createPopover({
		forceVisible: true
	});
</script>

<button
	type="button"
	class="inline-flex square-8 items-center justify-center rounded-full bg-ctp-mauve p-0
        text-sm font-medium text-ctp-mantle hover:opacity-75 active:opacity-50 transition-opacity"
	use:melt={$trigger}
	aria-label="Update dimensions"
>
	<User size="20" />
	<span class="sr-only">Open profile popover</span>
</button>

{#if $open}
	<div
		use:melt={$content}
		transition:fade={{ duration: 100 }}
		class="z-30 w-48 rounded-[4px] bg-ctp-base p-5 shadow-md shadow-ctp-crust"
	>
		<div use:melt={$arrow} />
		<div class="flex flex-col gap-2.5">
			<p class="mb-2 font-semibold text-ctp-text">Profile</p>
			<ul>
				<li class="mb-2 text-ctp-text hover:text-ctp-blue transition-colors">
					<a href="/settings" class="flex justify-start items-center gap-2">
						<Cog size="18" stroke-width="3" />
						<span>Settings</span>
					</a>
				</li>
				<li class="mb-2 text-ctp-text hover:text-ctp-blue transition-colors">
					<button
						type="button"
						class="flex justify-start items-center gap-2"
						on:click={logOut}
						data-umami-event="logout"
					>
						<LogOut size="18" stroke-width="3" />
						<span>Log out</span>
					</button>
				</li>
			</ul>
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
