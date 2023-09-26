<script lang="ts">
  import { display } from 'mathlifier';
	import Plot from "./Plot.svelte";

  $: relu = (x: number) => Math.max(0, x);
</script>

<h2 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">
  Options
</h2>
{@html display(`\\phi(x)=max(0, x)`)}

<h2 class="mt-8 mb-4 text-lg font-semibold text-ctp-lavender">
  Result
</h2>
<Plot fx={relu} />
