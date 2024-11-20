<script lang="ts">
	import { FileSpreadsheet } from 'lucide-svelte';
	import type { WordCloudWord } from '../utils';
	import Button from '$lib/components/Button.svelte';

	export let id: string;
	export let data: WordCloudWord[];

	const exportCSV = (e: Event) => {
		e.preventDefault();

		const csv = [
			['Word', 'IP address', 'Date', 'User agent'],
			...data.map((d) => [d.text, d.ip, d.createdAt, d.userAgent])
		]
			.map((d) => d.join(','))
			.join('\n');

		const blob = new Blob([csv], { type: 'text/csv' });
		const url = window.URL.createObjectURL(blob);

		const a = document.createElement('a');
		a.setAttribute('hidden', '');
		a.setAttribute('href', url);
		a.setAttribute('download', `word-cloud-${id}.csv`);
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
	};
</script>

<div class="mb-8 w-full p-4 bg-ctp-mantle rounded-md">
	<div class="mb-4 flex justify-between">
		<h2 class="text-2xl font-bold text-ctp-lavender">Raw data</h2>
		<Button on:click={exportCSV}>
			<span>Export as CSV</span>
			<FileSpreadsheet size="18" />
		</Button>
	</div>
	<table class="w-full table-auto">
		<thead>
			<tr>
				<th class="px-4 py-2 text-start">Word</th>
				<th class="px-4 py-2 text-start">UUID</th>
				<th class="px-4 py-2 text-start">OS</th>
				<th class="px-4 py-2 text-start">Browser</th>
			</tr>
		</thead>
		<tbody class="divide-y divide-ctp-surface0">
			{#each data as d}
				<tr class="hover:bg-ctp-crust">
					<td class="px-4 py-1">{d.text}</td>
					<td class="px-4 py-1">{d.uuid}</td>
					<td class="px-4 py-1">{d.userAgent.match(/\(([^)]+)\)/)[1].split(';')[1]}</td>
					<td class="px-4 py-1">{d.userAgent.split(' ').at(-1)}</td>
				</tr>{/each}
		</tbody>
	</table>
</div>
