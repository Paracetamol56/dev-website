<script lang="ts">
	import type { SelectOption } from '@melt-ui/svelte';
	import { writable } from 'svelte/store';
	import api from '$lib/api';
	import { addToast } from '$lib/toast';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Loader from '$lib/components/ui/Loader.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Text from '$lib/components/ui/Text.svelte';
	import { Tone } from '$lib/components/ui/styles';
	import { TextColor } from '$lib/components/ui/typography';
	import { formatDate, type AuditEntry } from './types';

	export let userId: string = '';
	export let refresh: number = 0;

	const LIMIT = 25;
	const actions = [
		'auth.link_requested',
		'auth.login',
		'auth.login_failed',
		'user.created',
		'user.identity_linked',
		'user.identity_unlinked',
		'user.passkey_added',
		'user.passkey_removed',
		'user.profile_updated',
		'user.data_exported',
		'user.deleted',
		'wordcloud.created',
		'wordcloud.duplicated',
		'wordcloud.updated',
		'wordcloud.deleted',
		'wordcloud.word_hidden',
		'wordcloud.words_merged',
		'contact.sent',
		'admin.role_changed',
		'access.denied',
		'access.rate_limited'
	];
	const options: SelectOption<string>[] = [
		{ value: '', label: 'All actions' },
		...actions.map((action) => ({ value: action, label: action }))
	];
	const action = writable(options[0]);
	const page = writable(1);

	let entries: AuditEntry[] = [];
	let names: Record<string, { name: string; email: string }> = {};
	let total = 0;
	let loading = true;
	let request = 0;

	const tone = (action: string) => {
		if (action === 'auth.login_failed' || action.startsWith('access.')) return Tone.Danger;
		if (action === 'user.deleted' || action.startsWith('admin.')) return Tone.Warning;
		if (action === 'auth.login' || action === 'user.created') return Tone.Success;
		return Tone.Neutral;
	};

	$: label = (id: string) => names[id]?.email ?? id;

	const details = (entry: AuditEntry) =>
		Object.entries(entry.meta ?? {})
			.map(([key, value]) => `${key}: ${Array.isArray(value) ? value.join(', ') : value}`)
			.join(' · ');

	const load = async (action: string, page: number, userId: string) => {
		const current = ++request;
		loading = true;
		try {
			const query = new URLSearchParams({ page: String(page), limit: String(LIMIT) });
			if (action) query.set('action', action);
			if (userId) query.set('userId', userId);
			const response = await api.callWithAuth('get', `/admin/audit?${query}`);
			if (current !== request) return;
			entries = response.data.entries;
			names = response.data.users ?? {};
			total = response.data.total;
		} catch (error) {
			console.error(error);
			addToast({
				data: {
					title: 'Failed to load activity',
					description: 'Please try again later',
					color: 'bg-ctp-red'
				}
			});
		} finally {
			if (current === request) loading = false;
		}
	};

	let selected = '';
	$: if ($action.value !== selected) {
		selected = $action.value;
		page.set(1);
	}
	$: refresh, load(selected, $page, userId);
</script>

<div class="mb-4 flex flex-wrap items-end justify-between gap-4">
	<div class="w-full sm:w-72">
		<Select value={action} {options} label="Action" />
	</div>
	<Text color={TextColor.Muted}>{total} {total === 1 ? 'entry' : 'entries'}</Text>
</div>

<div class="overflow-x-auto">
	<table class="w-full text-left text-sm">
		<thead class="text-ctp-subtext0">
			<tr>
				<th class="px-3 py-2 font-semibold">When</th>
				<th class="px-3 py-2 font-semibold">Action</th>
				<th class="px-3 py-2 font-semibold">User</th>
				<th class="px-3 py-2 font-semibold">Details</th>
				<th class="px-3 py-2 font-semibold">IP</th>
			</tr>
		</thead>
		<tbody>
			{#each entries as entry (entry.id)}
				<tr class="border-t border-ctp-surface0 align-top">
					<td class="whitespace-nowrap px-3 py-2 text-ctp-subtext0"
						>{formatDate(entry.createdAt)}</td
					>
					<td class="px-3 py-2"><Badge tone={tone(entry.action)}>{entry.action}</Badge></td>
					<td class="px-3 py-2">
						{#if entry.userId}
							<a class="hover:text-ctp-mauve" href="/admin/users/{entry.userId}">
								{label(entry.userId)}
							</a>
						{:else}
							<span class="text-ctp-overlay1">Anonymous</span>
						{/if}
					</td>
					<td class="px-3 py-2 text-ctp-subtext0">
						{#if entry.target}
							{#if entry.target.type === 'user'}
								<a class="hover:text-ctp-mauve" href="/admin/users/{entry.target.id}">
									on {label(entry.target.id)}
								</a>
							{:else}
								<span class="font-mono text-xs">{entry.target.type} {entry.target.id}</span>
							{/if}
							<br />
						{/if}
						{details(entry)}
					</td>
					<td class="whitespace-nowrap px-3 py-2 font-mono text-xs text-ctp-subtext0"
						>{entry.ip ?? ''}</td
					>
				</tr>
			{/each}
		</tbody>
	</table>
</div>

{#if loading && entries.length === 0}
	<div class="flex justify-center py-8"><Loader /></div>
{:else if entries.length === 0}
	<Text color={TextColor.Muted} class="py-8 text-center">Nothing recorded.</Text>
{/if}

{#if total > LIMIT}
	<div class="mt-6">
		<Pagination {page} count={total} perPage={LIMIT} />
	</div>
{/if}
