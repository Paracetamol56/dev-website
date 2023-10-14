<script lang="ts">
	import Button from '$lib/components/Button.svelte';
</script>

<div>
	<h3>Draw New Image</h3>
	<div>
		<div>
			<canvas id="predict-canvas" width="140" height="140" style="border:1px solid #000000;" />
		</div>
		<div style="text-align:center">
			<div>
				preview <br />
				<canvas id="preview-canvas" width="28" height="28" style="border:1px solid #000000;" />
			</div>
			<h4><div id="prediction" /></h4>
		</div>
		<div>
			Prediction Class Histogram
			<div id="predict-graph" />
		</div>
	</div>
	<Button>Clear</Button>
	<Button>Predict</Button>
</div>
