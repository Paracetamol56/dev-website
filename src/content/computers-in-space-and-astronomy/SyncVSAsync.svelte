<script lang="ts">
  import { onMount } from 'svelte';
  import * as Plot from '@observablehq/plot';
  import * as d3 from 'd3';
	import { variants } from '@catppuccin/palette';
	import { user } from '$lib/store';

  const syncData = [
    { name: "Tâche 1", start: 0, end: 0.8 },
    { name: "Tâche 2", start: 0.8, end: 2.1 },
    { name: "Tâche 3", start: 2.1, end: 3.2 },
    { name: "Tâche 4", start: 3.2, end: 4.5 },
    { name: "Tâche 5", start: 4.5, end: 5.5 }
  ];
  const asyncData = [
    { name: "Tâche 1", start: 0, end: 1.2 },
    { name: "Tâche 2", start: 0, end: 0.7 },
    { name: "Tâche 3", start: 0, end: 1.5 },
    { name: "Tâche 4", start: 0, end: 2.5 },
    { name: "Tâche 5", start: 0, end: 3.5 }
  ]

  let syncDiv: HTMLDivElement;
  let asyncDiv: HTMLDivElement;
  let width: number;
  $: {
    syncDiv?.firstChild?.remove();
    syncDiv?.append(
      Plot.plot({
        height: 300,
        width,
        marginLeft: 50,
        axis: null,
        y: {
          axis: 'left',
          label: null,
          grid: true
        },
        marks: [
          Plot.barX(syncData, {
            x: "start",
            x2: "end",
            y: "name",
            fill: variants[$user.flavour].blue.rgb
          })
        ]
      })
    );
  }
  $:{
    asyncDiv?.append(
      Plot.plot({
        height: 300,
        width,
        marginLeft: 50,
        axis: null,
        y: {
          axis: 'left',
          label: null,
          grid: true
        },
        marks: [
          Plot.barX(asyncData, {
            x: "start",
            x2: "end",
            y: "name",
            fill: variants[$user.flavour].blue.rgb
          })
        ]
      })
    );
  }
</script>

<div class="w-full flex flex-col md:flex-row gap-4">
  <div class="w-full">
    <h4>Synchrone</h4>
    <div bind:this={syncDiv} class="flex justify-start w-full" role="img" bind:clientWidth={width} />
  </div>
  <div class="w-full">
    <h4>Asynchrone</h4>
    <div bind:this={asyncDiv} class="flex justify-start w-full" role="img" />
  </div>
</div>