<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { writable } from 'svelte/store';
	import Select from '$lib/components/ui/Select.svelte';
	import Slider from '$lib/components/ui/Slider.svelte';
	import Switch from '$lib/components/ui/Switch.svelte';
	import {
		easings,
		highlightThemes,
		sizes,
		speeds,
		themes,
		transitions,
		type DeckSettings,
		type Option
	} from '$lib/components/deck/settings';

	export let settings: DeckSettings;

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
		highlightTheme: $highlightTheme.value
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

<form class="flex flex-col gap-8 p-4" on:submit|preventDefault>
	<fieldset class="flex flex-col gap-4">
		<legend class="mb-4 text-sm font-bold uppercase tracking-wide text-ctp-subtext0">
			Appearance
		</legend>
		<Select label="Theme" options={themes} value={theme} />
		<Select label="Slide size" options={sizeOptions} value={size} />
		<Select label="Code colours" options={highlightThemes} value={highlightTheme} />
	</fieldset>

	<fieldset class="flex flex-col gap-4">
		<legend class="mb-4 text-sm font-bold uppercase tracking-wide text-ctp-subtext0">
			Transitions
		</legend>
		<Select label="Slide transition" options={transitions} value={transition} />
		<Select label="Speed" options={speeds} value={transitionSpeed} />
		<Select label="Background transition" options={transitions} value={backgroundTransition} />
	</fieldset>

	<fieldset class="flex flex-col gap-4">
		<legend class="mb-4 text-sm font-bold uppercase tracking-wide text-ctp-subtext0">
			Navigation
		</legend>
		<Switch name="deck-controls" label="Arrows" checked={controls} />
		<Switch name="deck-progress" label="Progress bar" checked={progress} />
		<Switch name="deck-slide-number" label="Slide numbers" checked={slideNumber} />
		<Switch name="deck-center" label="Center slides vertically" checked={center} />
		<Switch name="deck-loop" label="Loop back to the first slide" checked={loop} />
	</fieldset>

	<fieldset class="flex flex-col gap-4">
		<legend class="mb-4 text-sm font-bold uppercase tracking-wide text-ctp-subtext0">
			Auto-animate
		</legend>
		<Switch
			name="deck-auto-animate"
			label="Animate between matching slides"
			checked={autoAnimate}
		/>
		<Select label="Easing" options={easings} value={autoAnimateEasing} disabled={!$autoAnimate} />
		<Slider
			label="Duration"
			min={0.1}
			max={5}
			step={0.1}
			value={autoAnimateDuration}
			format={(value) => `${value.toFixed(1)} s`}
			disabled={!$autoAnimate}
		/>
	</fieldset>

	<fieldset class="flex flex-col gap-4">
		<legend class="mb-4 text-sm font-bold uppercase tracking-wide text-ctp-subtext0">
			Plugins
		</legend>
		<Switch name="deck-markdown" label="Markdown sections" checked={markdown} />
		<Switch name="deck-highlight" label="Code highlighting" checked={highlight} />
	</fieldset>
</form>
