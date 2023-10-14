<script lang="ts">
	import Button from '$lib/components/Button.svelte';

	let maxEpoch: number = 1;
	let maxEpochError: string = '';

	let batchSize: number = 1;
	let batchSizeError: string = '';
</script>

<div>
	<div class="flex flex-col lg:flex-row gap-4">
		<div class="lg:w-64">
			<h3>Train Model</h3>
			<div class="mb-2">
				<fieldset class="w-full">
					<label for="epoch" class="mb-2 text-sm font-semibold">Max Epoch</label>
					<input
						id="epoch"
						type="number"
						bind:value={maxEpoch}
						class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0
                  px-3 pr-12 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
					/>
					<p class="text-left text-sm font-semibold text-ctp-red">
						{maxEpochError}
					</p>
				</fieldset>

				<fieldset class="w-full">
					<label for="batch" class="mb-2 text-sm font-semibold">Batch Size</label>
					<input
						id="batch"
						type="number"
						bind:value={batchSize}
						class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0
                  px-3 pr-12 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
					/>
					<p class="text-left text-sm font-semibold text-ctp-red">
						{batchSizeError}
					</p>
				</fieldset>
			</div>
			<Button>Train Model</Button>
		</div>
		<div class="w-full">
			<h3>Training Progress</h3>
			<div class="w-full flex flex-col gap-4 md:flex-row justify-stretch">
        <div class="w-full h-64 p-4 rounded-md bg-ctp-mantle">
          <p>Loss</p>
        </div>
        <div class="w-full h-64 p-4 rounded-md bg-ctp-mantle">
          <p>Accuracy</p>
        </div>
      </div>
		</div>
	</div>
</div>
