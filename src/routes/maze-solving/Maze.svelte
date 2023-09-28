<script lang="ts">
	import { Maze } from "./maze";
	import type { State } from "./utils";

  export let mazeString = "#####B#/##### #/####  #/#### ##/     ##/A######";

  let maze = new Maze(mazeString);
  let lastMove: State = maze.move();
  let solved = false;
  const handleAIMove = () => {
    if (solved) return;
    
    try {
      lastMove = maze.move();
    } catch (e) {
      console.log(e);
    }

    if (lastMove.x == maze.end.x && lastMove.y == maze.end.y) {
      solved = true;
    }
  }
  $: {
    for (let i = 0; i < maze.height; i++) {
      for (let j = 0; j < maze.width; j++) {
        const cell = maze.cells[i][j];

        if (!cell.element) continue;

        if (cell.wall) {
          cell.element!.classList.add("bg-ctp-crust");
        } else {
          cell.element!.classList.add("bg-ctp-base");

          while (cell.element!.firstChild) {
            cell.element!.removeChild(cell.element!.firstChild);
          }

          if (i == maze.start.x && j == maze.start.y) {
            cell.element!.classList.add("border-2", "border-ctp-blue");
          } else if (i == maze.end.x && j == maze.end.y) {
            cell.element!.classList.add("border-2", "border-ctp-green");
          }

          if (i == lastMove.x && j == lastMove.y) {
            const stateElement = document.createElement("div");
            stateElement.classList.add("rounded-full", "bg-ctp-mauve", "aspect-square", "w-1/2", "h-1/2");
            cell.element!.appendChild(stateElement);
          } else if (cell.visited) {
            const stateElement = document.createElement("div");
            stateElement.classList.add("rounded-full", "bg-ctp-yellow", "w-1/2", "h-1/2");
            cell.element!.appendChild(stateElement);
          }
        }
      }
    }
  }
</script>

<div class="relative p-2 rounded-md bg-ctp-mantle">
  <div class="grid gap-2" style={`grid-template-columns: repeat(${maze.width}, minmax(0, 1fr)); grid-template-rows: repeat(${maze.height}, minmax(0, 1fr));`}>
    {#each maze.cells as row, i}
      {#each row as cell, j}
        <div class="aspect-square rounded-md flex justify-center items-center" bind:this={cell.element}></div>
      {/each}
    {/each}
  </div>
  
  {#if solved}
    <div class="absolute rounded-md inset-0 flex flex-col justify-center items-center bg-ctp-green/25">
      <h2 class="mb-8 text-4xl font-semibold text-ctp-green">
        Solved!
      </h2>
      <button
        class="rounded-md bg-ctp-mauve px-3 py-1 font-medium text-ctp-mantle hover:opacity-75 active:opacity-50 transition-opacity"
        on:click={() => {
          maze = new Maze(mazeString);
          lastMove = maze.move();
          solved = false;
        }}
      >
        Reset
      </button>
    </div>
  {/if}
</div>

<div class="my-8 flex justify-center">
  <button
    class="rounded-md bg-ctp-mauve px-3 py-1 font-medium text-ctp-mantle hover:opacity-75 active:opacity-50 transition-opacity"
    on:click={() => handleAIMove()}
  >
    AI move
  </button>
</div>
