<script lang="ts">
	import TodoList from "./TodoList.svelte";
	import type { PageData } from "./$types";

	export let data: PageData;
</script>

<svelte:head>
	<title>Ormi</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-6xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
        Ormi
      </span>
		</h1>
	</hgroup>
</section>

<section class="container mx-auto">
	<div class="w-full flex justify-stretch gap-4">
		<div class="w-full relative">
			<TodoList todos={data.todos} />
		</div>
		<div class="w-full h-96 overflow-y-auto bg-ctp-mantle rounded-md p-2">
			Detail
		</div>
	</div>
</section>