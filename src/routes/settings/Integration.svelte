<script lang="ts">
	import { Check, Github } from 'lucide-svelte';
	import type { Writable } from 'svelte/store';
	import type { UserSettings } from './userSettings';

	const GITHUB_CLIENT_ID = '566de517d2c2d47ad218';

	export let userSettings: Writable<UserSettings | null>;
</script>

<h4 class="flex items-center gap-1 text-base font-semibold mb-2">
	<Github size="18" /> GitHub integration
</h4>
<p class="mb-2">
	Connect your GitHub account to enable features such as activity feed and repositories
	informations.
</p>
{#if $userSettings !== null}
	{#if $userSettings.github}
		<p class="flex items-center gap-1">
			<Check size="16" class="text-ctp-green" />
			<strong>
				Connected to GitHub as <a
					class="text-ctp-blue"
					href="https://github.com/{$userSettings.github.login}"
					target="_blank">{$userSettings.github.login}</a
				>.
			</strong>
		</p>
	{:else}
		<div class="flex justify-start">
			<a
				href={`https://github.com/login/oauth/authorize?client_id=${GITHUB_CLIENT_ID}&redirect_uri=${window.location.origin}/verify/github?path=/settings&scope=user:email`}
				class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1
  font-semibold text-ctp-mantle
  shadow-md shadow-ctp-crust transition-opacity
        hover:opacity-80 active:opacity-60"
			>
				<span>Connect to GitHub</span>
			</a>
		</div>
	{/if}
{/if}
