<script lang="ts">
	import { Heading, Text, TextColor, TextFont, TextSize } from '$lib/components/ui';

	export let name: string;
	export let legacy: boolean = false;
	export let wide: boolean = false;
	export let path: string = `$lib/components/${legacy ? '' : 'ui/'}${name}.svelte`;
</script>

<div class="flex flex-col gap-4 {wide ? 'md:col-span-2 xl:col-span-3' : ''}">
	<header>
		<Heading level={3} size={TextSize.Lg}>{name}</Heading>
		<Text size={TextSize.Xs} color={TextColor.Muted} font={TextFont.Mono}>
			{path}
		</Text>
	</header>
	<div class="flex flex-col gap-4">
		<slot />
	</div>
	{#if $$slots.state}
		<Text
			size={TextSize.Xs}
			color={TextColor.Muted}
			font={TextFont.Mono}
			class="break-all rounded-md bg-ctp-mantle px-3 py-2"
		>
			<slot name="state" />
		</Text>
	{/if}
</div>
