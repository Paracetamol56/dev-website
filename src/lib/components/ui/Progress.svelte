<script lang="ts">
	import { labelClass } from './styles';
	import { TextColor, TextSize } from './typography';

	export let value: number = 0;
	export let max: number = 100;
	export let label: string;
	export let hideLabel: boolean = false;

	$: percentage = max > 0 ? Math.min(100, Math.max(0, (value / max) * 100)) : 0;
</script>

<div class="flex flex-col gap-1">
	<div class="flex items-baseline justify-between gap-2 {hideLabel ? 'sr-only' : ''}">
		<span class={labelClass}>{label}</span>
		<span class="{TextSize.Sm} {TextColor.Muted}">{Math.round(percentage)}%</span>
	</div>
	<div
		role="progressbar"
		aria-label={label}
		aria-valuemin="0"
		aria-valuemax={max}
		aria-valuenow={value}
		class="h-2 w-full overflow-hidden rounded-full bg-ctp-mauve/20"
	>
		<div class="h-full rounded-full bg-ctp-mauve transition-[width]" style:width="{percentage}%" />
	</div>
</div>
