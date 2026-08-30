<script lang="ts">
	import { createRadioGroup, melt } from '@melt-ui/svelte';
	import type { Writable } from 'svelte/store';

	export let options: string[];
	export let required: boolean = false;
	export let disabled: boolean = false;
	export let orientation: 'horizontal' | 'vertical' = 'horizontal';
	export let value: Writable<string>;
	export let name: string = 'radio-group';

	const {
		elements: { root, item, hiddenInput },
		helpers: { isChecked }
	} = createRadioGroup({
		required,
		disabled,
		orientation,
		value
	});
</script>

<div
	use:melt={$root}
	class="flex flex-col gap-3 data-[orientation=horizontal]:flex-row"
	aria-label={name}
>
	{#each options as option}
		<div class="flex items-center gap-2">
			<button
				use:melt={$item(option)}
				class="grid square-5 cursor-default place-items-center rounded-full bg-ctp-surface0 shadow-sm shadow-ctp-crust
                  hover:bg-ctp-surface1 border-2 border-transparent"
				id="{name}-{option}"
				aria-labelledby="{name}-{option}-label"
			>
				{#if $isChecked(option)}
					<div class="square-3 rounded-full bg-ctp-mauve" />
				{/if}
			</button>
			<label
				class="leading-none capitalize font-semibold"
				for="{name}-{option}"
				id="{name}-{option}-label"
			>
				{option}
			</label>
		</div>
	{/each}
	<input {name} use:melt={$hiddenInput} />
</div>
