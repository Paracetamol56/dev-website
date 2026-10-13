<script lang="ts">
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import Activity from './Activity.svelte';
	import Users from './Users.svelte';

	const tabs = [
		{ id: 'users', title: 'Users' },
		{ id: 'activity', title: 'Activity' }
	];
</script>

<Tabs {tabs} label="Admin sections" let:id>
	{#if id === 'users'}
		<Users />
	{:else}
		<Activity />
	{/if}
</Tabs>
