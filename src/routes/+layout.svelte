<script lang="ts">
	import '../app.postcss';
	import { createToaster } from '@melt-ui/svelte';
	import type { ToastData } from '$lib/toastData';
	import { flip } from 'svelte/animate';
	import Toast from './Toast.svelte';
	import { writable } from 'svelte/store';
	import Navbar from './Navbar.svelte';
	import Footer from './Footer.svelte';
	import { user } from '$lib/stores';
	import { onMount } from 'svelte';
	import { variants } from '@catppuccin/palette';

	$: theme = writable($user! ? $user!.flavour : 'mocha');

	onMount(() => {
		theme.subscribe((value) => {
			const body = document.querySelector('body');
			body?.classList.remove(...Object.keys(variants).map((v) => `ctp-${v}`));
			body?.classList.add(`ctp-${value}`);
		});
	});

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

<Navbar theme={theme} />
<main class="mb-auto">
	<slot />
</main>
<Footer />
