<script lang="ts">
	import type { Maze } from "./maze";
  import * as d3 from "d3";
  
  export let maze: Maze;
  export let margin = { top: 40, right: 40, bottom: 40, left: 40 };

  let containerWidth: number;

  $: width = containerWidth;
  $: height = containerWidth / 16 * 9;

  let tree: d3.TreeLayout<unknown>;
  $: {
    tree = d3.tree()
      .size([
        height - margin.top - margin.bottom,
        width - margin.left - margin.right,
      ]);
    updateGraph();
  }
  
  let root: d3.HierarchyNode<unknown>;
  const updateGraph = () => {
    root = d3.stratify()
      .id((d: any) => d.state.x + "," + d.state.y)
      .parentId((d: any) => d.parent ? d.parent?.state.x + "," + d.parent?.state.y : "")
    (maze.explored);
    tree(root);
  }

  maze.stepCallback = updateGraph;
</script>

<div class="w-full" bind:clientWidth={containerWidth}>
  <svg
    class="bg-ctp-mantle rounded-md"
    width={width}
    height={height}
  >
    <g transform={`translate(${margin.left}, ${margin.top})`}>
      {#each root.descendants().slice(1) as link}
        <path
          class="stroke-ctp-lavender fill-none"
          d={`
            M ${link.y}, ${link.x}
            C ${(link.y + link.parent.y) / 2} ${link.x},
              ${(link.y + link.parent.y) / 2} ${link.parent.x},
              ${link.parent.y} ${link.parent.x}
          `}
        >
        </path>
      {/each}
      {#each root.descendants() as node}
        <circle
          cx={node.y}
          cy={node.x}
          r="5"
          class="fill-ctp-mauve"
        />
        <text
          x={node.y}
          y={node.x + 14}
          dy="0.31em"
          class="fill-ctp-mauve text-sm"
          text-anchor="middle"
        >
          ({node.data.state.x},{node.data.state.y})
        </text>
      {/each}
    </g>
  </svg>
</div>
