<script lang="ts">
	import '../app.postcss';
	import { createToaster, createToolbar, melt } from '@melt-ui/svelte';
	import LoginDialog from './LoginDialog.svelte';
	import type { ToastData } from '$lib/toastData';
	import { flip } from 'svelte/animate';
	import Toast from './Toast.svelte';
	import ThemePopover from './ThemePopover.svelte';
	import { writable } from 'svelte/store';
	import { user } from '$lib/stores';

	$: theme = writable('mocha')

	const {
    elements: { root, link },
  } = createToolbar();

	export { theme };
</script>

<script lang="ts" context="module">
  const {
		elements,
		helpers: { addToast },
		states: { toasts },
		actions: { portal }
	} = createToaster<ToastData>();

  export { addToast };
</script>

<svelte:head>
	<title>Dev - Mathéo Galuba</title>
	<meta name="description" content="A showcase of my little dev projects" />
</svelte:head>

<!-- Toast list -->
<div
	class="fixed right-0 top-0 z-50 m-4
				flex flex-col items-end gap-2
				md:bottom-0 md:top-auto"
	use:portal
>
	{#each $toasts as toast (toast.id)}
		<div animate:flip={{ duration: 500 }}>
			<Toast {elements} {toast} />
		</div>
	{/each}
</div>


<nav
	class="min-w-max p-3"
	use:melt={$root}
>
	<div class="flex items-center gap-4 rounded-md bg-ctp-mantle p-3 text-ctp-text">
		<a href="/">
			<svg class="square-8" width="750" height="606" viewBox="0 0 750 606" fill="none" xmlns="http://www.w3.org/2000/svg">
				<g clip-path="url(#clip0_22_3)">
					<path class="fill-ctp-text" d="M240.326 193.691L355.66 0H750V121.2H427.826L312.5 314.891L240.326 193.691Z"/>
					<path class="fill-ctp-text" d="M312.5 436.024L125 121.2H0L125 363.6L240.326 557.284L312.5 436.024Z"/>
					<path class="fill-ctp-text" d="M750 181.733V606H437.5L500 484.8H625V302.933L536.049 303L608.236 181.766L750 181.733Z"/>
					<path class="fill-ctp-subtext0" d="M463.875 181.8L240.326 557.284L269.34 606H355.625L608.236 181.766L463.875 181.8Z"/>
					<path class="fill-ctp-subtext0" d="M0 606H125V363.6L0 121.2V606Z"/>
				</g>
				<defs>
					<clipPath id="clip0_22_3">
						<rect width="750" height="606" fill="white"/>
					</clipPath>
				</defs>
			</svg>
		</a>
		<div class="flex items-center gap-3 text-md font-semibold" >
			<a href="https://matheo-galuba.com" target="_blank" class="" use:melt={$link}>Portfolio</a>
			<a href="/map" class="" use:melt={$link}>Main map</a>
			<a href="/contact" class="" use:melt={$link}>Contact</a>
		</div>
		<div class="ml-auto flex items-center gap-2" >
			<ThemePopover {theme} />
			{#if user !== null}
			{$user}
			{:else}
			<LoginDialog />
			{/if}
		</div>
	</div>
</nav>
<main>
	<slot />
</main>
