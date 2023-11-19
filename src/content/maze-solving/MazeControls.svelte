<script lang="ts">
	import Button from "$lib/components/Button.svelte";
	import MeltSelect from "$lib/components/MeltSelect.svelte";
	import { RedoDot, Rocket } from "lucide-svelte";
	import type { Maze } from "./maze";
	import type { Writable } from "svelte/store";
	import type { SelectOption } from "@melt-ui/svelte";
	import type { Readable } from "svelte/motion";
	import type { Frontier } from "./utils";

  export let maze: Readable<Maze<Frontier>>;
  export let presetMaze: Writable<number>;
  export let algorithm: Writable<SelectOption<string>>;
</script>

<div class="mb-8 p-4 rounded-md bg-ctp-mantle shadow-md shadow-ctp-crust flex flex-col items-start md:flex-row md:items-end justify-center gap-2">
  <div class="flex flex-col gap-1">
    <span class="text-ctp-text font-semibold text-sm">Load a preset maze</span>
    <div class="flex items-center gap-2">
      <Button on:click={() => presetMaze.set(1)}>
        <span>1</span>
      </Button>
      <Button on:click={() => presetMaze.set(2)}>
        <span>2</span>
      </Button>
      <Button on:click={() => presetMaze.set(3)}>
        <span>3</span>
      </Button>
      <Button on:click={() => presetMaze.set(4)}>
        <span>4</span>
      </Button>
    </div>
  </div>
  <MeltSelect name="Search algorithm" options={["Depth-first search", "Breadth-first search", "Greedy best-first search", "A* search"]} value={algorithm} />
  <Button
    on:click={() => $maze.aiStep()}
  >
    <span>Single step</span>
    <RedoDot size="18" />
  </Button>
  <Button
    on:click={() => $maze.aiSolve()}
  >
    <span>Solve</span>
    <Rocket size="18" />
  </Button>
</div>