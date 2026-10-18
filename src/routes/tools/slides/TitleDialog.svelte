<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { writable, type Writable } from 'svelte/store';
	import Button from '$lib/components/ui/Button.svelte';
	import Dialog from '$lib/components/ui/Dialog.svelte';
	import TextInput from '$lib/components/ui/TextInput.svelte';
	import { Tone } from '$lib/components/ui/styles';

	export let open: Writable<boolean>;
	export let heading: string;
	export let confirm: string;
	export let initial = '';
	export let loading = false;

	const dispatch = createEventDispatcher<{ submit: string }>();
	const title = writable(initial);
	let error = '';

	$: if ($open) {
		title.set(initial);
		error = '';
	}

	function submit() {
		const value = $title.trim();
		if (value.length === 0) error = 'Give the deck a title';
		else if (value.length > 100) error = 'The title must be 100 characters at most';
		else dispatch('submit', value);
	}
</script>

<Dialog {open} title={heading}>
	<form class="flex flex-col gap-6" on:submit|preventDefault={submit}>
		<TextInput label="Title" value={title} {error} on:input={() => (error = '')} />
		<div class="flex justify-end gap-2">
			<Button tone={Tone.Neutral} on:click={() => open.set(false)}>Cancel</Button>
			<Button type="submit" {loading}>{confirm}</Button>
		</div>
	</form>
</Dialog>
