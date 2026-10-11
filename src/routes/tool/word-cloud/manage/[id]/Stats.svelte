<script lang="ts">
	import { Hash, MessageSquareText, Radio, Users } from 'lucide-svelte';
	import { onDestroy } from 'svelte';
	import Kpi from '../../Kpi.svelte';
	import type { Distribution, WordCloudWord } from '../../utils';

	export let open: boolean;
	export let code: string;
	export let words: WordCloudWord[];
	export let participants: number;
	export let distribution: Distribution;

	const MINUTE = 60_000;
	let now = Date.now();
	const clock = setInterval(() => (now = Date.now()), 10_000);
	onDestroy(() => clearInterval(clock));

	$: hidden = words.filter((word) => word.hidden).length;
	$: lastMinute = words.filter((word) => now - Date.parse(word.createdAt) < MINUTE).length;
	$: average = participants ? (words.length / participants).toFixed(1) : '0';
	$: top = distribution[0];

	const plural = (count: number, word: string) => `${count} ${word}${count === 1 ? '' : 's'}`;
</script>

<div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
	<Kpi
		icon={Radio}
		iconClass={open ? 'text-ctp-green' : 'text-ctp-red'}
		label="Status"
		value={open ? 'Open' : 'Closed'}
	>
		Code <span class="font-mono font-semibold text-ctp-mauve">{code}</span>
	</Kpi>
	<Kpi
		icon={MessageSquareText}
		iconClass="text-ctp-blue"
		label="Submissions"
		value={words.length}
		detail="{lastMinute} in the last minute{hidden ? ` · ${hidden} hidden` : ''}"
	/>
	<Kpi
		icon={Users}
		iconClass="text-ctp-peach"
		label="Participants"
		value={participants}
		detail="{average} words each on average"
	/>
	<Kpi
		icon={Hash}
		iconClass="text-ctp-mauve"
		label="Unique words"
		value={distribution.length}
		detail={top ? `Top: ${top.text} (${plural(top.occurence, 'time')})` : 'No words yet'}
	/>
</div>
