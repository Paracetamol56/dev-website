<script lang="ts">
	import { Tone, toneText } from './styles';

	export let label: string;
	export let type: 'button' | 'submit' | 'reset' = 'button';
	export let tone: Tone = Tone.Neutral;
	export let disabled: boolean = false;
	let className = '';
	export { className as class };
</script>

<button
	{type}
	{disabled}
	aria-label={label}
	class="grid place-items-center rounded-full p-1 {toneText[tone]} transition-colors
		hover:bg-ctp-surface0 focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve
		disabled:cursor-not-allowed disabled:opacity-50 {className}"
	on:click
	{...$$restProps}
>
	<slot />
</button>
