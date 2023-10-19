<script lang="ts">
	import '../app.postcss';
	import { createToaster } from '@melt-ui/svelte';
	import type { ToastData } from '$lib/toastData';
	import { flip } from 'svelte/animate';
	import Toast from './Toast.svelte';
	import Navbar from './Navbar.svelte';
	import Transition from './Transition.svelte';
	import Footer from './Footer.svelte';

	export let data;
	export const prerender = true
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
	<script
		data-host="https://cdn.micrometrics.es"
		data-dnt="false"
		src="https://cdn.micrometrics.es/js/script.js"
		id="ZwSg9rf6GA"
		async
		defer
	></script>
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

<Navbar />

<main class="mb-auto">
	<Transition url={data.url}>
		<slot />
	</Transition>
</main>

<Footer />
