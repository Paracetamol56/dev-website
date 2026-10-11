<script lang="ts">
	import api from '$lib/api';
	import Button from '$lib/components/Button.svelte';
	import { EyeOff } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { writable } from 'svelte/store';
	import { addToast } from '../../../../+layout.svelte';
	import Dialog from '../../Dialog.svelte';
	import type { WordCloudWord } from '../../utils';

	export let id: string;
	export let words: WordCloudWord[];
	export let participants: Map<string, number>;
	export let participant: string | null;

	const dispatch = createEventDispatcher<{ words: WordCloudWord[]; participant: string }>();
	const open = writable(false);
	$: open.set(participant !== null);
	$: if (!$open) participant = null;

	$: own = words
		.filter((word) => word.uuid === participant)
		.sort((a, b) => a.createdAt.localeCompare(b.createdAt));
	$: first = own[0];
	$: last = own[own.length - 1];
	$: devices = new Set(own.map((word) => word.device).filter(Boolean));
	$: networks = new Set(own.map((word) => word.network).filter(Boolean));
	$: sharing = (hashes: Set<string | undefined>, key: 'device' | 'network') =>
		[
			...new Set(
				words
					.filter((word) => word.uuid !== participant && hashes.has(word[key]))
					.map((word) => word.uuid)
			)
		].sort((a, b) => (participants.get(a) ?? 0) - (participants.get(b) ?? 0));
	$: sameDevice = sharing(devices, 'device');
	$: sameNetwork = sharing(networks, 'network');

	const os = (userAgent: string) =>
		/Windows/.test(userAgent)
			? 'Windows'
			: /Android/.test(userAgent)
			? 'Android'
			: /iPhone|iPad|iPod/.test(userAgent)
			? 'iOS'
			: /Mac OS X/.test(userAgent)
			? 'macOS'
			: /Linux/.test(userAgent)
			? 'Linux'
			: 'Unknown';
	const browser = (userAgent: string) =>
		/Edg\//.test(userAgent)
			? 'Edge'
			: /OPR\//.test(userAgent)
			? 'Opera'
			: /Firefox\//.test(userAgent)
			? 'Firefox'
			: /Chrome\//.test(userAgent)
			? 'Chrome'
			: /Safari\//.test(userAgent)
			? 'Safari'
			: 'Unknown';
	const dateTime = (date: string) => new Date(date).toLocaleString();

	let hiding = false;
	async function hideAll() {
		hiding = true;
		let latest: WordCloudWord[] | null = null;
		for (const word of own.filter((word) => !word.hidden)) {
			try {
				latest = (
					await api.callWithAuth('PATCH', `/word-cloud/${id}/words/${word.id}`, { hidden: true })
				).data.words;
			} catch {
				addToast({
					data: {
						title: 'Error',
						description: `“${word.text}” could not be hidden`,
						color: 'bg-ctp-red'
					}
				});
			}
		}
		if (latest) dispatch('words', latest);
		hiding = false;
	}
</script>

<Dialog {open} title="Participant #{participant ? participants.get(participant) : ''}">
	{#if participant && first}
		<div class="flex max-h-[70vh] flex-col gap-4 overflow-y-auto">
			<dl class="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
				<div>
					<dt class="text-ctp-subtext0">Words</dt>
					<dd class="font-semibold">
						{own.length}{own.some((w) => w.hidden)
							? ` (${own.filter((w) => w.hidden).length} hidden)`
							: ''}
					</dd>
				</div>
				<div>
					<dt class="text-ctp-subtext0">Browser</dt>
					<dd class="font-semibold">{browser(first.userAgent)} on {os(first.userAgent)}</dd>
				</div>
				<div>
					<dt class="text-ctp-subtext0">First word</dt>
					<dd class="font-semibold">{dateTime(first.createdAt)}</dd>
				</div>
				<div>
					<dt class="text-ctp-subtext0">Last word</dt>
					<dd class="font-semibold">{dateTime(last.createdAt)}</dd>
				</div>
			</dl>

			{#if sameDevice.length}
				<p class="rounded-md bg-ctp-peach/15 px-3 py-2 text-sm">
					<strong class="text-ctp-peach">Same device</strong> as
					{#each sameDevice as other, i}
						<button
							type="button"
							class="font-mono font-semibold text-ctp-mauve hover:underline"
							on:click={() => dispatch('participant', other)}>#{participants.get(other)}</button
						>{i < sameDevice.length - 1 ? ', ' : ''}
					{/each}: probably one person with several identities.
				</p>
			{/if}
			{#if sameNetwork.length}
				<p class="text-sm text-ctp-subtext0">
					Same network as {sameNetwork.length} other participant{sameNetwork.length === 1
						? ''
						: 's'}, which is expected when the audience shares a Wi-Fi.
				</p>
			{/if}

			<ul class="flex flex-wrap gap-2">
				{#each own as word (word.id)}
					<li
						class="rounded-md px-2 py-1 text-sm {word.hidden
							? 'bg-ctp-surface0 text-ctp-overlay1 line-through'
							: 'bg-ctp-mauve/20 font-semibold text-ctp-text'}"
						title={dateTime(word.createdAt)}
					>
						{word.text}
					</li>
				{/each}
			</ul>

			<div class="flex justify-end">
				<Button on:click={hideAll} disabled={hiding || own.every((w) => w.hidden)}>
					<EyeOff size="16" />
					<span>{hiding ? 'Hiding...' : 'Hide all their words'}</span>
				</Button>
			</div>
		</div>
	{/if}
</Dialog>
