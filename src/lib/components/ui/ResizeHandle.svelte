<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Writable } from 'svelte/store';

	export let value: Writable<number>;
	export let min: number;
	export let max: number;
	export let defaultValue: number;
	export let label: string;
	export let orientation: 'vertical' | 'horizontal' = 'vertical';
	export let unit: 'fraction' | 'px' = 'fraction';
	export let invert = false;
	export let step = unit === 'fraction' ? 0.02 : 8;
	let className = '';
	export { className as class };

	const dispatch = createEventDispatcher<{ dragging: boolean }>();

	let element: HTMLDivElement;
	let dragging = false;
	let start = { pointer: 0, value: 0 };

	const clamp = (next: number) => Math.min(max, Math.max(min, next));
	const position = (event: PointerEvent) =>
		orientation === 'vertical' ? event.clientX : event.clientY;

	function down(event: PointerEvent) {
		if (event.button !== 0) return;
		event.preventDefault();
		element.setPointerCapture(event.pointerId);
		start = { pointer: position(event), value: $value };
		dragging = true;
		document.body.classList.add('select-none');
		dispatch('dragging', true);
	}

	function move(event: PointerEvent) {
		if (!dragging) return;
		if (unit === 'fraction') {
			const rect = element.parentElement?.getBoundingClientRect();
			if (!rect) return;
			const size = orientation === 'vertical' ? rect.width : rect.height;
			const offset = position(event) - (orientation === 'vertical' ? rect.left : rect.top);
			value.set(clamp(offset / size));
		} else {
			const distance = (position(event) - start.pointer) * (invert ? -1 : 1);
			value.set(clamp(start.value + distance));
		}
	}

	function up(event: PointerEvent) {
		if (!dragging) return;
		element.releasePointerCapture(event.pointerId);
		dragging = false;
		document.body.classList.remove('select-none');
		dispatch('dragging', false);
	}

	function key(event: KeyboardEvent) {
		const smaller = orientation === 'vertical' ? 'ArrowLeft' : 'ArrowUp';
		const larger = orientation === 'vertical' ? 'ArrowRight' : 'ArrowDown';
		const direction = invert ? -1 : 1;
		const moves: Record<string, () => number> = {
			[smaller]: () => $value - step * direction,
			[larger]: () => $value + step * direction,
			Home: () => min,
			End: () => max,
			Enter: () => defaultValue
		};
		if (!moves[event.key]) return;
		event.preventDefault();
		value.set(clamp(moves[event.key]()));
	}

	const percent = (number: number) =>
		unit === 'fraction' ? Math.round(number * 100) : Math.round(number);
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex a11y_no_noninteractive_element_interactions -->
<div
	bind:this={element}
	role="separator"
	tabindex="0"
	aria-label={label}
	aria-orientation={orientation}
	aria-valuenow={percent($value)}
	aria-valuemin={percent(min)}
	aria-valuemax={percent(max)}
	data-dragging={dragging ? '' : undefined}
	class="group flex touch-none items-center justify-center focus:outline-none
		{orientation === 'vertical' ? 'w-3 cursor-col-resize' : 'h-3 cursor-row-resize'} {className}"
	on:pointerdown={down}
	on:pointermove={move}
	on:pointerup={up}
	on:pointercancel={up}
	on:keydown={key}
	on:dblclick={() => value.set(defaultValue)}
>
	<span
		class="rounded-full bg-ctp-surface1 transition-colors group-hover:bg-ctp-mauve group-focus-visible:bg-ctp-mauve
			group-data-[dragging]:bg-ctp-mauve {orientation === 'vertical' ? 'h-12 w-1' : 'h-1 w-12'}"
	/>
</div>
