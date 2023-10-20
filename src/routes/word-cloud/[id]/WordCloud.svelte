<script lang="ts">
	import Button from '$lib/components/Button.svelte';
  import cloud from 'd3-cloud'
	import { FileCode } from 'lucide-svelte';

  export let data: {text: string, occurence: number}[];
  export let margin: { top: number; right: number; bottom: number; left: number; } = {
    top: 40,
    right: 40,
    bottom: 40,
    left: 40
  };

  let containerWidth: number;
  let containerHeight: number;

  $: width = containerWidth;
  $: height = width * 9 / 16;

  const colors = [
    "fill-ctp-rosewater",
    "fill-ctp-flamingo",
    "fill-ctp-pink",
    "fill-ctp-mauve",
    "fill-ctp-red",
    "fill-ctp-maroon",
    "fill-ctp-peach",
    "fill-ctp-yellow",
    "fill-ctp-green",
    "fill-ctp-teal",
    "fill-ctp-sky",
    "fill-ctp-sapphire",
    "fill-ctp-blue",
    "fill-ctp-lavender",
  ]

  let words: cloud.Word[] = [];
  $: {
    if (width && height && data) {
      words = [];
      cloud()
        .size([
          width - margin.left - margin.right,
          height - margin.top - margin.bottom
        ])
        .words(data.map(d => ({ text: d.text, size: d.occurence })))
        .padding(2)
        .rotate(0)
        .font("Helvetica")
        .fontSize(d => Math.sqrt(d.size!) * 30)
        .on("word", (word) => {
          words = [...words, word];
        })
        .start();
    }
  }

  const exportSVG = (e: Event) => {
    e.preventDefault();

    const svg = `<?xml version="1.0" encoding="utf-8"?>
    <svg
      xmlns="http://www.w3.org/2000/svg" 
      version="1.1"
      width="${width}"
      height="${height}"
      viewBox="0 0 ${width} ${height}"
      text-anchor="middle"
    >
      <style>
        .fill-ctp-rosewater{ fill: rgb(220, 138, 120); }
        .fill-ctp-flamingo{ fill: rgb(221, 120, 120); }
        .fill-ctp-pink{ fill: rgb(234, 118, 203); }
        .fill-ctp-mauve{ fill: rgb(136, 57, 239); }
        .fill-ctp-red{ fill: rgb(210, 15, 57); }
        .fill-ctp-maroon{ fill: rgb(230, 69, 83); }
        .fill-ctp-peach{ fill: rgb(254, 100, 11); }
        .fill-ctp-yellow{ fill: rgb(223, 142, 29); }
        .fill-ctp-green{ fill: rgb(64, 160, 43); }
        .fill-ctp-teal{ fill: rgb(23, 146, 153); }
        .fill-ctp-sky{ fill: rgb(4, 165, 229); }
        .fill-ctp-sapphire{ fill: rgb(32, 159, 181); }
        .fill-ctp-blue{ fill: rgb(30, 102, 245); }
        .fill-ctp-lavender{ fill: rgb(114, 135, 253); }
        .fill-ctp-text{ fill: rgb(76, 79, 105); }
        .fill-ctp-subtext1{ fill: rgb(92, 95, 119); }
        .fill-ctp-subtext0{ fill: rgb(108, 111, 133); }
        .fill-ctp-overlay2{ fill: rgb(124, 127, 147); }
        .fill-ctp-overlay1{ fill: rgb(140, 143, 161); }
        .fill-ctp-overlay0{ fill: rgb(156, 160, 176); }
        .fill-ctp-surface2{ fill: rgb(172, 176, 190); }
        .fill-ctp-surface1{ fill: rgb(188, 192, 204); }
        .fill-ctp-surface0{ fill: rgb(204, 208, 218); }
        .fill-ctp-base{ fill: rgb(239, 241, 245); }
        .fill-ctp-mantle{ fill: rgb(230, 233, 239); }
        .fill-ctp-crust{ fill: rgb(220, 224, 232); }
      </style>
      ${document.getElementById("word-cloud")?.innerHTML}
    </svg>`;
    const blob = new Blob([svg], { type: "image/svg+xml" });
    const url = window.URL.createObjectURL(blob);

    const a = document.createElement("a");
    a.setAttribute("hidden", "");
    a.setAttribute("href", url);
    a.setAttribute("download", `word-cloud.svg`);
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  }
</script>

<div class="mb-4 p-4 bg-ctp-mantle rounded-md">
  <div class="flex justify-between">
    <h2 class="text-2xl font-bold text-ctp-lavender">
      Word cloud
    </h2>
    <Button
      on:click={exportSVG}
    >
      <span>Export as SVG</span>
      <FileCode size="18" />
  </Button>
  </div>
  <div
    class="mb-8 w-full"
    bind:clientWidth={containerWidth}
    bind:clientHeight={containerHeight}
  >
    <svg
      id="word-cloud"
      class="text-xs max-w-full"
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      text-anchor="middle"
    >
      <g
        x={width / 2}
        y={height / 2}
        transform={`translate(${width / 2}, ${height / 2})`}
      >
        {#each words as word}
          <text
            class={colors[Math.floor(Math.random() * colors.length)]}
            font-size={word.size}
            font-family={word.font}
            transform={`translate(${word.x}, ${word.y}) rotate(${word.rotate})`}
          >
            {word.text}
          </text>
        {/each}
      </g>
    </svg>
  </div>
</div>

