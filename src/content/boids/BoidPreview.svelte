<script lang="ts">
	import type { Writable } from "svelte/store";


  export let speed: Writable<number[]>;
  export let cohesion: Writable<number[]>;
  export let alignment: Writable<number[]>;
  export let separation: Writable<number[]>;
  export let vision: Writable<number[]>;

  $: cohesionRadius = $vision[0] * $cohesion[0];
  $: alignmentRadius = $vision[0] * $alignment[0] * $cohesion[0];
  $: separationRadius = $vision[0] * $separation[0] * $cohesion[0] * $alignment[0];

  const size: number = 200
  const halfSize: number = size / 2;
</script>

<div class="flex flex-col justify-center items-center">
  <svg width={size} height={size}>
    <g transform="translate(0, 0)">
      <circle cx={halfSize} cy={halfSize} r={cohesionRadius} class="fill-ctp-lavender/25 stroke-none" />
      <circle cx={halfSize} cy={halfSize} r={alignmentRadius} class="fill-ctp-lavender/25 stroke-none" />
      <circle cx={halfSize} cy={halfSize} r={separationRadius} class="fill-ctp-lavender/25 stroke-none" />
      <!--<path d={`M ${width / 2} ${height / 2} L ${90} ${height + 1} L ${width - 90} ${height + 1} Z`} class="fill-ctp-mantle" />-->
    </g>
    <g transform={`translate(${halfSize}, ${halfSize})`}>
      <path d="M 0 0 L -5 10 L 5 10 Z" class="fill-ctp-mauve stroke-none" />
    </g>
  </svg>
  <div class="grow-0 flex flex-row flex-wrap justify-center gap-2">
    <div class="flex gap-1 items-baseline">
      <div class="square-3 rounded-full bg-ctp-lavender/25"></div>
      <span class="text-ctp-text font-semibold text-sm">Cohesion</span>
    </div>
    <div class="flex">
      <div class="flex gap-1 items-baseline">
        <div class="square-3 rounded-full bg-ctp-lavender/40"></div>
        <span class="text-ctp-text font-semibold text-sm">Alignment</span>
      </div>
    </div>
    <div class="flex">
      <div class="flex gap-1 items-baseline">
        <div class="square-3 rounded-full bg-ctp-lavender/60"></div>
        <span class="text-ctp-text font-semibold text-sm">Separation</span>
      </div>
    </div>
  </div>
</div>
