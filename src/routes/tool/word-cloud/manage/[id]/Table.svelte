<script lang="ts">
	import { FileSpreadsheet } from 'lucide-svelte';
	import type { WordCloudWord } from '../../utils';
	import Button from '$lib/components/Button.svelte';
	import { downloadBlob } from '$lib/download';

	export let id: string;
	export let data: WordCloudWord[];

	// Quoted for commas and newlines; a leading = + - @ would run as a spreadsheet formula
	const csvCell = (value: unknown) => {
		const text = String(value ?? '');
		return `"${(/^[=+\-@]/.test(text) ? `'${text}` : text).replaceAll('"', '""')}"`;
	};

	const exportCSV = (e: Event) => {
		e.preventDefault();
		const rows = [
			['Word', 'Participant', 'Date', 'User agent'],
			...data.map((d) => [d.text, d.uuid, new Date(d.createdAt).toISOString(), d.userAgent])
		];
		const csv = rows.map((row) => row.map(csvCell).join(',')).join('\n');
		downloadBlob(new Blob([csv], { type: 'text/csv' }), `word-cloud-${id}.csv`);
	};

	const os = (userAgent: string) =>
		userAgent
			.match(/\(([^)]+)\)/)?.[1]
			.split(';')[1]
			?.trim() ?? '';
	const browser = (userAgent: string) => userAgent.split(' ').at(-1) ?? '';
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
				<th class="px-4 py-2 text-start">Participant</th>
				<th class="px-4 py-2 text-start">OS</th>
				<th class="px-4 py-2 text-start">Browser</th>
			</tr>
		</thead>
		<tbody class="divide-y divide-ctp-surface0">
			{#each data as d}
				<tr class="hover:bg-ctp-crust">
					<td class="px-4 py-1">{d.text}</td>
					<td class="px-4 py-1">{d.uuid}</td>
					<td class="px-4 py-1">{os(d.userAgent)}</td>
					<td class="px-4 py-1">{browser(d.userAgent)}</td>
				</tr>{/each}
		</tbody>
	</table>
</div>
