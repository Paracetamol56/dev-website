<script context="module" lang="ts">
	export type TextLevel = 'small' | 'normal' | 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';
	export type TextFormat = {
		bold: boolean;
		italic: boolean;
		underline: boolean;
		strikeThrough: boolean;
		align: 'left' | 'center' | 'right' | null;
		level: TextLevel | null;
	};
	export type FormatCommand = { command: string; value?: string };
</script>

<script lang="ts">
	import { createSelect, createToolbar, melt, type SelectOption } from '@melt-ui/svelte';
	import {
		AlignCenter,
		AlignLeft,
		AlignRight,
		Bold,
		Check,
		ChevronDown,
		Italic,
		Strikethrough,
		Underline
	} from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { writable } from 'svelte/store';

	export let format: TextFormat;

	const dispatch = createEventDispatcher<{ format: FormatCommand }>();
	const styles = [
		{ value: 'bold', label: 'Bold', icon: Bold },
		{ value: 'italic', label: 'Italic', icon: Italic },
		{ value: 'underline', label: 'Underline', icon: Underline },
		{ value: 'strikeThrough', label: 'Strikethrough', icon: Strikethrough }
	] as const;
	const alignments = [
		{ value: 'left', label: 'Align left', icon: AlignLeft },
		{ value: 'center', label: 'Align center', icon: AlignCenter },
		{ value: 'right', label: 'Align right', icon: AlignRight }
	] as const;
	const levels: { value: TextLevel; label: string; class: string }[] = [
		{ value: 'small', label: 'Small', class: 'text-xs' },
		{ value: 'normal', label: 'Normal', class: 'text-sm' },
		{ value: 'h4', label: 'Heading 4', class: 'text-sm font-bold' },
		{ value: 'h3', label: 'Heading 3', class: 'text-base font-bold' },
		{ value: 'h2', label: 'Heading 2', class: 'text-lg font-bold' },
		{ value: 'h1', label: 'Heading 1', class: 'text-xl font-bold' }
	];
	const levelLabels: Record<TextLevel, string> = {
		small: 'Small',
		normal: 'Normal',
		h1: 'Heading 1',
		h2: 'Heading 2',
		h3: 'Heading 3',
		h4: 'Heading 4',
		h5: 'Heading 5',
		h6: 'Heading 6'
	};

	// The frame reports the formatting after each command, which keeps the pressed states right
	const fontValue = writable<string[]>([]);
	const alignValue = writable<string | undefined>(undefined);
	const levelValue = writable<SelectOption<TextLevel>>({ value: 'normal', label: 'Normal' });

	const {
		elements: { root, separator },
		builders: { createToolbarGroup }
	} = createToolbar();
	const {
		elements: { group: fontGroup, item: fontItem }
	} = createToolbarGroup({ type: 'multiple', value: fontValue });
	const {
		elements: { group: alignGroup, item: alignItem }
	} = createToolbarGroup({ value: alignValue });
	// The menu stays inside the toolbar: a click in it is not a click outside the edited text
	const {
		elements: { trigger, menu, option },
		states: { open },
		helpers: { isSelected },
		options: { disabled: levelDisabled }
	} = createSelect<TextLevel>({
		selected: levelValue,
		portal: null,
		forceVisible: true,
		positioning: { placement: 'bottom-start' },
		onSelectedChange: ({ next }) => {
			if (next) dispatch('format', { command: 'level', value: next.value });
			return next;
		}
	});

	// melt handles Enter and Space itself, so they never become a click
	const activate = (command: FormatCommand) => ({
		click: () => dispatch('format', command),
		keydown: (event: KeyboardEvent) => {
			if (event.key === 'Enter' || event.key === ' ') dispatch('format', command);
		}
	});

	$: fontValue.set(styles.filter((style) => format[style.value]).map((style) => style.value));
	$: alignValue.set(format.align ?? undefined);
	$: if (format.level) levelValue.set({ value: format.level, label: levelLabels[format.level] });
	$: levelDisabled.set(format.level === null);
</script>

<!-- Pressing a button must not take the focus from the text being edited in the frame -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	use:melt={$root}
	aria-label="Text formatting"
	class="flex items-center gap-1.5 rounded-md bg-ctp-base px-1.5 py-1 text-ctp-subtext1 shadow-lg shadow-ctp-crust ring-1 ring-ctp-surface0"
	on:mousedown|preventDefault
>
	<button
		use:melt={$trigger}
		aria-label="Text level"
		title="Text level"
		class="flex w-28 items-center justify-between gap-1 rounded-md px-2 py-1 text-sm transition-colors
			hover:bg-ctp-surface0 hover:text-ctp-text focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve
			disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:bg-transparent"
	>
		<span class="truncate">{format.level ? levelLabels[format.level] : 'Text'}</span>
		<ChevronDown size="14" class="shrink-0" />
	</button>
	{#if $open}
		<div
			use:melt={$menu}
			class="z-20 flex min-w-36 flex-col gap-0.5 rounded-md bg-ctp-base p-1 shadow-lg shadow-ctp-crust ring-1 ring-ctp-surface0 focus:!ring-0"
		>
			{#each levels as level}
				<div
					use:melt={$option({ value: level.value, label: level.label })}
					class="relative cursor-pointer rounded-md py-1 pl-7 pr-3 text-ctp-text
						data-[highlighted]:bg-ctp-mauve/20 data-[highlighted]:text-ctp-mauve {level.class}"
				>
					{#if $isSelected(level.value)}
						<Check size="14" class="absolute left-2 top-1/2 -translate-y-1/2 text-ctp-mauve" />
					{/if}
					{level.label}
				</div>
			{/each}
		</div>
	{/if}
	<div class="w-px self-stretch bg-ctp-surface1" use:melt={$separator} />
	<div class="flex items-center gap-0.5" use:melt={$fontGroup}>
		{#each styles as style}
			<button
				class="item"
				aria-label={style.label}
				title={style.label}
				use:melt={$fontItem(style.value)}
				on:click={activate({ command: style.value }).click}
				on:keydown={activate({ command: style.value }).keydown}
			>
				<svelte:component this={style.icon} size="16" />
			</button>
		{/each}
	</div>
	<div class="w-px self-stretch bg-ctp-surface1" use:melt={$separator} />
	<div class="flex items-center gap-0.5" use:melt={$alignGroup}>
		{#each alignments as alignment}
			<button
				class="item"
				aria-label={alignment.label}
				title={alignment.label}
				use:melt={$alignItem(alignment.value)}
				on:click={activate({ command: 'align', value: alignment.value }).click}
				on:keydown={activate({ command: 'align', value: alignment.value }).keydown}
			>
				<svelte:component this={alignment.icon} size="16" />
			</button>
		{/each}
	</div>
</div>

<style lang="postcss">
	.item {
		@apply rounded-md p-1.5 transition-colors hover:bg-ctp-surface0 hover:text-ctp-text focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve;
	}

	.item[data-state='on'] {
		@apply bg-ctp-mauve/20 text-ctp-mauve;
	}
</style>
