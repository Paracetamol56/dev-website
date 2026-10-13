<script lang="ts">
	import { createSelect, melt, type SelectOption } from '@melt-ui/svelte';
	import { Check, ChevronDown } from 'lucide-svelte';
	import { writable, type Writable } from 'svelte/store';
	import { fade } from 'svelte/transition';
	import { labelClass } from './styles';
	import { TextColor, TextSize } from './typography';

	export let value: Writable<SelectOption<string>> = writable({ value: '', label: '' });
	export let options: (string | SelectOption<string>)[];
	export let label: string;
	export let hideLabel: boolean = false;
	export let placeholder: string = 'Select an option';
	export let disabled: boolean = false;

	const {
		elements: { trigger, menu, option, label: melted },
		states: { selectedLabel, open },
		helpers: { isSelected }
	} = createSelect({
		selected: value,
		forceVisible: true,
		disabled,
		positioning: { placement: 'bottom', fitViewport: true, sameWidth: true }
	});

	$: items = options.map((option) =>
		typeof option === 'string' ? { value: option, label: option } : option
	);
</script>

<div class="flex flex-col gap-1">
	<!-- svelte-ignore a11y_label_has_associated_control -->
	<label use:melt={$melted} class="{labelClass} {hideLabel ? 'sr-only' : ''}">{label}</label>
	<button
		use:melt={$trigger}
		class="flex min-w-[220px] items-center justify-between gap-2 rounded-md bg-ctp-surface0 px-3 py-1 text-left
			shadow-md shadow-ctp-crust transition-opacity {TextSize.Base} {TextColor.Default}
			hover:opacity-90 focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve
			disabled:cursor-not-allowed disabled:opacity-50"
	>
		<span class="truncate {$selectedLabel ? '' : TextColor.Subtle}">
			{$selectedLabel || placeholder}
		</span>
		<ChevronDown class="square-5 shrink-0" />
	</button>
	{#if $open}
		<div
			use:melt={$menu}
			class="z-50 flex max-h-[300px] flex-col gap-1 overflow-y-auto rounded-md bg-ctp-surface0 p-1 shadow-md shadow-ctp-crust focus:!ring-0"
			transition:fade={{ duration: 150 }}
		>
			{#each items as item}
				<div
					use:melt={$option(item)}
					class="relative cursor-pointer rounded-md py-1 pl-8 pr-4 {TextColor.Default}
						data-[highlighted]:bg-ctp-mauve/25 data-[selected]:bg-ctp-mauve/25
						data-[highlighted]:text-ctp-mauve data-[selected]:text-ctp-mauve"
				>
					{#if $isSelected(item.value)}
						<Check class="absolute left-2 top-1/2 square-4 -translate-y-1/2 {TextColor.Accent}" />
					{/if}
					{item.label}
				</div>
			{/each}
		</div>
	{/if}
</div>
