<script lang="ts">
	import MeltCheckbox from '$lib/components/MeltCheckbox.svelte';
	import NumberInput from '$lib/components/NumberInput.svelte';
	import { writable } from 'svelte/store';
	import type { Settings } from './utils';

	export let initial: Settings;
	export let settings: Settings = initial;

	const filterProfanity = writable(initial.filterProfanity);
	const showCloud = writable(initial.showCloudToAudience);
	const allowUpvotes = writable(initial.allowUpvotes);
	const sentiment = writable(initial.sentiment);
	const sentimentInPresenter = writable(initial.sentimentInPresenter);
	const maxWords = writable(initial.maxWords ?? 20);

	$: settings = {
		...initial,
		filterProfanity: $filterProfanity,
		showCloudToAudience: $showCloud,
		allowUpvotes: $allowUpvotes,
		sentiment: $sentiment,
		sentimentInPresenter: $sentiment && $sentimentInPresenter,
		maxWords: Number.isFinite($maxWords) ? Math.min(100, Math.max(0, Math.round($maxWords))) : 20
	};
</script>

<fieldset class="grid gap-3">
	<legend class="mb-2 text-sm font-semibold">Settings</legend>
	<div>
		<MeltCheckbox
			name="settings-filter"
			label="Hide offensive words automatically"
			checked={filterProfanity}
		/>
		<p class="ml-6 mt-1 text-sm text-ctp-subtext0">
			28 languages, you can show them again. Lists by
			<a
				class="underline"
				href="https://github.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words"
				target="_blank"
				rel="noreferrer">LDNOOBW</a
			> (CC BY 4.0)
		</p>
	</div>
	<div>
		<MeltCheckbox
			name="settings-show-cloud"
			label="Show the live cloud to participants"
			checked={showCloud}
		/>
		<p class="ml-6 mt-1 text-sm text-ctp-subtext0">Below their form, without hidden words</p>
	</div>
	<div>
		<MeltCheckbox
			name="settings-upvotes"
			label="Let participants upvote recent answers"
			checked={allowUpvotes}
		/>
		<p class="ml-6 mt-1 text-sm text-ctp-subtext0">
			They see the last answers of others and can add them as their own
		</p>
	</div>
	<div>
		<MeltCheckbox
			name="settings-sentiment"
			label="Analyse the mood of the answers"
			checked={sentiment}
		/>
		<p class="ml-6 mt-1 text-sm text-ctp-subtext0">Seven languages, shown on this dashboard</p>
	</div>
	{#if $sentiment}
		<div class="ml-6">
			<MeltCheckbox
				name="settings-sentiment-presenter"
				label="Show the mood in presenter mode"
				checked={sentimentInPresenter}
			/>
		</div>
	{/if}
	<div class="w-56">
		<NumberInput label="Words per participant" value={maxWords} min={0} max={100} />
		<p class="mt-1 text-sm text-ctp-subtext0">0: no limit</p>
	</div>
</fieldset>
