<script lang="ts">
	import { writable, type Writable } from "svelte/store";
	import EvaluateModel from "./EvaluateModel.svelte";
	import InitModel from "./InitModel.svelte";
	import LoadDataset from "./LoadDataset.svelte";
	import SaveLoadModel from "./SaveLoadModel.svelte";
	import TestModel from "./TestModel.svelte";
	import TrainModel from "./TrainModel.svelte";
	import { MnistData } from "./mnistData";
	import * as tf from "@tensorflow/tfjs";

  const data: Writable<MnistData> = writable(new MnistData());
	const model: Writable<tf.Sequential> = writable(tf.sequential());

	export { data, model };
</script>

<LoadDataset {data} />
<InitModel {model}/>
<TrainModel {data} {model}/>
<EvaluateModel {data} {model} />
<SaveLoadModel {model} />
<TestModel {model} />