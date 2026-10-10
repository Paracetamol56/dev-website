<script lang="ts">
	export let size: number = 24;
	export let color: string = 'currentColor';
	export let strokeWidth: number = 2;
	export let absoluteStrokeWidth: boolean = false;
	export let defaultClass: string = 'lucide-icon';
</script>

<svg
	xmlns="http://www.w3.org/2000/svg"
	width={size}
	height={size}
	viewBox="0 0 24 24"
	fill="none"
	stroke={color}
	stroke-width={absoluteStrokeWidth ? (Number(strokeWidth) * 24) / Number(size) : strokeWidth}
	stroke-linecap="round"
	stroke-linejoin="round"
	class={defaultClass}
>
	<path d="M18.36 5.64A9 9 0 1 0 21 12h-9" />
</svg>
