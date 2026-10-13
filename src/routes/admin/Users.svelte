<script lang="ts">
	import { writable } from 'svelte/store';
	import api from '$lib/api';
	import { addToast } from '$lib/toast';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Loader from '$lib/components/ui/Loader.svelte';
	import Pagination from '$lib/components/ui/Pagination.svelte';
	import Text from '$lib/components/ui/Text.svelte';
	import TextInput from '$lib/components/ui/TextInput.svelte';
	import { Tone } from '$lib/components/ui/styles';
	import { TextColor } from '$lib/components/ui/typography';
	import { formatDate, type AdminUser } from './types';

	const LIMIT = 20;
	const search = writable('');
	const page = writable(1);

	let users: AdminUser[] = [];
	let total = 0;
	let loading = true;
	let request = 0;

	const load = async (search: string, page: number) => {
		const current = ++request;
		loading = true;
		try {
			const query = new URLSearchParams({ search, page: String(page), limit: String(LIMIT) });
			const response = await api.callWithAuth('get', `/admin/users?${query}`);
			if (current !== request) return;
			users = response.data.users;
			total = response.data.total;
		} catch (error) {
			console.error(error);
			addToast({
				data: {
					title: 'Failed to load users',
					description: 'Please try again later',
					color: 'bg-ctp-red'
				}
			});
		} finally {
			if (current === request) loading = false;
		}
	};

	let timeout: ReturnType<typeof setTimeout>;
	let applied = '';
	$: if ($search !== applied) {
		clearTimeout(timeout);
		const typed = $search;
		timeout = setTimeout(() => {
			applied = typed;
			page.set(1);
		}, 300);
	}
	$: load(applied, $page);
</script>

<div class="mb-4 flex flex-wrap items-end justify-between gap-4">
	<div class="w-full sm:w-72">
		<TextInput value={search} type="search" label="Search" placeholder="Email or name" />
	</div>
	<Text color={TextColor.Muted}>{total} {total === 1 ? 'user' : 'users'}</Text>
</div>

<div class="overflow-x-auto">
	<table class="w-full text-left text-sm">
		<thead class="text-ctp-subtext0">
			<tr>
				<th class="px-3 py-2 font-semibold">User</th>
				<th class="px-3 py-2 font-semibold">Sign-in</th>
				<th class="px-3 py-2 font-semibold">Created</th>
				<th class="px-3 py-2 font-semibold">Last login</th>
			</tr>
		</thead>
		<tbody>
			{#each users as user (user.id)}
				<tr class="border-t border-ctp-surface0 transition-colors hover:bg-ctp-surface0/40">
					<td class="px-3 py-2">
						<a class="block font-semibold hover:text-ctp-mauve" href="/admin/users/{user.id}">
							{user.name || 'No name'}
							{#if user.role === 'admin'}
								<Badge class="ml-1">Admin</Badge>
							{/if}
						</a>
						<span class="text-ctp-subtext0">{user.email}</span>
					</td>
					<td class="px-3 py-2">
						<div class="flex flex-wrap gap-1">
							{#each user.identities ?? [] as identity}
								<Badge tone={Tone.Neutral}>{identity.provider}</Badge>
							{/each}
							{#if user.passkeys?.length}
								<Badge tone={Tone.Neutral}>passkey</Badge>
							{/if}
						</div>
					</td>
					<td class="whitespace-nowrap px-3 py-2 text-ctp-subtext0">{formatDate(user.createdAt)}</td
					>
					<td class="whitespace-nowrap px-3 py-2 text-ctp-subtext0">{formatDate(user.lastLogin)}</td
					>
				</tr>
			{/each}
		</tbody>
	</table>
</div>

{#if loading && users.length === 0}
	<div class="flex justify-center py-8"><Loader /></div>
{:else if users.length === 0}
	<Text color={TextColor.Muted} class="py-8 text-center">No user found.</Text>
{/if}

{#if total > LIMIT}
	<div class="mt-6">
		<Pagination {page} count={total} perPage={LIMIT} />
	</div>
{/if}
