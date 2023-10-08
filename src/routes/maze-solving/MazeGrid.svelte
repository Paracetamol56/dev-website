<script lang="ts">
	import type { Cell, Maze } from "./maze";
	import Button from "$lib/components/Button.svelte";
	import type { Readable } from "svelte/store";

  export let maze: Readable<Maze>;
  let solved = false;
  let solvable = true;
  let cells: Cell[][] = [];
  $: cells = $maze.cells;

  export const updateGrid = () => {
    if ($maze.solved) {
      solved = true;
    } else if (!$maze.solvable) {
      solvable = false;
    }
    cells = $maze.cells;
  };
</script>

<div class="mb-8 relative p-1 md:p-2 rounded-md bg-ctp-mantle">
  <div class="mx-auto lg:w-2/3 grid gap-1 md:gap-2" style={`grid-template-columns: repeat(${$maze.width}, minmax(0, 1fr)); grid-template-rows: repeat(${$maze.height}, minmax(0, 1fr));`}>
    {#each cells as row, i}
      {#each row as cell, j}
        <div
          class="aspect-square rounded-md flex justify-center items-center
                {cell.wall ? 'bg-ctp-crust' : 'bg-ctp-base'}
                {i == $maze.start.x && j == $maze.start.y ? 'border-2 border-ctp-blue' : ''}
                {i == $maze.end.x && j == $maze.end.y ? 'border-2 border-ctp-green' : ''}"
        >
          {#if i == $maze.lastNode?.state.x && j == $maze.lastNode?.state.y}
            <div class="rounded-full bg-ctp-mauve aspect-square w-1/2 h-1/2"></div>
          {:else if cell.visited}
            {#if cell.solution}
              <div class="rounded-full bg-ctp-green w-1/2 h-1/2"></div>
            {:else}
              <div class="rounded-full bg-ctp-yellow w-1/2 h-1/2"></div>
            {/if}
          {/if}
        </div>
      {/each}
    {/each}
  </div>
  
  {#if solved}
    <div class="absolute rounded-md inset-0 flex flex-col justify-center items-center bg-ctp-green/25">
      <h2 class="mb-4 text-4xl font-semibold text-ctp-green">
        Solved!
      </h2>
      <p class="mb-4 font-semibold text-ctp-mantle">
        Explored states: {$maze.explored.length}
      </p>
      <Button
        on:click={() => {
          solved = false;
          $maze.reset();
        }}
      >
        <span>Reset</span>
      </Button>
    </div>
  {/if}
  {#if !solvable}
    <div class="absolute rounded-md inset-0 flex flex-col justify-center items-center bg-ctp-red/25">
      <h2 class="mb-4 text-4xl font-semibold text-ctp-red">
        No solution
      </h2>
      <p class="mb-4 font-semibold text-ctp-mantle">
        Explored states: {$maze.explored.length}
      </p>
      <Button
        on:click={() => {
          solvable = true;
          $maze.reset();
        }}
      >
        <span>Reset</span>
      </Button>
    </div>
  {/if}
</div>