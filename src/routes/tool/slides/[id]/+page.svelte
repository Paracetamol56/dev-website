<script lang="ts">
  import CodeMirror from "svelte-codemirror-editor";
  import { html } from "@codemirror/lang-html";
  import { catppuccinMocha } from "@catppuccin/codemirror"
  import { melt, createScrollArea, createToolbar } from '@melt-ui/svelte';
  import MeltSelect from "$lib/components/MeltSelect.svelte";

  import Slides from "$lib/components/deck/Slides.svelte"
  import Slide from "$lib/components/deck/Slide.svelte"

  let value: string = "<section>\n<h1>Hello world !</h1>\n</section>\n<section>\n<h2>Test</h2>\n</section>";

  const {
    elements: {
      root,
      content,
      viewport,
      corner,
      scrollbarY,
      thumbY,
      thumbX,
      scrollbarX,
    },
  } = createScrollArea({
    type: 'auto',
    dir: 'ltr',
  });

  const {
    elements: { root: toolbarRoot, button, link, separator },
    builders: { createToolbarGroup },
  } = createToolbar();
</script>


<main class="bg-ctp-mantle h-screen max-h-screen grid grid-cols-2 grid-rows-[auto_1fr_auto] gap-4 p-4">
  <div class="bg-ctp-base text-ctp-text rounded-md shadow-md shadow-ctp-crust bg-ctp-crust/50 p-4">
    Toolbar
  </div>
  <div use:melt={$root} class="col-start-1 row-start-2 h-full overflow-hidden bg-ctp-base rounded-md shadow-md shadow-ctp-crust">
    <div use:melt={$viewport} class="h-full">
      <div use:melt={$content} class="h-full">
        <CodeMirror bind:value lang={html()} theme={catppuccinMocha} class="h-full" />
      </div>
    </div>
    <div
      use:melt={$scrollbarY}
      class="flex h-full w-3.5 touch-none select-none border-l border-l-transparent p-1 transition-colors"
    >
      <div
        use:melt={$thumbY}
        class="relative flex-1 rounded-full bg-ctp-surface1"
      />
    </div>
    <div
      use:melt={$scrollbarX}
      class="flex h-3.5 w-full touch-none select-none border-t border-t-transparent p-1"
    >
      <div use:melt={$thumbX} class="relative rounded-full bg-ctp-surface1" />
    </div>
    <div use:melt={$corner} />
  </div>
  <div class="row-span-2 col-start-2 row-start-1 rounded-md overflow-hidden bg-ctp-base shadow-md shadow-ctp-crust">
    <Slides>
      <Slide>
        {@html value}
      </Slide>
    </Slides>
  </div>
  <div class="col-span-2 row-start-3 bg-ctp-base rounded-md shadow-md shadow-ctp-crust bg-ctp-crust/50 min-h-12">
  </div>
</main>
