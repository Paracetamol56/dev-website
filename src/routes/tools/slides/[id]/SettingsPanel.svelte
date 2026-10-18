<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { writable } from 'svelte/store';
	import Select from '$lib/components/ui/Select.svelte';
	import Slider from '$lib/components/ui/Slider.svelte';
	import Switch from '$lib/components/ui/Switch.svelte';
	import {
		defaultEditorSettings,
		easings,
		highlightThemes,
		indents,
		sizes,
		speeds,
		themes,
		transitions,
		type DeckSettings,
		type EditorSettings,
		type Option
	} from '$lib/components/deck/settings';

	export let settings: DeckSettings;
	export let disabled = false;

	const dispatch = createEventDispatcher<{ change: DeckSettings }>();

	const pick = (options: Option[], value: string) =>
		options.find((option) => option.value === value) ?? { value, label: value };

	const sizeOptions: Option[] = sizes.map(({ width, height, label }) => ({
		value: `${width}x${height}`,
		label
	}));
	const currentSize = `${settings.width}x${settings.height}`;
	if (!sizeOptions.some((option) => option.value === currentSize)) {
		sizeOptions.push({
			value: currentSize,
			label: `Custom (${settings.width} × ${settings.height})`
		});
	}

	const theme = writable(pick(themes, settings.theme));
	const size = writable(pick(sizeOptions, currentSize));
	const highlightTheme = writable(pick(highlightThemes, settings.highlightTheme));
	const transition = writable(pick(transitions, settings.transition));
	const transitionSpeed = writable(pick(speeds, settings.transitionSpeed));
	const backgroundTransition = writable(pick(transitions, settings.backgroundTransition));
	const controls = writable(settings.controls);
	const progress = writable(settings.progress);
	const slideNumber = writable(settings.slideNumber);
	const center = writable(settings.center);
	const loop = writable(settings.loop);
	const autoAnimate = writable(settings.autoAnimate);
	const autoAnimateEasing = writable(pick(easings, settings.autoAnimateEasing));
	const autoAnimateDuration = writable([settings.autoAnimateDuration]);
	const markdown = writable(settings.markdown);
	const highlight = writable(settings.highlight);
	const editor = settings.editor ?? defaultEditorSettings;
	const vim = writable(editor.vim);
	const lineWrapping = writable(editor.lineWrapping);
	const lineNumbers = writable(editor.lineNumbers);
	const fontSize = writable([editor.fontSize]);
	const indent = writable(pick(indents, editor.indent));

	$: [width, height] = $size.value.split('x').map(Number);
	$: next = {
		theme: $theme.value,
		transition: $transition.value,
		transitionSpeed: $transitionSpeed.value,
		backgroundTransition: $backgroundTransition.value,
		controls: $controls,
		progress: $progress,
		slideNumber: $slideNumber,
		center: $center,
		loop: $loop,
		autoAnimate: $autoAnimate,
		autoAnimateEasing: $autoAnimateEasing.value,
		autoAnimateDuration: Math.round($autoAnimateDuration[0] * 10) / 10,
		width,
		height,
		markdown: $markdown,
		highlight: $highlight,
		highlightTheme: $highlightTheme.value,
		editor: {
			vim: $vim,
			lineWrapping: $lineWrapping,
			lineNumbers: $lineNumbers,
			fontSize: $fontSize[0],
			indent: $indent.value as EditorSettings['indent']
		}
	} satisfies DeckSettings;
	let known = JSON.stringify(settings);

	function adopt(incoming: DeckSettings) {
		theme.set(pick(themes, incoming.theme));
		size.set(pick(sizeOptions, `${incoming.width}x${incoming.height}`));
		highlightTheme.set(pick(highlightThemes, incoming.highlightTheme));
		transition.set(pick(transitions, incoming.transition));
		transitionSpeed.set(pick(speeds, incoming.transitionSpeed));
		backgroundTransition.set(pick(transitions, incoming.backgroundTransition));
		controls.set(incoming.controls);
		progress.set(incoming.progress);
		slideNumber.set(incoming.slideNumber);
		center.set(incoming.center);
		loop.set(incoming.loop);
		autoAnimate.set(incoming.autoAnimate);
		autoAnimateEasing.set(pick(easings, incoming.autoAnimateEasing));
		autoAnimateDuration.set([incoming.autoAnimateDuration]);
		markdown.set(incoming.markdown);
		highlight.set(incoming.highlight);
		const editor = incoming.editor ?? defaultEditorSettings;
		vim.set(editor.vim);
		lineWrapping.set(editor.lineWrapping);
		lineNumbers.set(editor.lineNumbers);
		fontSize.set([editor.fontSize]);
		indent.set(pick(indents, editor.indent));
	}

	$: if (JSON.stringify(settings) !== known) {
		known = JSON.stringify(settings);
		adopt(settings);
	}
	$: if (JSON.stringify(next) !== known) {
		known = JSON.stringify(next);
		dispatch('change', next);
	}
</script>

{#if disabled}
	<p class="mx-4 mt-4 rounded-md bg-ctp-mantle px-3 py-2 text-sm text-ctp-subtext0">
		Only the owner and the editors can change the settings.
	</p>
{/if}
<form
	class="flex flex-col gap-10 p-4 transition-opacity {disabled ? 'opacity-60' : ''}"
	inert={disabled}
	on:submit|preventDefault
>
	<section class="flex flex-col gap-4">
		<h3 class="text-lg font-semibold text-ctp-text">Appearance</h3>
		<div class="fields">
			<Select label="Theme" options={themes} value={theme} />
			<Select label="Slide size" options={sizeOptions} value={size} />
		</div>
		<p class="text-xs text-ctp-subtext0">
			Slides are scaled to fit the screen, the size sets their proportions and text size.
		</p>
	</section>

	<section class="flex flex-col gap-4">
		<h3 class="text-lg font-semibold text-ctp-text">Transitions</h3>
		<div class="fields">
			<Select label="Slide transition" options={transitions} value={transition} />
			<Select label="Speed" options={speeds} value={transitionSpeed} />
			<Select label="Background transition" options={transitions} value={backgroundTransition} />
		</div>
		<!-- Its settings are dimmed and cannot be reached while auto-animate is off -->
		<section class="flex flex-col gap-4 rounded-md bg-ctp-mantle p-4 shadow-md shadow-ctp-crust">
			<header class="flex flex-wrap items-center justify-between gap-2">
				<h4 class="font-semibold text-ctp-text">Auto-animate</h4>
				<Switch
					name="deck-auto-animate"
					label="Animate between matching slides"
					hideLabel
					checked={autoAnimate}
				/>
			</header>
			<div
				class="flex flex-col gap-4 transition-opacity {$autoAnimate ? '' : 'opacity-40'}"
				inert={!$autoAnimate}
			>
				<p class="text-xs text-ctp-subtext0">
					Elements found on two following slides with the <code>data-auto-animate</code> attribute move
					from one to the other.
				</p>
				<div class="fields">
					<Select label="Easing" options={easings} value={autoAnimateEasing} />
					<Slider
						label="Duration"
						min={0.1}
						max={5}
						step={0.1}
						defaultValue={1}
						value={autoAnimateDuration}
						format={(value) => `${value.toFixed(1)} s`}
					/>
				</div>
			</div>
		</section>
	</section>

	<section class="flex flex-col gap-4">
		<h3 class="text-lg font-semibold text-ctp-text">Navigation</h3>
		<div class="fields">
			<Switch name="deck-controls" label="Arrows" checked={controls} />
			<Switch name="deck-progress" label="Progress bar" checked={progress} />
			<Switch name="deck-slide-number" label="Slide numbers" checked={slideNumber} />
			<Switch name="deck-center" label="Center slides vertically" checked={center} />
			<Switch name="deck-loop" label="Loop back to the first slide" checked={loop} />
		</div>
	</section>

	<section class="flex flex-col gap-4">
		<h3 class="text-lg font-semibold text-ctp-text">Plugins</h3>
		<div class="flex flex-col gap-1">
			<Switch name="deck-markdown" label="Markdown" checked={markdown} />
			<p class="pl-12 text-xs text-ctp-subtext0">
				Write a slide in Markdown with <code>&lt;section data-markdown&gt;</code>.
			</p>
		</div>
		<div class="flex flex-col gap-1">
			<Switch name="deck-highlight" label="Code highlighting" checked={highlight} />
			<p class="pl-12 text-xs text-ctp-subtext0">
				Colours the code in <code>&lt;pre&gt;&lt;code&gt;</code> blocks.
			</p>
		</div>
		<div class="pl-12 transition-opacity {$highlight ? '' : 'opacity-40'}" inert={!$highlight}>
			<Select label="Code colours" options={highlightThemes} value={highlightTheme} />
		</div>
	</section>

	<section class="flex flex-col gap-4">
		<h3 class="text-lg font-semibold text-ctp-text">Editor</h3>
		<p class="text-xs text-ctp-subtext0">The code editor of everyone working on this deck.</p>
		<div class="fields">
			<Switch name="deck-editor-vim" label="Vim mode" checked={vim} />
			<Switch name="deck-editor-wrapping" label="Wrap long lines" checked={lineWrapping} />
			<Switch name="deck-editor-line-numbers" label="Line numbers" checked={lineNumbers} />
		</div>
		<div class="fields">
			<Select label="Indent with" options={indents} value={indent} />
			<Slider
				label="Font size"
				min={10}
				max={24}
				step={1}
				defaultValue={defaultEditorSettings.fontSize}
				value={fontSize}
				format={(value) => `${value} px`}
			/>
		</div>
	</section>
</form>

<style lang="postcss">
	.fields {
		@apply grid grid-cols-[repeat(auto-fill,minmax(14rem,1fr))] items-end gap-4;
	}
</style>
