<script lang="ts">
	import Button from '$lib/components/Button.svelte';

  const save = (e: Event) => {
    e.preventDefault();
    // TODO: save model
  }
</script>

<div>
	<div class="flex flex-col md:flex-row gap-4">
		<div class="w-full">
			<h3>Save Model</h3>
			<Button on:click={save}>Save Model</Button>
			<h4>
				<div id="saved" style="display:none;">Model Saved</div>
			</h4>
		</div>
		<div class="w-full">
			<h3>Load Model</h3>
      <div class="flex flex-wrap gap-2">
        <label class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1 font-semibold text-ctp-mantle
                      shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60" for="json-upload"
        >
          Browse JSON 
        </label>
        <input type="file" id="json-upload" class="hidden" />
        <label class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1 font-semibold text-ctp-mantle
                      shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60" for="weights-upload">
          Browse Weights 
        </label>
        <input type="file" id="weights-upload" class="hidden" />
        <Button>
          Load Model
        </Button>
      </div>
      <p class="mt-2 text-ctp-green font-semibold">Loaded</p>
		</div>
	</div>
</div>
