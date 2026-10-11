<script lang="ts">
	import { createTagsInput, melt, type Tag } from '@melt-ui/svelte';
	import { X } from 'lucide-svelte';
	import { writable, type Writable } from 'svelte/store';

	export let tags: Writable<Tag[]> = writable([]);
	export let placeholder: string = 'Enter tags...';
	export let suggest: ((query: string) => Promise<string[]>) | null = null;

	const {
		elements: { root, input, tag, deleteTrigger, edit },
		states: { inputValue }
	} = createTagsInput({
		tags,
		unique: true,
		add(tag) {
			return { id: tag, value: tag };
		},
		addOnPaste: true
	});

	let inputElement: HTMLInputElement;
	let focused = false;
	let suggestions: string[] = [];
	let highlighted = -1;
	let timeout: ReturnType<typeof setTimeout>;
	let lastRequest = 0;

	const loadSuggestions = (query: string) => {
		if (!suggest) return;
		clearTimeout(timeout);
		timeout = setTimeout(() => {
			const request = ++lastRequest;
			suggest(query)
				.then((results) => {
					if (request !== lastRequest) return;
					suggestions = results;
					highlighted = -1;
				})
				.catch(() => (suggestions = []));
		}, 150);
	};

	$: if (focused) loadSuggestions($inputValue);
	$: visibleSuggestions = suggestions.filter(
		(suggestion) => !$tags.some((t) => t.value === suggestion)
	);

	const pick = (suggestion: string) => {
		$tags = [...$tags, { id: suggestion, value: suggestion }];
		inputElement.value = '';
		inputElement.dispatchEvent(new Event('input', { bubbles: true }));
		highlighted = -1;
	};

	const handleKeydown = (event: KeyboardEvent) => {
		if (!focused || visibleSuggestions.length === 0) return;
		if (event.key === 'ArrowDown') {
			event.preventDefault();
			highlighted = (highlighted + 1) % visibleSuggestions.length;
		} else if (event.key === 'ArrowUp') {
			event.preventDefault();
			highlighted = highlighted <= 0 ? visibleSuggestions.length - 1 : highlighted - 1;
		} else if (event.key === 'Enter' && highlighted >= 0) {
			event.preventDefault();
			event.stopPropagation();
			pick(visibleSuggestions[highlighted]);
		} else if (event.key === 'Escape') {
			suggestions = [];
		}
	};
</script>

<!-- svelte-ignore a11y-no-static-element-interactions -->
<div
	class="relative"
	on:keydown|capture={handleKeydown}
	on:focusin={() => (focused = true)}
	on:focusout={() => (focused = false)}
>
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
				<span class="flex items-center px-1.5">
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
			bind:this={inputElement}
			type="text"
			{placeholder}
			class="min-w-[4.5rem] shrink grow basis-0 border-0 bg-transparent text-ctp-text outline-none focus:!ring-0 data-[invalid]:text-ctp-red"
		/>
	</div>

	{#if focused && visibleSuggestions.length > 0}
		<ul
			class="absolute left-0 right-0 top-full z-10 mt-1 max-h-48 overflow-y-auto rounded-md bg-ctp-mantle p-1 shadow-md shadow-ctp-crust"
		>
			{#each visibleSuggestions as suggestion, index}
				<li>
					<button
						type="button"
						tabindex="-1"
						class="w-full rounded-md px-2 py-1 text-left text-sm hover:bg-ctp-mauve/20
							{index === highlighted ? 'bg-ctp-mauve/20' : ''}"
						on:mousedown|preventDefault={() => pick(suggestion)}
					>
						{suggestion}
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>
