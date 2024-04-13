<script lang="ts">
	import { writable, type Writable } from "svelte/store";
	import type { PageData } from "./$types";
	import Calendar from "./Calendar.svelte";
	import TodoList from "./TodoList.svelte";

	export let data: PageData;
	const focusedTodoId: Writable<string | null> = writable(null)
	$: focusedTodo = $focusedTodoId === null ? null : data.todos!.find(todo => todo.id === $focusedTodoId)

	$: console.log($focusedTodoId)
</script>



<svelte:head>
	<title>Ormi - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-6xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender"
				>Ormi</span
			>
		</h1>
	</hgroup>
</section>

<section class="container mx-auto mb-32">
	<p class="text-justify max-w-xl mx-auto mb-16">
		Okay so as a developer, I have to make a todo app at some point...<br>
		<strong>Ορμή</strong> means "momentum" in greek, this is because when you dive into a project,
		you have a sudden burst of energy and motivation comming from nowhere, Ormi 
		aims to help you keep that momentum going. The github like calendar below and
		the streak counter are here to motivate you to keep working on your project.
	</p>

	<!--<TodoList todos={data.todos} {focusedTodoId} />-->
</section>

<section class="container mx-auto">
	<Calendar />
</section>