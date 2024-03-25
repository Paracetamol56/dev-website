<script lang="ts">
	import { createTagsInput, melt } from '@melt-ui/svelte';
	import { X } from 'lucide-svelte';

	const {
		elements: { root, input, tag, deleteTrigger, edit },
		states: { tags }
	} = createTagsInput({
		defaultTags: ['Svelte', 'Typescript'],
		unique: true,
		add(tag) {
			return { id: tag, value: tag };
		},
		addOnPaste: true
	});
</script>

<div class="flex flex-col items-start justify-center gap-2">
	<div
		use:melt={$root}
		class="flex w-full flex-row flex-wrap gap-2 rounded-md bg-ctp-surface0 px-3 py-2 text-ctp-text
    shadow-md shadow-ctp-crust focus-within:ring-2 focus-within:ring-ctp-mauve"
	>
		{#each $tags as t}
			<div
				use:melt={$tag(t)}
				class="flex items-center overflow-hidden rounded-md bg-ctp-mauve text-ctp-mantle [word-break:break-word]
        data-[disabled]:bg-ctp-mauve/25 data-[selected]:bg-ctp-mauve/75 data-[disabled]:hover:cursor-default
        data-[disabled]:focus:!outline-none data-[disabled]:focus:!ring-0"
			>
				<span class="flex items-center  px-1.5">
					{t.value}
				</span>
				<button
					use:melt={$deleteTrigger(t)}
					class="flex h-full items-center px-1 enabled:hover:bg-ctp-mauve/75"
				>
					<X size="16" />
				</button>
			</div>
			<div
				use:melt={$edit(t)}
				class="flex items-center overflow-hidden rounded-md px-1.5 [word-break:break-word] data-[invalid-edit]:focus:!ring-ctp-red"
			/>
		{/each}

		<input
			use:melt={$input}
			type="text"
			placeholder="Enter labels..."
			class="min-w-[4.5rem] shrink grow basis-0 border-0 bg-transparent text-ctp-text outline-none focus:!ring-0 data-[invalid]:text-ctp-red"
		/>
	</div>
</div>
