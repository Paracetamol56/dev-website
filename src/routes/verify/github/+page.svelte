<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import api from '$lib/api';
	import axios from 'axios';
	import { goto } from '$app/navigation';

	onMount(() => {
		const code: string = $page.url.searchParams.get('code') ?? '';
		const path: string = $page.url.searchParams.get('path') ?? '/';

		console.log(code, path);
		axios
			.post('/api/auth/github', { code })
			.then((res) => {
				api.persistUser(res);
				goto(path);
			})
			.catch((err) => {
				console.error(err);
				goto(path);
			});
	});
</script>
