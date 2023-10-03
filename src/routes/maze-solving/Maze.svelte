<script lang="ts">
	import { RedoDot, Rocket } from "lucide-svelte";
	import { Maze } from "./maze";
	import type { Node, State } from "./utils";

  export let mazeString = "#####B#/##### #/####  #/#### ##/     ##/A######";

  let maze = new Maze(mazeString);
  let lastNode: Node = maze.move();
  let solved: boolean = false;
  let solvable: boolean = true;

  const getSolvedPath = (endNode: Node) => {
    let solvedPath: State[] = [];
    let node: Node = endNode;
    while (node.parent) {
      solvedPath.push(node.state);
      maze.cells[node.state.x][node.state.y].solution = true;
      node = node.parent;
    }
    solvedPath.push(node.state);
    maze.cells[node.state.x][node.state.y].solution = true;
    return solvedPath;
  }

  const handleAIStep = () => {
    if (solved) return;
    
    try {
      lastNode = maze.move();
    } catch (e) {
      console.log(e);
      if (e instanceof Error) {
        if (e.message == "No solution") {
          solvable = false;
        }
      }
    }

    if (lastNode.state.x == maze.end.x && lastNode.state.y == maze.end.y) {
      solved = true;
      getSolvedPath(lastNode);
    }
  };

  const handleAISolve = () => {
    if (solved) return;

    try {
      while (true) {
        lastNode = maze.move();
        if (lastNode.state.x == maze.end.x && lastNode.state.y == maze.end.y) {
          solved = true;
          getSolvedPath(lastNode);
          break;
        }
      }
    } catch (e) {
      console.log(e);
      if (e instanceof Error) {
        if (e.message == "No solution") {
          solvable = false;
        }
      }
    }
  };

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

          if (i == lastNode.state.x && j == lastNode.state.y) {
            const stateElement = document.createElement("div");
            stateElement.classList.add("rounded-full", "bg-ctp-mauve", "aspect-square", "w-1/2", "h-1/2");
            cell.element!.appendChild(stateElement);
          } else if (cell.visited) {
            if (cell.solution) {
              const stateElement = document.createElement("div");
              stateElement.classList.add("rounded-full", "bg-ctp-green", "w-1/2", "h-1/2");
              cell.element!.appendChild(stateElement);
            } else {
              const stateElement = document.createElement("div");
              stateElement.classList.add("rounded-full", "bg-ctp-yellow", "w-1/2", "h-1/2");
              cell.element!.appendChild(stateElement);
            }
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
      <h2 class="mb-4 text-4xl font-semibold text-ctp-green">
        Solved!
      </h2>
      <p class="mb-4">
        Explored states: {maze.explored.length}
      </p>
      <button
        class="rounded-md bg-ctp-mauve px-3 py-1 font-medium text-ctp-mantle hover:opacity-75 active:opacity-50 transition-opacity"
        on:click={() => {
          maze = new Maze(mazeString);
          lastNode = maze.move();
          solved = false;
        }}
      >
        Reset
      </button>
    </div>
  {/if}
  {#if !solvable}
    <div class="absolute rounded-md inset-0 flex flex-col justify-center items-center bg-ctp-red/25">
      <h2 class="mb-4 text-4xl font-semibold text-ctp-red">
        No solution
      </h2>
      <p class="mb-4">
        Explored states: {maze.explored.length}
      </p>
      <button
        class="rounded-md bg-ctp-mauve px-3 py-1 font-medium text-ctp-mantle hover:opacity-75 active:opacity-50 transition-opacity"
        on:click={() => {
          maze = new Maze(mazeString);
          lastNode = maze.move();
          solved = false;
          solvable = true;
        }}
      >
        Reset
      </button>
    </div>
  {/if}
</div>

<div class="my-8 flex justify-center gap-2">
  <button
    class="flex items-center rounded-md bg-ctp-mauve px-3 py-1 font-medium text-ctp-mantle hover:opacity-75 active:opacity-50 transition-opacity"
    on:click={() => handleAIStep()}
  >
    Single step&nbsp;<RedoDot size="18" />
  </button>
  <button
    class="flex items-center rounded-md bg-ctp-mauve px-3 py-1 font-medium text-ctp-mantle hover:opacity-75 active:opacity-50 transition-opacity"
    on:click={() => handleAISolve()}
  >
    Solve&nbsp;<Rocket size="18" />
  </button>
</div>
