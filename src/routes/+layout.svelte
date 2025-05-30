<script lang="ts" context="module">
	import { addToast, elements, portal, toasts } from '$lib/toast';

	export { addToast };
</script>

<script lang="ts">
	import '../app.postcss';
	import { flip } from 'svelte/animate';
	import Toast from '$lib/components/Toast.svelte';
	import Navbar from '$lib/components/Navbar.svelte';
	import Transition from '$lib/components/Transition.svelte';
	import Footer from '$lib/components/Footer.svelte';

	export let data;
</script>

<svelte:head>
	<script
		async
		src="https://analytics.matheo-galuba.com/script.js"
		data-website-id="c99860ae-f5a2-4f86-9dd2-5697c596efb4"
	></script>
	<title>Dev - Mathéo Galuba</title>
	<meta name="description" content="A showcase of my little dev projects" />
</svelte:head>

{#if data.url.includes('/embed')}
	<slot />
{:else}
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

	{#if data.url.includes('/tool/slides')}
		<main class="pt-[4.25rem] mb-auto">
			<Transition url={data.url}>
				<slot />
			</Transition>
		</main>
	{:else}
		<main class="py-32 mb-auto">
			<Transition url={data.url}>
				<slot />
			</Transition>
		</main>

		<Footer />
	{/if}
{/if}
