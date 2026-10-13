<script lang="ts">
	import { ArrowLeft, ShieldCheck, ShieldOff, Trash2 } from 'lucide-svelte';
	import { writable } from 'svelte/store';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import api from '$lib/api';
	import { user as session } from '$lib/store';
	import { addToast } from '$lib/toast';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Dialog from '$lib/components/ui/Dialog.svelte';
	import Heading from '$lib/components/ui/Heading.svelte';
	import Link from '$lib/components/ui/Link.svelte';
	import Loader from '$lib/components/ui/Loader.svelte';
	import Text from '$lib/components/ui/Text.svelte';
	import { ButtonVariant, Tone } from '$lib/components/ui/styles';
	import { TextColor, TextSize } from '$lib/components/ui/typography';
	import Activity from '../../Activity.svelte';
	import { formatDate, type AdminUser } from '../../types';

	const confirmDelete = writable(false);

	let user: AdminUser | null = null;
	let missing = false;
	let busy = false;
	let refresh = 0;

	$: id = $page.params.id ?? '';
	$: own = id === $session.id;
	$: admin = user?.role === 'admin';

	const failed = (title: string, error: any) => {
		console.error(error);
		addToast({
			data: {
				title,
				description: error.response?.data?.error ?? 'Please try again later',
				color: 'bg-ctp-red'
			}
		});
	};

	const load = async (id: string) => {
		user = null;
		missing = false;
		try {
			user = (await api.callWithAuth('get', `/admin/users/${id}`)).data;
		} catch (error) {
			missing = true;
		}
	};
	$: load(id);

	const setRole = async (role: string) => {
		busy = true;
		try {
			user = (await api.callWithAuth('patch', `/admin/users/${id}/role`, { role })).data;
			refresh++;
			addToast({
				data: {
					title: role ? 'Promoted to admin' : 'Admin role removed',
					description: user?.email ?? '',
					color: 'bg-ctp-green'
				}
			});
		} catch (error) {
			failed('Failed to change the role', error);
		} finally {
			busy = false;
		}
	};

	const remove = async () => {
		busy = true;
		try {
			await api.callWithAuth('delete', `/admin/users/${id}`);
			confirmDelete.set(false);
			addToast({
				data: { title: 'Account deleted', description: user?.email ?? '', color: 'bg-ctp-green' }
			});
			goto('/admin');
		} catch (error) {
			failed('Failed to delete the account', error);
		} finally {
			busy = false;
		}
	};
</script>

<Link href="/admin" class="mb-6 inline-flex items-center gap-1">
	<ArrowLeft size="16" />
	All users
</Link>

{#if user}
	<Card class="mb-8">
		<div class="flex flex-wrap items-start justify-between gap-4">
			<div class="min-w-0">
				<Heading level={2} size={TextSize.Xl2} class="mb-1">
					{user.name || 'No name'}
					{#if admin}
						<Badge class="ml-2 align-middle">Admin</Badge>
					{/if}
				</Heading>
				<Text color={TextColor.Muted}>{user.email}</Text>
				<Text size={TextSize.Xs} color={TextColor.Subtle} class="font-mono">{user.id}</Text>
			</div>
			{#if !own}
				<div class="flex flex-wrap gap-2">
					<Button
						variant={ButtonVariant.Soft}
						loading={busy}
						on:click={() => setRole(admin ? '' : 'admin')}
					>
						{#if admin}
							<ShieldOff size="16" />
							<span>Remove admin</span>
						{:else}
							<ShieldCheck size="16" />
							<span>Make admin</span>
						{/if}
					</Button>
					<Button
						variant={ButtonVariant.Soft}
						tone={Tone.Danger}
						disabled={busy}
						on:click={() => confirmDelete.set(true)}
					>
						<Trash2 size="16" />
						<span>Delete</span>
					</Button>
				</div>
			{:else}
				<Text size={TextSize.Sm} color={TextColor.Muted}>This is your account.</Text>
			{/if}
		</div>

		<dl class="mt-6 grid gap-x-8 gap-y-4 text-sm sm:grid-cols-2 lg:grid-cols-4">
			<div>
				<dt class="font-semibold text-ctp-subtext0">Created</dt>
				<dd>{formatDate(user.createdAt)}</dd>
			</div>
			<div>
				<dt class="font-semibold text-ctp-subtext0">Last login</dt>
				<dd>{formatDate(user.lastLogin)}</dd>
			</div>
			<div>
				<dt class="font-semibold text-ctp-subtext0">Sign-in providers</dt>
				<dd>
					{#each user.identities ?? [] as identity}
						<div>
							{identity.provider}
							<span class="text-ctp-subtext0">
								{identity.username || identity.email} · {formatDate(identity.lastLogin)}
							</span>
						</div>
					{:else}
						None verified
					{/each}
				</dd>
			</div>
			<div>
				<dt class="font-semibold text-ctp-subtext0">Passkeys</dt>
				<dd>
					{#each user.passkeys ?? [] as passkey}
						<div>
							{passkey.name}
							<span class="text-ctp-subtext0">· used {formatDate(passkey.lastUsed)}</span>
						</div>
					{:else}
						None
					{/each}
				</dd>
			</div>
		</dl>
	</Card>

	<Heading level={3} size={TextSize.Xl} class="mb-4">Activity</Heading>
	<Card>
		<Activity userId={id} {refresh} />
	</Card>

	<Dialog
		open={confirmDelete}
		title="Delete this account?"
		description="{user.email} loses access at once and is told by email. The data is erased 30 days later."
	>
		<svelte:fragment slot="footer">
			<Button
				variant={ButtonVariant.Ghost}
				tone={Tone.Neutral}
				on:click={() => confirmDelete.set(false)}
			>
				Cancel
			</Button>
			<Button tone={Tone.Danger} loading={busy} on:click={remove}>Delete account</Button>
		</svelte:fragment>
	</Dialog>
{:else if missing}
	<Text color={TextColor.Muted} class="py-16 text-center">This user does not exist.</Text>
{:else}
	<div class="flex justify-center py-16"><Loader /></div>
{/if}
