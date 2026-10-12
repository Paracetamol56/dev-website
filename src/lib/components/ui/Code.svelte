<script lang="ts">
	import { TextColor, TextFont, TextSize } from './typography';

	export let size: TextSize = TextSize.Sm;
	export let color: TextColor = TextColor.Accent;
</script>

<code
	class="{TextFont.Mono} {size} {color} rounded-md bg-ctp-mantle px-2 py-0.5 shadow-md shadow-ctp-crust"
>
	<slot />
</code>
