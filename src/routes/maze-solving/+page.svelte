<script lang="ts">
	import { onMount } from "svelte";
	import MazeControls from "./MazeControls.svelte";
  import MazeGraph from "./MazeGraph.svelte";
	import MazeGrid from "./MazeGrid.svelte";
	import { Maze } from "./maze";
	import { StackFrontier, type Frontier, type Node, QueueFrontier, PriorityQueueFrontier } from "./utils";
	import { readable, writable, type Readable, type Writable } from "svelte/store";
	import type { SelectOption } from "@melt-ui/svelte";

  const updateMaze = (_: Node) => {
    mazeGridRef.updateGrid();
    mazeGraphRef.updateGraph();
  };

  const createMaze = () => {
    switch ($algorithm.value) {
      case "Depth-first search":
        return new Maze<StackFrontier>(new StackFrontier(), updateMaze);
      case "Breadth-first search":
        return new Maze<QueueFrontier>(new QueueFrontier(), updateMaze);
      case "Greedy best-first search":
        return new Maze<PriorityQueueFrontier>(new PriorityQueueFrontier(), updateMaze);
      case "A* search":
        return new Maze<PriorityQueueFrontier>(new PriorityQueueFrontier(), updateMaze);
      default:
        throw new Error("Invalid algorithm");
    }
  };

  const loadMaze = () => {
    const newMaze = createMaze();
    switch ($presetMaze) {
      case 1:
        newMaze.loadFromString(
` # # ###  #B
 # #   # ## 
   # #   ## 
# ## # #    
#    # #####
### ## #####
A   ##      `
        );
        break;
      case 2:
        newMaze.loadFromString(
` #####
      
 ## ##
 ##B##
 #  ##
 # ###
A  ###`
        );
        break;
      case 3:
        newMaze.loadFromString(``);
        break;
      case 4:
        newMaze.loadFromString(
`          B
 #### #### 
 #     # # 
 # ### #   
   #     # 
## # ##### 
A      #   `);
        break;
      default:
        return;
    }

    maze = readable(newMaze, (set) => {
      set(newMaze);
    });
    if (mazeGridRef) mazeGridRef.updateGrid();
    if (mazeGraphRef) mazeGraphRef.updateGraph();
  };

  let presetMaze: Writable<number> = writable(1);
  let algorithm: Writable<SelectOption<string>> = writable({value: "Depth-first search", label: "Depth-first search"});
  let maze: Readable<Maze<Frontier>> = readable(createMaze(), (set) => {
    set(createMaze());
  });
  let mazeGridRef: MazeGrid;
  let mazeGraphRef: MazeGraph;

  presetMaze.subscribe((value) => {
    loadMaze();
  });

  algorithm.subscribe((value) => {
    maze = readable(createMaze(), (set) => {
      set(createMaze());
    });
    loadMaze();
    if (mazeGridRef) mazeGridRef.updateGrid();
    if (mazeGraphRef) mazeGraphRef.updateGraph();
  });

  loadMaze();
  onMount(() => {
    mazeGridRef.updateGrid();
    mazeGraphRef.updateGraph();
  });
</script>

<div class="container mx-auto">
  <h1 class="mt-16 mb-16 text-6xl font-semibold text-center">
    <span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">Maze solving</span>
  </h1>
  <section id="maze">
    <MazeControls maze={maze} presetMaze={presetMaze} algorithm={algorithm} />
    <MazeGrid maze={maze} bind:this={mazeGridRef} />
    <MazeGraph maze={maze} bind:this={mazeGraphRef} />
  </section>
</div>
