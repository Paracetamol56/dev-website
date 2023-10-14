<script lang="ts">
	import EvaluateModel from "./EvaluateModel.svelte";
	import InitModel from "./InitModel.svelte";
	import LoadDataset from "./LoadDataset.svelte";
	import SaveLoadModel from "./SaveLoadModel.svelte";
	import TestModel from "./TestModel.svelte";
	import TrainModel from "./TrainModel.svelte";
	import { MnistData } from "./mnistData";

  let data: MnistData = new MnistData();
</script>

<LoadDataset data={data} />
<InitModel />
<TrainModel />
<EvaluateModel />
<SaveLoadModel />
<TestModel />