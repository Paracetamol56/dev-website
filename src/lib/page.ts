type Categories = 'sveltekit' | 'svelte';

type Page = {
	title: string;
	slug: string;
	description: string;
	categories: Categories[];
	listed: boolean;
};

export type { Categories, Page };
