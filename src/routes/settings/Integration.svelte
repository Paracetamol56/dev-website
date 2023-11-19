<script lang="ts">
	import { user } from '$lib/stores';
	import axios from 'axios';
	import { Github } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { addToast } from '../+layout.svelte';

	const GITHUB_CLIENT_ID = '566de517d2c2d47ad218';
	const redirect_uri = window.location.origin + '/api/auth/github';

	let loading = true;
	let githubUser: any = null;
	onMount(() => {
		axios
      .get(`/api/user/${$user!.id}`)
      .then((res) => {
        if (res.data.github) {
          githubUser = res.data.github;
        }
        loading = false;
      })
      .catch((err) => {
        console.error(err);
        addToast({
          data: {
            title: 'Error',
            description: 'Failed load user data',
            color: 'bg-ctp-red'
          }
        });
      });
	});
</script>

<h4 class="flex items-center gap-1 text-base font-semibold mb-2"><Github size="18" /> GitHub integration</h4>
{#if !loading}
	{#if githubUser}
		<p>
			Connected to GitHub as <a
				class="text-ctp-blue"
				href="https://github.com/{githubUser.login}"
				target="_blank">{githubUser.login}</a
			>.
		</p>
	{:else}
		<div class="flex justify-start">
			<a
				href={`https://github.com/login/oauth/authorize?client_id=${GITHUB_CLIENT_ID}&redirect_uri=${redirect_uri}?path=/settings&scope=user:email`}
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
