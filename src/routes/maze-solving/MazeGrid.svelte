<script lang="ts">
	import { RedoDot, Rocket } from "lucide-svelte";
	import type { Maze } from "./maze";
	import type { Node, State } from "./utils";
	import MeltSelect from "$lib/components/MeltSelect.svelte";
	import Button from "$lib/components/Button.svelte";
	import { writable, type Writable } from "svelte/store";
	import type { SelectOption } from "@melt-ui/svelte";

  export let maze: Maze;
  export let loadMaze: (index: number) => void;

  let algorithm: Writable<SelectOption<string>> = writable({value: "Depth-first search", label: "Depth-first search"});
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

<div class="mb-8 p-4 rounded-md bg-ctp-mantle flex flex-col items-start md:flex-row md:items-end justify-center gap-2">
  <div class="flex flex-col gap-1">
    <p class="block text-ctp-text font-semibold text-sm">Load a preset maze</p>
    <div class="flex items-center gap-2">
      <Button on:click={() => loadMaze(1)}>
        <span>1</span>
      </Button>
      <Button on:click={() => loadMaze(2)}>
        <span>2</span>
      </Button>
      <Button on:click={() => loadMaze(3)}>
        <span>3</span>
      </Button>
      <Button on:click={() => loadMaze(4)}>
        <span>4</span>
      </Button>
    </div>
  </div>
  <MeltSelect name="Search algorithm" options={["Depth-first search", "Breadth-first search", "Greedy best-first search", "A* search"]} value={algorithm} />
  <Button
    on:click={() => handleAIStep()}
  >
    <span>Single step</span>
    <RedoDot size="18" />
  </Button>
  <Button
    on:click={() => handleAISolve()}
  >
    <span>Solve</span>
    <Rocket size="18" />
  </Button>
</div>

<div class="mb-8 relative p-2 rounded-md bg-ctp-mantle">
  <div class="grid gap-1 md:gap-2" style={`grid-template-columns: repeat(${maze.width}, minmax(0, 1fr)); grid-template-rows: repeat(${maze.height}, minmax(0, 1fr));`}>
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
      <p class="mb-4 font-semibold text-ctp-mantle">
        Explored states: {maze.explored.length}
      </p>
      <Button
        on:click={() => {
          maze.reset();
          lastNode = maze.move();
          solved = false;
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
        Explored states: {maze.explored.length}
      </p>
      <Button
        on:click={() => {
          maze.reset();
          lastNode = maze.move();
          solved = false;
          solvable = true;
        }}
      >
        <span>Reset</span>
      </Button>
    </div>
  {/if}
</div>