<script lang="ts">
	import { Divider, Heading, Text, TextColor } from '$lib/components/ui';

	export let title: string;
	export let description: string = '';
</script>

<section id={title.toLowerCase()} class="container mx-auto mb-20 scroll-mt-24">
	<Heading level={2}>{title}</Heading>
	{#if description}
		<Text color={TextColor.Muted} class="mt-1">{description}</Text>
	{/if}
	<Divider class="mb-8 mt-4" />
	<div class="grid grid-cols-1 gap-x-12 gap-y-10 md:grid-cols-2 xl:grid-cols-3">
		<slot />
	</div>
</section>
