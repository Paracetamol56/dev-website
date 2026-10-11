<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { user } from '$lib/store';
	import type { PageData } from './$types';
	import Stats from './Stats.svelte';
	import TodoList from './TodoList.svelte';

	export let data: PageData;

	$: if ($user.id && !data.loggedIn) invalidateAll();
</script>

<svelte:head>
	<title>Ormi - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-16">
	<hgroup class="text-center">
		<h1 class="mb-2 text-6xl font-bold">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender"
				>Ormi</span
			>
		</h1>
		<p class="text-xl font-semibold text-ctp-subtext0" lang="el">Ορμή</p>
	</hgroup>
</section>

<section class="container mx-auto mb-16">
	{#if data.loggedIn && data.stats}
		<Stats stats={data.stats} />
		<TodoList refresh={data} />
	{:else}
		<p class="text-center font-semibold">Log in to start using Ormi.</p>
	{/if}
</section>

<section class="container mx-auto">
	<p class="max-w-xl mx-auto text-center text-sm text-ctp-overlay1">
		Okay so as a developer, I have to make a todo app at some point...<br />
		<span lang="el">Ορμή</span> means "momentum" in greek: when you dive into a project, you have a sudden
		burst of energy and motivation coming from nowhere. Ormi aims to help you keep that momentum going,
		the streak counter and the github like calendar are here to motivate you to keep working on your
		project.
	</p>
</section>
