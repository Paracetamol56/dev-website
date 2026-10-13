<script lang="ts">
	import { Loader2 } from 'lucide-svelte';
	import { TextColor } from './typography';

	export let size: number = 32;
	export let color: TextColor = TextColor.Accent;
	export let label: string = 'Loading';
</script>

<span role="status" aria-label={label} class="inline-flex">
	<Loader2 class="animate-spin {color}" {size} />
</span>
