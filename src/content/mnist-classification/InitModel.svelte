<script lang="ts">
	import Button from "$lib/components/Button.svelte";
	import { createRadioGroup, melt } from "@melt-ui/svelte";

  const {
    elements: { root, item, hiddenInput },
    helpers: { isChecked },
  } = createRadioGroup({
    defaultValue: 'default',
  });

  const radioOptions = ['ANN', 'CNN'];
</script>

<div>
	<h3>Model Architecture</h3>
	<div>
		<textarea class="flex h-32 w-full items-center justify-between rounded-md bg-ctp-surface0 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-ctp-mauve" id="model"> model = tf.sequential(); </textarea>
	</div>
	<div class="my-4 p-4 rounded-md bg-ctp-mantle flex flex-col md:flex-row gap-4 justify-between items-center">
		<div
      use:melt={$root}
      class="flex gap-4"
      aria-label="View density"
    >
      {#each radioOptions as option}
        <div class="flex items-center gap-2">
          <button
            use:melt={$item(option)}
            class="grid square-3.5 cursor-default place-items-center rounded-full bg-ctp-surface0 shadow-sm shadow-ctp-crust
                  hover:bg-ctp-surface1"
            id={option}
            aria-labelledby="{option}-label"
          >
            {#if $isChecked(option)}
              <div class="square-2 rounded-full bg-ctp-mauve" />
            {/if}
          </button>
          <label
            class="capitalize leading-none font-semibold"
            for={option}
            id="{option}-label"
          >
            {option}
          </label>
        </div>
      {/each}
      <input name="line-height" use:melt={$hiddenInput} />
    </div>
    <Button>
      Initialize Model
    </Button>
	</div>
	<h3>Model Summary</h3>
	<div id="summary" />
</div>
