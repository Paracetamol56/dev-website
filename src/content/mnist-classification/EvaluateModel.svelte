<script lang="ts">
	import Button from '$lib/components/Button.svelte';
</script>

<div>
	<h3>Evaluate Model</h3>
  <Button>Evaluate Model</Button>

	<div class="w-full flex flex-col md:flex-row gap-4">
		<div class="w-full">
			<h3>Class Accuracy</h3>
			<div id="class-accuracy" class="w-full h-64 p-4 rounded-md bg-ctp-mantle" />
		</div>
		<div class="w-full">
			<h3>Confusion Matrix</h3>
			<div id="confusion-matrix" class="w-full h-64 p-4 rounded-md bg-ctp-mantle" />
		</div>
	</div>

	<h3>Prediction Example</h3>
	<Button>Show Prediction</Button>
	<div id="example-preview"  class="mt-4 w-full h-96 p-4 rounded-md bg-ctp-mantle" />
</div>
