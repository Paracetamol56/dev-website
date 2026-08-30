<script lang="ts">
	import { Search } from 'lucide-svelte';
	import TextInput from '$lib/components/TextInput.svelte';
	import type { Writable } from 'svelte/store';

	export let iconName: Writable<string>;
</script>

<div class="relative">
	<Search class="absolute left-3 top-1/2 -translate-y-1/2 text-ctp-subtext0" size="16" />
	<TextInput
		label="Search icons"
		bind:value={iconName}
		placeholder="Search for an icon..."
		class="pl-9"
	/>
</div>
