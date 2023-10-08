<script lang="ts">
	import { onMount } from "svelte";
	import MazeControls from "./MazeControls.svelte";
  import MazeGraph from "./MazeGraph.svelte";
	import MazeGrid from "./MazeGrid.svelte";
	import { Maze } from "./maze";
	import type { Node } from "./utils";
	import { readable, type Readable } from "svelte/store";

  const updateMaze = (node: Node) => {
    mazeGridRef.updateGrid();
    mazeGraphRef.updateGraph();
  };

  let maze: Readable<Maze> = readable(new Maze(updateMaze), (set) => {
    set(new Maze(updateMaze));
  });
  let mazeGridRef: MazeGrid;
  let mazeGraphRef: MazeGraph;


  const loadMaze = (index: number) => {
    const newMaze = new Maze(updateMaze);
    switch (index) {
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

  loadMaze(2);
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
    <MazeControls maze={maze} loadMaze={loadMaze}  />
    <MazeGrid maze={maze} bind:this={mazeGridRef} />
    <MazeGraph maze={maze} bind:this={mazeGraphRef} />
  </section>
</div>
