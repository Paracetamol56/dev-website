<script lang="ts">
	import type { SelectOption } from '@melt-ui/svelte';
	import { Copy, Link2, Mail, RefreshCw, Trash2 } from 'lucide-svelte';
	import { writable, type Writable } from 'svelte/store';
	import api from '$lib/api';
	import Button from '$lib/components/ui/Button.svelte';
	import Dialog from '$lib/components/ui/Dialog.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import TextInput from '$lib/components/ui/TextInput.svelte';
	import { ButtonVariant, Tone } from '$lib/components/ui/styles';
	import { addToast } from '$lib/toast';
	import {
		createLink,
		emailLink,
		errorMessage,
		joinUrl,
		notifyError,
		removeMember,
		revokeLink,
		setMemberRole
	} from '../actions';
	import type { Deck, LinkRole, Links, Member } from '../types';

	export let open: Writable<boolean>;
	export let deckId: string;

	const roles: { role: LinkRole; title: string; description: string }[] = [
		{
			role: 'editor',
			title: 'Editor link',
			description: 'Can edit, rename and duplicate the deck'
		},
		{
			role: 'viewer',
			title: 'Viewer link',
			description: 'Can follow the deck live, without editing'
		}
	];
	const roleOptions: SelectOption<string>[] = [
		{ value: 'editor', label: 'Can edit' },
		{ value: 'viewer', label: 'Can view' }
	];

	let links: Links = {};
	let members: Member[] = [];
	let loading = false;
	let busy: string | null = null;

	const email = writable('');
	const emailRole = writable(roleOptions[1]);
	let emailError = '';

	const notify = (title: string, description: string) =>
		addToast({ data: { title, description, color: 'bg-ctp-green' } });

	async function load() {
		loading = true;
		try {
			const deck = (await api.callWithAuth('GET', `/slides/${deckId}`)).data as Deck;
			links = deck.links ?? {};
			members = deck.members ?? [];
		} catch (error) {
			notifyError(errorMessage(error, 'The sharing settings could not be loaded'));
		} finally {
			loading = false;
		}
	}

	$: if ($open) load();

	async function run(key: string, action: () => Promise<unknown>) {
		busy = key;
		await action();
		busy = null;
	}

	const turnOn = (role: LinkRole) =>
		run(`link-${role}`, async () => {
			const next = await createLink(deckId, role);
			if (next) links = next;
		});

	const renew = (role: LinkRole) =>
		run(`link-${role}`, async () => {
			const next = await createLink(deckId, role);
			if (next) {
				links = next;
				notify('New link', 'The previous link does not work anymore');
			}
		});

	const turnOff = (role: LinkRole) =>
		run(`link-${role}`, async () => {
			if (await revokeLink(deckId, role)) links = { ...links, [role]: undefined };
		});

	async function copy(token: string) {
		try {
			await navigator.clipboard.writeText(joinUrl(token));
			notify('Link copied', 'Send it to the people you want to invite');
		} catch {
			notifyError('The link could not be copied, select it and copy it by hand');
		}
	}

	function send() {
		const address = $email.trim();
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(address)) {
			emailError = 'Enter a valid email address';
			return;
		}
		run('email', async () => {
			if (await emailLink(deckId, $emailRole.value as LinkRole, address)) {
				notify('Email sent', `${address} received a link to the deck`);
				email.set('');
				await load();
			}
		});
	}

	const changeRole = (member: Member, role: LinkRole) =>
		run(`member-${member.user}`, async () => {
			if (role !== member.role && (await setMemberRole(deckId, member.user, role))) {
				members = members.map((other) => (other.user === member.user ? { ...other, role } : other));
			}
		});

	const remove = (member: Member) =>
		run(`member-${member.user}`, async () => {
			if (await removeMember(deckId, member.user)) {
				members = members.filter((other) => other.user !== member.user);
			}
		});
</script>

<Dialog
	{open}
	title="Share this deck"
	description="People who open a link join the deck once they are logged in."
>
	<div class="flex flex-col gap-8">
		<section class="flex flex-col gap-4">
			{#each roles as { role, title, description }}
				{@const token = links[role]}
				<div class="flex flex-col gap-3 rounded-md bg-ctp-mantle p-4 shadow-md shadow-ctp-crust">
					<div class="flex flex-wrap items-center justify-between gap-2">
						<div>
							<h3 class="font-semibold text-ctp-text">{title}</h3>
							<p class="text-xs text-ctp-subtext0">{description}</p>
						</div>
						{#if token}
							<Button
								variant={ButtonVariant.Ghost}
								tone={Tone.Danger}
								disabled={busy !== null}
								on:click={() => turnOff(role)}
							>
								Turn off
							</Button>
						{:else}
							<Button
								variant={ButtonVariant.Soft}
								loading={busy === `link-${role}`}
								disabled={loading || busy !== null}
								on:click={() => turnOn(role)}
							>
								<Link2 size="16" />
								Create link
							</Button>
						{/if}
					</div>
					{#if token}
						<div class="flex items-center gap-2">
							<input
								readonly
								value={joinUrl(token)}
								aria-label="{title} address"
								class="min-w-0 flex-1 truncate rounded-md bg-ctp-surface0 px-3 py-1 text-sm text-ctp-text shadow-md shadow-ctp-crust
									focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve"
								on:focus={(event) => event.currentTarget.select()}
							/>
							<IconButton label="Copy the {title.toLowerCase()}" on:click={() => copy(token)}>
								<Copy size="16" />
							</IconButton>
							<IconButton
								label="Replace the {title.toLowerCase()}"
								disabled={busy !== null}
								on:click={() => renew(role)}
							>
								<RefreshCw size="16" />
							</IconButton>
						</div>
					{/if}
				</div>
			{/each}
		</section>

		<form class="flex flex-col gap-4" on:submit|preventDefault={send}>
			<h3 class="text-lg font-semibold text-ctp-text">Send a link by email</h3>
			<div class="flex flex-wrap items-end gap-3">
				<div class="min-w-[12rem] flex-1">
					<TextInput
						type="email"
						label="Email address"
						placeholder="ada@example.com"
						value={email}
						error={emailError}
						on:input={() => (emailError = '')}
					/>
				</div>
				<Select label="Access" options={roleOptions} value={emailRole} />
			</div>
			<div class="flex justify-end">
				<Button type="submit" loading={busy === 'email'} disabled={busy !== null}>
					<Mail size="16" />
					Send
				</Button>
			</div>
		</form>

		<section class="flex flex-col gap-3">
			<h3 class="text-lg font-semibold text-ctp-text">People with access</h3>
			{#if members.length}
				<ul class="flex flex-col gap-2">
					{#each members as member (member.user)}
						<li class="flex items-center gap-3">
							<span class="min-w-0 flex-1 truncate text-sm font-semibold text-ctp-text">
								{member.name}
							</span>
							<div
								role="group"
								aria-label="Access of {member.name}"
								class="flex gap-0.5 rounded-md bg-ctp-mantle p-0.5"
							>
								{#each roleOptions as option}
									<button
										type="button"
										aria-pressed={member.role === option.value}
										disabled={busy !== null}
										class="rounded px-2 py-0.5 text-xs font-semibold text-ctp-subtext0 transition-colors
											hover:text-ctp-text focus:outline-none focus-visible:ring-2 focus-visible:ring-ctp-mauve
											aria-pressed:bg-ctp-mauve/20 aria-pressed:text-ctp-mauve disabled:opacity-50"
										on:click={() => changeRole(member, option.value as LinkRole)}
									>
										{option.label}
									</button>
								{/each}
							</div>
							<IconButton
								label="Remove {member.name}"
								tone={Tone.Danger}
								disabled={busy !== null}
								on:click={() => remove(member)}
							>
								<Trash2 size="16" />
							</IconButton>
						</li>
					{/each}
				</ul>
			{:else}
				<p class="text-sm text-ctp-subtext0">
					{loading ? 'Loading…' : 'Nobody joined this deck yet.'}
				</p>
			{/if}
		</section>
	</div>
</Dialog>
