<script lang="ts">
  import * as Plot from '@observablehq/plot';
	import { variants } from '@catppuccin/palette';
	import { user } from '$lib/store';

  const data = [
    { name: "Viking 1", type: "Voyage", start: new Date("1975-08-20T00:00:00"), end: new Date("1976-07-20T00:00:00") },
    { name: "Viking 1", type: "Lander", start: new Date("1976-07-20T00:00:00"), end: new Date("1982-11-11T00:00:00") },
    { name: "Viking 2", type: "Voyage", start: new Date("1975-09-09T00:00:00"), end: new Date("1976-09-03T00:00:00") },
    { name: "Viking 2", type: "Lander", start: new Date("1976-09-03T00:00:00"), end: new Date("1980-04-12T00:00:00") },
    { name: "Mars Pathfinder", type: "Voyage", start: new Date("1996-09-04T00:00:00"), end: new Date("1997-07-04T00:00:00") },
    { name: "Mars Pathfinder", type: "Lander", start: new Date("1997-07-04T00:00:00"), end: new Date("1997-09-27T00:00:00") },
    { name: "Mars Pathfinder", type: "Rover", start: new Date("1997-07-04T00:00:00"), end: new Date("1997-09-27T00:00:00") },
    { name: "Spirit", type: "Voyage", start: new Date("2003-06-10T00:00:00"), end: new Date("2004-01-04T00:00:00") },
    { name: "Spirit", type: "Rover", start: new Date("2004-01-04T00:00:00"), end: new Date("2010-03-22T00:00:00") },
    { name: "Opportunity", type: "Voyage", start: new Date("2003-07-08T00:00:00"), end: new Date("2004-01-25T00:00:00") },
    { name: "Opportunity", type: "Rover", start: new Date("2004-01-25T00:00:00"), end: new Date("2018-06-10T00:00:00") },
    { name: "Phoenix", type: "Voyage", start: new Date("2007-08-04T00:00:00"), end: new Date("2008-05-25T00:00:00") },
    { name: "Phoenix", type: "Lander", start: new Date("2008-05-25T00:00:00"), end: new Date("2008-11-10T00:00:00") },
    { name: "Curiosity", type: "Voyage", start: new Date("2011-11-26T00:00:00"), end: new Date("2012-08-06T00:00:00") },
    { name: "Curiosity", type: "Rover", start: new Date("2012-08-06T00:00:00"), end: new Date() },
    { name: "InSight", type: "Voyage", start: new Date("2018-05-05T00:00:00"), end: new Date("2018-11-26T00:00:00") },
    { name: "InSight", type: "Lander", start: new Date("2018-11-26T00:00:00"), end: new Date("2022-12-15T00:00:00") },
    { name: "Tianwen-1", type: "Voyage", start: new Date("2020-07-23T00:00:00"), end: new Date("2021-02-10T00:00:00") },
    { name: "Tianwen-1", type: "Rover", start: new Date("2021-02-10T00:00:00"), end: new Date("2022-05-05T00:00:00") },
    { name: "Perseverance", type: "Voyage", start: new Date("2020-07-30T00:00:00"), end: new Date("2021-02-18T00:00:00") },
    { name: "Perseverance", type: "Rover", start: new Date("2021-02-18T00:00:00"), end: new Date() },
    { name: "Perseverance", type: "Rover & Helicoptère", start: new Date("2021-04-03T00:00:00"), end: new Date("2024-01-18T00:00:00") },
  ];

  let div: HTMLDivElement;

  let width: number;
  let height: number;

  $: {
    div?.firstChild?.remove();
    div?.append(
      Plot.plot({
        width,
        height,
        marginLeft: 100,
        x: {
          axis: 'top',
          label: null,
          grid: true
        },
        y: {
          axis: 'left',
          label: null,
          grid: false
        },
        color: {
          legend: true,
          range: [
						variants[$user.flavour].peach.rgb,
						variants[$user.flavour].red.rgb,
						variants[$user.flavour].blue.rgb,
						variants[$user.flavour].green.rgb,
					],
					domain: ['Voyage', 'Rover', 'Lander', 'Rover & Helicoptère']
        },
        marks: [
          Plot.barX(data, {
            x: "start",
            x2: "end",
            y: "name",
            fill: "type",
            sort: { y: "x" }
          })
        ]
      })
    );
  }
</script>

<div bind:this={div} class="flex justify-center aspect-video" role="img" aria-label="Mars missions timeline" bind:clientWidth={width} bind:clientHeight={height}></div>