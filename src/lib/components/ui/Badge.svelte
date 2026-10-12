<script lang="ts">
	import { Tone, toneSoft } from './styles';
	import { TextSize, TextWeight } from './typography';

	export let tone: Tone = Tone.Accent;
	export let size: TextSize = TextSize.Xs;
	let className = '';
	export { className as class };
</script>

<span
	class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 {size} {TextWeight.Medium} {toneSoft[
		tone
	]} {className}"
>
	<slot />
</span>
