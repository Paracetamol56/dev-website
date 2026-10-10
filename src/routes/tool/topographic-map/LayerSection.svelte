<script lang="ts">
	import MeltCheckbox from '$lib/components/MeltCheckbox.svelte';
	import type { Writable } from 'svelte/store';

	export let name: string;
	export let label: string;
	export let enabled: Writable<boolean>;
</script>

<section class="flex flex-col gap-4 rounded-md bg-ctp-base p-4 shadow-md shadow-ctp-crust">
	<header class="flex items-center justify-between">
		<h3 class="text-lg font-semibold text-ctp-text">{label}</h3>
		<MeltCheckbox name="layer-{name}" label={$enabled ? 'Shown' : 'Hidden'} checked={enabled} />
	</header>
	<!-- A hidden layer's settings are dimmed and cannot be reached -->
	<div
		class="flex flex-col gap-4 transition-opacity {$enabled ? '' : 'opacity-40'}"
		inert={!$enabled}
	>
		<slot />
	</div>
</section>
