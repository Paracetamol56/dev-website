<script lang="ts">
  import MazeGraph from "./MazeGraph.svelte";
	import MazeGrid from "./MazeGrid.svelte";
	import { Maze } from "./maze";

  let maze: Maze = new Maze();

  // Maze 1:
  //  # # ###  #B
  //  # #   # ## 
  //    # #   ## 
  // # ## # #    
  // #    # #####
  // ### ## #####
  // A   ##      

  // Maze 2:
  //  #####
  //       
  //  ## ##
  //  ##B##
  //  #  ##
  //  # ###
  // A  ###

  const loadMaze = (index: number) => {
    maze = new Maze();
    switch (index) {
      case 1:
        maze.loadFromString(
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
        maze.loadFromString(
` #####
      
 ## ##
 ##B##
 #  ##
 # ###
A  ###`
        );
        break;
      case 3:
        maze.loadFromString(``);
        break;
      case 4:
        maze.loadFromString(
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
  };

  loadMaze(1);
</script>

<div class="container mx-auto">
  <h1 class="mt-16 mb-16 text-6xl font-semibold text-center">
    <span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">Maze solving</span>
  </h1>
  <section id="maze">
    <MazeGrid maze={maze} loadMaze={loadMaze} />
    <MazeGraph maze={maze} />
  </section>
</div>
