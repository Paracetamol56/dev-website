<script lang="ts">
  import type { PageData } from "./$types";
	import Diagram from "./Diagram.svelte";

  export let data: PageData;
</script>

<div class="container mx-auto">
  <h1 class="my-16 text-6xl font-semibold text-center">
    Herzsprung-Russell diagram
  </h1>
  <Diagram data={data.stars} />
</div>