<script lang="ts">
	import { Loader2 } from 'lucide-svelte';
	import { ButtonVariant, Tone, toneSoft, toneSolid, toneText } from './styles';
	import { TextSize, TextWeight } from './typography';

	export let type: 'button' | 'submit' | 'reset' = 'button';
	export let href: string | undefined = undefined;
	export let variant: ButtonVariant = ButtonVariant.Solid;
	export let tone: Tone = Tone.Accent;
	export let size: TextSize = TextSize.Base;
	export let disabled: boolean = false;
	export let loading: boolean = false;
	let className = '';
	export { className as class };

	$: inactive = disabled || loading;
	$: look = {
		[ButtonVariant.Solid]: `${toneSolid[tone]} shadow-md shadow-ctp-crust`,
		[ButtonVariant.Soft]: toneSoft[tone],
		[ButtonVariant.Ghost]: `${toneText[tone]} hover:bg-ctp-surface0`
	}[variant];
</script>

<svelte:element
	this={href ? 'a' : 'button'}
	href={inactive ? undefined : href}
	type={href ? undefined : type}
	disabled={href ? undefined : inactive}
	aria-disabled={href && inactive ? 'true' : undefined}
	aria-busy={loading ? 'true' : undefined}
	class="{href
		? 'inline-flex'
		: 'flex'} items-center justify-center gap-1 rounded-md px-[0.75em] py-[0.25em]
		{size} {TextWeight.Semibold} {look} transition-opacity
		focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve focus-visible:ring-offset-2 focus-visible:ring-offset-ctp-base
		{inactive ? 'cursor-not-allowed opacity-50' : 'hover:opacity-80 active:opacity-60'} {className}"
	on:click
	{...$$restProps}
>
	{#if loading}
		<Loader2 class="animate-spin" size="1em" />
	{/if}
	<slot />
</svelte:element>
